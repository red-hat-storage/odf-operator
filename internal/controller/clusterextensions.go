/*
Copyright 2026 Data Foundation.

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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"

	"github.com/go-logr/logr"
	opv1a1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ocv1 "github.com/operator-framework/operator-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/red-hat-storage/odf-operator/internal/util"
)

type clusterExtensionConfig struct {
	DeploymentConfig opv1a1.SubscriptionConfig `json:"deploymentConfig,omitempty"`
}

func (c *clusterExtensionConfig) hasPodPlacements() bool {

	if c.DeploymentConfig.Affinity != nil ||
		len(c.DeploymentConfig.Tolerations) > 0 ||
		len(c.DeploymentConfig.NodeSelector) > 0 {
		return true
	}

	return false
}

func ensureClusterExtensions(
	ctx context.Context, cli client.Client, logger logr.Logger, olmPkgRecords []*OlmPkgRecord, publisherName string) error {

	odfOperatorClusterExtension, err := getOdfOperatorClusterExtension(ctx, cli, logger)
	if err != nil {
		logger.Error(err, "failed to get odf-operator ClusterExtension")
		return err
	}

	for _, record := range olmPkgRecords {

		if slices.Contains(ExtendedFdfPackages, record.Package) && publisherName != PublisherNameIBM {
			// Do not install extended pkgs
			continue
		}

		desiredClusterExtension, err := getDesiredClusterExtension(logger, record, odfOperatorClusterExtension)
		if err != nil {
			logger.Error(err, "failed to get desired ClusterExtension", "ClusterExtension", record.Package)
			return err
		}

		if err := createOrUpdateClusterExtension(ctx, cli, logger, desiredClusterExtension); err != nil {
			logger.Error(err, "failed to create or update ClusterExtension", "ClusterExtension", record.Package)
			return err
		}
	}

	return nil
}

func getOdfOperatorClusterExtension(ctx context.Context, cli client.Client, logger logr.Logger) (*ocv1.ClusterExtension, error) {

	clusterExtensionList := &ocv1.ClusterExtensionList{}
	if err := cli.List(ctx, clusterExtensionList); err != nil {
		logger.Error(err, "failed to list ClusterExtensions")
		return nil, err
	}

	for _, ce := range clusterExtensionList.Items {
		if ce.Spec.Source.Catalog != nil && ce.Spec.Source.Catalog.PackageName == OdfOperatorPackageName {
			logger.Info("found odf-operator ClusterExtension", "name", ce.Name)
			return &ce, nil
		}
	}

	err := fmt.Errorf("could not found odf-operator ClusterExtension")

	logger.Error(err, "odf-operator ClusterExtension not found")

	return nil, err
}

func getDesiredClusterExtension(
	logger logr.Logger, record *OlmPkgRecord, odfOperatorClusterExtension *ocv1.ClusterExtension) (*ocv1.ClusterExtension, error) {

	clusterExtension := &ocv1.ClusterExtension{
		ObjectMeta: metav1.ObjectMeta{
			Name: record.Package,
			Labels: map[string]string{
				ManagedByKey: ManagedByValOdfOperator,
			},
		},
		Spec: ocv1.ClusterExtensionSpec{
			Namespace: record.Namespace,
			ServiceAccount: ocv1.ServiceAccountReference{
				Name: odfOperatorClusterExtension.Spec.ServiceAccount.Name,
			},
			Source: ocv1.SourceConfig{
				SourceType: ocv1.SourceTypeCatalog,
				Catalog: &ocv1.CatalogFilter{
					PackageName: record.Package,
					Version:     record.Version,
				},
			},
		},
	}

	if err := handleClusterExtensionConfigSpecialCases(logger, clusterExtension); err != nil {
		logger.Error(err, "failed to handle special cases for ClusterExtension", "ClusterExtension", clusterExtension.Name)
		return nil, err
	}

	if err := inheritOdfOperatorClusterExtensionPlacements(logger, clusterExtension, odfOperatorClusterExtension); err != nil {
		logger.Error(err, "failed to inherit tolerations for ClusterExtension", "ClusterExtension", clusterExtension.Name)
		return nil, err
	}

	return clusterExtension, nil
}

func handleClusterExtensionConfigSpecialCases(logger logr.Logger, clusterExtension *ocv1.ClusterExtension) error {

	if clusterExtension.Spec.Source.Catalog == nil {
		return nil
	}

	config := &clusterExtensionConfig{}

	switch clusterExtension.Spec.Source.Catalog.PackageName {

	case "noobaa-operator", "mcg-operator":

		var envVars []corev1.EnvVar
		for _, name := range []string{
			"ROLEARN", "CLIENTID", "TENANTID", "SUBSCRIPTIONID", "RESOURCEGROUP",
			"PROJECT_NUMBER", "POOL_ID", "PROVIDER_ID", "SERVICE_ACCOUNT_EMAIL",
		} {
			if value := os.Getenv(name); value != "" {
				envVars = append(envVars, corev1.EnvVar{Name: name, Value: value})
			}
		}

		if len(envVars) == 0 {
			return nil
		}

		config.DeploymentConfig = opv1a1.SubscriptionConfig{
			Env: envVars,
		}

	case "csi-addons", "odf-csi-addons-operator", "cephcsi-operator":

		config.DeploymentConfig = opv1a1.SubscriptionConfig{
			Tolerations: []corev1.Toleration{
				{
					Key:      "node.ocs.openshift.io/storage",
					Operator: "Equal",
					Value:    "true",
					Effect:   "NoSchedule",
				},
			},
		}

	default:
		return nil
	}

	rawConfig, err := json.Marshal(config)
	if err != nil {
		logger.Error(err, "failed to marshal config")
		return err
	}

	clusterExtension.Spec.Config = &ocv1.ClusterExtensionConfig{
		ConfigType: ocv1.ClusterExtensionConfigTypeInline,
		Inline: &apiextensionsv1.JSON{
			Raw: rawConfig,
		},
	}

	return nil
}

func inheritOdfOperatorClusterExtensionPlacements(
	logger logr.Logger, targetClusterExtension, odfOperatorClusterExtension *ocv1.ClusterExtension) error {

	odfConfig, err := parseClusterExtensionConfig(logger, odfOperatorClusterExtension)
	if err != nil {
		logger.Error(err, "failed to get ClusterExtension config", "ClusterExtension", odfOperatorClusterExtension.Name)
		return err
	}

	if !odfConfig.hasPodPlacements() {
		return nil
	}

	targetConfig, err := parseClusterExtensionConfig(logger, targetClusterExtension)
	if err != nil {
		logger.Error(err, "failed to get ClusterExtension config", "ClusterExtension", targetClusterExtension.Name)
		return err
	}

	targetConfig.DeploymentConfig.Tolerations = util.AppendUniqueFunc(
		targetConfig.DeploymentConfig.Tolerations, odfConfig.DeploymentConfig.Tolerations,
		func(a, b corev1.Toleration) bool {
			return reflect.DeepEqual(a, b)
		},
	)

	targetConfig.DeploymentConfig.Affinity = odfConfig.DeploymentConfig.Affinity

	targetConfig.DeploymentConfig.NodeSelector = odfConfig.DeploymentConfig.NodeSelector

	rawConfig, err := json.Marshal(targetConfig)
	if err != nil {
		logger.Error(err, "failed to marshal ClusterExtension config")
		return err
	}

	targetClusterExtension.Spec.Config = &ocv1.ClusterExtensionConfig{
		ConfigType: ocv1.ClusterExtensionConfigTypeInline,
		Inline: &apiextensionsv1.JSON{
			Raw: rawConfig,
		},
	}

	return nil
}

func parseClusterExtensionConfig(logger logr.Logger, clusterExtension *ocv1.ClusterExtension) (*clusterExtensionConfig, error) {
	if clusterExtension.Spec.Config == nil ||
		clusterExtension.Spec.Config.Inline == nil ||
		len(clusterExtension.Spec.Config.Inline.Raw) == 0 {
		return &clusterExtensionConfig{}, nil
	}

	config := &clusterExtensionConfig{}
	if err := json.Unmarshal(clusterExtension.Spec.Config.Inline.Raw, &config); err != nil {
		logger.Error(err, "failed to unmarshal ClusterExtension inline raw config", "ClusterExtension", clusterExtension.Name)
		return &clusterExtensionConfig{}, err
	}

	return config, nil
}

func createOrUpdateClusterExtension(
	ctx context.Context, cli client.Client, logger logr.Logger, desiredClusterExtension *ocv1.ClusterExtension) error {

	clusterExtension := &ocv1.ClusterExtension{}
	clusterExtension.Name = desiredClusterExtension.Name

	op, err := controllerutil.CreateOrUpdate(ctx, cli, clusterExtension, func() error {

		if clusterExtension.Annotations != nil {
			val, ok := clusterExtension.Annotations[ManagedByKey]
			if ok && val == ManagedByValUser {
				logger.Info("ClusterExtension is managed by the user, skipping update", "ClusterExtension", clusterExtension.Name)
				return nil
			}
		}

		if clusterExtension.Annotations == nil {
			clusterExtension.Annotations = map[string]string{}
		}

		if clusterExtension.Labels == nil {
			clusterExtension.Labels = map[string]string{}
		}

		maps.Copy(clusterExtension.Annotations, desiredClusterExtension.Annotations)
		maps.Copy(clusterExtension.Labels, desiredClusterExtension.Labels)

		clusterExtension.Spec = desiredClusterExtension.Spec
		return nil
	})

	if err != nil {
		logger.Error(err, "failed to create or update ClusterExtension", "ClusterExtension", clusterExtension.Name)
		return err
	}

	logger.Info("ClusterExtension reconciled successfully", "ClusterExtension", clusterExtension.Name, "operation", op)

	return nil
}
