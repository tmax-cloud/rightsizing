package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/applyconfigurations/core/v1"

	"rightsizing/pkg/constants"
	"rightsizing/pkg/utils"
)

// Status struct
//   status: ?
//   conditions:
//   -
//   results:
//   - query: ~~
//     forecast:
//       value: 1.2424
//     optimization:
//       value: 123.124
// RightsizingRequestStatus defines the observed state of RightsizingRequest
type RightsizingRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status Status `json:"status" description:"status of the object"`
	// 각 service 별 상태 표기
	QueryStatuses map[string]QueryCondition `json:"conditions,omitempty" description:"condition list of each service"`
	// 각 service 결과 표기
	QueryResults map[string]QueryResult `json:"results,omitempty" description:"result list of each service"`
}

// Even if you pass multiple options, just recognize first option.
func (rs *RightsizingRequestStatus) PropagateStatus(request *RightsizingRequest, pod *corev1.Pod) {
	// server container 상태 확인
	serverContainerStatus := utils.GetContainerStatus(pod, constants.ServerContainerName)
	if serverContainerStatus != nil && utils.IsContainerFailed(*serverContainerStatus) {
		rs.Status = ServiceFailed
		return
	}
	// request container 상태 확인
	// containerStatus와 container 순서가 동일하다고 가정
	for i, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name != constants.ServerContainerName {
			query := request.Spec.Queries[i].Query
			queryCondition := GetQueryCondition(query, containerStatus)
			if queryCondition.Status == ConditionCompleted {
				result := GetQueryResult(query, containerStatus.Name, pod)
				rs.addOrUpdateResult(containerStatus.Name, result)
			}
			rs.addOrUpdateCondition(containerStatus.Name, queryCondition)
		}
	}
	rs.setStatus()
}

func (rs *RightsizingRequestStatus) setStatus() {
	var status = ServiceUnknown

	for _, condition := range rs.QueryStatuses {
		// 하나라도 waiting이면 waiting 상태
		if condition.Status == ConditionWaiting {
			status = ServiceWaiting
			break
		}
		if condition.Status == ConditionRunning {
			status = ServiceRunning
		}
	}

	for _, condition := range rs.QueryStatuses {
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

func (rs *RightsizingRequestStatus) addOrUpdateCondition(containerName string, condition QueryCondition) {
	if rs.QueryStatuses == nil {
		rs.QueryStatuses = make(map[string]QueryCondition)
	}

	if existCondition, exist := rs.QueryStatuses[containerName]; exist {
		if !QueryConditionEquality(existCondition, condition) {
			rs.QueryStatuses[containerName] = condition
		}
	} else {
		rs.QueryStatuses[containerName] = condition
	}
}

func (rs *RightsizingRequestStatus) addOrUpdateResult(containerName string, result *QueryResult) {
	if rs.QueryResults == nil {
		rs.QueryResults = make(map[string]QueryResult)
	}

	if result != nil {
		if _, exist := rs.QueryResults[containerName]; !exist {
			rs.QueryResults[containerName] = *result
		}
	}
}

func (rs *RightsizingRequestStatus) isCompleted() bool {
	if len(rs.QueryResults) != len(rs.QueryStatuses) {
		return false
	}

	completedQueryCnt := 0
	for _, query := range rs.QueryStatuses {
		if query.Status == ConditionCompleted {
			completedQueryCnt++
		}
	}

	if completedQueryCnt == 0 {
		return false
	}
	return true
}
