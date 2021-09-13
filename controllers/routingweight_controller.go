/*
Copyright 2021.

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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	routingv1alpha1 "github/lunarway/cluster-routing-controller/api/v1alpha1"
)

// RoutingWeightReconciler reconciles a RoutingWeight object
type RoutingWeightReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Log logr.Logger

	ClusterName string
}

//+kubebuilder:rbac:groups=routing.lunar.tech,resources=routingweights,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=routing.lunar.tech,resources=routingweights/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=routing.lunar.tech,resources=routingweights/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RoutingWeight object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RoutingWeightReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	routingWeight := &routingv1alpha1.RoutingWeight{}
	err := r.Get(ctx, req.NamespacedName, routingWeight)
	if err != nil {
		r.Log.Error(err, "fetch RoutingWeight from kubernetes API")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.Log.Info("Reconciling RoutingWeight: %s", routingWeight.ClusterName)
	if routingWeight.ClusterName != r.ClusterName{
		r.Log.Info("RoutingWeight ClusterName did not match current cluster name. Skipping.")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoutingWeightReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&routingv1alpha1.RoutingWeight{}).
		Complete(r)
}