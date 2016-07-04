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

package limitrange

import (
	"fmt"

	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/api/validation"
	"github.com/ttysteale/kubernetes-api/fields"
	"github.com/ttysteale/kubernetes-api/labels"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/util"
	"github.com/ttysteale/kubernetes-api/util/validation/field"
)

type limitrangeStrategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// Strategy is the default logic that applies when creating and updating
// LimitRange objects via the REST API.
var Strategy = limitrangeStrategy{api.Scheme, api.SimpleNameGenerator}

func (limitrangeStrategy) NamespaceScoped() bool {
	return true
}

func (limitrangeStrategy) PrepareForCreate(obj runtime.Object) {
	limitRange := obj.(*api.LimitRange)
	if len(limitRange.Name) == 0 {
		limitRange.Name = string(util.NewUUID())
	}
}

func (limitrangeStrategy) PrepareForUpdate(obj, old runtime.Object) {
}

func (limitrangeStrategy) Validate(ctx api.Context, obj runtime.Object) field.ErrorList {
	limitRange := obj.(*api.LimitRange)
	return validation.ValidateLimitRange(limitRange)
}

// Canonicalize normalizes the object after validation.
func (limitrangeStrategy) Canonicalize(obj runtime.Object) {
}

func (limitrangeStrategy) AllowCreateOnUpdate() bool {
	return true
}

func (limitrangeStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	limitRange := obj.(*api.LimitRange)
	return validation.ValidateLimitRange(limitRange)
}

func (limitrangeStrategy) AllowUnconditionalUpdate() bool {
	return true
}

func LimitRangeToSelectableFields(limitRange *api.LimitRange) fields.Set {
	return fields.Set{}
}

func (limitrangeStrategy) Export(runtime.Object, bool) error {
	// Copied from OpenShift exporter
	// TODO: this needs to be fixed
	//  limitrange.Strategy.PrepareForCreate(obj)
	return nil
}

func MatchLimitRange(label labels.Selector, field fields.Selector) generic.Matcher {
	return &generic.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
			lr, ok := obj.(*api.LimitRange)
			if !ok {
				return nil, nil, fmt.Errorf("given object is not a limit range.")
			}
			return labels.Set(lr.ObjectMeta.Labels), LimitRangeToSelectableFields(lr), nil
		},
	}
}
