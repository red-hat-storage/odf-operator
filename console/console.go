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
	"context"

	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DEPLOYMENT_NAMESPACE = "openshift-storage"

func GetService(serviceName string, port int, owner metav1.ObjectMeta) apiv1.Service {
	return apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName + "-service",
			Namespace: DEPLOYMENT_NAMESPACE,
			Annotations: map[string]string{
				"service.alpha.openshift.io/serving-cert-secret-name": serviceName + "-serving-cert",
			},
			Labels: map[string]string{
				"app": "odf-console",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "odf-console",
					UID:        owner.UID,
				},
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

func GetConsolePluginCR(pluginName string, displayName string, consolePort int, serviceName string, owner metav1.ObjectMeta) consolev1alpha1.ConsolePlugin {
	return consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: pluginName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "odf-console",
					UID:        owner.UID,
				},
			},
		},
		Spec: consolev1alpha1.ConsolePluginSpec{
			DisplayName: displayName,
			Service: consolev1alpha1.ConsolePluginService{
				Name:      serviceName,
				Namespace: DEPLOYMENT_NAMESPACE,
				Port:      int32(consolePort),
			},
		},
	}
}

//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=console.openshift.io,resources=consoleplugins,verbs=*

func InitConsole(client client.Client, odfPort int) error {
	deployment := appsv1.Deployment{}
	if err := client.Get(context.TODO(), types.NamespacedName{
		Name:      "odf-console",
		Namespace: DEPLOYMENT_NAMESPACE,
	}, &deployment); err != nil {
		return err
	}
	// Create core ODF console Service
	odfService := GetService("odf-console", odfPort, deployment.ObjectMeta)
	if err := client.Create(context.TODO(), &odfService); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	// Create core ODF Plugin
	odfConsolePlugin := GetConsolePluginCR("odf-console", "ODF Plugin", odfPort, odfService.ObjectMeta.Name, deployment.ObjectMeta)
	if err := client.Create(context.TODO(), &odfConsolePlugin); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
