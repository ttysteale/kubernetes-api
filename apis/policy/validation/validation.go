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

package validation

import (
	"reflect"

	unversionedvalidation "github.com/ttysteale/kubernetes-api/api/unversioned/validation"
	extensionsvalidation "github.com/ttysteale/kubernetes-api/apis/extensions/validation"
	"github.com/ttysteale/kubernetes-api/apis/policy"
	"github.com/ttysteale/kubernetes-api/util/validation/field"
)

func ValidatePodDisruptionBudget(pdb *policy.PodDisruptionBudget) field.ErrorList {
	allErrs := ValidatePodDisruptionBudgetSpec(pdb.Spec, field.NewPath("spec"))
	return allErrs
}

func ValidatePodDisruptionBudgetUpdate(pdb, oldPdb *policy.PodDisruptionBudget) field.ErrorList {
	allErrs := field.ErrorList{}

	restoreGeneration := pdb.Generation
	pdb.Generation = oldPdb.Generation

	if !reflect.DeepEqual(pdb, oldPdb) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec"), "updates to poddisruptionbudget spec are forbidden."))
	}

	pdb.Generation = restoreGeneration
	return allErrs
}

func ValidatePodDisruptionBudgetSpec(spec policy.PodDisruptionBudgetSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, extensionsvalidation.ValidatePositiveIntOrPercent(spec.MinAvailable, fldPath.Child("minAvailable"))...)
	allErrs = append(allErrs, extensionsvalidation.IsNotMoreThan100Percent(spec.MinAvailable, fldPath.Child("minAvailable"))...)
	allErrs = append(allErrs, unversionedvalidation.ValidateLabelSelector(spec.Selector, fldPath.Child("selector"))...)

	return allErrs
}
