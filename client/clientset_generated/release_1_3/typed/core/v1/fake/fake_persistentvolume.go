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

package fake

import (
	api "github.com/ttysteale/kubernetes-api/api"
	unversioned "github.com/ttysteale/kubernetes-api/api/unversioned"
	v1 "github.com/ttysteale/kubernetes-api/api/v1"
	core "github.com/ttysteale/kubernetes-api/client/testing/core"
	labels "github.com/ttysteale/kubernetes-api/labels"
	watch "github.com/ttysteale/kubernetes-api/watch"
)

// FakePersistentVolumes implements PersistentVolumeInterface
type FakePersistentVolumes struct {
	Fake *FakeCore
}

var persistentvolumesResource = unversioned.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumes"}

func (c *FakePersistentVolumes) Create(persistentVolume *v1.PersistentVolume) (result *v1.PersistentVolume, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootCreateAction(persistentvolumesResource, persistentVolume), &v1.PersistentVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.PersistentVolume), err
}

func (c *FakePersistentVolumes) Update(persistentVolume *v1.PersistentVolume) (result *v1.PersistentVolume, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootUpdateAction(persistentvolumesResource, persistentVolume), &v1.PersistentVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.PersistentVolume), err
}

func (c *FakePersistentVolumes) UpdateStatus(persistentVolume *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	obj, err := c.Fake.
		Invokes(core.NewRootUpdateSubresourceAction(persistentvolumesResource, "status", persistentVolume), &v1.PersistentVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.PersistentVolume), err
}

func (c *FakePersistentVolumes) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewRootDeleteAction(persistentvolumesResource, name), &v1.PersistentVolume{})
	return err
}

func (c *FakePersistentVolumes) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewRootDeleteCollectionAction(persistentvolumesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1.PersistentVolumeList{})
	return err
}

func (c *FakePersistentVolumes) Get(name string) (result *v1.PersistentVolume, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootGetAction(persistentvolumesResource, name), &v1.PersistentVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.PersistentVolume), err
}

func (c *FakePersistentVolumes) List(opts api.ListOptions) (result *v1.PersistentVolumeList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootListAction(persistentvolumesResource, opts), &v1.PersistentVolumeList{})
	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.PersistentVolumeList{}
	for _, item := range obj.(*v1.PersistentVolumeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested persistentVolumes.
func (c *FakePersistentVolumes) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewRootWatchAction(persistentvolumesResource, opts))
}

// Patch applies the patch and returns the patched persistentVolume.
func (c *FakePersistentVolumes) Patch(name string, pt api.PatchType, data []byte) (result *v1.PersistentVolume, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootPatchAction(persistentvolumesResource, name, data), &v1.PersistentVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.PersistentVolume), err
}
