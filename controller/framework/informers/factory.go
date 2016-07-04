/*
Copyright 2015 The Kubernetes Authors.

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

package informers

import (
	"time"

	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/client/cache"
	clientset "github.com/ttysteale/kubernetes-api/client/clientset_generated/internalclientset"
	"github.com/ttysteale/kubernetes-api/controller/framework"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/watch"
)

// CreateSharedPodInformer returns a SharedIndexInformer that lists and watches all pods
func CreateSharedPodInformer(client clientset.Interface, resyncPeriod time.Duration) framework.SharedIndexInformer {
	sharedInformer := framework.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Core().Pods(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Core().Pods(api.NamespaceAll).Watch(options)
			},
		},
		&api.Pod{},
		resyncPeriod,
		cache.Indexers{},
	)

	return sharedInformer
}

// CreateSharedPodIndexInformer returns a SharedIndexInformer that lists and watches all pods
func CreateSharedPodIndexInformer(client clientset.Interface, resyncPeriod time.Duration) framework.SharedIndexInformer {
	sharedIndexInformer := framework.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Core().Pods(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Core().Pods(api.NamespaceAll).Watch(options)
			},
		},
		&api.Pod{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	return sharedIndexInformer
}

// CreateSharedNodeIndexInformer returns a SharedIndexInformer that lists and watches all nodes
func CreateSharedNodeIndexInformer(client clientset.Interface, resyncPeriod time.Duration) framework.SharedIndexInformer {
	sharedIndexInformer := framework.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Core().Nodes().List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Core().Nodes().Watch(options)
			},
		},
		&api.Node{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	return sharedIndexInformer
}

// CreateSharedPVCIndexInformer returns a SharedIndexInformer that lists and watches all PVCs
func CreateSharedPVCIndexInformer(client clientset.Interface, resyncPeriod time.Duration) framework.SharedIndexInformer {
	sharedIndexInformer := framework.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Core().PersistentVolumeClaims(api.NamespaceAll).List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Core().PersistentVolumeClaims(api.NamespaceAll).Watch(options)
			},
		},
		&api.PersistentVolumeClaim{},
		resyncPeriod,
		cache.Indexers{})

	return sharedIndexInformer
}

// CreateSharedPVIndexInformer returns a SharedIndexInformer that lists and watches all PVs
func CreateSharedPVIndexInformer(client clientset.Interface, resyncPeriod time.Duration) framework.SharedIndexInformer {
	sharedIndexInformer := framework.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
				return client.Core().PersistentVolumes().List(options)
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return client.Core().PersistentVolumes().Watch(options)
			},
		},
		&api.PersistentVolume{},
		resyncPeriod,
		cache.Indexers{})

	return sharedIndexInformer
}
