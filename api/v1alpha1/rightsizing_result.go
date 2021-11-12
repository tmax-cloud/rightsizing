package v1alpha1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rightsizing/constants"
)

type ServiceResult struct {
	Describe string `json:"describe" description:"description of value"`
	// 대표 값 표기
	Values ServiceData `json:"value" description:"value which container result"`
	// 기록된 시각
	RecordedTime v1.Time `json:"recordedTime" description:"time when the result is recorded"`
}

type resultData interface {
	Data() *ServiceData
}

type forecastData struct {
	Datetimes []string  `json:"ds"`
	Yhat      []float32 `json:"yhat"`
	YhatLower []float32 `json:"yhat_lower"`
	YhatUpper []float32 `json:"yhat_upper"`
}

type forecastResult struct {
	CPU    forecastData `json:"cpu"`
	Memory forecastData `json:"memory"`
}

func (r forecastResult) Data() *ServiceData {
	if r.CPU.Yhat != nil && r.Memory.Yhat != nil {
		return &ServiceData{
			CPU:    fmt.Sprintf("%.4f", r.CPU.Yhat[len(r.CPU.Yhat)-1]),
			Memory: fmt.Sprintf("%.4f", r.Memory.Yhat[len(r.Memory.Yhat)-1]),
		}
	}
	return nil
}

type optimizationResult struct {
	CPU    float32 `json:"cpu"`
	Memory float32 `json:"memory"`
}

type ServiceData struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

func (r optimizationResult) Data() *ServiceData {
	return &ServiceData{
		CPU:    fmt.Sprintf("%.4f", r.CPU),
		Memory: fmt.Sprintf("%.4f", r.Memory),
	}
}

func GetPodUrl(pod *corev1.Pod) string {
	if pod.Status.PodIPs == nil {
		return ""
	}
	return fmt.Sprintf("http://%s:%s", pod.Status.PodIPs[0].IP, constants.ServerContainerPort)
}

func GetServiceResult(serviceType ServiceType, pod *corev1.Pod) *ServiceResult {
	url := GetPodUrl(pod)
	if url == "" {
		return nil
	}
	// example: http://10.0.0.1:8000/forecast
	url = fmt.Sprintf("%s/%s", url, serviceType)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result resultData
	switch serviceType {
	case ForecastService:
		result = &forecastResult{}
	case OptimizationService:
		result = &optimizationResult{}
	}
	if err := json.Unmarshal(data, result); err != nil {
		return nil
	}

	if d := result.Data(); d != nil {
		return &ServiceResult{
			Describe:     fmt.Sprintf("result of %s service", serviceType),
			Values:       *d,
			RecordedTime: v1.NewTime(time.Now()),
		}
	}
	return nil
}
