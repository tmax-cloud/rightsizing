package v1alpha1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"rightsizing/pkg/constants"
)

// Query struct
// Queries:
//   queryName1:
//     forecast:
//       value: 1.2424
//     optimization:
//       value: 123.124
type QueryResult struct {
	Query string `json:"query" description:"description of value"`
	// 대표 값 표기 (float 같은 데이터는 openapi 변환이 안되 string으로 저장)
	Data ResultData `json:"data" description:"result data of service"`
	// 기록된 시각
	RecordedTime v1.Time `json:"recordedTime" description:"time when the result is recorded"`
}

type ForecastResult struct {
	Datetime  string  `json:"ds"`
	Yhat      float32 `json:"yhat"`
	YhatLower float32 `json:"yhat_lower"`
	YhatUpper float32 `json:"yhat_upper"`
}

type OptimizationResult struct {
	Value float32 `json:"data"`
}

type ResultData struct {
	ForecastResult     *string `json:"forecast,omitempty"`
	OptimizationResult *string `json:"optimization,omitempty"`
}

type ResultItem struct {
	ForecastResult     *ForecastResult     `json:"forecast,omitempty"`
	OptimizationResult *OptimizationResult `json:"optimization,omitempty"`
}

func (r ResultItem) Data() ResultData {
	var forecast *string
	var optimization *string

	if r.ForecastResult != nil {
		forecast = new(string)
		*forecast = fmt.Sprintf("%g", r.ForecastResult.Yhat)
	}
	if r.OptimizationResult != nil {
		optimization = new(string)
		*optimization = fmt.Sprintf("%g", r.OptimizationResult.Value)
	}
	return ResultData{
		ForecastResult:     forecast,
		OptimizationResult: optimization,
	}
}

func GetPodUri(pod *corev1.Pod) string {
	if pod.Status.PodIP == "" {
		return ""
	}
	return fmt.Sprintf("http://%s:%s", pod.Status.PodIP, constants.ServerContainerPort)
}

func GetQueryResult(query, containerName string, pod *corev1.Pod) *QueryResult {
	url := GetPodUri(pod)
	if url == "" {
		return nil
	}
	// example: http://10.0.0.1:8000/queries/forecast, http://10.0.0.1:8000/queries/query-1
	url = fmt.Sprintf("%s/%s/%s", url, constants.QueryEndpoint, containerName)

	client := http.Client{
		Timeout: time.Second * 2,
	}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result ResultItem
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return &QueryResult{
		Query:        query,
		Data:         result.Data(),
		RecordedTime: v1.NewTime(time.Now()),
	}
}
