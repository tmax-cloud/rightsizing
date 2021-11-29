/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rightsizingrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"rightsizing/pkg/api/v1alpha1"
	"rightsizing/pkg/constants"
	"rightsizing/pkg/utils"
)

// RightsizingRequestReconciler reconciles a RightsizingRequest object
type RightsizingRequestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

const (
	finalizerName = "rightsizingrequest.finalizers"
)

// +kubebuilder:rbac:groups=rightsizing.tmax.io,resources=rightsizingrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rightsizing.tmax.io,resources=rightsizingrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rightsizing.tmax.io,resources=rightsizingrequests/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RightsizingRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RightsizingRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Rightsizing service instance 가져옴
	rsreq := &v1alpha1.RightsizingRequest{}
	if err := r.Get(ctx, req.NamespacedName, rsreq); err != nil {
		// 삭제 (또는 변경사항)이 생긴 object
		if apierr.IsNotFound(err) {
			r.Log.Error(err, "Failed to get instance", "apiVersion", rsreq.APIVersion, "rsreq", rsreq.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// // 반드시 필요한 부분은 아님
	// // DeletionTimestamp 확인해서 deletion 상태인지 확인
	if rsreq.ObjectMeta.DeletionTimestamp.IsZero() {
		// 삭제되지 않은 경우 controller에서 관리하기 위해서 finalizer 마킹
		if !utils.IsInclude(rsreq.ObjectMeta.Finalizers, finalizerName) {
			rsreq.ObjectMeta.Finalizers = append(rsreq.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.TODO(), rsreq); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// 삭제된 object가 실제 삭제될 수 있도록 finalizer 삭제
		if utils.IsInclude(rsreq.ObjectMeta.Finalizers, finalizerName) {
			if err := r.deleteResources(rsreq); err != nil {
				return ctrl.Result{}, err
			}
			// finalizer 제거해서 kubernetes에서 삭제될 수 있도록 함
			rsreq.ObjectMeta.Finalizers = utils.RemoveItem(rsreq.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.TODO(), rsreq); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	r.Log.Info("Reconciling rightsizingrequest service", "apiVersion", rsreq.APIVersion, "rzreq", rsreq.Name)
	// 존재하는지, 업데이트가 필요한 오브젝트가 있는지 확인
	checkResult, pod, err := r.checkPodExist(rsreq)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Pod 생성이 필요한 경우
	if checkResult == constants.CheckResultCreate {
		ctrl.SetControllerReference(rsreq, pod, r.Scheme)
		if err := r.Create(context.Background(), pod); err != nil {
			r.Log.Error(err, "Failed to create new pod", "apiVersion", pod.APIVersion, "pod", pod.Name)
			r.Recorder.Eventf(rsreq, v1.EventTypeWarning, "InternalError", err.Error())
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(rsreq, v1.EventTypeNormal, "CreatePod", "Successfully create pod")
	}
	// Rightsizing status 업데이트
	rsreq.Status.PropagateStatus(rsreq, pod)
	// 업데이트한 내용 반영하여 실제 상태 업데이트
	if err := r.updateStatus(rsreq); err != nil {
		r.Recorder.Eventf(rsreq, v1.EventTypeWarning, "InternalError", err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RightsizingRequestReconciler) RightsizingRequestPodTemplate(request *v1alpha1.RightsizingRequest) (*v1.Pod, error) {
	pod := constants.RightsizingRequestDefaultPodTemplate(request.ObjectMeta)

	config, err := v1alpha1.NewRightsizingRequestConfig(r.Client)
	if err != nil {
		return nil, err
	}

	url := config.PrometheusConfig.DefaultPrometheusUri
	if request.Spec.PrometheusUri != nil && *request.Spec.PrometheusUri != "" {
		url = *request.Spec.PrometheusUri
	}

	for i, query := range request.Spec.Queries {
		containerName := fmt.Sprintf("%s-%d", "query", i)
		// generate real query
		labels := make([]string, 0)
		for key, value := range query.Labels {
			labels = append(labels, fmt.Sprintf(`%s="%s"`, key, value))
		}
		fullQuery := fmt.Sprintf("%s{%s}", query.Query, strings.Join(labels, ","))

		container := v1.Container{
			Name:  containerName,
			Image: fmt.Sprintf("%s:%s", config.QueryConfig.ContainerImage, config.QueryConfig.Version),
			Args: []string{
				"python", "main.py",
				"--url", url,
				"-q", fullQuery,
				"-server_url", fmt.Sprintf("http://127.0.0.1:%s", constants.ServerContainerPort),
			},
			Env: []v1.EnvVar{
				{
					Name:  "CONTAINER_NAME",
					Value: containerName,
				},
			},
		}
		if doForecast := utils.ExtractBooleanValue(request.Spec.Forecast, query.Forecast); doForecast {
			container.Args = append(container.Args, "--forecast")
		}
		if doOptimization := utils.ExtractBooleanValue(request.Spec.Optimization, query.Optimization); doOptimization {
			container.Args = append(container.Args, "--optimization")
		}
		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}
	pod.Spec.Containers = append(pod.Spec.Containers, constants.ServerContainerTemplate())
	return &pod, nil
}

func (r *RightsizingRequestReconciler) checkPodExist(request *v1alpha1.RightsizingRequest) (constants.CheckResultType, *v1.Pod, error) {
	found := constants.RightsizingRequestDefaultPodTemplate(request.ObjectMeta)
	// 존재하지 않는 경우
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: found.Namespace, Name: found.Name}, &found); err != nil {
		if apierr.IsNotFound(err) {
			pod, err := r.RightsizingRequestPodTemplate(request)
			if err != nil {
				return constants.CheckResultError, nil, err
			}
			return constants.CheckResultCreate, pod, nil
		} else {
			r.Log.Error(err, "Failed to get pod", "Namespace", found.Namespace, "Name", found.Name)
			return constants.CheckResultError, nil, err
		}
	}
	return constants.CheckResultExisted, &found, nil
}

func (r *RightsizingRequestReconciler) updateStatus(desiredObject *v1alpha1.RightsizingRequest) error {
	existingObject := &v1alpha1.RightsizingRequest{}
	namespacedName := types.NamespacedName{Namespace: desiredObject.Namespace, Name: desiredObject.Name}
	if err := r.Get(context.TODO(), namespacedName, existingObject); err != nil {
		return err
	}

	if !equality.Semantic.DeepEqual(existingObject.Status, desiredObject.Status) {
		r.Recorder.Eventf(desiredObject, v1.EventTypeNormal, "UpdateStatus", fmt.Sprintf("Update state [%s]", desiredObject.Status.Status))
		if err := r.Status().Update(context.TODO(), desiredObject); err != nil {
			return err
		}
	}
	return nil
}

func (r *RightsizingRequestReconciler) deleteResources(request *v1alpha1.RightsizingRequest) error {
	found := constants.RightsizingRequestDefaultPodTemplate(request.ObjectMeta)
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: found.Namespace, Name: found.Name}, &found)
	if err != nil {
		r.Log.Error(err, "Failed to find old pod", "pod", found.Name)
		return nil
	}
	if err := r.Delete(context.TODO(), &found, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
		r.Log.Error(err, "Failed to delete pod", "pod", found.Name)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RightsizingRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RightsizingRequest{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
