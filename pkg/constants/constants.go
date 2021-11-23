package constants

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RightSizingLabelName = "rightsizing-service"
)

const (
	DefaultPrometheusUri = "http://prometheus-k8s.monitoring.svc.cluster.local"
)

const (
	OptimizationServiceContainerName  = "optimization"
	OptimizationServiceContainerImage = "docker.io/dbdydgur2244/rightsizing-optimization"
	// OptimizationServiceContainerImage = "tmaxcloudck/rightsizing-optimization"
)

const (
	ForecastServiceContainerName  = "forecast"
	ForecastServiceContainerImage = "docker.io/dbdydgur2244/rightsizing-forecast"
	// ForecastServiceContainerImage = "tmaxcloudck/rightsizing-forecast"
)

const (
	RequestServiceContainerName  = "request"
	RequestServiceContainerImage = "dbdydgur2244/rightsizing-request"
)

const (
	RequestServerContainerName  = "server"
	RequestServerContainerImage = "dbdydgur2244/rightsizing-request-server"
	RequestServerContainerPort  = "8000"
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

var ServiceContainerList = []string{
	OptimizationServiceContainerName,
	ForecastServiceContainerName,
}

var ServiceList = []string{
	OptimizationServiceContainerName,
	ForecastServiceContainerName,
}

// var ManagerServerContainer = v1.Container{}

func OptimizationContainerTemplate(url, namespace, name string) v1.Container {
	container := v1.Container{
		Name:  OptimizationServiceContainerName,
		Image: OptimizationServiceContainerImage,
		Command: []string{
			"python",
			"main.py",
		},
		Args: []string{
			"-url", url,
			"-ns", namespace,
			"-n", name,
			"-server_url", fmt.Sprintf("http://127.0.0.1:%s", ServerContainerPort),
		},
	}
	return container
}

func ForecastContainerTemplate(url, namespace, name string) v1.Container {
	container := v1.Container{
		Name:  ForecastServiceContainerName,
		Image: ForecastServiceContainerImage,
		Command: []string{
			"python",
			"main.py",
		},
		Args: []string{
			"-url", url,
			"-ns", namespace,
			"-n", name,
			"-server_url", fmt.Sprintf("http://127.0.0.1:%s", ServerContainerPort),
		},
	}
	return container
}

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
	pod.ObjectMeta.Labels[RightSizingLabelName] = meta.Name
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
	pod.ObjectMeta.Labels[RightSizingLabelName] = meta.Name
	return pod
}
