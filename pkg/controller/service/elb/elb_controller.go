package elb

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"

	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
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

	"github.com/go-logr/logr"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	return add(mgr, newReconciler(mgr, ctx))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) *ReconcileELB {
	recon := &ReconcileELB{
		cloud:            ctx.Provider(),
		kubeClient:       mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		logger:           ctrl.Log.WithName("controller").WithName("elb-controller"),
		record:           mgr.GetEventRecorderFor("elb-controller"),
		finalizerManager: helper.NewDefaultFinalizerManager(mgr.GetClient()),
	}

	elbManager := NewELBManager(recon.cloud)
	eipManager := NewEIPManager(recon.cloud)
	listenerManager := NewListenerManager(recon.cloud)
	serverGroupManager := NewServerGroupManager(recon.kubeClient, recon.cloud)

	recon.builder = NewModelBuilder(elbManager, eipManager, listenerManager, serverGroupManager)
	recon.applier = NewModelApplier(elbManager, eipManager, listenerManager, serverGroupManager)
	return recon
}

type elbController struct {
	c     controller.Controller
	recon *ReconcileELB
}

func (elbCtl elbController) Start(ctx context.Context) error {
	return elbCtl.c.Start(ctx)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileELB) error {
	rateLimit := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)

	recoverPanic := true
	// Create a new controller
	c, err := controller.NewUnmanaged(
		"elb-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 2,
			RateLimiter:             rateLimit,
			RecoverPanic:            &recoverPanic,
		},
	)
	if err != nil {
		return err
	}

	if err := c.Watch(&source.Kind{Type: &v1.Service{}},
		NewEnqueueRequestForServiceEvent(mgr.GetEventRecorderFor("elb-controller"))); err != nil {
		return fmt.Errorf("watch resource svc error: %s", err.Error())
	}

	if utilfeature.DefaultFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		// watch endpointslice
		if err := c.Watch(&source.Kind{Type: &discovery.EndpointSlice{}},
			NewEnqueueRequestForEndpointSliceEvent(mgr.GetEventRecorderFor("elb-controller"))); err != nil {
			return fmt.Errorf("watch resource endpointslice error: %s", err.Error())
		}
	} else {
		// watch endpoints
		if err := c.Watch(&source.Kind{Type: &v1.Endpoints{}},
			NewEnqueueRequestForEndpointEvent(mgr.GetEventRecorderFor("elb-controller"))); err != nil {
			return fmt.Errorf("watch resource endpoint error: %s", err.Error())
		}
	}

	if err := c.Watch(&source.Kind{Type: &v1.Node{}},
		NewEnqueueRequestForNodeEvent(mgr.GetEventRecorderFor("elb-controller"))); err != nil {
		return fmt.Errorf("watch resource node error: %s", err.Error())
	}
	return mgr.Add(&elbController{c: c, recon: r})
}

// ReconcileService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileELB{}

// ReconcileService reconciles a AutoRepair object
type ReconcileELB struct {
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

func (r ReconcileELB) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, r.reconcile(request)
}

func (r ReconcileELB) reconcile(request reconcile.Request) error {
	r.logger.Info("reconcile: start reconcile service", "service", request.NamespacedName)
	startTime := time.Now()
	svc := &v1.Service{}
	err := r.kubeClient.Get(context.Background(), request.NamespacedName, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.logger.Info("service not found, skip", "service", request.NamespacedName)
			return nil
		}
		r.logger.Error(err, "reconcile: get service failed", "service", request.NamespacedName)
	}

	anno := annotation.NewAnnotationRequest(svc)
	ctx := context.Background()
	ctx = context.WithValue(ctx, dryrun.ContextService, svc)
	reqCtx := &svcCtx.RequestContext{
		Ctx:      ctx,
		Service:  svc,
		Anno:     anno,
		Log:      r.logger.WithValues("service", util.Key(svc)),
		Recorder: r.record,
	}

	klog.Infof("%s: ensure loadbalancer with service details, \n%+v", util.Key(svc), util.PrettyJson(svc))

	if needDeleteLoadBalancer(svc) {
		err = r.cleanupLoadBalancerResources(reqCtx)
	} else {
		err = r.reconcileLoadBalancerResources(reqCtx)
	}
	if err != nil {
		return err
	}

	reqCtx.Log.Info("successfully reconcile")
	metric.SLBLatency.WithLabelValues("reconcile").Observe(metric.MsSince(startTime))
	return nil
}

func needDeleteLoadBalancer(svc *v1.Service) bool {
	return svc.DeletionTimestamp != nil || svc.Spec.Type != v1.ServiceTypeLoadBalancer
}

func (r *ReconcileELB) cleanupLoadBalancerResources(req *svcCtx.RequestContext) error {
	req.Log.Info("service do not need lb any more, try to delete it")
	if helper.HasFinalizer(req.Service, ELBFinalizer) {
		lb, err := r.buildAndApplyModel(req)
		if err != nil && !strings.Contains(err.Error(), "not exist") {
			r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedCleanLB,
				fmt.Sprintf("Error deleting load balancer [%s]: %s",
					lb.GetLoadBalancerId(), helper.GetLogMessage(err)))
			return err
		}
		if err = r.removeServiceLabels(req.Service); err != nil {
			r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedRemoveHash,
				fmt.Sprintf("Error removing service hash: %s", err.Error()))
			return err
		}

		// When service type changes from LoadBalancer to NodePort,
		// we need to clean Ingress attribute in service status
		if err = r.removeServiceStatus(req, req.Service); err != nil {
			r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedUpdateStatus,
				fmt.Sprintf("Error removing load balancer status: %s", err.Error()))
			return err
		}

		if err = r.finalizerManager.RemoveFinalizers(req.Ctx, req.Service, ELBFinalizer); err != nil {
			r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedRemoveFinalizer,
				fmt.Sprintf("Error removing load balancer finalizer: %v", err.Error()))
			return err
		}
	}
	return nil
}

func (r *ReconcileELB) reconcileLoadBalancerResources(req *svcCtx.RequestContext) error {
	// 1.add finalizer of elb
	if err := r.finalizerManager.AddFinalizers(req.Ctx, req.Service, ELBFinalizer); err != nil {
		r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedAddFinalizer,
			fmt.Sprintf("Error adding finalizer: %s", err.Error()))
		return err
	}

	// 2. build and apply edge lb model
	lb, err := r.buildAndApplyModel(req)
	if err != nil {
		r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedSyncLB,
			fmt.Sprintf("Error syncing load balancer [%s]: %s",
				lb.GetLoadBalancerId(), helper.GetLogMessage(err)))
		return err
	}

	// 3. add labels for service
	if err := r.addServiceLabels(req.Service, lb.GetLoadBalancerId()); err != nil {
		r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedAddHash,
			fmt.Sprintf("Error adding service hash: %s", err.Error()))
		return err
	}

	// 4. update status for service
	if err := r.updateServiceStatus(req, req.Service, lb); err != nil {
		r.record.Event(req.Service, v1.EventTypeWarning, helper.FailedUpdateStatus,
			fmt.Sprintf("Error updating load balancer status: %s", err.Error()))
		return err
	}

	r.record.Event(req.Service, v1.EventTypeNormal, helper.SucceedSyncLB,
		fmt.Sprintf("Ensured load balancer [%s]", lb.LoadBalancerAttribute.LoadBalancerId))
	return nil
}

func (r *ReconcileELB) buildAndApplyModel(reqCtx *svcCtx.RequestContext) (*elbmodel.EdgeLoadBalancer, error) {
	// build local model
	localModels, err := r.builder.BuildModel(reqCtx, LocalModel)
	if err != nil {
		return nil, fmt.Errorf("build load balancer local model error: %s", err.Error())
	}
	mdlJson, err := json.Marshal(localModels)
	if err != nil {
		return nil, fmt.Errorf("marshal load balancer model error: %s", err.Error())
	}

	r.logger.V(5).Info(fmt.Sprintf("local build: %s", mdlJson))

	// apply model
	remoteModels, err := r.applier.Apply(reqCtx, localModels)
	if err != nil {
		return remoteModels, fmt.Errorf("apply model error: %s", err.Error())
	}
	return remoteModels, nil

}

func (r *ReconcileELB) addServiceLabels(svc *v1.Service, lbId string) error {
	updated := svc.DeepCopy()
	if updated.Labels == nil {
		updated.Labels = make(map[string]string)
	}
	serviceHash := helper.GetServiceHash(svc)
	updated.Labels[helper.LabelServiceHash] = serviceHash
	if lbId != "" {
		updated.Labels[helper.LabelLoadBalancerId] = lbId
	}
	if err := r.kubeClient.Status().Patch(context.Background(), updated, client.MergeFrom(svc)); err != nil {
		return fmt.Errorf("%s failed to add service hash:, error: %s", util.Key(svc), err.Error())
	}
	return nil
}

func (r *ReconcileELB) removeServiceLabels(svc *v1.Service) error {
	updated := svc.DeepCopy()
	needUpdated := false
	if _, ok := updated.Labels[helper.LabelServiceHash]; ok {
		delete(updated.Labels, helper.LabelServiceHash)
		needUpdated = true
	}

	delete(updated.Labels, annotation.LoadBalancerId)

	if needUpdated {
		err := r.kubeClient.Status().Patch(context.TODO(), updated, client.MergeFrom(svc))
		if err != nil {
			return fmt.Errorf("%s failed to remove service hash:, error: %s", util.Key(svc), err.Error())
		}
	}
	return nil
}

func (r *ReconcileELB) updateServiceStatus(reqCtx *svcCtx.RequestContext, svc *v1.Service, lb *elbmodel.EdgeLoadBalancer) error {
	preStatus := svc.Status.LoadBalancer.DeepCopy()
	newStatus := &v1.LoadBalancerStatus{}
	if lb == nil {
		return fmt.Errorf("lb not found, cannot not patch service status")
	}
	var serviceIp string
	if reqCtx.Anno.Get(annotation.EdgeEipAssociate) != "" && !isTrue(reqCtx.Anno.Get(annotation.EdgeEipAssociate)) {
		serviceIp = lb.LoadBalancerAttribute.Address
	} else {
		serviceIp = lb.GetEipAddress()
	}
	ingress, err := r.setExternalIP(reqCtx.Ctx, lb.GetLoadBalancerName(), serviceIp)
	if err != nil {
		reqCtx.Recorder.Event(svc, v1.EventTypeWarning, "FailedSetEIPAddress", "get eip error, set external ip to slb ip")
	}
	newStatus.Ingress = ingress

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
				retErr = r.kubeClient.Get(reqCtx.Ctx, util.NamespacedName(svc), svcOld)
				if retErr != nil {
					return fmt.Errorf("error to get svc %s", util.Key(svc))
				}
				updated := svcOld.DeepCopy()
				updated.Status.LoadBalancer = *newStatus
				reqCtx.Log.Info(fmt.Sprintf("LoadBalancer: %v", updated.Status.LoadBalancer))
				retErr = r.kubeClient.Status().Patch(reqCtx.Ctx, updated, client.MergeFrom(svcOld))
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

func (r *ReconcileELB) setExternalIP(ctx context.Context, lbName, ip string) ([]v1.LoadBalancerIngress, error) {
	mdl := new(elbmodel.EdgeLoadBalancer)
	if ip == "" {
		return nil, fmt.Errorf("elb %s has no eip, svc external ip cannot be set to loadbalance address", mdl.GetEipAddress())
	}
	if lbName == "" {
		return nil, fmt.Errorf("elb %s has no name, svc host name cannot be set to loadbalance name", lbName)
	}

	var ingress []v1.LoadBalancerIngress
	ingress = append(ingress,
		v1.LoadBalancerIngress{
			IP: ip,
		})

	return ingress, nil
}

func (r *ReconcileELB) removeServiceStatus(reqCtx *svcCtx.RequestContext, svc *v1.Service) error {
	preStatus := svc.Status.LoadBalancer.DeepCopy()
	newStatus := &v1.LoadBalancerStatus{}
	if !v1helper.LoadBalancerStatusEqual(preStatus, newStatus) {
		util.ServiceLog.Info(fmt.Sprintf("status: [%v] [%v]", preStatus, newStatus))
	}
	return helper.Retry(
		&wait.Backoff{Duration: 1 * time.Second, Steps: 3, Factor: 2, Jitter: 4},
		func(svc *v1.Service) error {
			oldSvc := &v1.Service{}
			err := r.kubeClient.Get(reqCtx.Ctx, util.NamespacedName(svc), oldSvc)
			if err != nil {
				return fmt.Errorf("get svc %s, err", util.Key(svc))
			}
			updated := oldSvc.DeepCopy()
			updated.Status.LoadBalancer = *newStatus
			reqCtx.Log.Info(fmt.Sprintf("LoadBalancer: %v", updated.Status.LoadBalancer))
			err = r.kubeClient.Status().Patch(reqCtx.Ctx, updated, client.MergeFrom(oldSvc))
			if err != nil {
				if apierrors.IsNotFound(err) {
					util.ServiceLog.Info("not persisting update to service that no longer exists")
					return nil
				}
				if apierrors.IsConflict(err) {
					return fmt.Errorf("not persisting update to service %s that "+
						"has been changed since we received it: %v", util.Key(svc), err)
				}
				reqCtx.Log.Error(err, "failed to persist updated LoadBalancerStatus"+
					" after creating its load balancer")
				return fmt.Errorf("retry with %s, %s", err.Error(), helper.TRY_AGAIN)
			}
			return nil
		},
		svc)
}
