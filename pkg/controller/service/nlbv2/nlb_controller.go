package nlbv2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/metric"
	"k8s.io/klog/v2"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	reconciler, err := newReconciler(mgr, ctx)
	if err != nil {
		return fmt.Errorf("new nlb reconciler error: %s", err.Error())
	}
	return add(mgr, reconciler)
}

func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) (*ReconcileNLB, error) {
	recon := &ReconcileNLB{
		cloud:            ctx.Provider(),
		kubeClient:       mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		logger:           ctrl.Log.WithName("controller").WithName("nlb-controller"),
		record:           mgr.GetEventRecorderFor("nlb-controller"),
		finalizerManager: helper.NewDefaultFinalizerManager(mgr.GetClient()),
	}

	nlbManager := NewNLBManager(recon.cloud)
	listenerManager := NewListenerManager(recon.cloud)
	serverGroupManager, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	if err != nil {
		return nil, fmt.Errorf("NewServerGroupManager error:%s", err.Error())
	}
	recon.builder = NewModelBuilder(nlbManager, listenerManager, serverGroupManager)
	recon.applier = NewModelApplier(nlbManager, listenerManager, serverGroupManager)
	return recon, nil
}

type nlbController struct {
	c     controller.Controller
	recon *ReconcileNLB
}

func (n nlbController) Start(ctx context.Context) error {
	return n.c.Start(ctx)
}

func add(mgr manager.Manager, r *ReconcileNLB) error {
	rateLimit := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
	recoverPanic := true
	// Create a new controller
	c, err := controller.NewUnmanaged(
		"nlb-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: ctrlCfg.CloudCFG.Global.ServiceMaxConcurrentReconciles,
			RateLimiter:             rateLimit,
			RecoverPanic:            &recoverPanic,
		},
	)
	if err != nil {
		return err
	}

	if err := c.Watch(source.Kind(mgr.GetCache(), &v1.Service{}),
		NewEnqueueRequestForServiceEvent(mgr.GetEventRecorderFor("nlb-controller"))); err != nil {
		return fmt.Errorf("watch resource svc error: %s", err.Error())
	}

	if utilfeature.DefaultFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		// watch endpointslice
		if err := c.Watch(source.Kind(mgr.GetCache(), &discovery.EndpointSlice{}),
			NewEnqueueRequestForEndpointSliceEvent(mgr.GetClient(), mgr.GetEventRecorderFor("service-controller"))); err != nil {
			return fmt.Errorf("watch resource endpointslice error: %s", err.Error())
		}
	} else {
		// watch endpoints
		if err := c.Watch(source.Kind(mgr.GetCache(), &v1.Endpoints{}),
			NewEnqueueRequestForEndpointEvent(mgr.GetClient(), mgr.GetEventRecorderFor("service-controller"))); err != nil {
			return fmt.Errorf("watch resource endpoint error: %s", err.Error())
		}
	}

	if err := c.Watch(source.Kind(mgr.GetCache(), &v1.Node{}),
		NewEnqueueRequestForNodeEvent(mgr.GetClient(), mgr.GetEventRecorderFor("nlb-controller"))); err != nil {
		return fmt.Errorf("watch resource node error: %s", err.Error())
	}

	return mgr.Add(&nlbController{c: c, recon: r})
}

var _ reconcile.Reconciler = &ReconcileNLB{}

type ReconcileNLB struct {
	scheme  *runtime.Scheme
	builder *ModelBuilder
	applier *ModelApplier

	// client
	cloud      prvd.Provider
	kubeClient client.Client

	logger logr.Logger

	//record event recorder
	record           record.EventRecorder
	finalizerManager helper.FinalizerManager
}

func (m *ReconcileNLB) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return util.HandleReconcileResult(request, m.reconcile(ctx, request))
}

func (m *ReconcileNLB) reconcile(c context.Context, request reconcile.Request) error {
	startTime := time.Now()

	reconcileID := controller.ReconcileIDFromContext(c)
	nlbLog := util.NLBLog.WithValues("service", request.NamespacedName.String(), "reconcileID", reconcileID)
	nlbLog.Info("starting reconcile service")

	svc := &v1.Service{}
	err := m.kubeClient.Get(context.Background(), request.NamespacedName, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			m.logger.Info("service not found, skip", "service", request.NamespacedName)
			return nil
		}
		m.logger.Error(err, "reconcile: get service failed", "service", request.NamespacedName)
	}

	anno := &annotation.AnnotationRequest{Service: svc}
	// new context for each request
	ctx := context.Background()
	ctx = context.WithValue(ctx, dryrun.ContextService, svc)
	reqCtx := &svcCtx.RequestContext{
		Ctx:         ctx,
		ReconcileID: string(reconcileID),
		Service:     svc,
		Anno:        anno,
		Log:         nlbLog,
		Recorder:    m.record,
	}

	klog.Infof("%s: ensure loadbalancer with service details, reconcileID: %s\n%+v\n", util.Key(svc), reconcileID, util.PrettyJson(svc))

	if helper.NeedDeleteLoadBalancer(svc) || !helper.NeedNLB(svc) {
		err = m.cleanupLoadBalancerResources(reqCtx)
	} else {
		err = m.reconcileLoadBalancerResources(reqCtx)
	}

	var needRequeue *util.ReconcileNeedRequeue

	if err != nil && !errors.As(err, &needRequeue) {
		return err
	}

	reqCtx.Log.Info("successfully reconcile", "elapsedTime", time.Since(startTime).Seconds())
	metric.SLBLatency.WithLabelValues(metric.NLBType, "reconcile").Observe(metric.MsSince(startTime))

	if needRequeue != nil {
		reqCtx.Log.Info("requeue needed", "reason", needRequeue.Error(), "service", util.Key(reqCtx.Service))
		return needRequeue
	}

	return nil
}

func (m *ReconcileNLB) cleanupLoadBalancerResources(reqCtx *svcCtx.RequestContext) error {
	reqCtx.Log.Info("service do not need lb any more, try to delete it")
	if helper.HasFinalizer(reqCtx.Service, helper.NLBFinalizer) {
		lb, _, err := m.buildAndApplyModel(reqCtx)
		if err != nil && !strings.Contains(err.Error(), "ResourceNotFound.loadBalancer") {
			m.record.Event(reqCtx.Service, v1.EventTypeWarning, helper.FailedCleanLB,
				fmt.Sprintf("Error deleting load balancer [%s]: %s",
					lb.GetLoadBalancerId(), helper.GetLogMessage(err)))
			return err
		}

		if err := m.removeServiceLabels(reqCtx.Service); err != nil {
			m.record.Event(reqCtx.Service, v1.EventTypeWarning, helper.FailedRemoveHash,
				fmt.Sprintf("Error removing service hash: %s", err.Error()))
			return err
		}

		// When service type changes from LoadBalancer to NodePort,
		// we need to clean Ingress attribute in service status
		if err := m.removeServiceStatus(reqCtx, reqCtx.Service); err != nil {
			m.record.Event(reqCtx.Service, v1.EventTypeWarning, helper.FailedUpdateStatus,
				fmt.Sprintf("Error removing load balancer status: %s", err.Error()))
			return err
		}

		if err := m.finalizerManager.RemoveFinalizers(reqCtx.Ctx, reqCtx.Service, helper.NLBFinalizer); err != nil {
			m.record.Event(reqCtx.Service, v1.EventTypeWarning, helper.FailedRemoveFinalizer,
				fmt.Sprintf("Error removing load balancer finalizer: %v", err.Error()))
			return err
		}
	}
	m.record.Event(reqCtx.Service, v1.EventTypeNormal, helper.SucceedCleanLB, "Clean load balancer")
	return nil
}

func (m *ReconcileNLB) reconcileLoadBalancerResources(req *svcCtx.RequestContext) error {

	if err := m.finalizerManager.AddFinalizers(req.Ctx, req.Service, helper.NLBFinalizer); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedAddFinalizer,
			fmt.Sprintf("Error adding finalizer: %s", err.Error()))
		return err
	}

	lb, sgs, err := m.buildAndApplyModel(req)
	if err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedSyncLB,
			fmt.Sprintf("Error syncing load balancer [%s]: %s",
				lb.GetLoadBalancerId(), helper.GetLogMessage(err)))
		return err
	}

	if err := m.updateReadinessCondition(req, sgs); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedUpdateReadinessGate,
			fmt.Sprintf("Error updating pod readiness gates for service [%s]: %s", util.Key(req.Service), err.Error()))
		return err
	}

	if err := m.addServiceLabels(req.Service, lb.GetLoadBalancerId()); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedAddHash,
			fmt.Sprintf("Error adding service hash: %s", err.Error()))
		return err
	}

	if err := m.updateServiceStatus(req, req.Service, lb); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedUpdateStatus,
			fmt.Sprintf("Error updating load balancer status: %s", err.Error()))
		return err
	}

	m.record.Event(req.Service, v1.EventTypeNormal, helper.SucceedSyncLB,
		fmt.Sprintf("Ensured load balancer [%s]", lb.LoadBalancerAttribute.LoadBalancerId))
	return nil
}

func (m *ReconcileNLB) buildAndApplyModel(reqCtx *svcCtx.RequestContext) (*nlbmodel.NetworkLoadBalancer, []*nlbmodel.ServerGroup, error) {

	// build local model
	localModel, err := m.builder.BuildModel(reqCtx, LocalModel)
	if err != nil {
		return nil, nil, fmt.Errorf("build lb local model error: %s", err.Error())
	}
	mdlJson, err := json.Marshal(localModel)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	reqCtx.Log.V(5).Info(fmt.Sprintf("local build: %s", mdlJson))

	// apply model
	remoteModel, err := m.applier.Apply(reqCtx, localModel)
	if err != nil {
		return remoteModel, nil, fmt.Errorf("apply model error: %s", err.Error())
	}
	return remoteModel, localModel.ServerGroups, nil
}

func (m *ReconcileNLB) updateServiceStatus(reqCtx *svcCtx.RequestContext, svc *v1.Service, lb *nlbmodel.NetworkLoadBalancer) error {
	preStatus := svc.Status.LoadBalancer.DeepCopy()
	newStatus := &v1.LoadBalancerStatus{}
	if lb == nil {
		return fmt.Errorf("lb not found, cannot not patch service status")
	}

	// SLB ExternalIPType, display the slb ip as service external ip
	// If the length of elastic ip is 0, display the slb ip
	if len(newStatus.Ingress) == 0 {
		newStatus.Ingress = append(newStatus.Ingress,
			v1.LoadBalancerIngress{
				Hostname: lb.LoadBalancerAttribute.DNSName,
			})
	}

	// Write the state if changed
	// TODO: Be careful here ... what if there were other changes to the service?
	if !v1helper.LoadBalancerStatusEqual(preStatus, newStatus) {
		util.ServiceLog.Info(fmt.Sprintf("status: [%v] [%v]", preStatus, newStatus))
		var retErr error
		_ = helper.Retry(
			&wait.Backoff{
				Duration: 1 * time.Second,
				Steps:    3,
				Factor:   2,
				Jitter:   4,
			},
			func(svc *v1.Service) error {
				// get latest svc from the shared informer cache
				svcOld := &v1.Service{}
				retErr = m.kubeClient.Get(reqCtx.Ctx, util.NamespacedName(svc), svcOld)
				if retErr != nil {
					return fmt.Errorf("error to get svc %s", util.Key(svc))
				}
				updated := svcOld.DeepCopy()
				updated.Status.LoadBalancer = *newStatus
				reqCtx.Log.Info(fmt.Sprintf("LoadBalancer: %v", updated.Status.LoadBalancer))
				retErr = m.kubeClient.Status().Patch(reqCtx.Ctx, updated, client.MergeFrom(svcOld))
				if retErr == nil {
					return nil
				}

				// If the object no longer exists, we don't want to recreate it. Just bail
				// out so that we can process the delete, which we should soon be receiving
				// if we haven't already.
				if apierrors.IsNotFound(retErr) {
					util.ServiceLog.Error(retErr, "not persisting update to service that no longer exists")
					retErr = nil
					return nil
				}
				// TODO: Try to resolve the conflict if the change was unrelated to load
				// balancer status. For now, just pass it up the stack.
				if apierrors.IsConflict(retErr) {
					return fmt.Errorf("not persisting update to service %s that "+
						"has been changed since we received it: %v", util.Key(svc), retErr)
				}
				reqCtx.Log.Error(retErr, "failed to persist updated LoadBalancerStatus"+
					" after creating its load balancer")
				return fmt.Errorf("retry with %s, %s", retErr.Error(), helper.TRY_AGAIN)
			},
			svc,
		)
		return retErr
	}
	return nil

}

func (m *ReconcileNLB) removeServiceStatus(reqCtx *svcCtx.RequestContext, svc *v1.Service) error {
	preStatus := svc.Status.LoadBalancer.DeepCopy()
	newStatus := &v1.LoadBalancerStatus{}

	// Write the state if changed
	// TODO: Be careful here ... what if there were other changes to the service?
	if !v1helper.LoadBalancerStatusEqual(preStatus, newStatus) {
		util.ServiceLog.Info(fmt.Sprintf("status: [%v] [%v]", preStatus, newStatus))
		return helper.Retry(
			&wait.Backoff{
				Duration: 1 * time.Second,
				Steps:    3,
				Factor:   2,
				Jitter:   4,
			},
			func(svc *v1.Service) error {
				// get latest svc from the shared informer cache
				svcOld := &v1.Service{}
				err := m.kubeClient.Get(reqCtx.Ctx, util.NamespacedName(svc), svcOld)
				if err != nil {
					return fmt.Errorf("error to get svc %s", util.Key(svc))
				}
				updated := svcOld.DeepCopy()
				updated.Status.LoadBalancer = *newStatus
				reqCtx.Log.Info(fmt.Sprintf("LoadBalancer: %v", updated.Status.LoadBalancer))
				err = m.kubeClient.Status().Patch(reqCtx.Ctx, updated, client.MergeFrom(svcOld))
				if err == nil {
					return nil
				}

				// If the object no longer exists, we don't want to recreate it. Just bail
				// out so that we can process the delete, which we should soon be receiving
				// if we haven't already.
				if apierrors.IsNotFound(err) {
					util.ServiceLog.Error(err, "not persisting update to service that no longer exists")
					return nil
				}
				// TODO: Try to resolve the conflict if the change was unrelated to load
				// balancer status. For now, just pass it up the stack.
				if apierrors.IsConflict(err) {
					return fmt.Errorf("not persisting update to service %s that "+
						"has been changed since we received it: %v", util.Key(svc), err)
				}
				reqCtx.Log.Error(err, "failed to persist updated LoadBalancerStatus"+
					" after creating its load balancer")
				return fmt.Errorf("retry with %s, %s", err.Error(), helper.TRY_AGAIN)
			},
			svc,
		)
	}
	return nil

}

func (m *ReconcileNLB) addServiceLabels(svc *v1.Service, lbId string) error {
	updated := svc.DeepCopy()
	if updated.Labels == nil {
		updated.Labels = make(map[string]string)
	}
	serviceHash := helper.GetServiceHash(svc)
	updated.Labels[helper.LabelServiceHash] = serviceHash
	if lbId != "" {
		updated.Labels[helper.LabelLoadBalancerId] = lbId
	}
	if err := m.kubeClient.Status().Patch(context.Background(), updated, client.MergeFrom(svc)); err != nil {
		return fmt.Errorf("%s failed to add service hash:, error: %s", util.Key(svc), err.Error())
	}
	return nil
}

func (m *ReconcileNLB) removeServiceLabels(svc *v1.Service) error {
	updated := svc.DeepCopy()
	needUpdate := false
	if _, ok := updated.Labels[helper.LabelServiceHash]; ok {
		delete(updated.Labels, helper.LabelServiceHash)
		needUpdate = true
	}
	if _, ok := updated.Labels[helper.LabelLoadBalancerId]; ok {
		delete(updated.Labels, helper.LabelLoadBalancerId)
		needUpdate = true
	}
	if needUpdate {
		if err := m.kubeClient.Status().Patch(context.Background(), updated, client.MergeFrom(svc)); err != nil {
			return fmt.Errorf("%s failed to remove service hash:, error: %s", util.Key(svc), err.Error())
		}
	}
	return nil
}

func (m *ReconcileNLB) updateReadinessCondition(reqCtx *svcCtx.RequestContext, sgs []*nlbmodel.ServerGroup) error {
	var errs []error
	cond := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, reqCtx.Service.Name)
	a := map[string]bool{}
	for _, sg := range sgs {
		for _, b := range sg.InitialServers {
			if b.TargetRef == nil {
				reqCtx.Log.Info("backend TargetRef is nil, skip update readiness gates")
				continue
			}
			key := types.NamespacedName{Namespace: b.TargetRef.Namespace, Name: b.TargetRef.Name}
			if _, ok := a[key.String()]; ok {
				continue
			}

			pod := &v1.Pod{}
			err := m.kubeClient.Get(reqCtx.Ctx, key, pod)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			err = helper.UpdateReadinessConditionForPod(reqCtx.Ctx, m.kubeClient, pod, cond,
				helper.ConditionReasonServerRegistered, helper.ConditionMessageServerRegistered)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			a[key.String()] = true
		}
	}
	return utilerrors.NewAggregate(errs)
}
