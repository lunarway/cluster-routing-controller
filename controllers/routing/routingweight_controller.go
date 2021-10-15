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

package routing

import (
	"context"

	routingv1alpha1 "github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"
	"github/lunarway/cluster-routing-controller/internal/operator"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RoutingWeightReconciler reconciles a RoutingWeight object
type RoutingWeightReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ClusterName string
}

//+kubebuilder:rbac:groups=routing.lunar.tech,resources=routingweights,verbs=get;list;watch;update;patch

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
	logger := log.FromContext(ctx)

	routingWeight := &routingv1alpha1.RoutingWeight{}
	err := r.Get(ctx, req.NamespacedName, routingWeight)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		logger.Error(err, "fetch RoutingWeight from kubernetes API")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	logger.Info("Reconciling RoutingWeight")
	if !operator.IsLocalClusterName(*routingWeight, r.ClusterName) {
		logger.Info("RoutingWeight ClusterName did not match current cluster name. Skipping.", "RoutingWeightClusterName", routingWeight.Spec.ClusterName, "ClusterName", r.ClusterName)
		return ctrl.Result{}, nil
	}

	// Handle ingresses
	ingressList, err := r.getIngressList(ctx)
	if err != nil {
		logger.Error(err, "get ingressList")
		return reconcile.Result{}, err
	}

	ingresses := getControlledIngresses(ingressList.Items)
	for _, ingress := range ingresses {
		ingressPtr := &ingress
		operator.SetIngressAnnotations(ctx, ingressPtr, *routingWeight)

		err = operator.UpdateIngress(ctx, r.Client, *routingWeight, ingressPtr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Handle services
	serviceList, err := r.getServiceList(ctx)
	if err != nil {
		logger.Error(err, "get serviceList")
		return reconcile.Result{}, err
	}

	services := getControlledServices(serviceList.Items)
	for _, service := range services {
		servicesPtr := &service
		operator.SetServiceAnnotations(ctx, servicesPtr, *routingWeight)

		err = operator.UpdateService(ctx, r.Client, *routingWeight, servicesPtr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func getControlledIngresses(items []networkingv1.Ingress) []networkingv1.Ingress {
	var ingresses []networkingv1.Ingress

	for _, ingress := range items {
		if !operator.IsResourceControlled(ingress.Annotations) {
			continue
		}

		ingresses = append(ingresses, ingress)
	}

	return ingresses
}

func getControlledServices(items []corev1.Service) []corev1.Service {
	var services []corev1.Service

	for _, service := range items {
		if !operator.IsResourceControlled(service.Annotations) {
			continue
		}

		services = append(services, service)
	}

	return services
}

func (r *RoutingWeightReconciler) getIngressList(ctx context.Context) (*networkingv1.IngressList, error) {
	ingressList := &networkingv1.IngressList{}
	if err := r.List(ctx, ingressList); err != nil {
		return nil, err
	}
	return ingressList, nil
}

func (r *RoutingWeightReconciler) getServiceList(ctx context.Context) (*corev1.ServiceList, error) {
	serviceList := corev1.ServiceList{}
	if err := r.List(ctx, &serviceList); err != nil {
		return nil, err
	}
	return &serviceList, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoutingWeightReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&routingv1alpha1.RoutingWeight{}).
		Complete(r)
}
