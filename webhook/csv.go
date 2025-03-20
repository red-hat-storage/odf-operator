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
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ClusterServiceVersionDeploymentScaler struct {
	client.Client

	Decoder           admission.Decoder
	OperatorNamespace string
}

// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;patch

func (r *ClusterServiceVersionDeploymentScaler) Handle(ctx context.Context, req admission.Request) admission.Response {

	logger := log.FromContext(ctx)
	logger.Info("request received for csv review")

	csv := &opv1a1.ClusterServiceVersion{}
	if err := r.Decoder.Decode(req, csv); err != nil {
		logger.Error(err, "failed decoding admission review as csv")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed decoding admission review as csv: %v", err))
	}

	if ok := r.isCsvManagedByOdf(csv); !ok {
		logger.Info("ignoring csv as it is not a csv managed by ODF")
		return admission.Allowed("csv is not managed by ODF")
	}

	running, err := r.isPreviousCsvHasRunningDeployments(ctx, logger, csv)
	if err != nil {
		logger.Error(err, "failed getting replicas from csv")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed getting replicas from csv: %v", err))
	}

	if running {
		logger.Info("ignoring csv as the previous csv deployments are running")
		return admission.Allowed("previous csv deployments are running")
	}

	r.scaleDownCsvDeployments(logger, csv)

	marshaledCsv, err := json.Marshal(csv)
	if err != nil {
		logger.Error(err, "failed marshaling csv")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed marshaling csv: %v", err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledCsv)
}

func (r *ClusterServiceVersionDeploymentScaler) isPreviousCsvHasRunningDeployments(ctx context.Context, logger logr.Logger, csv *opv1a1.ClusterServiceVersion) (bool, error) {

	if csv.Spec.Replaces == "" {
		logger.Info("csv.Spec.Replaces is not populated")
		return false, nil
	}

	prevCsv := &opv1a1.ClusterServiceVersion{}
	key := client.ObjectKey{Name: csv.Spec.Replaces, Namespace: csv.Namespace}

	if err := r.Client.Get(ctx, key, prevCsv); errors.IsNotFound(err) {
		// new install where an previous csv does not exists
		return false, nil
	} else if err != nil {
		logger.Error(err, "failed getting previous csv")
		return false, err
	}

	deployments := prevCsv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs
	for i := range deployments {
		deployment := &deployments[i]
		if *deployment.Spec.Replicas > 0 {
			// upgrade case where an older csv is found with replica 1
			return true, nil
		}
	}

	// upgrade case where an older csv is found with replica 0
	return false, nil
}

func (r *ClusterServiceVersionDeploymentScaler) scaleDownCsvDeployments(logger logr.Logger, csv *opv1a1.ClusterServiceVersion) {

	logger.Info("mutating requested csv")

	for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec.Replicas = ptr.To(int32(0))
	}
}

func (r *ClusterServiceVersionDeploymentScaler) isCsvManagedByOdf(csv *opv1a1.ClusterServiceVersion) bool {

	if csv.Namespace != r.OperatorNamespace {
		return false
	}

	for i := range controllers.ResourceMappingList {
		for _, pkgName := range controllers.ResourceMappingList[i].PkgNames {
			if strings.HasPrefix(csv.Name, pkgName) {
				return true
			}
		}
	}

	return false
}

func (r *ClusterServiceVersionDeploymentScaler) SetupWebhookWithManager(mgr ctrl.Manager) error {

	mgr.GetWebhookServer().Register(controllers.CsvWebhookPath, &webhook.Admission{Handler: r})

	return nil
}
