/*
Copyright 2025 Red Hat OpenShift Data Foundation.

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

	"github.com/go-logr/logr"
	admrv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	CsvWebhookPath = "/mutate-operators-coreos-com-v1alpha1-csv"
)

var (
	csvWebhookService = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "odf-operator-webhook-server-service",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       443,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt32(9443),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": "odf-operator",
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	csvWebhook = admrv1.MutatingWebhook{
		Name: "csv.odf.openshift.io",
		ClientConfig: admrv1.WebhookClientConfig{
			Service: &admrv1.ServiceReference{
				Name: csvWebhookService.Name,
				Path: ptr.To(CsvWebhookPath),
				Port: ptr.To(int32(443)),
			},
		},
		Rules: []admrv1.RuleWithOperations{{
			Rule: admrv1.Rule{
				APIGroups:   []string{"operators.coreos.com"},
				APIVersions: []string{"v1alpha1"},
				Resources:   []string{"clusterserviceversions"},
				Scope:       ptr.To(admrv1.NamespacedScope),
			},
			Operations: []admrv1.OperationType{admrv1.Create},
		}},
		SideEffects:             ptr.To(admrv1.SideEffectClassNone),
		TimeoutSeconds:          ptr.To(int32(30)),
		AdmissionReviewVersions: []string{"v1"},
		// fail the admission if webhook can't be reached
		FailurePolicy: ptr.To(admrv1.Fail),
	}
)

func reconcileCsvWebhook(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string, targetNamespaces []string) error {

	if err := reconcileCsvMutatingWebhookService(ctx, cli, logger, operatorNamespace); err != nil {
		logger.Error(err, "unable to reconcile webhook service")
		return err
	}

	if err := reconcileCsvMutatingWebhookConfiguration(ctx, cli, logger, operatorNamespace, targetNamespaces); err != nil {
		logger.Error(err, "unable to register csv mutating webhook")
		return err
	}

	return nil
}

func reconcileCsvMutatingWebhookService(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string) error {

	svc := &corev1.Service{}
	svc.Name = csvWebhookService.Name
	svc.Namespace = operatorNamespace

	res, err := controllerutil.CreateOrUpdate(ctx, cli, svc, func() error {
		if svc.Annotations == nil {
			svc.Annotations = map[string]string{}
		}
		svc.Annotations["service.beta.openshift.io/serving-cert-secret-name"] = "odf-operator-webhook-server-cert"
		csvWebhookService.Spec.DeepCopyInto(&svc.Spec)
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info("successfully created or updated webhook service", "operation", res, "name", svc.Name)
	return nil
}

func reconcileCsvMutatingWebhookConfiguration(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string, targetNamespaces []string) error {

	whConfig := &admrv1.MutatingWebhookConfiguration{}
	whConfig.Name = csvWebhook.Name

	res, err := controllerutil.CreateOrUpdate(ctx, cli, whConfig, func() error {

		if whConfig.Annotations == nil {
			whConfig.Annotations = map[string]string{}
		}
		// openshift fills in the ca on finding this annotation
		whConfig.Annotations["service.beta.openshift.io/inject-cabundle"] = "true"

		var caBundle []byte
		if len(whConfig.Webhooks) == 0 {
			whConfig.Webhooks = make([]admrv1.MutatingWebhook, 1)
		} else {
			// do not mutate CA bundle that was injected by openshift
			caBundle = whConfig.Webhooks[0].ClientConfig.CABundle
		}

		// webhook desired state
		wh := &whConfig.Webhooks[0]
		csvWebhook.DeepCopyInto(wh)

		wh.Name = whConfig.Name
		// only send requests received from own namespace
		wh.NamespaceSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "kubernetes.io/metadata.name",
					Operator: metav1.LabelSelectorOpIn,
					Values:   targetNamespaces,
				},
			},
		}

		wh.ObjectSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "olm.copiedFrom",
					Operator: metav1.LabelSelectorOpDoesNotExist,
				},
				{
					Key:      "odf.openshift.io/odf-operator",
					Operator: metav1.LabelSelectorOpDoesNotExist,
				},
			},
		}

		/*// create cel rules
		var conditions []string
		for _, pkg := range PkgNames {
			labelKey := fmt.Sprintf(CsvLabelKey, pkg, operatorNamespace)
			conditions = append(conditions, fmt.Sprintf(`object.metadata.labels.contains("%s")`, labelKey))
		}

		wh.MatchConditions = []admrv1.MatchCondition{{
			Name:       "has_at_least_one_target_label",
			Expression: strings.Join(conditions, " || "),
		}}*/

		// preserve the existing (injected) CA bundle if any
		wh.ClientConfig.CABundle = caBundle
		// send request to the service running in own namespace
		wh.ClientConfig.Service.Namespace = operatorNamespace

		return nil
	})
	if err != nil {
		return err
	}

	logger.Info("successfully created or updated webhook", "operation", res, "name", whConfig.Name)
	return nil
}
