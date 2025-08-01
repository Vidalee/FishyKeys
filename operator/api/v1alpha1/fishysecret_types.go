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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FishySecretSpec defines the desired state of FishySecret.
type FishySecretSpec struct {
	// Token used to authenticate to the external secret manager
	Token string `json:"token,omitempty"`

	// Server is the endpoint of the external secret manager
	Server string `json:"server"`

	// Target defines the Kubernetes Secret to create
	Target SecretTarget `json:"target"`

	// Data is the list of key-paths to fetch from the secret manager
	// Each item becomes a key in the resulting Kubernetes Secret
	Data []SecretKeyMapping `json:"data"`
}

// SecretTarget defines where to create the Kubernetes Secret
type SecretTarget struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// SecretKeyMapping defines a mapping between a key in the secret manager
// and a key in the resulting Kubernetes Secret
type SecretKeyMapping struct {
	// SecretPath is the path in the secret manager
	SecretPath string `json:"secretPath"`

	// SecretKeyName is the field name in the resulting K8s Secret
	SecretKeyName string `json:"secretKeyName"`
}

// FishySecretStatus defines the observed state of FishySecret.
type FishySecretStatus struct {
	// Conditions represent the latest available observations of the FishySecret's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastSyncedTime is the last time the secret was successfully synced
	LastSyncedTime *metav1.Time `json:"lastSyncedTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// FishySecret is the Schema for the fishysecrets API.
// +kubebuilder:subresource:status
type FishySecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FishySecretSpec   `json:"spec,omitempty"`
	Status FishySecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FishySecretList contains a list of FishySecret.
type FishySecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FishySecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FishySecret{}, &FishySecretList{})
}
