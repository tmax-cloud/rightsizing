package controllers

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rightsizing/api/v1alpha1"
	"rightsizing/constants"
)

func RightSizingDefaultPodTemplate(service *v1alpha1.Rightsizing) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Labels:      service.Labels,
			Annotations: service.Annotations,
		},
		Spec: v1.PodSpec{RestartPolicy: v1.RestartPolicyNever},
	}

	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = make(map[string]string)
	}
	pod.ObjectMeta.Labels[constants.RightSizingLabelName] = service.Name
	return pod
}

func PodTemplate(service *v1alpha1.Rightsizing) *v1.Pod {
	pod := RightSizingDefaultPodTemplate(service)

	url := constants.DefaultPrometheusUri
	if service.Spec.PrometheusUri != nil {
		url = *service.Spec.PrometheusUri
	}
	namespace := service.Spec.PodNamespace
	name := service.Spec.PodName

	if service.Spec.Forecast != nil {
		pod.Spec.Containers = append(pod.Spec.Containers, constants.ForecastContainerTemplate(url, namespace, name))
	}
	if service.Spec.Optimization != nil {
		pod.Spec.Containers = append(pod.Spec.Containers, constants.OptimizationContainerTemplate(url, namespace, name))
	}
	if service.Spec.Trace != nil {
		if *service.Spec.Trace {
			pod.Spec.Containers = append(pod.Spec.Containers, constants.TraceContainerTemplate(url, namespace, name))
		}
	}
	pod.Spec.Containers = append(pod.Spec.Containers, constants.ServerContainerTemplate())
	return &pod
}
