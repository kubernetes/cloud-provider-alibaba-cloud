package servicemanager

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/backend"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
)

type Builder interface {
	Build(ctx context.Context, svcStackCtx *alb.ServiceStackContext) (*alb.ServiceManager, error)
}

var _ Builder = &defaultServiceManagerBuilder{}

type defaultServiceManagerBuilder struct {
	backendMgr *backend.Manager
}

func NewDefaultServiceStackBuilder(backendMgr *backend.Manager) *defaultServiceManagerBuilder {
	return &defaultServiceManagerBuilder{
		backendMgr: backendMgr,
	}
}

func (b defaultServiceManagerBuilder) Build(ctx context.Context, svcStackCtx *alb.ServiceStackContext) (*alb.ServiceManager, error) {
	serverStack := new(alb.ServiceManager)
	serverStack.ClusterID = svcStackCtx.ClusterID
	serverStack.Namespace = svcStackCtx.ServiceNamespace
	serverStack.Name = svcStackCtx.ServiceName
	port2ServerGroup := make(map[int32]*alb.ServerGroupWithIngress)
	port2Backends := make(map[int32][]alb.BackendItem)
	containsPotentialReadyEndpoints := false
	if !svcStackCtx.IsServiceNotFound {
		policy, err := helper.GetServiceTrafficPolicy(svcStackCtx.Service)
		if err != nil {
			return nil, err
		}
		serverStack.TrafficPolicy = string(policy)
		port2Backends, containsPotentialReadyEndpoints, err = b.backendMgr.BuildServicePortsToSDKBackends(ctx, svcStackCtx.Service)
		if err != nil {
			return nil, fmt.Errorf("build servicePortToServerGroup error: %v", err)
		}
		serverStack.ContainsPotentialReadyEndpoints = containsPotentialReadyEndpoints
	}
	for port, ingressNames := range svcStackCtx.ServicePortToIngressNames {
		port2ServerGroup[port] = new(alb.ServerGroupWithIngress)
		port2ServerGroup[port].IngressNames = ingressNames

		if backends, ok := port2Backends[port]; ok {
			port2ServerGroup[port].Backends = backends
		}
	}
	serverStack.PortToServerGroup = port2ServerGroup

	return serverStack, nil
}
