package controllers

import (
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

const (
	random30CharacterString = "KP7TThmSTZegSGmHuPKLnSaaAHSG3RSgqw6akBj0oVk"
)

func getUXBackendServerDeployment() *appsv1.Deployment {

	labels := map[string]string{
		"app.kubernetes.io/component": "ux-backend-server",
		"app.kubernetes.io/name":      "ux-backend-server",
		"app":                         "ux-backend-server",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ux-backend-server",
			Namespace: OperatorNamespace,
			Labels:    labels,
		},
	}
	deploymentSpec := &appsv1.DeploymentSpec{
		Replicas: ptr.To(int32(1)),
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "ux-backend-server",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "onboarding-private-key",
								MountPath: "/etc/private-key",
							},
							{
								Name:      "ux-cert-secret",
								MountPath: "/etc/tls/private",
							},
						},
						Image:           os.Getenv("UX_BACKEND_SERVER_IMAGE"),
						ImagePullPolicy: "IfNotPresent",
						Command:         []string{"/usr/local/bin/ux-backend-server"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8080,
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "ONBOARDING_TOKEN_LIFETIME",
								Value: os.Getenv("ONBOARDING_TOKEN_LIFETIME"),
							},
							{
								Name:  "UX_BACKEND_PORT",
								Value: os.Getenv("UX_BACKEND_PORT"),
							},
							{
								Name:  "TLS_ENABLED",
								Value: os.Getenv("TLS_ENABLED"),
							},
							{
								Name:  "DEVICEFINDER_IMAGE",
								Value: os.Getenv("DEVICEFINDER_IMAGE"),
							},
							{
								Name: "POD_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("50m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("250m"),
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:           ptr.To(true),
							ReadOnlyRootFilesystem: ptr.To(true),
						},
					},
					{
						Name: "oauth-proxy",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "ux-proxy-secret",
								MountPath: "/etc/proxy/secrets",
							},
							{
								Name:      "ux-cert-secret",
								MountPath: "/etc/tls/private",
							},
						},
						Image:           os.Getenv("UX_BACKEND_OAUTH_IMAGE"),
						ImagePullPolicy: "IfNotPresent",
						Args: []string{"-provider=openshift",
							"-https-address=:8888",
							"-http-address=", "-email-domain=*",
							"-upstream=http://localhost:8080/",
							"-tls-cert=/etc/tls/private/tls.crt",
							"-tls-key=/etc/tls/private/tls.key",
							"-cookie-secret-file=/etc/proxy/secrets/session_secret",
							"-openshift-service-account=ux-backend-server",
							`-openshift-delegate-urls={"/":{"group":"ocs.openshift.io","resource":"storageclusters","namespace":"openshift-storage","verb":"create"},"/info/":{"group":"authorization.k8s.io","resource":"selfsubjectaccessreviews","verb":"create"}}`,
							"-openshift-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8888,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("5m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("25m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:           ptr.To(true),
							ReadOnlyRootFilesystem: ptr.To(true),
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "onboarding-private-key",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "onboarding-private-key",
								Optional:   ptr.To(true),
							},
						},
					},
					{
						Name: "ux-proxy-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "ux-backend-proxy",
							},
						},
					},
					{
						Name: "ux-cert-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "ux-cert-secret",
							},
						},
					},
				},
				PriorityClassName:  "system-cluster-critical",
				ServiceAccountName: "ux-backend-server",
			},
		},
	}

	deployment.Spec = *deploymentSpec
	return deployment
}

func getUXBackendServerSecret() *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ux-backend-proxy",
			Namespace: OperatorNamespace,
		},
	}
	secret.StringData = map[string]string{
		"session_secret": random30CharacterString,
	}
	return secret
}

func getUXBackendServerService() *corev1.Service {
	service := &corev1.Service{}
	service.Name = "ux-backend-proxy"
	service.Namespace = OperatorNamespace
	service.Annotations = map[string]string{
		"service.beta.openshift.io/serving-cert-secret-name": "ux-cert-secret",
	}
	service.Spec = corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name:     "proxy",
				Port:     8888,
				Protocol: corev1.ProtocolTCP,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8888,
				},
			},
		},
		Selector:        map[string]string{"app": "ux-backend-server"},
		SessionAffinity: "None",
		Type:            "ClusterIP",
	}
	return service
}
