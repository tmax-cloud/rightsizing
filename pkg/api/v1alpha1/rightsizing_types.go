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

// RightsizingSpec defines the desired state of Rightsizing
type RightsizingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	QueryParam `json:",inline"`
	// Pod 관련 정보
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
	Trace        *bool  `json:"trace,omitempty"`
	// +kubebuilder:validation:Optional
	TraceCycle *string `json:"traceCycle,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Rightsizing is the Schema for the rightsizings API
// +kubebuilder:resource:shortName=rz
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Optimization",type=boolean,JSONPath=`.spec.optimization`
// +kubebuilder:printcolumn:name="Forecast",type=boolean,JSONPath=`.spec.forecast`
// +kubebuilder:printcolumn:name="Trace",type=boolean,JSONPath=`.spec.trace`
// +kubebuilder:printcolumn:name="Trace Cycle",type=string,JSONPath=`.spec.traceCycle`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Rightsizing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RightsizingSpec   `json:"spec,omitempty"`
	Status RightsizingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RightsizingList contains a list of Rightsizing
type RightsizingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rightsizing `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rightsizing{}, &RightsizingList{})
}
