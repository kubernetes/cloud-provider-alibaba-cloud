package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	discovery "k8s.io/api/discovery/v1beta1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/metric"
	"k8s.io/klog"
	"os"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	return add(mgr, newReconciler(mgr, ctx))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) *ReconcileService {
	recon := &ReconcileService{
		cloud:            ctx.Provider(),
		kubeClient:       mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		logger:           ctrl.Log.WithName("controller").WithName("service-controller"),
		record:           mgr.GetEventRecorderFor("service-controller"),
		finalizerManager: helper.NewDefaultFinalizerManager(mgr.GetClient()),
	}

	slbManager := NewLoadBalancerManager(recon.cloud)
	listenerManager := NewListenerManager(recon.cloud)
	vGroupManager := NewVGroupManager(recon.kubeClient, recon.cloud)
	recon.builder = NewModelBuilder(slbManager, listenerManager, vGroupManager)
	recon.applier = NewModelApplier(slbManager, listenerManager, vGroupManager)
	return recon
}

type serviceController struct {
	c     controller.Controller
	recon *ReconcileService
}

func (svcC serviceController) Start(ctx context.Context) error {
	if ctrlCfg.ControllerCFG.DryRun {
		initMap(svcC.recon.kubeClient)
	}
	return svcC.c.Start(ctx)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r *ReconcileService) error {
	rateLimit := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)

	// Create a new controller
	c, err := controller.NewUnmanaged(
		"service-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 2,
			RateLimiter:             rateLimit,
		},
	)
	if err != nil {
		return err
	}

	if err := c.Watch(&source.Kind{Type: &v1.Service{}},
		NewEnqueueRequestForServiceEvent(mgr.GetEventRecorderFor("service-controller"))); err != nil {
		return fmt.Errorf("watch resource svc error: %s", err.Error())
	}

	if utilfeature.DefaultFeatureGate.Enabled(helper.EndpointSlice) {
		// watch endpointslice
		if err := c.Watch(&source.Kind{Type: &discovery.EndpointSlice{}},
			NewEnqueueRequestForEndpointSliceEvent(mgr.GetEventRecorderFor("service-controller"))); err != nil {
			return fmt.Errorf("watch resource endpointslice error: %s", err.Error())
		}
	} else {
		// watch endpoints
		if err := c.Watch(&source.Kind{Type: &v1.Endpoints{}},
			NewEnqueueRequestForEndpointEvent(mgr.GetEventRecorderFor("service-controller"))); err != nil {
			return fmt.Errorf("watch resource endpoint error: %s", err.Error())
		}
	}

	if err := c.Watch(&source.Kind{Type: &v1.Node{}},
		NewEnqueueRequestForNodeEvent(mgr.GetEventRecorderFor("service-controller"))); err != nil {
		return fmt.Errorf("watch resource node error: %s", err.Error())
	}
	return mgr.Add(&serviceController{c: c, recon: r})
}

// ReconcileService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a AutoRepair object
type ReconcileService struct {
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

func (m *ReconcileService) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, m.reconcile(request)
}

type RequestContext struct {
	Ctx      context.Context
	Service  *v1.Service
	Anno     *AnnotationRequest
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (m *ReconcileService) reconcile(request reconcile.Request) (err error) {
	startTime := time.Now()

	defer func() {
		if ctrlCfg.ControllerCFG.DryRun {
			initial.Store(request.String(), 1)
			if mapfull() {
				util.ServiceLog.Info("ccm initial process finished.")
				err := dryrun.ResultEvent(m.kubeClient, dryrun.SUCCESS, "ccm initial process finished")
				if err != nil {
					util.ServiceLog.Error(err, "write precheck event failed")
				}
				os.Exit(0)
			}
			err = nil
		}
	}()

	svc := &v1.Service{}
	err = m.kubeClient.Get(context.Background(), request.NamespacedName, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			util.ServiceLog.Info("service not found, skip", "service", request.NamespacedName)
			// Request object not found, could have been deleted
			// after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil
		}
		util.ServiceLog.Error(err, "reconcile: get service failed", "service", request.NamespacedName)
		return err
	}
	anno := &AnnotationRequest{svc: svc}

	// disable public address
	if anno.Get(AddressType) == "" ||
		anno.Get(AddressType) == string(model.InternetAddressType) {
		if ctrlCfg.CloudCFG.Global.DisablePublicSLB {
			m.record.Event(svc, v1.EventTypeWarning, helper.FailedSyncLB, "create public address slb is not allowed")
			// do not support create public address slb, return and don't requeue
			return nil
		}
	}

	// new context for each request
	ctx := context.Background()
	ctx = context.WithValue(ctx, dryrun.ContextService, svc)

	reqContext := &RequestContext{
		Ctx:      ctx,
		Service:  svc,
		Anno:     anno,
		Log:      util.ServiceLog.WithValues("service", util.Key(svc)),
		Recorder: m.record,
	}

	klog.Infof("%s: ensure loadbalancer with service details, \n%+v", util.Key(svc), util.PrettyJson(svc))

	if ctrlCfg.ControllerCFG.DryRun {
		if _, err = m.buildAndApplyModel(reqContext); err != nil {
			m.record.Event(reqContext.Service, v1.EventTypeWarning, helper.FailedSyncLB,
				fmt.Sprintf("DryRun: Error syncing load balancer: %s", helper.GetLogMessage(err)))
		}
		return nil
	}

	// check to see whither if loadbalancer deletion is needed
	if needDeleteLoadBalancer(svc) {
		err = m.cleanupLoadBalancerResources(reqContext)
	} else {
		err = m.reconcileLoadBalancerResources(reqContext)
	}
	if err != nil {
		reqContext.Log.Error(err, "reconcile loadbalancer failed")
		return err
	}

	reqContext.Log.Info("successfully reconcile")
	metric.SLBLatency.WithLabelValues("reconcile").Observe(metric.MsSince(startTime))

	return nil
}

func (m *ReconcileService) cleanupLoadBalancerResources(reqCtx *RequestContext) error {
	reqCtx.Log.Info("service do not need lb any more, try to delete it")
	if helper.HasFinalizer(reqCtx.Service, ServiceFinalizer) {
		_, err := m.buildAndApplyModel(reqCtx)
		if err != nil && !strings.Contains(err.Error(), "LoadBalancerId does not exist") {
			m.record.Event(reqCtx.Service, v1.EventTypeWarning, helper.FailedCleanLB,
				fmt.Sprintf("Error deleting load balancer: %s", helper.GetLogMessage(err)))
			return err
		}

		if err := m.removeServiceHash(reqCtx.Service); err != nil {
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

		if err := m.finalizerManager.RemoveFinalizers(reqCtx.Ctx, reqCtx.Service, ServiceFinalizer); err != nil {
			m.record.Event(reqCtx.Service, v1.EventTypeWarning, helper.FailedRemoveFinalizer,
				fmt.Sprintf("Error removing load balancer finalizer: %v", err.Error()))
			return err
		}
	}
	m.record.Event(reqCtx.Service, v1.EventTypeNormal, helper.SucceedCleanLB, "Clean load balancer")
	return nil
}

func (m *ReconcileService) reconcileLoadBalancerResources(req *RequestContext) error {

	if err := m.finalizerManager.AddFinalizers(req.Ctx, req.Service, ServiceFinalizer); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedAddFinalizer,
			fmt.Sprintf("Error adding finalizer: %s", err.Error()))
		return err
	}

	lb, err := m.buildAndApplyModel(req)
	if err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedSyncLB,
			fmt.Sprintf("Error syncing load balancer: %s", helper.GetLogMessage(err)))
		return err
	}

	if err := m.addServiceHash(req.Service); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedAddHash,
			fmt.Sprintf("Error adding service hash: %s", err.Error()))
		return err
	}

	if err := m.updateServiceStatus(req, req.Service, lb); err != nil {
		m.record.Event(req.Service, v1.EventTypeWarning, helper.FailedUpdateStatus,
			fmt.Sprintf("Error updating load balancer status: %s", err.Error()))
		return err
	}

	m.record.Event(req.Service, v1.EventTypeNormal, helper.SucceedSyncLB, "Ensured load balancer")
	return nil
}

func (m *ReconcileService) buildAndApplyModel(reqCtx *RequestContext) (*model.LoadBalancer, error) {

	// build local model
	localModel, err := m.builder.BuildModel(reqCtx, LocalModel)
	if err != nil {
		return nil, fmt.Errorf("build lb local model error: %s", err.Error())
	}
	mdlJson, err := json.Marshal(localModel)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	util.ServiceLog.V(5).Info(fmt.Sprintf("local build: %s", mdlJson))

	// apply model
	remoteModel, err := m.applier.Apply(reqCtx, localModel)
	if err != nil {
		return nil, fmt.Errorf("apply model error: %s", err.Error())
	}
	return remoteModel, nil
}

func (m *ReconcileService) updateServiceStatus(reqCtx *RequestContext, svc *v1.Service, lb *model.LoadBalancer) error {
	preStatus := svc.Status.LoadBalancer.DeepCopy()
	newStatus := &v1.LoadBalancerStatus{}
	if lb == nil {
		return fmt.Errorf("lb not found, cannot not patch service status")
	}

	// EIP ExternalIPType, display the slb associated elastic ip as service external ip
	if reqCtx.Anno.Get(ExternalIPType) == "eip" {
		ingress, err := m.setEIPAsExternalIP(reqCtx.Ctx, lb.LoadBalancerAttribute.LoadBalancerId)
		if err != nil {
			reqCtx.Recorder.Event(svc, v1.EventTypeWarning, "FailedSetEIPAddress", "get eip error, set external ip to slb ip")
		}
		newStatus.Ingress = ingress
	}

	// SLB ExternalIPType, display the slb ip as service external ip
	// If the length of elastic ip is 0, display the slb ip
	if len(newStatus.Ingress) == 0 {
		newStatus.Ingress = append(newStatus.Ingress,
			v1.LoadBalancerIngress{
				IP: lb.LoadBalancerAttribute.Address,
			})
	}

	// Write the state if changed
	// TODO: Be careful here ... what if there were other changes to the service?
	if !v1helper.LoadBalancerStatusEqual(preStatus, newStatus) {
		util.ServiceLog.Info(fmt.Sprintf("status: [%v] [%v]", preStatus, newStatus))
		return retry(
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
				return fmt.Errorf("retry with %s, %s", err.Error(), TRY_AGAIN)
			},
			svc,
		)
	}
	return nil

}

func (m *ReconcileService) removeServiceStatus(reqCtx *RequestContext, svc *v1.Service) error {
	preStatus := svc.Status.LoadBalancer.DeepCopy()
	newStatus := &v1.LoadBalancerStatus{}

	// Write the state if changed
	// TODO: Be careful here ... what if there were other changes to the service?
	if !v1helper.LoadBalancerStatusEqual(preStatus, newStatus) {
		util.ServiceLog.Info(fmt.Sprintf("status: [%v] [%v]", preStatus, newStatus))
		return retry(
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
				return fmt.Errorf("retry with %s, %s", err.Error(), TRY_AGAIN)
			},
			svc,
		)
	}
	return nil

}

func (m *ReconcileService) addServiceHash(svc *v1.Service) error {
	updated := svc.DeepCopy()
	if updated.Labels == nil {
		updated.Labels = make(map[string]string)
	}
	serviceHash := getServiceHash(svc)
	updated.Labels[LabelServiceHash] = serviceHash
	if err := m.kubeClient.Status().Patch(context.Background(), updated, client.MergeFrom(svc)); err != nil {
		return fmt.Errorf("%s failed to add service hash:, error: %s", util.Key(svc), err.Error())
	}
	return nil
}

func (m *ReconcileService) removeServiceHash(svc *v1.Service) error {
	updated := svc.DeepCopy()
	if _, ok := updated.Labels[LabelServiceHash]; ok {
		delete(updated.Labels, LabelServiceHash)
		if err := m.kubeClient.Status().Patch(context.Background(), updated, client.MergeFrom(svc)); err != nil {
			return fmt.Errorf("%s failed to remove service hash:, error: %s", util.Key(svc), err.Error())
		}
	}
	return nil
}

func (m *ReconcileService) setEIPAsExternalIP(ctx context.Context, lbId string) ([]v1.LoadBalancerIngress, error) {

	eips, err := m.cloud.DescribeEipAddresses(ctx, string(vpc.SlbInstance), lbId)
	if err != nil {
		return nil, err
	}
	if len(eips) == 0 {
		return nil, fmt.Errorf("slb %s has no eip, svc external ip cannot be set to eip address", lbId)
	}
	if len(eips) > 1 {
		util.ServiceLog.Info(fmt.Sprintf(" slb %s has multiple eips, len [%d]", lbId, len(eips)))
	}

	var ingress []v1.LoadBalancerIngress
	for _, eip := range eips {
		ingress = append(ingress,
			v1.LoadBalancerIngress{
				IP: eip,
			})
	}
	return ingress, nil
}
