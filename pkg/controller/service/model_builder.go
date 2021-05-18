package service

import (
	"encoding/json"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
)

type IModelBuilder interface {
	Build(reqCtx *RequestContext) (*model.LoadBalancer, error)
}

type ModelType string

const (
	// LOCAL_MODEL, model built based on cluster information
	LOCAL_MODEL = ModelType("local")

	// REMOTE_MODEL, Model built based on cloud information
	REMOTE_MODEL = ModelType("remote")
)

type ModelBuilder struct {
	slbMgr    *LoadBalancerManager
	lisMgr    *ListenerManager
	vGroupMgr *VGroupManager
}

// NewDefaultModelBuilder construct a new defaultModelBuilder
func NewModelBuilder(slbMgr *LoadBalancerManager, lisMgr *ListenerManager, vGroupMgr *VGroupManager) *ModelBuilder {
	return &ModelBuilder{
		slbMgr:    slbMgr,
		lisMgr:    lisMgr,
		vGroupMgr: vGroupMgr,
	}
}

func (builder *ModelBuilder) Instance(modelType ModelType) IModelBuilder {
	switch modelType {
	case LOCAL_MODEL:
		return &localModel{builder}
	case REMOTE_MODEL:
		return &remoteModel{builder}
	}
	return &localModel{builder}
}

func (builder *ModelBuilder) BuildModel(reqCtx *RequestContext, modelType ModelType) (*model.LoadBalancer, error) {
	return builder.Instance(modelType).Build(reqCtx)
}

// localModel build model according to the Kubernetes cluster info
type localModel struct{ *ModelBuilder }

func (c localModel) Build(reqCtx *RequestContext) (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
	}
	// if the service do not need loadbalancer any more, return directly.
	if !isSLBNeeded(reqCtx.Service) {
		if reqCtx.Anno.Get(LoadBalancerId) != "" {
			lbMdl.LoadBalancerAttribute.IsUserManaged = true
		}
		return lbMdl, nil
	}
	if err := c.slbMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build slb attribute error: %s", err.Error())
	}
	if err := c.vGroupMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build slb listener error: %s", err.Error())
	}
	if err := c.lisMgr.BuildLocalModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("builid vserver groups error: %s", err.Error())
	}

	lbMdlJson, err := json.Marshal(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	klog.Infof("cluster build: %s", lbMdlJson)
	return lbMdl, nil
}

// localModel build model according to the cloud loadbalancer info
type remoteModel struct{ *ModelBuilder }

func (c remoteModel) Build(reqCtx *RequestContext) (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
	}

	err := c.slbMgr.BuildRemoteModel(reqCtx, lbMdl)
	if err != nil {
		return nil, fmt.Errorf("can not get load balancer attribute from cloud, error: %s", err.Error())
	}
	if lbMdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return lbMdl, nil
	}

	if err := c.vGroupMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("build backend from remote error: %s", err.Error())
	}

	if err := c.lisMgr.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		return nil, fmt.Errorf("can not build listener attribute from cloud, error: %s", err.Error())
	}

	lbMdlJson, err := json.Marshal(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	klog.Infof("cloud model build: %s", lbMdlJson)
	return lbMdl, nil
}
