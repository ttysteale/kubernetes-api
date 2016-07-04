/*
Copyright 2014 The Kubernetes Authors.

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
	"github.com/ttysteale/kubernetes-api/api/v1"
	"github.com/ttysteale/kubernetes-api/client/restclient"
	"github.com/ttysteale/kubernetes-api/client/testing/core"
)

func (c *FakePods) Bind(binding *v1.Binding) error {
	action := core.CreateActionImpl{}
	action.Verb = "create"
	action.Resource = podsResource
	action.Subresource = "bindings"
	action.Object = binding

	_, err := c.Fake.Invokes(action, binding)
	return err
}

func (c *FakePods) GetLogs(name string, opts *v1.PodLogOptions) *restclient.Request {
	action := core.GenericActionImpl{}
	action.Verb = "get"
	action.Namespace = c.ns
	action.Resource = podsResource
	action.Subresource = "logs"
	action.Value = opts

	_, _ = c.Fake.Invokes(action, &v1.Pod{})
	return &restclient.Request{}
}
