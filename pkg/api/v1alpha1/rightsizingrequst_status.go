package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"

	"rightsizing/pkg/constants"
	"rightsizing/pkg/utils"
)

// RightsizingRequestStatus defines the observed state of RightsizingRequest
type RightsizingRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status Status `json:"status" description:"status of the object"`
	// 각 service 별 상태 표기
	ServiceStatuses map[ServiceType]ServiceStatus `json:"conditions,omitempty" description:"condition list of each service"`
	// 각 service 결과 표기
	ServiceResults map[ServiceType]RequestServiceResult `json:"queryResults,omitempty" description:"result list of each service"`
}

// Even if you pass multiple options, just recognize first option.
func (rs *RightsizingRequestStatus) PropagateStatus(pod *corev1.Pod) {
	var condition = ConditionUnknown
	var serviceStatuses []ServiceStatus
	// server container 상태 확인
	serverContainerStatus := utils.GetContainerStatus(pod, constants.RequestServerContainerName)
	if serverContainerStatus != nil && utils.IsContainerFailed(*serverContainerStatus) {
		condition = ConditionFailed
		serviceStatuses = append(serviceStatuses, NewCondition(ServerService, ConditionFailed, ServiceStatusOption{
			Reason:  serverContainerStatus.State.Terminated.Reason,
			Message: serverContainerStatus.State.Terminated.Message,
		}))
		rs.Status = condition
		return
	}
	// request container 상태 확인
	serviceContainerStatus := utils.GetContainerStatus(pod, constants.RequestServiceContainerName)
	if serviceContainerStatus != nil {
		serviceStatus := GetContainerStatus(RequestService, *serviceContainerStatus)
		// update results
		switch serviceStatus.Status {
		case ConditionWaiting, ConditionRunning:
			condition = serviceStatus.Status
		case ConditionCompleted:
			for _, serviceName := range constants.ServiceList {
				serviceType := GetServiceType(serviceName)
				if _, exist := rs.ServiceResults[serviceType]; !exist {
					result := GetRequestServiceResult(serviceType, pod)
					rs.SetResult(serviceType, result)
				}
			}
		}
		serviceStatuses = append(serviceStatuses, serviceStatus)
		condition = serviceStatus.Status
	}

	rs.Status = condition
	// update condition of each service
	for _, status := range serviceStatuses {
		rs.SetServiceCondition(status)
	}
}

func (rs *RightsizingRequestStatus) SetServiceCondition(status ServiceStatus) {
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

func (rs *RightsizingRequestStatus) SetResult(serviceType ServiceType, result *RequestServiceResult) {
	if result != nil {
		if rs.ServiceResults == nil {
			rs.ServiceResults = make(map[ServiceType]RequestServiceResult)
		}
		rs.ServiceResults[serviceType] = *result
	}
}
