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

package unversioned

import (
	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/api/unversioned"
	"github.com/ttysteale/kubernetes-api/apimachinery/registered"
	"github.com/ttysteale/kubernetes-api/apis/batch"
	"github.com/ttysteale/kubernetes-api/apis/batch/v2alpha1"
	"github.com/ttysteale/kubernetes-api/client/restclient"
)

type BatchInterface interface {
	JobsNamespacer
	ScheduledJobsNamespacer
}

// BatchClient is used to interact with Kubernetes batch features.
type BatchClient struct {
	*restclient.RESTClient
}

func (c *BatchClient) Jobs(namespace string) JobInterface {
	return newJobsV1(c, namespace)
}

func (c *BatchClient) ScheduledJobs(namespace string) ScheduledJobInterface {
	return newScheduledJobs(c, namespace)
}

func NewBatch(c *restclient.Config) (*BatchClient, error) {
	config := *c
	if err := setBatchDefaults(&config, nil); err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &BatchClient{client}, nil
}

func NewBatchV2Alpha1(c *restclient.Config) (*BatchClient, error) {
	config := *c
	if err := setBatchDefaults(&config, &v2alpha1.SchemeGroupVersion); err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &BatchClient{client}, nil
}

func NewBatchOrDie(c *restclient.Config) *BatchClient {
	var (
		client *BatchClient
		err    error
	)
	if c.ContentConfig.GroupVersion != nil && *c.ContentConfig.GroupVersion == v2alpha1.SchemeGroupVersion {
		client, err = NewBatchV2Alpha1(c)
	} else {
		client, err = NewBatch(c)
	}
	if err != nil {
		panic(err)
	}
	return client
}

func setBatchDefaults(config *restclient.Config, gv *unversioned.GroupVersion) error {
	// if batch group is not registered, return an error
	g, err := registered.Group(batch.GroupName)
	if err != nil {
		return err
	}
	config.APIPath = defaultAPIPath
	if config.UserAgent == "" {
		config.UserAgent = restclient.DefaultKubernetesUserAgent()
	}
	// TODO: Unconditionally set the config.Version, until we fix the config.
	//if config.Version == "" {
	copyGroupVersion := g.GroupVersion
	if gv != nil {
		copyGroupVersion = *gv
	}
	config.GroupVersion = &copyGroupVersion
	//}

	config.Codec = api.Codecs.LegacyCodec(*config.GroupVersion)
	config.NegotiatedSerializer = api.Codecs
	return nil
}
