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

package v1alpha1

import (
	"github.com/ttysteale/kubernetes-api/api/unversioned"
	"github.com/ttysteale/kubernetes-api/api/v1"
	"github.com/ttysteale/kubernetes-api/apis/rbac"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/watch/versioned"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = unversioned.GroupVersion{Group: rbac.GroupName, Version: "v1alpha1"}

func AddToScheme(scheme *runtime.Scheme) {
	addKnownTypes(scheme)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Role{},
		&RoleBinding{},
		&RoleBindingList{},
		&RoleList{},

		&ClusterRole{},
		&ClusterRoleBinding{},
		&ClusterRoleBindingList{},
		&ClusterRoleList{},

		&v1.ListOptions{},
		&v1.DeleteOptions{},
		&v1.ExportOptions{},
	)
	versioned.AddToGroupVersion(scheme, SchemeGroupVersion)
}
