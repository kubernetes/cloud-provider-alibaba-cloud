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
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"net"
	"time"

	"strings"

	"github.com/golang/glog"
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
	v1node "k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/metrics"
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
	RouteTables(clusterName string) ([]string, error)
	// ListRoutes lists all managed routes that belong to the specified clusterName
	ListRoutes(clusterName string, table string) ([]*cloudprovider.Route, error)
	// CreateRoute creates the described managed route
	// route.Name will be ignored, although the cloud-provider may use nameHint
	// to create a more user-meaningful name.
	CreateRoute(clusterName string, nameHint string, table string, route *cloudprovider.Route) error
	// DeleteRoute deletes the specified managed route
	// Route should be as returned by ListRoutes
	DeleteRoute(clusterName string, table string, route *cloudprovider.Route) error
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
}

// New new route controller
func New(routes Routes,
	kubeClient clientset.Interface,
	nodeInformer coreinformers.NodeInformer,
	clusterName string, clusterCIDR *net.IPNet) (*RouteController, error) {

	if kubeClient != nil && kubeClient.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage(ROUTE_CONTROLLER, kubeClient.CoreV1().RESTClient().GetRateLimiter())
	}

	if clusterCIDR == nil {
		return nil, fmt.Errorf("RouteController: Must specify clusterCIDR")
	}

	eventer, caster := broadcaster()

	return &RouteController{
		routes:           routes,
		kubeClient:       kubeClient,
		clusterName:      clusterName,
		clusterCIDR:      clusterCIDR,
		nodeLister:       nodeInformer.Lister(),
		nodeListerSynced: nodeInformer.Informer().HasSynced,
		broadcaster:      caster,
		recorder:         eventer,
	}, nil
}

// Run start route controller
func (rc *RouteController) Run(stopCh <-chan struct{}, syncPeriod time.Duration) {
	defer utilruntime.HandleCrash()

	glog.Info("starting route controller")
	defer glog.Info("shutting down route controller")

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
			glog.Errorf("Couldn't reconcile node routes: %v", err)
		}
	}, syncPeriod, stopCh)

	<-stopCh
}

func (rc *RouteController) reconcile() error {
	nodes, err := rc.nodeLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error listing nodes: %v", err)
	}
	tabs, err := rc.routes.RouteTables(rc.clusterName)
	if err != nil {
		return fmt.Errorf("RouteTables: %s", err.Error())
	}
	for _, table := range tabs {
		//ListRoutes & Sync
		routeList, err := rc.routes.ListRoutes(rc.clusterName, table)
		if err != nil {
			return fmt.Errorf("error listing routes: %v", err)
		}
		if err := rc.sync(table, nodes, routeList); err != nil {
			return fmt.Errorf("reconcile route for table [%s] error: %s", table, err.Error())
		}
	}
	return nil
}

// Aoxn: Alibaba cloud does not support concurrent route operation
func (rc *RouteController) sync(table string, nodes []*v1.Node, routes []*cloudprovider.Route) error {

	//try delete conflicted route from vpc route table.
	for _, route := range routes {
		if !rc.isResponsibleForRoute(route) {
			continue
		}

		// Check if this route is a blackhole, or applies to a node we know about & has an incorrect CIDR.
		if route.Blackhole || rc.isRouteConflicted(nodes, route) {

			// Aoxn: Alibaba cloud does not support concurrent route operation
			glog.Infof("Deleting route %s %s", route.Name, route.DestinationCIDR)
			if err := rc.routes.DeleteRoute(rc.clusterName, table, route); err != nil {
				glog.Errorf("Could not delete route %s %s from table %s, %s", route.Name, route.DestinationCIDR, table, err.Error())
				continue
			}
			glog.Infof("Delete route %s %s from table %s SUCCESS.", route.Name, route.DestinationCIDR, table)
		}
	}
	cached := RouteCacheMap(routes)
	// try create desired routes
	for _, node := range nodes {

		if _, exclude := node.Labels[utils.LabelNodeRoleExcludeNode]; exclude {
			continue
		}
		if node.Spec.PodCIDR == "" {
			rc.updateNetworkingCondition(types.NodeName(node.Name), false)
			continue
		}
		if node.Spec.ProviderID == "" {
			glog.Errorf("Node %s has no Provider ID, skip it", node.Name)
			continue
		}
		// ignore error return. Try it next time anyway.
		rc.tryCreateRoute(table, node, cached)
	}
	return nil
}

// RouteCacheMap return cached map for routes
func RouteCacheMap(routes []*cloudprovider.Route) map[types.NodeName]*cloudprovider.Route {
	// routeMap maps routeTargetNode->route
	routeMap := make(map[types.NodeName]*cloudprovider.Route)
	for _, route := range routes {
		if route.TargetNode != "" {
			routeMap[route.TargetNode] = route
		}
	}
	return routeMap
}

func (rc *RouteController) tryCreateRoute(table string,
	node *v1.Node,
	cache map[types.NodeName]*cloudprovider.Route) error {

	_, condition := v1node.GetNodeCondition(&node.Status, v1.NodeReady);
	if condition != nil && condition.Status == v1.ConditionUnknown {
		glog.Infof("node %s is in unknown status.Skip creating route.", node.Name)
		return nil
	}

	if node.Spec.PodCIDR == "" {
		return rc.updateNetworkingCondition(types.NodeName(node.Name), false)
	}

	if node.Spec.ProviderID == "" {
		glog.Warningf("node %s has no node.Spec.ProviderID, skip it", node.Name)
		return nil
	}
	providerID := types.NodeName(node.Spec.ProviderID)
	// Check if we have a route for this node w/ the correct CIDR.
	r := cache[providerID]
	if r == nil || r.DestinationCIDR != node.Spec.PodCIDR {
		// If not, create the route.
		route := &cloudprovider.Route{
			TargetNode:      providerID,
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

			glog.Infof("Creating route for node %s %s with hint %s", node.Name, route.DestinationCIDR, node.Name)
			err := rc.routes.CreateRoute(rc.clusterName, node.Name, table, route)
			if err != nil {
				lasterr = err
				if strings.Contains(err.Error(), "not found") {
					glog.Infof("not found route %s", err.Error())
					return true, nil
				}
				glog.Errorf("Backoff creating route: %s", err.Error())
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			msg := fmt.Sprintf("could not create route %s for node %s: %v -> %v", route.DestinationCIDR, node.Name, err, lasterr)
			if rc.recorder != nil {
				rc.recorder.Eventf(
					&v1.ObjectReference{
						Kind:      "Node",
						Name:      node.Name,
						UID:       node.UID,
						Namespace: "",
					}, v1.EventTypeWarning, "FailedToCreateRoute", msg)
			}
			glog.Error(msg)
		}
		glog.Infof("Created route for %s with %s -> %s", table, node.Name, node.Spec.PodCIDR)
		return rc.updateNetworkingCondition(types.NodeName(node.Name), err == nil)
	}
	// Update condition only if it doesn't reflect the current state.
	_, condition = v1node.GetNodeCondition(&node.Status, v1.NodeNetworkUnavailable)
	if condition != nil &&
		condition.Status == v1.ConditionTrue {
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
			msg := fmt.Sprintf("isRouteConflicted: node.Spec.PodCIDR=%s -> "+
				"route.CIDR=%s, %s", node.Spec.PodCIDR, route.DestinationCIDR, err.Error())
			if rc.recorder != nil {
				rc.recorder.Eventf(
					&v1.ObjectReference{
						Kind:      "Node",
						Name:      node.Name,
						UID:       node.UID,
						Namespace: "",
					}, v1.EventTypeWarning, "FailedToCreateRoute", msg)
			}
			glog.Error(msg)
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
			glog.Errorf("Error updating node %s: %v", nodeName, err)
			return err
		}
		glog.V(4).Infof("Error updating node %s, retrying: %v", nodeName, err)
	}
	glog.Errorf("Error updating node %s: %v", nodeName, err)
	return err
}

func (rc *RouteController) isResponsibleForRoute(route *cloudprovider.Route) bool {
	_, cidr, err := net.ParseCIDR(route.DestinationCIDR)
	if err != nil {
		glog.Errorf("Ignoring route %s, unparsable CIDR: %v", route.Name, err)
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
	caster.StartLogging(glog.Infof)
	source := v1.EventSource{Component: ROUTE_CONTROLLER}
	return caster.NewRecorder(scheme.Scheme, source), caster
}
