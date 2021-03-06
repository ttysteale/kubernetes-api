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
	"github.com/ttysteale/kubernetes-api/api/rest"
	"github.com/ttysteale/kubernetes-api/apis/certificates"
	"github.com/ttysteale/kubernetes-api/fields"
	"github.com/ttysteale/kubernetes-api/labels"
	"github.com/ttysteale/kubernetes-api/registry/cachesize"
	csrregistry "github.com/ttysteale/kubernetes-api/registry/certificates"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	"github.com/ttysteale/kubernetes-api/registry/generic/registry"
	"github.com/ttysteale/kubernetes-api/runtime"
)

// REST implements a RESTStorage for CertificateSigningRequest against etcd
type REST struct {
	*registry.Store
}

// NewREST returns a registry which will store CertificateSigningRequest in the given helper
func NewREST(opts generic.RESTOptions) (*REST, *StatusREST, *ApprovalREST) {
	prefix := "/certificatesigningrequests"

	newListFunc := func() runtime.Object { return &certificates.CertificateSigningRequestList{} }
	storageInterface := opts.Decorator(opts.Storage, cachesize.GetWatchCacheSizeByResource(cachesize.CertificateSigningRequests), &certificates.CertificateSigningRequest{}, prefix, csrregistry.Strategy, newListFunc)

	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &certificates.CertificateSigningRequest{} },
		NewListFunc: newListFunc,
		KeyRootFunc: func(ctx api.Context) string {
			return prefix
		},
		KeyFunc: func(ctx api.Context, id string) (string, error) {
			return registry.NoNamespaceKeyFunc(ctx, prefix, id)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*certificates.CertificateSigningRequest).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return csrregistry.Matcher(label, field)
		},
		QualifiedResource:       certificates.Resource("certificatesigningrequests"),
		DeleteCollectionWorkers: opts.DeleteCollectionWorkers,

		CreateStrategy: csrregistry.Strategy,
		UpdateStrategy: csrregistry.Strategy,

		Storage: storageInterface,
	}

	// Subresources use the same store and creation strategy, which only
	// allows empty subs. Updates to an existing subresource are handled by
	// dedicated strategies.
	statusStore := *store
	statusStore.UpdateStrategy = csrregistry.StatusStrategy

	approvalStore := *store
	approvalStore.UpdateStrategy = csrregistry.ApprovalStrategy

	return &REST{store}, &StatusREST{store: &statusStore}, &ApprovalREST{store: &approvalStore}
}

// StatusREST implements the REST endpoint for changing the status of a CSR.
type StatusREST struct {
	store *registry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &certificates.CertificateSigningRequest{}
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx api.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}

// ApprovalREST implements the REST endpoint for changing the approval state of a CSR.
type ApprovalREST struct {
	store *registry.Store
}

func (r *ApprovalREST) New() runtime.Object {
	return &certificates.CertificateSigningRequest{}
}

// Update alters the approval subset of an object.
func (r *ApprovalREST) Update(ctx api.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}
