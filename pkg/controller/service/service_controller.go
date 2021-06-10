package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
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
func newReconciler(mgr manager.Manager, ctx *shared.SharedContext) reconcile.Reconciler {
	recon := &ReconcileService{
		cloud:            ctx.Provider(),
		kubeClient:       mgr.GetClient(),
		scheme:           mgr.GetScheme(),
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

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	if ctx2.GlobalFlag.DryRun {
		if err := initMap(mgr.GetClient()); err != nil {
		}
	}

	// Create a new controller
	c, err := controller.New(
		"service-controller", mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: 1,
		},
	)
	if err != nil {
		return err
	}
	hand := NewMapHandler()
	if err := c.Watch(&source.Kind{Type: &v1.Service{}}, hand,
		NewPredicateForServiceEvent(mgr.GetEventRecorderFor("service-controller"))); err != nil {
		return fmt.Errorf("watch resource svc error: %s", err.Error())
	}

	if err := c.Watch(&source.Kind{Type: &v1.Endpoints{}}, hand,
		NewPredicateForEndpointEvent(mgr.GetClient())); err != nil {
		return fmt.Errorf("watch resource endpoint error: %s", err.Error())
	}

	if err := c.Watch(&source.Kind{Type: &v1.Node{}}, hand,
		NewPredicateForNodeEvent(mgr.GetEventRecorderFor("service-controller"))); err != nil {
		return fmt.Errorf("watch resource node error: %s", err.Error())
	}
	return nil
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

	//record event recorder
	record           record.EventRecorder
	finalizerManager helper.FinalizerManager
}

func (m *ReconcileService) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	klog.Infof("do reconcile service %s", request.NamespacedName)
	return reconcile.Result{}, m.reconcile(request)
}

type RequestContext struct {
	Ctx      context.Context
	Service  *v1.Service
	Anno     *AnnotationRequest
	Log      *util.Log
	Recorder record.EventRecorder
}

func (m *ReconcileService) reconcile(request reconcile.Request) error {

	defer func() {
		if ctx2.GlobalFlag.DryRun {
			initial.Store(request, 1)
			if mapfull() {
				klog.Infof("ccm initial process finished.")
				err := dryrun.ResultEvent(m.kubeClient, dryrun.SUCCESS, "ccm initial process finished")
				if err != nil {
					klog.Errorf("write precheck event fail: %s", err.Error())
				}
				os.Exit(0)
			}
		}
	}()

	svc := &v1.Service{}
	err := m.kubeClient.Get(context.Background(), request.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("service %s not found, skip", request.NamespacedName)
			// Request object not found, could have been deleted
			// after reconcile request.
			// Owned objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil
		}
		return err
	}
	anno := &AnnotationRequest{svc: svc}

	// disable public address
	if anno.Get(AddressType) == "" ||
		anno.Get(AddressType) == string(model.InternetAddressType) {
		if ctx2.CFG.Global.DisablePublicSLB {
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
		Log:      util.NewReqLog(fmt.Sprintf("[%s] ", request.String())),
		Recorder: m.record,
	}

	reqContext.Log.Infof("ensure loadbalancer with service details, \n%+v", util.PrettyJson(svc))
	// check to see whither if loadbalancer deletion is needed
	if needDeleteLoadBalancer(svc) {
		return m.cleanupLoadBalancerResources(reqContext)
	}
	err = m.reconcileLoadBalancerResources(reqContext)
	if err != nil {
		klog.Errorf("[%s]: reconcile loadbalancer error: %s", request.NamespacedName, err.Error())
	}
	return err
}

func (m *ReconcileService) cleanupLoadBalancerResources(reqCtx *RequestContext) error {
	reqCtx.Log.Infof("service do not need lb any more, try to delete it")
	if helper.HasFinalizer(reqCtx.Service, ServiceFinalizer) {
		_, err := m.buildAndApplyModel(reqCtx)
		if err != nil && !strings.Contains(err.Error(), "LoadBalancerId does not exist") {
			m.record.Eventf(reqCtx.Service, v1.EventTypeWarning, helper.FailedCleanLB,
				fmt.Sprintf("Error deleting load balancer: %s", helper.GetLogMessage(err)))
			return err
		}

		if err := m.removeServiceHash(reqCtx.Service); err != nil {
			m.record.Eventf(reqCtx.Service, v1.EventTypeWarning, helper.FailedRemoveHash,
				fmt.Sprintf("Error removing service hash: %s", err.Error()))
			return err
		}

		if err := m.finalizerManager.RemoveFinalizers(reqCtx.Ctx, reqCtx.Service, ServiceFinalizer); err != nil {
			m.record.Eventf(reqCtx.Service, v1.EventTypeWarning, helper.FailedRemoveFinalizer,
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
		m.record.Eventf(req.Service, v1.EventTypeWarning, helper.FailedAddHash,
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
	localModel, err := m.builder.BuildModel(reqCtx, LOCAL_MODEL)
	if err != nil {
		return nil, fmt.Errorf("build lb local model error: %s", err.Error())
	}
	mdlJson, err := json.Marshal(localModel)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	klog.V(5).Infof("local build: %s", mdlJson)

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
		klog.Infof("status: [%v] [%v]", preStatus, newStatus)
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
				klog.Infof("%s, LoadBalancer: %s", util.Key(updated), updated.Status.LoadBalancer)
				err = m.kubeClient.Status().Patch(reqCtx.Ctx, updated, client.MergeFrom(svcOld))
				if err == nil {
					return nil
				}

				// If the object no longer exists, we don't want to recreate it. Just bail
				// out so that we can process the delete, which we should soon be receiving
				// if we haven't already.
				if errors.IsNotFound(err) {
					klog.Warningf("not persisting update to service that no "+
						"longer exists: %v", err)
					return nil
				}
				// TODO: Try to resolve the conflict if the change was unrelated to load
				// balancer status. For now, just pass it up the stack.
				if errors.IsConflict(err) {
					return fmt.Errorf("not persisting update to service %s that "+
						"has been changed since we received it: %v", util.Key(svc), err)
				}
				klog.Warningf("failed to persist updated LoadBalancerStatus to "+
					"service %s after creating its load balancer: %v", util.Key(svc), err)
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
		klog.Warningf(" slb %s has multiple eips, len [%d]", lbId, len(eips))
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
