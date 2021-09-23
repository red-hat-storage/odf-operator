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
	"testing"

	"github.com/stretchr/testify/assert"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
)

func TestEnsureSubscription(t *testing.T) {

	testCases := []struct {
		label                    string
		kind                     odfv1alpha1.StorageKind
		subscriptionAlreadyExist bool
		expectedSubscription     bool
	}{
		{
			label:                    "do not create subscription for StorageCluster",
			kind:                     VendorStorageCluster(),
			subscriptionAlreadyExist: false,
			expectedSubscription:     false,
		},
		{
			label:                    "create subscription for FlashSystemCluster if does not exist one",
			kind:                     VendorFlashSystemCluster(),
			subscriptionAlreadyExist: false,
			expectedSubscription:     true,
		},
		{
			label:                    "do not create subscription for FlashSystemCluster if it does exist",
			kind:                     VendorFlashSystemCluster(),
			subscriptionAlreadyExist: true,
			expectedSubscription:     true,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		fakeReconciler, fakeStorageSystem := GetFakeStorageSystemReconciler()
		err := operatorv1alpha1.AddToScheme(fakeReconciler.Scheme)
		assert.NoError(t, err)

		if tc.kind == VendorFlashSystemCluster() {
			fakeStorageSystem.Spec.Kind = tc.kind
		}

		if tc.subscriptionAlreadyExist {
			subscription := GetFlashSystemClusterSubscriptions()[0]
			subscription.Spec.Channel = "fake-channel"
			err := fakeReconciler.Client.Create(context.TODO(), subscription)
			assert.NoError(t, err)
		}

		err = fakeReconciler.ensureSubscription(fakeStorageSystem, fakeReconciler.Log)

		if tc.kind == VendorStorageCluster() {
			assert.Error(t, err)
			assert.Equal(t, "No subscription found with package name ocs-operator", err.Error())
		} else {
			assert.NoError(t, err)
		}

		existingSubscriptions := &operatorv1alpha1.SubscriptionList{}
		err = fakeReconciler.Client.List(context.TODO(), existingSubscriptions)
		assert.NoError(t, err)

		if !tc.expectedSubscription {
			assert.Equal(t, 0, len(existingSubscriptions.Items))
		} else {
			assert.Equal(t, 1, len(existingSubscriptions.Items))
			assert.Equal(t, IbmSubscriptionPackage, existingSubscriptions.Items[0].Spec.Package)
			assert.Equal(t, IbmSubscriptionChannel, existingSubscriptions.Items[0].Spec.Channel)
		}
	}
}
