package service

import (
	"context"
	"fmt"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
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

func (m *ModelApplier) Apply(reqCtx *RequestContext, local *model.LoadBalancer) (*model.LoadBalancer, error) {
	remote := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
	}

	err := m.slbMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return remote, fmt.Errorf("get load balancer attribute from cloud, error: %s", err.Error())
	}

	serviceHashChanged := isServiceHashChanged(reqCtx.Service)
	// apply sequence can not change, apply lb first, then vgroup, listener at last
	if serviceHashChanged || ctx2.GlobalFlag.DryRun {
		if err := m.applyLoadBalancerAttribute(reqCtx, local, remote); err != nil {
			return remote, fmt.Errorf("update lb attribute error: %s", err.Error())
		}
		// if remote slb is not exist, return
		if remote.LoadBalancerAttribute.LoadBalancerId == "" {
			return remote, nil
		}
	}
	reqCtx.Ctx = context.WithValue(reqCtx.Ctx, dryrun.ContextSLB, remote.LoadBalancerAttribute.LoadBalancerId)

	if err := m.vGroupMgr.BuildRemoteModel(reqCtx, remote); err != nil {
		return remote, fmt.Errorf("get lb backend from remote error: %s", err.Error())
	}
	if err := m.applyVGroups(reqCtx, local, remote); err != nil {
		return remote, fmt.Errorf("update lb backends error: %s", err.Error())
	}

	if serviceHashChanged || ctx2.GlobalFlag.DryRun {
		if err := m.lisMgr.BuildRemoteModel(reqCtx, remote); err != nil {
			return remote, fmt.Errorf("get lb listeners from cloud, error: %s", err.Error())
		}
		if err := m.applyListeners(reqCtx, local, remote); err != nil {
			return remote, fmt.Errorf("update lb listeners error: %s", err.Error())
		}
	}

	if err := m.cleanup(reqCtx, local, remote); err != nil {
		return remote, fmt.Errorf("update lb listeners error: %s", err.Error())
	}

	return remote, nil
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
	if needDeleteLoadBalancer(reqCtx.Service) {
		if !local.LoadBalancerAttribute.IsUserManaged {
			err := m.slbMgr.Delete(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("delete loadbalancer [%s] error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}
			remote.LoadBalancerAttribute.LoadBalancerId = ""
			reqCtx.Log.Infof("successfully delete slb %s", remote.LoadBalancerAttribute.LoadBalancerId)
			return nil
		}
	}

	// create slb
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		if isServiceOwnIngress(reqCtx.Service) {
			return fmt.Errorf("alicloud: not able to find loadbalancer, but it's defined in service.loaderbalancer.ingress [%v]"+
				"this may happen when you delete the loadbalancer", reqCtx.Service.Status.LoadBalancer.Ingress)
		}

		if err := m.slbMgr.Create(reqCtx, local); err != nil {
			return fmt.Errorf("create lb error: %s", err.Error())
		}
		reqCtx.Log.Infof("successfully create lb %s", local.LoadBalancerAttribute.LoadBalancerId)
		// update remote model
		remote.LoadBalancerAttribute.LoadBalancerId = local.LoadBalancerAttribute.LoadBalancerId
		if err := m.slbMgr.Find(reqCtx, remote); err != nil {
			return fmt.Errorf("update remote model for lbId %s, error: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
		}
		return nil
	}

	// update slb
	if local.LoadBalancerAttribute.IsUserManaged {
		tags, err := m.slbMgr.cloud.DescribeTags(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId)
		if err != nil {
			return fmt.Errorf("describe slb tags error: %s", err.Error())
		}

		if ok, reason := isLoadBalancerReusable(tags); !ok {
			return fmt.Errorf("alicloud: the loadbalancer %s can not be reused, %s", remote.LoadBalancerAttribute.LoadBalancerId, reason)
		}
	}

	return m.slbMgr.Update(reqCtx, local, remote)

}

func (m *ModelApplier) applyVGroups(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
	for i := range local.VServerGroups {
		found := false
		var old model.VServerGroup
		for _, rv := range remote.VServerGroups {
			// find by vgroup id first
			if local.VServerGroups[i].VGroupId != "" &&
				local.VServerGroups[i].VGroupId == rv.VGroupId {
				found = true
				old = rv
				break
			}
			// find by vgroup name
			if local.VServerGroups[i].VGroupName == rv.VGroupName {
				found = true
				local.VServerGroups[i].VGroupId = rv.VGroupId
				old = rv
				break
			}
		}

		// update
		if found {
			if err := m.vGroupMgr.UpdateVServerGroup(reqCtx, local.VServerGroups[i], old); err != nil {
				return fmt.Errorf("EnsureVGroupUpdated error: %s", err.Error())
			}
		}

		// create
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
			reqCtx.Log.Infof("listener override is false, skip reconcile listeners", local.NamespacedName)
			return nil
		}
	}
	createActions, updateActions, deleteActions, err := buildActionsForListeners(reqCtx, local, remote)
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

func (m *ModelApplier) cleanup(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
	// clean up vServerGroup
	vgs := remote.VServerGroups
	for _, vg := range vgs {
		if !isVGroupManagedByMyService(vg, reqCtx.Service) {
			reqCtx.Log.Infof("delete vgroup: [%s] description [%s] is managed by user, skip delete", vg.VGroupName, vg.VGroupId)
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
			reqCtx.Log.Infof("delete vgroup: [%s], [%s]", vg.NamedKey.Key(), vg.VGroupId)
			err := m.vGroupMgr.DeleteVServerGroup(reqCtx, vg.VGroupId)
			if err != nil {
				return fmt.Errorf("delete vgroup %s failed, error: %s", vg.VGroupId, err.Error())
			}
		}
	}
	return nil
}
