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
	"sync"

	"github.com/go-logr/logr"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/red-hat-storage/odf-operator/controllers"
)

type ClusterServiceVersionDeploymentScaler struct {
	client.Client

	Decoder           admission.Decoder
	OperatorNamespace string

	odfOperatorConfigAccessMutex        sync.Mutex
	odfOperatorConfigMapResourceVersion string
	odfOwnedCsvNames                    map[string]bool
}

//+kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;patch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get

func (r *ClusterServiceVersionDeploymentScaler) Handle(ctx context.Context, req admission.Request) admission.Response {

	logger := log.FromContext(ctx)
	logger.Info("request received for csv review")

	csv := &opv1a1.ClusterServiceVersion{}
	if err := r.Decoder.Decode(req, csv); err != nil {
		logger.Error(err, "failed decoding admission review as csv")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed decoding admission review as csv: %v", err))
	}

	if err := r.loadOdfConfigMapData(ctx, logger); err != nil {
		logger.Error(err, "failed to build config")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to build config: %v", err))
	}

	if ok := r.isCsvManagedByOdf(csv); !ok {
		logger.Info("ignoring csv as it is not a csv managed by ODF")
		return admission.Allowed("csv is not managed by ODF")
	}

	var isPrevCsvHasRunningDeployments bool
	if csv.Spec.Replaces != "" {
		prevCsv := &opv1a1.ClusterServiceVersion{}
		key := client.ObjectKey{Name: csv.Spec.Replaces, Namespace: csv.Namespace}
		if err := r.Client.Get(ctx, key, prevCsv); err != nil {
			if client.IgnoreNotFound(err) != nil {
				logger.Error(err, "failed to get previous CSV", "csv", csv.Spec.Replaces)
				return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed getting previous csv %q: %w", csv.Spec.Replaces, err))
			}
		} else {
			logger.Info("previous CSV found", "csv", csv.Spec.Replaces)
			isPrevCsvHasRunningDeployments = r.isCsvHasRunningDeployments(prevCsv)
			r.syncNewCsvWithPrevCsv(prevCsv, csv)
		}
	}

	if !isPrevCsvHasRunningDeployments {
		logger.Info("scaling down deployments")
		r.scaleDownCsvDeployments(logger, csv)
	}

	marshaledCsv, err := json.Marshal(csv)
	if err != nil {
		logger.Error(err, "failed marshaling csv")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed marshaling csv: %v", err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledCsv)
}

func (r *ClusterServiceVersionDeploymentScaler) loadOdfConfigMapData(ctx context.Context, logger logr.Logger) error {

	configmap, err := controllers.GetOdfConfigMap(ctx, r.Client, logger)
	if err != nil {
		return err
	}

	r.odfOperatorConfigAccessMutex.Lock()
	defer r.odfOperatorConfigAccessMutex.Unlock()

	if configmap.ResourceVersion == r.odfOperatorConfigMapResourceVersion {
		return nil
	}

	r.odfOwnedCsvNames = map[string]bool{}
	controllers.ParseOdfConfigMapRecords(logger, configmap, func(record *controllers.OdfOperatorConfigMapRecord, key, rawValue string) {
		if record.Csv == "" {
			logger.Info("skipping the record from the configmap", "key", key, "value", rawValue)
			return
		}

		r.odfOwnedCsvNames[record.Csv] = true
	})

	r.odfOperatorConfigMapResourceVersion = configmap.ResourceVersion
	logger.Info("webhook csv records", "records", r.odfOwnedCsvNames)

	return nil
}

// syncNewCsvWithPrevCsv copies the required fields from the previous csv that are required after upgrade in the new csv
func (r *ClusterServiceVersionDeploymentScaler) syncNewCsvWithPrevCsv(prevCsv *opv1a1.ClusterServiceVersion, newCsv *opv1a1.ClusterServiceVersion) {

	prevDeployments := prevCsv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs
	newDeployments := newCsv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs

	for i := range newDeployments {
		newDeployment := &newDeployments[i]

		for j := range prevDeployments {
			prevDeployment := &prevDeployments[j]

			if newDeployment.Name == prevDeployment.Name {
				newDeployment.Spec.Template.Spec.HostNetwork = prevDeployment.Spec.Template.Spec.HostNetwork
				newDeployment.Spec.Template.Spec.DNSPolicy = prevDeployment.Spec.Template.Spec.DNSPolicy
				newDeployment.Spec.Replicas = prevDeployment.Spec.Replicas
				break
			}
		}
	}
}

func (r *ClusterServiceVersionDeploymentScaler) isCsvHasRunningDeployments(csv *opv1a1.ClusterServiceVersion) bool {

	deployments := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs
	for i := range deployments {
		deployment := &deployments[i]
		if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas > 0 {
			return true
		}
	}

	return false
}

func (r *ClusterServiceVersionDeploymentScaler) scaleDownCsvDeployments(logger logr.Logger, csv *opv1a1.ClusterServiceVersion) {

	logger.Info("mutating requested csv")

	for i := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec.Replicas = ptr.To(int32(0))
	}
}

func (r *ClusterServiceVersionDeploymentScaler) isCsvManagedByOdf(csv *opv1a1.ClusterServiceVersion) bool {

	return r.odfOwnedCsvNames[csv.Name]
}

func (r *ClusterServiceVersionDeploymentScaler) SetupWebhookWithManager(mgr ctrl.Manager) error {

	mgr.GetWebhookServer().Register(controllers.CsvWebhookPath, &webhook.Admission{Handler: r})

	return nil
}
