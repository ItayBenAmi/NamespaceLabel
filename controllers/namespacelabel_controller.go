/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"sort"
	"time"

	"go.uber.org/zap"
	kcore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "tutorial.kubebuilder.io/project/api/v1"
	utils "tutorial.kubebuilder.io/project/internal/utils"
)

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// getProtectedLabels returns a slice of all the protected labels, which we can not allow to be overriden by the NamespaceLabels.
func getProtectedLabels() []string {
	return []string{"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component",
		"app.kubernetes.io/part-of", "app.kubernetes.io/managed-by"}
}

// filterNamespaceLabels filters out NamespaceLabels with nil values for LastModifiedTime
func filterNamespaceLabels(namespaceLabels []corev1.NamespaceLabel) []corev1.NamespaceLabel {
	filteredNamespaceLabels := []corev1.NamespaceLabel{}

	for index := range namespaceLabels {
		if namespaceLabels[index].Status.LastModifiedTime != nil {
			filteredNamespaceLabels = append(filteredNamespaceLabels, namespaceLabels[index])
		}
	}

	return filteredNamespaceLabels
}

// sortNamespaceLabelsByLastModifiedTime receives a slice of NamespaceLabels and returns the slice sorted by
// the LastModifiedTime
func sortNamespaceLabelsByLastModifiedTime(namespaceLabels []corev1.NamespaceLabel) []corev1.NamespaceLabel {
	sortedNameSpaceLabels := make([]corev1.NamespaceLabel, len(namespaceLabels))
	copy(sortedNameSpaceLabels, namespaceLabels)

	sort.Slice(sortedNameSpaceLabels, func(i, j int) bool {
		return sortedNameSpaceLabels[j].Status.LastModifiedTime.After(sortedNameSpaceLabels[i].Status.LastModifiedTime.Time)
	})

	return sortedNameSpaceLabels
}

// getFinalLabelsMap receives a NamespaceLabelList, sorts the items based on their LastModifiedTime field
// and constructs a labels map containing the labels to set in the NamespaceLabel.
// Two things are taken into cosideration here:
// 1. To not include protected labels.
// 2. If two or more NamespaceLabels have the same label, take the label from the most recently modified NamespaceLabel.
func getFinalLabelsMap(namespaceLabelList corev1.NamespaceLabelList) map[string]string {
	var protectedLabels = getProtectedLabels()

	namespaceLabels := make([]corev1.NamespaceLabel, len(namespaceLabelList.Items))
	copy(namespaceLabels, namespaceLabelList.Items)
	namespaceLabels = filterNamespaceLabels(namespaceLabels)
	namespaceLabels = sortNamespaceLabelsByLastModifiedTime(namespaceLabels)

	finalNamespaceLabels := make(map[string]string)
	for _, namespaceLabel := range namespaceLabels {
		for labelName, labelValue := range namespaceLabel.Spec.Labels {
			if utils.Contains(protectedLabels, labelName) == false {
				finalNamespaceLabels[labelName] = labelValue
			}
		}
	}

	return finalNamespaceLabels
}

//+kubebuilder:rbac:groups=core.tutorial.kubebuilder.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tutorial.kubebuilder.io,resources=namespacelabels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.tutorial.kubebuilder.io,resources=namespacelabels/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=namespaces/status,verbs=get

func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var logger = utils.SetUpLogger()
	logger = *logger.With(zap.String("Namespace", req.Namespace), zap.String("NamespaceLabel", req.Name))
	logger.Info("Beginning reconcile logic for NamespaceLabel")

	// Fetch the current NamespaceLabel
	logger.Debug("Fetching the current NamespaceLabel")
	var namespaceLabel corev1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		logger.Error("Unable to fetch NamespaceLabel", zap.Error(errors.New(err.Error())))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Debug("Finished fetching the current NamespaceLabel")

	// Update the current NamespaceLabel's LastModifiedTime field
	logger.Debug("Updating the NamespaceLabel's LastModified field")
	namespaceLabel.Status.LastModifiedTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, &namespaceLabel); err != nil {
		logger.Error("Unable to update NamespaceLabel", zap.Error(errors.New(err.Error())))
		return ctrl.Result{}, err
	}
	logger.Debug("Finished updating the NamespaceLabel's LastModified field")

	// List all the NamespaceLabel objects in the current namespace.
	logger.Debug("Fetching all NamespaceLabels")
	var namespaceLabels corev1.NamespaceLabelList
	if err := r.List(ctx, &namespaceLabels, client.InNamespace(req.Namespace)); err != nil {
		logger.Error("Unable to list all NamespaceLabels", zap.Error(errors.New(err.Error())))
		return ctrl.Result{}, err
	}
	logger.Debug("Finished fetching all NamespaceLabels")

	// Fetch the current Namespace
	logger.Debug("Fetching the current Namespace")
	var namespace kcore.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: req.Namespace}, &namespace); err != nil {
		logger.Error("Unable to fetch Namespace", zap.Error(errors.New(err.Error())))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Debug("Finished fetching the current Namespace")

	// Update the labels in the current Namespace
	logger.Debug("Updating the Namespace with the new labels")
	namespaceLabelsToAdd := getFinalLabelsMap(namespaceLabels)
	namespace.Labels = namespaceLabelsToAdd
	if err := r.Update(ctx, &namespace); err != nil {
		logger.Error("Unable to update Namespace", zap.Error(errors.New(err.Error())))
		return ctrl.Result{}, err
	}
	logger.Debug("Finished updating the Namespace with the new labels")

	logger.Info("Finished Reconciler logic for the NamespaceLabel")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.NamespaceLabel{}).
		Complete(r)
}
