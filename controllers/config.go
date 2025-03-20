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
	odfOperatorConfigMapName = "odf-operator-manager-config"
)

var (
	// configMapResorceVersion will keep the local copy of the odf-operator-manager-config configmap resource version
	configMapResorceVersion string
	// OlmPkgRecordList will keep the local copy of parsed data from odf-operator-manager-config configmap
	OlmPkgRecordList = []*OlmPkgRecord{}
	// CrdToPackageRecordList will keep the local copy of parsed data from OlmPkgRecordList
	CrdToPackageRecordList = []*CrdToPackageRecord{}
)

type OlmPkgRecord struct {
	/* example
	   channel: alpha
	   csv: ocs-operator.v4.18.0
	   pkg: ocs-operator
	   scaleUpOnInstanceOf:
	     - storageclusters.ocs.openshift.io
	   ---------------------------------------
	   channel: beta
	   csv: odf-prometheus-operator.v4.18.0
	   pkg: odf-prometheus-operator
	   scaleUpOnInstanceOf:
	     - alertmanagers.monitoring.coreos.com
	     - prometheuses.monitoring.coreos.com
	*/

	Channel             string   `yaml:"channel"`
	Csv                 string   `yaml:"csv"`
	Pkg                 string   `yaml:"pkg"`
	ScaleUpOnInstanceOf []string `yaml:"ScaleUpOnInstanceOf"`
}

type CrdToPackageRecord struct {
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
	PkgNames   []string
}

func GetOdfOperatorConfigMap(ctx context.Context, cli client.Client, logger logr.Logger) (corev1.ConfigMap, error) {

	cm := corev1.ConfigMap{}
	cm.Name = odfOperatorConfigMapName
	cm.Namespace = OperatorNamespace

	if err := cli.Get(ctx, client.ObjectKeyFromObject(&cm), &cm); err != nil {
		logger.Error(err, "failed to get configmap", "configmap", cm.Name)
		return corev1.ConfigMap{}, err
	}

	logger.Info("found configmap successfully", "configmap", cm.Name)
	return cm, nil
}

func ParseOlmPkgRecords(ctx context.Context, cli client.Client, logger logr.Logger, configmap corev1.ConfigMap) error {

	if configmap.ResourceVersion == configMapResorceVersion {
		return nil
	}

	// parse the ConfigMap data and skip any keys that fail to parse
	var olmPkgRecordList []*OlmPkgRecord
	for key, value := range configmap.Data {

		// skip parsing known environment variable keys from the configmap.
		// first condition can be removed once the older keys are removed from the configmap.
		if _, ok := DefaultValMap[key]; ok || key == "controller_manager_config.yaml" {
			continue
		}

		var record OlmPkgRecord
		if err := yaml.Unmarshal([]byte(value), &record); err != nil {
			logger.Error(err, "failed to unmarshal configmap data", "key", key)
		} else if record.Channel == "" || record.Csv == "" || record.Pkg == "" {
			logger.Error(fmt.Errorf("missing required fields in configmap data"), "failed to parse configmap data", "key", key)
		} else {
			olmPkgRecordList = append(olmPkgRecordList, &record)
		}
	}

	logger.Info("parsed configmap data successfully", "configMapData", olmPkgRecordList)

	// update the global copy of the parsed data
	OlmPkgRecordList = olmPkgRecordList

	return nil
}

func ParseCrdToPackageRecords(ctx context.Context, cli client.Client, logger logr.Logger) error {

	crdMapping := map[string]*CrdToPackageRecord{}

	for _, pkgInfo := range OlmPkgRecordList {
		for _, crdName := range pkgInfo.ScaleUpOnInstanceOf {

			if _, ok := crdMapping[crdName]; !ok {

				record := &CrdToPackageRecord{}
				record.CrdName = crdName
				record.PkgNames = []string{pkgInfo.Pkg}

				crd := &extv1.CustomResourceDefinition{}
				if err := cli.Get(ctx, client.ObjectKey{Name: crdName}, crd); errors.IsNotFound(err) {
					logger.Info("CRD not found, populating details without API version and kind", "crdName", crdName)
				} else if err != nil {
					logger.Error(err, "failed to get CRD", "crdName", crdName)
					return err
				} else {
					record.ApiVersion = crd.Spec.Group + "/" + crd.Spec.Versions[0].Name
					record.Kind = crd.Spec.Names.Kind
				}
				crdMapping[crdName] = record
			} else {
				crdMapping[crdName].PkgNames = append(crdMapping[crdName].PkgNames, pkgInfo.Pkg)
			}
		}
	}

	logger.Info("parsed resource mapping successfully", "resourceMapping", crdMapping)

	// update the global copy of the parsed data
	CrdToPackageRecordList = []*CrdToPackageRecord{}
	for _, value := range crdMapping {
		CrdToPackageRecordList = append(CrdToPackageRecordList, value)
	}

	return nil
}

func ParseRecords(ctx context.Context, cli client.Client, logger logr.Logger) error {

	configmap, err := GetOdfOperatorConfigMap(ctx, cli, logger)
	if err != nil {
		logger.Error(err, "failed to get configmap")
		return err
	}

	if configmap.ResourceVersion != configMapResorceVersion {
		if err := ParseOlmPkgRecords(ctx, cli, logger, configmap); err != nil {
			logger.Error(err, "failed to parse olm pkg records")
			return err
		}
		// update the global copy of the configmap resource version
		configMapResorceVersion = configmap.ResourceVersion
		return nil
	}

	if err := ParseCrdToPackageRecords(ctx, cli, logger); err != nil {
		logger.Error(err, "failed to parse crd to package records")
		return err
	}

	return nil
}

func GetCrdNamesMap() map[string]bool {

	crdNames := map[string]bool{}
	for _, record := range OlmPkgRecordList {
		for _, crdName := range record.ScaleUpOnInstanceOf {
			crdNames[crdName] = true
		}
	}

	return crdNames
}

func GetCsvNamesMap() map[string]bool {

	csvNames := map[string]bool{}
	for _, record := range OlmPkgRecordList {
		csvNames[record.Csv] = true
	}

	return csvNames
}

func GetOdfDependenciesSubConfig(logger logr.Logger) (OlmPkgRecord, error) {

	for _, record := range OlmPkgRecordList {
		if record.Pkg == "odf-dependencies" {
			logger.Info("parsed odf-dependencies config successfully", "config", record)
			return OlmPkgRecord{Channel: record.Channel, Csv: record.Csv, Pkg: record.Pkg}, nil
		}
	}

	err := fmt.Errorf("odf-dependencies config not found in configmap")
	logger.Error(err, "failed to get odf-dependencies config")
	return OlmPkgRecord{}, err
}
