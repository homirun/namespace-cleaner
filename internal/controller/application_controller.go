/*
Copyright 2024.

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

package controller

import (
	"context"

	argoprojv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=argoproj.io,resources=applications/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Application")

	app := &argoprojv1alpha1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		logger.Error(err, "Unable to fetch Application")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ラベルの存在確認
	if val, ok := app.Labels["namespace-cleaner.homi.run/managed"]; ok && val == "true" {
		// finalizerの名前を定義
		finalizerName := "namespace-cleaner.homi.run/finalizer"
		// 存在しない場合は追加
		if !controllerutil.ContainsFinalizer(app, finalizerName) {
			controllerutil.AddFinalizer(app, finalizerName)
			if err := r.Update(ctx, app); err != nil {
				logger.Error(err, "Failed to add finalizer")
				return ctrl.Result{}, err
			}
			logger.Info("Added finalizer to Application")
		}

		// finalizerの処理
		if app.DeletionTimestamp != nil {
			logger.Info("Deleting Application")
			// spec.destination.namespaceに指定されたnamespaceを削除
			if err := r.Delete(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: app.Spec.Destination.Namespace}}); err != nil {
				logger.Error(err, "Failed to delete namespace")
				return ctrl.Result{}, err
			}
			logger.Info("Deleted namespace")
			// finalizerを削除
			controllerutil.RemoveFinalizer(app, finalizerName)
			if err := r.Update(ctx, app); err != nil {
				logger.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
			logger.Info("Removed finalizer from Application")
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&argoprojv1alpha1.Application{}).
		Complete(r)
}
