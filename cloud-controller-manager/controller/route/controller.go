/*
Copyright 2018 The Kubernetes Authors.

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

package route

import (
	"context"
	"fmt"
	"k8s.io/client-go/util/workqueue"
	queue "k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils/metric"
	"k8s.io/klog"
	"net"
	"reflect"
	"time"

	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	//v1node "k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/cloud-provider"
	"k8s.io/cloud-provider/node/helpers"
	metrics "k8s.io/component-base/metrics/prometheus/ratelimiter"
	controller "k8s.io/kube-aggregator/pkg/controllers"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
)

const (
	// ROUTE_CONTROLLER route controller name
	ROUTE_CONTROLLER = "route-controller"

	// Maximum number of retries of node status update.
	updateNodeStatusMaxRetries int = 3
)

// Routes is an abstract, pluggable interface for advanced routing rules.
type Routes interface {
	// RouteTables get all available route tables.
	RouteTables(ctx context.Context, clusterName string) ([]string, error)
	// ListRoutes lists all managed routes that belong to the specified clusterName
	ListRoutes(ctx context.Context, clusterName string, table string) ([]*cloudprovider.Route, error)
	// CreateRoute creates the described managed route
	// route.Name will be ignored, although the cloud-provider may use nameHint
	// to create a more user-meaningful name.
	CreateRoute(ctx context.Context, clusterName string, nameHint string, table string, route *cloudprovider.Route) error
	// DeleteRoute deletes the specified managed route
	// Route should be as returned by ListRoutes
	DeleteRoute(ctx context.Context, clusterName string, table string, route *cloudprovider.Route) error
}

// RouteController response for route reconcile
type RouteController struct {
	routes           Routes
	kubeClient       clientset.Interface
	clusterName      string
	clusterCIDR      *net.IPNet
	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced
	broadcaster      record.EventBroadcaster
	recorder         record.EventRecorder
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

const NODE_QUEUE = "node.queue"

// New new route controller
func New(routes Routes,
	kubeClient clientset.Interface,
	nodeInformer coreinformers.NodeInformer,
	clusterName string, clusterCIDR *net.IPNet) (*RouteController, error) {

	if kubeClient != nil && kubeClient.CoreV1().RESTClient().GetRateLimiter() != nil {
		err := metrics.RegisterMetricAndTrackRateLimiterUsage(
			ROUTE_CONTROLLER,
			kubeClient.CoreV1().RESTClient().GetRateLimiter(),
		)
		if err != nil {
			klog.Warningf("metrics initialized fail. %s", err.Error())
		}
	}

	if clusterCIDR == nil {
		return nil, fmt.Errorf("RouteController: Must specify clusterCIDR")
	}

	eventer, caster := broadcaster()

	rc := &RouteController{
		routes:           routes,
		kubeClient:       kubeClient,
		clusterName:      clusterName,
		clusterCIDR:      clusterCIDR,
		nodeLister:       nodeInformer.Lister(),
		nodeListerSynced: nodeInformer.Informer().HasSynced,
		broadcaster:      caster,
		recorder:         eventer,
		queues: map[string]queue.DelayingInterface{
			NODE_QUEUE: workqueue.NewNamedDelayingQueue(NODE_QUEUE),
		},
	}

	rc.HandlerForNodeDeletion(
		rc.queues[NODE_QUEUE],
		nodeInformer.Informer(),
	)

	return rc, nil
}

func (rc *RouteController) HandlerForNodeDeletion(
	que queue.DelayingInterface,
	informer cache.SharedIndexInformer,
) {
	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(nodec interface{}) {
				node, ok := nodec.(*v1.Node)
				if !ok {
					klog.Infof("not node type: %s\n", reflect.TypeOf(nodec))
					return
				}
				if _, exclude := node.Labels[utils.LabelNodeRoleExcludeNode]; exclude {
					klog.Infof("ignore node with exclude node label %s", node.Name)
					return
				}
				que.Add(node)
				klog.Infof("node deletion event: %s, %s", node.Name, node.Spec.ProviderID)
			},
		},
	)
}

// Run start route controller
func (rc *RouteController) Run(stopCh <-chan struct{}, syncPeriod time.Duration) {
	defer utilruntime.HandleCrash()

	klog.Info("starting route controller")
	defer klog.Info("shutting down route controller")

	if !controller.WaitForCacheSync(ROUTE_CONTROLLER, stopCh, rc.nodeListerSynced) {
		return
	}

	if rc.broadcaster != nil {
		sink := &v1core.EventSinkImpl{
			Interface: v1core.New(rc.kubeClient.CoreV1().RESTClient()).Events(""),
		}
		rc.broadcaster.StartRecordingToSink(sink)
	}

	// TODO: If we do just the full Resync every 5 minutes (default value)
	// that means that we may wait up to 5 minutes before even starting
	// creating a route for it. This is bad.
	// We should have a watch on node and if we observe a new node (with CIDR?)
	// trigger reconciliation for that node.
	go wait.NonSlidingUntil(func() {
		if err := rc.reconcile(); err != nil {
			klog.Errorf("Couldn't reconcile node routes: %v", err)
		}
	}, syncPeriod, stopCh)

	go wait.Until(
		func() {
			que := rc.queues[NODE_QUEUE]
			for {
				func() {
					// Workerqueue ensures that a single key would not be process
					// by two worker concurrently, so multiple workers is safe here.
					key, quit := que.Get()
					if quit {
						return
					}
					defer que.Done(key)
					node, ok := key.(*v1.Node)
					if !ok {
						klog.Errorf("not type of *v1.Node, %s", reflect.TypeOf(key))
						return
					}
					klog.Infof("worker: queued sync for [%s] node deletion with route", node.Name)
					start := time.Now()
					if err := rc.syncd(node); err != nil {
						que.AddAfter(key, 2*time.Minute)
						klog.Errorf("requeue: sync route for node %s, error %v", node.Name, err)
					}
					metric.RouteLatency.WithLabelValues("delete").Observe(metric.MsSince(start))
				}()
			}
		},
		2*time.Second,
		stopCh,
	)
	<-stopCh
}

func (rc *RouteController) syncd(node *v1.Node) error {
	if _, exclude := node.Labels[utils.LabelNodeRoleExcludeNode]; exclude {
		return nil
	}
	if node.Spec.PodCIDR == "" {
		klog.Warningf("Node %s PodCIDR is nil, skip delete route", node.Name)
		return nil
	}
	if node.Spec.ProviderID == "" {
		klog.Warningf("Node %s has no Provider ID, skip delete route", node.Name)
		return nil
	}

	ctx := context.Background()
	tabs, err := rc.routes.RouteTables(ctx, rc.clusterName)
	if err != nil {
		return fmt.Errorf("RouteTables: %s", err.Error())
	}
	for _, table := range tabs {
		route := &cloudprovider.Route{
			Name:            node.Spec.ProviderID,
			TargetNode:      types.NodeName(node.Spec.ProviderID),
			DestinationCIDR: node.Spec.PodCIDR,
		}
		if err := rc.routes.DeleteRoute(
			ctx, rc.clusterName, table, route,
		); err != nil {
			klog.Errorf(
				"delete route %s %s from table %s, %s", route.Name, route.DestinationCIDR, table, err.Error())
			return fmt.Errorf("node deletion, delete route error: %s", err.Error())
		}
		klog.Infof("node deletion: delete route %s %s from table %s SUCCESS.", route.Name, route.DestinationCIDR, table)
	}
	return nil
}

func (rc *RouteController) reconcile() error {
	ctx := context.Background()
	start := time.Now()
	nodes, err := rc.nodeLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listing nodes: %v", err)
	}
	tabs, err := rc.routes.RouteTables(ctx, rc.clusterName)
	if err != nil {
		return fmt.Errorf("RouteTables: %s", err.Error())
	}
	for _, table := range tabs {
		//ListRoutes & Sync
		routeList, err := rc.routes.ListRoutes(ctx, rc.clusterName, table)
		if err != nil {
			return fmt.Errorf("error listing routes: %v", err)
		}
		if err := rc.sync(ctx, table, nodes, routeList); err != nil {
			return fmt.Errorf("reconcile route for table [%s] error: %s", table, err.Error())
		}
	}
	metric.RouteLatency.WithLabelValues("reconcile").Observe(metric.MsSince(start))
	return nil
}

// Aoxn: Alibaba cloud does not support concurrent route operation
func (rc *RouteController) sync(ctx context.Context, table string, nodes []*v1.Node, routes []*cloudprovider.Route) error {

	//try delete conflicted route from vpc route table.
	for _, route := range routes {
		if !rc.isResponsibleForRoute(route) {
			continue
		}

		// Check if this route is a blackhole, or applies to a node we know about & has an incorrect CIDR.
		if route.Blackhole || rc.isRouteConflicted(nodes, route) {

			// Aoxn: Alibaba cloud does not support concurrent route operation
			klog.Infof("Deleting route %s %s", route.Name, route.DestinationCIDR)
			if err := rc.routes.DeleteRoute(ctx, rc.clusterName, table, route); err != nil {
				klog.Errorf("Could not delete route %s %s from table %s, %s", route.Name, route.DestinationCIDR, table, err.Error())
				continue
			}
			klog.Infof("Delete route %s %s from table %s SUCCESS.", route.Name, route.DestinationCIDR, table)
		}
	}
	cached := RouteCacheMap(routes)
	// try create desired routes
	for _, node := range nodes {

		if _, exclude := node.Labels[utils.LabelNodeRoleExcludeNode]; exclude {
			continue
		}
		if node.Spec.PodCIDR == "" {
			err := rc.updateNetworkingCondition(types.NodeName(node.Name), false)
			if err != nil {
				klog.Errorf("route, update network condition error: %s", err.Error())
			}
			continue
		}
		if node.Spec.ProviderID == "" {
			klog.Errorf("Node %s has no Provider ID, skip it", node.Name)
			continue
		}
		// ignore error return. Try it next time anyway.
		err := rc.tryCreateRoute(ctx, table, node, cached)
		if err != nil {
			klog.Errorf("try create route error: %s", err.Error())
		}
	}
	return nil
}

// RouteCacheMap return cached map for routes
func RouteCacheMap(routes []*cloudprovider.Route) map[string]*cloudprovider.Route {
	// routeMap maps routeTargetNode+routeDestinationCIDR->route
	routeMap := make(map[string]*cloudprovider.Route)
	for _, route := range routes {
		if route.TargetNode != "" && route.DestinationCIDR != "" {
			routeKey := fmt.Sprintf("%s-%s", route.TargetNode, route.DestinationCIDR)
			routeMap[routeKey] = route
		}
	}
	return routeMap
}

func (rc *RouteController) tryCreateRoute(
	ctx context.Context,
	table string,
	node *v1.Node,
	cache map[string]*cloudprovider.Route,
) error {

	_, condition := helpers.GetNodeCondition(&node.Status, v1.NodeReady)
	if condition != nil && condition.Status == v1.ConditionUnknown {
		klog.Infof("node %s is in unknown status.Skip creating route.", node.Name)
		return nil
	}

	if node.Spec.PodCIDR == "" {
		return rc.updateNetworkingCondition(types.NodeName(node.Name), false)
	}

	if node.Spec.ProviderID == "" {
		klog.Warningf("node %s has no node.Spec.ProviderID, skip it", node.Name)
		return nil
	}
	providerID := node.Spec.ProviderID
	destinationCIDR := node.Spec.PodCIDR
	// Check if we have a route for this node w/ the correct CIDR.
	routeKey := fmt.Sprintf("%s-%s", providerID, destinationCIDR)
	r := cache[routeKey]
	if r == nil || r.DestinationCIDR != node.Spec.PodCIDR {
		start := time.Now()
		// If not, create the route.
		route := &cloudprovider.Route{
			TargetNode:      types.NodeName(providerID),
			DestinationCIDR: node.Spec.PodCIDR,
		}

		backoff := wait.Backoff{
			Duration: 4 * time.Second,
			Steps:    3,
			Factor:   2,
			Jitter:   1,
		}
		var lasterr error
		err := wait.ExponentialBackoff(backoff, func() (bool, error) {

			klog.Infof("Creating route for node %s %s with hint %s", node.Name, route.DestinationCIDR, node.Name)
			err := rc.routes.CreateRoute(ctx, rc.clusterName, node.Name, table, route)
			if err != nil {
				lasterr = err
				if strings.Contains(err.Error(), "not found") {
					klog.Infof("not found route %s", err.Error())
					return true, nil
				}
				klog.Errorf("Backoff creating route: %s", err.Error())
				return false, nil
			}
			return true, nil
		})

		ref := &v1.ObjectReference{
			Kind:      "Node",
			Name:      node.Name,
			UID:       node.UID,
			Namespace: "",
		}
		if err != nil {
			rc.recorder.Eventf(
				ref,
				v1.EventTypeWarning,
				"Failed",
				"Fail to create route, error: %s",
				err.Error())
			klog.Errorf("could not create route %s for node %s: %v -> %v", route.DestinationCIDR, node.Name, err, lasterr)
		} else {
			rc.recorder.Eventf(
				ref,
				v1.EventTypeNormal,
				"SuccessfulCreate",
				"Created route for %s with %s -> %s successfully",
				table, node.Name, node.Spec.PodCIDR,
			)
			klog.Infof("Created route for %s with %s -> %s", table, node.Name, node.Spec.PodCIDR)
		}
		metric.RouteLatency.WithLabelValues("create").Observe(metric.MsSince(start))
		klog.Infof("Created route for %s with %s -> %s", table, node.Name, node.Spec.PodCIDR)
		return rc.updateNetworkingCondition(types.NodeName(node.Name), err == nil)
	}
	// Update condition only if it doesn't reflect the current state.
	_, condition = helpers.GetNodeCondition(&node.Status, v1.NodeNetworkUnavailable)
	if condition != nil &&
		condition.Status == v1.ConditionFalse {
		return nil
	}
	return rc.updateNetworkingCondition(types.NodeName(node.Name), true)
}

func (rc *RouteController) isRouteConflicted(nodes []*v1.Node, route *cloudprovider.Route) bool {
	for _, node := range nodes {
		// skip node without podcidr
		if node.Spec.PodCIDR == "" {
			continue
		}

		if node.Spec.PodCIDR == route.DestinationCIDR &&
			!strings.Contains(node.Spec.ProviderID, string(route.TargetNode)) {
			// conflicted with exist route.
			return true
		}
		contains, err := RealContainsCidr(node.Spec.PodCIDR, route.DestinationCIDR)
		if err != nil {
			// record event an error out.
			if rc.recorder != nil {
				rc.recorder.Eventf(
					&v1.ObjectReference{
						Kind:      "Node",
						Name:      node.Name,
						UID:       node.UID,
						Namespace: "",
					},
					v1.EventTypeWarning,
					"Failed",
					"Fail to reconcile route, route conflict error:  %s",
					err.Error(),
				)
			}
			klog.Errorf("route conflicted: node.Spec.PodCIDR=%s -> "+
				"route.CIDR=%s, %s", node.Spec.PodCIDR, route.DestinationCIDR, err.Error())
			return false
		}
		if contains {
			return true
		}
	}
	return false
}

func (rc *RouteController) updateNetworkingCondition(nodeName types.NodeName, routeCreated bool) error {
	var err error
	for i := 0; i < updateNodeStatusMaxRetries; i++ {
		// Patch could also fail, even though the chance is very slim. So we still do
		// patch in the retry loop.
		currentTime := metav1.Now()
		if routeCreated {
			err = nodeutil.SetNodeCondition(rc.kubeClient, nodeName, v1.NodeCondition{
				Type:               v1.NodeNetworkUnavailable,
				Status:             v1.ConditionFalse,
				Reason:             "RouteCreated",
				Message:            "RouteController created a route",
				LastTransitionTime: currentTime,
			})
		} else {
			err = nodeutil.SetNodeCondition(rc.kubeClient, nodeName, v1.NodeCondition{
				Type:               v1.NodeNetworkUnavailable,
				Status:             v1.ConditionTrue,
				Reason:             "NoRouteCreated",
				Message:            "RouteController failed to create a route",
				LastTransitionTime: currentTime,
			})
		}
		if err == nil {
			return nil
		}
		if !errors.IsConflict(err) {
			klog.Errorf("Error updating node %s: %v", nodeName, err)
			return err
		}
		klog.V(4).Infof("Error updating node %s, retrying: %v", nodeName, err)
	}
	klog.Errorf("Error updating node %s: %v", nodeName, err)
	return err
}

func (rc *RouteController) isResponsibleForRoute(route *cloudprovider.Route) bool {
	_, cidr, err := net.ParseCIDR(route.DestinationCIDR)
	if err != nil {
		klog.Errorf("Ignoring route %s, unparsable CIDR: %v", route.Name, err)
		return false
	}
	// Not responsible if this route's CIDR is not within our clusterCIDR
	lastIP := make([]byte, len(cidr.IP))
	for i := range lastIP {
		lastIP[i] = cidr.IP[i] | ^cidr.Mask[i]
	}
	if !rc.clusterCIDR.Contains(cidr.IP) || !rc.clusterCIDR.Contains(lastIP) {
		return false
	}
	return true
}

func broadcaster() (record.EventRecorder, record.EventBroadcaster) {
	caster := record.NewBroadcaster()
	caster.StartLogging(klog.Infof)
	source := v1.EventSource{Component: ROUTE_CONTROLLER}
	return caster.NewRecorder(scheme.Scheme, source), caster
}
