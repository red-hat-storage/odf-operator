/**
 * Copyright contributors to the ibm-storage-odf-operator project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageClassConfig struct {
	StorageClassName string `json:"storageclassName"`
	PoolName         string `json:"poolName"`
	// +kubebuilder:validation:Enum=ext4;xfs
	FsType string `json:"fsType,omitempty"`
	// +kubebuilder:validation:MaxLength=20
	VolumeNamePrefix string `json:"volumeNamePrefix,omitempty"`
	// +kubebuilder:validation:Enum=thick;thin;compressed;deduplicated
	SpaceEfficiency string `json:"spaceEfficiency,omitempty"`
}

// FlashSystemClusterSpec defines the desired state of FlashSystemCluster
type FlashSystemClusterSpec struct {
	// Name is the name of the flashsystem storage cluster
	Name string `json:"name"`
	// Secret refers to a secret that has the credentials for flashsystem csi storageclass
	Secret corev1.SecretReference `json:"secret"`
	// InsecureSkipVerify disables target certificate validation if true
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// DefaultPool has the configuration to create default storage class
	DefaultPool *StorageClassConfig `json:"defaultPool,omitempty"`
}

// FlashSystemClusterStatus defines the observed state of FlashSystemCluster
type FlashSystemClusterStatus struct {
	// Phase describes the Phase of FlashSystemCluster
	// This is used by OLM UI to provide status information
	// to the user
	Phase string `json:"phase,omitempty"`

	// Conditions describes the state of the FlashSystemCluster resource.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// RelatedObjects is a list of objects created and maintained by this
	// operator. Object references will be added to this list after they have
	// been created AND found in the cluster.
	// +optional
	// RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=.metadata.creationTimestamp
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=.status.phase,description="Current Phase"
//+kubebuilder:printcolumn:name="Created At",type=string,JSONPath=.metadata.creationTimestamp

// FlashSystemCluster is the Schema for the flashsystemclusters API
type FlashSystemCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlashSystemClusterSpec   `json:"spec,omitempty"`
	Status FlashSystemClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlashSystemClusterList contains a list of FlashSystemCluster
type FlashSystemClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlashSystemCluster `json:"items"`
}

const (
	// PhaseProgressing is used during launch exporter & CSI CR creation phase
	PhaseProgressing = "Progressing"
	// PhaseError is used when reconcile fails or there is any of false ready from conditions
	PhaseError = "Error"
	// PhaseReady is used when reconcile is successful
	PhaseReady = "Ready"
	// PhaseNotReady is used if reconcile fails
	PhaseNotReady = "Not Ready"
	// PhaseDeleting is used if deleting flashsystemcluster is happening
	PhaseDeleting = "Deleting"
	// PhaseConnecting is reserved for later usage
	PhaseConnecting = "Connecting"
)

func init() {
	SchemeBuilder.Register(&FlashSystemCluster{}, &FlashSystemClusterList{})
}
