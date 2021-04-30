package service

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog"
)

func NewLBModelApplier(ctx context.Context, cloud prvd.Provider, svc *v1.Service, anno *AnnotationRequest) *modelApplier {
	return &modelApplier{
		ctx:   ctx,
		cloud: cloud,
		svc:   svc,
		anno:  anno,
	}
}

type modelApplier struct {
	ctx   context.Context
	cloud prvd.Provider
	svc   *v1.Service
	anno  *AnnotationRequest
}

func (m *modelApplier) Apply(local *model.LoadBalancer, remote *model.LoadBalancer) (*model.LoadBalancer, error) {

	local, err := m.updateLoadBalancerAttribute(local, remote)
	if err != nil {
		return local, fmt.Errorf("update lb attribute error: %s", err.Error())
	}
	//_, err = m.updateBackend(local, remote)
	//if err != nil {
	//	return local, fmt.Errorf("update lb backends error: %s", err.Error())
	//}
	//_, err = m.updateListener(local, remote)
	//if err != nil {
	//	return local, fmt.Errorf("update lb listeners error: %s", err.Error())
	//}
	return local, nil
}

func (m *modelApplier) updateLoadBalancerAttribute(local *model.LoadBalancer, remote *model.LoadBalancer) (*model.LoadBalancer, error) {

	if remote == nil || remote.LoadBalancerAttribute.LoadBalancerId == "" {
		setModelDefaultValue(local, m.anno)
		klog.Info("not found load balancer, try to create slb. model [%v]", local)
		return m.cloud.CreateSLB(m.ctx, local)
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

func setModelDefaultValue(mdl *model.LoadBalancer, anno *AnnotationRequest) {
	if mdl.LoadBalancerAttribute.AddressType == nil {
		v := anno.GetDefaultValue(AddressType)
		mdl.LoadBalancerAttribute.AddressType = &v
	}

	if mdl.LoadBalancerAttribute.LoadBalancerName == nil {
		v := anno.GetDefaultLoadBalancerName()
		mdl.LoadBalancerAttribute.LoadBalancerName = &v
	}

	// TODO ecs模式下获取vpc id & vsw id
	if *mdl.LoadBalancerAttribute.AddressType == string(model.IntranetAddressType) {
		mdl.LoadBalancerAttribute.VpcId = ctx2.CFG.Global.VpcID
		if mdl.LoadBalancerAttribute.VSwitchId == nil {
			v := ctx2.CFG.Global.VswitchID
			mdl.LoadBalancerAttribute.VSwitchId = &v
		}
	}

	if mdl.LoadBalancerAttribute.LoadBalancerSpec == nil {
		v := anno.GetDefaultValue(Spec)
		mdl.LoadBalancerAttribute.LoadBalancerSpec = &v
	}

	if mdl.LoadBalancerAttribute.DeleteProtection == nil {
		v := anno.GetDefaultValue(DeleteProtection)
		mdl.LoadBalancerAttribute.DeleteProtection = &v
	}

	if mdl.LoadBalancerAttribute.ModificationProtectionStatus == nil {
		v := anno.GetDefaultValue(ModificationProtection)
		mdl.LoadBalancerAttribute.ModificationProtectionStatus = &v
		mdl.LoadBalancerAttribute.ModificationProtectionReason = ModificationProtectionReason
	}

	mdl.LoadBalancerAttribute.Tags = append(mdl.LoadBalancerAttribute.Tags, anno.GetDefaultTags()...)
}

func (m *modelApplier) updateBackend(local *model.LoadBalancer, remote *model.LoadBalancer) (*model.LoadBalancer, error) {
	return nil, nil

}

func (m *modelApplier) updateListener(local *model.LoadBalancer, remote *model.LoadBalancer) (*model.LoadBalancer, error) {
	return nil, nil
}
