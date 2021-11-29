package v1alpha1

type QueryParam struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="http://prometheus-k8s.monitoring.svc.cluster.local"
	PrometheusUri *string `json:"prometheusUri,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Optimization *bool `json:"optimization,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Forecast *bool `json:"forecast,omitempty"`
}

type PrometheusQuery struct {
	Query  string            `json:"query,omitempty" description:"prometheus query"`
	Labels map[string]string `json:"labels,omitempty" description:"prometheus label set which container key-value"`
	// +kubebuilder:validation:Optional
	Optimization *bool `json:"optimization,omitempty" description:"overwrite upper layer optimization value"`
	// +kubebuilder:validation:Optional
	Forecast *bool `json:"forecast,omitempty" description:"overwrite upper layer forecast value"`
}
