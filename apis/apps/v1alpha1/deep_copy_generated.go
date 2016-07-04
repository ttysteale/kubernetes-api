// +build !ignore_autogenerated

/*
Copyright 2016 The Kubernetes Authors.

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	api "github.com/ttysteale/kubernetes-api/api"
	unversioned "github.com/ttysteale/kubernetes-api/api/unversioned"
	v1 "github.com/ttysteale/kubernetes-api/api/v1"
	conversion "github.com/ttysteale/kubernetes-api/conversion"
)

func init() {
	if err := api.Scheme.AddGeneratedDeepCopyFuncs(
		DeepCopy_v1alpha1_PetSet,
		DeepCopy_v1alpha1_PetSetList,
		DeepCopy_v1alpha1_PetSetSpec,
		DeepCopy_v1alpha1_PetSetStatus,
	); err != nil {
		// if one of the deep copy functions is malformed, detect it immediately.
		panic(err)
	}
}

func DeepCopy_v1alpha1_PetSet(in PetSet, out *PetSet, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := v1.DeepCopy_v1_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_v1alpha1_PetSetSpec(in.Spec, &out.Spec, c); err != nil {
		return err
	}
	if err := DeepCopy_v1alpha1_PetSetStatus(in.Status, &out.Status, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_v1alpha1_PetSetList(in PetSetList, out *PetSetList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]PetSet, len(in))
		for i := range in {
			if err := DeepCopy_v1alpha1_PetSet(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_v1alpha1_PetSetSpec(in PetSetSpec, out *PetSetSpec, c *conversion.Cloner) error {
	if in.Replicas != nil {
		in, out := in.Replicas, &out.Replicas
		*out = new(int32)
		**out = *in
	} else {
		out.Replicas = nil
	}
	if in.Selector != nil {
		in, out := in.Selector, &out.Selector
		*out = new(unversioned.LabelSelector)
		if err := unversioned.DeepCopy_unversioned_LabelSelector(*in, *out, c); err != nil {
			return err
		}
	} else {
		out.Selector = nil
	}
	if err := v1.DeepCopy_v1_PodTemplateSpec(in.Template, &out.Template, c); err != nil {
		return err
	}
	if in.VolumeClaimTemplates != nil {
		in, out := in.VolumeClaimTemplates, &out.VolumeClaimTemplates
		*out = make([]v1.PersistentVolumeClaim, len(in))
		for i := range in {
			if err := v1.DeepCopy_v1_PersistentVolumeClaim(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.VolumeClaimTemplates = nil
	}
	out.ServiceName = in.ServiceName
	return nil
}

func DeepCopy_v1alpha1_PetSetStatus(in PetSetStatus, out *PetSetStatus, c *conversion.Cloner) error {
	if in.ObservedGeneration != nil {
		in, out := in.ObservedGeneration, &out.ObservedGeneration
		*out = new(int64)
		**out = *in
	} else {
		out.ObservedGeneration = nil
	}
	out.Replicas = in.Replicas
	return nil
}
