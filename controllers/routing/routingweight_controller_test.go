package routing

import (
	"context"
	"testing"

	"github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
		controlledService = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "controlledServiceName",
				Namespace: "serviceNamespace",
				Annotations: map[string]string{
					"routing.lunar.tech/controlled": "true",
				},
			},
			Spec:   corev1.ServiceSpec{},
			Status: corev1.ServiceStatus{},
		}

		controlledServiceWithWeights = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "controlledServiceName",
				Namespace: "serviceNamespace",
				Annotations: map[string]string{
					"routing.lunar.tech/controlled": "true",
					"key":                           "value2",
				},
			},
			Spec:   corev1.ServiceSpec{},
			Status: corev1.ServiceStatus{},
		}

		nonControlledService = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nonControlledServiceName",
				Namespace: "serviceNamespace",
			},
			Spec:   corev1.ServiceSpec{},
			Status: corev1.ServiceStatus{},
		}

		ctx = context.Background()
	)

	t.Run("Set annotations on ingress when control annotation is set", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, controlledIngress)

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
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledIngress.Name,
			Namespace: controlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualIngress.Annotations)
	})

	t.Run("Does not set annotation on ingress when cluster names does not match", func(t *testing.T) {
		sut := createSut(t, "Another cluster name", routingWeightResource, controlledIngress)

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
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledIngress.Name,
			Namespace: controlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, controlledIngress, actualIngress)
	})

	t.Run("Does not set annotation on ingress when control annotation is not set", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, nonControlledIngress)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: routingWeightResource.Namespace,
				Name:      routingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualIngress := &networkingv1.Ingress{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      nonControlledIngress.Name,
			Namespace: nonControlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, nonControlledIngress, actualIngress)
	})

	t.Run("Does not set annotation on ingress when control annotation is set but is in dryRun mode", func(t *testing.T) {
		sut := createSut(t, clusterName, dryRunRoutingWeightResource, controlledIngress)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: dryRunRoutingWeightResource.Namespace,
				Name:      dryRunRoutingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualIngress := &networkingv1.Ingress{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledIngress.Name,
			Namespace: controlledIngress.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, controlledIngress, actualIngress)
	})

	t.Run("Updates annotation on ingress when annotation already exists with different value", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, controlledIngressWithWeights)

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
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledIngressWithWeights.Name,
			Namespace: controlledIngressWithWeights.Namespace,
		}, actualIngress)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualIngress.Annotations)
	})

	t.Run("Does nothing when no resources exist", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource)

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

	t.Run("Set annotations on service when control annotation is set", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, controlledService)

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

		actualService := corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, &actualService)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualService.Annotations)
	})

	t.Run("Does not set annotation on service when cluster names does not match", func(t *testing.T) {
		sut := createSut(t, "Another cluster name", routingWeightResource, controlledService)

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

		actualService := corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, &actualService)
		assert.NoError(t, err)

		assert.Equal(t, controlledService, actualService)
	})

	t.Run("Does not set annotation on service when control annotation is not set", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, nonControlledService)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: routingWeightResource.Namespace,
				Name:      routingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      nonControlledService.Name,
			Namespace: nonControlledService.Namespace,
		}, &actualService)
		assert.NoError(t, err)

		assert.Equal(t, nonControlledService, actualService)
	})

	t.Run("Does not set annotation on service when control annotation is set but is in dryRun mode", func(t *testing.T) {
		sut := createSut(t, clusterName, dryRunRoutingWeightResource, controlledService)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: dryRunRoutingWeightResource.Namespace,
				Name:      dryRunRoutingWeightResource.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, &actualService)
		assert.NoError(t, err)

		assert.Equal(t, controlledService, actualService)
	})

	t.Run("Updates annotation on service when annotation already exists with different value", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, controlledServiceWithWeights)

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

		actualService := corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledServiceWithWeights.Name,
			Namespace: controlledServiceWithWeights.Namespace,
		}, &actualService)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualService.Annotations)
	})
}

func createSut(t *testing.T, clusterName string, objects ...client.Object) *RoutingWeightReconciler {
	t.Helper()

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeight{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
	s.AddKnownTypes(networkingv1.SchemeGroupVersion, &networkingv1.Ingress{})
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Service{})

	client := fake.NewClientBuilder().
		WithObjects(objects...).
		Build()

	return &RoutingWeightReconciler{
		Client:      client,
		Scheme:      s,
		ClusterName: clusterName,
	}
}
