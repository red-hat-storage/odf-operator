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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfcontrollers "github.com/red-hat-data-services/odf-operator/controllers"
)

// log is for logging in this package.
var subscriptionlog = logf.Log.WithName("subscription-resource")

//+kubebuilder:webhook:path=/mutate-operators-coreos-com-v1alpha1-subscription,mutating=true,failurePolicy=fail,sideEffects=None,groups=operators.coreos.com,resources=subscriptions,verbs=create;update,versions=v1alpha1,name=msubscription.kb.io,admissionReviewVersions={v1,v1beta1}

type SubscriptionDefaulter struct {
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &SubscriptionDefaulter{}

func (r *SubscriptionDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {

	instance := &operatorv1alpha1.Subscription{}
	err := r.decoder.Decode(req, instance)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	r.Default(instance)

	marshaledInstance, err := json.Marshal(instance)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledInstance)
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SubscriptionDefaulter) Default(instance *operatorv1alpha1.Subscription) {

	subscriptionlog.Info("default", "name", instance.Name)

	if instance.Spec.Package == odfcontrollers.IbmSubscriptionPackage && instance.Spec.Channel != odfcontrollers.IbmSubscriptionChannel {
		instance.Spec.Channel = odfcontrollers.IbmSubscriptionChannel
	}
}

func (r *SubscriptionDefaulter) SetupWebhookWithManager(mgr ctrl.Manager) error {

	mgr.GetWebhookServer().
		Register("/mutate-operators-coreos-com-v1alpha1-subscription",
			&webhook.Admission{Handler: r})

	return nil
}

func (r *SubscriptionDefaulter) InjectDecoder(decoder *admission.Decoder) error {
	r.decoder = decoder
	return nil
}
