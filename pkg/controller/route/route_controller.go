package route

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/metric"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = klogr.New().WithName("route-controller")

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	requeueChan := make(chan event.GenericEvent, ctrlCfg.ControllerCFG.RouteReconcileBatchSize*ctrlCfg.CloudCFG.Global.RouteMaxConcurrentReconciles)
	rateLimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed, and it's only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)

	r := newReconciler(mgr, ctx, requeueChan, rateLimiter)

	recoverPanic := true
	// Create a new controller
	c, err := controller.NewUnmanaged(
		"route-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
			RecoverPanic:            &recoverPanic,
		},
	)
	if err != nil {
		return err
	}

	enqueueRequest := &enqueueRequestForNodeEvent{
		rateLimiter: rateLimiter,
	}
	// Watch for changes to primary resource AutoRepair
	if err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Node{}),
		enqueueRequest,
		&predicateForNodeEvent{},
	); err != nil {
		return err
	}

	if err := c.Watch(
		&source.Channel{Source: requeueChan},
		enqueueRequest); err != nil {
		return err
	}

	return mgr.Add(&routeController{c: c, recon: r})
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext, requeue chan<- event.GenericEvent, rateLimiter workqueue.RateLimiter) *ReconcileRoute {

	recon := &ReconcileRoute{
		cloud:           ctx.Provider(),
		client:          mgr.GetClient(),
		scheme:          mgr.GetScheme(),
		record:          mgr.GetEventRecorderFor("route-controller"),
		nodeCache:       cmap.New(),
		configRoutes:    ctrlCfg.ControllerCFG.ConfigureCloudRoutes,
		reconcilePeriod: ctrlCfg.ControllerCFG.RouteReconciliationPeriod.Duration,
		requeueChan:     requeue,
		requestChan:     make(chan reconcile.Request, ctrlCfg.ControllerCFG.RouteReconcileBatchSize*ctrlCfg.CloudCFG.Global.RouteMaxConcurrentReconciles),
		rateLimiter:     rateLimiter,
	}
	return recon
}

type routeController struct {
	c     controller.Controller
	recon *ReconcileRoute
}

// Start function will not be called until the resource lock is acquired
func (controller routeController) Start(ctx context.Context) error {
	if controller.recon.configRoutes {
		controller.recon.periodicalSync()
		for i := range ctrlCfg.CloudCFG.Global.RouteMaxConcurrentReconciles {
			go controller.recon.batchWorker(ctx, i)
		}
	}
	return controller.c.Start(ctx)
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

	requeueChan chan<- event.GenericEvent
	requestChan chan reconcile.Request

	rateLimiter workqueue.RateLimiter
}

func (r *ReconcileRoute) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// if do not need route, skip all node events
	if !r.configRoutes {
		return reconcile.Result{}, nil
	}
	log.Info("enqueue route reconcile request", "node", request.Name)
	r.requestChan <- request
	return reconcile.Result{}, nil
}

func (r *ReconcileRoute) batchWorker(ctx context.Context, idx int) {
	log.Info("starting batch worker", "worker", idx)
outer:
	for {
		var requests []reconcile.Request
		var names []string
		var requestMap map[string]bool
		select {
		case m := <-r.requestChan:
			requests = []reconcile.Request{m}
			names = []string{m.Name}
			requestMap = map[string]bool{m.Name: true}
			// sleep to aggregate recently events as many as possible
			time.Sleep(time.Duration(ctrlCfg.ControllerCFG.NodeEventAggregationWaitSeconds) * time.Second)
		inner:
			for {
				if len(requests) >= ctrlCfg.ControllerCFG.RouteReconcileBatchSize {
					break
				}
				select {
				case m = <-r.requestChan:
					if !requestMap[m.Name] {
						requests = append(requests, m)
						names = append(names, m.Name)
						requestMap[m.Name] = true
					} else {
						log.V(4).Info("duplicated request", "request", m)
					}
				default:
					break inner
				}
			}
		case <-ctx.Done():
			log.Info("context done, batch worker is shutting down")
			break outer
		}

		startTime := time.Now()
		reconcileID := uuid.New().String()
		log.Info("batch reconcile routes",
			"length", len(requests), "names", names, "worker", idx, "reconcileID", reconcileID)
		if err := r.batchSyncCloudRoutes(ctx, reconcileID, requests); err != nil {
			log.Error(err, "Sync routes error, requeue", "names", names, "reconcileID", reconcileID)
			r.record.Eventf(
				&corev1.Event{ObjectMeta: metav1.ObjectMeta{Name: "route-controller"}},
				corev1.EventTypeWarning, helper.FailedSyncRoute,
				"Reconciling route error: %s",
				err.Error(),
			)
			for _, req := range requests {
				r.requeueNode(&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: req.Name,
					},
				})
			}
			continue
		}
		log.Info("Successfully reconcile routes",
			"nodes", names, "worker", idx,
			"elapsedTime", time.Since(startTime).Seconds(), "reconcileID", reconcileID)
	}
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
	go func() {
		time.Sleep(r.reconcilePeriod)
		wait.Until(r.reconcileForCluster, r.reconcilePeriod, wait.NeverStop)
	}()
}

func (r *ReconcileRoute) reconcileForCluster() {
	ctx := context.Background()
	start := time.Now()
	defer func() {
		metric.RouteLatency.WithLabelValues("reconcile").Observe(metric.MsSince(start))
	}()

	nodes, err := node.NodeList(r.client, true)
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

func (r *ReconcileRoute) batchSyncCloudRoutes(ctx context.Context, reconcileID string, requests []reconcile.Request) error {
	startTime := time.Now()
	defer func() {
		metric.RouteLatency.WithLabelValues("reconcile").Observe(metric.MsSince(startTime))
	}()
	var toAddedNodes []*corev1.Node
	toAdd := map[string][]*model.Route{}
	var toDelete []*model.Route

	for _, request := range requests {
		n := &corev1.Node{}
		err := r.client.Get(ctx, request.NamespacedName, n)
		if err != nil {
			// todo: check deletion timestamp
			if errors.IsNotFound(err) {
				if o, ok := r.nodeCache.Get(request.Name); ok {
					if route, ok := o.(*model.Route); ok {
						toDelete = append(toDelete, &model.Route{
							Name:            route.Name,
							DestinationCIDR: route.DestinationCIDR,
							ProviderId:      route.ProviderId,
							NodeReference: &corev1.Node{
								ObjectMeta: metav1.ObjectMeta{
									Name: request.Name,
								},
							},
						})
					}
				}
				// node not found, ignore
				continue
			}
			log.Error(err, "error getting node, requeue", "node", request.Name, "reconcileID", reconcileID)
			r.requeueNode(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: request.Name,
				},
			})
		}
		if !needSyncRoute(n) {
			continue
		}
		if n.Spec.ProviderID == "" {
			log.Info("node's provider id is not exist, skip creating route", "node", n.Name, "reconcileID", reconcileID)
			continue
		}

		if _, _, err := util.NodeFromProviderID(n.Spec.ProviderID); err != nil {
			log.Error(err, "invalid providerID, requeue", "node", n.Name, "providerID", n.Spec.ProviderID, "reconcileID", reconcileID)
			r.requeueChan <- event.GenericEvent{Object: n}
			continue

		}

		_, ipv4RouteCidr, err := getIPv4RouteForNode(n)
		if err != nil || ipv4RouteCidr == "" {
			log.Error(err, "node parse podCIDR error, skip creating route",
				"node", n.Name, "cidr", n.Spec.PodCIDR, "reconcileID", reconcileID)
			if err1 := r.updateNetworkingCondition(ctx, n, false); err1 != nil {
				klog.Errorf("route, update network condition error: %v", err1)
			}
			r.requeueChan <- event.GenericEvent{Object: n}
			continue
		}

		toAddedNodes = append(toAddedNodes, n)
	}

	if len(toAddedNodes) == 0 && len(toDelete) == 0 {
		log.Info("no route need to be added.", "reconcileID", reconcileID)
		return nil
	}

	tables, err := getRouteTables(ctx, r.cloud)
	if err != nil {
		// todo: add event
		return err
	}

	cachedRoutes := map[string][]*model.Route{}
	for _, t := range tables {
		routes, err := r.cloud.ListRoute(ctx, t)
		if err != nil {
			return err
		}
		cachedRoutes[t] = routes
	}

	for _, n := range toAddedNodes {
		for _, table := range tables {
			_, ipv4RouteCidr, _ := getIPv4RouteForNode(n)

			route, err := findRoute(ctx, table, n.Spec.ProviderID, ipv4RouteCidr, cachedRoutes[table], r.cloud)
			if err != nil {
				log.Error(err, "error find route existence for instance", "providerID", n.Spec.ProviderID, "reconcileID", reconcileID)
				nodeRef := &corev1.ObjectReference{
					Kind: "Node",
					Name: n.Name,
					UID:  n.UID,
				}
				r.record.Event(
					nodeRef,
					corev1.EventTypeWarning,
					"DescriberRouteFailed",
					fmt.Sprintf("Describe Route Failed for %s reason: %s", table, helper.GetLogMessage(err)),
				)
				r.requeueNode(n)
				continue
			}

			if route == nil {
				toAdd[table] = append(toAdd[table], &model.Route{
					Name:            fmt.Sprintf("%s-%s", n.Spec.ProviderID, ipv4RouteCidr),
					DestinationCIDR: ipv4RouteCidr,
					ProviderId:      n.Spec.ProviderID,
					NodeReference:   n,
				})
			} else {
				r.rateLimiter.Forget(reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: n.Name,
					},
				})
			}
		}
	}

	preparedTime := time.Now()
	log.Info("Start sync routes", "reconcileID", reconcileID)

	var errs []error
	for _, t := range tables {
		err = r.batchDeleteRoutes(ctx, reconcileID, t, toDelete)
		if err != nil {
			var toDeleteNames []string
			for _, d := range toDelete {
				toDeleteNames = append(toDeleteNames, d.Name)
			}
			log.Error(err, "batch delete routes failed, requeue all routes", "table", t, "entries", toDeleteNames, "reconcileID", reconcileID)
			for _, route := range toDelete {
				r.requeueNode(route.NodeReference)
			}
			errs = append(errs, err)
		}
		err = r.batchAddRoutes(ctx, reconcileID, t, toAdd[t])
		if err != nil {
			var toAddNames []string
			for _, d := range toAdd[t] {
				toAddNames = append(toAddNames, d.Name)
			}
			log.Error(err, "batch add routes failed, requeue all routes", "table", t, "entries", toAdd[t], "reconcileID", reconcileID)
			for _, route := range toAdd[t] {
				r.requeueNode(route.NodeReference)
			}
			errs = append(errs, err)
			continue
		}
	}

	log.Info("Sync cloud routes done", "reconcileID", reconcileID,
		"prepareTime", preparedTime.Sub(startTime).Seconds(), "syncTime", time.Now().Sub(preparedTime).Seconds())

	return utilerrors.NewAggregate(errs)
}

func (r *ReconcileRoute) requeueNode(n *corev1.Node) {
	// if the channel is full, drop the event to prevent blocking
	select {
	case r.requeueChan <- event.GenericEvent{Object: n}:
		break
	default:
		log.Info("requeue channel is full, drop the request", "node", n.Name)
	}
}
