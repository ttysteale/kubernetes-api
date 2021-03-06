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

// FakeServices implements ServiceInterface
type FakeServices struct {
	Fake *FakeCore
	ns   string
}

var servicesResource = unversioned.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

func (c *FakeServices) Create(service *v1.Service) (result *v1.Service, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(servicesResource, c.ns, service), &v1.Service{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Service), err
}

func (c *FakeServices) Update(service *v1.Service) (result *v1.Service, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(servicesResource, c.ns, service), &v1.Service{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Service), err
}

func (c *FakeServices) UpdateStatus(service *v1.Service) (*v1.Service, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction(servicesResource, "status", c.ns, service), &v1.Service{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Service), err
}

func (c *FakeServices) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(servicesResource, c.ns, name), &v1.Service{})

	return err
}

func (c *FakeServices) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction(servicesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1.ServiceList{})
	return err
}

func (c *FakeServices) Get(name string) (result *v1.Service, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(servicesResource, c.ns, name), &v1.Service{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Service), err
}

func (c *FakeServices) List(opts api.ListOptions) (result *v1.ServiceList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(servicesResource, c.ns, opts), &v1.ServiceList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ServiceList{}
	for _, item := range obj.(*v1.ServiceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested services.
func (c *FakeServices) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(servicesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched service.
func (c *FakeServices) Patch(name string, pt api.PatchType, data []byte) (result *v1.Service, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchAction(servicesResource, c.ns, name, data), &v1.Service{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Service), err
}
