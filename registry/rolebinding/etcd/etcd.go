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

package etcd

import (
	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/apis/rbac"
	"github.com/ttysteale/kubernetes-api/fields"
	"github.com/ttysteale/kubernetes-api/labels"
	"github.com/ttysteale/kubernetes-api/registry/cachesize"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	"github.com/ttysteale/kubernetes-api/registry/generic/registry"
	"github.com/ttysteale/kubernetes-api/registry/rolebinding"
	"github.com/ttysteale/kubernetes-api/runtime"
)

// REST implements a RESTStorage for RoleBinding against etcd
type REST struct {
	*registry.Store
}

// NewREST returns a RESTStorage object that will work against RoleBinding objects.
func NewREST(opts generic.RESTOptions) *REST {
	prefix := "/rolebindings"

	newListFunc := func() runtime.Object { return &rbac.RoleBindingList{} }
	storageInterface := opts.Decorator(
		opts.Storage,
		cachesize.GetWatchCacheSizeByResource(cachesize.RoleBindings),
		&rbac.RoleBinding{},
		prefix,
		rolebinding.Strategy,
		newListFunc,
	)

	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &rbac.RoleBinding{} },
		NewListFunc: newListFunc,
		KeyRootFunc: func(ctx api.Context) string {
			return registry.NamespaceKeyRootFunc(ctx, prefix)
		},
		KeyFunc: func(ctx api.Context, id string) (string, error) {
			return registry.NamespaceKeyFunc(ctx, prefix, id)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*rbac.RoleBinding).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return rolebinding.Matcher(label, field)
		},
		QualifiedResource:       rbac.Resource("rolebindings"),
		DeleteCollectionWorkers: opts.DeleteCollectionWorkers,

		CreateStrategy: rolebinding.Strategy,
		UpdateStrategy: rolebinding.Strategy,
		DeleteStrategy: rolebinding.Strategy,

		Storage: storageInterface,
	}

	return &REST{store}
}
