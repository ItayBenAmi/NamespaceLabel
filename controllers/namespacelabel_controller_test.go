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
		NamespaceLabelName       = "test-namespacelabel"
		SecondNamespaceLabelName = "test-namespacelabel2"
		Namespace                = "default"
		NameLabel                = "app.kubernetes.io/name"
		ProtectedLabelKey        = "app.kubernetes.io/instance"
		ProtectedLabelValue      = "test"

		timeout  = time.Second * 100
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

			// Delete the namespacelabel for cleanup
			Expect(k8sClient.Delete(ctx, createdNamespaceLabel)).Should(Succeed())
		})
		It("Should not override the protected labels", func() {
			By("Creating a new NamespaceLabel with a protected label")
			ctx := context.Background()

			// Make a deep copy of the original defaultLabels map
			labelsToAdd := make(map[string]string)
			for k, v := range defaultLabels {
				labelsToAdd[k] = v
			}

			labelsToAdd[ProtectedLabelKey] = ProtectedLabelValue
			namespaceLabel := &namespacelabelv1.NamespaceLabel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "core.tutorial.kubebuilder.io/v1",
					Kind:       "NamespaceLabel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecondNamespaceLabelName,
					Namespace: Namespace,
				},
				Spec: namespacelabelv1.NamespaceLabelSpec{
					Labels: labelsToAdd,
				},
			}
			Expect(k8sClient.Create(ctx, namespaceLabel)).Should(Succeed())
			namespacelabelLookupKey := types.NamespacedName{Name: SecondNamespaceLabelName, Namespace: Namespace}
			createdNamespaceLabel := &namespacelabelv1.NamespaceLabel{}

			// We'll need to retry getting this newly created NamespaceLabel, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespacelabelLookupKey, createdNamespaceLabel)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Checking the new namespace doesn't include the protected label.")
			namespaceLookupKey := types.NamespacedName{Name: Namespace}
			createdNamespace := &corev1.Namespace{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespaceLookupKey, createdNamespace)
				if err != nil {
					return false
				}
				return reflect.DeepEqual(createdNamespace.Labels, defaultLabels)
			}, timeout, interval).Should(BeTrue())

			// Delete the namespacelabel for cleanup
			Expect(k8sClient.Delete(ctx, createdNamespaceLabel)).Should(Succeed())
		})
	})
	Context("When creating two NamespaceLabels", func() {
		const (
			FirstNamespaceLabelName  = "bob"
			SecondNamespaceLabelName = "larry"
		)
		It("Should join all their Labels to in the namespace's Labels", func() {
			By("Creating two new NamespaceLabels")
			firstNamespaceLabels := map[string]string{"bob": "8", "gary": "5"}
			secondNamespaceLabels := map[string]string{"bob": "7", "larry": "6"}

			expectedLables := make(map[string]string)

			// Join the maps to create the expected labels
			for k, v := range firstNamespaceLabels {
				expectedLables[k] = v
			}

			for k, v := range secondNamespaceLabels {
				expectedLables[k] = v
			}
			ctx := context.Background()

			// Create the first Namespace label
			firstNamespaceLabel := &namespacelabelv1.NamespaceLabel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "core.tutorial.kubebuilder.io/v1",
					Kind:       "NamespaceLabel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      FirstNamespaceLabelName,
					Namespace: Namespace,
				},
				Spec: namespacelabelv1.NamespaceLabelSpec{
					Labels: firstNamespaceLabels,
				},
			}
			Expect(k8sClient.Create(ctx, firstNamespaceLabel)).Should(Succeed())

			firstNamespacelabelLookupKey := types.NamespacedName{Name: FirstNamespaceLabelName, Namespace: Namespace}
			firstCreatedNamespaceLabel := &namespacelabelv1.NamespaceLabel{}

			// We'll need to retry getting this newly created NamespaceLabel, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, firstNamespacelabelLookupKey, firstCreatedNamespaceLabel)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// Create the second Namespace label
			secondNamespaceLabel := &namespacelabelv1.NamespaceLabel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "core.tutorial.kubebuilder.io/v1",
					Kind:       "NamespaceLabel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecondNamespaceLabelName,
					Namespace: Namespace,
				},
				Spec: namespacelabelv1.NamespaceLabelSpec{
					Labels: secondNamespaceLabels,
				},
			}
			Expect(k8sClient.Create(ctx, secondNamespaceLabel)).Should(Succeed())

			secondNamespacelabelLookupKey := types.NamespacedName{Name: SecondNamespaceLabelName, Namespace: Namespace}
			secondCreatedNamespaceLabel := &namespacelabelv1.NamespaceLabel{}

			// We'll need to retry getting this newly created NamespaceLabel, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secondNamespacelabelLookupKey, secondCreatedNamespaceLabel)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Checking the namespace's labels have been updated correctly")
			namespaceLookupKey := types.NamespacedName{Name: Namespace}
			createdNamespace := &corev1.Namespace{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespaceLookupKey, createdNamespace)
				if err != nil || createdNamespace.Labels == nil {
					return false
				}
				return reflect.DeepEqual(createdNamespace.Labels, expectedLables)
			}, timeout, interval).Should(BeTrue())

			// Delete the namespacelabel for cleanup
			Expect(k8sClient.Delete(ctx, firstNamespaceLabel)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, secondNamespaceLabel)).Should(Succeed())

		})
	})

})
