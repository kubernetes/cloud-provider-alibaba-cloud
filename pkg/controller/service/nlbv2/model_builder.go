package nlbv2

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

type ModelType string

const (
	// LOCAL_MODEL, model built based on cluster information
	LocalModel = ModelType("local")

	// REMOTE_MODEL, Model built based on cloud information
	RemoteModel = ModelType("remote")
)

type IModelBuilder interface {
	Build(reqCtx *svcCtx.RequestContext) (*nlbmodel.NetworkLoadBalancer, error)
}

type ModelBuilder struct {
	NLBMgr *NLBManager
	LisMgr *ListenerManager
	SGMgr  *ServerGroupManager
}

// NewDefaultModelBuilder construct a new defaultModelBuilder
func NewModelBuilder(nlbMgr *NLBManager, lisMgr *ListenerManager, sgMgr *ServerGroupManager) *ModelBuilder {
	return &ModelBuilder{
		NLBMgr: nlbMgr,
		LisMgr: lisMgr,
		SGMgr:  sgMgr,
	}
}

func (builder *ModelBuilder) Instance(modelType ModelType) IModelBuilder {
	switch modelType {
	case LocalModel:
		return &localModel{builder}
	case RemoteModel:
		return &remoteModel{builder}
	}
	return &localModel{builder}
}

func (builder *ModelBuilder) BuildModel(reqCtx *svcCtx.RequestContext, modelType ModelType) (*nlbmodel.NetworkLoadBalancer, error) {
	return builder.Instance(modelType).Build(reqCtx)
}

// localModel build model according to the Kubernetes cluster info
type localModel struct{ *ModelBuilder }

func (c localModel) Build(reqCtx *svcCtx.RequestContext) (*nlbmodel.NetworkLoadBalancer, error) {
	lbMdl := &nlbmodel.NetworkLoadBalancer{
		NamespacedName:        util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
	}
	// if the service do not need loadbalancer any more, return directly.
	if helper.NeedDeleteLoadBalancer(reqCtx.Service) {
		if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
			lbMdl.LoadBalancerAttribute.IsUserManaged = true
		}
		if reqCtx.Anno.Get(annotation.PreserveLBOnDelete) != "" {
			lbMdl.LoadBalancerAttribute.PreserveOnDelete = true
		}
		return lbMdl, nil
	}
	if err := c.NLBMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build nlb attribute error: %s", err.Error())
	}
	if err := c.LisMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build nlb server group error: %s", err.Error())
	}
	if err := c.SGMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("builid nlb listener error: %s", err.Error())
	}

	return lbMdl, nil
}

// localModel build model according to the cloud loadbalancer info
type remoteModel struct{ *ModelBuilder }

func (c remoteModel) Build(reqCtx *svcCtx.RequestContext) (*nlbmodel.NetworkLoadBalancer, error) {
	lbMdl := &nlbmodel.NetworkLoadBalancer{
		NamespacedName:        util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
	}

	err := c.NLBMgr.BuildRemoteModel(reqCtx, lbMdl)
	if err != nil {
		return nil, fmt.Errorf("can not get nlb attribute from cloud, error: %s", err.Error())
	}
	if lbMdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return lbMdl, nil
	}

	if err := c.SGMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build server group from remote error: %s", err.Error())
	}

	if err := c.LisMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("can not build nlb listener attribute from cloud, error: %s", err.Error())
	}

	return lbMdl, nil
}
