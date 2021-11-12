package v1alpha1

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rightsizing/constants"
	"rightsizing/utils"
)

type RightsizingStatus struct {
	// Rightsizing overall status
	Status Status `json:"status" description:"status of the object"`
	// 각 service 별 상태 표기
	ServiceStatuses map[ServiceType]ServiceStatus `json:"conditions,omitempty" description:"condition list of the services (e.g. forecast, optimization)"`
	// 각 service 결과 표기
	ServiceResults map[ServiceType]ServiceResult `json:"results,omitempty" description:"result list of the services (e.g. forecast, optimization)"`
}

type Status string

const (
	// Unknown: pod의 상태를 모르는 상태 (pod이 실행전, pod을 GET하지 못한 경우)
	ConditionUnknown Status = "Unknown"
	// Running: Rightsizing service 수행중인 상황
	ConditionRunning Status = "Running"
	// Complete: Rightsizing service 수행을 완료하여 서비스 내역에 해당 정보를 기록한 상황 (Service 준비 상태 (ready 보다는 완성되었다는 느낌을 주기위해서))
	ConditionComplete Status = "Complete"
	// Error: Rightsizing service 수행 중 에러가 발생한 상황
	ConditionFailed Status = "Failed"
)

type ServiceType string

const (
	ForecastService     = "forecast"
	OptimizationService = "optimization"
	// 모르는 서비스.... 문제 있는 상황
	UnknownService = "unknown"
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
	default:
		serviceType = UnknownService
	}
	return
}

func GetServiceContainerStatuses(pod *corev1.Pod) (condition Status, serviceStatuses []ServiceStatus) {
	condition = ConditionComplete // 하나라도 complete 아니면 ServiceComplete 상태 X

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
				serviceStatuses = append(serviceStatuses, NewCondition(serviceType, ConditionComplete, opt))
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

// Even if you pass multiple options, just recognize first option.
func (rs *RightsizingStatus) PropagateStatus(resultType constants.CheckResultType, pod *corev1.Pod) {
	var condition Status
	var serviceStatuses []ServiceStatus

	if resultType == constants.CheckResultCreate {
		condition = ConditionUnknown
	} else { // constants.CheckResultExist
		condition, serviceStatuses = GetServiceContainerStatuses(pod)
	}

	rs.Status = condition
	for _, status := range serviceStatuses {
		rs.SetCondition(status)
		if status.Status == ConditionComplete {
			if _, exist := rs.ServiceResults[status.Type]; !exist {
				result := GetServiceResult(status.Type, pod)
				rs.SetResult(status.Type, result)
			}
		}
	}
}

func (rs *RightsizingStatus) SetStatus(condition Status, serviceStatuses ...ServiceStatus) {
	if rs.Status == condition {
		return
	}
	if len(serviceStatuses) > 0 {
		for _, condition := range serviceStatuses {
			rs.SetCondition(condition)
		}
	}
}

func (rs *RightsizingStatus) SetCondition(status ServiceStatus) {
	if rs.ServiceStatuses == nil {
		rs.ServiceStatuses = make(map[ServiceType]ServiceStatus)
	}
	if existServiceStatus, exist := rs.ServiceStatuses[status.Type]; exist {
		if !ServiceStatusEquality(existServiceStatus, status) {
			rs.ServiceStatuses[status.Type] = status
		}
	} else {
		rs.ServiceStatuses[status.Type] = status
	}
}

func (rs *RightsizingStatus) SetResult(serviceType ServiceType, result *ServiceResult) {
	if rs.ServiceResults == nil {
		rs.ServiceResults = make(map[ServiceType]ServiceResult)
	}
	if result != nil {
		rs.ServiceResults[serviceType] = *result
	}
}
