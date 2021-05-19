package service

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/klog"
	"sort"
)

func NewModelApplier(slbMgr *LoadBalancerManager, lisMgr *ListenerManager, vGroupMgr *VGroupManager) *ModelApplier {
	return &ModelApplier{
		slbMgr:    slbMgr,
		lisMgr:    lisMgr,
		vGroupMgr: vGroupMgr,
	}
}

type ModelApplier struct {
	slbMgr    *LoadBalancerManager
	lisMgr    *ListenerManager
	vGroupMgr *VGroupManager
}

func (m *ModelApplier) Apply(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {

	if err := m.applyLoadBalancerAttribute(reqCtx, local, remote); err != nil {
		return fmt.Errorf("update lb attribute error: %s", err.Error())
	}
	// if remote slb is not exist, return
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	if err := m.applyVGroups(reqCtx, local, remote); err != nil {
		return fmt.Errorf("update lb backends error: %s", err.Error())
	}

	if err := m.applyListeners(reqCtx, local, remote); err != nil {
		return fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	if err := m.cleanup(reqCtx, local, remote); err != nil {
		return fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	return nil
}

func (m *ModelApplier) applyLoadBalancerAttribute(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
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

	// delete slb
	if !isSLBNeeded(reqCtx.Service) {
		if !local.LoadBalancerAttribute.IsUserManaged {
			err := m.slbMgr.Delete(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("EnsureLoadBalancerDeleted [%s] error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}
			remote.LoadBalancerAttribute.LoadBalancerId = ""
			return nil
		}
	}

	// create slb
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		if err := m.slbMgr.Create(reqCtx, local); err != nil {
			return fmt.Errorf("EnsureLoadBalancerCreated error: %s", err.Error())
		}
		// update remote model
		remote.LoadBalancerAttribute.LoadBalancerId = local.LoadBalancerAttribute.LoadBalancerId
		if err := m.slbMgr.Find(reqCtx, remote); err != nil {
			return fmt.Errorf("update remote model for lbId %s, error: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
		}
		return nil
	}

	return m.slbMgr.Update(reqCtx, local, remote)

}

func (m *ModelApplier) applyVGroups(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
	for i := range local.VServerGroups {
		found := false
		for j := range remote.VServerGroups {
			if local.VServerGroups[i].VGroupName == remote.VServerGroups[j].VGroupName {
				found = true
				local.VServerGroups[i].VGroupId = remote.VServerGroups[j].VGroupId
				if err := m.vGroupMgr.UpdateVServerGroup(reqCtx, local.VServerGroups[i], remote.VServerGroups[j]); err != nil {
					return fmt.Errorf("EnsureVGroupUpdated error: %s", err.Error())
				}
			}
		}

		// if not exist, create
		if !found {
			err := m.vGroupMgr.CreateVServerGroup(reqCtx, &local.VServerGroups[i], remote.LoadBalancerAttribute.LoadBalancerId)
			if err != nil {
				return fmt.Errorf("EnsureVGroupCreated error: %s", err.Error())
			}
			remote.VServerGroups = append(remote.VServerGroups, local.VServerGroups[i])
		}
	}

	return nil
}

func (m *ModelApplier) applyListeners(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
	if local.LoadBalancerAttribute.IsUserManaged {
		if reqCtx.Anno.isForceOverride() {
			klog.Infof("[%s] override is false, skip reconcile listeners", local.NamespacedName)
			return nil
		}
	}
	klog.Infof("not user defined loadbalancer[%s], start to apply listener.", remote.LoadBalancerAttribute.LoadBalancerId)
	createActions, updateActions, deleteActions, err := buildActionsForListeners(local, remote)
	if err != nil {
		return fmt.Errorf("merge listener: %s", err.Error())
	}
	// make https come first.
	// ensure https listeners to be created first for http forward
	sort.SliceStable(
		createActions,
		func(i, j int) bool {
			if createActions[i].listener.Protocol == model.HTTPS {
				return true
			}
			return false
		},
	)
	// make https at last.
	// ensure https listeners to be deleted late for http forward
	sort.SliceStable(
		deleteActions,
		func(i, j int) bool {
			if deleteActions[i].listener.Protocol == model.HTTPS {
				return false
			}
			return true
		},
	)

	// Pls be careful of the sequence. deletion first,then addition, last updation
	for _, action := range deleteActions {
		err := m.lisMgr.Delete(reqCtx, action)
		if err != nil {
			return fmt.Errorf("delete listener [%d] error: %s", action.listener.ListenerPort, err.Error())
		}
	}

	for _, action := range createActions {
		err := m.lisMgr.Create(reqCtx, action)
		if err != nil {
			return fmt.Errorf("create listener [%d] error: %s", action.listener.ListenerPort, err.Error())
		}
	}

	for _, action := range updateActions {
		err := m.lisMgr.Update(reqCtx, action)
		if err != nil {
			return fmt.Errorf("update listener [%d] error: %s", action.local.ListenerPort, err.Error())
		}
	}

	return nil
}

// TODO notes
func (m *ModelApplier) cleanup(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
	// clean up vservergroup
	vgs := remote.VServerGroups
	for _, vg := range vgs {
		if vg.NamedKey.ServiceName != local.NamespacedName.Name ||
			vg.NamedKey.Namespace != local.NamespacedName.Namespace ||
			vg.NamedKey.CID != alibaba.CLUSTER_ID {
			// skip those which does not belong to this service
			continue
		}
		found := false
		for _, l := range local.VServerGroups {
			if vg.VGroupId == l.VGroupId {
				found = true
				break
			}
		}
		if !found {
			klog.Infof("try to remove unused vserver group, [%s][%s]", vg.NamedKey.Key(), vg.VGroupId)
			err := m.vGroupMgr.DeleteVServerGroup(reqCtx, vg.VGroupId)
			if err != nil {
				klog.Infof("Error: cleanup vgroup warining: "+
					"failed to remove vgroup[%s]. wait for next try. %s", vg.NamedKey.Key(), err.Error())
				return err
			}
		}
	}
	return nil
}
