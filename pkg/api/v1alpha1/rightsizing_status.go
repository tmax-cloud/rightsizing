package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"

	"rightsizing/pkg/constants"
)

type RightsizingStatus struct {
	// Rightsizing overall status
	Status Status `json:"status" description:"status of the object"`
	// 각 service 별 상태 표기
	ServiceStatuses map[ServiceType]ServiceStatus `json:"conditions,omitempty" description:"condition list of the services (e.g. forecast, optimization)"`
	// 각 service 결과 표기
	ServiceResults map[ServiceType]ServiceResult `json:"results,omitempty" description:"result list of the services (e.g. forecast, optimization)"`
}

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
		rs.SetServiceCondition(status)
		if status.Status == ConditionCompleted {
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
			rs.SetServiceCondition(condition)
		}
	}
}

func (rs *RightsizingStatus) SetServiceCondition(status ServiceStatus) {
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
	if result != nil {
		if rs.ServiceResults == nil {
			rs.ServiceResults = make(map[ServiceType]ServiceResult)
		}
		rs.ServiceResults[serviceType] = *result
	}
}
