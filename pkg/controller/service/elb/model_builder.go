package elb

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

type ModelType string

const (
	// LocalModel  is model built based on cluster information
	LocalModel = ModelType("local")

	// RemoteModel is model built based on cloud information
	RemoteModel = ModelType("remote")
)

type IModelBuilder interface {
	Build(reqCtx *svcCtx.RequestContext) (*elbmodel.EdgeLoadBalancer, error)
}

type ModelBuilder struct {
	ELBMgr *ELBManager
	EIPMgr *EIPManager
	LisMgr *ListenerManager
	SGMgr  *ServerGroupManager
}

func NewModelBuilder(elbMgr *ELBManager, eipMgr *EIPManager, lisMgr *ListenerManager, sgMgr *ServerGroupManager) *ModelBuilder {
	return &ModelBuilder{
		ELBMgr: elbMgr,
		EIPMgr: eipMgr,
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

func (builder *ModelBuilder) BuildModel(reqCtx *svcCtx.RequestContext, modelType ModelType) (*elbmodel.EdgeLoadBalancer, error) {
	return builder.Instance(modelType).Build(reqCtx)
}

type localModel struct{ *ModelBuilder }

func (l localModel) Build(reqCtx *svcCtx.RequestContext) (*elbmodel.EdgeLoadBalancer, error) {
	lbMdl := &elbmodel.EdgeLoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: elbmodel.EdgeLoadBalancerAttribute{
			IsUserManaged: true,
			IsReUsed:      false,
		},
		EipAttribute: elbmodel.EdgeEipAttribute{IsUserManaged: true},
		ServerGroup:  elbmodel.EdgeServerGroup{},
		Listeners:    elbmodel.EdgeListeners{},
	}
	if needDeleteLoadBalancer(reqCtx.Service) {
		if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
			lbMdl.LoadBalancerAttribute.IsUserManaged = true
		}
		return lbMdl, nil
	}

	if err := l.ELBMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build elb attribute error: %s", err.Error())
	}

	if err := l.EIPMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build eip attribute error: %s", err.Error())
	}

	if err := l.SGMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build server group error: %s", err.Error())
	}

	if err := l.LisMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build elb listener error: %s", err.Error())
	}

	return lbMdl, nil
}

type remoteModel struct{ *ModelBuilder }

func (r remoteModel) Build(reqCtx *svcCtx.RequestContext) (*elbmodel.EdgeLoadBalancer, error) {
	lbMdl := &elbmodel.EdgeLoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: elbmodel.EdgeLoadBalancerAttribute{
			IsUserManaged: true,
			IsReUsed:      false,
		},
		EipAttribute: elbmodel.EdgeEipAttribute{IsUserManaged: true},
		ServerGroup:  elbmodel.EdgeServerGroup{},
		Listeners:    elbmodel.EdgeListeners{},
	}

	if err := r.ELBMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build elb attribute error: %s", err.Error())
	}

	if err := r.EIPMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build eip attribute error: %s", err.Error())
	}

	if err := r.SGMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build server group error: %s", err.Error())
	}

	if err := r.LisMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build elb listener error: %s", err.Error())
	}

	return lbMdl, nil
}
