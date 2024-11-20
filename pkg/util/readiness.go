/*
Copyright 2024 Red Hat OpenShift Data Foundation.

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
	"fmt"
	"net/http"
	"strings"

	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

func CheckCSVPhase(c client.Client, namespace string, csvNames ...string) healthz.Checker {

	csvMap := map[string]struct{}{}
	for _, name := range csvNames {
		csvMap[name] = struct{}{}
	}
	return func(r *http.Request) error {
		csvList := &opv1a1.ClusterServiceVersionList{}
		if err := c.List(r.Context(), csvList, client.InNamespace(namespace)); err != nil {
			return err
		}

		// Check if it is upgrade from 4.17 to 4.18
		// The new CSVs won't exists while upgrading
		// They will exists only after new operator has created a new subscription
		if AreMultipleOdfOperatorCsvsPresent(csvList) {
			return nil
		}

		for idx := range csvList.Items {
			csv := &csvList.Items[idx]
			_, exists := csvMap[csv.Name]
			if exists {
				if csv.Status.Phase != opv1a1.CSVPhaseSucceeded {
					return fmt.Errorf("CSV %s is not in Succeeded phase", csv.Name)
				} else if csv.Status.Phase == opv1a1.CSVPhaseSucceeded {
					delete(csvMap, csv.Name)
				}
			}
		}
		if len(csvMap) != 0 {
			for csvName := range csvMap {
				return fmt.Errorf("CSV %s is not found", csvName)
			}
		}
		return nil
	}
}

func AreMultipleOdfOperatorCsvsPresent(csvs *opv1a1.ClusterServiceVersionList) bool {

	count := 0

	for _, csv := range csvs.Items {
		if strings.HasPrefix(csv.Name, "odf-operator") {
			count += 1
		}
	}

	return count > 1
}
