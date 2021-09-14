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

		clusterName = "clusterName"
		annotation  = v1alpha1.Annotation{
			Key:   "key",
			Value: "value",
		}

		routingWeightResource = &v1alpha1.RoutingWeight{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      "routingWeight",
				Namespace: "routingWeightNamespace",
			},
			Spec: v1alpha1.RoutingWeightSpec{
				ClusterName: clusterName,
				DryRun:      false,
				Annotations: []v1alpha1.Annotation{annotation},
			},
			Status: v1alpha1.RoutingWeightStatus{},
		}

		dryRunRoutingWeightResource = &v1alpha1.RoutingWeight{
			TypeMeta: typeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dryRunRoutingWeight",
				Namespace: "routingWeightNamespace",
			},
			Spec: v1alpha1.RoutingWeightSpec{
				ClusterName: clusterName,
				DryRun:      true,
				Annotations: []v1alpha1.Annotation{annotation},
			},
			Status: v1alpha1.RoutingWeightStatus{},
		}
		controlledIngress = &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "controlledIngressName",
				Namespace: "ingressNamespace",
				Annotations: map[string]string{
					"routing.lunar.tech/controlled": "true",
				},
			},
			Spec:   networkingv1.IngressSpec{},
			Status: networkingv1.IngressStatus{},
		}

		controlledIngressWithWeights = &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "controlledIngressWithWeightsName",
				Namespace: "ingressNamespace",
				Annotations: map[string]string{
					"routing.lunar.tech/controlled": "true",
					"key":                           "value2",
				},
			},
			Spec:   networkingv1.IngressSpec{},
			Status: networkingv1.IngressStatus{},
		}

		nonControlledIngress = &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: "networking.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nonControlledIngressName",
				Namespace: "ingressNamespace",
			},
			Spec:   networkingv1.IngressSpec{},
			Status: networkingv1.IngressStatus{},
		}
		ctx = context.Background()
	)

	t.Run("Set annotations on controlledIngress when control annotation is set", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource, controlledIngress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: routingWeightResource.Namespace,
				Name:      routingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		expectedAnnotations := map[string]string{
			"key":                           "value",
			"routing.lunar.tech/controlled": "true",
		}

		actualIngress := &networkingv1.Ingress{}
		err = cl.Get(ctx, types.NamespacedName{
			Name:      controlledIngress.Name,
			Namespace: controlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualIngress.Annotations)
	})

	t.Run("Does not set annotation on controlledIngress when cluster names does not match", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource, controlledIngress).
			Build()
		sut := createSut(cl, s, "Another cluster name")

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: metav1.ObjectMeta{
					Name:      "routingWeight",
					Namespace: "routingWeightNamespace",
				}.Namespace,
				Name: metav1.ObjectMeta{
					Name:      "routingWeight",
					Namespace: "routingWeightNamespace",
				}.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualIngress := &networkingv1.Ingress{}
		err = cl.Get(ctx, types.NamespacedName{
			Name:      controlledIngress.Name,
			Namespace: controlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, controlledIngress, actualIngress)
	})

	t.Run("Does not set annotation on ingress when control annotation is not set", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, nonControlledIngress)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource, nonControlledIngress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: routingWeightResource.Namespace,
				Name:      routingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualIngress := &networkingv1.Ingress{}
		err = cl.Get(ctx, types.NamespacedName{
			Name:      nonControlledIngress.Name,
			Namespace: nonControlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, nonControlledIngress, actualIngress)
	})

	t.Run("Does not set annotation on ingress when control annotation is set but is in dryRun mode", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, dryRunRoutingWeightResource)
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		cl := fake.NewClientBuilder().
			WithObjects(dryRunRoutingWeightResource, controlledIngress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: dryRunRoutingWeightResource.Namespace,
				Name:      dryRunRoutingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualIngress := &networkingv1.Ingress{}
		err = cl.Get(ctx, types.NamespacedName{
			Name:      controlledIngress.Name,
			Namespace: controlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, controlledIngress, actualIngress)
	})

	t.Run("Updates annotation on ingress when annotation value has changed", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngressWithWeights)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource, controlledIngressWithWeights).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: routingWeightResource.Namespace,
				Name:      routingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		expectedAnnotations := map[string]string{
			"key":                           "value",
			"routing.lunar.tech/controlled": "true",
		}

		actualIngress := &networkingv1.Ingress{}
		err = cl.Get(ctx, types.NamespacedName{
			Name:      controlledIngressWithWeights.Name,
			Namespace: controlledIngressWithWeights.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualIngress.Annotations)
	})

	t.Run("Does nothing when no ingresses exist", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: metav1.ObjectMeta{
					Name:      "routingWeight",
					Namespace: "routingWeightNamespace",
				}.Namespace,
				Name: metav1.ObjectMeta{
					Name:      "routingWeight",
					Namespace: "routingWeightNamespace",
				}.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)
	})
}

func createSut(c client.WithWatch, s *runtime.Scheme, clusterName string) *RoutingWeightReconciler {
	return &RoutingWeightReconciler{
		Client:      c,
		Scheme:      s,
		ClusterName: clusterName,
	}
}
