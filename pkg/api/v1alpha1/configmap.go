package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"rightsizing/pkg/constants"
)

const (
	PrometheusConfigKeyName = "prometheus"
	QueryConfigKeyName      = "query"
)

type PrometheusConfig struct {
	DefaultPrometheusUri string `json:"defaultPrometheusUri"`
}

type ResourceQueryConfig struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type RightsizingQueryConfig struct {
	ContainerImage  string              `json:"image"`
	Version         string              `json:"defaultImageVersion" description:"default image tag"`
	ResourceQueries ResourceQueryConfig `json:"resourceQuery" description:"default query of each resources"`
}

type RightsizingConfig struct {
	PrometheusConfig PrometheusConfig
	QueryConfig      RightsizingQueryConfig
}

type RightsizingRequestConfig struct {
	PrometheusConfig PrometheusConfig
	QueryConfig      RightsizingQueryConfig
}

func NewRightsizngConfig(cli client.Client) (*RightsizingConfig, error) {
	configMap := &corev1.ConfigMap{}
	err := cli.Get(context.TODO(), types.NamespacedName{Namespace: constants.RightsizingConfigMapNamespace, Name: constants.RightsizingConfigMapName}, configMap)
	if err != nil {
		return nil, err
	}

	prometheusConfig := PrometheusConfig{}
	if prometheus, ok := configMap.Data[PrometheusConfigKeyName]; ok {
		err := json.Unmarshal([]byte(prometheus), &prometheusConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse prometheus config json: %v", err)
		}
	}

	queryConfig := RightsizingQueryConfig{}
	if query, ok := configMap.Data[QueryConfigKeyName]; ok {
		err := json.Unmarshal([]byte(query), &queryConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse query config json: %v", err)
		}
	}

	return &RightsizingConfig{
		PrometheusConfig: prometheusConfig,
		QueryConfig:      queryConfig,
	}, nil
}

func NewRightsizingRequestConfig(cli client.Client) (*RightsizingRequestConfig, error) {
	configMap := &corev1.ConfigMap{}
	err := cli.Get(context.TODO(), types.NamespacedName{Namespace: constants.RightsizingConfigMapNamespace, Name: constants.RightsizingConfigMapName}, configMap)
	if err != nil {
		return nil, err
	}

	prometheusConfig := PrometheusConfig{}
	if prometheus, ok := configMap.Data[PrometheusConfigKeyName]; ok {
		err := json.Unmarshal([]byte(prometheus), &prometheusConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse prometheus config json: %v", err)
		}
	}

	queryConfig := RightsizingQueryConfig{}
	if query, ok := configMap.Data[QueryConfigKeyName]; ok {
		err := json.Unmarshal([]byte(query), &queryConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse query config json: %v", err)
		}
	}

	return &RightsizingRequestConfig{
		PrometheusConfig: prometheusConfig,
		QueryConfig:      queryConfig,
	}, nil
}
