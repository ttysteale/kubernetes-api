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

// FakeServiceAccounts implements ServiceAccountInterface
type FakeServiceAccounts struct {
	Fake *FakeCore
	ns   string
}

var serviceaccountsResource = unversioned.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}

func (c *FakeServiceAccounts) Create(serviceAccount *v1.ServiceAccount) (result *v1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(serviceaccountsResource, c.ns, serviceAccount), &v1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ServiceAccount), err
}

func (c *FakeServiceAccounts) Update(serviceAccount *v1.ServiceAccount) (result *v1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(serviceaccountsResource, c.ns, serviceAccount), &v1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ServiceAccount), err
}

func (c *FakeServiceAccounts) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(serviceaccountsResource, c.ns, name), &v1.ServiceAccount{})

	return err
}

func (c *FakeServiceAccounts) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction(serviceaccountsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1.ServiceAccountList{})
	return err
}

func (c *FakeServiceAccounts) Get(name string) (result *v1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(serviceaccountsResource, c.ns, name), &v1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ServiceAccount), err
}

func (c *FakeServiceAccounts) List(opts api.ListOptions) (result *v1.ServiceAccountList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(serviceaccountsResource, c.ns, opts), &v1.ServiceAccountList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ServiceAccountList{}
	for _, item := range obj.(*v1.ServiceAccountList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceAccounts.
func (c *FakeServiceAccounts) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(serviceaccountsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched serviceAccount.
func (c *FakeServiceAccounts) Patch(name string, pt api.PatchType, data []byte) (result *v1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchAction(serviceaccountsResource, c.ns, name, data), &v1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ServiceAccount), err
}
