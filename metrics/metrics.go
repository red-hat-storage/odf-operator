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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	storageSystemMap = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "odf",
		Subsystem: "",
		Name:      "system_map",
		Help:      "Map of ODF StorageSystems to their target Custom Resource",
	}, []string{"storage_system", "target_name", "target_namespace", "target_kind", "target_apiversion"})
)

func init() {
	metrics.Registry.MustRegister(storageSystemMap)
}

func ReportODFSystemMapMetrics(storageSystem, name, namespace, kind, apiVersion string) {
	storageSystemMap.With(prometheus.Labels{
		"storage_system":    storageSystem,
		"target_name":       name,
		"target_namespace":  namespace,
		"target_kind":       kind,
		"target_apiversion": apiVersion,
	}).Set(1)
}
