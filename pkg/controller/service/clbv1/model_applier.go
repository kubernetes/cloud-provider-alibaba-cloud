package clbv1

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrlcfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/parallel"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"sync"
)

type serverGroupApplyResult struct {
	VGroupName string
	VGroupID   string
	Err        error
}

type ModelApplier struct {
	slbMgr    *LoadBalancerManager
	lisMgr    *ListenerManager
	vGroupMgr *VGroupManager
}

func NewModelApplier(slbMgr *LoadBalancerManager, lisMgr *ListenerManager, vGroupMgr *VGroupManager) *ModelApplier {
	return &ModelApplier{
		slbMgr:    slbMgr,
		lisMgr:    lisMgr,
		vGroupMgr: vGroupMgr,
	}
}

func (m *ModelApplier) Apply(reqCtx *svcCtx.RequestContext, local *model.LoadBalancer) (*model.LoadBalancer, error) {
	if local == nil {
		return nil, fmt.Errorf("local model is nil")
	}
	if util.NamespacedName(reqCtx.Service).String() != local.NamespacedName.String() {
		return nil, fmt.Errorf("local model namespaced name %s not match request context service %s",
			local.NamespacedName, util.NamespacedName(reqCtx.Service))
	}

	hashChangedOrDryRun := helper.IsServiceHashChanged(reqCtx.Service) || ctrlcfg.ControllerCFG.DryRun
	remote := &model.LoadBalancer{
		NamespacedName:                  util.NamespacedName(reqCtx.Service),
		ContainsPotentialReadyEndpoints: local.ContainsPotentialReadyEndpoints,
	}

	err := m.buildRemoteModel(reqCtx, local, remote)
	if err != nil {
		return remote, err
	}

	if remote.LoadBalancerAttribute.LoadBalancerId != "" && local.LoadBalancerAttribute.PreserveOnDelete {
		reqCtx.Recorder.Eventf(reqCtx.Service, v1.EventTypeWarning, helper.PreservedOnDelete,
			"The lb [%s] will be preserved after the service is deleted.", remote.LoadBalancerAttribute.LoadBalancerId)
	}

	var tasks []func() error

	if hashChangedOrDryRun {
		err := m.applyLoadBalancer(reqCtx, local, remote)
		if err != nil {
			return remote, fmt.Errorf("update lb attribute error: %s", err.Error())
		}

		tasks = append(tasks, func() error {
			if err := m.slbMgr.Update(reqCtx, local, remote); err != nil {
				return fmt.Errorf("update lb attribute error: %s", err.Error())
			}
			return nil
		})
	}

	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		if helper.NeedDeleteLoadBalancer(reqCtx.Service) {
			return remote, nil
		}
		return remote, fmt.Errorf("alicloud: can not find loadbalancer by tag [%s:%s]",
			helper.TAGKEY, reqCtx.Anno.GetDefaultLoadBalancerName())
	}

	var sgChannel chan serverGroupApplyResult
	if hashChangedOrDryRun {
		sgChannel = make(chan serverGroupApplyResult, len(local.VServerGroups))
	}

	actions, err := buildVGroupCreateAndUpdateActions(reqCtx, local, remote)
	tasks = append(tasks, func() error {
		return m.applyVGroups(reqCtx, actions, sgChannel)
	})

	if hashChangedOrDryRun {
		if local.LoadBalancerAttribute.IsUserManaged && !reqCtx.Anno.IsForceOverride() {
			reqCtx.Log.Info("listener override is false, skip reconcile listeners")
		} else {
			createActions, updateActions, deleteActions, err := buildActionsForListeners(reqCtx, local, remote)
			if err != nil {
				return remote, fmt.Errorf("merge listener: %s", err.Error())
			}

			tasks = append(tasks, func() error {
				return m.applyListeners(reqCtx, createActions, updateActions, deleteActions, sgChannel)
			})
		}
	}

	err = parallel.Do(tasks...)
	if err != nil {
		return remote, err
	}

	if hashChangedOrDryRun {
		if err := m.cleanup(reqCtx, local, remote); err != nil {
			return remote, fmt.Errorf("cleanup lb error: %s", err)
		}
	}

	return remote, err
}

func (m *ModelApplier) buildRemoteModel(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) error {
	err := m.slbMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return fmt.Errorf("get load balancer attribute from cloud, error: %s", err.Error())
	}

	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	klog.Infof("%s find clb with result, reconcileID: %s\n%+v", util.Key(reqCtx.Service), reqCtx.ReconcileID, util.PrettyJson(remote))

	tasks := []func() error{
		func() error {
			tags, err := m.slbMgr.cloud.ListCLBTagResources(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId)
			if err != nil {
				return fmt.Errorf("DescribeTags: %s", err.Error())
			}
			remote.LoadBalancerAttribute.Tags = tags
			return nil
		},
		func() error {
			if err := m.vGroupMgr.BuildRemoteModel(reqCtx, remote); err != nil {
				return fmt.Errorf("get lb backend from remote error: %s", err.Error())
			}
			return nil
		},
	}

	if helper.IsServiceHashChanged(reqCtx.Service) || ctrlcfg.ControllerCFG.DryRun {
		if !local.LoadBalancerAttribute.IsUserManaged || reqCtx.Anno.IsForceOverride() {
			tasks = append(tasks, func() error {
				if err := m.lisMgr.BuildRemoteModel(reqCtx, remote); err != nil {
					return fmt.Errorf("get lb listeners from cloud, error: %s", err.Error())
				}
				return nil
			})
		}
	}
	return parallel.Do(tasks...)
}

func (m *ModelApplier) applyLoadBalancer(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) error {
	// delete
	if helper.NeedDeleteLoadBalancer(reqCtx.Service) {
		err := m.deleteLoadBalancer(reqCtx, local, remote)
		if err != nil {
			return err
		}
		return nil
	}

	// create
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		if helper.IsServiceOwnIngress(reqCtx.Service) {
			return fmt.Errorf("alicloud: can not find loadbalancer, but it's defined in service [%v] "+
				"this may happen when you delete the loadbalancer", reqCtx.Service.Status.LoadBalancer.Ingress[0].IP)
		}

		if err := m.slbMgr.Create(reqCtx, local); err != nil {
			return fmt.Errorf("create lb error: %s", err.Error())
		}
		reqCtx.Log.Info(fmt.Sprintf("successfully create lb %s", local.LoadBalancerAttribute.LoadBalancerId))
		// update remote model
		remote.LoadBalancerAttribute.LoadBalancerId = local.LoadBalancerAttribute.LoadBalancerId
		if err := m.slbMgr.Find(reqCtx, remote); err != nil {
			return fmt.Errorf("update remote model for lbId %s, error: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
		}
	}

	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		// update loadbalancer: return error
		return fmt.Errorf("alicloud: can not find loadbalancer by tag [%s:%s]",
			helper.TAGKEY, reqCtx.Anno.GetDefaultLoadBalancerName())
	}

	// check whether slb can be reused
	if local.LoadBalancerAttribute.IsUserManaged {
		if ok, reason := isLoadBalancerReusable(reqCtx, remote.LoadBalancerAttribute.Tags, remote.LoadBalancerAttribute.Address); !ok {
			return fmt.Errorf("alicloud: the loadbalancer %s can not be reused, %s",
				remote.LoadBalancerAttribute.LoadBalancerId, reason)
		}
	}
	return nil
}

func (m *ModelApplier) deleteLoadBalancer(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) error {
	if !local.LoadBalancerAttribute.IsUserManaged {
		if local.LoadBalancerAttribute.PreserveOnDelete {
			err := m.slbMgr.SetProtectionsOff(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("set loadbalancer [%s] protections off error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}

			err = m.slbMgr.CleanupLoadBalancerTags(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("cleanup loadbalancer [%s] tags error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}
			reqCtx.Log.Info(fmt.Sprintf("successfully cleanup preserved slb %s", remote.LoadBalancerAttribute.LoadBalancerId))
		} else {
			err := m.slbMgr.Delete(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("delete loadbalancer [%s] error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}
			reqCtx.Log.Info(fmt.Sprintf("successfully delete slb %s", remote.LoadBalancerAttribute.LoadBalancerId))
			remote.LoadBalancerAttribute.LoadBalancerId = ""
			remote.LoadBalancerAttribute.Address = ""
		}
		return nil
	}

	reqCtx.Log.Info(fmt.Sprintf("slb %s is reused, skip delete it", remote.LoadBalancerAttribute.LoadBalancerId))
	return nil
}

func (m *ModelApplier) applyVGroups(reqCtx *svcCtx.RequestContext, actions []vGroupAction, serverGroupChannel chan serverGroupApplyResult) error {
	if serverGroupChannel != nil {
		defer close(serverGroupChannel)
	}

	errs := m.vGroupMgr.ParallelUpdateVServerGroup(reqCtx, actions, serverGroupChannel)
	return utilerrors.NewAggregate(errs)
}

const (
	listenerActionTypeCreate = "create"
	listenerActionTypeUpdate = "update"
	listenerActionTypeDelete = "delete"
)

type listenerAction struct {
	ActionType string
	Create     CreateAction
	Update     UpdateAction
	Delete     DeleteAction
}

func (m *ModelApplier) applyListeners(reqCtx *svcCtx.RequestContext,
	createActions []CreateAction, updateActions []UpdateAction, deleteActions []DeleteAction, serverGroupChannel chan serverGroupApplyResult) error {

	var errs []error
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}

	beforeActions, actions, afterActions := buildActionsToUpdate(reqCtx, createActions, updateActions, deleteActions)

	listenerMap := map[string][]listenerAction{}
	for _, a := range actions {
		switch a.ActionType {
		case listenerActionTypeCreate:
			listenerMap[a.Create.listener.VGroupName] = append(listenerMap[a.Create.listener.VGroupName], a)
		case listenerActionTypeUpdate:
			listenerMap[a.Update.local.VGroupName] = append(listenerMap[a.Update.local.VGroupName], a)
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		retErrs := m.parallelDoListenerActions(reqCtx, beforeActions)
		if len(retErrs) != 0 {
			mtx.Lock()
			errs = append(errs, retErrs...)
			mtx.Unlock()
		}
	}()

	for n := range serverGroupChannel {
		reqCtx.Log.V(5).Info("receive server group apply result", "result", n)
		if n.Err != nil {
			reqCtx.Log.Error(n.Err, "error applying servergroups, skip listener actions",
				"vgroupName", n.VGroupName)
			continue
		}
		actions := listenerMap[n.VGroupName]
		if len(actions) == 0 {
			continue
		}
		for i := range actions {
			switch actions[i].ActionType {
			case listenerActionTypeCreate:
				actions[i].Create.listener.VGroupId = n.VGroupID
			case listenerActionTypeUpdate:
				actions[i].Update.local.VGroupId = n.VGroupID
			}
		}
		delete(listenerMap, n.VGroupName)
		wg.Add(1)
		go func() {
			defer wg.Done()
			retErrs := m.parallelDoListenerActions(reqCtx, actions)
			if len(retErrs) != 0 {
				mtx.Lock()
				errs = append(errs, retErrs...)
				mtx.Unlock()
			}
		}()
	}

	wg.Wait()

	// do later actions
	// listeners with empty vgroup name, eg. forward to http
	retErrs := m.parallelDoListenerActions(reqCtx, afterActions)
	errs = append(errs, retErrs...)

	if len(listenerMap) != 0 {
		errs = append(errs, fmt.Errorf("there are unprocessed listeners in reconcile: %+v", listenerMap))
	}

	return utilerrors.NewAggregate(errs)
}

func (m *ModelApplier) parallelDoListenerActions(reqCtx *svcCtx.RequestContext, actions []listenerAction) []error {
	if len(actions) == 0 {
		return nil
	}
	errs := make([]error, len(actions))
	parallel.DoPiece(reqCtx.Ctx, ctrlcfg.ControllerCFG.MaxConcurrentActions, len(actions), func(i int) {
		switch actions[i].ActionType {
		case listenerActionTypeCreate:
			if err := m.lisMgr.Create(reqCtx, actions[i].Create); err != nil {
				errs[i] = fmt.Errorf("create listener error: %s", err.Error())
			}
		case listenerActionTypeUpdate:
			if err := m.lisMgr.Update(reqCtx, actions[i].Update); err != nil {
				errs[i] = fmt.Errorf("update listener error: %s", err.Error())
			}
		case listenerActionTypeDelete:
			if err := m.lisMgr.Delete(reqCtx, actions[i].Delete); err != nil {
				errs[i] = fmt.Errorf("delete listener error: %s", err.Error())
			}
		}
	})
	var ret []error
	for _, e := range errs {
		if e != nil {
			ret = append(ret, e)
		}
	}
	return ret
}

func buildVGroupCreateAndUpdateActions(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) ([]vGroupAction, error) {
	var actions []vGroupAction
	updatedVGroups := map[string]bool{}

	for i := range local.VServerGroups {
		found := false
		var old *model.VServerGroup
		sgKey := "name-" + local.VServerGroups[i].VGroupName
		if local.VServerGroups[i].VGroupId != "" {
			sgKey = "id-" + local.VServerGroups[i].VGroupId
		}
		for _, rv := range remote.VServerGroups {
			// for reuse vgroup case, find by vgroup id first
			if local.VServerGroups[i].VGroupId != "" &&
				local.VServerGroups[i].VGroupId == rv.VGroupId {
				found = true
				old = &rv
				break
			}
			// find by vgroup name
			if local.VServerGroups[i].VGroupId == "" &&
				local.VServerGroups[i].VGroupName == rv.VGroupName {
				local.VServerGroups[i].VGroupId = rv.VGroupId
				found = true
				old = &rv
				break
			}
		}

		if updatedVGroups[sgKey] {
			reqCtx.Log.Info("already updated vgroup, skip",
				"vgroupID", local.VServerGroups[i].VGroupId, "vgroupName", local.VServerGroups[i].VGroupName)
			continue
		}

		// update
		if found {
			add, del, update := diff(*old, local.VServerGroups[i])
			if len(add) == 0 && len(del) == 0 && len(update) == 0 {
				reqCtx.Log.Info(fmt.Sprintf("reconcile vgroup: [%s] not change, skip reconcile", old.VGroupId),
					"vgroupName", old.VGroupName)
				continue
			}
			actions = append(actions, vGroupAction{
				Action: vGroupActionUpdate,
				LBId:   remote.LoadBalancerAttribute.LoadBalancerId,
				Local:  &local.VServerGroups[i],
				Remote: old,
			})
			updatedVGroups[sgKey] = true
		}

		// create
		if !found {
			reqCtx.Log.Info(fmt.Sprintf("try to create vgroup %s", local.VServerGroups[i].VGroupName))
			actions = append(actions, vGroupAction{
				Action: vGroupActionCreateAndAddBackendServers,
				LBId:   remote.LoadBalancerAttribute.LoadBalancerId,
				Local:  &local.VServerGroups[i],
			})
			updatedVGroups[sgKey] = true
		}
	}

	return actions, nil
}

func (m *ModelApplier) cleanup(reqCtx *svcCtx.RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) error {
	// clean up vServerGroup
	vgs := remote.VServerGroups
	var actions []vGroupAction
	for _, vg := range vgs {
		found := false
		for _, l := range local.VServerGroups {
			if vg.VGroupId == l.VGroupId {
				found = true
				break
			}
		}

		// delete unused vgroup
		if !found {
			// do not delete user managed vgroup, but need to clean the backends in the vgroup
			if !isVGroupManagedByMyService(vg, reqCtx.Service) {
				reqCtx.Log.Info(fmt.Sprintf("try to delete vgroup: [%s] description [%s] is managed by user, skip delete",
					vg.VGroupName, vg.VGroupId))
				var del []model.BackendAttribute
				for _, remote := range vg.Backends {
					if !remote.IsUserManaged {
						del = append(del, remote)
					}
				}
				if len(del) > 0 {
					if err := m.vGroupMgr.BatchRemoveVServerGroupBackendServers(reqCtx, vg, del); err != nil {
						return err
					}
				}
				continue
			}

			reqCtx.Log.Info(fmt.Sprintf("delete vgroup: [%s], [%s]", vg.NamedKey.Key(), vg.VGroupId))
			//err := m.vGroupMgr.DeleteVServerGroup(reqCtx, vg.VGroupId)
			actions = append(actions, vGroupAction{
				Action: vGroupActionDelete,
				Remote: &model.VServerGroup{
					VGroupId: vg.VGroupId,
				},
			})
		}
	}

	errs := m.vGroupMgr.ParallelUpdateVServerGroup(reqCtx, actions, nil)
	return utilerrors.NewAggregate(errs)
}

func isLoadBalancerReusable(reqCtx *svcCtx.RequestContext, tags []tag.Tag, lbIp string) (bool, string) {
	for _, tag := range tags {
		// the tag of the apiserver slb is "ack.aliyun.com": "${clusterid}",
		// so can not reuse slbs which have ack.aliyun.com tag key.
		if tag.Key == helper.TAGKEY || tag.Key == util.ClusterTagKey {
			return false, "can not reuse loadbalancer created by kubernetes."
		}
	}

	// if use eip as externalIPType, ingress IP is eip, skip to check
	if reqCtx.Anno.Get(annotation.ExternalIPType) == "eip" {
		return true, ""
	}

	service := reqCtx.Service
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		found := false
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP == lbIp || (ingress.Hostname != "" && ingress.IP == "") {
				found = true
			}
		}
		if !found {
			return false, fmt.Sprintf("service has been associated with ip [%v], cannot be bound to ip [%s]",
				service.Status.LoadBalancer.Ingress[0].IP, lbIp)
		}
	}

	return true, ""
}

func hasDeleteActionToPort(port int, actions []DeleteAction) bool {
	for _, a := range actions {
		if a.listener.ListenerPort == port {
			return true
		}
	}
	return false
}

func hasDeleteActionForwardedTo(action DeleteAction, actions []DeleteAction) bool {
	port := action.listener.ListenerPort
	for _, a := range actions {
		if a.listener.ListenerForward != model.OnFlag {
			continue
		}

		if a.listener.ForwardPort == port {
			return true
		}
	}
	return false
}

func buildActionsToUpdate(reqCtx *svcCtx.RequestContext, createActions []CreateAction, updateActions []UpdateAction, deleteActions []DeleteAction) ([]listenerAction, []listenerAction, []listenerAction) {
	var beforeActions []listenerAction
	var actions []listenerAction
	var afterActions []listenerAction
	for _, action := range updateActions {
		act := listenerAction{
			ActionType: listenerActionTypeUpdate,
			Update:     action,
		}
		// vgroup is not ready now
		if action.local.VGroupId != "" {
			beforeActions = append(beforeActions, act)
		} else {
			actions = append(actions, act)
		}
	}

	for _, action := range deleteActions {
		act := listenerAction{
			ActionType: listenerActionTypeDelete,
			Delete:     action,
		}
		if hasDeleteActionForwardedTo(action, deleteActions) {
			afterActions = append(afterActions, act)
		} else {
			beforeActions = append(beforeActions, act)
		}
	}

	for _, action := range createActions {
		act := listenerAction{
			ActionType: listenerActionTypeCreate,
			Create:     action,
		}
		if action.listener.ListenerForward == model.OnFlag {
			reqCtx.Log.V(5).Info("listener will be created later because it depends on another listener",
				"listener", action.listener.ListenerPort, "forwardPort", action.listener.ForwardPort)
			afterActions = append(afterActions, act)
		} else if hasDeleteActionToPort(action.listener.ListenerPort, deleteActions) && action.listener.VGroupId != "" {
			reqCtx.Log.V(5).Info("listener will be created later because the deletion action of the same port need to be done",
				"listener", action.listener.ListenerPort)
			afterActions = append(afterActions, act)
		} else if action.listener.VGroupId != "" {
			beforeActions = append(beforeActions, act)
		} else {
			actions = append(actions, act)
		}
	}

	return beforeActions, actions, afterActions
}
