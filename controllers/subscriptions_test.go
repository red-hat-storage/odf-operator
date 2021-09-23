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

	var err error
	testCases := []struct {
		label                    string
		kind                     odfv1alpha1.StorageKind
		subscriptionAlreadyExist bool
		expectedSubscriptions    int
	}{
		{
			label:                    "create subscriptions for StorageCluster",
			kind:                     VendorStorageCluster(),
			subscriptionAlreadyExist: false,
			expectedSubscriptions:    2,
		},
		{
			label:                    "create subscription for FlashSystemCluster if does not exist one",
			kind:                     VendorFlashSystemCluster(),
			subscriptionAlreadyExist: false,
			expectedSubscriptions:    1,
		},
		{
			label:                    "update subscription for FlashSystemCluster if it does exist",
			kind:                     VendorFlashSystemCluster(),
			subscriptionAlreadyExist: true,
			expectedSubscriptions:    1,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		fakeStorageSystem := GetFakeStorageSystem(tc.kind)
		fakeReconciler := GetFakeStorageSystemReconciler(t, fakeStorageSystem)

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
		assert.NoError(t, err)

		existingSubscriptions := &operatorv1alpha1.SubscriptionList{}
		err = fakeReconciler.Client.List(context.TODO(), existingSubscriptions)
		assert.NoError(t, err)

		assert.Equal(t, tc.expectedSubscriptions, len(existingSubscriptions.Items))

		if tc.kind == VendorFlashSystemCluster() {
			assert.Equal(t, IbmSubscriptionPackage, existingSubscriptions.Items[0].Spec.Package)
			assert.Equal(t, IbmSubscriptionChannel, existingSubscriptions.Items[0].Spec.Channel)
		}
	}
}
