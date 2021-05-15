package service

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/klog"
)

func NewLBModelApplier(reqCtx *RequestContext) *modelApplier {
	return &modelApplier{
		reqCtx,
	}
}

type modelApplier struct {
	reqCtx *RequestContext
}

func (m *modelApplier) Apply(local *model.LoadBalancer, remote *model.LoadBalancer) error {

	if err := m.applyLoadBalancerAttribute(local, remote); err != nil {
		return fmt.Errorf("update lb attribute error: %s", err.Error())
	}

	if err := m.applyVGroups(local, remote);
		err != nil {
		return fmt.Errorf("update lb backends error: %s", err.Error())
	}

	if err := m.applyListeners(local, remote); err != nil {
		return fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	if err := m.cleanup(local, remote); err != nil {
		return fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	return nil
}

func (m *modelApplier) applyLoadBalancerAttribute(local *model.LoadBalancer, remote *model.LoadBalancer) error {
	if local == nil {
		return fmt.Errorf("local model is nil")
	}
	if remote == nil {
		return fmt.Errorf("remote model is nil")
	}
	if local.NamespacedName.String() != remote.NamespacedName.String() {
		return fmt.Errorf("models for different svc, local [%s], remote [%s]",
			local.NamespacedName, remote.NamespacedName)
	}

	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		if err := m.reqCtx.EnsureLoadBalancerCreated(local); err != nil {
			return fmt.Errorf("EnsureLoadBalancerCreated error: %s", err.Error())
		}
		// update remote model
		remote.LoadBalancerAttribute.LoadBalancerId = local.LoadBalancerAttribute.LoadBalancerId
		if err := m.reqCtx.FindLoadBalancer(remote); err != nil {
			return fmt.Errorf("update remote model for lbId %s, error: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
		}
		return nil
	}

	klog.Infof("found load balancer [%s], try to update load balancer attribute",
		remote.LoadBalancerAttribute.LoadBalancerId)

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
	return m.reqCtx.EnsureLoadBalancerUpdated(local, remote)

}

func (m *modelApplier) applyVGroups(local *model.LoadBalancer, remote *model.LoadBalancer) error {
	for i := range local.VServerGroups {
		found := false
		for j := range remote.VServerGroups {
			if local.VServerGroups[i].VGroupName == remote.VServerGroups[j].VGroupName {
				found = true
				local.VServerGroups[i].VGroupId = remote.VServerGroups[j].VGroupId
				if err := m.reqCtx.EnsureVGroupUpdated(local.VServerGroups[i], remote.VServerGroups[j]); err != nil {
					return fmt.Errorf("EnsureVGroupUpdated error: %s", err.Error())
				}
			}
		}

		// if not exist, create
		if !found {
			err := m.reqCtx.EnsureVGroupCreated(&local.VServerGroups[i], remote.LoadBalancerAttribute.LoadBalancerId)
			if err != nil {
				return fmt.Errorf("EnsureVGroupCreated error: %s", err.Error())
			}
			remote.VServerGroups = append(remote.VServerGroups, local.VServerGroups[i])
		}
	}

	return nil
}

func (m *modelApplier) applyListeners(local *model.LoadBalancer, remote *model.LoadBalancer) error {

	createActions, updateActions, deleteActions, err := buildActionsForListeners(local, remote)
	if err != nil {
		return fmt.Errorf("merge listener: %s", err.Error())
	}
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

	// Pls be careful of the sequence. deletion first,then addition, last updation
	for _, action := range deleteActions {
		err := m.reqCtx.EnsureListenerDeleted(action)
		if err != nil {
			return fmt.Errorf("delete listener [%d] error: %s", action.listener.ListenerPort, err.Error())
		}
	}

	for _, action := range createActions {
		err := m.reqCtx.EnsureListenerCreated(action)
		if err != nil {
			return fmt.Errorf("create listener [%d] error: %s", action.listener.ListenerPort, err.Error())
		}
	}

	for _, action := range updateActions {
		err := m.reqCtx.EnsureListenerUpdated(action)
		if err != nil {
			return fmt.Errorf("update listener [%d] error: %s", action.local.ListenerPort, err.Error())
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
