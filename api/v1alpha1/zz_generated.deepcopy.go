//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rightsizing) DeepCopyInto(out *Rightsizing) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rightsizing.
func (in *Rightsizing) DeepCopy() *Rightsizing {
	if in == nil {
		return nil
	}
	out := new(Rightsizing)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Rightsizing) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RightsizingList) DeepCopyInto(out *RightsizingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Rightsizing, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RightsizingList.
func (in *RightsizingList) DeepCopy() *RightsizingList {
	if in == nil {
		return nil
	}
	out := new(RightsizingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RightsizingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RightsizingSpec) DeepCopyInto(out *RightsizingSpec) {
	*out = *in
	if in.PrometheusUri != nil {
		in, out := &in.PrometheusUri, &out.PrometheusUri
		*out = new(string)
		**out = **in
	}
	if in.Optimization != nil {
		in, out := &in.Optimization, &out.Optimization
		*out = new(bool)
		**out = **in
	}
	if in.Forecast != nil {
		in, out := &in.Forecast, &out.Forecast
		*out = new(bool)
		**out = **in
	}
	if in.Trace != nil {
		in, out := &in.Trace, &out.Trace
		*out = new(bool)
		**out = **in
	}
	if in.TraceCycle != nil {
		in, out := &in.TraceCycle, &out.TraceCycle
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RightsizingSpec.
func (in *RightsizingSpec) DeepCopy() *RightsizingSpec {
	if in == nil {
		return nil
	}
	out := new(RightsizingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RightsizingStatus) DeepCopyInto(out *RightsizingStatus) {
	*out = *in
	if in.ServiceStatuses != nil {
		in, out := &in.ServiceStatuses, &out.ServiceStatuses
		*out = make(map[ServiceType]ServiceStatus, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.ServiceResults != nil {
		in, out := &in.ServiceResults, &out.ServiceResults
		*out = make(map[ServiceType]ServiceResult, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RightsizingStatus.
func (in *RightsizingStatus) DeepCopy() *RightsizingStatus {
	if in == nil {
		return nil
	}
	out := new(RightsizingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceData) DeepCopyInto(out *ServiceData) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceData.
func (in *ServiceData) DeepCopy() *ServiceData {
	if in == nil {
		return nil
	}
	out := new(ServiceData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceResult) DeepCopyInto(out *ServiceResult) {
	*out = *in
	out.Values = in.Values
	in.RecordedTime.DeepCopyInto(&out.RecordedTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceResult.
func (in *ServiceResult) DeepCopy() *ServiceResult {
	if in == nil {
		return nil
	}
	out := new(ServiceResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceStatus) DeepCopyInto(out *ServiceStatus) {
	*out = *in
	if in.URL != nil {
		in, out := &in.URL, &out.URL
		*out = new(string)
		**out = **in
	}
	if in.Reason != nil {
		in, out := &in.Reason, &out.Reason
		*out = new(string)
		**out = **in
	}
	if in.Message != nil {
		in, out := &in.Message, &out.Message
		*out = new(string)
		**out = **in
	}
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceStatus.
func (in *ServiceStatus) DeepCopy() *ServiceStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceStatusOption) DeepCopyInto(out *ServiceStatusOption) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceStatusOption.
func (in *ServiceStatusOption) DeepCopy() *ServiceStatusOption {
	if in == nil {
		return nil
	}
	out := new(ServiceStatusOption)
	in.DeepCopyInto(out)
	return out
}