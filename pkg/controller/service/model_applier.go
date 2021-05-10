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

func (m *modelApplier) Apply(local *model.LoadBalancer, remote *model.LoadBalancer) error {

	if err := m.updateLoadBalancerAttribute(local, remote); err != nil {
		return fmt.Errorf("update lb attribute error: %s", err.Error())
	}

	if err := m.updateBackend(local, remote); err != nil {
		return fmt.Errorf("update lb backends error: %s", err.Error())
	}

	if err := m.updateListener(local, remote); err != nil {
		return fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	if err := m.cleanup(local, remote); err != nil {
		return fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	return nil
}

func (m *modelApplier) updateLoadBalancerAttribute(local *model.LoadBalancer, remote *model.LoadBalancer) error {
	var err error
	if local == nil {
		return fmt.Errorf("local model is nil")
	}
	if remote == nil {
		return fmt.Errorf("remote model is nil")
	}

	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		setModelDefaultValue(local, m.anno)
		klog.Info("not found load balancer, try to create slb")
		err = m.cloud.CreateSLB(m.ctx, local)
		if err != nil {
			return fmt.Errorf("create slb error: %s", err.Error())
		}
		remote.LoadBalancerAttribute.LoadBalancerId = local.LoadBalancerAttribute.LoadBalancerId
		err = m.cloud.DescribeSLB(m.ctx, remote)
		if err != nil {
			return fmt.Errorf("describe slb error: %s", err.Error())
		}
		return nil
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
	return nil

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
	// TODO remove me
	//if mdl.LoadBalancerAttribute.DeleteProtection == nil {
	//	v := anno.GetDefaultValue(DeleteProtection)
	//	mdl.LoadBalancerAttribute.DeleteProtection = &v
	//}
	//
	//if mdl.LoadBalancerAttribute.ModificationProtectionStatus == nil {
	//	v := anno.GetDefaultValue(ModificationProtection)
	//	mdl.LoadBalancerAttribute.ModificationProtectionStatus = &v
	//	mdl.LoadBalancerAttribute.ModificationProtectionReason = ModificationProtectionReason
	//}

	mdl.LoadBalancerAttribute.Tags = append(mdl.LoadBalancerAttribute.Tags, anno.GetDefaultTags()...)
}

func (m *modelApplier) updateBackend(local *model.LoadBalancer, remote *model.LoadBalancer) error {
	for i := range local.VServerGroups {
		found := false
		for j := range remote.VServerGroups {
			if local.VServerGroups[i].VGroupName == remote.VServerGroups[j].VGroupName {
				found = true

				add, del, update := diff(local.VServerGroups[i], remote.VServerGroups[j])
				if len(add) == 0 && len(del) == 0 && len(update) == 0 {
					klog.Infof("update: no backend need to be added for vgroupid")
				}
				//if len(add) > 0 {
				//	if err := batchAddVServerGroupBackendServers(ctx, add); err != nil {
				//		break
				//	}
				//}
				//if len(del) > 0 {
				//	if err := batchRemoveVServerGroupBackendServers(ctx, del); err != nil {
				//		break
				//	}
				//}
				//if len(update) > 0 {
				//	batchUpdateVServerGroupBackendServers(ctx, update)
				//}

				local.VServerGroups[i].VGroupId = remote.VServerGroups[j].VGroupId
			}
		}

		// if not exist, create
		if !found {
			if err := m.cloud.CreateVServerGroup(m.ctx,
				&local.VServerGroups[i], remote.LoadBalancerAttribute.LoadBalancerId); err != nil {
				return fmt.Errorf("create vserver group error: %s", err.Error())
			}
			remote.VServerGroups = append(remote.VServerGroups, local.VServerGroups[i])
		}
	}

	return nil
}

func diff(remote, local model.VServerGroup) (
	[]model.BackendAttribute, []model.BackendAttribute, []model.BackendAttribute) {

	var (
		addition  []model.BackendAttribute
		deletions []model.BackendAttribute
		updates   []model.BackendAttribute
	)

	for _, r := range remote.Backends {
		//// skip nodes which does not belong to the cluster
		//if isUserManagedNode(api.Description, v.NamedKey.Key()) {
		//	continue
		//}
		found := false
		for _, l := range local.Backends {
			if l.Type == "eni" {
				if r.ServerId == l.ServerId &&
					r.ServerIp == l.ServerIp {
					found = true
					break
				}
			} else {
				if r.ServerId == l.ServerId {
					found = true
					break
				}
			}

		}
		if !found {
			deletions = append(deletions, r)
		}
	}
	for _, l := range local.Backends {
		found := false
		for _, r := range remote.Backends {
			if l.Type == "eni" {
				if r.ServerId == l.ServerId &&
					r.ServerIp == l.ServerIp {
					found = true
					break
				}
			} else {
				if r.ServerId == l.ServerId {
					found = true
					break
				}
			}

		}
		if !found {
			addition = append(addition, l)
		}
	}
	for _, l := range local.Backends {
		for _, r := range remote.Backends {
			//if isUserManagedNode(api.Description, v.NamedKey.Key()) {
			//	continue
			//}
			if l.Type == "eni" {
				if l.ServerId == r.ServerId &&
					l.ServerIp == r.ServerIp &&
					l.Weight != r.Weight {
					updates = append(updates, l)
					break
				}
			} else {
				if l.ServerId == r.ServerId &&
					l.Weight != r.Weight {
					updates = append(updates, l)
					break
				}
			}
		}
	}
	return addition, deletions, updates
}

func (m *modelApplier) updateListener(local *model.LoadBalancer, remote *model.LoadBalancer) error {
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
		addition []model.ListenerAttribute
		updation []model.ListenerAttribute
		deletion []model.ListenerAttribute
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
				updation = append(updation, llis)
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
			deletion = append(deletion, rlis)
			klog.Infof("not found listener, do delete")
		} else {
			klog.Infof("port [%d] not managed by my service [%s/%s], skip processing.", rlis.ListenerPort)
		}
	}

	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// For additions
	for _, llis := range local.Listeners {
		found := false
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
			addition = append(addition, llis)
		}
	}

	for _, l := range addition {
		klog.Infof("add new listener %d", l.ListenerPort)
		lbId := remote.LoadBalancerAttribute.LoadBalancerId
		if err := findVServerGroup(local.VServerGroups, &l); err != nil {
			return err
		}
		if err := m.cloud.CreateLoadBalancerTCPListener(m.ctx, lbId, &l); err != nil {
			return fmt.Errorf("create tcp listener [%d] error: %s", l.ListenerPort, err.Error())
		}
		if err := m.cloud.StartLoadBalancerListener(m.ctx, lbId, l.ListenerPort); err != nil {
			return fmt.Errorf("start tcp listener [%d] error: %s", l.ListenerPort, err.Error())
		}
	}
	return nil
}

func (m *modelApplier) cleanup(local *model.LoadBalancer, remote *model.LoadBalancer) error {
	// clean up vservergroup
	/*lbId := remote.LoadBalancerAttribute.LoadBalancerId
	vgs, err := m.cloud.DescribeVServerGroups(m.ctx, lbId)
	if err != nil {
		return err
	}
	for _, vg := range vgs {
		if vg != local.NamespacedName.Name ||
			vg.NamedKey.Namespace != service.Namespace ||
			vg.NamedKey.CID != CLUSTER_ID {
			// skip those which does not belong to this service
			continue
		}
		found := false
		for _, l := range local.VServerGroups {
			if vg.NamedKey.Port == l.VGroupName {
				found = true
				break
			}
		}
		if !found {
			rem.Logf("try to remove unused vserver group, [%s][%s]", rem.NamedKey.Key(), rem.VGroupId)
			err := rem.Remove(ctx)
			if err != nil {
				rem.Logf("Error: cleanup vgroup warining: "+
					"failed to remove vgroup[%s]. wait for next try. %s", rem.NamedKey.Key(), err.Error())
				return err
			}
		}
	}*/
	return nil
}

func findVServerGroup(vgs []model.VServerGroup, port *model.ListenerAttribute) error {
	for _, vg := range vgs {
		if vg.VGroupName == port.VGroupName {
			port.VGroupId = vg.VGroupId
			return nil
		}
	}
	return fmt.Errorf("can not find vgroup by name %s", port.VGroupName)
}
