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
	"net/http"
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

func TestHandleValidator(t *testing.T) {

	testCases := []struct {
		label     string
		channel   string
		Operation admissionv1.Operation
		Response  bool
	}{
		{
			label:     "make sure it accept the create request if channel is as expected",
			channel:   odfcontrollers.IbmSubscriptionChannel,
			Operation: admissionv1.Create,
			Response:  true,
		},
		{
			label:     "make sure it accept the update request if channel is as expected",
			channel:   odfcontrollers.IbmSubscriptionChannel,
			Operation: admissionv1.Update,
			Response:  true,
		},
		{
			label:     "make sure it accept the delete request if channel is as expected",
			channel:   odfcontrollers.IbmSubscriptionChannel,
			Operation: admissionv1.Delete,
			Response:  true,
		},
		{
			label:     "make sure it decline the create request if channel is not as expected",
			channel:   "fake-channel",
			Operation: admissionv1.Create,
			Response:  false,
		},
		{
			label:     "make sure it decline the update request if channel is not as expected",
			channel:   "fake-channel",
			Operation: admissionv1.Update,
			Response:  false,
		},
		{
			label:     "make sure it accept the delete request if channel is not as expected",
			channel:   "fake-channel",
			Operation: admissionv1.Delete,
			Response:  true,
		},
	}

	for i, tc := range testCases {
		t.Logf("Case %d: %s\n", i+1, tc.label)

		scheme := runtime.NewScheme()
		utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))

		decoder, err := admission.NewDecoder(scheme)
		assert.NoError(t, err)

		r := &SubscriptionValidator{decoder: decoder}

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
				OldObject: func() runtime.RawExtension {
					if tc.Operation == admissionv1.Update {
						return runtime.RawExtension{Raw: rawSubscription}
					} else {
						return runtime.RawExtension{}
					}
				}(),
				Operation: tc.Operation,
			},
		}

		response := r.Handle(context.TODO(), request)
		assert.Equal(t, tc.Response, response.Allowed)

		if tc.Response {
			assert.Equal(t, int32(http.StatusOK), response.Result.Code)
		} else {
			assert.Equal(t, int32(http.StatusBadRequest), response.Result.Code)
		}
	}
}
