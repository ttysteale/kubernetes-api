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
	core "github.com/ttysteale/kubernetes-api/client/testing/core"
	labels "github.com/ttysteale/kubernetes-api/labels"
	watch "github.com/ttysteale/kubernetes-api/watch"
)

// FakeLimitRanges implements LimitRangeInterface
type FakeLimitRanges struct {
	Fake *FakeCore
	ns   string
}

var limitrangesResource = unversioned.GroupVersionResource{Group: "", Version: "", Resource: "limitranges"}

func (c *FakeLimitRanges) Create(limitRange *api.LimitRange) (result *api.LimitRange, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(limitrangesResource, c.ns, limitRange), &api.LimitRange{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.LimitRange), err
}

func (c *FakeLimitRanges) Update(limitRange *api.LimitRange) (result *api.LimitRange, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(limitrangesResource, c.ns, limitRange), &api.LimitRange{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.LimitRange), err
}

func (c *FakeLimitRanges) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(limitrangesResource, c.ns, name), &api.LimitRange{})

	return err
}

func (c *FakeLimitRanges) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction(limitrangesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &api.LimitRangeList{})
	return err
}

func (c *FakeLimitRanges) Get(name string) (result *api.LimitRange, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(limitrangesResource, c.ns, name), &api.LimitRange{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.LimitRange), err
}

func (c *FakeLimitRanges) List(opts api.ListOptions) (result *api.LimitRangeList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(limitrangesResource, c.ns, opts), &api.LimitRangeList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &api.LimitRangeList{}
	for _, item := range obj.(*api.LimitRangeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested limitRanges.
func (c *FakeLimitRanges) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(limitrangesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched limitRange.
func (c *FakeLimitRanges) Patch(name string, pt api.PatchType, data []byte) (result *api.LimitRange, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchAction(limitrangesResource, c.ns, name, data), &api.LimitRange{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.LimitRange), err
}
