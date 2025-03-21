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
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	odfOperatorConfigmapName = "odf-operator-manager-config"
)

type ConfigData struct {
	/* example
	   channel: alpha
	   csv: ocs-operator.v4.18.0
	   pkg: ocs-operator
	   scalerCrds:
	     - storageclusters.ocs.openshift.io
	   ---------------------------------------
	   channel: beta
	   csv: odf-prometheus-operator.v4.18.0
	   pkg: odf-prometheus-operator
	   scalerCrds:
	     - alertmanagers.monitoring.coreos.com
	     - prometheuses.monitoring.coreos.com
	*/

	Channel    string   `yaml:"channel"`
	Csv        string   `yaml:"csv"`
	Pkg        string   `yaml:"pkg"`
	ScalerCrds []string `yaml:"scalerCrds"`
}

func GetConfig(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string) (map[string]ConfigData, error) {

	// read configmap
	cm := &corev1.ConfigMap{}
	cm.Name = odfOperatorConfigmapName
	cm.Namespace = operatorNamespace
	if err := cli.Get(ctx, client.ObjectKeyFromObject(cm), cm); err != nil {
		logger.Error(err, "failed to get configmap")
		return nil, err
	}

	// parse the ConfigMap data and skip any keys that fail to parse
	configMapData := make(map[string]ConfigData)
	for key, value := range cm.Data {

		// skip parsing known environment variable keys from the configmap.
		// first condition can be removed once the older keys are removed from the configmap.
		if _, ok := DefaultValMap[key]; ok || key == "controller_manager_config.yaml" {
			continue
		}

		var config ConfigData
		if err := yaml.Unmarshal([]byte(value), &config); err != nil {
			logger.Error(err, "failed to unmarshal configmap data", "key", key)
			continue
		}
		configMapData[key] = config
	}

	logger.Info("parsed configmap data successfully", "configMapData", configMapData)

	return configMapData, nil
}

type CrdAndPackages struct {
	/* examples
	   CrdName:    "storageclusters.ocs.openshift.io",
	   ApiVersion: "ocs.openshift.io/v1",
	   Kind:       "StorageCluster",
	   PkgNames:   []string{OcsSubscriptionPackage},
	   -----------------------------------------------
	   CrdName:    "cephclusters.ceph.rook.io",
	   ApiVersion: "ceph.rook.io/v1",
	   Kind:       "CephCluster",
	   PkgNames:   []string{RookSubscriptionPackage, CephCSISubscriptionPackage, CSIAddonsSubscriptionPackage},
	*/

	CrdName    string
	ApiVersion string
	Kind       string
	CrdPresent bool
	PkgNames   []string
}

func GetCrdAndPackagesMappingList(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string) ([]CrdAndPackages, error) {

	configMapData, err := GetConfig(ctx, cli, logger, operatorNamespace)
	if err != nil {
		logger.Error(err, "failed to get configmap data")
		return nil, err
	}

	// build a mapping of CRD names to the corresponding package names
	// this allows us to associate each CRD with the multiple packages
	crdToPkgNames := map[string][]string{}
	for _, value := range configMapData {
		for _, crd := range value.ScalerCrds {
			crdToPkgNames[crd] = append(crdToPkgNames[crd], value.Pkg)
		}
	}

	// create a slice of ResourceMappingRecord
	resourceMapping := []CrdAndPackages{}
	for crd, pkgNames := range crdToPkgNames {
		resourceMapping = append(resourceMapping, CrdAndPackages{
			CrdName:  crd,
			PkgNames: pkgNames,
		})
	}

	// get the CRD details for each CRD (API version and Kind)
	for i, record := range resourceMapping {
		crd := &extv1.CustomResourceDefinition{}
		if err := cli.Get(ctx, client.ObjectKey{Name: record.CrdName}, crd); err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "CRD not found", "crdName", record.CrdName)
				continue
			}
			logger.Error(err, "failed to get CRD", "crdName", record.CrdName)
			return nil, err
		}
		resourceMapping[i].CrdPresent = true
		resourceMapping[i].ApiVersion = crd.Spec.Group + "/" + crd.Spec.Versions[0].Name
		resourceMapping[i].Kind = crd.Spec.Names.Kind
	}

	logger.Info("parsed resource mapping successfully", "resourceMapping", resourceMapping)

	return resourceMapping, nil
}

func GetCrdNamesMap(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string) (map[string]bool, error) {

	configMapData, err := GetConfig(ctx, cli, logger, operatorNamespace)
	if err != nil {
		logger.Error(err, "failed to get configmap data")
		return nil, err
	}

	crdNames := map[string]bool{}
	for _, value := range configMapData {
		for _, crd := range value.ScalerCrds {
			crdNames[crd] = true
		}
	}

	logger.Info("parsed CRD names successfully", "crdNames", crdNames)

	return crdNames, nil
}

func GetCsvNamesMap(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string) (map[string]bool, error) {

	configMapData, err := GetConfig(ctx, cli, logger, operatorNamespace)
	if err != nil {
		logger.Error(err, "failed to get configmap data")
		return nil, err
	}

	csvNames := map[string]bool{}
	for _, value := range configMapData {
		csvNames[value.Csv] = true
	}

	logger.Info("parsed csv names successfully", "csvNames", csvNames)

	return csvNames, nil
}

func GetOdfDependenciesConfig(ctx context.Context, cli client.Client, logger logr.Logger, operatorNamespace string) (ConfigData, error) {

	// read configmap
	cm := &corev1.ConfigMap{}
	cm.Name = odfOperatorConfigmapName
	cm.Namespace = operatorNamespace
	if err := cli.Get(ctx, client.ObjectKeyFromObject(cm), cm); err != nil {
		logger.Error(err, "failed to get configmap")
		return ConfigData{}, err
	}

	// parse the ConfigMap data and skip any keys that fail to parse
	for key, value := range cm.Data {

		// skip parsing known environment variable keys from the ConfigMap.
		// first condition can be removed once the older keys are removed from the ConfigMap.
		if _, ok := DefaultValMap[key]; ok || key == "controller_manager_config.yaml" {
			continue
		}

		var config ConfigData
		if err := yaml.Unmarshal([]byte(value), &config); err != nil {
			logger.Error(err, "failed to unmarshal configmap data", "key", key)
			continue
		}

		if config.Pkg == "odf-dependencies" {
			logger.Info("parsed odf-dependencies config successfully", "config", config)
			return config, nil
		}
	}

	err := fmt.Errorf("odf-dependencies config not found in configmap")
	logger.Error(err, "failed to get odf-dependencies config")
	return ConfigData{}, err
}
