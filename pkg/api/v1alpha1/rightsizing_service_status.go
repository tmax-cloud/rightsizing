package v1alpha1

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rightsizing/pkg/constants"
	"rightsizing/pkg/utils"
)

type Status string

const (
	// Unknown: pod의 상태를 모르는 상태 (pod이 실행전, pod을 GET하지 못한 경우)
	ConditionUnknown Status = "Unknown"
	ConditionWaiting Status = "Waiting"
	// Running: Rightsizing service 수행중인 상황
	ConditionRunning Status = "Running"
	// Complete: Rightsizing service 수행을 완료하여 서비스 내역에 해당 정보를 기록한 상황 (Service 준비 상태 (ready 보다는 완성되었다는 느낌을 주기위해서))
	ConditionCompleted Status = "Complete"
	// Error: Rightsizing service 수행 중 에러가 발생한 상황
	ConditionFailed Status = "Failed"
)

type ServiceType string

const (
	ForecastService     ServiceType = "forecast"
	OptimizationService ServiceType = "optimization"
	ServerService       ServiceType = "server"
	RequestService      ServiceType = "request"
	// 모르는 서비스.... 문제 있는 상황
	UnknownService ServiceType = "unknownservice"
)

type ServiceStatus struct {
	Type    ServiceType `json:"type"`
	Status  Status      `json:"state"`
	URL     *string     `json:"url,omitempty"`
	Reason  *string     `json:"reason" description:"one-word CamelCase reason for the condition's last transition"`
	Message *string     `json:"message" description:"human-readable message indicating details about last transition"`
	//
	LastTransitionTime v1.Time `json:"lastTransitionTime,omitempty" description:"last time the condition transit from one status to another"`
}

type ServiceStatusOption struct {
	Url     string
	Reason  string
	Message string
}

// Even if you pass multiple options, just recognize first option.
func NewCondition(serviceType ServiceType, status Status, opt ...ServiceStatusOption) ServiceStatus {
	var reason *string = nil
	var message *string = nil
	var url *string = nil

	if len(opt) > 0 {
		reason = new(string)
		message = new(string)
		*reason = opt[0].Reason
		*message = opt[0].Message
		if opt[0].Url != "" {
			url = new(string)
			*url = opt[0].Url
		}
	}

	return ServiceStatus{
		Type:               serviceType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		URL:                url,
		LastTransitionTime: v1.NewTime(time.Now()),
	}
}

func GetServiceType(containerName string) (serviceType ServiceType) {
	switch containerName {
	case constants.OptimizationServiceContainerName:
		serviceType = OptimizationService
	case constants.ForecastServiceContainerName:
		serviceType = ForecastService
	case constants.RequestServiceContainerName:
		serviceType = RequestService
	default:
		serviceType = UnknownService
	}
	return
}

func GetContainerStatus(serviceType ServiceType, containerStatus corev1.ContainerStatus) ServiceStatus {
	var status Status
	var opt ServiceStatusOption

	if utils.IsContainerFailed(containerStatus) {
		status = ConditionFailed
		opt = ServiceStatusOption{
			Reason:  containerStatus.State.Terminated.Reason,
			Message: containerStatus.State.Terminated.Message,
		}
	} else if utils.IsContainerRunning(containerStatus) {
		status = ConditionRunning
		opt = ServiceStatusOption{
			Reason:  "Running",
			Message: "",
		}
	} else if utils.IsContainerSucceeded(containerStatus) {
		status = ConditionCompleted
		opt = ServiceStatusOption{
			Reason:  "Succeeded",
			Message: "",
		}
	} else {
		if containerStatus.State.Waiting != nil {
			status = ConditionWaiting
			opt = ServiceStatusOption{
				Reason:  containerStatus.State.Waiting.Reason,
				Message: containerStatus.State.Waiting.Message,
			}
		} else {
			status = ConditionUnknown
			opt = ServiceStatusOption{
				Reason:  "Unknown",
				Message: "can't recognize container status",
			}
		}
	}
	return NewCondition(serviceType, status, opt)
}

func GetServiceContainerStatuses(pod *corev1.Pod) (condition Status, serviceStatuses []ServiceStatus) {
	condition = ConditionWaiting // 하나라도 complete 아니면 ServiceComplete 상태 X

	for _, containerName := range constants.ServiceContainerList {
		serviceType := GetServiceType(containerName)
		if serviceType == UnknownService {
			continue
		}
		// ServiceType에 존재하는 container들
		if status := utils.GetContainerStatus(pod, containerName); status != nil {
			if utils.IsContainerFailed(*status) {
				condition = ConditionFailed
				opt := ServiceStatusOption{
					Reason:  status.State.Terminated.Reason,
					Message: status.State.Terminated.Message,
				}
				serviceStatuses = append(serviceStatuses, NewCondition(serviceType, ConditionFailed, opt))
			} else if utils.IsContainerRunning(*status) {
				condition = ConditionRunning
				opt := ServiceStatusOption{
					Reason:  fmt.Sprintf("%s%s", containerName, "Running"),
					Message: "",
				}
				serviceStatuses = append(serviceStatuses, NewCondition(serviceType, ConditionRunning, opt))
			} else if utils.IsContainerSucceeded(*status) {
				opt := ServiceStatusOption{
					Reason:  fmt.Sprintf("%s%s", containerName, "Succeeded"),
					Message: "",
				}
				serviceStatuses = append(serviceStatuses, NewCondition(serviceType, ConditionCompleted, opt))
			} else {
				condition = ConditionUnknown
				if status.State.Waiting != nil {
					opt := ServiceStatusOption{
						Reason:  status.State.Waiting.Reason,
						Message: status.State.Waiting.Message,
					}
					serviceStatuses = append(serviceStatuses, NewCondition(serviceType, ConditionUnknown, opt))
				} else {
					opt := ServiceStatusOption{
						Reason:  fmt.Sprintf("%s%s", containerName, "Unknown"),
						Message: fmt.Sprintf("can't recognize %s container status", containerName),
					}
					serviceStatuses = append(serviceStatuses, NewCondition(serviceType, ConditionUnknown, opt))
				}
			}
		}
	}

	anyServiceNotCompleted := false
	for _, status := range serviceStatuses {
		if status.Status != ConditionCompleted {
			anyServiceNotCompleted = true
		}
	}
	if !anyServiceNotCompleted {
		condition = ConditionCompleted
	}
	return
}

func ServiceStatusEquality(c1, c2 ServiceStatus) bool {
	if c1.Status != c2.Status {
		return false
	}
	if c1.Reason != nil && c2.Reason != nil {
		if *c1.Reason != *c2.Reason {
			return false
		}
	} else if c1.Reason != c2.Reason {
		return false
	}
	if c1.Message != nil && c2.Message != nil {
		if *c1.Message != *c2.Message {
			return false
		}
	} else if c1.Reason != c2.Reason {
		return false
	}
	return true
}
