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

func logPrefix(dryRun bool) string {
	logPrefix := "dryRun=false"
	if dryRun {
		logPrefix = "dryRun=true"
	}

	return logPrefix
}

func IsLocalClusterName(routingWeight routingv1alpha1.RoutingWeight, clusterName string) bool {
	return routingWeight.Spec.ClusterName == clusterName
}

func UpdateIngress(ctx context.Context, apiClient client.Client, dryRun bool, ingress *networkingv1.Ingress) error {
	logger := log.FromContext(ctx)
	logPrefix := logPrefix(dryRun)

	logger.Info(fmt.Sprintf("%s Updating ingress object in api server", logPrefix))
	if dryRun {
		logger.Info(fmt.Sprintf("%s Dryrun of change. Doing nothing", logPrefix))
		return nil
	}

	return apiClient.Update(ctx, ingress)
}

func SetIngressAnnotations(ctx context.Context, ingress *networkingv1.Ingress, routingWeight routingv1alpha1.RoutingWeight) {
	logPrefix := "dryRun=false"
	if routingWeight.Spec.DryRun {
		logPrefix = "dryRun=true"
	}

	logger := log.FromContext(ctx)
	for _, annotation := range routingWeight.Spec.Annotations {
		value, ok := ingress.Annotations[annotation.Key]
		if ok {
			logger.Info(fmt.Sprintf("%s Existing annotation found on ingress: %s:'%s'", logPrefix, annotation.Key, value), "ingress", ingress.Name)
		}

		logger.Info(fmt.Sprintf("%s Setting annotation on ingress", logPrefix), "ingress", ingress.Name, "annotation", annotation.Value)
		ingress.Annotations[annotation.Key] = annotation.Value
	}
}

func IsIngressControlled(ingress networkingv1.Ingress) bool {
	value, ok := ingress.Annotations[controlledByAnnotationKey]

	return ok && value == "true"
}

func HandleIngress(ctx context.Context, client client.Client, clusterName string, ingress *networkingv1.Ingress) (v1alpha1.RoutingWeight, bool, error) {
	logger := log.FromContext(ctx)

	if !IsIngressControlled(*ingress) {
		logger.Info("Ingress is not controlled. skipping.", "ingress", ingress.Name)
		return v1alpha1.RoutingWeight{}, false, nil
	}

	routingWeights, err := GetRoutingWeightList(ctx, client)
	if err != nil {
		return v1alpha1.RoutingWeight{}, false, err
	}

	// If more than one in cluster, then error
	var localRoutingWeights []v1alpha1.RoutingWeight
	for _, routingWeight := range routingWeights.Items {
		if !IsLocalClusterName(routingWeight, clusterName) {
			continue
		}

		logger.Info("Found local cluster routingWeight", "routingWeight", routingWeight.Name)
		localRoutingWeights = append(localRoutingWeights, routingWeight)
	}

	if len(localRoutingWeights) == 0 {
		logger.Info("Found no local routingWeights. skipping.")
		return v1alpha1.RoutingWeight{}, false, nil
	}

	if len(localRoutingWeights) != 1 {
		return v1alpha1.RoutingWeight{}, false, fmt.Errorf("more than one local cluster routing weight found, existing due to possible conflicts in annotations")
	}

	routingWeight := localRoutingWeights[0]
	SetIngressAnnotations(ctx, ingress, routingWeight)

	return routingWeight, true, nil
}

func GetRoutingWeightList(ctx context.Context, client client.Client) (*v1alpha1.RoutingWeightList, error) {
	list := &v1alpha1.RoutingWeightList{}
	if err := client.List(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}
