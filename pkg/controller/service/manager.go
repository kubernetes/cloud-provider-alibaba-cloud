package service

import (
	"context"
	"fmt"
	"github.com/denverdino/aliyungo/slb"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)
type EndpointMgr struct {
	isEniBackendType bool

	req       *AnnotationRequest
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
		req:       &AnnotationRequest{svc: svc},
		localMode: isLocalModeService(svc),
		// backend type
		isEniBackendType: IsENIBackendType(svc),
	}
	nodes := v1.NodeList{}
	err := mc.List(context.TODO(), &nodes)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("svc: %s", key(svc)))
	}
	ns, err := filterOutByLabel(nodes.Items, mgr.req.Get(BackendLabel))
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
	svc *v1.Service
	req *AnnotationRequest

	cloud       provider.Provider
	client      client.Client
	endpointMgr *EndpointMgr

	//record event recorder
	record record.EventRecorder
}

func NewContextMgr(
	svc *v1.Service,
	cloud provider.Provider,
	mclient client.Client,
	record record.EventRecorder,
) *ContextManager {
	return &ContextManager{
		svc:    svc,
		client: mclient,
		cloud:  cloud,
		record: record,
		req:    &AnnotationRequest{svc: svc},
	}
}

func (m *ContextManager) Reconcile() error {
	err := m.validate()
	if err != nil {
		return errors.Wrap(err, "validate")
	}

	// 1. check to see whither if loadbalancer deletion is needed
	if !isSLBNeeded(m.svc) {
		// todo: do slb delete
		return nil
	}

	m.endpointMgr, err = NewEndpointMgr(m.client, m.svc)
	if err != nil {
		return errors.Wrap(err, "ContextManager")
	}
	//m.cloud.ListSLB(context.TODO(),)

	return nil
}

func (m *ContextManager) validate() error {
	// safety check.
	if m.svc == nil {
		return fmt.Errorf("service could not be empty")
	}

	// disable public address
	if m.req.Get(AddressType) == string(slb.InternetAddressType) {
		if ctx2.CFG.Global.DisablePublicSLB {
			return fmt.Errorf("PublicAddress SLB is Not allowed")
		}
	}
	return nil
}

