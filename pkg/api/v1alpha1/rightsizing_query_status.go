package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rightsizing/pkg/utils"
)

type Condition string

const (
	// Unknown: pod의 상태를 모르는 상태 (pod이 실행전, pod을 GET하지 못한 경우)
	ConditionUnknown Condition = "Unknown"
	ConditionWaiting Condition = "Waiting"
	// Running: Rightsizing service 수행중인 상황
	ConditionRunning Condition = "Running"
	// Complete: Rightsizing service 수행을 완료하여 서비스 내역에 해당 정보를 기록한 상황 (Service 준비 상태 (ready 보다는 완성되었다는 느낌을 주기위해서))
	ConditionCompleted Condition = "Complete"
	// Error: Rightsizing service 수행 중 에러가 발생한 상황
	ConditionFailed Condition = "Failed"
)

type QueryCondition struct {
	Query   string    `json:"query"`
	Status  Condition `json:"status"`
	URL     *string   `json:"url,omitempty"`
	Reason  *string   `json:"reason,omitempty"  description:"one-word CamelCase reason for the condition's last transition"`
	Message *string   `json:"message,omitempty" description:"human-readable message indicating details about last transition"`
	//
	LastTransitionTime v1.Time `json:"lastTransitionTime,omitempty" description:"last time the condition transit from one status to another"`
}

type QueryConditionOption struct {
	Url     string
	Reason  string
	Message string
}

// Even if you pass multiple options, just recognize first option.
func NewQueryCondition(condition Condition, query string, opt ...QueryConditionOption) QueryCondition {
	var reason *string
	var message *string
	var url *string

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

	return QueryCondition{
		Query:              query,
		Status:             condition,
		Reason:             reason,
		Message:            message,
		URL:                url,
		LastTransitionTime: v1.NewTime(time.Now()),
	}
}

func GetQueryCondition(query string, containerStatus corev1.ContainerStatus) QueryCondition {
	var condition Condition
	var opt QueryConditionOption

	if utils.IsContainerFailed(containerStatus) {
		condition = ConditionFailed
		opt = QueryConditionOption{
			Reason:  containerStatus.State.Terminated.Reason,
			Message: containerStatus.State.Terminated.Message,
		}
	} else if utils.IsContainerRunning(containerStatus) {
		condition = ConditionRunning
		opt = QueryConditionOption{
			Reason:  "Running",
			Message: "Start to query",
		}
	} else if utils.IsContainerSucceeded(containerStatus) {
		condition = ConditionCompleted
		opt = QueryConditionOption{
			Reason:  "Succeeded",
			Message: "successfully analyze",
		}
	} else {
		if containerStatus.State.Waiting != nil {
			condition = ConditionWaiting
			opt = QueryConditionOption{
				Reason:  containerStatus.State.Waiting.Reason,
				Message: containerStatus.State.Waiting.Message,
			}
		} else {
			condition = ConditionUnknown
			opt = QueryConditionOption{
				Reason:  "Unknown",
				Message: "can't recognize container status",
			}
		}
	}
	return NewQueryCondition(condition, query, opt)
}

func QueryConditionEquality(c1, c2 QueryCondition) bool {
	if c1.Query != c2.Query {
		return false
	}
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
