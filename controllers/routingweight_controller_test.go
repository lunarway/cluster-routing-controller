package controllers

import (
	"context"
	"testing"

	"github/lunarway/cluster-routing-controller/api/v1alpha1"

	networkingv1 "k8s.io/api/networking/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

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

		routingWeightResource = &v1alpha1.RoutingWeight{
			TypeMeta:   typeMeta,
			ObjectMeta: metadata,
			Spec: v1alpha1.RoutingWeightSpec{
				TargetCluster: clusterName,
				DryRun:        false,
				Annotations:   []v1alpha1.Annotation{annotation},
			},
			Status: v1alpha1.RoutingWeightStatus{},
		}
		ingress = &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-name",
				Namespace: "ingress-namespace",
				Annotations: map[string]string{
					"routing.lunar.tech/controlled": "true",
				},
			},
			Spec:   networkingv1.IngressSpec{},
			Status: networkingv1.IngressStatus{},
		}
		ctx = context.Background()
	)

	t.Run("Set annotations on ingress when control annotation is set", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, ingress)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource, ingress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: namespacedName,
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualIngress := &networkingv1.Ingress{}
		err = cl.Get(ctx, types.NamespacedName{
			Name:      ingress.Name,
			Namespace: ingress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		expectedIngress := ingress
		for key, value := range routingWeightResource.Annotations {
			expectedIngress.Annotations[key] = value
		}

		assert.Equal(t, expectedIngress, actualIngress)
	})

	t.Run("Does not set annotation on ingress when cluster names does not match", func(t *testing.T) {

	})

	t.Run("Does not set annotation on ingress when control annotation is not set", func(t *testing.T) {

	})

	t.Run("Updates annotation on ingress when annotation value has changed", func(t *testing.T) {

	})

	t.Run("Does nothing when no ingresses exist", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: namespacedName,
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)
	})
}

func createSut(c client.WithWatch, s *runtime.Scheme, clusterName string) RoutingWeightReconciler {
	return RoutingWeightReconciler{
		Client:      c,
		Scheme:      s,
		ClusterName: clusterName,
	}
}
