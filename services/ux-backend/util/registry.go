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
		return nil, fmt.Errorf("failed to unmarshal docker config: %w", err)
	}
	return config, nil
}

func getRegistryCredentials(ctx context.Context, secretName string, secretNamespace string, client client.Client) (*types.DockerAuthConfig, error) {
	secret := &corev1.Secret{}
	if err := client.Get(ctx, k8stypes.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret); err != nil {
		klog.Errorf("failed to get secret: %v", err)
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	config, err := parseDockerRegistrySecret(secret)
	if err != nil {
		klog.Errorf("failed to parse docker registry secret: %v", err)
		return nil, fmt.Errorf("failed to parse docker registry secret: %w", err)
	}
	return config, nil
}

func TestRegistryConnection(ctx context.Context, registryURL string, registryRepositoryName string, secretKey string, secretNamespace string, client client.Client) error {
	credentials, err := getRegistryCredentials(ctx, secretKey, secretNamespace, client)
	if err != nil {
		klog.Errorf("failed to get registry credentials: %v", err)
		return fmt.Errorf("failed to get registry credentials: %w", err)
	}

	// Validate registry authentication.
	if err := docker.CheckAuth(ctx, &types.SystemContext{}, credentials.Username, credentials.Password, registryURL); err != nil {
		klog.Errorf("failed to authenticate to registry %s: %v", registryURL, err)
		return fmt.Errorf("failed to authenticate to registry: %w", err)
	}

	// Validate repository access. An empty tag list is considered success.
	ref, err := docker.ParseReference(fmt.Sprintf("//%s/%s", registryURL, registryRepositoryName))
	if err != nil {
		klog.Errorf("failed to parse repository reference %s/%s: %v", registryURL, registryRepositoryName, err)
		return fmt.Errorf("invalid registry URL or repository name: %w", err)
	}

	sys := &types.SystemContext{
		DockerAuthConfig: credentials,
	}
	if _, err := docker.GetRepositoryTags(ctx, sys, ref); err != nil {
		klog.Errorf("failed to access repository %s/%s: %v", registryURL, registryRepositoryName, err)
		return fmt.Errorf("failed to access repository: %w", err)
	}

	return nil
}
