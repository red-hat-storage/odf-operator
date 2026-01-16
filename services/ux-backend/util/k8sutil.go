package util

import (
	"context"
	"fmt"
	"os"

	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// PodNamespaceEnvVar is the env variable for the pod namespace
	PodNamespaceEnvVar = "POD_NAMESPACE"
)

var podNamespace = os.Getenv(PodNamespaceEnvVar)

// GetPodNamespace returns the namespace where the pod is deployed
func GetPodNamespace() string {
	if podNamespace == "" {
		panic(fmt.Errorf("%s must be set", PodNamespaceEnvVar))
	}
	return podNamespace
}

func GetStorageClusterInNamespace(ctx context.Context, cl client.Client, namespace string) (*ocsv1.StorageCluster, error) {
	storageClusterList := &ocsv1.StorageClusterList{}
	err := cl.List(ctx, storageClusterList, client.InNamespace(namespace), client.Limit(1))
	if err != nil {
		return nil, fmt.Errorf("unable to list storageCluster(s) in namespace %s: %v", namespace, err)
	}

	if len(storageClusterList.Items) == 0 {
		return nil, fmt.Errorf("no storageCluster found in namespace %s", namespace)
	}
	if storageClusterList.Items[0].Status.Phase == PhaseIgnored {
		return nil, fmt.Errorf("storageCluster with Phase 'Ignored' found. Please delete the storageCluster to proceed")
	}

	return &storageClusterList.Items[0], nil
}
