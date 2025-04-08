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
	"slices"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	odfOperatorConfigMapName = "odf-operator-manager-config"
)

var (
	configMapIgnoreKeys             = []string{"controller_manager_config.yaml"}
	EmptyOdfOperatorConfigMapRecord = OdfOperatorConfigMapRecord{}
)

type OdfOperatorConfigMapRecord struct {
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

func GetOdfConfigMap(ctx context.Context, cli client.Client, logger logr.Logger) (corev1.ConfigMap, error) {

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

func ParseOdfConfigMapRecords(logger logr.Logger, configmap corev1.ConfigMap, fn func(*OdfOperatorConfigMapRecord, string, string)) {

	var record OdfOperatorConfigMapRecord

	for key, value := range configmap.Data {

		// skip parsing known environment variable and keys from the configmap.
		if slices.Contains(configMapIgnoreKeys, key) {
			continue
		}

		record = EmptyOdfOperatorConfigMapRecord
		if err := yaml.Unmarshal([]byte(value), &record); err != nil {
			logger.Error(err, "failed to unmarshal configmap data", "key", key)
			continue
		}

		fn(&record, key, value)
	}
}
