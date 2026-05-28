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

package config

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var (
	EmptyPkgConfigMapRecord = PkgConfigMapRecord{}
)

type PkgConfigMapRecord struct {
	/* example
	   package: ocs-operator
	   version: v5.1.0
	   namespace: openshift-storage
	   ---------------------------------------
	   package: odf-prometheus-operator
	   version: v5.1.0
	   namespace: "" (empty will be treated as operator namespace)
	*/

	Package   string `yaml:"package"`
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"`
}

func GetConfigMap(ctx context.Context, cli client.Client, logger logr.Logger, configMapName string, namespace string) (corev1.ConfigMap, error) {

	cm := corev1.ConfigMap{}
	cm.Name = configMapName
	cm.Namespace = namespace

	if err := cli.Get(ctx, client.ObjectKeyFromObject(&cm), &cm); err != nil {
		logger.Error(err, "failed to get configmap", "configmap", cm.Name)
		return corev1.ConfigMap{}, err
	}

	logger.Info("found configmap successfully", "configmap", cm.Name)
	return cm, nil
}

func ParsePkgsConfigMapRecords(logger logr.Logger, configmap corev1.ConfigMap, operatorNamespace string, fn func(*PkgConfigMapRecord, string, string)) {

	var record PkgConfigMapRecord

	for key, value := range configmap.Data {

		record = EmptyPkgConfigMapRecord
		if err := yaml.Unmarshal([]byte(value), &record); err != nil {
			logger.Error(err, "failed to unmarshal configmap data", "key", key)
			continue
		}

		if record.Namespace == "" {
			record.Namespace = operatorNamespace
		}

		fn(&record, key, value)
	}
}
