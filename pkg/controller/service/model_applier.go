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
	_, err = m.updateListener(local, remote)
	if err != nil {
		return local, fmt.Errorf("update lb listeners error: %s", err.Error())
	}
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
	return local, nil

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
	klog.Infof("try to update listener. local: [%v], remote: [%v]", local.Listeners, remote.Listeners)
	// make https come first.
	// ensure https listeners to be created first for http forward
	//sort.SliceStable(
	//	updates,
	//	func(i, j int) bool {
	//		// 1. https comes first.
	//		// 2. DELETE action comes before https
	//		if isDeleteAction(updates[i].Action) {
	//			return true
	//		}
	//		if isDeleteAction(updates[j].Action) {
	//			return false
	//		}
	//		if strings.ToUpper(
	//			updates[i].TransforedProto,
	//		) == "HTTPS" {
	//			return true
	//		}
	//		return false
	//	},
	//)
	var (
		addition []*model.ListenerAttribute
		updation []*model.ListenerAttribute
		deletion []*model.ListenerAttribute
	)

	for _, rlis := range remote.Listeners {
		found := false
		for _, llis := range local.Listeners {
			if rlis.ListenerPort == llis.ListenerPort {
				found = true
				// port matched. that is where the conflict case begin.
				// 1. check protocol match.
				//if isProtocolMatch(local, remote) {
				//	// protocol match, need to do update operate no matter managed by whom
				//	// consider override annotation & user defined loadbalancer
				//	if !override && isUserDefinedLoadBalancer(svc) {
				//		// port conflict with user managed slb or listener.
				//		return nil, fmt.Errorf("PortProtocolConflict] port matched, but conflict with user managed listener. "+
				//			"Port:%d, ListenerName:%s, svc: %s. Protocol:[source:%s dst:%s]",
				//			remote.Port, remote.Name, local.NamedKey.Key(), remote.TransforedProto, local.TransforedProto)
				//	}

				// do update operate
				updation = append(updation, &llis)
				klog.Infof("found listener with port & protocol match, do update")
			} else {
				//// protocol not match, need to recreate
				//if !override && isUserDefinedLoadBalancer(svc) {
				//	return nil, fmt.Errorf("[PortProtocolConflict] port matched, "+
				//		"while protocol does not. force override listener %t. source[%v], target[%v]", override, local.NamedKey, remote.NamedKey)
				//}
				klog.Infof("found listener with protocol match, need recreate")
			}
		}
		if !found {
			deletion = append(deletion, &rlis)
			klog.Infof("not found listener, do delete")
		} else {
			klog.Infof("port [%d] not managed by my service [%s/%s], skip processing.", rlis.ListenerPort)
		}
	}

	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// For additions
	for _, llis := range local.Listeners {
		found := false
		klog.Infof("local port: %d", llis.ListenerPort)
		for _, rlis := range remote.Listeners {
			if llis.ListenerPort == rlis.ListenerPort {
				// port match
				//if !isProtocolMatch(remote, local) {
				//	// protocol does not match, do add listener
				//	break
				//}
				//// port matched. updated . skip
				found = true
				break
			}
		}
		if !found {
			addition = append(addition, &llis)
			klog.Infof("add new listener %d", llis.ListenerPort)
		}
	}

	for _, l := range addition {
		// TODO
		l.LoadBalancerId = remote.LoadBalancerAttribute.LoadBalancerId
		if err := m.cloud.CreateLoadBalancerTCPListener(m.ctx, l); err != nil {
			return nil, fmt.Errorf("create tcp listener error: %s", err.Error())
		}
		if err := m.cloud.StartLoadBalancerListener(m.ctx, l.LoadBalancerId, l.ListenerPort); err != nil {
			return nil, fmt.Errorf("start tcp listener error: %s", err.Error())
		}

	}

	//

	return nil, nil
}
