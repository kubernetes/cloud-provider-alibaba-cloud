package service

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	queue "k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/metrics"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/util/workqueue"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	//SERVICE_SYNC_PERIOD Interval of synchronizing service status from apiserver
	SERVICE_SYNC_PERIOD = 1 * time.Minute
	SERVICE_QUEUE       = "service-queue"
	NODE_QUEUE          = "node-queue"
	SERVICE_CONTROLLER  = "service-controller"

	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	// LabelNodeRoleExcludeBalancer specifies that the node should be
	// exclude from load balancers created by a cloud provider.
	LabelNodeRoleExcludeBalancer = "alpha.service-controller.kubernetes.io/exclude-balancer"
	CCM_CLASS                    = "service.beta.kubernetes.io/class"
)

type EnsureENI interface {
	// EnsureLoadbalancerWithENI
	// creates a new load balancer 'name', or updates the existing one. Returns the
	// status of the balancer. Implementations must treat the *v1.svc and *v1.Endpoints
	// parameters as read-only and not modify them.
	// Parameter 'name' is the name of the cluster as presented to kube-controller-manager
	EnsureLoadBalancerWithENI(name string, svc *v1.Service, endpoint *v1.Endpoints) (*v1.LoadBalancerStatus, error)

	// UpdateLoadBalancerWithENI updates hosts under the specified load balancer.
	// Implementations must treat the *v1.Service and *v1.Enodpoints
	// parameters as read-only and not modify them.
	// Parameter 'name' is the name of the cluster as presented to kube-controller-manager
	UpdateLoadBalancerWithENI(name string, service *v1.Service, endpoint *v1.Endpoints) error
}

type Controller struct {
	cloud       cloudprovider.LoadBalancer
	client      clientset.Interface
	ifactory    informers.SharedInformerFactory
	clusterName string
	local       *Context
	caster      record.EventBroadcaster
	recorder    record.EventRecorder

	// Package workqueue provides a simple queue that supports the following
	// features:
	//  * Fair: items processed in the order in which they are added.
	//  * Stingy: a single item will not be processed multiple times concurrently,
	//      and if an item is added multiple times before it can be processed, it
	//      will only be processed once.
	//  * Multiple consumers and producers. In particular, it is allowed for an
	//      item to be reenqueued while it is being processed.
	//  * Shutdown notifications.
	queues map[string]queue.DelayingInterface
}

func NewController(
	cloud cloudprovider.LoadBalancer,
	client clientset.Interface,
	ifactory informers.SharedInformerFactory,
	clusterName string,
) (*Controller, error) {

	recorder, caster := broadcaster()
	rate := client.CoreV1().RESTClient().GetRateLimiter()
	if client != nil && rate != nil {
		if err := metrics.RegisterMetricAndTrackRateLimiterUsage("service_controller", rate); err != nil {
			return nil, err
		}
	}

	con := &Controller{
		cloud:       cloud,
		clusterName: clusterName,
		ifactory:    ifactory,
		local:       &Context{},
		caster:      caster,
		recorder:    recorder,
		client:      client,
		queues: map[string]queue.DelayingInterface{
			NODE_QUEUE:    workqueue.NewNamedDelayingQueue(NODE_QUEUE),
			SERVICE_QUEUE: workqueue.NewNamedDelayingQueue(SERVICE_QUEUE),
		},
	}
	con.HandlerForEndpointChange(
		con.local,
		con.queues[NODE_QUEUE],
		con.ifactory.Core().V1().Endpoints().Informer(),
	)
	con.HandlerForNodesChange(
		con.local,
		con.queues[NODE_QUEUE],
		con.ifactory.Core().V1().Nodes().Informer(),
	)
	con.HandlerForServiceChange(
		con.local,
		con.queues[SERVICE_QUEUE],
		con.ifactory.Core().V1().Services().Informer(),
		recorder,
	)
	return con, nil
}

func (con *Controller) Run(stopCh <-chan struct{}, workers int) {
	defer runtime.HandleCrash()
	defer func() {
		for _, que := range con.queues {
			que.ShutDown()
		}
	}()

	glog.Info("starting service controller")
	defer glog.Info("shutting down service controller")

	if !controller.WaitForCacheSync(
		"service",
		stopCh,
		con.ifactory.Core().V1().Services().Informer().HasSynced,
		con.ifactory.Core().V1().Nodes().Informer().HasSynced,
	) {
		glog.Error("service and nodes cache has not been syncd")
		return
	}

	tasks := map[string]SyncTask{
		NODE_QUEUE:    con.NodeSyncTask,
		SERVICE_QUEUE: con.ServiceSyncTask,
	}
	for i := 0; i < workers; i++ {
		// run service&node sync worker
		for que, task := range tasks {
			go wait.Until(
				WorkerFunc(
					con.local,
					con.queues[que],
					task,
				),
				2*time.Second,
				stopCh,
			)
		}
	}

	glog.Info("service controller started")
	<-stopCh
}

func broadcaster() (record.EventRecorder, record.EventBroadcaster) {
	caster := record.NewBroadcaster()
	caster.StartLogging(glog.Infof)
	source := v1.EventSource{Component: SERVICE_CONTROLLER}
	return caster.NewRecorder(scheme.Scheme, source), caster
}

func key(svc *v1.Service) string {
	return fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
}

func Enqueue(queue queue.DelayingInterface, k interface{}) {
	glog.Infof("controller: enqueue object %s for service", k.(string))
	queue.Add(k.(string))
}

func (con *Controller) HandlerForNodesChange(
	ctx *Context,
	que queue.DelayingInterface,
	informer cache.SharedIndexInformer,
) {

	syncNodes := func(object interface{}) {
		node, ok := object.(*v1.Node)
		if !ok || node == nil {
			glog.Info("node change: node object is nil, skip")
			return
		}
		// node change may affect any service that concerns
		// eg. Need LoadBalancer
		ctx.Range(
			func(k string, svc *v1.Service) bool {
				if !NeedLoadBalancer(svc) {
					utils.Logf(svc, "node change: loadbalancer is not needed, skip")
					return true
				}
				if !isProcessNeeded(svc) {
					utils.Logf(svc, "node change: class not empty, skip process ")
					return true
				}
				utils.Logf(svc, "node change: enque service")
				Enqueue(que, key(svc))
				return true
			},
		)
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: syncNodes,
			UpdateFunc: func(obja, objb interface{}) {
				node1, ok1 := obja.(*v1.Node)
				node2, ok2 := objb.(*v1.Node)
				if ok1 && ok2 &&
					NodeSpecChanged(node1, node2) {
					// label and schedulable changed .
					// status healthy should be considered
					glog.Infof("controller: node[%s/%s] update event", node1.Namespace, node1.Name)
					syncNodes(node1)
				}
			},
			DeleteFunc: syncNodes,
		},
		SERVICE_SYNC_PERIOD,
	)
}

func (con *Controller) HandlerForEndpointChange(
	context *Context,
	que queue.DelayingInterface,
	informer cache.SharedIndexInformer,
) {
	syncEndpoints := func(epd interface{}) {
		ep, ok := epd.(*v1.Endpoints)
		if !ok || ep == nil {
			glog.Info("endpoints change: endpoint object is nil, skip")
			return
		}
		svc := context.Get(fmt.Sprintf("%s/%s", ep.Namespace, ep.Name))
		if svc == nil {
			glog.Infof("endpoint change: can not get cached service for "+
				"endpoints[%s/%s], enqueue for default endpoint.\n", ep.Namespace, ep.Name)
			var err error
			svc, err = con.client.CoreV1().Services(ep.Namespace).Get(ep.Name,v12.GetOptions{})
			if err!=nil {
				glog.Warningf("can not get service %s/%s. ",ep.Namespace,ep.Name)
				return
			}
		}
		if !isProcessNeeded(svc) {
			utils.Logf(svc, "endpoint: class not empty, skip process ")
			return
		}
		if !NeedLoadBalancer(svc) {
			// we are safe here to skip process syncEnpoint.
			utils.Logf(svc, "endpoint change: loadBalancer is not needed, skip")
			return
		}
		utils.Logf(svc, "enqueue endpoint")
		Enqueue(que, key(svc))
	}
	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: syncEndpoints,
			UpdateFunc: func(obja, objb interface{}) {
				ep1, ok1 := obja.(*v1.Endpoints)
				ep2, ok2 := objb.(*v1.Endpoints)
				if ok1 && ok2 && !reflect.DeepEqual(ep1.Subsets, ep2.Subsets) {
					glog.Infof("controller: endpoints update event, endpoints [%s/%s]\n", ep1.Namespace, ep1.Name)
					syncEndpoints(ep1)
				}
			},
			DeleteFunc: syncEndpoints,
		},
		SERVICE_SYNC_PERIOD,
	)
}

func (con *Controller) HandlerForServiceChange(
	context *Context,
	que queue.DelayingInterface,
	informer cache.SharedIndexInformer,
	record record.EventRecorder,
) {
	syncService := func(svc *v1.Service) {
		if !isProcessNeeded(svc) {
			utils.Logf(svc, "class not empty, skip process")
			return
		}
		Enqueue(que, key(svc))
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(add interface{}) {
				svc, ok := add.(*v1.Service)
				if !ok {
					utils.Logf(svc, "add: not type service %s, skip", reflect.TypeOf(add))
					return
				}
				utils.Logf(svc, "service addiontion event received %s", reflect.TypeOf(svc))
				syncService(svc)
			},
			UpdateFunc: func(old, cur interface{}) {
				oldd, ok1 := old.(*v1.Service)
				curr, ok2 := cur.(*v1.Service)
				if ok1 && ok2 &&
					NeedUpdate(oldd, curr, record) {
					utils.Logf(curr, "service update event")
					syncService(curr)
				}
			},
			DeleteFunc: func(cur interface{}) {
				svc, ok := cur.(*v1.Service)
				if !ok {
					utils.Logf(svc, "delete: not type service %s, skip", reflect.TypeOf(cur))
					return
				}
				utils.Logf(svc, "controller: service deletion received, %s", utils.PrettyJson(svc))
				// recorder service in local context
				context.Set(key(svc), svc)
				syncService(svc)
			},
		},
		SERVICE_SYNC_PERIOD,
	)
}

func WorkerFunc(
	contex *Context,
	queue queue.DelayingInterface,
	syncd SyncTask,
) func() {

	return func() {
		for {
			func() {
				// Workerqueue ensures that a single key would not be process
				// by two worker concurrently, so multiple workers is safe here.
				key, quit := queue.Get()
				if quit {
					return
				}
				defer queue.Done(key)

				glog.Infof("[%s] worker: queued sync for service", key)

				if err := syncd(key.(string)); err != nil {
					queue.AddAfter(key, 4*time.Second)
					glog.Errorf("requeue: sync error for service %s %v", key, err)
				}
			}()
		}
	}
}

type SyncTask func(key string) error

// ------------------------------------------------- Where sync begins ---------------------------------------------------------

// SyncService Entrance for syncing service
func (con *Controller) ServiceSyncTask(k string) error {
	startTime := time.Now()

	ns, name, err := cache.SplitMetaNamespaceKey(k)
	if err != nil {
		return fmt.Errorf("unexpected key format %s for syncing service", k)
	}

	// local cache might be nil on first process which is expected
	cached := con.local.Get(k)

	defer func() {
		glog.Infof("[%s] finished syncing service (%v)", k, time.Now().Sub(startTime))
	}()

	// service holds the latest service info from apiserver
	service, err := con.
		ifactory.
		Core().
		V1().
		Services().
		Lister().
		Services(ns).
		Get(name)
	switch {
	case errors.IsNotFound(err):

		if cached == nil {
			glog.Errorf("unexpected nil cached service for deletion, wait retry %s", k)
			return nil
		}
		// service absence in store means watcher caught the deletion, ensure LB
		// info is cleaned delete error would cause ReEnqueue svc, which mean retry.
		utils.Logf(cached, "service has been deleted %v", key(cached))
		return retry(nil, con.delete, cached)
	case err != nil:
		return fmt.Errorf("failed to load service from local context: %s", err.Error())
	default:
		// catch unexpected service
		if service == nil {
			glog.Errorf("unexpected nil service for update, wait retry. %s", k)
			return fmt.Errorf("retry unexpected nil service %s. ", k)
		}
		return con.update(cached, service)
	}
}

func isProcessNeeded(svc *v1.Service) bool { return svc.Annotations[CCM_CLASS] == "" }

func retry(
	backoff *wait.Backoff,
	fun func(svc *v1.Service) error,
	svc *v1.Service,
) error {
	if backoff == nil {
		backoff = &wait.Backoff{
			Duration: 1 * time.Second,
			Steps:    8,
			Factor:   2,
			Jitter:   4,
		}
	}
	return wait.ExponentialBackoff(
		*backoff,
		func() (bool, error) {
			err := fun(svc)
			if err != nil &&
				strings.Contains(err.Error(), "try again") {
				glog.Errorf("retry with error: %s", err.Error())
				return false, nil
			}
			if err != nil {
				glog.Errorf("retry error: NotRetry, %s", err.Error())
			}
			return true, nil
		},
	)
}

func (con *Controller) update(cached, svc *v1.Service) error {

	// Save the state so we can avoid a write if it doesn't change
	pre := v1helper.LoadBalancerStatusDeepCopy(&svc.Status.LoadBalancer)
	if cached != nil &&
		cached.UID != svc.UID {
		con.recorder.Eventf(
			cached,
			v1.EventTypeNormal,
			"UIDChanged",
			"uid: %s -> %s, try delete old service first",
			cached.UID,
			svc.UID,
		)
		return retry(nil, con.delete, svc)
	}
	var newm *v1.LoadBalancerStatus
	if !NeedLoadBalancer(svc) {
		// delete loadbalancer which is no longer needed
		utils.Logf(svc, "try delete loadbalancer which no longer needed for service.")

		if err := retry(nil, con.delete, svc); err != nil {
			return err
		}
		// continue for updating service status.
		newm = &v1.LoadBalancerStatus{}
	} else {

		utils.Logf(svc, "start to ensure loadbalancer")
		var (
			err error
		)
		if IsENIBackendType(svc) {
			// Ensure ENI type backend
			eni, ok := con.cloud.(EnsureENI)
			if !ok {
				return fmt.Errorf("cloud does not implement EnsureENI interface")
			}
			var eps *v1.Endpoints
			eps, err = con.ifactory.Core().V1().Endpoints().Lister().Endpoints(svc.Namespace).Get(svc.Name)
			if err != nil {
				return fmt.Errorf("get available endpoints for eni: %s", err.Error())
			}
			newm, err = eni.EnsureLoadBalancerWithENI(con.clusterName, svc, eps)
		} else {
			var nodes []*v1.Node
			nodes, err = AvailableNodes(svc, con.ifactory)
			if err != nil {
				return fmt.Errorf("error get available nodes %s", err.Error())
			}
			// Fire warning event if there are no available nodes
			// for loadbalancer service
			if len(nodes) == 0 {
				con.recorder.Eventf(
					svc,
					v1.EventTypeWarning,
					"UnAvailableLoadBalancer",
					"There are no available backend loadbalancer nodes for service %s",
					key(svc),
				)
			}
			newm, err = con.cloud.EnsureLoadBalancer(con.clusterName, svc, nodes)
		}
		if err != nil {
			return fmt.Errorf("ensure loadbalancer error: %s", err)
		}
		con.recorder.Eventf(
			svc,
			v1.EventTypeNormal,
			"EnsuredLoadBalancer",
			"Finished ensured loadbalancer, %s",
			key(svc),
		)
	}
	if err := con.updateStatus(svc, pre, newm); err != nil {
		return fmt.Errorf("update service status: %s", err.Error())
	}
	// Always update the cache upon success.
	// NOTE: Since we update the cached service if and only if we successfully
	// processed it, a cached service being nil implies that it hasn't yet
	// been successfully processed.
	con.local.Set(key(svc), svc)
	return nil
}
func (con *Controller) updateStatus(svc *v1.Service, pre, newm *v1.LoadBalancerStatus) error {
	if newm == nil {
		return fmt.Errorf("status not updated for nil status reason")
	}
	// Write the state if changed
	// TODO: Be careful here ... what if there were other changes to the service?
	if !v1helper.LoadBalancerStatusEqual(pre, newm) {
		// Make a copy so we don't mutate the shared informer cache
		service := svc.DeepCopy()

		// Update the status on the copy
		service.Status.LoadBalancer = *newm

		return retry(
			&wait.Backoff{
				Duration: 1 * time.Second,
				Steps:    3,
				Factor:   2,
				Jitter:   4,
			},
			func(svc *v1.Service) error {
				_, err := con.
					client.
					CoreV1().
					Services(service.Namespace).
					UpdateStatus(service)
				if err == nil {
					return nil
				}
				// If the object no longer exists, we don't want to recreate it. Just bail
				// out so that we can process the delete, which we should soon be receiving
				// if we haven't already.
				if errors.IsNotFound(err) {
					utils.Logf(service, "not persisting update to service that no "+
						"longer exists: %v", err)
					return nil
				}
				// TODO: Try to resolve the conflict if the change was unrelated to load
				// balancer status. For now, just pass it up the stack.
				if errors.IsConflict(err) {
					return fmt.Errorf("not persisting update to service %s that "+
						"has been changed since we received it: %v", key(svc), err)
				}
				glog.Warningf("failed to persist updated LoadBalancerStatus to "+
					"service %s after creating its load balancer: %v", key(svc), err)
				return fmt.Errorf("retry with %s", err.Error())
			},
			service,
		)
	}
	utils.Logf(svc, "not persisting unchanged LoadBalancerStatus for service to registry.")
	return nil
}

func (con *Controller) delete(svc *v1.Service) error {

	// do not check for the neediness of loadbalancer, delete anyway.
	con.recorder.Eventf(
		svc,
		v1.EventTypeNormal,
		"DeletingLoadBalancer",
		"for service: %s",
		key(svc),
	)
	err := con.cloud.EnsureLoadBalancerDeleted(con.clusterName, svc)
	if err != nil {
		con.recorder.Eventf(
			svc,
			v1.EventTypeWarning,
			"DeletingLoadBalancerFailed",
			"Error deleting: %s",
			err.Error(),
		)
		return fmt.Errorf("please retry")
	}
	con.recorder.Eventf(
		svc,
		v1.EventTypeNormal,
		"DeletedLoadBalancer",
		"LoadBalancer Deleted SUCCESS. %s",
		key(svc),
	)
	con.local.Remove(key(svc))
	return nil
}

func (con *Controller) NodeSyncTask(k string) error {
	glog.Infof("[%s] start sync backend for service.\n", k)

	ns, name, err := cache.SplitMetaNamespaceKey(k)
	if err != nil {
		return fmt.Errorf("unexpected key format %s for syncing backends", k)
	}

	// service holds the latest service info from apiserver
	service, err := con.
		ifactory.
		Core().
		V1().
		Services().
		Lister().
		Services(ns).
		Get(name)
	switch {
	case errors.IsNotFound(err):
		glog.Errorf("service %s not found for backend sync, finished", k)
		return nil
	case err != nil:
		return fmt.Errorf("failed to load service from local context for backend: %s", err.Error())
	default:
		// catch unexpected service
		if service == nil {
			glog.Errorf("unexpected nil service for update, wait retry. %s", k)
			return nil
		}

		defer utils.Logf(service, "finish sync backend for service\n\n")
		if IsENIBackendType(service) {
			eni, ok := con.cloud.(EnsureENI)
			if !ok {
				return fmt.Errorf("cloud does not implement EnsureENI interface")
			}
			eps, err := con.ifactory.Core().V1().Endpoints().Lister().Endpoints(service.Namespace).Get(service.Name)
			if err != nil {
				return fmt.Errorf("get available endpoints for eni: %s", err.Error())
			}
			return eni.UpdateLoadBalancerWithENI(con.clusterName, service, eps)
		}
		nodes, err := AvailableNodes(service, con.ifactory)
		if err != nil {
			return fmt.Errorf("get available nodes: %s", err.Error())
		}
		// Warning for zero length nodes
		if len(nodes) == 0 {
			con.recorder.Eventf(
				service,
				v1.EventTypeWarning,
				"UnAvailableLoadBalancer",
				"There are no available nodes for LoadBalancer service %s, NodeSyncTask",
				key(service),
			)
		}
		return con.cloud.UpdateLoadBalancer(con.clusterName, service, nodes)
	}
}

func AvailableNodes(
	svc *v1.Service,
	ifactory informers.SharedInformerFactory,
) ([]*v1.Node, error) {
	predicate, err := NodeConditionPredicate(svc)
	if err != nil {
		return nil, fmt.Errorf("error get predicate: %s", err.Error())
	}
	nodes, err := ifactory.
		Core().
		V1().
		Nodes().
		Lister().
		ListWithPredicate(predicate)
	if err != nil {
		return nodes, err
	}
	if ServiceModeLocal(svc) {
		nodes, err = AvailableNodeModeLocal(nodes, svc, ifactory)
	}
	return nodes, err
}

func NodeConditionPredicate(svc *v1.Service) (corelisters.NodeConditionPredicate, error) {

	predicate := func(node *v1.Node) bool {
		// Filter unschedulable node.
		if node.Spec.Unschedulable {
			if svc.Annotations[utils.ServiceAnnotationLoadBalancerRemoveUnscheduledBackend] == "on" {
				utils.Logf(svc, "ignore node %s with unschedulable condition", node.Name)
				return false
			}
		}

		// As of 1.6, we will taint the master, but not necessarily mark
		// it unschedulable. Recognize nodes labeled as master, and filter
		// them also, as we were doing previously.
		if _, isMaster := node.Labels[LabelNodeRoleMaster]; isMaster {
			return false
		}

		if _, exclude := node.Labels[LabelNodeRoleExcludeBalancer]; exclude {
			utils.Logf(svc, "ignore node with exclude label %s", node.Name)
			return false
		}

		// If we have no info, don't accept
		if len(node.Status.Conditions) == 0 {
			return false
		}
		for _, cond := range node.Status.Conditions {
			// We consider the node for load balancing only when its NodeReady
			// condition status is ConditionTrue
			if cond.Type == v1.NodeReady &&
				cond.Status != v1.ConditionTrue {
				utils.Logf(svc, "ignoring node %v with %v condition "+
					"status %v", node.Name, cond.Type, cond.Status)
				return false
			}
		}

		return true
	}

	return predicate, nil
}

func AvailableNodeModeLocal(nodes []*v1.Node, svc *v1.Service,
	ifactory informers.SharedInformerFactory) ([]*v1.Node, error) {
	var (
		availableNodes    []*v1.Node
		records           []string
		availableNodesMap = make(map[string]float64)
		endpoints         = ifactory.Core().V1().Endpoints().Lister()
		podNum            = 0.0
	)

	ep, err := endpoints.Endpoints(svc.Namespace).Get(svc.Name)
	if err != nil {
		return nil, fmt.Errorf("find endpoints for service "+
			"[%s] with error [%s]", key(svc), err.Error())
	}

	utils.Logf(svc, "endpoint has [%d] subsets. ", len(ep.Subsets))
	for _, sub := range ep.Subsets {
		for _, add := range sub.Addresses {
			utils.Logf(svc, "prepare to add node [%s] for service backend", *add.NodeName)
			if _, exist := availableNodesMap[*add.NodeName]; !exist {
				availableNodesMap[*add.NodeName] = 1
			} else {
				availableNodesMap[*add.NodeName] += 1
			}
		}
	}

	for _, node := range nodes {
		if _, exist := availableNodesMap[node.Name]; exist {
			availableNodes = append(availableNodes, node)
			podNum += availableNodesMap[node.Name]
			records = append(records, node.Name)
		}
	}

	utils.Logf(svc, "predicate: local mode service should accept node %v for service\n", records)

	for _, node := range availableNodes {
		node.Labels["weight"] = strconv.Itoa(int(availableNodesMap[node.Name] / podNum * 100.0))
	}
	return availableNodes, nil
}
