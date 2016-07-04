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

package master

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ttysteale/kubernetes-api/api"
	"github.com/ttysteale/kubernetes-api/api/meta"
	"github.com/ttysteale/kubernetes-api/api/rest"
	"github.com/ttysteale/kubernetes-api/api/unversioned"
	apiv1 "github.com/ttysteale/kubernetes-api/api/v1"
	"github.com/ttysteale/kubernetes-api/apimachinery/registered"
	"github.com/ttysteale/kubernetes-api/apis/apps"
	appsapi "github.com/ttysteale/kubernetes-api/apis/apps/v1alpha1"
	"github.com/ttysteale/kubernetes-api/apis/autoscaling"
	autoscalingapiv1 "github.com/ttysteale/kubernetes-api/apis/autoscaling/v1"
	"github.com/ttysteale/kubernetes-api/apis/batch"
	batchapiv1 "github.com/ttysteale/kubernetes-api/apis/batch/v1"
	batchapiv2alpha1 "github.com/ttysteale/kubernetes-api/apis/batch/v2alpha1"
	"github.com/ttysteale/kubernetes-api/apis/certificates"
	certificatesapiv1alpha1 "github.com/ttysteale/kubernetes-api/apis/certificates/v1alpha1"
	"github.com/ttysteale/kubernetes-api/apis/extensions"
	extensionsapiv1beta1 "github.com/ttysteale/kubernetes-api/apis/extensions/v1beta1"
	"github.com/ttysteale/kubernetes-api/apis/policy"
	policyapiv1alpha1 "github.com/ttysteale/kubernetes-api/apis/policy/v1alpha1"
	"github.com/ttysteale/kubernetes-api/apis/rbac"
	rbacapi "github.com/ttysteale/kubernetes-api/apis/rbac/v1alpha1"
	rbacvalidation "github.com/ttysteale/kubernetes-api/apis/rbac/validation"
	"github.com/ttysteale/kubernetes-api/apiserver"
	apiservermetrics "github.com/ttysteale/kubernetes-api/apiserver/metrics"
	"github.com/ttysteale/kubernetes-api/genericapiserver"
	"github.com/ttysteale/kubernetes-api/healthz"
	kubeletclient "github.com/ttysteale/kubernetes-api/kubelet/client"
	"github.com/ttysteale/kubernetes-api/master/ports"
	certificateetcd "github.com/ttysteale/kubernetes-api/registry/certificates/etcd"
	"github.com/ttysteale/kubernetes-api/registry/clusterrole"
	clusterroleetcd "github.com/ttysteale/kubernetes-api/registry/clusterrole/etcd"
	clusterrolepolicybased "github.com/ttysteale/kubernetes-api/registry/clusterrole/policybased"
	"github.com/ttysteale/kubernetes-api/registry/clusterrolebinding"
	clusterrolebindingetcd "github.com/ttysteale/kubernetes-api/registry/clusterrolebinding/etcd"
	clusterrolebindingpolicybased "github.com/ttysteale/kubernetes-api/registry/clusterrolebinding/policybased"
	"github.com/ttysteale/kubernetes-api/registry/componentstatus"
	configmapetcd "github.com/ttysteale/kubernetes-api/registry/configmap/etcd"
	controlleretcd "github.com/ttysteale/kubernetes-api/registry/controller/etcd"
	deploymentetcd "github.com/ttysteale/kubernetes-api/registry/deployment/etcd"
	"github.com/ttysteale/kubernetes-api/registry/endpoint"
	endpointsetcd "github.com/ttysteale/kubernetes-api/registry/endpoint/etcd"
	eventetcd "github.com/ttysteale/kubernetes-api/registry/event/etcd"
	expcontrolleretcd "github.com/ttysteale/kubernetes-api/registry/experimental/controller/etcd"
	"github.com/ttysteale/kubernetes-api/registry/generic"
	ingressetcd "github.com/ttysteale/kubernetes-api/registry/ingress/etcd"
	jobetcd "github.com/ttysteale/kubernetes-api/registry/job/etcd"
	limitrangeetcd "github.com/ttysteale/kubernetes-api/registry/limitrange/etcd"
	"github.com/ttysteale/kubernetes-api/registry/namespace"
	namespaceetcd "github.com/ttysteale/kubernetes-api/registry/namespace/etcd"
	networkpolicyetcd "github.com/ttysteale/kubernetes-api/registry/networkpolicy/etcd"
	"github.com/ttysteale/kubernetes-api/registry/node"
	nodeetcd "github.com/ttysteale/kubernetes-api/registry/node/etcd"
	pvetcd "github.com/ttysteale/kubernetes-api/registry/persistentvolume/etcd"
	pvcetcd "github.com/ttysteale/kubernetes-api/registry/persistentvolumeclaim/etcd"
	petsetetcd "github.com/ttysteale/kubernetes-api/registry/petset/etcd"
	podetcd "github.com/ttysteale/kubernetes-api/registry/pod/etcd"
	poddisruptionbudgetetcd "github.com/ttysteale/kubernetes-api/registry/poddisruptionbudget/etcd"
	pspetcd "github.com/ttysteale/kubernetes-api/registry/podsecuritypolicy/etcd"
	podtemplateetcd "github.com/ttysteale/kubernetes-api/registry/podtemplate/etcd"
	replicasetetcd "github.com/ttysteale/kubernetes-api/registry/replicaset/etcd"
	resourcequotaetcd "github.com/ttysteale/kubernetes-api/registry/resourcequota/etcd"
	"github.com/ttysteale/kubernetes-api/registry/role"
	roleetcd "github.com/ttysteale/kubernetes-api/registry/role/etcd"
	rolepolicybased "github.com/ttysteale/kubernetes-api/registry/role/policybased"
	"github.com/ttysteale/kubernetes-api/registry/rolebinding"
	rolebindingetcd "github.com/ttysteale/kubernetes-api/registry/rolebinding/etcd"
	rolebindingpolicybased "github.com/ttysteale/kubernetes-api/registry/rolebinding/policybased"
	secretetcd "github.com/ttysteale/kubernetes-api/registry/secret/etcd"
	"github.com/ttysteale/kubernetes-api/registry/service"
	etcdallocator "github.com/ttysteale/kubernetes-api/registry/service/allocator/etcd"
	serviceetcd "github.com/ttysteale/kubernetes-api/registry/service/etcd"
	ipallocator "github.com/ttysteale/kubernetes-api/registry/service/ipallocator"
	serviceaccountetcd "github.com/ttysteale/kubernetes-api/registry/serviceaccount/etcd"
	thirdpartyresourceetcd "github.com/ttysteale/kubernetes-api/registry/thirdpartyresource/etcd"
	"github.com/ttysteale/kubernetes-api/registry/thirdpartyresourcedata"
	thirdpartyresourcedataetcd "github.com/ttysteale/kubernetes-api/registry/thirdpartyresourcedata/etcd"
	"github.com/ttysteale/kubernetes-api/runtime"
	"github.com/ttysteale/kubernetes-api/storage"
	etcdmetrics "github.com/ttysteale/kubernetes-api/storage/etcd/metrics"
	etcdutil "github.com/ttysteale/kubernetes-api/storage/etcd/util"
	"github.com/ttysteale/kubernetes-api/util/wait"

	daemonetcd "github.com/ttysteale/kubernetes-api/registry/daemonset/etcd"
	horizontalpodautoscaleretcd "github.com/ttysteale/kubernetes-api/registry/horizontalpodautoscaler/etcd"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ttysteale/kubernetes-api/registry/service/allocator"
	"github.com/ttysteale/kubernetes-api/registry/service/portallocator"
)

const (
	// DefaultEndpointReconcilerInterval is the default amount of time for how often the endpoints for
	// the kubernetes Service are reconciled.
	DefaultEndpointReconcilerInterval = 10 * time.Second
)

type Config struct {
	*genericapiserver.Config

	EnableCoreControllers    bool
	EndpointReconcilerConfig EndpointReconcilerConfig
	DeleteCollectionWorkers  int
	EventTTL                 time.Duration
	KubeletClient            kubeletclient.KubeletClient
	// Used to start and monitor tunneling
	Tunneler genericapiserver.Tunneler

	disableThirdPartyControllerForTesting bool
}

// EndpointReconcilerConfig holds the endpoint reconciler and endpoint reconciliation interval to be
// used by the master.
type EndpointReconcilerConfig struct {
	Reconciler EndpointReconciler
	Interval   time.Duration
}

// Master contains state for a Kubernetes cluster master/api server.
type Master struct {
	*genericapiserver.GenericAPIServer

	// Map of v1 resources to their REST storages.
	v1ResourcesStorage map[string]rest.Storage

	enableCoreControllers   bool
	deleteCollectionWorkers int
	// registries are internal client APIs for accessing the storage layer
	// TODO: define the internal typed interface in a way that clients can
	// also be replaced
	nodeRegistry              node.Registry
	namespaceRegistry         namespace.Registry
	serviceRegistry           service.Registry
	endpointRegistry          endpoint.Registry
	serviceClusterIPAllocator service.RangeRegistry
	serviceNodePortAllocator  service.RangeRegistry

	// storage for third party objects
	thirdPartyStorage storage.Interface
	// map from api path to a tuple of (storage for the objects, APIGroup)
	thirdPartyResources map[string]thirdPartyEntry
	// protects the map
	thirdPartyResourcesLock sync.RWMutex
	// Useful for reliable testing.  Shouldn't be used otherwise.
	disableThirdPartyControllerForTesting bool

	// Used to start and monitor tunneling
	tunneler genericapiserver.Tunneler
}

// thirdPartyEntry combines objects storage and API group into one struct
// for easy lookup.
type thirdPartyEntry struct {
	storage *thirdpartyresourcedataetcd.REST
	group   unversioned.APIGroup
}

// New returns a new instance of Master from the given config.
// Certain config fields will be set to a default value if unset.
// Certain config fields must be specified, including:
//   KubeletClient
func New(c *Config) (*Master, error) {
	if c.KubeletClient == nil {
		return nil, fmt.Errorf("Master.New() called with config.KubeletClient == nil")
	}

	s, err := genericapiserver.New(c.Config)
	if err != nil {
		return nil, err
	}

	m := &Master{
		GenericAPIServer:        s,
		enableCoreControllers:   c.EnableCoreControllers,
		deleteCollectionWorkers: c.DeleteCollectionWorkers,
		tunneler:                c.Tunneler,

		disableThirdPartyControllerForTesting: c.disableThirdPartyControllerForTesting,
	}
	m.InstallAPIs(c)

	// TODO: Attempt clean shutdown?
	if m.enableCoreControllers {
		m.NewBootstrapController(c.EndpointReconcilerConfig).Start()
	}

	return m, nil
}

var defaultMetricsHandler = prometheus.Handler().ServeHTTP

// MetricsWithReset is a handler that resets metrics when DELETE is passed to the endpoint.
func MetricsWithReset(w http.ResponseWriter, req *http.Request) {
	if req.Method == "DELETE" {
		apiservermetrics.Reset()
		etcdmetrics.Reset()
		io.WriteString(w, "metrics reset\n")
		return
	}
	defaultMetricsHandler(w, req)
}

func (m *Master) InstallAPIs(c *Config) {
	apiGroupsInfo := []genericapiserver.APIGroupInfo{}

	// Install v1 unless disabled.
	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(apiv1.SchemeGroupVersion) {
		// Install v1 API.
		m.initV1ResourcesStorage(c)
		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *registered.GroupOrDie(api.GroupName),
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1": m.v1ResourcesStorage,
			},
			IsLegacyGroup:        true,
			Scheme:               api.Scheme,
			ParameterCodec:       api.ParameterCodec,
			NegotiatedSerializer: api.Codecs,
		}
		if autoscalingGroupVersion := (unversioned.GroupVersion{Group: "autoscaling", Version: "v1"}); registered.IsEnabledVersion(autoscalingGroupVersion) {
			apiGroupInfo.SubresourceGroupVersionKind = map[string]unversioned.GroupVersionKind{
				"replicationcontrollers/scale": autoscalingGroupVersion.WithKind("Scale"),
			}
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	// Run the tunneler.
	healthzChecks := []healthz.HealthzChecker{}
	if m.tunneler != nil {
		m.tunneler.Run(m.getNodeAddresses)
		healthzChecks = append(healthzChecks, healthz.NamedCheck("SSH Tunnel Check", m.IsTunnelSyncHealthy))
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "apiserver_proxy_tunnel_sync_latency_secs",
			Help: "The time since the last successful synchronization of the SSH tunnels for proxy requests.",
		}, func() float64 { return float64(m.tunneler.SecondsSinceSync()) })
	}
	healthz.InstallHandler(m.MuxHelper, healthzChecks...)

	if c.EnableProfiling {
		m.MuxHelper.HandleFunc("/metrics", MetricsWithReset)
	} else {
		m.MuxHelper.HandleFunc("/metrics", defaultMetricsHandler)
	}

	// Install extensions unless disabled.
	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(extensionsapiv1beta1.SchemeGroupVersion) {
		var err error
		m.thirdPartyStorage, err = c.StorageFactory.New(extensions.Resource("thirdpartyresources"))
		if err != nil {
			glog.Fatalf("Error getting third party storage: %v", err)
		}
		m.thirdPartyResources = map[string]thirdPartyEntry{}

		extensionResources := m.getExtensionResources(c)
		extensionsGroupMeta := registered.GroupOrDie(extensions.GroupName)

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *extensionsGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1beta1": extensionResources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	// Install autoscaling unless disabled.
	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(autoscalingapiv1.SchemeGroupVersion) {
		autoscalingResources := m.getAutoscalingResources(c)
		autoscalingGroupMeta := registered.GroupOrDie(autoscaling.GroupName)

		// Hard code preferred group version to autoscaling/v1
		autoscalingGroupMeta.GroupVersion = autoscalingapiv1.SchemeGroupVersion

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *autoscalingGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1": autoscalingResources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	// Install batch unless disabled.
	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(batchapiv1.SchemeGroupVersion) ||
		c.APIResourceConfigSource.AnyResourcesForVersionEnabled(batchapiv2alpha1.SchemeGroupVersion) {
		batchv1Resources := m.getBatchResources(c, batchapiv1.SchemeGroupVersion)
		batchGroupMeta := registered.GroupOrDie(batch.GroupName)

		// Hard code preferred group version to batch/v1
		batchGroupMeta.GroupVersion = batchapiv1.SchemeGroupVersion

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *batchGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1": batchv1Resources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(batchapiv2alpha1.SchemeGroupVersion) {
			batchv2alpha1Resources := m.getBatchResources(c, batchapiv2alpha1.SchemeGroupVersion)
			apiGroupInfo.VersionedResourcesStorageMap["v2alpha1"] = batchv2alpha1Resources
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(policyapiv1alpha1.SchemeGroupVersion) {
		policyResources := m.getPolicyResources(c)
		policyGroupMeta := registered.GroupOrDie(policy.GroupName)

		// Hard code preferred group version to policy/v1alpha1
		policyGroupMeta.GroupVersion = policyapiv1alpha1.SchemeGroupVersion

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *policyGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1alpha1": policyResources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(appsapi.SchemeGroupVersion) {
		appsResources := m.getAppsResources(c)
		appsGroupMeta := registered.GroupOrDie(apps.GroupName)

		// Hard code preferred group version to apps/v1alpha1
		appsGroupMeta.GroupVersion = appsapi.SchemeGroupVersion

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *appsGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1alpha1": appsResources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(certificatesapiv1alpha1.SchemeGroupVersion) {
		certificateResources := m.getCertificateResources(c)
		certificatesGroupMeta := registered.GroupOrDie(certificates.GroupName)

		// Hard code preferred group version to certificates/v1alpha1
		certificatesGroupMeta.GroupVersion = certificatesapiv1alpha1.SchemeGroupVersion

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *certificatesGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1alpha1": certificateResources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	if c.APIResourceConfigSource.AnyResourcesForVersionEnabled(rbacapi.SchemeGroupVersion) {
		rbacResources := m.getRBACResources(c)
		rbacGroupMeta := registered.GroupOrDie(rbac.GroupName)

		// Hard code preferred group version to rbac/v1alpha1
		rbacGroupMeta.GroupVersion = rbacapi.SchemeGroupVersion

		apiGroupInfo := genericapiserver.APIGroupInfo{
			GroupMeta: *rbacGroupMeta,
			VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
				"v1alpha1": rbacResources,
			},
			OptionsExternalVersion: &registered.GroupOrDie(api.GroupName).GroupVersion,
			Scheme:                 api.Scheme,
			ParameterCodec:         api.ParameterCodec,
			NegotiatedSerializer:   api.Codecs,
		}
		apiGroupsInfo = append(apiGroupsInfo, apiGroupInfo)
	}

	if err := m.InstallAPIGroups(apiGroupsInfo); err != nil {
		glog.Fatalf("Error in registering group versions: %v", err)
	}
}

func (m *Master) initV1ResourcesStorage(c *Config) {
	restOptions := func(resource string) generic.RESTOptions {
		return m.GetRESTOptionsOrDie(c, api.Resource(resource))
	}

	podTemplateStorage := podtemplateetcd.NewREST(restOptions("podTemplates"))

	eventStorage := eventetcd.NewREST(restOptions("events"), uint64(c.EventTTL.Seconds()))
	limitRangeStorage := limitrangeetcd.NewREST(restOptions("limitRanges"))

	resourceQuotaStorage, resourceQuotaStatusStorage := resourcequotaetcd.NewREST(restOptions("resourceQuotas"))
	secretStorage := secretetcd.NewREST(restOptions("secrets"))
	serviceAccountStorage := serviceaccountetcd.NewREST(restOptions("serviceAccounts"))
	persistentVolumeStorage, persistentVolumeStatusStorage := pvetcd.NewREST(restOptions("persistentVolumes"))
	persistentVolumeClaimStorage, persistentVolumeClaimStatusStorage := pvcetcd.NewREST(restOptions("persistentVolumeClaims"))
	configMapStorage := configmapetcd.NewREST(restOptions("configMaps"))

	namespaceStorage, namespaceStatusStorage, namespaceFinalizeStorage := namespaceetcd.NewREST(restOptions("namespaces"))
	m.namespaceRegistry = namespace.NewRegistry(namespaceStorage)

	endpointsStorage := endpointsetcd.NewREST(restOptions("endpoints"))
	m.endpointRegistry = endpoint.NewRegistry(endpointsStorage)

	nodeStorage := nodeetcd.NewStorage(restOptions("nodes"), c.KubeletClient, m.ProxyTransport)
	m.nodeRegistry = node.NewRegistry(nodeStorage.Node)

	podStorage := podetcd.NewStorage(
		restOptions("pods"),
		kubeletclient.ConnectionInfoGetter(nodeStorage.Node),
		m.ProxyTransport,
	)

	serviceRESTStorage, serviceStatusStorage := serviceetcd.NewREST(restOptions("services"))
	m.serviceRegistry = service.NewRegistry(serviceRESTStorage)

	var serviceClusterIPRegistry service.RangeRegistry
	serviceClusterIPRange := m.ServiceClusterIPRange
	if serviceClusterIPRange == nil {
		glog.Fatalf("service clusterIPRange is nil")
		return
	}

	serviceStorage, err := c.StorageFactory.New(api.Resource("services"))
	if err != nil {
		glog.Fatal(err.Error())
	}

	serviceClusterIPAllocator := ipallocator.NewAllocatorCIDRRange(serviceClusterIPRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewAllocationMap(max, rangeSpec)
		// TODO etcdallocator package to return a storage interface via the storageFactory
		etcd := etcdallocator.NewEtcd(mem, "/ranges/serviceips", api.Resource("serviceipallocations"), serviceStorage)
		serviceClusterIPRegistry = etcd
		return etcd
	})
	m.serviceClusterIPAllocator = serviceClusterIPRegistry

	var serviceNodePortRegistry service.RangeRegistry
	serviceNodePortAllocator := portallocator.NewPortAllocatorCustom(m.ServiceNodePortRange, func(max int, rangeSpec string) allocator.Interface {
		mem := allocator.NewAllocationMap(max, rangeSpec)
		// TODO etcdallocator package to return a storage interface via the storageFactory
		etcd := etcdallocator.NewEtcd(mem, "/ranges/servicenodeports", api.Resource("servicenodeportallocations"), serviceStorage)
		serviceNodePortRegistry = etcd
		return etcd
	})
	m.serviceNodePortAllocator = serviceNodePortRegistry

	controllerStorage := controlleretcd.NewStorage(restOptions("replicationControllers"))

	serviceRest := service.NewStorage(m.serviceRegistry, m.endpointRegistry, serviceClusterIPAllocator, serviceNodePortAllocator, m.ProxyTransport)

	// TODO: Factor out the core API registration
	m.v1ResourcesStorage = map[string]rest.Storage{
		"pods":             podStorage.Pod,
		"pods/attach":      podStorage.Attach,
		"pods/status":      podStorage.Status,
		"pods/log":         podStorage.Log,
		"pods/exec":        podStorage.Exec,
		"pods/portforward": podStorage.PortForward,
		"pods/proxy":       podStorage.Proxy,
		"pods/binding":     podStorage.Binding,
		"bindings":         podStorage.Binding,

		"podTemplates": podTemplateStorage,

		"replicationControllers":        controllerStorage.Controller,
		"replicationControllers/status": controllerStorage.Status,

		"services":        serviceRest.Service,
		"services/proxy":  serviceRest.Proxy,
		"services/status": serviceStatusStorage,

		"endpoints": endpointsStorage,

		"nodes":        nodeStorage.Node,
		"nodes/status": nodeStorage.Status,
		"nodes/proxy":  nodeStorage.Proxy,

		"events": eventStorage,

		"limitRanges":                   limitRangeStorage,
		"resourceQuotas":                resourceQuotaStorage,
		"resourceQuotas/status":         resourceQuotaStatusStorage,
		"namespaces":                    namespaceStorage,
		"namespaces/status":             namespaceStatusStorage,
		"namespaces/finalize":           namespaceFinalizeStorage,
		"secrets":                       secretStorage,
		"serviceAccounts":               serviceAccountStorage,
		"persistentVolumes":             persistentVolumeStorage,
		"persistentVolumes/status":      persistentVolumeStatusStorage,
		"persistentVolumeClaims":        persistentVolumeClaimStorage,
		"persistentVolumeClaims/status": persistentVolumeClaimStatusStorage,
		"configMaps":                    configMapStorage,

		"componentStatuses": componentstatus.NewStorage(func() map[string]apiserver.Server { return m.getServersToValidate(c) }),
	}
	if registered.IsEnabledVersion(unversioned.GroupVersion{Group: "autoscaling", Version: "v1"}) {
		m.v1ResourcesStorage["replicationControllers/scale"] = controllerStorage.Scale
	}
}

// NewBootstrapController returns a controller for watching the core capabilities of the master.  If
// endpointReconcilerConfig.Interval is 0, the default value of DefaultEndpointReconcilerInterval
// will be used instead.  If endpointReconcilerConfig.Reconciler is nil, the default
// MasterCountEndpointReconciler will be used.
func (m *Master) NewBootstrapController(endpointReconcilerConfig EndpointReconcilerConfig) *Controller {
	if endpointReconcilerConfig.Interval == 0 {
		endpointReconcilerConfig.Interval = DefaultEndpointReconcilerInterval
	}

	if endpointReconcilerConfig.Reconciler == nil {
		// use a default endpoint	reconciler if nothing is set
		// m.endpointRegistry is set via m.InstallAPIs -> m.initV1ResourcesStorage
		endpointReconcilerConfig.Reconciler = NewMasterCountEndpointReconciler(m.MasterCount, m.endpointRegistry)
	}

	return &Controller{
		NamespaceRegistry: m.namespaceRegistry,
		ServiceRegistry:   m.serviceRegistry,

		EndpointReconciler: endpointReconcilerConfig.Reconciler,
		EndpointInterval:   endpointReconcilerConfig.Interval,

		SystemNamespaces:         []string{api.NamespaceSystem},
		SystemNamespacesInterval: 1 * time.Minute,

		ServiceClusterIPRegistry: m.serviceClusterIPAllocator,
		ServiceClusterIPRange:    m.ServiceClusterIPRange,
		ServiceClusterIPInterval: 3 * time.Minute,

		ServiceNodePortRegistry: m.serviceNodePortAllocator,
		ServiceNodePortRange:    m.ServiceNodePortRange,
		ServiceNodePortInterval: 3 * time.Minute,

		PublicIP: m.ClusterIP,

		ServiceIP:                 m.ServiceReadWriteIP,
		ServicePort:               m.ServiceReadWritePort,
		ExtraServicePorts:         m.ExtraServicePorts,
		ExtraEndpointPorts:        m.ExtraEndpointPorts,
		PublicServicePort:         m.PublicReadWritePort,
		KubernetesServiceNodePort: m.KubernetesServiceNodePort,
	}
}

func (m *Master) getServersToValidate(c *Config) map[string]apiserver.Server {
	serversToValidate := map[string]apiserver.Server{
		"controller-manager": {Addr: "127.0.0.1", Port: ports.ControllerManagerPort, Path: "/healthz"},
		"scheduler":          {Addr: "127.0.0.1", Port: ports.SchedulerPort, Path: "/healthz"},
	}

	for ix, machine := range c.StorageFactory.Backends() {
		etcdUrl, err := url.Parse(machine)
		if err != nil {
			glog.Errorf("Failed to parse etcd url for validation: %v", err)
			continue
		}
		var port int
		var addr string
		if strings.Contains(etcdUrl.Host, ":") {
			var portString string
			addr, portString, err = net.SplitHostPort(etcdUrl.Host)
			if err != nil {
				glog.Errorf("Failed to split host/port: %s (%v)", etcdUrl.Host, err)
				continue
			}
			port, _ = strconv.Atoi(portString)
		} else {
			addr = etcdUrl.Host
			port = 4001
		}
		// TODO: etcd health checking should be abstracted in the storage tier
		serversToValidate[fmt.Sprintf("etcd-%d", ix)] = apiserver.Server{
			Addr:        addr,
			EnableHTTPS: etcdUrl.Scheme == "https",
			Port:        port,
			Path:        "/health",
			Validate:    etcdutil.EtcdHealthCheck,
		}
	}
	return serversToValidate
}

// HasThirdPartyResource returns true if a particular third party resource currently installed.
func (m *Master) HasThirdPartyResource(rsrc *extensions.ThirdPartyResource) (bool, error) {
	_, group, err := thirdpartyresourcedata.ExtractApiGroupAndKind(rsrc)
	if err != nil {
		return false, err
	}
	path := makeThirdPartyPath(group)
	services := m.HandlerContainer.RegisteredWebServices()
	for ix := range services {
		if services[ix].RootPath() == path {
			return true, nil
		}
	}
	return false, nil
}

func (m *Master) removeThirdPartyStorage(path string) error {
	m.thirdPartyResourcesLock.Lock()
	defer m.thirdPartyResourcesLock.Unlock()
	storage, found := m.thirdPartyResources[path]
	if found {
		if err := m.removeAllThirdPartyResources(storage.storage); err != nil {
			return err
		}
		delete(m.thirdPartyResources, path)
		m.RemoveAPIGroupForDiscovery(getThirdPartyGroupName(path))
	}
	return nil
}

// RemoveThirdPartyResource removes all resources matching `path`.  Also deletes any stored data
func (m *Master) RemoveThirdPartyResource(path string) error {
	if err := m.removeThirdPartyStorage(path); err != nil {
		return err
	}

	services := m.HandlerContainer.RegisteredWebServices()
	for ix := range services {
		root := services[ix].RootPath()
		if root == path || strings.HasPrefix(root, path+"/") {
			m.HandlerContainer.Remove(services[ix])
		}
	}
	return nil
}

func (m *Master) removeAllThirdPartyResources(registry *thirdpartyresourcedataetcd.REST) error {
	ctx := api.NewDefaultContext()
	existingData, err := registry.List(ctx, nil)
	if err != nil {
		return err
	}
	list, ok := existingData.(*extensions.ThirdPartyResourceDataList)
	if !ok {
		return fmt.Errorf("expected a *ThirdPartyResourceDataList, got %#v", list)
	}
	for ix := range list.Items {
		item := &list.Items[ix]
		if _, err := registry.Delete(ctx, item.Name, nil); err != nil {
			return err
		}
	}
	return nil
}

// ListThirdPartyResources lists all currently installed third party resources
func (m *Master) ListThirdPartyResources() []string {
	m.thirdPartyResourcesLock.RLock()
	defer m.thirdPartyResourcesLock.RUnlock()
	result := []string{}
	for key := range m.thirdPartyResources {
		result = append(result, key)
	}
	return result
}

func (m *Master) addThirdPartyResourceStorage(path string, storage *thirdpartyresourcedataetcd.REST, apiGroup unversioned.APIGroup) {
	m.thirdPartyResourcesLock.Lock()
	defer m.thirdPartyResourcesLock.Unlock()
	m.thirdPartyResources[path] = thirdPartyEntry{storage, apiGroup}
	m.AddAPIGroupForDiscovery(apiGroup)
}

// InstallThirdPartyResource installs a third party resource specified by 'rsrc'.  When a resource is
// installed a corresponding RESTful resource is added as a valid path in the web service provided by
// the master.
//
// For example, if you install a resource ThirdPartyResource{ Name: "foo.company.com", Versions: {"v1"} }
// then the following RESTful resource is created on the server:
//   http://<host>/apis/company.com/v1/foos/...
func (m *Master) InstallThirdPartyResource(rsrc *extensions.ThirdPartyResource) error {
	kind, group, err := thirdpartyresourcedata.ExtractApiGroupAndKind(rsrc)
	if err != nil {
		return err
	}
	plural, _ := meta.KindToResource(unversioned.GroupVersionKind{
		Group:   group,
		Version: rsrc.Versions[0].Name,
		Kind:    kind,
	})

	thirdparty := m.thirdpartyapi(group, kind, rsrc.Versions[0].Name, plural.Resource)
	if err := thirdparty.InstallREST(m.HandlerContainer); err != nil {
		glog.Fatalf("Unable to setup thirdparty api: %v", err)
	}
	path := makeThirdPartyPath(group)
	groupVersion := unversioned.GroupVersionForDiscovery{
		GroupVersion: group + "/" + rsrc.Versions[0].Name,
		Version:      rsrc.Versions[0].Name,
	}
	apiGroup := unversioned.APIGroup{
		Name:             group,
		Versions:         []unversioned.GroupVersionForDiscovery{groupVersion},
		PreferredVersion: groupVersion,
	}
	apiserver.AddGroupWebService(api.Codecs, m.HandlerContainer, path, apiGroup)

	m.addThirdPartyResourceStorage(path, thirdparty.Storage[plural.Resource].(*thirdpartyresourcedataetcd.REST), apiGroup)
	apiserver.InstallServiceErrorHandler(api.Codecs, m.HandlerContainer, m.NewRequestInfoResolver(), []string{thirdparty.GroupVersion.String()})
	return nil
}

func (m *Master) thirdpartyapi(group, kind, version, pluralResource string) *apiserver.APIGroupVersion {
	resourceStorage := thirdpartyresourcedataetcd.NewREST(
		generic.RESTOptions{
			Storage:                 m.thirdPartyStorage,
			Decorator:               generic.UndecoratedStorage,
			DeleteCollectionWorkers: m.deleteCollectionWorkers,
		},
		group,
		kind,
	)

	apiRoot := makeThirdPartyPath("")

	storage := map[string]rest.Storage{
		pluralResource: resourceStorage,
	}

	optionsExternalVersion := registered.GroupOrDie(api.GroupName).GroupVersion
	internalVersion := unversioned.GroupVersion{Group: group, Version: runtime.APIVersionInternal}
	externalVersion := unversioned.GroupVersion{Group: group, Version: version}

	return &apiserver.APIGroupVersion{
		Root:                apiRoot,
		GroupVersion:        externalVersion,
		RequestInfoResolver: m.NewRequestInfoResolver(),

		Creater:   thirdpartyresourcedata.NewObjectCreator(group, version, api.Scheme),
		Convertor: api.Scheme,
		Copier:    api.Scheme,
		Typer:     api.Scheme,

		Mapper:                 thirdpartyresourcedata.NewMapper(registered.GroupOrDie(extensions.GroupName).RESTMapper, kind, version, group),
		Linker:                 registered.GroupOrDie(extensions.GroupName).SelfLinker,
		Storage:                storage,
		OptionsExternalVersion: &optionsExternalVersion,

		Serializer:     thirdpartyresourcedata.NewNegotiatedSerializer(api.Codecs, kind, externalVersion, internalVersion),
		ParameterCodec: thirdpartyresourcedata.NewThirdPartyParameterCodec(api.ParameterCodec),

		Context: m.RequestContextMapper,

		MinRequestTimeout: m.MinRequestTimeout,
	}
}

func (m *Master) GetRESTOptionsOrDie(c *Config, resource unversioned.GroupResource) generic.RESTOptions {
	storage, err := c.StorageFactory.New(resource)
	if err != nil {
		glog.Fatalf("Unable to find storage destination for %v, due to %v", resource, err.Error())
	}

	return generic.RESTOptions{
		Storage:                 storage,
		Decorator:               m.StorageDecorator(),
		DeleteCollectionWorkers: m.deleteCollectionWorkers,
	}
}

// getExperimentalResources returns the resources for extensions api
func (m *Master) getExtensionResources(c *Config) map[string]rest.Storage {
	restOptions := func(resource string) generic.RESTOptions {
		return m.GetRESTOptionsOrDie(c, extensions.Resource(resource))
	}

	// TODO update when we support more than one version of this group
	version := extensionsapiv1beta1.SchemeGroupVersion

	storage := map[string]rest.Storage{}

	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("horizontalpodautoscalers")) {
		hpaStorage, hpaStatusStorage := horizontalpodautoscaleretcd.NewREST(restOptions("horizontalpodautoscalers"))
		storage["horizontalpodautoscalers"] = hpaStorage
		storage["horizontalpodautoscalers/status"] = hpaStatusStorage

		controllerStorage := expcontrolleretcd.NewStorage(m.GetRESTOptionsOrDie(c, api.Resource("replicationControllers")))
		storage["replicationcontrollers"] = controllerStorage.ReplicationController
		storage["replicationcontrollers/scale"] = controllerStorage.Scale
	}
	thirdPartyResourceStorage := thirdpartyresourceetcd.NewREST(restOptions("thirdpartyresources"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("thirdpartyresources")) {
		thirdPartyControl := ThirdPartyController{
			master: m,
			thirdPartyResourceRegistry: thirdPartyResourceStorage,
		}
		if !m.disableThirdPartyControllerForTesting {
			go wait.Forever(func() {
				if err := thirdPartyControl.SyncResources(); err != nil {
					glog.Warningf("third party resource sync failed: %v", err)
				}
			}, 10*time.Second)
		}
		storage["thirdpartyresources"] = thirdPartyResourceStorage
	}

	daemonSetStorage, daemonSetStatusStorage := daemonetcd.NewREST(restOptions("daemonsets"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("daemonsets")) {
		storage["daemonsets"] = daemonSetStorage
		storage["daemonsets/status"] = daemonSetStatusStorage
	}
	deploymentStorage := deploymentetcd.NewStorage(restOptions("deployments"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("deployments")) {
		storage["deployments"] = deploymentStorage.Deployment
		storage["deployments/status"] = deploymentStorage.Status
		storage["deployments/rollback"] = deploymentStorage.Rollback
		storage["deployments/scale"] = deploymentStorage.Scale
	}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("jobs")) {
		jobsStorage, jobsStatusStorage := jobetcd.NewREST(restOptions("jobs"))
		storage["jobs"] = jobsStorage
		storage["jobs/status"] = jobsStatusStorage
	}
	ingressStorage, ingressStatusStorage := ingressetcd.NewREST(restOptions("ingresses"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("ingresses")) {
		storage["ingresses"] = ingressStorage
		storage["ingresses/status"] = ingressStatusStorage
	}
	podSecurityPolicyStorage := pspetcd.NewREST(restOptions("podsecuritypolicy"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("podsecuritypolicy")) {
		storage["podSecurityPolicies"] = podSecurityPolicyStorage
	}
	replicaSetStorage := replicasetetcd.NewStorage(restOptions("replicasets"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("replicasets")) {
		storage["replicasets"] = replicaSetStorage.ReplicaSet
		storage["replicasets/status"] = replicaSetStorage.Status
		storage["replicasets/scale"] = replicaSetStorage.Scale
	}
	networkPolicyStorage := networkpolicyetcd.NewREST(restOptions("networkpolicies"))
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("networkpolicies")) {
		storage["networkpolicies"] = networkPolicyStorage
	}

	return storage
}

// getAutoscalingResources returns the resources for autoscaling api
func (m *Master) getAutoscalingResources(c *Config) map[string]rest.Storage {
	// TODO update when we support more than one version of this group
	version := autoscalingapiv1.SchemeGroupVersion

	storage := map[string]rest.Storage{}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("horizontalpodautoscalers")) {
		hpaStorage, hpaStatusStorage := horizontalpodautoscaleretcd.NewREST(m.GetRESTOptionsOrDie(c, autoscaling.Resource("horizontalpodautoscalers")))
		storage["horizontalpodautoscalers"] = hpaStorage
		storage["horizontalpodautoscalers/status"] = hpaStatusStorage
	}
	return storage
}

// getCertificateResources returns the resources for certificates API
func (m *Master) getCertificateResources(c *Config) map[string]rest.Storage {
	restOptions := func(resource string) generic.RESTOptions {
		return m.GetRESTOptionsOrDie(c, certificates.Resource(resource))
	}

	// TODO update when we support more than one version of this group
	version := certificatesapiv1alpha1.SchemeGroupVersion

	storage := map[string]rest.Storage{}

	csrStorage, csrStatusStorage, csrApprovalStorage := certificateetcd.NewREST(restOptions("certificatesigningrequests"))

	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("certificatesigningrequests")) {
		storage["certificatesigningrequests"] = csrStorage
		storage["certificatesigningrequests/status"] = csrStatusStorage
		storage["certificatesigningrequests/approval"] = csrApprovalStorage
	}

	return storage
}

// getBatchResources returns the resources for batch api
func (m *Master) getBatchResources(c *Config, version unversioned.GroupVersion) map[string]rest.Storage {
	storage := map[string]rest.Storage{}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("jobs")) {
		jobsStorage, jobsStatusStorage := jobetcd.NewREST(m.GetRESTOptionsOrDie(c, batch.Resource("jobs")))
		storage["jobs"] = jobsStorage
		storage["jobs/status"] = jobsStatusStorage
	}
	return storage
}

// getPolicyResources returns the resources for policy api
func (m *Master) getPolicyResources(c *Config) map[string]rest.Storage {
	// TODO update when we support more than one version of this group
	version := policyapiv1alpha1.SchemeGroupVersion

	storage := map[string]rest.Storage{}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("poddisruptionbudgets")) {
		poddisruptionbudgetStorage, poddisruptionbudgetStatusStorage := poddisruptionbudgetetcd.NewREST(m.GetRESTOptionsOrDie(c, policy.Resource("poddisruptionbudgets")))
		storage["poddisruptionbudgets"] = poddisruptionbudgetStorage
		storage["poddisruptionbudgets/status"] = poddisruptionbudgetStatusStorage
	}
	return storage
}

// getAppsResources returns the resources for apps api
func (m *Master) getAppsResources(c *Config) map[string]rest.Storage {
	// TODO update when we support more than one version of this group
	version := appsapi.SchemeGroupVersion

	storage := map[string]rest.Storage{}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("petsets")) {
		petsetStorage, petsetStatusStorage := petsetetcd.NewREST(m.GetRESTOptionsOrDie(c, apps.Resource("petsets")))
		storage["petsets"] = petsetStorage
		storage["petsets/status"] = petsetStatusStorage
	}
	return storage
}

func (m *Master) getRBACResources(c *Config) map[string]rest.Storage {
	version := rbacapi.SchemeGroupVersion

	once := new(sync.Once)
	var authorizationRuleResolver rbacvalidation.AuthorizationRuleResolver
	newRuleValidator := func() rbacvalidation.AuthorizationRuleResolver {
		once.Do(func() {
			authorizationRuleResolver = rbacvalidation.NewDefaultRuleResolver(
				role.NewRegistry(roleetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("roles")))),
				rolebinding.NewRegistry(rolebindingetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("rolebindings")))),
				clusterrole.NewRegistry(clusterroleetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("clusterroles")))),
				clusterrolebinding.NewRegistry(clusterrolebindingetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("clusterrolebindings")))),
			)
		})
		return authorizationRuleResolver
	}

	storage := map[string]rest.Storage{}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("roles")) {
		rolesStorage := roleetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("roles")))
		storage["roles"] = rolepolicybased.NewStorage(rolesStorage, newRuleValidator(), c.AuthorizerRBACSuperUser)
	}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("rolebindings")) {
		roleBindingsStorage := rolebindingetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("rolebindings")))
		storage["rolebindings"] = rolebindingpolicybased.NewStorage(roleBindingsStorage, newRuleValidator(), c.AuthorizerRBACSuperUser)
	}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("clusterroles")) {
		clusterRolesStorage := clusterroleetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("clusterroles")))
		storage["clusterroles"] = clusterrolepolicybased.NewStorage(clusterRolesStorage, newRuleValidator(), c.AuthorizerRBACSuperUser)
	}
	if c.APIResourceConfigSource.ResourceEnabled(version.WithResource("clusterrolebindings")) {
		clusterRoleBindingsStorage := clusterrolebindingetcd.NewREST(m.GetRESTOptionsOrDie(c, rbac.Resource("clusterrolebindings")))
		storage["clusterrolebindings"] = clusterrolebindingpolicybased.NewStorage(clusterRoleBindingsStorage, newRuleValidator(), c.AuthorizerRBACSuperUser)
	}
	return storage
}

// findExternalAddress returns ExternalIP of provided node with fallback to LegacyHostIP.
func findExternalAddress(node *api.Node) (string, error) {
	var fallback string
	for ix := range node.Status.Addresses {
		addr := &node.Status.Addresses[ix]
		if addr.Type == api.NodeExternalIP {
			return addr.Address, nil
		}
		if fallback == "" && addr.Type == api.NodeLegacyHostIP {
			fallback = addr.Address
		}
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("Couldn't find external address: %v", node)
}

func (m *Master) getNodeAddresses() ([]string, error) {
	nodes, err := m.nodeRegistry.ListNodes(api.NewDefaultContext(), nil)
	if err != nil {
		return nil, err
	}
	addrs := []string{}
	for ix := range nodes.Items {
		node := &nodes.Items[ix]
		addr, err := findExternalAddress(node)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func (m *Master) IsTunnelSyncHealthy(req *http.Request) error {
	if m.tunneler == nil {
		return nil
	}
	lag := m.tunneler.SecondsSinceSync()
	if lag > 600 {
		return fmt.Errorf("Tunnel sync is taking to long: %d", lag)
	}
	sshKeyLag := m.tunneler.SecondsSinceSSHKeySync()
	if sshKeyLag > 600 {
		return fmt.Errorf("SSHKey sync is taking to long: %d", sshKeyLag)
	}
	return nil
}

func DefaultAPIResourceConfigSource() *genericapiserver.ResourceConfig {
	ret := genericapiserver.NewResourceConfig()
	ret.EnableVersions(apiv1.SchemeGroupVersion, extensionsapiv1beta1.SchemeGroupVersion, batchapiv1.SchemeGroupVersion, autoscalingapiv1.SchemeGroupVersion, appsapi.SchemeGroupVersion, policyapiv1alpha1.SchemeGroupVersion, rbacapi.SchemeGroupVersion, certificatesapiv1alpha1.SchemeGroupVersion)

	// all extensions resources except these are disabled by default
	ret.EnableResources(
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("daemonsets"),
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("deployments"),
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("horizontalpodautoscalers"),
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("ingresses"),
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("jobs"),
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("replicasets"),
		extensionsapiv1beta1.SchemeGroupVersion.WithResource("thirdpartyresources"),
	)

	return ret
}
