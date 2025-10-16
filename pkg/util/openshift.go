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
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	opv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lib/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOpenShiftVersion(ctx context.Context, cl client.Client) (string, error) {
	clusterVersion := &configv1.ClusterVersion{}
	clusterVersion.Name = "version"
	if err := cl.Get(ctx, client.ObjectKeyFromObject(clusterVersion), clusterVersion); err != nil {
		return "", err
	}

	// Look for the latest completed version in history
	for _, historyEntry := range clusterVersion.Status.History {
		if historyEntry.State == configv1.CompletedUpdate {
			return historyEntry.Version, nil
		}
	}

	return "", fmt.Errorf("no completed version found in clusterVersion status history")
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
	return getConditionFactory(client).NewCondition(opv2.ConditionType(opv2.Upgradeable))
}
