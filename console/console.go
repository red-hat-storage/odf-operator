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

package console

import (
	"strings"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const MAIN_BASE_PATH = "/"
const COMPATIBILITY_BASE_PATH = "/compatibility/"

func GetDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "odf-console",
			Namespace: namespace,
		},
	}
}

func GetNginxConfiguration() string {
	return NginxConf
}

func GetNginxConfConfigMap(namespace string) *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "odf-console-nginx-conf",
			Namespace: namespace,
		},
		Data: map[string]string{
			"nginx.conf": NginxConf,
		},
	}
}

func GetService(port int, namespace string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "odf-console-service",
			Namespace: namespace,
			Annotations: map[string]string{
				"service.alpha.openshift.io/serving-cert-secret-name": "odf-console-serving-cert",
			},
			Labels: map[string]string{
				"app": "odf-console",
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Protocol: "TCP",
					TargetPort: intstr.IntOrString{IntVal: int32(port)},
					Port:       int32(port),
					Name:       "console-port",
				},
			},
			Selector: map[string]string{
				"app": "odf-console",
			},
			Type: "ClusterIP",
		},
	}
}

func GetConsolePluginCR(consolePort int, serviceNamespace string) *consolev1alpha1.ConsolePlugin {
	return &consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: "odf-console",
		},
		Spec: consolev1alpha1.ConsolePluginSpec{
			DisplayName: "ODF Plugin",
			Service: consolev1alpha1.ConsolePluginService{
				Name:      "odf-console-service",
				Namespace: serviceNamespace,
				Port:      int32(consolePort),
			},
		},
	}
}

func GetBasePath(clusterVersion string) string {
	if strings.Contains(clusterVersion, "4.16") {
		return COMPATIBILITY_BASE_PATH
	}

	return MAIN_BASE_PATH
}
