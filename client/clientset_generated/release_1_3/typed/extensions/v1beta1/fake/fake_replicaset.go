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
	v1beta1 "github.com/ttysteale/kubernetes-api/apis/extensions/v1beta1"
	core "github.com/ttysteale/kubernetes-api/client/testing/core"
	labels "github.com/ttysteale/kubernetes-api/labels"
	watch "github.com/ttysteale/kubernetes-api/watch"
)

// FakeReplicaSets implements ReplicaSetInterface
type FakeReplicaSets struct {
	Fake *FakeExtensions
	ns   string
}

var replicasetsResource = unversioned.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "replicasets"}

func (c *FakeReplicaSets) Create(replicaSet *v1beta1.ReplicaSet) (result *v1beta1.ReplicaSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(replicasetsResource, c.ns, replicaSet), &v1beta1.ReplicaSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReplicaSet), err
}

func (c *FakeReplicaSets) Update(replicaSet *v1beta1.ReplicaSet) (result *v1beta1.ReplicaSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(replicasetsResource, c.ns, replicaSet), &v1beta1.ReplicaSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReplicaSet), err
}

func (c *FakeReplicaSets) UpdateStatus(replicaSet *v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction(replicasetsResource, "status", c.ns, replicaSet), &v1beta1.ReplicaSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReplicaSet), err
}

func (c *FakeReplicaSets) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(replicasetsResource, c.ns, name), &v1beta1.ReplicaSet{})

	return err
}

func (c *FakeReplicaSets) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction(replicasetsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ReplicaSetList{})
	return err
}

func (c *FakeReplicaSets) Get(name string) (result *v1beta1.ReplicaSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(replicasetsResource, c.ns, name), &v1beta1.ReplicaSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReplicaSet), err
}

func (c *FakeReplicaSets) List(opts api.ListOptions) (result *v1beta1.ReplicaSetList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(replicasetsResource, c.ns, opts), &v1beta1.ReplicaSetList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ReplicaSetList{}
	for _, item := range obj.(*v1beta1.ReplicaSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested replicaSets.
func (c *FakeReplicaSets) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(replicasetsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched replicaSet.
func (c *FakeReplicaSets) Patch(name string, pt api.PatchType, data []byte) (result *v1beta1.ReplicaSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchAction(replicasetsResource, c.ns, name, data), &v1beta1.ReplicaSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ReplicaSet), err
}
