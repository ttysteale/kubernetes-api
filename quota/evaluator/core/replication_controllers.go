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

package core

import (
	"github.com/ttysteale/kubernetes-api/admission"
	"github.com/ttysteale/kubernetes-api/api"
	clientset "github.com/ttysteale/kubernetes-api/client/clientset_generated/internalclientset"
	"github.com/ttysteale/kubernetes-api/quota"
	"github.com/ttysteale/kubernetes-api/quota/generic"
	"github.com/ttysteale/kubernetes-api/runtime"
)

// NewReplicationControllerEvaluator returns an evaluator that can evaluate replication controllers
func NewReplicationControllerEvaluator(kubeClient clientset.Interface) quota.Evaluator {
	allResources := []api.ResourceName{api.ResourceReplicationControllers}
	return &generic.GenericEvaluator{
		Name:              "Evaluator.ReplicationController",
		InternalGroupKind: api.Kind("ReplicationController"),
		InternalOperationResources: map[admission.Operation][]api.ResourceName{
			admission.Create: allResources,
		},
		MatchedResourceNames: allResources,
		MatchesScopeFunc:     generic.MatchesNoScopeFunc,
		ConstraintsFunc:      generic.ObjectCountConstraintsFunc(api.ResourceReplicationControllers),
		UsageFunc:            generic.ObjectCountUsageFunc(api.ResourceReplicationControllers),
		ListFuncByNamespace: func(namespace string, options api.ListOptions) (runtime.Object, error) {
			return kubeClient.Core().ReplicationControllers(namespace).List(options)
		},
	}
}
