package route

import (
	"context"
	cmap "github.com/orcaman/concurrent-map"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	globalCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/metric"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	return add(mgr, newReconciler(mgr, ctx))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) reconcile.Reconciler {
	recon := &ReconcileRoute{
		cloud:        ctx.Provider(),
		client:       mgr.GetClient(),
		scheme:       mgr.GetScheme(),
		record:       mgr.GetEventRecorderFor("Route"),
		nodeCache:    cmap.New(),
		configRoutes: globalCtx.CFG.Global.ConfigureCloudRoutes,
	}

	if globalCtx.CFG.Global.RouteReconciliationPeriod != "" {
		duration, err := time.ParseDuration(globalCtx.CFG.Global.RouteReconciliationPeriod)
		if err == nil || duration != 0 {
			recon.reconcilePeriod = duration
		}
	}
	if recon.reconcilePeriod == 0 {
		recon.reconcilePeriod = 5 * time.Minute
	}

	if recon.configRoutes {
		recon.periodicalSync()
	}
	return recon
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(
		"route-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
		},
	)
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AutoRepair
	return c.Watch(
		&source.Kind{
			Type: &corev1.Node{},
		},
		&handler.EnqueueRequestForObject{},
		&predicateForNodeEvent{},
	)
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
	nodepool := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, nodepool)
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

	return reconcile.Result{}, r.syncCloudRoute(ctx, nodepool)
}

func (r *ReconcileRoute) syncCloudRoute(ctx context.Context, node *corev1.Node) error {
	klog.Infof("sync routes for node: %v", node.Name)
	if !r.configRoutes || helper.HasExcludeLabel(node) {
		return nil
	}

	readyCondition, ok := helper.FindCondition(node.Status.Conditions, corev1.NodeReady)
	if ok && readyCondition.Status == corev1.ConditionUnknown {
		klog.Infof("node %s is in unknown status.Skip creating route.", node.Name)
		return nil
	}

	prvdId := node.Spec.ProviderID
	if prvdId == "" {
		klog.Warningf("provider id not exist, skip %s route config", node.Name)
		return nil
	}

	_, ipv4RouteCidr, err := getIPv4RouteForNode(node)
	if err != nil || ipv4RouteCidr == "" {
		if err1 := r.updateNetworkingCondition(ctx, node, false); err1 != nil {
			klog.Errorf("route, update network condition error: %v", err1)
		}
		return err
	}

	tables, err := getRouteTables(ctx, r.cloud)
	if err != nil {
		r.record.Eventf(node, corev1.EventTypeWarning, "DescriberRouteTablesFailed", "Describe RouteTables Failed reason: %s", err)
		return err
	}
	var tablesErr []error
	for _, table := range tables {
		if node.DeletionTimestamp == nil {
			tablesErr = append(tablesErr, r.addRouteForNode(ctx, table, ipv4RouteCidr, prvdId, node, nil))
		}
	}
	if utilerrors.NewAggregate(tablesErr) != nil {
		return r.updateNetworkingCondition(ctx, node, false)
	} else {
		networkCondition, ok := helper.FindCondition(node.Status.Conditions, corev1.NodeNetworkUnavailable)
		if ok && networkCondition.Status != corev1.ConditionFalse {
			// Update condition only if it doesn't reflect the current state.
			return nil
		}
		return r.updateNetworkingCondition(ctx, node, true)
	}
}

func (r *ReconcileRoute) addRouteForNode(ctx context.Context, table, ipv4Cidr, prvdId string, node *corev1.Node, cachedRouteEntry []*model.Route) error {
	var err error
	start := time.Now()
	route, findErr := findRoute(ctx, table, prvdId, ipv4Cidr, cachedRouteEntry, r.cloud)
	if findErr != nil {
		klog.Errorf("error found exist route for instance: %v, %v", prvdId, findErr)
		r.record.Eventf(node, corev1.EventTypeWarning, "DescriberRouteFailed", "Describe Route Failed for %s reason: %s", table, findErr)
		return nil
	}
	if route == nil || route.DestinationCIDR != ipv4Cidr {
		route, err = createRouteForInstance(ctx, table, prvdId, ipv4Cidr, r.cloud)
		if err != nil {
			klog.Errorf("error add route for instance: %v, %v, %v", prvdId, table, err)
			r.record.Eventf(node, corev1.EventTypeWarning, "CreateRouteFailed", "Create Route Failed for %s reason: %s", table, err)
		} else {
			r.record.Eventf(node, corev1.EventTypeNormal, "CreatedRoute", "Created route for %s with %s -> %s successfully", table, node.Name, ipv4Cidr)
		}
	}
	if route != nil {
		r.nodeCache.SetIfAbsent(node.Name, route)
	}
	metric.RouteLatency.WithLabelValues("create").Observe(metric.MsSince(start))
	return err
}

func (r *ReconcileRoute) updateNetworkingCondition(ctx context.Context, node *corev1.Node, routeCreated bool) error {
	var err error
	for i := 0; i < updateNodeStatusMaxRetries; i++ {
		// Patch could also fail, even though the chance is very slim. So we still do
		// patch in the retry loop.
		patch := client.MergeFrom(node.DeepCopy())
		condition, ok := helper.FindCondition(node.Status.Conditions, corev1.NodeNetworkUnavailable)
		condition.Type = corev1.NodeNetworkUnavailable
		condition.LastHeartbeatTime = metav1.Now()
		condition.LastTransitionTime = metav1.Now()
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
			node.Status.Conditions = append(node.Status.Conditions, *condition)
		}
		err = r.client.Status().Patch(ctx, node, patch)
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

	nodes, err := node.NodeList(r.client)
	if err != nil {
		klog.Errorf("reconcile: error listing nodes: %v", err)
		return
	}

	tables, err := getRouteTables(ctx, r.cloud)
	if err != nil {
		klog.Errorf("reconcile: error get RouteTables: %v", err)
	}

	for _, table := range tables {
		// Sync for Nodes
		err := r.syncTableRoutes(ctx, table, nodes)
		if err != nil {
			klog.Errorf("reconcile route for table [%s] error: %s", table, err.Error())
		}
	}
	metric.RouteLatency.WithLabelValues("reconcile").Observe(metric.MsSince(start))
}
