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

package subscription

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
	odfcontrollers "github.com/red-hat-data-services/odf-operator/controllers"
)

func TestHandleDefaulter(t *testing.T) {

	testCases := []struct {
		label          string
		channel        string
		expectedChange bool
	}{
		{
			label:          "ensure it does not change channel if it is already as expected",
			channel:        odfcontrollers.IbmSubscriptionChannel,
			expectedChange: false,
		},
		{
			label:          "ensure it change the channel if it is not as expected",
			channel:        "fake-channel",
			expectedChange: true,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		scheme := runtime.NewScheme()
		utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))

		decoder, err := admission.NewDecoder(scheme)
		assert.NoError(t, err)

		r := &SubscriptionDefaulter{decoder: decoder}

		storageSystem := odfcontrollers.GetFakeStorageSystem()
		storageSystem.Spec.Kind = odfv1alpha1.FlashSystemCluster
		subscription := odfcontrollers.GetFlashSystemClusterSubscription(storageSystem)
		subscription.Spec.Channel = tc.channel
		rawSubscription, err := json.Marshal(subscription)
		assert.NoError(t, err)

		request := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: rawSubscription,
				},
			},
		}

		response := r.Handle(context.TODO(), request)
		assert.True(t, response.Allowed)

		if tc.expectedChange {
			for _, p := range response.Patches {
				assert.Equal(t, "replace", p.Operation)
				assert.Equal(t, "/spec/channel", p.Path)
				assert.Equal(t, odfcontrollers.IbmSubscriptionChannel, p.Value)
			}
		} else {
			assert.Equal(t, 0, len(response.Patches))
		}
	}
}
