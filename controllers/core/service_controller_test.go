package core

import (
	"context"
	"testing"

	"github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"

	networkingv1 "k8s.io/api/networking/v1"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestServiceController(t *testing.T) {
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
			},
			Status: v1alpha1.RoutingWeightStatus{},
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

		explicitNonControlledService = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nonControlledServiceName",
				Namespace: "serviceNamespace",
				Annotations: map[string]string{
					"routing.lunar.tech/controlled": "false",
				},
			},
			Spec:   corev1.ServiceSpec{},
			Status: corev1.ServiceStatus{},
		}

		ctx = context.Background()
	)

	t.Run("None controlled Service is kept unchanged", func(t *testing.T) {
		sut := createSut(t, clusterName, nonControlledService)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: nonControlledService.Namespace,
				Name:      nonControlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      nonControlledService.Name,
			Namespace: nonControlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, nonControlledService.Annotations, actualService.Annotations)
	})

	t.Run("Controlled Ingress is kept unchanged when no routingWeights are defined", func(t *testing.T) {
		sut := createSut(t, clusterName, controlledService)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledService.Namespace,
				Name:      controlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, controlledService.Annotations, actualService.Annotations)
	})

	t.Run("Controlled Ingress is kept unchanged when no local routingWeights are defined", func(t *testing.T) {
		sut := createSut(t, "Another cluster", controlledService, routingWeightResource)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledService.Namespace,
				Name:      controlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, controlledService.Annotations, actualService.Annotations)
	})

	t.Run("Controlled Ingress is kept unchanged when in dryRun mode", func(t *testing.T) {
		sut := createSut(t, clusterName, controlledService, dryRunRoutingWeightResource)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledService.Namespace,
				Name:      controlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, controlledService.Annotations, actualService.Annotations)
	})

	t.Run("None controlled Ingress is kept unchanged when local routingWeights are defined", func(t *testing.T) {
		sut := createSut(t, clusterName, nonControlledService, routingWeightResource)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: nonControlledService.Namespace,
				Name:      nonControlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      nonControlledService.Name,
			Namespace: nonControlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, nonControlledService.Annotations, actualService.Annotations)
	})

	t.Run("Controlled Ingress is changed when local routingWeights are defined", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, controlledService)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: controlledService.Namespace,
				Name:      controlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		expectedAnnotations := map[string]string{
			"key":                           "value",
			"routing.lunar.tech/controlled": "true",
		}

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      controlledService.Name,
			Namespace: controlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualService.Annotations)
	})

	t.Run("Explicitly non-controlled Ingress is changed when local routingWeights are defined", func(t *testing.T) {
		sut := createSut(t, clusterName, routingWeightResource, explicitNonControlledService)

		result, err := sut.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: explicitNonControlledService.Namespace,
				Name:      explicitNonControlledService.Name,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{}, result)

		expectedAnnotations := map[string]string{
			"routing.lunar.tech/controlled": "false",
		}

		actualService := &corev1.Service{}
		err = sut.Get(ctx, types.NamespacedName{
			Name:      explicitNonControlledService.Name,
			Namespace: explicitNonControlledService.Namespace,
		}, actualService)
		assert.NoError(t, err)

		assert.Equal(t, expectedAnnotations, actualService.Annotations)
	})
}

func createSut(t *testing.T, clusterName string, objects ...client.Object) *ServiceReconciler {
	t.Helper()

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeight{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
	s.AddKnownTypes(networkingv1.SchemeGroupVersion, &networkingv1.Ingress{})

	client := fake.NewClientBuilder().
		WithObjects(objects...).
		Build()

	return &ServiceReconciler{
		Client:      client,
		Scheme:      s,
		ClusterName: clusterName,
	}
}
