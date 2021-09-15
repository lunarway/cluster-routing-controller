package networking

import (
	"context"
	"testing"

	"github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIngressController(t *testing.T) {
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

	t.Run("None controlled Ingress is keept unchanged", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, nonControlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects(nonControlledIngress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: nonControlledIngress.Namespace,
				Name:      nonControlledIngress.Name,
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

		assert.Equal(t, nonControlledIngress.Annotations, actualIngress.Annotations)
	})

	t.Run("Controlled Ingress is keept unchanged when no routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		cl := fake.NewClientBuilder().
			WithObjects(controlledIngress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledIngress.Namespace,
				Name:      controlledIngress.Name,
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

		assert.Equal(t, controlledIngress.Annotations, actualIngress.Annotations)
	})

	t.Run("Controlled Ingress is keept unchanged when no local routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects(controlledIngress, routingWeightResource).
			Build()
		sut := createSut(cl, s, "Another cluster")

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledIngress.Namespace,
				Name:      controlledIngress.Name,
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

		assert.Equal(t, controlledIngress.Annotations, actualIngress.Annotations)
	})

	t.Run("Controlled Ingress is keept unchanged when in dryRun mode", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		s.AddKnownTypes(v1alpha1.GroupVersion, dryRunRoutingWeightResource)
		cl := fake.NewClientBuilder().
			WithObjects(controlledIngress, dryRunRoutingWeightResource).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledIngress.Namespace,
				Name:      controlledIngress.Name,
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

		assert.Equal(t, controlledIngress.Annotations, actualIngress.Annotations)
	})

	t.Run("None controlled Ingress is unchanged when local routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, nonControlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		cl := fake.NewClientBuilder().
			WithObjects(nonControlledIngress, routingWeightResource).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: nonControlledIngress.Namespace,
				Name:      nonControlledIngress.Name,
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

		assert.Equal(t, nonControlledIngress.Annotations, actualIngress.Annotations)
	})

	t.Run("Controlled Ingress is changed when local routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource, controlledIngress).
			Build()
		sut := createSut(cl, s, clusterName)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledIngress.Namespace,
				Name:      controlledIngress.Name,
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

}

func createSut(c client.WithWatch, s *runtime.Scheme, clusterName string) *IngressReconciler {
	return &IngressReconciler{
		Client:      c,
		Scheme:      s,
		ClusterName: clusterName,
	}
}
