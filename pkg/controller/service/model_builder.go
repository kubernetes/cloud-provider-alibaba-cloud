package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
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
	ctx  context.Context
	svc  *v1.Service
	anno *AnnotationRequest

	cloud      prvd.Provider
	kubeClient client.Client

	model string
}

// NewDefaultModelBuilder construct a new defaultModelBuilder
func NewModelBuilder(ctx context.Context, kubeClient client.Client, cloud prvd.Provider,
	svc *v1.Service, anno *AnnotationRequest, model string) *ModelBuilder {
	return &ModelBuilder{
		ctx:        ctx,
		kubeClient: kubeClient,
		cloud:      cloud,
		svc:        svc,
		anno:       anno,
		model:      model,
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
		NamespacedName: util.NamespacedName(c.svc),
	}

	if err := c.buildLoadBalancerAttribute(lbMdl); err != nil {
		return nil, fmt.Errorf("build slb attribute error: %s", err.Error())
	}

	if err := c.buildVServerGroups(lbMdl); err != nil {
		return nil, fmt.Errorf("builid vserver groups error: %s", err.Error())
	}

	if err := c.buildListener(lbMdl); err != nil {
		return nil, fmt.Errorf("build slb listener error: %s", err.Error())
	}

	lbMdlJson, err := json.Marshal(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	klog.Infof("cluster build: %s", lbMdlJson)
	return lbMdl, nil
}

func (c localModel) buildLoadBalancerAttribute(m *model.LoadBalancer) error {
	m.LoadBalancerAttribute.AddressType = c.anno.Get(AddressType)
	m.LoadBalancerAttribute.InternetChargeType = c.anno.Get(ChargeType)
	bandwidth := c.anno.Get(Bandwidth)
	if bandwidth != nil {
		i, err := strconv.Atoi(*bandwidth)
		if err != nil &&
			*(m.LoadBalancerAttribute.InternetChargeType) == string(model.PayByBandwidth) {
			return fmt.Errorf("bandwidth must be integer, got [%s], error: %s", *bandwidth, err.Error())
		}
		m.LoadBalancerAttribute.Bandwidth = &i
	}
	if c.anno.Get(LoadBalancerId) != nil {
		m.LoadBalancerAttribute.LoadBalancerId = *c.anno.Get(LoadBalancerId)
		m.LoadBalancerAttribute.IsUserManaged = true
	}
	m.LoadBalancerAttribute.LoadBalancerName = c.anno.Get(LoadBalancerName)
	m.LoadBalancerAttribute.VSwitchId = c.anno.Get(VswitchId)
	m.LoadBalancerAttribute.MasterZoneId = c.anno.Get(MasterZoneID)
	m.LoadBalancerAttribute.SlaveZoneId = c.anno.Get(SlaveZoneID)
	m.LoadBalancerAttribute.LoadBalancerSpec = c.anno.Get(Spec)
	m.LoadBalancerAttribute.ResourceGroupId = c.anno.Get(ResourceGroupId)
	m.LoadBalancerAttribute.AddressIPVersion = c.anno.Get(IPVersion)
	m.LoadBalancerAttribute.DeleteProtection = c.anno.Get(DeleteProtection)
	m.LoadBalancerAttribute.ModificationProtectionStatus = c.anno.Get(ModificationProtection)

	return nil
}

func (c localModel) buildVServerGroups(m *model.LoadBalancer) error {
	vg, err := buildVGroupsFromService(c)
	if err != nil {
		return err
	}
	m.VServerGroups = vg
	return nil
}

func (c localModel) buildListener(mdl *model.LoadBalancer) error {
	for _, port := range c.svc.Spec.Ports {
		listener, err := buildListenersFromService(c, port)
		if err != nil {
			return fmt.Errorf("build listener from service error: %s", err.Error())
		}
		mdl.Listeners = append(mdl.Listeners, listener)
	}
	return nil
}

// localModel build model according to the cloud loadbalancer info
type remoteModel struct{ *ModelBuilder }

func (c remoteModel) Build() (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(c.svc),
	}

	exist, err := c.buildLoadBalancerAttribute(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("can not get load balancer attribute from cloud, error: %s", err.Error())
	}
	if !exist {
		return lbMdl, nil
	}

	if err := c.buildVServerGroups(lbMdl); err != nil {
		return nil, fmt.Errorf("build backend from remote error: %s", err.Error())
	}

	if err := c.buildListener(lbMdl); err != nil {
		return nil, fmt.Errorf("can not build listener attribute from cloud, error: %s", err.Error())
	}

	lbMdlJson, err := json.Marshal(lbMdl)
	if err != nil {
		return nil, fmt.Errorf("marshal lbmdl error: %s", err.Error())
	}
	klog.Infof("cloud model build: %s", lbMdlJson)
	return lbMdl, nil
}

func (c remoteModel) buildLoadBalancerAttribute(lb *model.LoadBalancer) (bool, error) {
	// Initialize the slb model to find the slb associated with the svc
	// 1. set loadbalancer id
	if c.anno.Get(LoadBalancerId) != nil {
		lb.LoadBalancerAttribute.LoadBalancerId = *c.anno.Get(LoadBalancerId)
	}
	// 2. set default loadbalancer name
	v := c.anno.GetDefaultLoadBalancerName()
	lb.LoadBalancerAttribute.LoadBalancerName = &v
	// 3. set default loadbalancer tag
	lb.LoadBalancerAttribute.Tags = []slb.Tag{
		{
			TagKey:   TAGKEY,
			TagValue: v,
		},
	}
	return c.cloud.FindSLB(c.ctx, lb)
}

func (c remoteModel) buildVServerGroups(remote *model.LoadBalancer) error {
	vgs, err := c.cloud.DescribeVServerGroups(c.ctx, remote.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeVServerGroups error: %s", err.Error())
	}
	remote.VServerGroups = vgs
	return nil
}

func (c remoteModel) buildListener(remote *model.LoadBalancer) error {
	listeners, err := c.cloud.DescribeLoadBalancerListeners(c.ctx, remote.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeLoadBalancerListeners error:%s", err.Error())
	}
	remote.Listeners = listeners
	return nil
}
