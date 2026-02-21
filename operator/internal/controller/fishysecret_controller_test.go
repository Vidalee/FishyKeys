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
	pb "github.com/Vidalee/FishyKeys/operator/gen/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type mockSecretsServer struct {
	pb.UnimplementedSecretsServer
	expectedToken string
	// secrets maps base64-encoded secret paths to their values,
	// matching exactly what the controller sends over gRPC.
	secrets map[string]string
}

func (s *mockSecretsServer) OperatorGetSecretValue(ctx context.Context, req *pb.OperatorGetSecretValueRequest) (*pb.OperatorGetSecretValueResponse, error) {
	defer GinkgoRecover()

	md, _ := metadata.FromIncomingContext(ctx)
	Expect(md.Get("authorization")).To(ContainElement(s.expectedToken))

	value, ok := s.secrets[req.Path]
	Expect(ok).To(BeTrue(), "unexpected path requested: %s", req.Path)

	return &pb.OperatorGetSecretValueResponse{Value: &value}, nil
}

var _ = Describe("FishySecret Controller", func() {
	Context("When reconciling a resource", func() {
		const fishySecretName = "test-fishysecret"
		expectedServer := "localhost:8090"
		expectedToken := "secret token!"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      fishySecretName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating fishysecret-config Secret")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fishysecret-config",
					Namespace: typeNamespacedName.Namespace,
				},
				Data: map[string][]byte{
					"token": []byte(expectedToken),
					"url":   []byte(expectedServer),
				},
			}
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			By("creating a FishySecret resource")
			fishySecret := &fishykeysv1alpha1.FishySecret{}
			Expect(errors.IsNotFound(k8sClient.Get(ctx, typeNamespacedName, fishySecret))).To(BeTrue(), "FishySecret should not exist before the test")
			Expect(k8sClient.Create(ctx, &fishykeysv1alpha1.FishySecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fishySecretName,
					Namespace: "default",
				},
				Spec: fishykeysv1alpha1.FishySecretSpec{
					Target: fishykeysv1alpha1.SecretTarget{
						Name:      "test-secret",
						Namespace: "default",
					},
					Data: []fishykeysv1alpha1.SecretKeyMapping{
						{
							SecretPath:    "/db/user",
							SecretKeyName: "DB_USER",
						},
					},
				},
			})).To(Succeed())
		})

		AfterEach(func() {
			By("cleaning up the FishySecret resource and associated Secrets")
			fishySecret := &fishykeysv1alpha1.FishySecret{}
			err := k8sClient.Get(ctx, typeNamespacedName, fishySecret)
			Expect(err).NotTo(HaveOccurred())

			// Strip the finalizer so deletion is not blocked (controller is not auto-running in envtest)
			fishySecret.Finalizers = []string{}
			err = k8sClient.Update(ctx, fishySecret)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Delete(ctx, fishySecret)
			Expect(err).NotTo(HaveOccurred())

			// We are in envtest, so the garbage collector will not delete the associated Secrets
			// when the FishySecret is deleted. We need to manually delete them.
			secrets := &corev1.SecretList{}
			err = k8sClient.List(ctx, secrets, &client.ListOptions{
				Namespace: typeNamespacedName.Namespace,
			})
			Expect(err).NotTo(HaveOccurred())

			for _, s := range secrets.Items {
				err := k8sClient.Delete(ctx, &s)
				Expect(err).NotTo(HaveOccurred(), "Failed to manually delete secret %s", s.Name)
			}
		})
		It("should successfully reconcile the resource", func() {
			controllerReconciler := &FishySecretReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			srv := grpc.NewServer()
			pb.RegisterSecretsServer(srv, &mockSecretsServer{
				expectedToken: expectedToken,
				secrets: map[string]string{
					"L2RiL3VzZXI=": "secret value :p", // base64("/db/user")
				},
			})
			listener, err := net.Listen("tcp", expectedServer)
			Expect(err).ToNot(HaveOccurred())
			go func() {
				defer GinkgoRecover()
				_ = srv.Serve(listener) // ignore error, server is meant to run
			}()
			defer srv.Stop()

			time.Sleep(100 * time.Millisecond)

			// First reconcile: adds the finalizer and returns early
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Second reconcile: actually syncs the secret
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Check created FishySecret
			fishySecret := &fishykeysv1alpha1.FishySecret{}
			err = k8sClient.Get(ctx, typeNamespacedName, fishySecret)
			Expect(err).NotTo(HaveOccurred(), "Expected to get the FishySecret resource")

			Expect(fishySecret.Status.Conditions[0].Type).To(Equal("Ready"))
			Expect(fishySecret.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(fishySecret.Status.Conditions[0].Reason).To(Equal("SecretSynced"))
			Expect(fishySecret.Status.Conditions[0].Message).To(Equal("Secret successfully synced"))

			Expect(fishySecret.Status.LastSyncedTime).NotTo(BeNil(),
				"Expected the FishySecret status to have a LastSyncedTime set",
			)

			// Check created Secret
			actualSecret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "test-secret",
				Namespace: "default",
			}, actualSecret)
			Expect(err).NotTo(HaveOccurred(), "Expected the secret to be retrieved")

			Expect(actualSecret.Type).To(Equal(corev1.SecretTypeOpaque))
			Expect(actualSecret.Data).To(Equal(map[string][]byte{
				"DB_USER": []byte("secret value :p"),
			}))
			Expect(actualSecret.Labels).To(HaveKeyWithValue("fishykeys.2v.pm/owner-name", fishySecretName))
			Expect(actualSecret.Labels).To(HaveKeyWithValue("fishykeys.2v.pm/owner-namespace", "default"))
		})
	})
})
