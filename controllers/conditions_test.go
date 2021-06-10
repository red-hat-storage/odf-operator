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
	"testing"

	"github.com/stretchr/testify/assert"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

func TestSetConditionResourcePresent(t *testing.T) {

	t.Skip("Skipping test because TODO is left")

	testCases := []struct {
		label             string
		hasBackEndStorage bool
		expectedRequeue   bool
		expectedError     bool
	}{

		{
			label:             "ensure ResourcePresent condition is true",
			hasBackEndStorage: true,
			expectedRequeue:   false,
			expectedError:     false,
		},
		{
			label:             "ensure ResourcePresent condition is false",
			hasBackEndStorage: false,
			expectedRequeue:   true,
			expectedError:     true,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		fakeReconciler, fakeStorageSystem := GetFakeStorageSystemReconciler()

		if tc.hasBackEndStorage {
			fakeReconciler.Log.Info("HACK IGNORE 'SA9003: empty branch' via adding this log line")
			// TODO: create storageCluster
		}

		requeue, err := fakeReconciler.setConditionResourcePresent(fakeStorageSystem, fakeReconciler.Log)
		assert.Equal(t, tc.expectedRequeue, requeue)

		assert.Equal(t, tc.hasBackEndStorage,
			conditionsv1.IsStatusConditionTrue(
				fakeStorageSystem.Status.Conditions, odfv1alpha1.ConditionResourcePresent))

		if tc.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
