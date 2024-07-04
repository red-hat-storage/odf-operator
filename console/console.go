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

	consolev1 "github.com/openshift/api/console/v1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const MAIN_BASE_PATH = "/"
const COMPATIBILITY_BASE_PATH = "/compatibility/"
const ODF_CONSOLE = "odf-console"
const CUSTOMER_PORTAL_LINK = "https://access.redhat.com/downloads/content/547"

func GetDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ODF_CONSOLE,
			Namespace: namespace,
		},
	}
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
				"app": ODF_CONSOLE,
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
				"app": ODF_CONSOLE,
			},
			Type: "ClusterIP",
		},
	}
}

func GetConsolePluginProxy(serviceNamespace string) []consolev1.ConsolePluginProxy {
	return []consolev1.ConsolePluginProxy{
		{
			Alias: "provider-proxy",
			Endpoint: consolev1.ConsolePluginProxyEndpoint{
				Type: consolev1.ProxyTypeService,
				Service: &consolev1.ConsolePluginProxyServiceConfig{
					Name:      "ux-backend-proxy",
					Namespace: serviceNamespace,
					Port:      8888,
				},
			},
			Authorization: consolev1.UserToken,
		},
		{
			Alias: "rosa-prometheus",
			Endpoint: consolev1.ConsolePluginProxyEndpoint{
				Type: consolev1.ProxyTypeService,
				Service: &consolev1.ConsolePluginProxyServiceConfig{
					Name:      "prometheus",
					Namespace: serviceNamespace,
					Port:      9339,
				},
			},
			Authorization: consolev1.None,
		},
	}
}

func GetConsolePluginCR(consolePort int, serviceNamespace string) *consolev1.ConsolePlugin {
	return &consolev1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: ODF_CONSOLE,
		},
		Spec: consolev1.ConsolePluginSpec{
			DisplayName: "ODF Plugin",
			Backend: consolev1.ConsolePluginBackend{
				Service: &consolev1.ConsolePluginService{
					Name:      "odf-console-service",
					Namespace: serviceNamespace,
					Port:      int32(consolePort),
				},
			},
			I18n: consolev1.ConsolePluginI18n{
				LoadType: consolev1.Empty,
			},
			Proxy: GetConsolePluginProxy(serviceNamespace),
		},
	}
}

func GetBasePath(clusterVersion string) string {
	if strings.Contains(clusterVersion, "4.17") {
		return COMPATIBILITY_BASE_PATH
	}

	return MAIN_BASE_PATH
}

func GetConsoleCLIDownloadLinks() []consolev1.CLIDownloadLink {
	return []consolev1.CLIDownloadLink{
		{
			Href: CUSTOMER_PORTAL_LINK,
			Text: "Red Hat Customer Portal",
		},
	}
}

func GetConsoleCLIDownloadCR() *consolev1.ConsoleCLIDownload {
	return &consolev1.ConsoleCLIDownload{
		ObjectMeta: metav1.ObjectMeta{
			Name: "odf-cli-downloads",
		},
		Spec: consolev1.ConsoleCLIDownloadSpec{
			Description: "With the Data Foundation CLI tool, you can effectively manage and troubleshoot your Data Foundation environment from a terminal.\n\nYou can find a compatible version and download the CLI tool from the Red Hat Customer Portal:\n",
			DisplayName: "odf - Data Foundation Command Line Interface (CLI)",
			Links:       GetConsoleCLIDownloadLinks(),
		},
	}
}
