package service

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"k8s.io/klog"
	"sort"
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
	// if remote slb is not exist, return
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	if err := m.applyVGroups(local, remote); err != nil {
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

	// delete slb
	if !isSLBNeeded(m.reqCtx.svc) {
		if !local.LoadBalancerAttribute.IsUserManaged {
			err := m.reqCtx.EnsureLoadBalancerDeleted(remote)
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
	if local.LoadBalancerAttribute.IsUserManaged {
		if m.reqCtx.anno.isForceOverride() {
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

// TODO notes
func (m *modelApplier) cleanup(local *model.LoadBalancer, remote *model.LoadBalancer) error {
	// clean up vservergroup
	vgs := remote.VServerGroups
	for _, vg := range vgs {
		if vg.NamedKey.ServiceName != local.NamespacedName.Name ||
			vg.NamedKey.Namespace != local.NamespacedName.Namespace ||
			vg.NamedKey.CID != metadata.CLUSTER_ID {
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
			err := m.reqCtx.EnsureVGroupDeleted(vg.VGroupId)
			if err != nil {
				klog.Infof("Error: cleanup vgroup warining: "+
					"failed to remove vgroup[%s]. wait for next try. %s", vg.NamedKey.Key(), err.Error())
				return err
			}
		}
	}
	return nil
}
