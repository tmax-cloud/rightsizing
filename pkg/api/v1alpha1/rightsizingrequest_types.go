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

type RightsizingRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	RightsizingService `json:",inline"`
	Query              string `json:"query" description:"prometheus query (optional)"`
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty" description:"prometheus labels which are key-value set(e.g pod_name=test)"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RightsizingRequest is the Schema for the rightsizingrequests API
// +kubebuilder:resource:shortName=rzreq
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Optimization",type=boolean,JSONPath=`.spec.optimization`
// +kubebuilder:printcolumn:name="Forecast",type=boolean,JSONPath=`.spec.forecast`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RightsizingRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RightsizingRequestSpec   `json:"spec,omitempty"`
	Status RightsizingRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RightsizingRequestList contains a list of RightsizingRequest
type RightsizingRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RightsizingRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RightsizingRequest{}, &RightsizingRequestList{})
}
