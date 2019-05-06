// +build !ignore_autogenerated

/*
Copyright 2019 The OpenEBS Authors

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CStorPoolAttr) DeepCopyInto(out *CStorPoolAttr) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CStorPoolAttr.
func (in *CStorPoolAttr) DeepCopy() *CStorPoolAttr {
	if in == nil {
		return nil
	}
	out := new(CStorPoolAttr)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaim) DeepCopyInto(out *StoragePoolClaim) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaim.
func (in *StoragePoolClaim) DeepCopy() *StoragePoolClaim {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StoragePoolClaim) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimDisk) DeepCopyInto(out *StoragePoolClaimDisk) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimDisk.
func (in *StoragePoolClaimDisk) DeepCopy() *StoragePoolClaimDisk {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimDisk)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimDiskGroups) DeepCopyInto(out *StoragePoolClaimDiskGroups) {
	*out = *in
	if in.Disks != nil {
		in, out := &in.Disks, &out.Disks
		*out = make([]StoragePoolClaimDisk, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimDiskGroups.
func (in *StoragePoolClaimDiskGroups) DeepCopy() *StoragePoolClaimDiskGroups {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimDiskGroups)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimList) DeepCopyInto(out *StoragePoolClaimList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StoragePoolClaim, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimList.
func (in *StoragePoolClaimList) DeepCopy() *StoragePoolClaimList {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StoragePoolClaimList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimNodeSpec) DeepCopyInto(out *StoragePoolClaimNodeSpec) {
	*out = *in
	out.PoolSpec = in.PoolSpec
	if in.DiskGroups != nil {
		in, out := &in.DiskGroups, &out.DiskGroups
		*out = make([]StoragePoolClaimDiskGroups, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimNodeSpec.
func (in *StoragePoolClaimNodeSpec) DeepCopy() *StoragePoolClaimNodeSpec {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimNodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimSpec) DeepCopyInto(out *StoragePoolClaimSpec) {
	*out = *in
	if in.MaxPools != nil {
		in, out := &in.MaxPools, &out.MaxPools
		*out = new(int)
		**out = **in
	}
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]StoragePoolClaimNodeSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.PoolSpec = in.PoolSpec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimSpec.
func (in *StoragePoolClaimSpec) DeepCopy() *StoragePoolClaimSpec {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimStatus) DeepCopyInto(out *StoragePoolClaimStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimStatus.
func (in *StoragePoolClaimStatus) DeepCopy() *StoragePoolClaimStatus {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimStatus)
	in.DeepCopyInto(out)
	return out
}
