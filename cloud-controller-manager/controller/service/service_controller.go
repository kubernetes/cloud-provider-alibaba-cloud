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

package service

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	kubefeatures "k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/util/metrics"
)

const (
	// Interval of synchronizing service status from apiserver
	serviceSyncPeriod = 30 * time.Second
	// Interval of synchronizing node status from apiserver
	nodeSyncPeriod = 100 * time.Second

	// How long to wait before retrying the processing of a service change.
	// If this changes, the sleep in hack/jenkins/e2e.sh before downing a cluster
	// should be changed appropriately.
	minRetryDelay = 5 * time.Second
	maxRetryDelay = 300 * time.Second

	clientRetryCount    = 5
	clientRetryInterval = 5 * time.Second

	retryable    = true
	notRetryable = false

	doNotRetry = time.Duration(0)

	// LabelNodeRoleMaster specifies that a node is a master
	// It's copied over to kubeadm until it's merged in core: https://github.com/kubernetes/kubernetes/pull/39112
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	// LabelNodeRoleExcludeBalancer specifies that the node should be
	// exclude from load balancers created by a cloud provider.
	LabelNodeRoleExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
)

type cachedService struct {
	backend []*v1.Node
	// The cached state of the service
	state *v1.Service
	// Controls error back-off
	lastRetryDelay time.Duration
	lock           sync.Mutex
}

type serviceCache struct {
	mu         sync.Mutex // protects serviceMap
	serviceMap map[string]*cachedService
}

type ServiceController struct {
	cloud               cloudprovider.Interface
	knownHosts          []*v1.Node
	servicesToUpdate    []*v1.Service
	kubeClient          clientset.Interface
	clusterName         string
	balancer            cloudprovider.LoadBalancer
	cache               *serviceCache
	serviceLister       corelisters.ServiceLister
	serviceListerSynced cache.InformerSynced
	epLister            corelisters.EndpointsLister
	epListerSynced      cache.InformerSynced
	eventBroadcaster    record.EventBroadcaster
	eventRecorder       record.EventRecorder
	nodeLister          corelisters.NodeLister
	nodeListerSynced    cache.InformerSynced
	// services that need to be synced
	workingQueue workqueue.DelayingInterface
	nodesQueue   workqueue.DelayingInterface
}

// New returns a new service controller to keep cloud provider service resources
// (like load balancers) in sync with the registry.
func New(
	cloud cloudprovider.Interface,
	kubeClient clientset.Interface,
	serviceInformer coreinformers.ServiceInformer,
	nodeInformer coreinformers.NodeInformer,
	endpointsInformer coreinformers.EndpointsInformer,
	clusterName string,
) (*ServiceController, error) {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(glog.Infof)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "service-controller"})

	if kubeClient != nil && kubeClient.CoreV1().RESTClient().GetRateLimiter() != nil {
		if err := metrics.RegisterMetricAndTrackRateLimiterUsage("service_controller", kubeClient.CoreV1().RESTClient().GetRateLimiter()); err != nil {
			return nil, err
		}
	}

	s := &ServiceController{
		cloud:            cloud,
		knownHosts:       []*v1.Node{},
		kubeClient:       kubeClient,
		clusterName:      clusterName,
		cache:            &serviceCache{serviceMap: make(map[string]*cachedService)},
		eventBroadcaster: broadcaster,
		eventRecorder:    recorder,
		nodeLister:       nodeInformer.Lister(),
		nodeListerSynced: nodeInformer.Informer().HasSynced,
		epLister:         endpointsInformer.Lister(),
		epListerSynced:   endpointsInformer.Informer().HasSynced,
		workingQueue:     workqueue.NewNamedDelayingQueue("service"),
		nodesQueue:       workqueue.NewNamedDelayingQueue("nodes"),
	}
	endpointsInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(cur interface{}) {
				endpoint := cur.(*v1.Endpoints)
				glog.V(5).Infof("controller: Add event, enpoints [%s/%s]\n", endpoint.Namespace, endpoint.Name)
				s.syncEnpoints(cur)
			},
			UpdateFunc: func(obja, objb interface{}) {
				ep1, ok1 := obja.(*v1.Endpoints)
				ep2, ok2 := objb.(*v1.Endpoints)
				if ok1 && ok2 && !reflect.DeepEqual(ep1.Subsets, ep2.Subsets) {
					glog.V(5).Infof("controller: Update event, endpoints [%s/%s]\n", ep1.Namespace, ep1.Name)
					s.syncEnpoints(ep1)
				}
			},
			DeleteFunc: func(cur interface{}) {
				endpoint := cur.(*v1.Endpoints)
				glog.V(5).Infof("controller: Delete event, endpoints [%s/%s]\n", endpoint.Namespace, endpoint.Name)
				s.syncEnpoints(cur)
			},
		},
		serviceSyncPeriod,
	)
	nodeInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(cur interface{}) {
				node := cur.(*v1.Node)
				glog.V(5).Infof("controller: Add event, nodes [%s/%s]\n", node.Namespace, node.Name)
				s.syncNodes(cur)
			},
			UpdateFunc: func(obja, objb interface{}) {

				node1, ok1 := obja.(*v1.Node)
				node2, ok2 := objb.(*v1.Node)
				if ok1 && ok2 &&
					(!nodeLabelsEqual([]*v1.Node{node1}, []*v1.Node{node2}) ||
						node1.Spec.Unschedulable != node2.Spec.Unschedulable ||
						!nodeConditionEqual(node1.Status.Conditions, node2.Status.Conditions)) {
					// label and schedulable changed .
					// status healthy should be considered
					glog.V(5).Infof("controller: Update event, nodes [%s/%s]\n", node1.Namespace, node1.Name)
					s.syncNodes(node1)
				}
			},
			DeleteFunc: func(cur interface{}) {
				node := cur.(*v1.Node)
				glog.V(5).Infof("controller: Delete event, nodes [%s/%s]\n", node.Namespace, node.Name)
				s.syncNodes(cur)
			},
		},
		serviceSyncPeriod,
	)
	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(cur interface{}) {
				//if !needLoadBalancer(cur.(*v1.Service)) {
				//	key, _ := controller.KeyFunc(cur)
				//	glog.Infof("controller: do not need loadbalancer ,skip. %s\n",key)
				//	return
				//}
				svc := cur.(*v1.Service)
				glog.V(5).Infof("controller: Add event, service [%s/%s]\n", svc.Namespace, svc.Name)
				s.enqueueService(cur)
			},
			UpdateFunc: func(old, cur interface{}) {
				oldSvc, ok1 := old.(*v1.Service)
				curSvc, ok2 := cur.(*v1.Service)
				if ok1 && ok2 && s.needsUpdate(oldSvc, curSvc) {
					glog.V(5).Infof("controller: Update event, service [%s/%s]\n", oldSvc.Namespace, oldSvc.Name)
					s.enqueueService(cur)
				}
			},
			DeleteFunc: func(cur interface{}) {
				svc := cur.(*v1.Service)
				glog.V(5).Infof("controller: Delete event, service [%s/%s]\n", svc.Namespace, svc.Name)
				s.enqueueService(cur)
			},
		},
		serviceSyncPeriod,
	)
	s.serviceLister = serviceInformer.Lister()
	s.serviceListerSynced = serviceInformer.Informer().HasSynced

	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ServiceController) syncNodes(obj interface{}) {

	//glog.V(5).Infof("controller: syncNodes, enqueue node [%s]\n",obj.(*v1.Node).Name)
	errs := s.cache.lockedEach(
		func(svc *cachedService) error {
			if !needLoadBalancer(svc.state) {
				glog.V(6).Infof("syncNodes: service [%s] does not need loadbalancer , skip.\n", svc.state.Name)
				return nil
			}
			glog.V(5).Infof("controller: syncNodes, enqueue service[%s/%s]\n", svc.state.Namespace, svc.state.Name)
			s.enqueueServiceForNodes(svc.state)
			return nil
		},
	)
	if len(errs) > 0 {
		glog.Errorf("sync nodes queue error, %d service backend queued. %s\n", len(s.cache.serviceMap)-len(errs), errorutils.NewAggregate(errs).Error())
	}
}

func (s *ServiceController) syncEnpoints(obj interface{}) {
	ep, ok := obj.(*v1.Endpoints)
	if !ok {
		glog.Errorf("enpoint controller: wrong type [%s], expect v1.Endpoints\n", reflect.TypeOf(obj).String())
		return
	}
	svc, exist, err := s.cache.GetByKey(fmt.Sprintf("%s/%s", ep.Namespace, ep.Name))
	if err != nil || !exist {
		glog.Infof("controller: can not get cached service for endpoints[%s/%s], skip sync endpoint.\n", ep.Namespace, ep.Name)
		return
	}
	service := svc.(*cachedService).state
	if !needLoadBalancer(service) {
		// we are safe here to skip process syncEnpoint.
		glog.V(5).Infof("syncNodes: service [%s] does not need loadbalancer , skip.\n", service.Name)
		return
	}
	glog.V(5).Infof("controller: syncEndpoint, enqueue service [%s/%s]\n", service.Namespace, service.Name)
	s.enqueueServiceForNodes(service)
}

// obj could be an *v1.Service, or a DeletionFinalStateUnknown marker item.
func (s *ServiceController) enqueueService(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %#v: %v", obj, err)
		return
	}
	glog.Infof("working queue add service %s\n", key)
	s.workingQueue.Add(key)
}

// obj could be an *v1.Service, or a DeletionFinalStateUnknown marker item.
func (s *ServiceController) enqueueServiceForNodes(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %#v: %v", obj, err)
		return
	}
	s.nodesQueue.Add(key)
}

// Run starts a background goroutine that watches for changes to services that
// have (or had) LoadBalancers=true and ensures that they have
// load balancers created and deleted appropriately.
// serviceSyncPeriod controls how often we check the cluster's services to
// ensure that the correct load balancers exist.
// nodeSyncPeriod controls how often we check the cluster's nodes to determine
// if load balancers need to be updated to point to a new set.
//
// It's an error to call Run() more than once for a given ServiceController
// object.
func (s *ServiceController) Run(stopCh <-chan struct{}, workers int) {
	defer runtime.HandleCrash()
	defer s.workingQueue.ShutDown()
	defer s.nodesQueue.ShutDown()

	glog.Info("Starting service controller")
	defer glog.Info("Shutting down service controller")

	if !controller.WaitForCacheSync("service", stopCh, s.serviceListerSynced, s.nodeListerSynced) {
		return
	}
	for i := 0; i < workers; i++ {
		go wait.Until(s.worker, time.Second, stopCh)
		go wait.Until(s.nodesWorker, time.Second, stopCh)
	}

	//go wait.Until(s.nodeSyncLoop, nodeSyncPeriod, stopCh)

	<-stopCh
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (s *ServiceController) worker() {
	for {
		func() {
			key, quit := s.workingQueue.Get()
			if quit {
				return
			}
			defer s.workingQueue.Done(key)
			err := s.syncService(key.(string))
			if err != nil {
				glog.Errorf("Error syncing service: %v", err)
			}
		}()
	}
}
func (s *ServiceController) nodesWorker() {
	for {
		func() {
			key, quit := s.nodesQueue.Get()
			if quit {
				return
			}
			defer s.nodesQueue.Done(key)
			obj, exist, err := s.cache.GetByKey(key.(string))
			if err != nil || !exist {
				glog.V(4).Infof("controller: can not get cached service[%s], might been deleted concurrently or [%v]\n", key, err)
				return
			}
			glog.V(5).Infof("controller: node worker sync for [%s]", key)
			cached := obj.(*cachedService)
			err = s.syncBackend(cached)

			if err != nil {
				glog.Errorf("controller: error syncing backend, %v", err)
			}
		}()
	}
}

func (s *ServiceController) init() error {
	if s.cloud == nil {
		return fmt.Errorf("WARNING: no cloud provider provided, services of type LoadBalancer will fail")
	}

	balancer, ok := s.cloud.LoadBalancer()
	if !ok {
		return fmt.Errorf("the cloud provider does not support external load balancers")
	}
	s.balancer = balancer

	return nil
}

func nodeConditionEqual(a, b []v1.NodeCondition) bool {
	if len(a) != len(b) {
		return false
	}
	for _, cona := range a {
		found := false
		for _, conb := range b {
			if string(cona.Type) == string(conb.Type) &&
				string(cona.Status) == string(conb.Status) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Returns an error if processing the service update failed, along with a time.Duration
// indicating whether processing should be retried; zero means no-retry; otherwise
// we should retry in that Duration.
func (s *ServiceController) processServiceUpdate(cachedService *cachedService, service *v1.Service, key string) (error, time.Duration) {
	if cachedService.state != nil {
		if cachedService.state.UID != service.UID {
			err, retry := s.processLoadBalancerDelete(cachedService, key)
			if err != nil {
				return err, retry
			}
		}
	}
	// cache the service, we need the info for service deletion
	cachedService.state = service
	err, retry := s.createLoadBalancerIfNeeded(key, service, cachedService)
	if err != nil {
		message := "Error creating load balancer"
		var retryToReturn time.Duration
		if retry {
			message += " (will retry): "
			retryToReturn = cachedService.nextRetryDelay()
		} else {
			message += " (will not retry): "
			retryToReturn = doNotRetry
		}
		message += err.Error()
		s.eventRecorder.Event(service, v1.EventTypeWarning, "CreatingLoadBalancerFailed", message)
		glog.V(5).Infof("next retry delay is %d\n", retryToReturn)
		return err, retryToReturn
	}
	// Always update the cache upon success.
	// NOTE: Since we update the cached service if and only if we successfully
	// processed it, a cached service being nil implies that it hasn't yet
	// been successfully processed.
	s.cache.set(key, cachedService)

	cachedService.resetRetryDelay()
	return nil, doNotRetry
}

// Returns whatever error occurred along with a boolean indicator of whether it
// should be retried.
func (s *ServiceController) createLoadBalancerIfNeeded(key string, service *v1.Service, cachedService *cachedService) (error, bool) {
	// Note: It is safe to just call EnsureLoadBalancer.  But, on some clouds that requires a delete & create,
	// which may involve service interruption.  Also, we would like user-friendly events.

	// Save the state so we can avoid a write if it doesn't change
	previousState := v1helper.LoadBalancerStatusDeepCopy(&service.Status.LoadBalancer)
	var newState *v1.LoadBalancerStatus
	var err error

	if !needLoadBalancer(service) {
		_, exists, err := s.balancer.GetLoadBalancer(s.clusterName, service)
		if err != nil {
			return fmt.Errorf("error getting LB for service %s: %v", key, err), retryable
		}
		if exists {
			glog.Infof("Deleting existing load balancer for service %s that no longer needs a load balancer.", key)
			s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletingLoadBalancer", "Deleting load balancer")
			if err := s.balancer.EnsureLoadBalancerDeleted(s.clusterName, service); err != nil {
				return err, retryable
			}
			s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletedLoadBalancer", "Deleted load balancer")
		}

		newState = &v1.LoadBalancerStatus{}
	} else {
		glog.V(2).Infof("Ensuring LB for service %s", key)

		// TODO: We could do a dry-run here if wanted to avoid the spurious cloud-calls & events when we restart

		s.eventRecorder.Event(service, v1.EventTypeNormal, "EnsuringLoadBalancer", "Ensuring load balancer")
		newState, err = s.ensureLoadBalancer(service, cachedService)
		if err != nil {
			return fmt.Errorf("failed to ensure load balancer for service %s: %v", key, err), retryable
		}
		s.eventRecorder.Event(service, v1.EventTypeNormal, "EnsuredLoadBalancer", "Ensured load balancer")
	}

	// Write the state if changed
	// TODO: Be careful here ... what if there were other changes to the service?
	if !v1helper.LoadBalancerStatusEqual(previousState, newState) {
		// Make a copy so we don't mutate the shared informer cache
		service = service.DeepCopy()

		// Update the status on the copy
		service.Status.LoadBalancer = *newState

		if err := s.persistUpdate(service); err != nil {
			return fmt.Errorf("failed to persist updated status to apiserver, even after retries. Giving up: %v", err), notRetryable
		}
		// reset state in case to persist service change by createLoadBalancerIfNeeded.
		cachedService.state = service
	} else {
		glog.V(2).Infof("Not persisting unchanged LoadBalancerStatus for service %s to registry.", key)
	}

	return nil, notRetryable
}

func (s *ServiceController) persistUpdate(service *v1.Service) error {
	var err error
	for i := 0; i < clientRetryCount; i++ {
		_, err = s.kubeClient.CoreV1().Services(service.Namespace).UpdateStatus(service)
		if err == nil {
			return nil
		}
		// If the object no longer exists, we don't want to recreate it. Just bail
		// out so that we can process the delete, which we should soon be receiving
		// if we haven't already.
		if errors.IsNotFound(err) {
			glog.Infof("Not persisting update to service '%s/%s' that no longer exists: %v",
				service.Namespace, service.Name, err)
			return nil
		}
		// TODO: Try to resolve the conflict if the change was unrelated to load
		// balancer status. For now, just pass it up the stack.
		if errors.IsConflict(err) {
			return fmt.Errorf("not persisting update to service '%s/%s' that has been changed since we received it: %v",
				service.Namespace, service.Name, err)
		}
		glog.Warningf("Failed to persist updated LoadBalancerStatus to service '%s/%s' after creating its load balancer: %v",
			service.Namespace, service.Name, err)
		time.Sleep(clientRetryInterval)
	}
	return err
}

func (s *ServiceController) ensureLoadBalancer(service *v1.Service, cachedService *cachedService) (*v1.LoadBalancerStatus, error) {
	predicate, err := s.getNodeConditionPredicate(service)
	if err != nil {
		return nil, err
	}
	nodes, err := s.nodeLister.ListWithPredicate(predicate)
	if err != nil {
		return nil, err
	}

	// If there are no available nodes for LoadBalancer service, make a EventTypeWarning event for it.
	if len(nodes) == 0 {
		s.eventRecorder.Eventf(service, v1.EventTypeWarning, "UnAvailableLoadBalancer", "There are no available nodes for LoadBalancer service %s/%s", service.Namespace, service.Name)
	}

	// - Only one protocol supported per service
	// - Not all cloud providers support all protocols and the next step is expected to return
	//   an error for unsupported protocols
	status, err := s.balancer.EnsureLoadBalancer(s.clusterName, service, nodes)
	if err != nil {
		return nil, err
	}
	cachedService.backend = nodes
	return status, nil
}

func (s *ServiceController) syncBackend(service *cachedService) error {

	key, err := controller.KeyFunc(service.state)
	if err != nil {
		glog.Warningf("can not obtain service key ,%s\n", service.state.Name)
	}
	glog.V(4).Infof("start sync backend for service [%s].\n", key)
	service.lock.Lock()
	defer service.lock.Unlock()
	predicate, err := s.getNodeConditionPredicate(service.state)
	if err != nil {
		return err
	}
	nodes, err := s.nodeLister.ListWithPredicate(predicate)
	if err != nil {
		return err
	}
	// If there are no available nodes for LoadBalancer service, make a EventTypeWarning event for it.
	if len(nodes) == 0 {
		s.eventRecorder.Eventf(service.state, v1.EventTypeWarning, "UnAvailableLoadBalancer", "There are no available nodes for LoadBalancer service %s/%s", service.state.Namespace, service.state.Name)
	}
	// service holds the latest service info from apiserver
	_, err = s.serviceLister.Services(service.state.Namespace).Get(service.state.Name)
	if errors.IsNotFound(err) {
		glog.Infof("SyncBackend: Service has been deleted %v", key)
		return nil
	}
	// here we should not check for the neediness of updating loadbalancer.
	// Because, the provider may need to filter by label again.
	/*
		if !needUpdateBackend(service.backend, nodes) {
			return nil
		}
	*/

	// - Only one protocol supported per service
	// - Not all cloud providers support all protocols and the next step is expected to return
	//   an error for unsupported protocols
	if err := s.balancer.UpdateLoadBalancer(s.clusterName, service.state, nodes); err != nil {
		return err
	}
	glog.Infof("finish sync backend for service [%s]\n\n", key)
	service.backend = nodes
	return nil
}

// ListKeys implements the interface required by DeltaFIFO to list the keys we
// already know about.
func (s *serviceCache) ListKeys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.serviceMap))
	for k := range s.serviceMap {
		keys = append(keys, k)
	}
	return keys
}

// GetByKey returns the value stored in the serviceMap under the given key
func (s *serviceCache) GetByKey(key string) (interface{}, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.serviceMap[key]; ok {
		return v, true, nil
	}
	return nil, false, nil
}

// ListKeys implements the interface required by DeltaFIFO to list the keys we
// already know about.
func (s *serviceCache) ListAll() []*cachedService {
	s.mu.Lock()
	defer s.mu.Unlock()
	services := make([]*cachedService, 0, len(s.serviceMap))
	for _, v := range s.serviceMap {
		services = append(services, v)
	}
	return services
}

func (s *serviceCache) get(serviceName string) (*cachedService, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	service, ok := s.serviceMap[serviceName]
	return service, ok
}

func (s *serviceCache) getOrCreate(serviceName string) *cachedService {
	s.mu.Lock()
	defer s.mu.Unlock()
	service, ok := s.serviceMap[serviceName]
	if !ok {
		service = &cachedService{}
		s.serviceMap[serviceName] = service
	}
	return service
}

func (s *serviceCache) set(serviceName string, service *cachedService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceMap[serviceName] = service
}

func (s *serviceCache) delete(serviceName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceMap, serviceName)
}

// lockedEach rely on caller to deal with error
func (s *serviceCache) lockedEach(do func(service *cachedService) error) []error {
	s.mu.Lock()
	defer s.mu.Unlock()
	errs := []error{}
	if len(s.serviceMap) == 0 {
		glog.Infof("service cache: empty service cache , skip func call. ")
		return errs
	}
	for _, svc := range s.serviceMap {
		err := do(svc)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (s *serviceCache) locked(do func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return do()
}

func (s *ServiceController) needsUpdate(oldService *v1.Service, newService *v1.Service) bool {
	if !needLoadBalancer(oldService) && !needLoadBalancer(newService) {
		return false
	}
	if needLoadBalancer(oldService) != needLoadBalancer(newService) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "Type", "%v -> %v",
			oldService.Spec.Type, newService.Spec.Type)
		return true
	}

	if needLoadBalancer(newService) && !reflect.DeepEqual(oldService.Spec.LoadBalancerSourceRanges, newService.Spec.LoadBalancerSourceRanges) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "LoadBalancerSourceRanges", "%v -> %v",
			oldService.Spec.LoadBalancerSourceRanges, newService.Spec.LoadBalancerSourceRanges)
		return true
	}

	if !portsEqualForLB(oldService, newService) || oldService.Spec.SessionAffinity != newService.Spec.SessionAffinity {
		return true
	}
	if !loadBalancerIPsAreEqual(oldService, newService) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "LoadbalancerIP", "%v -> %v",
			oldService.Spec.LoadBalancerIP, newService.Spec.LoadBalancerIP)
		return true
	}
	if len(oldService.Spec.ExternalIPs) != len(newService.Spec.ExternalIPs) {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "ExternalIP", "Count: %v -> %v",
			len(oldService.Spec.ExternalIPs), len(newService.Spec.ExternalIPs))
		return true
	}
	for i := range oldService.Spec.ExternalIPs {
		if oldService.Spec.ExternalIPs[i] != newService.Spec.ExternalIPs[i] {
			s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "ExternalIP", "Added: %v",
				newService.Spec.ExternalIPs[i])
			return true
		}
	}
	if !reflect.DeepEqual(oldService.Annotations, newService.Annotations) {
		return true
	}
	if oldService.UID != newService.UID {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "UID", "%v -> %v",
			oldService.UID, newService.UID)
		return true
	}
	if oldService.Spec.ExternalTrafficPolicy != newService.Spec.ExternalTrafficPolicy {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "ExternalTrafficPolicy", "%v -> %v",
			oldService.Spec.ExternalTrafficPolicy, newService.Spec.ExternalTrafficPolicy)
		return true
	}
	if oldService.Spec.HealthCheckNodePort != newService.Spec.HealthCheckNodePort {
		s.eventRecorder.Eventf(newService, v1.EventTypeNormal, "HealthCheckNodePort", "%v -> %v",
			oldService.Spec.HealthCheckNodePort, newService.Spec.HealthCheckNodePort)
		return true
	}

	return false
}

func (s *ServiceController) loadBalancerName(service *v1.Service) string {
	return cloudprovider.GetLoadBalancerName(service)
}

func getPortsForLB(service *v1.Service) ([]*v1.ServicePort, error) {
	var protocol v1.Protocol

	ports := []*v1.ServicePort{}
	for i := range service.Spec.Ports {
		sp := &service.Spec.Ports[i]
		// The check on protocol was removed here.  The cloud provider itself is now responsible for all protocol validation
		ports = append(ports, sp)
		if protocol == "" {
			protocol = sp.Protocol
		} else if protocol != sp.Protocol && needLoadBalancer(service) {
			// TODO:  Convert error messages to use event recorder
			return nil, fmt.Errorf("mixed protocol external load balancers are not supported")
		}
	}
	return ports, nil
}

func portsEqualForLB(x, y *v1.Service) bool {
	xPorts, err := getPortsForLB(x)
	if err != nil {
		return false
	}
	yPorts, err := getPortsForLB(y)
	if err != nil {
		return false
	}
	return portSlicesEqualForLB(xPorts, yPorts)
}

func portSlicesEqualForLB(x, y []*v1.ServicePort) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if !portEqualForLB(x[i], y[i]) {
			return false
		}
	}
	return true
}

func portEqualForLB(x, y *v1.ServicePort) bool {
	// TODO: Should we check name?  (In theory, an LB could expose it)
	if x.Name != y.Name {
		return false
	}

	if x.Protocol != y.Protocol {
		return false
	}

	if x.Port != y.Port {
		return false
	}

	if x.NodePort != y.NodePort {
		return false
	}

	// We don't check TargetPort; that is not relevant for load balancing
	// TODO: Should we blank it out?  Or just check it anyway?

	return true
}

func nodeNames(nodes []*v1.Node) []string {
	ret := make([]string, len(nodes))
	for i, node := range nodes {
		ret[i] = node.Name
	}
	return ret
}

func nodeLabelsEqual(a, b []*v1.Node) bool {
	if len(a) != len(b) {
		return false
	}
	for _, na := range a {
		found := false
		for _, nb := range b {
			if na.Name != nb.Name {
				continue
			}
			// found and break
			found = true
			if !labelEqual(na.Labels, nb.Labels) {
				return false
			}
			//succeed, break for next
			break
		}
		if !found {
			return false
		}
	}
	return true
}

func labelEqual(label1, label2 map[string]string) bool {
	if len(label1) != len(label2) {
		return false
	}
	for k0, value0 := range label2 {
		if value1, exist := label2[k0]; !exist || value0 != value1 {
			return false
		}
	}
	return true
}

func needUpdateBackend(x, y []*v1.Node) bool {
	if len(x) != len(y) {
		return false
	}
	if !stringSlicesEqual(nodeNames(x), nodeNames(y)) {
		return false
	}
	return nodeLabelsEqual(x, y)
}

func stringSlicesEqual(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	if !sort.StringsAreSorted(x) {
		sort.Strings(x)
	}
	if !sort.StringsAreSorted(y) {
		sort.Strings(y)
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func (s *ServiceController) getNodeConditionPredicate(service *v1.Service) (corelisters.NodeConditionPredicate, error) {
	eplist, err := func() (map[string]string, error) {
		if service.Spec.ExternalTrafficPolicy !=
			v1.ServiceExternalTrafficPolicyTypeLocal {
			return nil, nil
		}
		nodes := make(map[string]string)
		ep, err := s.epLister.Endpoints(service.Namespace).Get(service.Name)
		if err != nil {
			//if err.(errors.StatusError).Status().Code == http.StatusNotFound {
			//	glog.Warningf("alicloud: ecounter service without an endpoints which is not " +
			//		"expected ! this will result in no backend of a loadbalancer.[%s/%s]\n",service.Namespace,service.Name)
			//	return nodes, nil
			//}
			glog.Errorf("alicloud: find endpoints for service [%s/%s] with error [%s]", service.Namespace, service.Name, err.Error())
			return nil, err
		}
		records := []string{}
		for _, sub := range ep.Subsets {
			for _, add := range sub.Addresses {
				nodes[*add.NodeName] = *add.NodeName
				records = append(records, *add.NodeName)
			}
		}
		glog.V(4).Infof("controller: node condition predicate, external traffic policy should accept node %v\n", records)
		return nodes, nil
	}()
	if err != nil {
		return nil, err
	}
	return func(node *v1.Node) bool {
		// We add the master to the node list, but its unschedulable.  So we use this to filter
		// the master.
		if node.Spec.Unschedulable {
			return false
		}

		// As of 1.6, we will taint the master, but not necessarily mark it unschedulable.
		// Recognize nodes labeled as master, and filter them also, as we were doing previously.
		if _, hasMasterRoleLabel := node.Labels[LabelNodeRoleMaster]; hasMasterRoleLabel {
			return false
		}

		if utilfeature.DefaultFeatureGate.Enabled(kubefeatures.ServiceNodeExclusion) {
			if _, hasExcludeBalancerLabel := node.Labels[LabelNodeRoleExcludeBalancer]; hasExcludeBalancerLabel {
				return false
			}
		}

		// If we have no info, don't accept
		if len(node.Status.Conditions) == 0 {
			return false
		}
		for _, cond := range node.Status.Conditions {
			// We consider the node for load balancing only when its NodeReady condition status
			// is ConditionTrue
			if cond.Type == v1.NodeReady && cond.Status != v1.ConditionTrue {
				glog.V(4).Infof("Ignoring node %v with %v condition status %v", node.Name, cond.Type, cond.Status)
				return false
			}
		}
		if eplist != nil {
			if _, exist := eplist[node.Name]; !exist {
				return false
			}
		}
		return true
	}, nil
}

//// nodeSyncLoop handles updating the hosts pointed to by all load
//// balancers whenever the set of nodes in the cluster changes.
//func (s *ServiceController) nodeSyncLoop() {
//	svcs := s.cache.ListAll()
//	if len(svcs) <= 0 {
//		// nothing to sync
//		return
//	}
//	updated := 0
//	total   := 0
//	errs := s.cache.lockedEach(
//		func(svc *cachedService) error {
//			if ! needLoadBalancer(svc.state) {
//				glog.V(6).Infof("service [%s] does not need loadbalancer , skip.\n",svc.state.Name)
//				return nil
//			}
//			predicate, err := s.getNodeConditionPredicate(svc.state)
//			if err != nil {
//				return err
//			}
//			newHosts, err := s.nodeLister.ListWithPredicate(predicate)
//			if needUpdateBackend(newHosts, svc.backend) {
//				total += 1
//				if err := s.balancer.UpdateLoadBalancer(s.clusterName, svc.state, newHosts) ;
//					err != nil {
//					s.eventRecorder.Event(svc.state, v1.EventTypeWarning, "UpdatedLoadBalancer", "Updated load balancer Fail")
//					return err
//				}
//				updated += 1
//				svc.backend = newHosts
//				s.eventRecorder.Event(svc.state, v1.EventTypeNormal, "UpdatedLoadBalancer", "Updated load balancer with new hosts")
//			}
//			return nil
//		},
//	)
//	if errs != nil {
//		glog.Errorf("error update %d service backend, " +
//			"will be resync in the next node loop\n",total-updated, errorutils.NewAggregate(errs).Error())
//	}
//	glog.Infof("Successfully updated %d out of %d load balancers to direct traffic to the updated set of nodes.", updated, total)
//}

// updateLoadBalancerHosts updates all existing load balancers so that
// they will match the list of hosts provided.
// Returns the list of services that couldn't be updated.
func (s *ServiceController) updateLoadBalancerHosts(services []*v1.Service, hosts []*v1.Node) (servicesToRetry []*v1.Service) {
	for _, service := range services {
		func() {
			if service == nil {
				return
			}
			if err := s.lockedUpdateLoadBalancerHosts(service, hosts); err != nil {
				glog.Errorf("External error while updating load balancer: %v.", err)
				servicesToRetry = append(servicesToRetry, service)
			}
		}()
	}
	return servicesToRetry
}

// Updates the load balancer of a service, assuming we hold the mutex
// associated with the service.
func (s *ServiceController) lockedUpdateLoadBalancerHosts(service *v1.Service, hosts []*v1.Node) error {
	if !needLoadBalancer(service) {
		return nil
	}

	// This operation doesn't normally take very long (and happens pretty often), so we only record the final event
	err := s.balancer.UpdateLoadBalancer(s.clusterName, service, hosts)
	if err == nil {
		// If there are no available nodes for LoadBalancer service, make a EventTypeWarning event for it.
		if len(hosts) == 0 {
			s.eventRecorder.Eventf(service, v1.EventTypeWarning, "UnAvailableLoadBalancer", "There are no available nodes for LoadBalancer service %s/%s", service.Namespace, service.Name)
		} else {
			s.eventRecorder.Event(service, v1.EventTypeNormal, "UpdatedLoadBalancer", "Updated load balancer with new hosts")
		}
		return nil
	}

	// It's only an actual error if the load balancer still exists.
	if _, exists, err := s.balancer.GetLoadBalancer(s.clusterName, service); err != nil {
		glog.Errorf("External error while checking if load balancer %q exists: name, %v", cloudprovider.GetLoadBalancerName(service), err)
	} else if !exists {
		return nil
	}

	s.eventRecorder.Eventf(service, v1.EventTypeWarning, "LoadBalancerUpdateFailed", "Error updating load balancer with new hosts %v: %v", nodeNames(hosts), err)
	return err
}

func needLoadBalancer(service *v1.Service) bool {
	return service.Spec.Type == v1.ServiceTypeLoadBalancer
}

func needLocalBackend(service *v1.Service) bool {
	return service.Spec.ExternalTrafficPolicy ==
		v1.ServiceExternalTrafficPolicyTypeLocal
}

func loadBalancerIPsAreEqual(oldService, newService *v1.Service) bool {
	return oldService.Spec.LoadBalancerIP == newService.Spec.LoadBalancerIP
}

// Computes the next retry, using exponential backoff
// mutex must be held.
func (s *cachedService) nextRetryDelay() time.Duration {
	s.lastRetryDelay = s.lastRetryDelay * 2
	if s.lastRetryDelay < minRetryDelay {
		s.lastRetryDelay = minRetryDelay
	}
	if s.lastRetryDelay > maxRetryDelay {
		s.lastRetryDelay = maxRetryDelay
	}
	return s.lastRetryDelay
}

// Resets the retry exponential backoff.  mutex must be held.
func (s *cachedService) resetRetryDelay() {
	s.lastRetryDelay = time.Duration(0)
}

// syncService will sync the Service with the given key if it has had its expectations fulfilled,
// meaning it did not expect to see any more of its pods created or deleted. This function is not meant to be
// invoked concurrently with the same key.
func (s *ServiceController) syncService(key string) error {
	startTime := time.Now()
	//var cachedService *cachedService

	cachedService := s.cache.getOrCreate(key)
	cachedService.lock.Lock()
	defer cachedService.lock.Unlock()
	var retryDelay time.Duration
	defer func() {
		glog.V(4).Infof("Finished syncing service %q (%v)\n\n", key, time.Now().Sub(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// service holds the latest service info from apiserver
	service, err := s.serviceLister.Services(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		// service absence in store means watcher caught the deletion, ensure LB info is cleaned
		glog.Infof("Service has been deleted %v", key)
		err, retryDelay = s.processServiceDeletion(key)
		return err
	case err != nil:
		glog.Infof("Unable to retrieve service %v from store: %v", key, err)
		s.workingQueue.Add(key)
		return err
	default:
		err, retryDelay = s.processServiceUpdate(cachedService, service, key)
	}

	if retryDelay != 0 {
		// Add the failed service back to the queue so we'll retry it.
		glog.Errorf("Failed to process service %v. Retrying in %s: %v", key, retryDelay, err)
		go func(obj interface{}, delay time.Duration) {
			// put back the service key to working queue, it is possible that more entries of the service
			// were added into the queue during the delay, but it does not mess as when handling the retry,
			// it always get the last service info from service store
			s.workingQueue.AddAfter(obj, delay)
		}(key, retryDelay)
	} else if err != nil {
		runtime.HandleError(fmt.Errorf("failed to process service %v. Not retrying: %v", key, err))
	}
	return nil
}

// Returns an error if processing the service deletion failed, along with a time.Duration
// indicating whether processing should be retried; zero means no-retry; otherwise
// we should retry after that Duration.
func (s *ServiceController) processServiceDeletion(key string) (error, time.Duration) {
	cachedService, ok := s.cache.get(key)
	if !ok {
		return fmt.Errorf("service %s not in cache even though the watcher thought it was. Ignoring the deletion", key), doNotRetry
	}
	return s.processLoadBalancerDelete(cachedService, key)
}

func (s *ServiceController) processLoadBalancerDelete(cachedService *cachedService, key string) (error, time.Duration) {
	service := cachedService.state
	// delete load balancer info only if the service type is LoadBalancer
	if !needLoadBalancer(service) {
		return nil, doNotRetry
	}
	s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletingLoadBalancer", "Deleting load balancer")
	err := s.balancer.EnsureLoadBalancerDeleted(s.clusterName, service)
	if err != nil {
		message := "Error deleting load balancer (will retry): " + err.Error()
		s.eventRecorder.Event(service, v1.EventTypeWarning, "DeletingLoadBalancerFailed", message)
		return err, cachedService.nextRetryDelay()
	}
	s.eventRecorder.Event(service, v1.EventTypeNormal, "DeletedLoadBalancer", "Deleted load balancer")
	s.cache.delete(key)

	cachedService.resetRetryDelay()
	return nil, doNotRetry
}
