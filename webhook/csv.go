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

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/red-hat-storage/odf-operator/controllers"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ClusterServiceVersionDefaulter struct {
	client.Client

	Decoder           admission.Decoder
	OperatorNamespace string

	ctx context.Context
	log logr.Logger
}

// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;patch

func (r *ClusterServiceVersionDefaulter) Handle(ctx context.Context, req admission.Request) admission.Response {

	r.ctx = ctx
	r.log = log.FromContext(ctx)
	r.log.Info("request received for csv review")

	csv := &opv1a1.ClusterServiceVersion{}
	if err := r.Decoder.Decode(req, csv); err != nil {
		r.log.Error(err, "failed decoding admission review as csv")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed decoding admission review as csv: %v", err))
	}

	if ok := r.isOdfManagedCsv(csv); !ok {
		r.log.Info("ignoring requested csv as it is not relevant")
		return admission.Allowed("csv is not relevant to the odf")
	}

	replicas, err := r.getOldCsvReplicas(csv)
	if err != nil {
		r.log.Error(err, "failed getting replicas from the csv")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed getting replicas from the csv: %v", err))
	}

	r.mutateCsv(csv, &replicas)

	marshaledCsv, err := json.Marshal(csv)
	if err != nil {
		r.log.Error(err, "failed marshaling csv")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed marshaling csv: %v", err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledCsv)
}

func (r *ClusterServiceVersionDefaulter) getOldCsvReplicas(csv *opv1a1.ClusterServiceVersion) (int32, error) {

	if csv.Spec.Replaces == "" {
		r.log.Info("csv.Spec.Replaces is not populated")
		return 0, nil
	}

	oldCsv := &opv1a1.ClusterServiceVersion{}
	key := client.ObjectKey{Name: csv.Spec.Replaces, Namespace: csv.Namespace}

	if err := r.Client.Get(r.ctx, key, oldCsv); err != nil {
		if errors.IsNotFound(err) {
			// new install where an older csv does not exists
			return 0, nil
		}
		r.log.Error(err, "failed getting older csv")
		return 0, err
	}

	for _, deployment := range oldCsv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		if *deployment.Spec.Replicas > 0 {
			// upgrade case where an older csv is found with replica 1
			return 1, nil
		}
	}

	// upgrade case where an older csv is found with replica 0
	return 0, nil
}

func (r *ClusterServiceVersionDefaulter) mutateCsv(csv *opv1a1.ClusterServiceVersion, replicas *int32) {

	r.log.Info("mutating requested csv")
	for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec.Replicas = replicas
	}
}

func (r *ClusterServiceVersionDefaulter) isOdfManagedCsv(csv *opv1a1.ClusterServiceVersion) bool {

	if csv.Namespace != r.OperatorNamespace {
		return false
	}

	for _, pkgName := range controllers.PkgNames {
		if strings.HasPrefix(csv.Name, pkgName) {
			return true
		}
	}

	return false
}

func (r *ClusterServiceVersionDefaulter) SetupWebhookWithManager(mgr ctrl.Manager) error {

	mgr.GetWebhookServer().Register(controllers.WebhookPath, &webhook.Admission{Handler: r})

	return nil
}
