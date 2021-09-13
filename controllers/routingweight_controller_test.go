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
	t.Run("RoutingWeight without ingress", func(t *testing.T) {
		s := scheme.Scheme

		routingWeightResource := &v1alpha1.RoutingWeight{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoutingWeight",
				APIVersion: "v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: v1alpha1.RoutingWeightSpec{
				TargetCluster: "targetCluster",
				DryRun:        false,
				Annotations: []v1alpha1.Annotation{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			Status: v1alpha1.RoutingWeightStatus{},
		}
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)

		// Add tracked objects to the fake client simulating their existence in a k8s
		// cluster
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut := RoutingWeightReconciler{
			Client:      cl,
			Scheme:      s,
			ClusterName: "clusterName",
		}

		result, err := sut.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "namespace",
				Name:      "name",
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)
	})
}
