package v1alpha1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RequestServiceResult struct {
	Describe string `json:"describe" description:"description of value"`
	// 대표 값 표기
	Data RequestServiceData `json:"data" description:"result data of service"`
	// 기록된 시각
	RecordedTime v1.Time `json:"recordedTime" description:"time when the result is recorded"`
}

type requestData interface {
	Data() RequestServiceData
}

type RequestServiceData struct {
	Value string `json:"value"`
}

type forecastRequest struct {
	forecastData
}

func (r forecastRequest) Data() RequestServiceData {
	return RequestServiceData{Value: fmt.Sprintf("%f", r.Yhat)}
}

type optimizationRequest struct {
	Value float32 `json:"data"`
}

func (r optimizationRequest) Data() RequestServiceData {
	return RequestServiceData{fmt.Sprintf("%f", r.Value)}
}

func GetRequestServiceResult(serviceType ServiceType, pod *corev1.Pod) *RequestServiceResult {
	url := GetPodUrl(pod)
	if url == "" {
		return nil
	}
	// example: http://10.0.0.1:8000/forecast
	url = fmt.Sprintf("%s/%s", url, serviceType)

	var result requestData
	switch serviceType {
	case ForecastService:
		result = &forecastRequest{}
		url = url + "/summary"
	case OptimizationService:
		result = &optimizationRequest{}
	}

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil
	}

	d := result.Data()
	return &RequestServiceResult{
		Data:         d,
		RecordedTime: v1.NewTime(time.Now()),
	}
}
