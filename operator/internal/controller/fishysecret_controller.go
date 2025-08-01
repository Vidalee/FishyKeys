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
	fishykeysv1alpha1 "github.com/Vidalee/FishyKeys/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// FishySecretReconciler reconciles a FishySecret object
type FishySecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fishykeys.2v.pm,resources=fishysecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fishykeys.2v.pm,resources=fishysecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fishykeys.2v.pm,resources=fishysecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FishySecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *FishySecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var fishySecret fishykeysv1alpha1.FishySecret
	if err := r.Get(ctx, req.NamespacedName, &fishySecret); err != nil {
		log.Error(err, "unable to fetch FishySecret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secretData := make(map[string][]byte)
	for _, mapping := range fishySecret.Spec.Data {
		value, err := fetchSecretFromManager(fishySecret.Spec.Server, fishySecret.Spec.Token, mapping.SecretPath)
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
		},
		Data: secretData,
		Type: corev1.SecretTypeOpaque,
	}

	if err := ctrl.SetControllerReference(&fishySecret, desiredSecret, r.Scheme); err != nil {
		log.Error(err, "unable to set owner reference")
		return ctrl.Result{}, err
	}

	var existingSecret corev1.Secret
	err := r.Get(ctx, client.ObjectKeyFromObject(desiredSecret), &existingSecret)
	if err != nil && apierrors.IsNotFound(err) {
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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FishySecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fishykeysv1alpha1.FishySecret{}).
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

func fetchSecretFromManager(server string, token string, path string) (string, error) {
	return "dummy", nil
}
