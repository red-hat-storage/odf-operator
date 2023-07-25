package deploymanager

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNamespace creates a namespace
func (d *DeployManager) CreateNamespace(name string) error {

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"openshift.io/cluster-monitoring":    "true",
				"pod-security.kubernetes.io/enforce": "privileged",
			},
		},
	}

	return d.Client.Create(d.Ctx, namespace)
}

// DeleteNamespaceAndWait deletes a namespace and waits on it to terminate
func (d *DeployManager) DeleteNamespaceAndWait(name string) error {

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"openshift.io/cluster-monitoring": "true",
			},
		},
	}

	err := d.Client.Delete(d.Ctx, namespace)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	timeout := 600 * time.Second
	interval := 10 * time.Second

	err = utilwait.PollUntilContextTimeout(d.Ctx, interval, timeout, true, func(context.Context) (done bool, err error) {

		existingNamespace := &corev1.Namespace{}
		err = d.Client.Get(d.Ctx, client.ObjectKeyFromObject(namespace), existingNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		}
		if err == nil {
			d.Log.Info("Waiting on namespace to be deleted")
			return false, nil
		}
		return true, err
	})

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
