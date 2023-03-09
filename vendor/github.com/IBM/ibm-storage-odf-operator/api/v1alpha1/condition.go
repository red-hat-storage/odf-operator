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

// Condition represents the state of the operator's
// reconciliation functionality.
// +k8s:deepcopy-gen=true
type Condition struct {
	Type ConditionType `json:"type" description:"type of condition."`

	Status corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`

	// +optional
	Reason string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`

	// +optional
	Message string `json:"message,omitempty" description:"human-readable message indicating details about last transition"`

	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime" description:"last time we got an update on a given condition"`

	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime" description:"last time the condition transit from one status to another"`
}

// ConditionType is the state of the operator's reconciliation functionality.
type ConditionType string

const (
	// ExporterCreated indicts exporter is launched by operator
	ExporterCreated ConditionType = "ExporterCreated"
	// ExporterReady is set from exporter and reason & message are provided if false condition
	ExporterReady ConditionType = "ExporterReady"
	// StorageClusterReady is set from exporter after query from FlashSystem
	StorageClusterReady ConditionType = "StorageClusterReady"
	// ProvisionerCreated indicts the FlashSystem CSI CR is created
	ProvisionerCreated ConditionType = "ProvisionerCreated"
	// ProvisionerReused indicts the existing FlashSystem CSI CR is reused
	ProvisionerReused ConditionType = "ProvisionerReused"
	// ProvisionerReady reflects the status of FlashSystem CSI CR
	ProvisionerReady ConditionType = "ProvisionerReady"
	// ConditionProgressing indicts the reconciling process is in progress
	ConditionProgressing ConditionType = "ReconcileProgressing"
	// ConditionReconcileComplete indicts the Reconcile function completes
	ConditionReconcileComplete ConditionType = "ReconcileComplete"
)

const (
	ReasonReconcileFailed    = "ReconcileFailed"
	ReasonReconcileInit      = "Init"
	ReasonReconcileCompleted = "ReconcileCompleted"
)
