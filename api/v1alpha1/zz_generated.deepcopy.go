//go:build !ignore_autogenerated

/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CascadeAutoOperator) DeepCopyInto(out *CascadeAutoOperator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CascadeAutoOperator.
func (in *CascadeAutoOperator) DeepCopy() *CascadeAutoOperator {
	if in == nil {
		return nil
	}
	out := new(CascadeAutoOperator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CascadeAutoOperator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CascadeAutoOperatorList) DeepCopyInto(out *CascadeAutoOperatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CascadeAutoOperator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CascadeAutoOperatorList.
func (in *CascadeAutoOperatorList) DeepCopy() *CascadeAutoOperatorList {
	if in == nil {
		return nil
	}
	out := new(CascadeAutoOperatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CascadeAutoOperatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CascadeAutoOperatorSpec) DeepCopyInto(out *CascadeAutoOperatorSpec) {
	*out = *in
	in.ScenarioConfig.DeepCopyInto(&out.ScenarioConfig)
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	in.Template.DeepCopyInto(&out.Template)
	in.Strategy.DeepCopyInto(&out.Strategy)
	if in.RevisionHistoryLimit != nil {
		in, out := &in.RevisionHistoryLimit, &out.RevisionHistoryLimit
		*out = new(int32)
		**out = **in
	}
	if in.ProgressDeadlineSeconds != nil {
		in, out := &in.ProgressDeadlineSeconds, &out.ProgressDeadlineSeconds
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CascadeAutoOperatorSpec.
func (in *CascadeAutoOperatorSpec) DeepCopy() *CascadeAutoOperatorSpec {
	if in == nil {
		return nil
	}
	out := new(CascadeAutoOperatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CascadeAutoOperatorStatus) DeepCopyInto(out *CascadeAutoOperatorStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CascadeAutoOperatorStatus.
func (in *CascadeAutoOperatorStatus) DeepCopy() *CascadeAutoOperatorStatus {
	if in == nil {
		return nil
	}
	out := new(CascadeAutoOperatorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CascadeModule) DeepCopyInto(out *CascadeModule) {
	*out = *in
	if in.Configuration != nil {
		in, out := &in.Configuration, &out.Configuration
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ActiveDeadlineSeconds != nil {
		in, out := &in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.BackoffLimit != nil {
		in, out := &in.BackoffLimit, &out.BackoffLimit
		*out = new(int32)
		**out = **in
	}
	if in.TTLSecondsAfterFinished != nil {
		in, out := &in.TTLSecondsAfterFinished, &out.TTLSecondsAfterFinished
		*out = new(int32)
		**out = **in
	}
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CascadeModule.
func (in *CascadeModule) DeepCopy() *CascadeModule {
	if in == nil {
		return nil
	}
	out := new(CascadeModule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CascadeScenario) DeepCopyInto(out *CascadeScenario) {
	*out = *in
	if in.CascadeModules != nil {
		in, out := &in.CascadeModules, &out.CascadeModules
		*out = make([]CascadeModule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CascadeScenario.
func (in *CascadeScenario) DeepCopy() *CascadeScenario {
	if in == nil {
		return nil
	}
	out := new(CascadeScenario)
	in.DeepCopyInto(out)
	return out
}
