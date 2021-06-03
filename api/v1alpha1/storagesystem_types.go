/*
Copyright 2021 Red Hat OpenShift Data Foundation.

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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//+kubebuilder:validation:Enum=FlashSystemCluster;StorageCluster

// StorageVendor captures the type of storage vendor
type StorageKind string

const (
	// FlashSystemCluster represents the ibm flashsystem
	FlashSystemCluster StorageKind = "FlashSystemCluster"

	// StorageCluster represents the openshift container storage
	StorageCluster StorageKind = "StorageCluster"
)

// StorageSystemSpec defines the desired state of StorageSystem
type StorageSystemSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required

	// Name describes the name of managed storage vendor CR
	Name string `json:"name,omitempty"`

	//+kubebuilder:validation:Optional

	// NameSpace describes the namespace of managed storage vendor CR
	NameSpace string `json:"nameSpace,omitempty"`

	//+kubebuilder:validation:Optional

	// Kind describes the kind of storage vendor
	Kind StorageKind `json:"kind,omitempty"`
}

// StorageSystemStatus defines the observed state of StorageSystem
type StorageSystemStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase describes the Phase of StorageSystem
	// This is used by OLM UI to provide status information
	// to the user
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=storsys

// StorageSystem is the Schema for the storagesystems API
type StorageSystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageSystemSpec   `json:"spec,omitempty"`
	Status StorageSystemStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StorageSystemList contains a list of StorageSystem
type StorageSystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageSystem `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageSystem{}, &StorageSystemList{})
}
