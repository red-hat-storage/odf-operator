package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// dockerConfigJSON mirrors the structure the containers/image library
// uses internally (pkg/docker/config.dockerConfigFile) but is unexported there.
type dockerConfigJSON struct {
	Auths map[string]dockerConfigEntry `json:"auths"`
}

// dockerConfigEntry mirrors pkg/docker/config.dockerAuthConfig.
type dockerConfigEntry struct {
	Auth string `json:"auth,omitempty"`
}

// decodeDockerConfigAuth decodes a base64-encoded "user:password" auth string,
// mirroring the logic in pkg/docker/config.decodeDockerAuth.
func decodeDockerConfigAuth(encoded string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode auth field: %w", err)
	}
	user, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return "", "", fmt.Errorf("invalid auth field: missing ':' separator")
	}
	return user, strings.Trim(password, "\x00"), nil
}

func parseDockerRegistrySecret(secret *corev1.Secret) (*types.DockerAuthConfig, error) {
	data, ok := secret.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return nil, fmt.Errorf("docker config json key not found in secret")
	}
	var config dockerConfigJSON
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal docker config: %w", err)
	}
	if len(config.Auths) == 0 {
		return nil, fmt.Errorf("no auth entries found in docker config")
	}
	// Use the first (and typically only) auth entry.
	for registry, entry := range config.Auths {
		if entry.Auth == "" {
			return nil, fmt.Errorf("empty auth field for registry %s", registry)
		}
		username, password, err := decodeDockerConfigAuth(entry.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to decode credentials for registry %s: %w", registry, err)
		}
		return &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}, nil
	}
	return nil, fmt.Errorf("no valid auth entries found in docker config")
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
