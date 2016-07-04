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

package clusterrole

import (
	"fmt"

	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/api/rest"
	"github.com/ttysteale/kubernetes-api/apis/rbac"
	"github.com/ttysteale/kubernetes-api/apis/rbac/validation"
	"github.com/ttysteale/kubernetes-api/fields"
	"github.com/ttysteale/kubernetes-api/labels"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/util/validation/field"
)

// strategy implements behavior for ClusterRoles
type strategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// strategy is the default logic that applies when creating and updating
// ClusterRole objects.
var Strategy = strategy{api.Scheme, api.SimpleNameGenerator}

// Strategy should implement rest.RESTCreateStrategy
var _ rest.RESTCreateStrategy = Strategy

// Strategy should implement rest.RESTUpdateStrategy
var _ rest.RESTUpdateStrategy = Strategy

// NamespaceScoped is true for ClusterRoles.
func (strategy) NamespaceScoped() bool {
	return false
}

// AllowCreateOnUpdate is true for ClusterRoles.
func (strategy) AllowCreateOnUpdate() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set by end users
// on creation.
func (strategy) PrepareForCreate(obj runtime.Object) {
	_ = obj.(*rbac.ClusterRole)
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (strategy) PrepareForUpdate(obj, old runtime.Object) {
	newClusterRole := obj.(*rbac.ClusterRole)
	oldClusterRole := old.(*rbac.ClusterRole)

	_, _ = newClusterRole, oldClusterRole
}

// Validate validates a new ClusterRole. Validation must check for a correct signature.
func (strategy) Validate(ctx api.Context, obj runtime.Object) field.ErrorList {
	clusterRole := obj.(*rbac.ClusterRole)
	return validation.ValidateClusterRole(clusterRole)
}

// Canonicalize normalizes the object after validation.
func (strategy) Canonicalize(obj runtime.Object) {
	_ = obj.(*rbac.ClusterRole)
}

// ValidateUpdate is the default update validation for an end user.
func (strategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	newObj := obj.(*rbac.ClusterRole)
	errorList := validation.ValidateClusterRole(newObj)
	return append(errorList, validation.ValidateClusterRoleUpdate(newObj, old.(*rbac.ClusterRole))...)
}

// If AllowUnconditionalUpdate() is true and the object specified by
// the user does not have a resource version, then generic Update()
// populates it with the latest version. Else, it checks that the
// version specified by the user matches the version of latest etcd
// object.
func (strategy) AllowUnconditionalUpdate() bool {
	return true
}

func (s strategy) Export(obj runtime.Object, exact bool) error {
	return nil
}

// Matcher returns a generic matcher for a given label and field selector.
func Matcher(label labels.Selector, field fields.Selector) generic.Matcher {
	return generic.MatcherFunc(func(obj runtime.Object) (bool, error) {
		sa, ok := obj.(*rbac.ClusterRole)
		if !ok {
			return false, fmt.Errorf("not a ClusterRole")
		}
		fields := SelectableFields(sa)
		return label.Matches(labels.Set(sa.Labels)) && field.Matches(fields), nil
	})
}

// SelectableFields returns a label set that can be used for filter selection
func SelectableFields(obj *rbac.ClusterRole) labels.Set {
	return labels.Set{}
}
