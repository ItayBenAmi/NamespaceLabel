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
	"sort"
	"time"

	kcore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "tutorial.kubebuilder.io/project/api/v1"
)

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func getProtectedLabels() []string {
	return []string{"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component",
		"app.kubernetes.io/part-of", "app.kubernetes.io/managed-by"}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

//+kubebuilder:rbac:groups=core.tutorial.kubebuilder.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.tutorial.kubebuilder.io,resources=namespacelabels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.tutorial.kubebuilder.io,resources=namespacelabels/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=namespaces/status,verbs=get

func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Fetch the current NamespaceLabel
	var namespaceLabel corev1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		// log.Error(err, "unable to fetch NamespaceLabel")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Update the current NamespaceLabel's LastModifiedTime field
	namespaceLabel.Status.LastModifiedTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, &namespaceLabel); err != nil {
		return ctrl.Result{}, err
	}

	var protectedLables = getProtectedLabels()

	var namespaceLabels corev1.NamespaceLabelList
	if err := r.List(ctx, &namespaceLabels, client.InNamespace(req.Namespace)); err != nil {
		// log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}

	sort.Slice(namespaceLabels.Items, func(i, j int) bool {
		return namespaceLabels.Items[j].Status.LastModifiedTime.After(namespaceLabels.Items[i].Status.LastModifiedTime.Time)
	})

	var namespace kcore.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: req.Namespace}, &namespace); err != nil {
		// log.Error(err, "unable to fetch NamespaceLabel")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	namespaceLabelsToAdd := make(map[string]string)
	for _, namespaceLabel := range namespaceLabels.Items {
		for labelName, labelValue := range namespaceLabel.Spec.Labels {
			if contains(protectedLables, labelName) == false {
				namespaceLabelsToAdd[labelName] = labelValue
			}
		}
	}

	namespace.Labels = namespaceLabelsToAdd
	if err := r.Update(ctx, &namespace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

var (
	namespaceLabelOwnerKey = ".metadata.controller"
	apiGVStr               = corev1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.NamespaceLabel{}).
		Complete(r)
}
