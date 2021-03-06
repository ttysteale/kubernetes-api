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

package daemonset

import (
	"testing"

	_ "github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/api/testapi"
	apitesting "github.com/ttysteale/kubernetes-api/api/testing"
	"github.com/ttysteale/kubernetes-api/apis/extensions"
	"github.com/ttysteale/kubernetes-api/labels"
)

func TestSelectableFieldLabelConversions(t *testing.T) {
	apitesting.TestSelectableFieldLabelConversionsOfKind(t,
		testapi.Extensions.GroupVersion().String(),
		"DaemonSet",
		labels.Set(DaemonSetToSelectableFields(&extensions.DaemonSet{})),
		nil,
	)
}
