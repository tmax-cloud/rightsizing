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

package controllers

import (
	"context"
	"fmt"

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

	"rightsizing/api/v1alpha1"
	"rightsizing/constants"
	"rightsizing/utils"
)

const (
	finalizerName = "rightsizing.finalizers"
)

// RightsizingReconciler reconciles a Rightsizing object
type RightsizingReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=rightsizing.tmax.io,resources=rightsizings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rightsizing.tmax.io,resources=rightsizings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rightsizing.tmax.io,resources=rightsizings/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Rightsizing object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RightsizingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Rightsizing service instance 가져옴
	rs := &v1alpha1.Rightsizing{}
	if err := r.Get(ctx, req.NamespacedName, rs); err != nil {
		// 삭제 (또는 변경사항)이 생긴 object
		if apierr.IsNotFound(err) {
			// r.Log.Error(err, "Failed to get instance", "apiVersion", rs.APIVersion, "rsvc", rs.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// // 반드시 필요한 부분은 아님
	// // DeletionTimestamp 확인해서 deletion 상태인지 확인
	if rs.ObjectMeta.DeletionTimestamp.IsZero() {
		// 삭제되지 않은 경우 controller에서 관리하기 위해서 finalizer 마킹
		if !utils.IsInclude(rs.ObjectMeta.Finalizers, finalizerName) {
			rs.ObjectMeta.Finalizers = append(rs.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.TODO(), rs); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// 삭제된 object가 실제 삭제될 수 있도록 finalizer 삭제
		if utils.IsInclude(rs.ObjectMeta.Finalizers, finalizerName) {
			if err := r.deleteResources(rs); err != nil {
				return ctrl.Result{}, err
			}
			// finalizer 제거해서 kubernetes에서 삭제될 수 있도록 함
			rs.ObjectMeta.Finalizers = utils.RemoveItem(rs.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.TODO(), rs); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	r.Log.Info("Reconciling rightsizing service", "apiVersion", rs.APIVersion, "rsvc", rs.Name)
	// 존재하는지, 업데이트가 필요한 오브젝트가 있는지 확인
	checkResult, pod, err := r.checkPodExist(rs)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Pod 생성이 필요한 경우
	if checkResult == constants.CheckResultCreate {
		ctrl.SetControllerReference(rs, pod, r.Scheme)
		if err := r.Create(context.TODO(), pod); err != nil {
			r.Log.Error(err, "Failed to create new pod", "apiVersion", pod.APIVersion, "pod", pod.Name)
			r.Recorder.Eventf(rs, v1.EventTypeWarning, "InternalError", err.Error())
			return ctrl.Result{}, err
		}
	}
	// Rightsizing status 업데이트
	rs.Status.PropagateStatus(checkResult, pod)
	r.Log.Info("Rightsizing status", "status", rs.Status)
	// 업데이트한 내용 반영하여 실제 상태 업데이트
	if err := r.updateStatus(rs); err != nil {
		r.Recorder.Eventf(rs, v1.EventTypeWarning, "InternalError", err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RightsizingReconciler) checkPodExist(service *v1alpha1.Rightsizing) (constants.CheckResultType, *v1.Pod, error) {
	found := &v1.Pod{}
	pod := PodTemplate(service)

	err := r.Get(context.TODO(), types.NamespacedName{Namespace: service.Namespace, Name: service.Name}, found)
	// 존재하지 않는 경우
	if err != nil {
		if apierr.IsNotFound(err) {
			return constants.CheckResultCreate, pod, nil
		} else {
			r.Log.Error(err, "Failed to get pod", "Namespace", service.Namespace, "Name", service.Name)
			return constants.CheckResultError, nil, err
		}
	}
	return constants.CheckResultExisted, found, nil
}

func (r *RightsizingReconciler) updateStatus(desiredService *v1alpha1.Rightsizing) error {
	existingService := &v1alpha1.Rightsizing{}
	namespacedName := types.NamespacedName{Namespace: desiredService.Namespace, Name: desiredService.Name}
	if err := r.Get(context.TODO(), namespacedName, existingService); err != nil {
		return err
	}

	if !equality.Semantic.DeepEqual(existingService.Status, desiredService.Status) {
		r.Recorder.Eventf(desiredService, v1.EventTypeNormal, "Update", fmt.Sprintf("state %s to %s", existingService.Status.Status, desiredService.Status.Status))
		if err := r.Status().Update(context.TODO(), desiredService); err != nil {
			return err
		}
	}
	return nil
}

func (r *RightsizingReconciler) deleteResources(service *v1alpha1.Rightsizing) error {
	found := &v1.Pod{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: service.Namespace, Name: service.Name}, found)
	if err != nil {
		r.Log.Error(err, "Failed to find old pod", "pod", found.Name)
		return err
	}
	if err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
		r.Log.Error(err, "Failed to delete pod", "pod", found.Name)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RightsizingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Rightsizing{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
