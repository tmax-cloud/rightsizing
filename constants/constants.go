package constants

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

const (
	RightSizingLabelName = "rightsizing-service"
)

const (
	DefaultTraceIntervalSeconds = "300"
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
	TraceServiceContainerName  = "trace"
	TraceServiceContainerImage = "docker.io/dbdydgur2244/rightsizing-trace"
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

func TraceContainerTemplate(url, namespace, name string, interval ...string) v1.Container {
	i := DefaultTraceIntervalSeconds
	if len(interval) > 0 {
		i = interval[0]
	}
	container := v1.Container{
		Name:  TraceServiceContainerName,
		Image: TraceServiceContainerImage,
		Command: []string{
			"python",
			"main.py",
		},
		Args: []string{
			"-url", url,
			"-ns", namespace,
			"-n", name,
			"-server_url", fmt.Sprintf("http://127.0.0.1:%s", ServerContainerPort),
			"-i", i,
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
