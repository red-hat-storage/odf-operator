package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func parseDockerRegistrySecret(secret *corev1.Secret) (*types.DockerAuthConfig, error) {
	data, ok := secret.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return nil, fmt.Errorf("docker config json key not found in secret")
	}
	var config *types.DockerAuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal docker config: %v", err)
	}
	return config, nil
}

func getRegistryCredentials(ctx context.Context, secretName string, secretNamespace string, client client.Client) (*types.DockerAuthConfig, error) {
	secret := &corev1.Secret{}
	if err := client.Get(ctx, k8stypes.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret); err != nil {
		klog.Errorf("failed to get secret: %v", err)
		return nil, fmt.Errorf("failed to get secret: %v", err)
	}
	config, err := parseDockerRegistrySecret(secret)
	if err != nil {
		klog.Errorf("failed to parse docker registry secret: %v", err)
		return nil, fmt.Errorf("failed to parse docker registry secret: %v", err)
	}
	return config, nil
}

func TestRegistryConnection(ctx context.Context, registryURL string, registryRepositoryName string, secretKey string, secretNamespace string, client client.Client) error {
	ref, err := docker.ParseReference(fmt.Sprintf("//%s/%s", registryURL, registryRepositoryName))
	if err != nil {
		klog.Errorf("failed to parse reference: %v", err)
		return fmt.Errorf("registry URL, repository name or tag is invalid: %v", err)
	}
	credentials, err := getRegistryCredentials(ctx, secretKey, secretNamespace, client)
	if err != nil {
		klog.Errorf("failed to get registry credentials: %v", err)
		return fmt.Errorf("failed to get registry credentials: %v", err)
	}
	sys := &types.SystemContext{
		DockerAuthConfig: credentials,
	}
	image, err := ref.NewImageSource(ctx, sys)
	if err != nil {
		klog.Errorf("failed to get image: %v", err)
		return fmt.Errorf("failed to get image: %v", err)
	}
	defer func() {
		if err := image.Close(); err != nil {
			klog.Errorf("failed to close image source: %v", err)
		}
	}()
	_, _, err = image.GetManifest(ctx, nil)
	if err != nil {
		klog.Errorf("failed to get manifest: %v", err)
		return fmt.Errorf("failed to get manifest: %v", err)
	}
	return nil
}
