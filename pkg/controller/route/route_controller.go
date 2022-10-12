package route

import (
	"context"
	"fmt"
	cmap "github.com/orcaman/concurrent-map"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/metric"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	return add(mgr, newReconciler(mgr, ctx))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) *ReconcileRoute {
	recon := &ReconcileRoute{
		cloud:           ctx.Provider(),
		client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		record:          mgr.GetEventRecorderFor("route-controller"),
		nodeCache:       cmap.New(),
		configRoutes:    ctrlCfg.ControllerCFG.ConfigureCloudRoutes,
		reconcilePeriod: ctrlCfg.ControllerCFG.RouteReconciliationPeriod.Duration,
	}
	return recon
}

type routeController struct {
	c     controller.Controller
	recon *ReconcileRoute
}

// Start() function will not be called until the resource lock is acquired
func (controller routeController) Start(ctx context.Context) error {
	if controller.recon.configRoutes {
		controller.recon.periodicalSync()
	}
	return controller.c.Start(ctx)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileRoute) error {
	// Create a new controller
	c, err := controller.NewUnmanaged(
		"route-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
			RecoverPanic:            true,
		},
	)
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AutoRepair
	err = c.Watch(
		&source.Kind{
			Type: &corev1.Node{},
		},
		&handler.EnqueueRequestForObject{},
		&predicateForNodeEvent{},
	)
	if err != nil {
		return err
	}

	return mgr.Add(&routeController{c: c, recon: r})
}

// ReconcileRoute implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRoute{}

// ReconcileRoute reconciles a AutoRepair object
type ReconcileRoute struct {
	cloud  prvd.Provider
	client client.Client
	scheme *runtime.Scheme

	// configuration fields
	reconcilePeriod time.Duration
	configRoutes    bool

	nodeCache cmap.ConcurrentMap

	//record event recorder
	record record.EventRecorder
}

func (r *ReconcileRoute) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// if do not need route, skip all node events
	if !r.configRoutes {
		return reconcile.Result{}, nil
	}

	reconcileNode := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, reconcileNode)
	if err != nil {
		if errors.IsNotFound(err) {
			if o, ok := r.nodeCache.Get(request.Name); ok {
				if route, ok := o.(*model.Route); ok {
					start := time.Now()
					tables, err := getRouteTables(ctx, r.cloud)
					if err != nil {
						klog.Errorf("error get route tables to delete node %s route %v, error: %v", request.Name, route, err)
						return reconcile.Result{}, err
					}
					var errList []error
					for _, table := range tables {
						if err = deleteRouteForInstance(ctx, table, route.ProviderId, route.DestinationCIDR, r.cloud); err != nil {
							errList = append(errList, err)
							klog.Errorf("error delete route entry for delete node %s route %v, error: %v", request.Name, route, err)
						} else {
							klog.Infof("successfully delete route entry for node %s route %s", request.Name, route)
						}
					}
					metric.RouteLatency.WithLabelValues("delete").Observe(metric.MsSince(start))
					if aggrErr := utilerrors.NewAggregate(errList); aggrErr == nil {
						r.nodeCache.Remove(request.Name)
					} else {
						// requeue for remove error
						return reconcile.Result{}, aggrErr
					}
				}
			}
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	err = r.syncCloudRoute(ctx, reconcileNode)
	if err != nil {
		klog.Errorf("add route for node %s failed, err: %s", reconcileNode.Name, err.Error())
		nodeRef := &corev1.ObjectReference{
			Kind:      "Node",
			Name:      reconcileNode.Name,
			UID:       types.UID(reconcileNode.Name),
			Namespace: "",
		}
		r.record.Event(nodeRef, corev1.EventTypeWarning, helper.FailedSyncRoute, "sync cloud route failed")
	}
	// no need to retry, reconcileForCluster() will reconcile routes periodically
	return reconcile.Result{}, nil
}

func (r *ReconcileRoute) syncCloudRoute(ctx context.Context, node *corev1.Node) error {
	if !needSyncRoute(node) {
		return nil
	}

	prvdId := node.Spec.ProviderID
	if prvdId == "" {
		klog.Warningf("node %s provider id is not exist, skip creating route", node.Name)
		return nil
	}

	_, ipv4RouteCidr, err := getIPv4RouteForNode(node)
	if err != nil || ipv4RouteCidr == "" {
		klog.Warningf("node %s parse podCIDR %s error, skip creating route", node.Name, node.Spec.PodCIDR)
		if err1 := r.updateNetworkingCondition(ctx, node, false); err1 != nil {
			klog.Errorf("route, update network condition error: %v", err1)
		}
		return err
	}

	tables, err := getRouteTables(ctx, r.cloud)
	if err != nil {
		return err
	}
	var tablesErr []error
	for _, table := range tables {
		tablesErr = append(tablesErr, r.addRouteForNode(ctx, table, ipv4RouteCidr, prvdId, node, nil))
	}
	if utilerrors.NewAggregate(tablesErr) != nil {
		err := r.updateNetworkingCondition(ctx, node, false)
		if err != nil {
			klog.Errorf("update network condition for node %s, error: %v", node.Name, err.Error())
		}
		return utilerrors.NewAggregate(tablesErr)
	} else {
		return r.updateNetworkingCondition(ctx, node, true)
	}
}

func (r *ReconcileRoute) addRouteForNode(
	ctx context.Context, table, ipv4Cidr, prvdId string, node *corev1.Node, cachedRouteEntry []*model.Route,
) error {
	var err error
	nodeRef := &corev1.ObjectReference{
		Kind:      "Node",
		Name:      node.Name,
		UID:       types.UID(node.Name),
		Namespace: "",
	}

	route, findErr := findRoute(ctx, table, prvdId, ipv4Cidr, cachedRouteEntry, r.cloud)
	if findErr != nil {
		klog.Errorf("error found exist route for instance: %v, %v", prvdId, findErr)
		r.record.Event(
			nodeRef,
			corev1.EventTypeWarning,
			"DescriberRouteFailed",
			fmt.Sprintf("Describe Route Failed for %s reason: %s", table, helper.GetLogMessage(findErr)),
		)
		return nil
	}

	// route not found, try to create route
	if route == nil || route.DestinationCIDR != ipv4Cidr {
		klog.Infof("create routes for node %s: %v - %v", node.Name, prvdId, ipv4Cidr)
		start := time.Now()
		route, err = createRouteForInstance(ctx, table, prvdId, ipv4Cidr, r.cloud)
		if err != nil {
			klog.Errorf("error create route for node %v : instance id [%v], route [%v], err: %s", node.Name, prvdId, table, err.Error())
			r.record.Event(
				nodeRef,
				corev1.EventTypeWarning,
				helper.FailedCreateRoute,
				fmt.Sprintf("Error creating route entry in %s: %s", table, helper.GetLogMessage(err)),
			)
		} else {
			klog.Infof("Created route for %s with %s - %s successfully", table, node.Name, ipv4Cidr)
			r.record.Event(
				nodeRef,
				corev1.EventTypeNormal,
				helper.SucceedCreateRoute,
				fmt.Sprintf("Created route for %s with %s -> %s successfully", table, node.Name, ipv4Cidr),
			)
		}
		metric.RouteLatency.WithLabelValues("create").Observe(metric.MsSince(start))
	}
	if route != nil {
		r.nodeCache.SetIfAbsent(node.Name, route)
	}
	return err
}

func (r *ReconcileRoute) updateNetworkingCondition(ctx context.Context, node *corev1.Node, routeCreated bool) error {
	networkCondition, ok := helper.FindCondition(node.Status.Conditions, corev1.NodeNetworkUnavailable)
	if routeCreated && ok && networkCondition.Status == corev1.ConditionFalse {
		klog.V(2).Infof("set node %v with NodeNetworkUnavailable=false was canceled because it is already set", node.Name)
		return nil
	}

	if !routeCreated && ok && networkCondition.Status == corev1.ConditionTrue {
		klog.V(2).Infof("set node %v with NodeNetworkUnavailable=true was canceled because it is already set", node.Name)
		return nil
	}

	klog.Infof("Patching node status %v with %v previous condition was:%+v", node.Name, routeCreated, networkCondition)
	var err error
	for i := 0; i < updateNodeStatusMaxRetries; i++ {
		// Patch could also fail, even though the chance is very slim. So we still do
		// patch in the retry loop.
		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*corev1.Node)
			condition, ok := helper.FindCondition(nins.Status.Conditions, corev1.NodeNetworkUnavailable)
			condition.Type = corev1.NodeNetworkUnavailable
			condition.LastTransitionTime = metav1.Now()
			condition.LastHeartbeatTime = metav1.Now()
			if routeCreated {
				condition.Status = corev1.ConditionFalse
				condition.Reason = "RouteCreated"
				condition.Message = "RouteController created a route"
			} else {
				condition.Status = corev1.ConditionTrue
				condition.Reason = "NoRouteCreated"
				condition.Message = "RouteController failed to create a route"
			}
			if !ok {
				nins.Status.Conditions = append(nins.Status.Conditions, *condition)
			}
			return nins, nil
		}
		err = helper.PatchM(r.client, node, diff, helper.PatchStatus)
		if err == nil {
			return nil
		}
		if !errors.IsConflict(err) {
			klog.Errorf("Error updating node %s: %v", node.Name, err)
			return err
		}
		klog.V(4).Infof("Error updating node %s, retrying: %v", node.Name, err)
	}
	klog.Errorf("Error updating node %s: %v", node.Name, err)
	return err
}

func (r *ReconcileRoute) periodicalSync() {
	go wait.Until(r.reconcileForCluster, r.reconcilePeriod, wait.NeverStop)
}

func (r *ReconcileRoute) reconcileForCluster() {
	ctx := context.Background()
	start := time.Now()
	defer func() {
		metric.RouteLatency.WithLabelValues("reconcile").Observe(metric.MsSince(start))
	}()

	nodes, err := node.NodeList(r.client)
	if err != nil {
		klog.Errorf("reconcile: error listing nodes: %v", err)
		return
	}

	tables, err := getRouteTables(ctx, r.cloud)
	if err != nil {
		klog.Errorf("sync route tables error: get RouteTables: %v", err)
		r.record.Event(
			&corev1.Event{ObjectMeta: metav1.ObjectMeta{Name: "route-controller"}},
			corev1.EventTypeWarning, helper.FailedSyncRoute,
			fmt.Sprintf("Reconciling route error: %s", err.Error()),
		)
		return
	}

	var failedTableIds []string
	for _, table := range tables {
		// Sync for nodes
		if err := r.syncTableRoutes(ctx, table, nodes); err != nil {
			failedTableIds = append(failedTableIds, table)
			klog.Errorf("sync route tables error: sync table [%s] error: %s", table, err.Error())
		}
	}

	if len(failedTableIds) != 0 {
		r.record.Event(
			&corev1.Event{ObjectMeta: metav1.ObjectMeta{Name: "route-controller"}},
			corev1.EventTypeWarning, helper.FailedSyncRoute,
			fmt.Sprintf("Error reconciling route, reconcile table [%s] failed", strings.Join(failedTableIds, ",")),
		)
	} else {
		klog.Infof("sync route tables successfully, tables [%s]", strings.Join(tables, ","))
	}
}
