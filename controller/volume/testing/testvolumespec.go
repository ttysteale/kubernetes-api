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

package testing

import (
	"fmt"

	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/client/clientset_generated/internalclientset/fake"
	"github.com/ttysteale/kubernetes-api/client/testing/core"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/types"
	"github.com/ttysteale/kubernetes-api/volume"
	"github.com/ttysteale/kubernetes-api/watch"
)

// GetTestVolumeSpec returns a test volume spec
func GetTestVolumeSpec(volumeName string, diskName api.UniqueVolumeName) *volume.Spec {
	return &volume.Spec{
		Volume: &api.Volume{
			Name: volumeName,
			VolumeSource: api.VolumeSource{
				GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
					PDName:   string(diskName),
					FSType:   "fake",
					ReadOnly: false,
				},
			},
		},
	}
}

func CreateTestClient() *fake.Clientset {
	fakeClient := &fake.Clientset{}

	fakeClient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		obj := &api.PodList{}
		podNamePrefix := "mypod"
		namespace := "mynamespace"
		for i := 0; i < 5; i++ {
			podName := fmt.Sprintf("%s-%d", podNamePrefix, i)
			pod := api.Pod{
				Status: api.PodStatus{
					Phase: api.PodRunning,
				},
				ObjectMeta: api.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					Labels: map[string]string{
						"name": podName,
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "containerName",
							Image: "containerImage",
							VolumeMounts: []api.VolumeMount{
								{
									Name:      "volumeMountName",
									ReadOnly:  false,
									MountPath: "/mnt",
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: "volumeName",
							VolumeSource: api.VolumeSource{
								GCEPersistentDisk: &api.GCEPersistentDiskVolumeSource{
									PDName:   "pdName",
									FSType:   "ext4",
									ReadOnly: false,
								},
							},
						},
					},
				},
			}
			obj.Items = append(obj.Items, pod)
		}
		return true, obj, nil
	})

	fakeWatch := watch.NewFake()
	fakeClient.AddWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))

	return fakeClient
}

// NewPod returns a test pod object
func NewPod(uid, name string) *api.Pod {
	return &api.Pod{
		ObjectMeta: api.ObjectMeta{
			UID:       types.UID(uid),
			Name:      name,
			Namespace: name,
		},
	}
}
