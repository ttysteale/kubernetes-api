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

package batch

import (
	api "github.com/ttysteale/kubernetes-api/api"
	unversioned "github.com/ttysteale/kubernetes-api/api/unversioned"
	conversion "github.com/ttysteale/kubernetes-api/conversion"
)

func init() {
	if err := api.Scheme.AddGeneratedDeepCopyFuncs(
		DeepCopy_batch_Job,
		DeepCopy_batch_JobCondition,
		DeepCopy_batch_JobList,
		DeepCopy_batch_JobSpec,
		DeepCopy_batch_JobStatus,
		DeepCopy_batch_JobTemplate,
		DeepCopy_batch_JobTemplateSpec,
		DeepCopy_batch_ScheduledJob,
		DeepCopy_batch_ScheduledJobList,
		DeepCopy_batch_ScheduledJobSpec,
		DeepCopy_batch_ScheduledJobStatus,
	); err != nil {
		// if one of the deep copy functions is malformed, detect it immediately.
		panic(err)
	}
}

func DeepCopy_batch_Job(in Job, out *Job, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_batch_JobSpec(in.Spec, &out.Spec, c); err != nil {
		return err
	}
	if err := DeepCopy_batch_JobStatus(in.Status, &out.Status, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_batch_JobCondition(in JobCondition, out *JobCondition, c *conversion.Cloner) error {
	out.Type = in.Type
	out.Status = in.Status
	if err := unversioned.DeepCopy_unversioned_Time(in.LastProbeTime, &out.LastProbeTime, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_Time(in.LastTransitionTime, &out.LastTransitionTime, c); err != nil {
		return err
	}
	out.Reason = in.Reason
	out.Message = in.Message
	return nil
}

func DeepCopy_batch_JobList(in JobList, out *JobList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]Job, len(in))
		for i := range in {
			if err := DeepCopy_batch_Job(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_batch_JobSpec(in JobSpec, out *JobSpec, c *conversion.Cloner) error {
	if in.Parallelism != nil {
		in, out := in.Parallelism, &out.Parallelism
		*out = new(int32)
		**out = *in
	} else {
		out.Parallelism = nil
	}
	if in.Completions != nil {
		in, out := in.Completions, &out.Completions
		*out = new(int32)
		**out = *in
	} else {
		out.Completions = nil
	}
	if in.ActiveDeadlineSeconds != nil {
		in, out := in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = *in
	} else {
		out.ActiveDeadlineSeconds = nil
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
	if in.ManualSelector != nil {
		in, out := in.ManualSelector, &out.ManualSelector
		*out = new(bool)
		**out = *in
	} else {
		out.ManualSelector = nil
	}
	if err := api.DeepCopy_api_PodTemplateSpec(in.Template, &out.Template, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_batch_JobStatus(in JobStatus, out *JobStatus, c *conversion.Cloner) error {
	if in.Conditions != nil {
		in, out := in.Conditions, &out.Conditions
		*out = make([]JobCondition, len(in))
		for i := range in {
			if err := DeepCopy_batch_JobCondition(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Conditions = nil
	}
	if in.StartTime != nil {
		in, out := in.StartTime, &out.StartTime
		*out = new(unversioned.Time)
		if err := unversioned.DeepCopy_unversioned_Time(*in, *out, c); err != nil {
			return err
		}
	} else {
		out.StartTime = nil
	}
	if in.CompletionTime != nil {
		in, out := in.CompletionTime, &out.CompletionTime
		*out = new(unversioned.Time)
		if err := unversioned.DeepCopy_unversioned_Time(*in, *out, c); err != nil {
			return err
		}
	} else {
		out.CompletionTime = nil
	}
	out.Active = in.Active
	out.Succeeded = in.Succeeded
	out.Failed = in.Failed
	return nil
}

func DeepCopy_batch_JobTemplate(in JobTemplate, out *JobTemplate, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_batch_JobTemplateSpec(in.Template, &out.Template, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_batch_JobTemplateSpec(in JobTemplateSpec, out *JobTemplateSpec, c *conversion.Cloner) error {
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_batch_JobSpec(in.Spec, &out.Spec, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_batch_ScheduledJob(in ScheduledJob, out *ScheduledJob, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_batch_ScheduledJobSpec(in.Spec, &out.Spec, c); err != nil {
		return err
	}
	if err := DeepCopy_batch_ScheduledJobStatus(in.Status, &out.Status, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_batch_ScheduledJobList(in ScheduledJobList, out *ScheduledJobList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]ScheduledJob, len(in))
		for i := range in {
			if err := DeepCopy_batch_ScheduledJob(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_batch_ScheduledJobSpec(in ScheduledJobSpec, out *ScheduledJobSpec, c *conversion.Cloner) error {
	out.Schedule = in.Schedule
	if in.StartingDeadlineSeconds != nil {
		in, out := in.StartingDeadlineSeconds, &out.StartingDeadlineSeconds
		*out = new(int64)
		**out = *in
	} else {
		out.StartingDeadlineSeconds = nil
	}
	out.ConcurrencyPolicy = in.ConcurrencyPolicy
	if in.Suspend != nil {
		in, out := in.Suspend, &out.Suspend
		*out = new(bool)
		**out = *in
	} else {
		out.Suspend = nil
	}
	if err := DeepCopy_batch_JobTemplateSpec(in.JobTemplate, &out.JobTemplate, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_batch_ScheduledJobStatus(in ScheduledJobStatus, out *ScheduledJobStatus, c *conversion.Cloner) error {
	if in.Active != nil {
		in, out := in.Active, &out.Active
		*out = make([]api.ObjectReference, len(in))
		for i := range in {
			if err := api.DeepCopy_api_ObjectReference(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Active = nil
	}
	if in.LastScheduleTime != nil {
		in, out := in.LastScheduleTime, &out.LastScheduleTime
		*out = new(unversioned.Time)
		if err := unversioned.DeepCopy_unversioned_Time(*in, *out, c); err != nil {
			return err
		}
	} else {
		out.LastScheduleTime = nil
	}
	return nil
}
