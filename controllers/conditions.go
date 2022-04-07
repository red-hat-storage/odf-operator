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

package controllers

import (
	corev1 "k8s.io/api/core/v1"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
)

func SetAvailableCondition(conditions *[]conditionsv1.Condition,
	status corev1.ConditionStatus, reason string, message string) {

	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionAvailable,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

func SetProgressingCondition(conditions *[]conditionsv1.Condition,
	status corev1.ConditionStatus, reason string, message string) {

	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionProgressing,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

func SetStorageSystemInvalidCondition(conditions *[]conditionsv1.Condition,
	status corev1.ConditionStatus, reason string, message string) {

	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    odfv1alpha1.ConditionStorageSystemInvalid,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

func SetVendorCsvReadyCondition(conditions *[]conditionsv1.Condition,
	status corev1.ConditionStatus, reason string, message string) {

	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    odfv1alpha1.ConditionVendorCsvReady,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

func SetVendorSystemPresentCondition(conditions *[]conditionsv1.Condition,
	status corev1.ConditionStatus, reason string, message string) {

	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    odfv1alpha1.ConditionVendorSystemPresent,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

func SetReconcileInitConditions(conditions *[]conditionsv1.Condition,
	reason string, message string) {

	SetAvailableCondition(
		conditions, corev1.ConditionFalse, reason, message)
	SetProgressingCondition(
		conditions, corev1.ConditionTrue, reason, message)

	SetStorageSystemInvalidCondition(
		conditions, corev1.ConditionUnknown, reason, message)
	SetVendorCsvReadyCondition(
		conditions, corev1.ConditionUnknown, reason, message)
	SetVendorSystemPresentCondition(
		conditions, corev1.ConditionUnknown, reason, message)
}

func SetReconcileStartConditions(conditions *[]conditionsv1.Condition,
	reason string, message string) {

	SetAvailableCondition(
		conditions, corev1.ConditionFalse, reason, message)
	SetProgressingCondition(
		conditions, corev1.ConditionTrue, reason, message)
}

func SetReconcileCompleteConditions(conditions *[]conditionsv1.Condition,
	reason string, message string) {

	SetAvailableCondition(
		conditions, corev1.ConditionTrue, reason, message)
	SetProgressingCondition(
		conditions, corev1.ConditionFalse, reason, message)
}

func SetDeletionInProgressConditions(conditions *[]conditionsv1.Condition,
	reason string, message string) {

	SetAvailableCondition(
		conditions, corev1.ConditionFalse, reason, message)
	SetProgressingCondition(
		conditions, corev1.ConditionTrue, reason, message)
}

func SetStorageSystemInvalidConditions(conditions *[]conditionsv1.Condition,
	reason string, message string) {

	SetProgressingCondition(
		conditions, corev1.ConditionUnknown, reason, message)

	SetStorageSystemInvalidCondition(
		conditions, corev1.ConditionTrue, reason, message)
}
