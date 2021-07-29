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
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	odfcontrollers "github.com/red-hat-data-services/odf-operator/controllers"
)

//+kubebuilder:webhook:path=/validate-operators-coreos-com-v1alpha1-subscription,mutating=false,failurePolicy=fail,sideEffects=None,groups=operators.coreos.com,resources=subscriptions,verbs=create;update,versions=v1alpha1,name=vsubscription.kb.io,admissionReviewVersions={v1,v1beta1}

type SubscriptionValidator struct {
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &SubscriptionValidator{}

func (r *SubscriptionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	instance := &operatorv1alpha1.Subscription{}
	if err := r.decoder.Decode(req, instance); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	var validateErr error
	switch req.Operation {
	case admissionv1.Create:
		validateErr = r.ValidateCreate(instance)
	case admissionv1.Delete:
		validateErr = r.ValidateDelete(instance)
	case admissionv1.Update:
		old := &operatorv1alpha1.Subscription{}
		if err := r.decoder.DecodeRaw(req.OldObject, old); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		validateErr = r.ValidateUpdate(instance, old)
	}

	if validateErr != nil {
		return admission.Errored(http.StatusBadRequest, validateErr)
	}
	return admission.Allowed("")
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SubscriptionValidator) ValidateCreate(instance *operatorv1alpha1.Subscription) error {

	subscriptionlog.Info("validate create", "name", instance.Name)
	return r.ValidateSubscription(instance)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SubscriptionValidator) ValidateUpdate(instance, old *operatorv1alpha1.Subscription) error {

	subscriptionlog.Info("validate update", "name", instance.Name)
	return r.ValidateSubscription(instance)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SubscriptionValidator) ValidateDelete(instance *operatorv1alpha1.Subscription) error {

	subscriptionlog.Info("validate delete", "name", instance.Name)
	return nil
}

// ValidateSubscription validates the subscription
func (r *SubscriptionValidator) ValidateSubscription(instance *operatorv1alpha1.Subscription) error {

	subscriptionlog.Info("validate Subscription", "name", instance.Name)

	if instance.Spec.Package == odfcontrollers.IbmSubscriptionPackage && instance.Spec.Channel != odfcontrollers.IbmSubscriptionChannel {
		return fmt.Errorf("Spec.Channel can not be %s", instance.Spec.Channel)
	}

	return nil
}

func (r *SubscriptionValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/validate-operators-coreos-com-v1alpha1-subscription", &webhook.Admission{Handler: r})
	return nil
}

func (r *SubscriptionValidator) InjectDecoder(decoder *admission.Decoder) error {
	r.decoder = decoder
	return nil
}
