package service

import (
	"encoding/json"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
)

type IModelBuilder interface {
	Build() (*model.LoadBalancer, error)
}

const (
	// LOCAL_MODEL, model built based on cluster information
	LOCAL_MODEL = "local"

	// REMOTE_MODEL, Model built based on cloud information
	REMOTE_MODEL = "remote"
)

type ModelBuilder struct {
	reqCtx *RequestContext
	model  string
}

// NewDefaultModelBuilder construct a new defaultModelBuilder
func NewModelBuilder(req *RequestContext, model string) *ModelBuilder {
	return &ModelBuilder{
		reqCtx: req,
		model:  model,
	}
}

func (builder *ModelBuilder) Instance() IModelBuilder {
	switch builder.model {
	case LOCAL_MODEL:
		return &localModel{builder}
	case REMOTE_MODEL:
		return &remoteModel{builder}
	}
	return &localModel{builder}
}

func (builder *ModelBuilder) Build() (*model.LoadBalancer, error) {
	return builder.Instance().Build()
}

// localModel build model according to the Kubernetes cluster info
type localModel struct{ *ModelBuilder }

func (c localModel) Build() (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(c.reqCtx.svc),
	}
	// if the service do not need loadbalancer any more, return directly.
	if !isSLBNeeded(c.reqCtx.svc) {
		if c.reqCtx.anno.Get(LoadBalancerId) != "" {
			lbMdl.LoadBalancerAttribute.IsUserManaged = true
		}
		return lbMdl, nil
	}
	if err := c.reqCtx.BuildLoadBalancerAttributeForLocalModel(lbMdl); err != nil {
		return nil, fmt.Errorf("build slb attribute error: %s", err.Error())
	}
	if err := c.reqCtx.BuildVGroupsForLocalModel(lbMdl); err != nil {
		return nil, fmt.Errorf("builid vserver groups error: %s", err.Error())
	}
	if err := c.reqCtx.BuildListenersForLocalModel(lbMdl); err != nil {
		return nil, fmt.Errorf("build slb listener error: %s", err.Error())
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

func (c remoteModel) Build() (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(c.reqCtx.svc),
	}

	err := c.reqCtx.BuildLoadBalancerAttributeForRemoteModel(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("can not get load balancer attribute from cloud, error: %s", err.Error())
	}
	if lbMdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return lbMdl, nil
	}

	if err := c.reqCtx.BuildVGroupsForRemoteModel(lbMdl); err != nil {
		return nil, fmt.Errorf("build backend from remote error: %s", err.Error())
	}

	if err := c.reqCtx.BuildListenersForRemoteModel(lbMdl); err != nil {
		return nil, fmt.Errorf("can not build listener attribute from cloud, error: %s", err.Error())
	}

	lbMdlJson, err := json.Marshal(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	klog.Infof("cloud model build: %s", lbMdlJson)
	return lbMdl, nil
}
