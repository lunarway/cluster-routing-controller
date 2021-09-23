package operator

import (
	"context"
	"fmt"

	"github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"
	routingv1alpha1 "github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	controlledByAnnotationKey = "routing.lunar.tech/controlled"
)

func logPrefix(routingWeight routingv1alpha1.RoutingWeight) string {
	logPrefix := ""
	if routingWeight.Spec.DryRun {
		logPrefix = "[dryRun]"
	}

	return logPrefix
}

func IsLocalClusterName(routingWeight routingv1alpha1.RoutingWeight, clusterName string) bool {
	return routingWeight.Spec.ClusterName == clusterName
}

func UpdateIngress(ctx context.Context, apiClient client.Client, routingWeight routingv1alpha1.RoutingWeight, ingress *networkingv1.Ingress) error {
	logger := log.FromContext(ctx)
	dryRun := routingWeight.Spec.DryRun
	logPrefix := logPrefix(routingWeight)

	logger.Info(fmt.Sprintf("%s Updating ingress object in api server", logPrefix))
	if dryRun {
		logger.Info(fmt.Sprintf("%s Dryrun of change. Doing nothing", logPrefix))
		return nil
	}

	return apiClient.Update(ctx, ingress)
}

func SetIngressAnnotations(ctx context.Context, ingress *networkingv1.Ingress, routingWeight routingv1alpha1.RoutingWeight) {
	logPrefix := logPrefix(routingWeight)

	logger := log.FromContext(ctx)
	for _, annotation := range routingWeight.Spec.Annotations {
		value, ok := ingress.Annotations[annotation.Key]
		if ok {
			logger.Info(fmt.Sprintf("%s Existing annotation found on ingress", logPrefix), "ingress", ingress.Name, "annotation", fmt.Sprintf("%s: %s", annotation.Key, value))
		}

		logger.Info(fmt.Sprintf("%s Setting annotation on ingress", logPrefix), "ingress", ingress.Name, "annotation", fmt.Sprintf("%s: %s", annotation.Key, annotation.Value))
		if !routingWeight.Spec.DryRun {
			ingress.Annotations[annotation.Key] = annotation.Value
		}
	}
}

func IsIngressControlled(ingress networkingv1.Ingress) bool {
	value, ok := ingress.Annotations[controlledByAnnotationKey]

	return ok && value == "true"
}

func DoesIngressNeedsUpdating(ctx context.Context, client client.Client, clusterName string, ingress *networkingv1.Ingress) (bool, v1alpha1.RoutingWeight, error) {
	logger := log.FromContext(ctx)

	if !IsIngressControlled(*ingress) {
		return false, v1alpha1.RoutingWeight{}, nil
	}

	routingWeights, err := GetRoutingWeightList(ctx, client)
	if err != nil {
		return false, v1alpha1.RoutingWeight{}, err
	}

	// If more than one in cluster, then error
	var localRoutingWeights []v1alpha1.RoutingWeight
	for _, routingWeight := range routingWeights.Items {
		if !IsLocalClusterName(routingWeight, clusterName) {
			continue
		}

		logger.Info("Found local cluster routingWeight")
		localRoutingWeights = append(localRoutingWeights, routingWeight)
	}

	if len(localRoutingWeights) == 0 {
		logger.Info("Found no local routingWeights. skipping.")
		return false, v1alpha1.RoutingWeight{}, nil
	}

	if len(localRoutingWeights) != 1 {
		return false, v1alpha1.RoutingWeight{}, fmt.Errorf("more than one local cluster routing weight found, existing due to possible conflicts in annotations")
	}

	routingWeight := localRoutingWeights[0]
	return true, routingWeight, nil
}

func GetRoutingWeightList(ctx context.Context, client client.Client) (*v1alpha1.RoutingWeightList, error) {
	list := &v1alpha1.RoutingWeightList{}
	if err := client.List(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}
