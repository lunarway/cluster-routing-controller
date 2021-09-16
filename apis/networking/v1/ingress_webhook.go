/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github/lunarway/cluster-routing-controller/internal/operator"

	admissionv1 "k8s.io/api/admission/v1"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var ingresslog = logf.Log.WithName("ingress-resource")

type IngressAnnotator struct {
	Client      client.Client
	decoder     *admission.Decoder
	ClusterName string
}

func (a *IngressAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

func (a *IngressAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &networkingv1.Ingress{}
	err := a.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	ingresslog.Info("IngressAnnotator found Ingress", "ingress", ingress.Name)
	if req.Operation != admissionv1.Create {
		return admission.Allowed("not a create operation")
	}

	_, _, err = operator.HandleIngress(ctx, a.Client, a.ClusterName, ingress)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshalledIngress, err := json.Marshal(ingress)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledIngress)
}

//+kubebuilder:webhook:path=/mutate-networking-k8s-io-v1-ingress,mutating=true,failurePolicy=fail,sideEffects=None,groups=networking.k8s.io,resources=ingresses,verbs=create;update,versions=v1,name=mingress.kb.io,admissionReviewVersions={v1,v1beta1}
