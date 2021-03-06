// +build !ignore_autogenerated

/*

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddon) DeepCopyInto(out *ClusterAddon) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddon.
func (in *ClusterAddon) DeepCopy() *ClusterAddon {
	if in == nil {
		return nil
	}
	out := new(ClusterAddon)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterAddon) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonAction) DeepCopyInto(out *ClusterAddonAction) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonAction.
func (in *ClusterAddonAction) DeepCopy() *ClusterAddonAction {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonAction)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonCondition) DeepCopyInto(out *ClusterAddonCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonCondition.
func (in *ClusterAddonCondition) DeepCopy() *ClusterAddonCondition {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonList) DeepCopyInto(out *ClusterAddonList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterAddon, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonList.
func (in *ClusterAddonList) DeepCopy() *ClusterAddonList {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterAddonList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonSource) DeepCopyInto(out *ClusterAddonSource) {
	*out = *in
	in.Action.DeepCopyInto(&out.Action)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonSource.
func (in *ClusterAddonSource) DeepCopy() *ClusterAddonSource {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonSpec) DeepCopyInto(out *ClusterAddonSpec) {
	*out = *in
	in.Target.DeepCopyInto(&out.Target)
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make(map[string]ClusterAddonSource, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonSpec.
func (in *ClusterAddonSpec) DeepCopy() *ClusterAddonSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonStatus) DeepCopyInto(out *ClusterAddonStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ClusterAddonCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonStatus.
func (in *ClusterAddonStatus) DeepCopy() *ClusterAddonStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddonTarget) DeepCopyInto(out *ClusterAddonTarget) {
	*out = *in
	if in.CACert != nil {
		in, out := &in.CACert, &out.CACert
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.ClientCert != nil {
		in, out := &in.ClientCert, &out.ClientCert
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.ClientKey != nil {
		in, out := &in.ClientKey, &out.ClientKey
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddonTarget.
func (in *ClusterAddonTarget) DeepCopy() *ClusterAddonTarget {
	if in == nil {
		return nil
	}
	out := new(ClusterAddonTarget)
	in.DeepCopyInto(out)
	return out
}
