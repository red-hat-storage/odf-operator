package console

import (
	"context"

	consolev1 "github.com/openshift/api/console/v1"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const DEPLOYMENT_NAME = "odf-console"
const DEPLOYMENT_NAMESPACE = "openshift-storage"
const ODF_CONSOLE_IMAGE = "docker.io/badhikar/odf-console:4.9"

func int32Ptr(i int32) *int32 { return &i }

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(consolev1alpha1.Install(scheme))
	//+kubebuilder:scaffold:scheme
}

var (
	serverDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DEPLOYMENT_NAME,
			Namespace: DEPLOYMENT_NAMESPACE,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "ui",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "ui",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "odf-console",
							Image: ODF_CONSOLE_IMAGE,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 9001,
								},
							},
						},
					},
				},
			},
		},
	}
	consolePlugin = &consolev1alpha1.ConsolePlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: "odf-console",
		},
		Spec: consolev1alpha1.ConsolePluginSpec{
			DisplayName: "ODF Console",
			Service: consolev1alpha1.ConsolePluginService{
				Name:      "odf-service",
				Namespace: "openshift-storage",
				Port:      int32(9001),
			},
		},
	}
	service = &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "odf-service",
			Namespace: DEPLOYMENT_NAMESPACE,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Name: "jack", Port: int32(9001)},
			},
		},
	}
)

func Runner() {
	cl, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		klog.Error(err)
	}
	err = cl.Create(context.TODO(), serverDeployment)
	if err != nil {
		klog.Error(err)
	}
	err = cl.Create(context.TODO(), service)
	if err != nil {
		klog.Error(err)
	}
	err = cl.Create(context.TODO(), consolePlugin)
	if err != nil {
		klog.Error(err)
	}

}
