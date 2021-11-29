package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"

	"rightsizing/pkg/constants"
	"rightsizing/pkg/utils"
)

type Status string

const (
	// Unknown: pod의 상태를 모르는 상태 (pod이 실행전, pod을 GET하지 못한 경우)
	ServiceUnknown Status = "Unknown"
	ServiceWaiting Status = "Waiting"
	// Running: Rightsizing service 수행중인 상황
	ServiceRunning Status = "Running"
	// Complete: Rightsizing service 수행을 완료하여 서비스 내역에 해당 정보를 기록한 상황 (Service 준비 상태 (ready 보다는 완성되었다는 느낌을 주기위해서))
	ServiceCompleted Status = "Complete"
	// Error: Rightsizing service 수행 중 에러가 발생한 상황
	ServiceFailed Status = "Failed"
)

type RightsizingStatus struct {
	// Rightsizing overall status
	Status Status `json:"status" description:"status of the object"`
	// 각 service 별 상태 표기
	ServiceStatuses map[string]QueryCondition `json:"conditions,omitempty" description:"condition list of the services (e.g. forecast, optimization)"`
	// 각 service 결과 표기
	ServiceResults map[string]QueryResult `json:"results,omitempty" description:"result list of the services (e.g. forecast, optimization)"`
}

func (rs *RightsizingStatus) PropagateStatus(pod *corev1.Pod) {
	// server container 상태 확인
	serverContainerStatus := utils.GetContainerStatus(pod, constants.ServerContainerName)
	if serverContainerStatus != nil && utils.IsContainerFailed(*serverContainerStatus) {
		rs.Status = ServiceFailed
		return
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name != constants.ServerContainerName {
			// Rightsizing의 경우 container name이 resource 관련으로 가서 cpu, memory 이런식으로 됨.
			queryCondition := GetQueryCondition(containerStatus.Name, containerStatus)
			// update results
			if queryCondition.Status == ConditionCompleted {
				result := GetQueryResult(containerStatus.Name, containerStatus.Name, pod)
				rs.addOrUpdateResult(containerStatus.Name, result)
			}
			rs.addOrUpdateCondition(containerStatus.Name, queryCondition)
		}
	}
	rs.setStatus()
}

func (rs *RightsizingStatus) addOrUpdateCondition(containerName string, condition QueryCondition) {
	if rs.ServiceStatuses == nil {
		rs.ServiceStatuses = make(map[string]QueryCondition)
	}

	if existCondition, exist := rs.ServiceStatuses[containerName]; exist {
		if !QueryConditionEquality(existCondition, condition) {
			rs.ServiceStatuses[containerName] = condition
		}
	} else {
		rs.ServiceStatuses[containerName] = condition
	}
}

func (rs *RightsizingStatus) addOrUpdateResult(containerName string, result *QueryResult) {
	if rs.ServiceResults == nil {
		rs.ServiceResults = make(map[string]QueryResult)
	}

	if result != nil {
		if _, exist := rs.ServiceResults[containerName]; !exist {
			rs.ServiceResults[containerName] = *result
		}
	}
}

func (rs *RightsizingStatus) setStatus() {
	var status = ServiceUnknown

	for _, condition := range rs.ServiceStatuses {
		// 하나라도 waiting이면 waiting 상태
		if condition.Status == ConditionWaiting {
			status = ServiceWaiting
			break
		}
		if condition.Status == ConditionRunning {
			status = ServiceRunning
		}
	}

	for _, condition := range rs.ServiceStatuses {
		if condition.Status == ConditionFailed {
			status = ServiceFailed
			break
		}
	}

	if rs.isCompleted() {
		status = ServiceCompleted
	}
	rs.Status = status
}

func (rs *RightsizingStatus) isCompleted() bool {
	if len(rs.ServiceResults) != len(rs.ServiceStatuses) {
		return false
	}

	completedQueryCnt := 0
	for _, query := range rs.ServiceStatuses {
		if query.Status == ConditionCompleted {
			completedQueryCnt++
		}
	}

	if completedQueryCnt == 0 {
		return false
	}
	return true
}
