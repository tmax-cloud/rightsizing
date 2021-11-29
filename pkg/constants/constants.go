package constants

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RightsizingLabelName = "rightsizing-service"
)

const (
	RightsizingConfigMapNamespace = "rightsizing-operator-system"
	RightsizingConfigMapName      = "rightsizing-operator-configmap"
)

const (
	QueryEndpoint = "queries"
)

const (
	ServerContainerName  = "server"
	ServerContainerImage = "dbdydgur2244/rightsizing-server"
	ServerContainerPort  = "8000"
)

type CheckResultType int

const (
	CheckResultCreate  CheckResultType = 0
	CheckResultExisted CheckResultType = 1
	CheckResultError   CheckResultType = 2
)

func ServerContainerTemplate() v1.Container {
	return v1.Container{
		Name:  ServerContainerName,
		Image: ServerContainerImage,
		Command: []string{
			"uvicorn",
			"main:app",
		},
		Args: []string{
			"--reload",
			"--host", "0.0.0.0",
			"--port", ServerContainerPort,
		},
	}
}

func RightsizingDefaultPodTemplate(meta metav1.ObjectMeta) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", "rightsizing", meta.Name),
			Namespace:   meta.Namespace,
			Labels:      meta.Labels,
			Annotations: meta.Annotations,
		},
		Spec: v1.PodSpec{RestartPolicy: v1.RestartPolicyNever},
	}

	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = make(map[string]string)
	}
	pod.ObjectMeta.Labels[RightsizingLabelName] = meta.Name
	return pod
}

func RightsizingRequestDefaultPodTemplate(meta metav1.ObjectMeta) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", "rightsizingrequest", meta.Name),
			Namespace:   meta.Namespace,
			Labels:      meta.Labels,
			Annotations: meta.Annotations,
		},
		Spec: v1.PodSpec{RestartPolicy: v1.RestartPolicyNever},
	}

	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = make(map[string]string)
	}
	pod.ObjectMeta.Labels[RightsizingLabelName] = meta.Name
	return pod
}
