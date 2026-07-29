package discovery

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("updateConfigMap", func() {
	const (
		namespace  = "openshift-storage"
		nodeName   = "worker-1"
		oldPodName = "devicefinder-abc12"
		oldPodUID  = "11111111-1111-1111-1111-111111111111"
		newPodName = "devicefinder-def34"
		newPodUID  = "22222222-2222-2222-2222-222222222222"
	)

	var discovery *DeviceDiscovery

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		staleCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapNamePrefix + nodeName,
				Namespace: namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       oldPodName,
						UID:        types.UID(oldPodUID),
					},
				},
			},
			Data: map[string]string{"discovered-devices": "[]"},
		}

		discovery = &DeviceDiscovery{
			kubeClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(staleCM).Build(),
		}

		t := GinkgoT()
		t.Setenv("NODE_NAME", nodeName)
		t.Setenv("POD_NAMESPACE", namespace)
		t.Setenv("POD_NAME", newPodName)
		t.Setenv("POD_UID", newPodUID)
	})

	It("reclaims ownership from a stale pod owner reference on update", func() {
		Expect(discovery.updateConfigMap()).To(Succeed())

		cm := &corev1.ConfigMap{}
		Expect(discovery.kubeClient.Get(context.TODO(), client.ObjectKey{
			Name:      configMapNamePrefix + nodeName,
			Namespace: namespace,
		}, cm)).To(Succeed())

		Expect(cm.OwnerReferences).To(HaveLen(1))
		ownerRef := cm.OwnerReferences[0]
		Expect(ownerRef.Kind).To(Equal("Pod"))
		Expect(ownerRef.Name).To(Equal(newPodName))
		Expect(ownerRef.UID).To(Equal(types.UID(newPodUID)))
		Expect(ownerRef.Controller).To(BeNil())
	})

	It("preserves user-added labels when updating the ConfigMap", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		existingCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapNamePrefix + nodeName,
				Namespace: namespace,
				Labels: map[string]string{
					"custom-label": "user-value",
				},
			},
			Data: map[string]string{"discovered-devices": "[]"},
		}

		discovery = &DeviceDiscovery{
			kubeClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingCM).Build(),
		}

		Expect(discovery.updateConfigMap()).To(Succeed())

		cm := &corev1.ConfigMap{}
		Expect(discovery.kubeClient.Get(context.TODO(), client.ObjectKey{
			Name:      configMapNamePrefix + nodeName,
			Namespace: namespace,
		}, cm)).To(Succeed())

		Expect(cm.Labels).To(HaveKeyWithValue("custom-label", "user-value"))
		Expect(cm.Labels).To(HaveKeyWithValue("app", "devicefinder"))
		Expect(cm.Labels).To(HaveKeyWithValue("node", nodeName))
	})
})
