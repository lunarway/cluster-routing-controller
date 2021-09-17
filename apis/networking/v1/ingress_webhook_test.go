package v1

import (
	"context"
	"encoding/json"
	"testing"

	"github/lunarway/cluster-routing-controller/apis/routing/v1alpha1"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIngressAnnotator(t *testing.T) {
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

	t.Run("None controlled Ingress is kept unchanged", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, nonControlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects().
			Build()
		sut, err := createSut(cl, s, clusterName)
		assert.NoError(t, err)

		marshalledIngress, err := json.Marshal(nonControlledIngress)
		assert.NoError(t, err)

		result := sut.Handle(ctx, admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object: runtime.RawExtension{
					Raw:    marshalledIngress,
					Object: nonControlledIngress,
				},
			},
		})

		assert.True(t, result.Allowed)
		assert.Nil(t, result.PatchType)
	})

	t.Run("Controlled Ingress is kept unchanged when no routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects().
			Build()
		sut, err := createSut(cl, s, clusterName)
		assert.NoError(t, err)

		marshalledIngress, err := json.Marshal(controlledIngress)
		assert.NoError(t, err)

		result := sut.Handle(ctx, admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object: runtime.RawExtension{
					Raw:    marshalledIngress,
					Object: controlledIngress,
				},
			},
		})

		assert.True(t, result.Allowed)
		assert.Nil(t, result.PatchType)
	})

	t.Run("Controlled Ingress is kept unchanged when no local routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut, err := createSut(cl, s, "Another Cluster")
		assert.NoError(t, err)

		marshalledIngress, err := json.Marshal(controlledIngress)
		assert.NoError(t, err)

		result := sut.Handle(ctx, admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object: runtime.RawExtension{
					Raw:    marshalledIngress,
					Object: controlledIngress,
				},
			},
		})

		assert.True(t, result.Allowed)
		assert.Nil(t, result.PatchType)
	})

	t.Run("Controlled Ingress is kept unchanged when in dryRun mode", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, dryRunRoutingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects(dryRunRoutingWeightResource).
			Build()
		sut, err := createSut(cl, s, clusterName)
		assert.NoError(t, err)

		marshalledIngress, err := json.Marshal(controlledIngress)
		assert.NoError(t, err)

		result := sut.Handle(ctx, admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object: runtime.RawExtension{
					Raw:    marshalledIngress,
					Object: controlledIngress,
				},
			},
		})

		assert.True(t, result.Allowed)
		assert.Nil(t, result.PatchType)
	})

	t.Run("None controlled Ingress is kept unchanged when local routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, nonControlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut, err := createSut(cl, s, clusterName)
		assert.NoError(t, err)

		marshalledIngress, err := json.Marshal(nonControlledIngress)
		assert.NoError(t, err)

		result := sut.Handle(ctx, admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object: runtime.RawExtension{
					Raw:    marshalledIngress,
					Object: controlledIngress,
				},
			},
		})

		assert.True(t, result.Allowed)
		assert.Nil(t, result.PatchType)
	})

	t.Run("Controlled Ingress is changed when local routingWeights are defined", func(t *testing.T) {
		s := scheme.Scheme
		s.AddKnownTypes(networkingv1.SchemeGroupVersion, controlledIngress)
		s.AddKnownTypes(v1alpha1.GroupVersion, routingWeightResource)
		s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.RoutingWeightList{})
		cl := fake.NewClientBuilder().
			WithObjects(routingWeightResource).
			Build()
		sut, err := createSut(cl, s, clusterName)
		assert.NoError(t, err)

		marshalledIngress, err := json.Marshal(controlledIngress)
		assert.NoError(t, err)

		result := sut.Handle(ctx, admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object: runtime.RawExtension{
					Raw:    marshalledIngress,
					Object: controlledIngress,
				},
			},
		})

		assert.True(t, result.Allowed)
		assert.NotNil(t, result.PatchType)
	})
}

func createSut(cl client.WithWatch, scheme *runtime.Scheme, clusterName string) (*IngressAnnotator, error) {
	decoder, err := admission.NewDecoder(scheme)
	if err != nil {
		return nil, err
	}

	return &IngressAnnotator{
		Client:      cl,
		decoder:     decoder,
		ClusterName: clusterName,
	}, nil
}
