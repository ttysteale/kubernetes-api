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

package lifecycle

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/ttysteale/kubernetes-api/api"
	kubecontainer "github.com/ttysteale/kubernetes-api/kubelet/container"
	kubetypes "github.com/ttysteale/kubernetes-api/kubelet/types"
	"github.com/ttysteale/kubernetes-api/kubelet/util/format"
	"github.com/ttysteale/kubernetes-api/kubelet/util/ioutils"
	"github.com/ttysteale/kubernetes-api/types"
	"github.com/ttysteale/kubernetes-api/util/intstr"
)

type HandlerRunner struct {
	httpGetter       kubetypes.HttpGetter
	commandRunner    kubecontainer.ContainerCommandRunner
	containerManager podStatusProvider
}

type podStatusProvider interface {
	GetPodStatus(uid types.UID, name, namespace string) (*kubecontainer.PodStatus, error)
}

func NewHandlerRunner(httpGetter kubetypes.HttpGetter, commandRunner kubecontainer.ContainerCommandRunner, containerManager podStatusProvider) kubecontainer.HandlerRunner {
	return &HandlerRunner{
		httpGetter:       httpGetter,
		commandRunner:    commandRunner,
		containerManager: containerManager,
	}
}

func (hr *HandlerRunner) Run(containerID kubecontainer.ContainerID, pod *api.Pod, container *api.Container, handler *api.Handler) (string, error) {
	switch {
	case handler.Exec != nil:
		var (
			buffer bytes.Buffer
			msg    string
		)
		output := ioutils.WriteCloserWrapper(&buffer)
		err := hr.commandRunner.ExecInContainer(containerID, handler.Exec.Command, nil, output, output, false)
		if err != nil {
			msg := fmt.Sprintf("Exec lifecycle hook (%v) for Container %q in Pod %q failed - %q", handler.Exec.Command, container.Name, format.Pod(pod), buffer.String())
			glog.V(1).Infof(msg)
		}
		return msg, err
	case handler.HTTPGet != nil:
		msg, err := hr.runHTTPHandler(pod, container, handler)
		if err != nil {
			msg := fmt.Sprintf("Http lifecycle hook (%s) for Container %q in Pod %q failed - %q", handler.HTTPGet.Path, container.Name, format.Pod(pod), msg)
			glog.V(1).Infof(msg)
		}
		return msg, err
	default:
		err := fmt.Errorf("Invalid handler: %v", handler)
		msg := fmt.Sprintf("Cannot run handler: %v", err)
		glog.Errorf(msg)
		return msg, err
	}
}

// resolvePort attempts to turn a IntOrString port reference into a concrete port number.
// If portReference has an int value, it is treated as a literal, and simply returns that value.
// If portReference is a string, an attempt is first made to parse it as an integer.  If that fails,
// an attempt is made to find a port with the same name in the container spec.
// If a port with the same name is found, it's ContainerPort value is returned.  If no matching
// port is found, an error is returned.
func resolvePort(portReference intstr.IntOrString, container *api.Container) (int, error) {
	if portReference.Type == intstr.Int {
		return portReference.IntValue(), nil
	}
	portName := portReference.StrVal
	port, err := strconv.Atoi(portName)
	if err == nil {
		return port, nil
	}
	for _, portSpec := range container.Ports {
		if portSpec.Name == portName {
			return int(portSpec.ContainerPort), nil
		}
	}
	return -1, fmt.Errorf("couldn't find port: %v in %v", portReference, container)
}

func (hr *HandlerRunner) runHTTPHandler(pod *api.Pod, container *api.Container, handler *api.Handler) (string, error) {
	host := handler.HTTPGet.Host
	if len(host) == 0 {
		status, err := hr.containerManager.GetPodStatus(pod.UID, pod.Name, pod.Namespace)
		if err != nil {
			glog.Errorf("Unable to get pod info, event handlers may be invalid.")
			return "", err
		}
		if status.IP == "" {
			return "", fmt.Errorf("failed to find networking container: %v", status)
		}
		host = status.IP
	}
	var port int
	if handler.HTTPGet.Port.Type == intstr.String && len(handler.HTTPGet.Port.StrVal) == 0 {
		port = 80
	} else {
		var err error
		port, err = resolvePort(handler.HTTPGet.Port, container)
		if err != nil {
			return "", err
		}
	}
	url := fmt.Sprintf("http://%s/%s", net.JoinHostPort(host, strconv.Itoa(port)), handler.HTTPGet.Path)
	resp, err := hr.httpGetter.Get(url)
	return getHttpRespBody(resp), err
}

func getHttpRespBody(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	defer resp.Body.Close()
	if bytes, err := ioutil.ReadAll(resp.Body); err == nil {
		return string(bytes)
	}
	return ""
}
