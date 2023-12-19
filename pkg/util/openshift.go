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

package util

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	operatorv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lib/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DetermineOpenShiftVersion(client client.Client) (string, error) {
	// Determine ocp version
	clusterVersionList := configv1.ClusterVersionList{}
	if err := client.List(context.TODO(), &clusterVersionList); err != nil {
		return "", err
	}
	clusterVersion := ""
	for _, version := range clusterVersionList.Items {
		clusterVersion = version.Status.Desired.Version
	}
	return clusterVersion, nil
}

func getConditionFactory(client client.Client) conditions.Factory {
	return conditions.InClusterFactory{Client: client}
}

func GetConditionName(client client.Client) (string, error) {
	namespacedName, err := getConditionFactory(client).GetNamespacedName()
	if err != nil {
		return "", err
	}
	return namespacedName.Name, nil
}

func NewUpgradeableCondition(client client.Client) (conditions.Condition, error) {
	return getConditionFactory(client).NewCondition(operatorv2.ConditionType(operatorv2.Upgradeable))
}
