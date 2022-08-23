package controllers

import (
	"context"
	"fmt"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	"github.com/go-logr/logr"
	odfv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1alpha1 "open-cluster-management.io/api/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ClusterClaimKey string

const (
	OdfVersion            ClusterClaimKey = "odf-version"
	StorageSystemName     ClusterClaimKey = "odf-storage-system-name"
	StorageClusterName    ClusterClaimKey = "odf-storage-cluster-name"
	CephFsid              ClusterClaimKey = "odf-ceph-fsid"
	RookCephMonSecretName string          = "rook-ceph-mon"
	FsidKey               string          = "fsid"
)

type ClusterClaimCreator struct {
	Context       context.Context
	Logger        logr.Logger
	Client        client.Client
	Values        map[ClusterClaimKey]string
	StorageSystem *odfv1alpha1.StorageSystem
}

func (r *StorageSystemReconciler) ensureClusterClaims(instance *odfv1alpha1.StorageSystem, logger logr.Logger) error {
	ctx := context.TODO()
	crd := extensionsv1.CustomResourceDefinition{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: "clusterclaims.cluster.open-cluster-management.io"}, &crd)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("clusterclaims.cluster.open-cluster-management.io CRD not found. skipping creation of clusterclaims")
			return nil
		}
		return err
	}

	creator := ClusterClaimCreator{
		Logger:        logger,
		Context:       ctx,
		Client:        r.Client,
		Values:        make(map[ClusterClaimKey]string),
		StorageSystem: instance,
	}

	odfVersion, err := creator.getOdfVersion()
	if len(odfVersion) == 0 || err != nil {
		logger.Info("failed to get odf version for operator. retrying again")
		return err
	}

	cephFsid, err := creator.getCephFsid()
	if len(cephFsid) == 0 || err != nil {
		logger.Info("failed to get ceph fsid from secret. retrying again")
		return err
	}

	err = creator.setStorageSystemName(instance.Name).
		setStorageClusterName(instance.Spec.Name).
		setOdfVersion(odfVersion).
		setCephFsid(cephFsid).
		create()

	return err
}

func (c *ClusterClaimCreator) create() error {
	for key, value := range c.Values {
		cc := clusterv1alpha1.ClusterClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: string(key),
				OwnerReferences: []metav1.OwnerReference{
					metav1.OwnerReference{
						APIVersion: c.StorageSystem.APIVersion,
						Kind:       c.StorageSystem.Kind,
						Name:       c.StorageSystem.Name,
						UID:        c.StorageSystem.UID,
					},
				},
			},
			Spec: clusterv1alpha1.ClusterClaimSpec{
				Value: value,
			},
		}

		_, err := controllerutil.CreateOrUpdate(c.Context, c.Client, &cc, func() error {
			return nil
		})

		if err != nil {
			c.Logger.Info("failed to create/update clusterclaim for %q", key)
			return err
		}
	}

	return nil
}
func (c *ClusterClaimCreator) getOdfVersion() (string, error) {
	var csvs operatorsv1alpha1.ClusterServiceVersionList
	err := c.Client.List(c.Context, &csvs, &client.ListOptions{Namespace: c.StorageSystem.Namespace})
	if err != nil {
		return "", err
	}

	for _, csv := range csvs.Items {
		if strings.HasPrefix(csv.Name, OperatorNamePrefix) {
			return csv.Name, nil
		}
	}

	return "", fmt.Errorf("failed to find csv with prefix %q", OperatorNamePrefix)
}

func (c *ClusterClaimCreator) getCephFsid() (string, error) {
	var rookCephMonSecret corev1.Secret
	err := c.Client.Get(c.Context, types.NamespacedName{Name: RookCephMonSecretName, Namespace: c.StorageSystem.Namespace}, &rookCephMonSecret)
	if err != nil {
		return "", err
	}
	if val, ok := rookCephMonSecret.Data[FsidKey]; ok {
		return string(val), nil
	}

	return "", fmt.Errorf("failed to fetch ceph fsid from %q secret", RookCephMonSecretName)
}

func (c *ClusterClaimCreator) setStorageSystemName(name string) *ClusterClaimCreator {
	c.Values[StorageSystemName] = name
	return c
}

func (c *ClusterClaimCreator) setOdfVersion(version string) *ClusterClaimCreator {
	c.Values[OdfVersion] = version
	return c
}

func (c *ClusterClaimCreator) setStorageClusterName(name string) *ClusterClaimCreator {
	c.Values[StorageClusterName] = name
	return c
}

func (c *ClusterClaimCreator) setCephFsid(fsid string) *ClusterClaimCreator {
	c.Values[CephFsid] = fsid
	return c
}
