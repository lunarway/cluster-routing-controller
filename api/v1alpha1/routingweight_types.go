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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Annotation defines the desired annotations that the operator should set on each ingress controlled by it
type Annotation struct {
	// Key defines the annotation key
	Key string `json:"key"`
	// Value defines the annotation value
	Value string `json:"value"`
}

// RoutingWeightSpec defines the desired state of RoutingWeight
type RoutingWeightSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ClusterName defines which cluster this routingWeight belongs to
	// The operator will only apply routingWeights where the clusterName matches its own identity
	ClusterName string `json:"clusterName"`
	// DryRun defines if a Weight should be applied or simulated
	// If set to true it only write changes to stdout and no changes to any Ingress objects are done
	DryRun bool `json:"dryRun,omitempty"`

	// Annotations defines the desired Annotations
	Annotations []Annotation `json:"annotations"`
}

// RoutingWeightStatus defines the observed state of RoutingWeight
type RoutingWeightStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RoutingWeight is the Schema for the routingweights API
type RoutingWeight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoutingWeightSpec   `json:"spec,omitempty"`
	Status RoutingWeightStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoutingWeightList contains a list of RoutingWeight
type RoutingWeightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoutingWeight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoutingWeight{}, &RoutingWeightList{})
}
