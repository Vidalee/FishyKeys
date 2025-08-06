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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("FishySecret Controller", func() {
	Context("When reconciling a resource", func() {
		const fishySecretName = "test-fishysecret"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      fishySecretName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating a FishySecret resource")
			fishySecret := &fishykeysv1alpha1.FishySecret{}
			err := k8sClient.Get(ctx, typeNamespacedName, fishySecret)
			Expect(err).To(HaveOccurred(), "Expected to not find the FishySecret resource")
			if err != nil && errors.IsNotFound(err) {
				fishySecretToCreate := &fishykeysv1alpha1.FishySecret{
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
								SecretPath:    "db/user",
								SecretKeyName: "DB_USER",
							},
							{
								SecretPath:    "db/password",
								SecretKeyName: "DB_PASSWORD",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, fishySecretToCreate)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("cleaning up the FishySecret resource and associated Secret")
			fishySecret := &fishykeysv1alpha1.FishySecret{}
			err := k8sClient.Get(ctx, typeNamespacedName, fishySecret)
			Expect(err).NotTo(HaveOccurred())

			secrets := &corev1.SecretList{}
			err = k8sClient.List(ctx, secrets, &client.ListOptions{
				Namespace: typeNamespacedName.Namespace,
			})
			Expect(err).NotTo(HaveOccurred())

			// We are in envtest, so the garbage collector will not delete the associated Secret
			// when the FishySecret is deleted. We need to manually delete it.
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

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
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
			dummySecretValue := "dummy"
			expectedSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         fishykeysv1alpha1.GroupVersion.String(),
							Kind:               "FishySecret",
							Name:               fishySecretName,
							UID:                fishySecret.GetUID(),
							Controller:         func(b bool) *bool { return &b }(true),
							BlockOwnerDeletion: func(b bool) *bool { return &b }(true),
						},
					},
				},
				Data: map[string][]byte{
					"DB_USER":     []byte(dummySecretValue),
					"DB_PASSWORD": []byte(dummySecretValue),
				},
				Type: corev1.SecretTypeOpaque,
			}

			actualSecret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      expectedSecret.Name,
				Namespace: expectedSecret.Namespace,
			}, actualSecret)
			Expect(err).NotTo(HaveOccurred(), "Expected the secret to be retrieved")

			Expect(actualSecret.Name).To(Equal(expectedSecret.Name))
			Expect(actualSecret.Namespace).To(Equal(expectedSecret.Namespace))
			Expect(actualSecret.Type).To(Equal(expectedSecret.Type))
			Expect(actualSecret.Data).To(Equal(expectedSecret.Data))
			Expect(actualSecret.OwnerReferences).To(ContainElement(expectedSecret.OwnerReferences[0]))
		})
	})
})
