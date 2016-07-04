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

package serviceaccount

import (
	"github.com/ttysteale/kubernetes-api/api"
	clientset "github.com/ttysteale/kubernetes-api/client/clientset_generated/internalclientset"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	"github.com/ttysteale/kubernetes-api/registry/secret"
	secretetcd "github.com/ttysteale/kubernetes-api/registry/secret/etcd"
	serviceaccountregistry "github.com/ttysteale/kubernetes-api/registry/serviceaccount"
	serviceaccountetcd "github.com/ttysteale/kubernetes-api/registry/serviceaccount/etcd"
	"github.com/ttysteale/kubernetes-api/serviceaccount"
	"github.com/ttysteale/kubernetes-api/storage"
)

// clientGetter implements ServiceAccountTokenGetter using a clientset.Interface
type clientGetter struct {
	client clientset.Interface
}

// NewGetterFromClient returns a ServiceAccountTokenGetter that
// uses the specified client to retrieve service accounts and secrets.
// The client should NOT authenticate using a service account token
// the returned getter will be used to retrieve, or recursion will result.
func NewGetterFromClient(c clientset.Interface) serviceaccount.ServiceAccountTokenGetter {
	return clientGetter{c}
}
func (c clientGetter) GetServiceAccount(namespace, name string) (*api.ServiceAccount, error) {
	return c.client.Core().ServiceAccounts(namespace).Get(name)
}
func (c clientGetter) GetSecret(namespace, name string) (*api.Secret, error) {
	return c.client.Core().Secrets(namespace).Get(name)
}

// registryGetter implements ServiceAccountTokenGetter using a service account and secret registry
type registryGetter struct {
	serviceAccounts serviceaccountregistry.Registry
	secrets         secret.Registry
}

// NewGetterFromRegistries returns a ServiceAccountTokenGetter that
// uses the specified registries to retrieve service accounts and secrets.
func NewGetterFromRegistries(serviceAccounts serviceaccountregistry.Registry, secrets secret.Registry) serviceaccount.ServiceAccountTokenGetter {
	return &registryGetter{serviceAccounts, secrets}
}
func (r *registryGetter) GetServiceAccount(namespace, name string) (*api.ServiceAccount, error) {
	ctx := api.WithNamespace(api.NewContext(), namespace)
	return r.serviceAccounts.GetServiceAccount(ctx, name)
}
func (r *registryGetter) GetSecret(namespace, name string) (*api.Secret, error) {
	ctx := api.WithNamespace(api.NewContext(), namespace)
	return r.secrets.GetSecret(ctx, name)
}

// NewGetterFromStorageInterface returns a ServiceAccountTokenGetter that
// uses the specified storage to retrieve service accounts and secrets.
func NewGetterFromStorageInterface(s storage.Interface) serviceaccount.ServiceAccountTokenGetter {
	return NewGetterFromRegistries(
		serviceaccountregistry.NewRegistry(serviceaccountetcd.NewREST(generic.RESTOptions{Storage: s, Decorator: generic.UndecoratedStorage})),
		secret.NewRegistry(secretetcd.NewREST(generic.RESTOptions{Storage: s, Decorator: generic.UndecoratedStorage})),
	)
}
