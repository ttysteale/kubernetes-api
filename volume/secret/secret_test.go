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

package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/ttysteale/kubernetes-api/api"
	clientset "github.com/ttysteale/kubernetes-api/client/clientset_generated/internalclientset"
	"github.com/ttysteale/kubernetes-api/client/clientset_generated/internalclientset/fake"
	"github.com/ttysteale/kubernetes-api/types"
	"github.com/ttysteale/kubernetes-api/volume"
	"github.com/ttysteale/kubernetes-api/volume/empty_dir"
	volumetest "github.com/ttysteale/kubernetes-api/volume/testing"
	"github.com/ttysteale/kubernetes-api/volume/util"

	"github.com/stretchr/testify/assert"
)

func TestMakePayload(t *testing.T) {
	cases := []struct {
		name     string
		mappings []api.KeyToPath
		secret   *api.Secret
		payload  map[string][]byte
		success  bool
	}{
		{
			name: "no overrides",
			secret: &api.Secret{
				Data: map[string][]byte{
					"foo": []byte("foo"),
					"bar": []byte("bar"),
				},
			},
			payload: map[string][]byte{
				"foo": []byte("foo"),
				"bar": []byte("bar"),
			},
			success: true,
		},
		{
			name: "basic 1",
			mappings: []api.KeyToPath{
				{
					Key:  "foo",
					Path: "path/to/foo.txt",
				},
			},
			secret: &api.Secret{
				Data: map[string][]byte{
					"foo": []byte("foo"),
					"bar": []byte("bar"),
				},
			},
			payload: map[string][]byte{
				"path/to/foo.txt": []byte("foo"),
			},
			success: true,
		},
		{
			name: "subdirs",
			mappings: []api.KeyToPath{
				{
					Key:  "foo",
					Path: "path/to/1/2/3/foo.txt",
				},
			},
			secret: &api.Secret{
				Data: map[string][]byte{
					"foo": []byte("foo"),
					"bar": []byte("bar"),
				},
			},
			payload: map[string][]byte{
				"path/to/1/2/3/foo.txt": []byte("foo"),
			},
			success: true,
		},
		{
			name: "subdirs 2",
			mappings: []api.KeyToPath{
				{
					Key:  "foo",
					Path: "path/to/1/2/3/foo.txt",
				},
			},
			secret: &api.Secret{
				Data: map[string][]byte{
					"foo": []byte("foo"),
					"bar": []byte("bar"),
				},
			},
			payload: map[string][]byte{
				"path/to/1/2/3/foo.txt": []byte("foo"),
			},
			success: true,
		},
		{
			name: "subdirs 3",
			mappings: []api.KeyToPath{
				{
					Key:  "foo",
					Path: "path/to/1/2/3/foo.txt",
				},
				{
					Key:  "bar",
					Path: "another/path/to/the/esteemed/bar.bin",
				},
			},
			secret: &api.Secret{
				Data: map[string][]byte{
					"foo": []byte("foo"),
					"bar": []byte("bar"),
				},
			},
			payload: map[string][]byte{
				"path/to/1/2/3/foo.txt":                []byte("foo"),
				"another/path/to/the/esteemed/bar.bin": []byte("bar"),
			},
			success: true,
		},
		{
			name: "non existent key",
			mappings: []api.KeyToPath{
				{
					Key:  "zab",
					Path: "path/to/foo.txt",
				},
			},
			secret: &api.Secret{
				Data: map[string][]byte{
					"foo": []byte("foo"),
					"bar": []byte("bar"),
				},
			},
			success: false,
		},
	}

	for _, tc := range cases {
		actualPayload, err := makePayload(tc.mappings, tc.secret)
		if err != nil && tc.success {
			t.Errorf("%v: unexpected failure making payload: %v", tc.name, err)
			continue
		}

		if err == nil && !tc.success {
			t.Errorf("%v: unexpected success making payload", tc.name)
			continue
		}

		if !tc.success {
			continue
		}

		if e, a := tc.payload, actualPayload; !reflect.DeepEqual(e, a) {
			t.Errorf("%v: expected and actual payload do not match", tc.name)
		}
	}
}

func newTestHost(t *testing.T, clientset clientset.Interface) (string, volume.VolumeHost) {
	tempDir, err := ioutil.TempDir("/tmp", "secret_volume_test.")
	if err != nil {
		t.Fatalf("can't make a temp rootdir: %v", err)
	}

	return tempDir, volumetest.NewFakeVolumeHost(tempDir, clientset, empty_dir.ProbeVolumePlugins(), "" /* rootContext */)
}

func TestCanSupport(t *testing.T) {
	pluginMgr := volume.VolumePluginMgr{}
	_, host := newTestHost(t, nil)
	pluginMgr.InitPlugins(ProbeVolumePlugins(), host)

	plugin, err := pluginMgr.FindPluginByName(secretPluginName)
	if err != nil {
		t.Errorf("Can't find the plugin by name")
	}
	if plugin.GetPluginName() != secretPluginName {
		t.Errorf("Wrong name: %s", plugin.GetPluginName())
	}
	if !plugin.CanSupport(&volume.Spec{Volume: &api.Volume{VolumeSource: api.VolumeSource{Secret: &api.SecretVolumeSource{SecretName: ""}}}}) {
		t.Errorf("Expected true")
	}
	if plugin.CanSupport(&volume.Spec{}) {
		t.Errorf("Expected false")
	}
}

func TestPlugin(t *testing.T) {
	var (
		testPodUID     = types.UID("test_pod_uid")
		testVolumeName = "test_volume_name"
		testNamespace  = "test_secret_namespace"
		testName       = "test_secret_name"

		volumeSpec    = volumeSpec(testVolumeName, testName)
		secret        = secret(testNamespace, testName)
		client        = fake.NewSimpleClientset(&secret)
		pluginMgr     = volume.VolumePluginMgr{}
		rootDir, host = newTestHost(t, client)
	)

	pluginMgr.InitPlugins(ProbeVolumePlugins(), host)

	plugin, err := pluginMgr.FindPluginByName(secretPluginName)
	if err != nil {
		t.Errorf("Can't find the plugin by name")
	}

	pod := &api.Pod{ObjectMeta: api.ObjectMeta{Namespace: testNamespace, UID: testPodUID}}
	mounter, err := plugin.NewMounter(volume.NewSpecFromVolume(volumeSpec), pod, volume.VolumeOptions{})
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mounter == nil {
		t.Errorf("Got a nil Mounter")
	}

	volumePath := mounter.GetPath()
	if !strings.HasSuffix(volumePath, fmt.Sprintf("pods/test_pod_uid/volumes/kubernetes.io~secret/test_volume_name")) {
		t.Errorf("Got unexpected path: %s", volumePath)
	}

	err = mounter.SetUp(nil)
	if err != nil {
		t.Errorf("Failed to setup volume: %v", err)
	}
	if _, err := os.Stat(volumePath); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("SetUp() failed, volume path not created: %s", volumePath)
		} else {
			t.Errorf("SetUp() failed: %v", err)
		}
	}

	// secret volume should create its own empty wrapper path
	podWrapperMetadataDir := fmt.Sprintf("%v/pods/test_pod_uid/plugins/kubernetes.io~empty-dir/wrapped_test_volume_name", rootDir)

	if _, err := os.Stat(podWrapperMetadataDir); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("SetUp() failed, empty-dir wrapper path is not created: %s", podWrapperMetadataDir)
		} else {
			t.Errorf("SetUp() failed: %v", err)
		}
	}
	doTestSecretDataInVolume(volumePath, secret, t)
	defer doTestCleanAndTeardown(plugin, testPodUID, testVolumeName, volumePath, t)

	// Metrics only supported on linux
	metrics, err := mounter.GetMetrics()
	if runtime.GOOS == "linux" {
		assert.NotEmpty(t, metrics)
		assert.NoError(t, err)
	} else {
		t.Skipf("Volume metrics not supported on %s", runtime.GOOS)
	}
}

// Test the case where the plugin's ready file exists, but the volume dir is not a
// mountpoint, which is the state the system will be in after reboot.  The dir
// should be mounter and the secret data written to it.
func TestPluginReboot(t *testing.T) {
	var (
		testPodUID     = types.UID("test_pod_uid3")
		testVolumeName = "test_volume_name"
		testNamespace  = "test_secret_namespace"
		testName       = "test_secret_name"

		volumeSpec    = volumeSpec(testVolumeName, testName)
		secret        = secret(testNamespace, testName)
		client        = fake.NewSimpleClientset(&secret)
		pluginMgr     = volume.VolumePluginMgr{}
		rootDir, host = newTestHost(t, client)
	)

	pluginMgr.InitPlugins(ProbeVolumePlugins(), host)

	plugin, err := pluginMgr.FindPluginByName(secretPluginName)
	if err != nil {
		t.Errorf("Can't find the plugin by name")
	}

	pod := &api.Pod{ObjectMeta: api.ObjectMeta{Namespace: testNamespace, UID: testPodUID}}
	mounter, err := plugin.NewMounter(volume.NewSpecFromVolume(volumeSpec), pod, volume.VolumeOptions{})
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mounter == nil {
		t.Errorf("Got a nil Mounter")
	}

	podMetadataDir := fmt.Sprintf("%v/pods/test_pod_uid3/plugins/kubernetes.io~secret/test_volume_name", rootDir)
	util.SetReady(podMetadataDir)
	volumePath := mounter.GetPath()
	if !strings.HasSuffix(volumePath, fmt.Sprintf("pods/test_pod_uid3/volumes/kubernetes.io~secret/test_volume_name")) {
		t.Errorf("Got unexpected path: %s", volumePath)
	}

	err = mounter.SetUp(nil)
	if err != nil {
		t.Errorf("Failed to setup volume: %v", err)
	}
	if _, err := os.Stat(volumePath); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("SetUp() failed, volume path not created: %s", volumePath)
		} else {
			t.Errorf("SetUp() failed: %v", err)
		}
	}

	doTestSecretDataInVolume(volumePath, secret, t)
	doTestCleanAndTeardown(plugin, testPodUID, testVolumeName, volumePath, t)
}

func volumeSpec(volumeName, secretName string) *api.Volume {
	return &api.Volume{
		Name: volumeName,
		VolumeSource: api.VolumeSource{
			Secret: &api.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func secret(namespace, name string) api.Secret {
	return api.Secret{
		ObjectMeta: api.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: map[string][]byte{
			"data-1": []byte("value-1"),
			"data-2": []byte("value-2"),
			"data-3": []byte("value-3"),
		},
	}
}

func doTestSecretDataInVolume(volumePath string, secret api.Secret, t *testing.T) {
	for key, value := range secret.Data {
		secretDataHostPath := path.Join(volumePath, key)
		if _, err := os.Stat(secretDataHostPath); err != nil {
			t.Fatalf("SetUp() failed, couldn't find secret data on disk: %v", secretDataHostPath)
		} else {
			actualSecretBytes, err := ioutil.ReadFile(secretDataHostPath)
			if err != nil {
				t.Fatalf("Couldn't read secret data from: %v", secretDataHostPath)
			}

			actualSecretValue := string(actualSecretBytes)
			if string(value) != actualSecretValue {
				t.Errorf("Unexpected value; expected %q, got %q", value, actualSecretValue)
			}
		}
	}
}

func doTestCleanAndTeardown(plugin volume.VolumePlugin, podUID types.UID, testVolumeName, volumePath string, t *testing.T) {
	unmounter, err := plugin.NewUnmounter(testVolumeName, podUID)
	if err != nil {
		t.Errorf("Failed to make a new Unmounter: %v", err)
	}
	if unmounter == nil {
		t.Errorf("Got a nil Unmounter")
	}

	if err := unmounter.TearDown(); err != nil {
		t.Errorf("Expected success, got: %v", err)
	}
	if _, err := os.Stat(volumePath); err == nil {
		t.Errorf("TearDown() failed, volume path still exists: %s", volumePath)
	} else if !os.IsNotExist(err) {
		t.Errorf("SetUp() failed: %v", err)
	}
}
