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
	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/apimachinery/registered"
	clientset "github.com/ttysteale/kubernetes-api/client/clientset_generated/release_1_2"
	v1core "github.com/ttysteale/kubernetes-api/client/clientset_generated/release_1_2/typed/core/v1"
	fakev1core "github.com/ttysteale/kubernetes-api/client/clientset_generated/release_1_2/typed/core/v1/fake"
	v1beta1extensions "github.com/ttysteale/kubernetes-api/client/clientset_generated/release_1_2/typed/extensions/v1beta1"
	fakev1beta1extensions "github.com/ttysteale/kubernetes-api/client/clientset_generated/release_1_2/typed/extensions/v1beta1/fake"
	"github.com/ttysteale/kubernetes-api/client/testing/core"
	"github.com/ttysteale/kubernetes-api/client/typed/discovery"
	fakediscovery "github.com/ttysteale/kubernetes-api/client/typed/discovery/fake"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/watch"
)

// Clientset returns a clientset that will respond with the provided objects
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := core.NewObjects(api.Scheme, api.Codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	fakePtr := core.Fake{}
	fakePtr.AddReactor("*", "*", core.ObjectReaction(o, registered.RESTMapper()))

	fakePtr.AddWatchReactor("*", core.DefaultWatchReactor(watch.NewFake(), nil))

	return &Clientset{fakePtr}
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	core.Fake
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return &fakediscovery.FakeDiscovery{Fake: &c.Fake}
}

var _ clientset.Interface = &Clientset{}

// Core retrieves the CoreClient
func (c *Clientset) Core() v1core.CoreInterface {
	return &fakev1core.FakeCore{Fake: &c.Fake}
}

// Extensions retrieves the ExtensionsClient
func (c *Clientset) Extensions() v1beta1extensions.ExtensionsInterface {
	return &fakev1beta1extensions.FakeExtensions{Fake: &c.Fake}
}
