package controllers

import (
	"context"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	namespacelabelv1 "tutorial.kubebuilder.io/project/api/v1"
)

var (
	defaultLabels = map[string]string{"bob": "5"}
)
var _ = Describe("NamespaceLabel controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		NamespaceLabelName = "test-namespacelabel"
		Namespace          = "default"
		NameLabel          = "app.kubernetes.io/name"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a NamespaceLabel", func() {
		It("Should update the Namespace's labels", func() {
			By("By creating a new NamespaceLabel")
			ctx := context.Background()
			namespaceLabel := &namespacelabelv1.NamespaceLabel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "core.tutorial.kubebuilder.io/v1",
					Kind:       "NamespaceLabel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      NamespaceLabelName,
					Namespace: Namespace,
				},
				Spec: namespacelabelv1.NamespaceLabelSpec{
					Labels: defaultLabels,
				},
			}
			Expect(k8sClient.Create(ctx, namespaceLabel)).Should(Succeed())
			namespacelabelLookupKey := types.NamespacedName{Name: NamespaceLabelName, Namespace: Namespace}
			createdNamespaceLabel := &namespacelabelv1.NamespaceLabel{}

			// We'll need to retry getting this newly created NamespaceLabel, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespacelabelLookupKey, createdNamespaceLabel)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// Let's make sure our Labels value was properly converted/handled.
			Expect(createdNamespaceLabel.Spec.Labels).Should(Equal(map[string]string{"bob": "5"}))

			By("Checking the Namespace's labels were updated")

			namespaceLookupKey := types.NamespacedName{Name: Namespace}
			createdNamespace := &corev1.Namespace{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespaceLookupKey, createdNamespace)
				if err != nil {
					return false
				}
				return reflect.DeepEqual(createdNamespace.Labels, defaultLabels)
			}, timeout, interval).Should(BeTrue())
		})
	})

})
