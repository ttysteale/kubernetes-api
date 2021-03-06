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

package network

import (
	"testing"

	"github.com/ttysteale/kubernetes-api/apis/componentconfig"
	nettest "github.com/ttysteale/kubernetes-api/kubelet/network/testing"
)

func TestSelectDefaultPlugin(t *testing.T) {
	all_plugins := []NetworkPlugin{}
	plug, err := InitNetworkPlugin(all_plugins, "", nettest.NewFakeHost(nil), componentconfig.HairpinNone, "10.0.0.0/8")
	if err != nil {
		t.Fatalf("Unexpected error in selecting default plugin: %v", err)
	}
	if plug == nil {
		t.Fatalf("Failed to select the default plugin.")
	}
	if plug.Name() != DefaultPluginName {
		t.Errorf("Failed to select the default plugin. Expected %s. Got %s", DefaultPluginName, plug.Name())
	}
}
