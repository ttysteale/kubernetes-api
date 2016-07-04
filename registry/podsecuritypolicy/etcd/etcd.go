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

package etcd

import (
	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/apis/extensions"
	"github.com/ttysteale/kubernetes-api/fields"
	"github.com/ttysteale/kubernetes-api/labels"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	"github.com/ttysteale/kubernetes-api/registry/generic/registry"
	"github.com/ttysteale/kubernetes-api/registry/podsecuritypolicy"
	"github.com/ttysteale/kubernetes-api/runtime"
)

// REST implements a RESTStorage for PodSecurityPolicies against etcd.
type REST struct {
	*registry.Store
}

const Prefix = "/podsecuritypolicies"

// NewREST returns a RESTStorage object that will work against PodSecurityPolicy objects.
func NewREST(opts generic.RESTOptions) *REST {
	newListFunc := func() runtime.Object { return &extensions.PodSecurityPolicyList{} }
	storageInterface := opts.Decorator(
		opts.Storage, 100, &extensions.PodSecurityPolicy{}, Prefix, podsecuritypolicy.Strategy, newListFunc)

	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &extensions.PodSecurityPolicy{} },
		NewListFunc: newListFunc,
		KeyRootFunc: func(ctx api.Context) string {
			return Prefix
		},
		KeyFunc: func(ctx api.Context, name string) (string, error) {
			return registry.NoNamespaceKeyFunc(ctx, Prefix, name)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*extensions.PodSecurityPolicy).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return podsecuritypolicy.MatchPodSecurityPolicy(label, field)
		},
		QualifiedResource:       extensions.Resource("podsecuritypolicies"),
		DeleteCollectionWorkers: opts.DeleteCollectionWorkers,

		CreateStrategy:      podsecuritypolicy.Strategy,
		UpdateStrategy:      podsecuritypolicy.Strategy,
		DeleteStrategy:      podsecuritypolicy.Strategy,
		ReturnDeletedObject: true,
		Storage:             storageInterface,
	}
	return &REST{store}
}
