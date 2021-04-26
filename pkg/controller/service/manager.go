package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type EndpointMgr struct {
	isEniBackendType bool

	anno      *AnnotationRequest
	localMode bool
	nodes     []v1.Node
	svc       *v1.Service
	endpoints *v1.Endpoints
}

func NewEndpointMgr(
	mc client.Client, svc *v1.Service,
) (*EndpointMgr, error) {
	mgr := &EndpointMgr{
		svc:       svc,
		anno:      &AnnotationRequest{svc: svc},
		localMode: isLocalModeService(svc),
		// backend type
		isEniBackendType: IsENIBackendType(svc),
	}
	nodes := v1.NodeList{}
	err := mc.List(context.TODO(), &nodes)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("svc: %s", key(svc)))
	}
	ns, err := filterOutByLabel(nodes.Items, *mgr.anno.Get(BackendLabel))
	if err != nil {
		return nil, errors.Wrap(err, BackendLabel)
	}
	mgr.nodes = ns

	eps := v1.Endpoints{}
	err = mc.Get(context.TODO(), client.ObjectKey{Namespace: svc.Namespace, Name: svc.Name}, &eps)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return nil, errors.Wrap(err, fmt.Sprintf("svc: %s", key(svc)))
		}
		klog.Warningf("endpoint not found: %s", key(svc))
	}
	mgr.endpoints = &eps
	return mgr, nil
}

type ContextManager struct {
	svc  *v1.Service
	anno *AnnotationRequest

	cloud       prvd.Provider
	kubeClient  client.Client
	endpointMgr *EndpointMgr

	//record event recorder
	record record.EventRecorder
}

func NewContextMgr(
	svc *v1.Service,
	cloud prvd.Provider,
	mclient client.Client,
	record record.EventRecorder,
) *ContextManager {
	return &ContextManager{
		svc:        svc,
		kubeClient: mclient,
		cloud:      cloud,
		record:     record,
		anno:       &AnnotationRequest{svc: svc},
	}
}

func (m *ContextManager) Reconcile() error {
	err := m.validate()
	if err != nil {
		return errors.Wrap(err, "validate")
	}

	// 1. check to see whither if loadbalancer deletion is needed
	if !isSLBNeeded(m.svc) {
		return m.cleanupLoadBalancerResources(m.svc)
	}
	err = m.reconcileLoadBalancerResources(m.svc)
	if err != nil {
		log.Infof("reconcile loadbalancer error: %s", err.Error())
	}
	return err
}

func (m *ContextManager) validate() error {
	// safety check.
	if m.svc == nil {
		return fmt.Errorf("service could not be empty")
	}

	// disable public address
	if m.anno.Get(AddressType) == nil || *m.anno.Get(AddressType) == "internet" {
		if ctx2.CFG.Global.DisablePublicSLB {
			return fmt.Errorf("PublicAddress SLB is Not allowed")
		}
	}
	return nil
}

func (m *ContextManager) cleanupLoadBalancerResources(svc *v1.Service) error {
	// TODO
	//if k8s.HasFinalizer(svc, serviceFinalizer) {
	//	_, _, err := r.buildAndDeployModel(ctx, svc)
	//	if err != nil {
	//		return err
	//	}
	//	if err := r.finalizerManager.RemoveFinalizers(ctx, svc, serviceFinalizer); err != nil {
	//		r.eventRecorder.Event(svc, corev1.EventTypeWarning, k8s.ServiceEventReasonFailedRemoveFinalizer, fmt.Sprintf("Failed remove finalizer due to %v", err))
	//		return err
	//	}
	//}
	return nil
}

func (m *ContextManager) reconcileLoadBalancerResources(svc *v1.Service) error {
	// TODO Finalizer
	//if err := m.finalizerManager.AddFinalizers(ctx, svc, serviceFinalizer); err != nil {
	//	r.eventRecorder.Event(svc, corev1.EventTypeWarning, k8s.ServiceEventReasonFailedAddFinalizer, fmt.Sprintf("Failed add finalizer due to %v", err))
	//	return err
	//}

	if err := m.buildAndEnsureModel(); err != nil {
		return err
	}
	klog.Infof("build & ensure successfully")

	if err := m.updateServiceStatus(svc); err != nil {
		//m.record.Event(svc, v1.EventTypeWarning, , fmt.Sprintf("Failed update status due to %v", err))
		return err
	}

	//m.record.Event(svc, v1.EventTypeNormal, k8s.ServiceEventReasonSuccessfullyReconciled, "Successfully reconciled")
	return nil
}

func (m *ContextManager) buildAndEnsureModel() error {
	clusterModel, err := NewClusterModelBuilder(m.kubeClient, m.cloud, m.svc, m.anno).Build()
	if err != nil {
		return fmt.Errorf("build cluster model error: %s", err.Error())
	}
	cloudModel, err := NewCloudModelBuilder(m.cloud, m.svc, m.anno).Build()
	if err != nil {
		return fmt.Errorf("build cloud model error: %s", err.Error())
	}

	updated, err := m.updateLoadBalancerAttribute(clusterModel, cloudModel)
	if err != nil {
		return err
	}
	updated, err = m.updateBackend(clusterModel, updated)
	if err != nil {
		return err
	}

	updated, err = m.updateListener(clusterModel, updated)
	if err != nil {
		return err
	}

	log.Infof("updated model: [%v]", updated)
	return nil
}

func (m *ContextManager) updateLoadBalancerAttribute(clusterMdl *model.LoadBalancer, cloudMdl *model.LoadBalancer) (*model.LoadBalancer, error) {
	ctx := context.TODO()

	if cloudMdl.LoadBalancerAttribute.LoadBalancerId == "" {
		klog.Info("not found load balancer, try to create")
		return createLoadBalancer(ctx, m.cloud, clusterMdl, m.anno)
	}

	klog.Infof("found load balancer, try to update load balancer attribute")

	//if cluster.LoadBalancerAttribute.RefLoadBalancerId != nil &&
	//	*cluster.LoadBalancerAttribute.RefLoadBalancerId != cloud.LoadBalancerAttribute.LoadBalancerId {
	//	return fmt.Errorf("can not found slb according to user defined loadbalancer id [%s]",
	//		*cluster.LoadBalancerAttribute.RefLoadBalancerId)
	//}
	//
	//if cluster.LoadBalancerAttribute.LoadBalancerName != nil &&
	//	*cluster.LoadBalancerAttribute.LoadBalancerName != *cloud.LoadBalancerAttribute.LoadBalancerName {
	//	m.cloud.CreateSLB()
	//
	//}
	return &model.LoadBalancer{}, nil

}

func (m *ContextManager) updateBackend(balancer *model.LoadBalancer, balancer2 *model.LoadBalancer) (*model.LoadBalancer, error) {
	return nil, nil

}

func (m *ContextManager) updateListener(balancer *model.LoadBalancer, balancer2 *model.LoadBalancer) (*model.LoadBalancer, error) {
	return nil, nil
}

func (m *ContextManager) updateServiceStatus(svc *v1.Service) error {
	// TODO
	//if len(svc.Status.LoadBalancer.Ingress) != 1 ||
	//	svc.Status.LoadBalancer.Ingress[0].IP != "" {
	//	svcOld := svc.DeepCopy()
	//	svc.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
	//		{
	//			Hostname: lbDNS,
	//		},
	//	}
	//	if err := m.kubeClient.Status().Patch(context.TODO(), svc, client.MergeFrom(svcOld)); err != nil {
	//		return errors.Wrapf(err, "failed to update service status: %v", NamespacedName(svc))
	//	}
	//}
	return nil

}

func createLoadBalancer(
	ctx context.Context,
	cloud prvd.Provider,
	lbMdl *model.LoadBalancer,
	anno *AnnotationRequest) (*model.LoadBalancer, error) {
	if lbMdl.LoadBalancerAttribute.LoadBalancerSpec == nil {
		*lbMdl.LoadBalancerAttribute.LoadBalancerSpec = anno.GetDefaultValue(Spec)
	}
	if lbMdl.LoadBalancerAttribute.AddressType == nil {
		v := anno.GetDefaultValue(AddressType)
		lbMdl.LoadBalancerAttribute.AddressType = &v

	}
	return cloud.CreateSLB(ctx, lbMdl)

}
