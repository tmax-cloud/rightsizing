package v1alpha1

type RightsizingService struct {
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
