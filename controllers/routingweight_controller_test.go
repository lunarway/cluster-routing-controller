package controllers

import (
	"context"
	"testing"

	"github/lunarway/cluster-routing-controller/api/v1alpha1"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestRoutingWeightController(t *testing.T) {
	var (
		typeMeta = metav1.TypeMeta{
			Kind:       "RoutingWeight",
			APIVersion: "v1alpha1",
		}
		metadata = metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		}
		clusterName    = "clusterName"
		namespacedName = types.NamespacedName{
			Namespace: "namespace",
			Name:      "name",
		}
		annotation = v1alpha1.Annotation{
			Key:   "key",
			Value: "value",
		}
	)

	t.Run("Set annotation on ingress when control annotation is set", func(t *testing.T) {

	})

	t.Run("Does not set annotation on ingress when cluster names does not match", func(t *testing.T) {

	})

	t.Run("Does not set annotation on ingress when control annotation is not set", func(t *testing.T) {

	})

	t.Run("Updates annotation on ingress when annotation value has changed", func(t *testing.T) {

	})

	t.Run("Does nothing when no ingresses exist", func(t *testing.T) {
		s := scheme.Scheme

		routingWeightResource := &v1alpha1.RoutingWeight{
			TypeMeta:   typeMeta,
			ObjectMeta: metadata,
			Spec: v1alpha1.RoutingWeightSpec{
				TargetCluster: clusterName,
				DryRun:        false,
				Annotations:   []v1alpha1.Annotation{annotation},
			},
			Status: v1alpha1.RoutingWeightStatus{},
		}
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)

		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut := RoutingWeightReconciler{
			Client:      cl,
			Scheme:      s,
			ClusterName: clusterName,
		}

		result, err := sut.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: namespacedName,
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)
	})
}
