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
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ibmv1alpha1 "github.com/IBM/ibm-storage-odf-operator/api/v1alpha1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	ocsv1 "github.com/openshift/ocs-operator/api/v1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

func GetCondition(condition conditionsv1.ConditionType, status corev1.ConditionStatus, reason, msg string) conditionsv1.Condition {
	return conditionsv1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: msg,
	}
}

func GetConditionResourcePresent(status corev1.ConditionStatus, reason, msg string) conditionsv1.Condition {
	return GetCondition(odfv1alpha1.ConditionResourcePresent, status, reason, msg)
}

func (r *StorageSystemReconciler) setConditionResourcePresent(instance *odfv1alpha1.StorageSystem, logger logr.Logger) (bool, error) {

	var requeue bool
	var err error

	if instance.Spec.Kind == odfv1alpha1.StorageCluster {
		logger.Info("get storageCluster")
		storageCluster := &ocsv1.StorageCluster{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Name, Namespace: instance.Spec.NameSpace}, storageCluster)
	} else if instance.Spec.Kind == odfv1alpha1.FlashSystemCluster {
		logger.Info("get flashSystemCluster")
		flashSystemCluster := &ibmv1alpha1.FlashSystemCluster{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Name, Namespace: instance.Spec.NameSpace}, flashSystemCluster)
	}

	if err == nil {
		logger.Info("set condition", string(odfv1alpha1.ConditionResourcePresent), corev1.ConditionTrue)
		conditionsv1.SetStatusCondition(&instance.Status.Conditions, GetConditionResourcePresent(corev1.ConditionTrue, "Found", ""))
		requeue = false
	} else if errors.IsNotFound(err) {
		logger.Error(err, "set condition", string(odfv1alpha1.ConditionResourcePresent), corev1.ConditionFalse)
		conditionsv1.SetStatusCondition(&instance.Status.Conditions, GetConditionResourcePresent(corev1.ConditionFalse, "NotFound", err.Error()))
		requeue = true
	} else {
		logger.Error(err, "set condition", string(odfv1alpha1.ConditionResourcePresent), corev1.ConditionUnknown)
		conditionsv1.SetStatusCondition(&instance.Status.Conditions, GetConditionResourcePresent(corev1.ConditionUnknown, "", err.Error()))
		requeue = true
	}

	return requeue, err
}
