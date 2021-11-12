package utils

import (
	v1 "k8s.io/api/core/v1"
)

func IsInclude(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func RemoveItem(slice []string, target string) (result []string) {
	for _, v := range slice {
		if v == target {
			continue
		}
		result = append(result, v)
	}
	return
}

func IsPodReadyCondition(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsPodContainerReady(pod *v1.Pod, containerName string) bool {
	if pod == nil {
		return false
	}

	for _, condition := range pod.Status.ContainerStatuses {
		if condition.Name == containerName {
			if condition.Ready {
				return true
			}
			break
		}
	}
	return false
}

func IsPodContainerSucceeded(pod *v1.Pod, containerName string) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			return IsContainerSucceeded(status)
		}
	}
	return false
}

func IsContainerRunning(status v1.ContainerStatus) bool {
	if status.Ready && status.State.Running != nil {
		return true
	}
	return false
}

func IsContainerSucceeded(status v1.ContainerStatus) bool {
	if !status.Ready && status.State.Terminated != nil {
		if status.State.Terminated.ExitCode == 0 && status.State.Terminated.Reason != "Error" {
			return true
		}
	}
	return false
}

func IsContainerFailed(status v1.ContainerStatus) bool {
	if status.State.Terminated != nil {
		if status.State.Terminated.Reason == "Error" {
			return true
		}
	}
	return false
}

func GetContainerStatus(pod *v1.Pod, containerName string) *v1.ContainerStatus {
	if pod == nil {
		return nil
	}

	for _, condition := range pod.Status.ContainerStatuses {
		if condition.Name == containerName {
			return &condition
		}
	}
	return nil
}
