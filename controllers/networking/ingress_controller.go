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

package networking

import (
	"context"
	"fmt"

	"github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"
	"github/lunarway/cluster-routing-controller/internal/operator"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ClusterName string
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		logger.Error(err, "fetch ingress from kubernetes API")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if !operator.IsIngressControlled(*ingress) {
		logger.Info("Ingress is not controlled. skipping.", "ingress", ingress.Name)
		return ctrl.Result{}, nil
	}

	routingWeights, err := r.getRoutingWeightList(ctx)
	if err != nil {
		logger.Error(err, "get routingWeights")
		return reconcile.Result{}, err
	}

	// If more than one in cluster, then error
	var localRoutingWeights []v1alpha1.RoutingWeight
	for _, routingWeight := range routingWeights.Items {
		if !operator.IsLocalClusterName(routingWeight, r.ClusterName) {
			continue
		}

		logger.Info("Found local cluster routingWeight", "routingWeight", routingWeight.Name)
		localRoutingWeights = append(localRoutingWeights, routingWeight)
	}

	if len(localRoutingWeights) == 0 {
		logger.Info("Found no local routingWeights. skipping.")
		return ctrl.Result{}, nil
	}

	if len(localRoutingWeights) != 1 {
		logger.Error(fmt.Errorf("more than one local cluster routing weight found"), "Existing due to possible conflicts in annotations")
		return reconcile.Result{}, err
	}

	err = operator.SetRoutingWeightAnnotations(ctx, r.Client, *ingress, localRoutingWeights[0])
	if err != nil {
		logger.Error(err, "set annotations")
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) getRoutingWeightList(ctx context.Context) (*v1alpha1.RoutingWeightList, error) {
	list := &v1alpha1.RoutingWeightList{}
	if err := r.List(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Complete(r)
}
