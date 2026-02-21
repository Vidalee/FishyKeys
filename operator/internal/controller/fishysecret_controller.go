/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	fishykeysv1alpha1 "github.com/Vidalee/FishyKeys/operator/api/v1alpha1"
	pb "github.com/Vidalee/FishyKeys/operator/gen/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const fishySecretFinalizer = "fishykeys.2v.pm/finalizer"

// FishySecretReconciler reconciles a FishySecret object
type FishySecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fishykeys.2v.pm,resources=fishysecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fishykeys.2v.pm,resources=fishysecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fishykeys.2v.pm,resources=fishysecrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *FishySecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var fishySecret fishykeysv1alpha1.FishySecret
	if err := r.Get(ctx, req.NamespacedName, &fishySecret); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource deleted, nothing to do
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch FishySecret")
		return ctrl.Result{}, err
	}

	if !fishySecret.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Finalizing FishySecret", "name", fishySecret.Name)

		target := fishySecret.Spec.Target
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      target.Name,
				Namespace: target.Namespace,
			},
		}
		// In case our Secret was deleted manually we may not find it, ignore this error
		if err := r.Delete(ctx, secret); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "unable to delete child Secret", "name", target.Name)
			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(&fishySecret, fishySecretFinalizer)
		if err := r.Update(ctx, &fishySecret); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&fishySecret, fishySecretFinalizer) {
		controllerutil.AddFinalizer(&fishySecret, fishySecretFinalizer)
		if err := r.Update(ctx, &fishySecret); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	secretData := make(map[string][]byte)

	serverUrl, token, err := getFishyKeysUrlAndToken(ctx, r)
	if err != nil {
		log.Error(err, "failed to get FishyKeys token")
		setStatusCondition(&fishySecret, "Ready", metav1.ConditionFalse, "TokenError", err.Error())
		_ = r.Status().Update(ctx, &fishySecret)
		return ctrl.Result{}, err
	}

	conn, err := grpc.NewClient(serverUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "failed to connect to FishyKeys server")
		setStatusCondition(&fishySecret, "Ready", metav1.ConditionFalse, "ConnectionError", err.Error())
		_ = r.Status().Update(ctx, &fishySecret)
		return ctrl.Result{}, err
	}
	defer conn.Close()

	grpcCtx, grpcCancel := context.WithTimeout(ctx, 5*time.Second)
	defer grpcCancel()
	grpcCtx = metadata.NewOutgoingContext(grpcCtx, metadata.Pairs("authorization", token))
	secretsClient := pb.NewSecretsClient(conn)

	for _, mapping := range fishySecret.Spec.Data {
		value, err := fetchSecretValue(grpcCtx, secretsClient, mapping.SecretPath)
		if err != nil {
			log.Error(err, "failed to fetch from secret manager", "path", mapping.SecretPath)
			setStatusCondition(&fishySecret, "Ready", metav1.ConditionFalse, "FetchError", err.Error())
			_ = r.Status().Update(ctx, &fishySecret)
			return ctrl.Result{}, err
		}
		secretData[mapping.SecretKeyName] = []byte(value)
	}

	desiredSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fishySecret.Spec.Target.Name,
			Namespace: fishySecret.Spec.Target.Namespace,
			Labels: map[string]string{
				"fishykeys.2v.pm/owner-name":      fishySecret.Name,
				"fishykeys.2v.pm/owner-namespace": fishySecret.Namespace,
			},
		},
		Data: secretData,
		Type: corev1.SecretTypeOpaque,
	}

	var existingSecret corev1.Secret
	err = r.Get(ctx, client.ObjectKeyFromObject(desiredSecret), &existingSecret)
	if apierrors.IsNotFound(err) {
		if err := r.Create(ctx, desiredSecret); err != nil {
			log.Error(err, "failed to create Secret")
			return ctrl.Result{}, err
		}
		log.Info("Secret created", "name", desiredSecret.Name)
	} else if err == nil {
		if !reflect.DeepEqual(existingSecret.Data, desiredSecret.Data) {
			existingSecret.Data = desiredSecret.Data
			if err := r.Update(ctx, &existingSecret); err != nil {
				log.Error(err, "failed to update Secret")
				return ctrl.Result{}, err
			}
			log.Info("Secret updated", "name", desiredSecret.Name)
		}
	} else {
		log.Error(err, "failed to get existing Secret")
		return ctrl.Result{}, err
	}

	setStatusCondition(&fishySecret, "Ready", metav1.ConditionTrue, "SecretSynced", "Secret successfully synced")
	fishySecret.Status.LastSyncedTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, &fishySecret); err != nil {
		log.Error(err, "failed to update FishySecret status")
		return ctrl.Result{}, err
	}

	// Schedule a reconciliation in 5 minutes, in case the corresponding secrets in are updated in the backend
	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// mapSecretToFishySecrets maps Secret events to owning FishySecret resources
func (r *FishySecretReconciler) mapSecretToFishySecrets(ctx context.Context, obj client.Object) []reconcile.Request {
	ns := obj.GetLabels()["fishykeys.2v.pm/owner-namespace"]
	name := obj.GetLabels()["fishykeys.2v.pm/owner-name"]

	if ns == "" || name == "" {
		return nil
	}

	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
	}}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FishySecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fishykeysv1alpha1.FishySecret{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.mapSecretToFishySecrets),
		).
		Named("fishysecret").
		Complete(r)
}

func setStatusCondition(f *fishykeysv1alpha1.FishySecret, conditionType string, status metav1.ConditionStatus, reason string, msg string) {
	meta.SetStatusCondition(&f.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	})
}

func fetchSecretValue(ctx context.Context, secretsClient pb.SecretsClient, path string) (string, error) {
	encodedPath := base64.StdEncoding.EncodeToString([]byte(path))

	resp, err := secretsClient.OperatorGetSecretValue(ctx, &pb.OperatorGetSecretValueRequest{
		Path: encodedPath,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	return resp.GetValue(), nil
}

func getFishyKeysUrlAndToken(ctx context.Context, r *FishySecretReconciler) (string, string, error) {
	operatorNamespace := os.Getenv("POD_NAMESPACE")
	if operatorNamespace == "" {
		if os.Getenv("ENV") == "DEV" {
			operatorNamespace = "default"
		} else {
			return "", "", fmt.Errorf("POD_NAMESPACE environment variable is not set")
		}
	}

	var tokenSecret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{
		Name:      "fishysecret-config",
		Namespace: operatorNamespace,
	}, &tokenSecret); err != nil {
		return "", "", fmt.Errorf("unable to get token secret in namespace %s: %w. Please create it", operatorNamespace, err)
	}

	tokenBytes, ok := tokenSecret.Data["token"]
	if !ok {
		return "", "", fmt.Errorf("invalid token secret: missing 'token' key in Secret fishysecret-config")
	}
	urlBytes, ok := tokenSecret.Data["url"]
	if !ok {
		return "", "", fmt.Errorf("invalid token secret: missing 'url' key in Secret fishysecret-config")
	}

	token := string(tokenBytes)
	url := string(urlBytes)
	return url, token, nil
}
