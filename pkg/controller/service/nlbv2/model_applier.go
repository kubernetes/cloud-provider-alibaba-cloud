package nlbv2

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/parallel"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sync"
)

type serverGroupCreateResult struct {
	ServerGroupName string
	ServerGroupID   string
	Err             error
}

type ParallelizeModelApplier struct {
	nlbMgr *NLBManager
	lisMgr *ListenerManager
	sgMgr  *ServerGroupManager
}

func NewModelApplier(nlbMgr *NLBManager, lisMgr *ListenerManager, serverGroupManager *ServerGroupManager) *ParallelizeModelApplier {
	return &ParallelizeModelApplier{
		nlbMgr: nlbMgr,
		lisMgr: lisMgr,
		sgMgr:  serverGroupManager,
	}
}

func (m *ParallelizeModelApplier) Apply(reqCtx *svcCtx.RequestContext, local *nlbmodel.NetworkLoadBalancer) (*nlbmodel.NetworkLoadBalancer, error) {
	if local == nil {
		return nil, fmt.Errorf("local or remote mdl is nil")
	}
	if util.NamespacedName(reqCtx.Service).String() != local.NamespacedName.String() {
		return nil, fmt.Errorf("models for different svc, local [%s], remote [%s]",
			local.NamespacedName, util.NamespacedName(reqCtx.Service))
	}

	hashChangedOrDryRun := helper.IsServiceHashChanged(reqCtx.Service) || ctrlCfg.ControllerCFG.DryRun
	remote := &nlbmodel.NetworkLoadBalancer{
		NamespacedName:                  util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute:           &nlbmodel.LoadBalancerAttribute{},
		ContainsPotentialReadyEndpoints: local.ContainsPotentialReadyEndpoints,
	}

	err := m.buildRemoteModel(reqCtx, local, remote)
	if err != nil {
		return remote, err
	}
	reqCtx.Ctx = context.WithValue(reqCtx.Ctx, dryrun.ContextNLB, remote.GetLoadBalancerId())

	if remote.LoadBalancerAttribute.LoadBalancerId != "" && local.LoadBalancerAttribute.PreserveOnDelete {
		reqCtx.Recorder.Eventf(reqCtx.Service, v1.EventTypeWarning, helper.PreservedOnDelete,
			"The lb [%s] will be preserved after the service is deleted.", remote.LoadBalancerAttribute.LoadBalancerId)
	}

	var tasks []func() error
	if hashChangedOrDryRun {
		err := m.applyLoadBalancer(reqCtx, local, remote)
		if err != nil {
			return remote, fmt.Errorf("update nlb attribute error: %s", err.Error())
		}

		tasks = append(tasks, func() error {
			if err := m.nlbMgr.Update(reqCtx, local, remote); err != nil {
				return fmt.Errorf("update nlb attribute error: %s", err.Error())
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

	var sgChannel chan serverGroupCreateResult
	if hashChangedOrDryRun {
		sgChannel = make(chan serverGroupCreateResult, len(local.ServerGroups))
	}

	actions, err := m.buildServerGroupCreateAndUpdateActions(reqCtx, local, remote)
	if err != nil {
		return remote, err
	}

	tasks = append(tasks, func() error {
		return m.applyServerGroups(reqCtx, actions, sgChannel)
	})

	if hashChangedOrDryRun {
		if local.LoadBalancerAttribute.IsUserManaged && !reqCtx.Anno.IsForceOverride() {
			reqCtx.Log.Info("listener override is false, skip reconcile listeners")
		} else {
			actions, err := buildActionsForListeners(reqCtx, local, remote)
			if err != nil {
				return remote, err
			}
			tasks = append(tasks, func() error {
				return m.applyListeners(reqCtx, actions, sgChannel)
			})
		}
	}

	err = parallel.Do(tasks...)
	if err != nil {
		return remote, err
	}

	if hashChangedOrDryRun {
		err = m.cleanup(reqCtx, local, remote)
		if err != nil {
			return remote, err
		}
	}

	return remote, err
}

func (m *ParallelizeModelApplier) buildRemoteModel(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	err := m.nlbMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return fmt.Errorf("get nlb attribute from cloud error: %s", err.Error())
	}

	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	tasks := []func() error{
		func() error {
			tags, err := m.nlbMgr.cloud.ListNLBTagResources(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId)
			if err != nil {
				return fmt.Errorf("ListNLBTagResources: %s", err.Error())
			}
			remote.LoadBalancerAttribute.Tags = tags
			return nil
		},
		func() error {
			if err := m.sgMgr.BuildRemoteModel(reqCtx, remote); err != nil {
				return fmt.Errorf("get lb backend from remote error: %s", err.Error())
			}
			return nil
		},
	}

	if helper.IsServiceHashChanged(reqCtx.Service) || ctrlCfg.ControllerCFG.DryRun {
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

func (m *ParallelizeModelApplier) applyLoadBalancer(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	// delete nlb
	if helper.NeedDeleteLoadBalancer(reqCtx.Service) {
		err := m.deleteLoadBalancer(reqCtx, local, remote)
		if err != nil {
			return err
		}
		return nil
	}

	// create nlb
	// todo: requeue lb if lb is not ready
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		if helper.IsServiceOwnIngress(reqCtx.Service) {
			return fmt.Errorf("alicloud: can not find loadbalancer, but it's defined in service [%v] "+
				"this may happen when you delete the loadbalancer", reqCtx.Service.Status.LoadBalancer.Ingress[0].IP)
		}

		if err := m.nlbMgr.Create(reqCtx, local); err != nil {
			return fmt.Errorf("create nlb error: %s", err.Error())
		}
		reqCtx.Log.Info(fmt.Sprintf("successfully create lb %s", local.LoadBalancerAttribute.LoadBalancerId))
		// update remote model
		remote.LoadBalancerAttribute.LoadBalancerId = local.LoadBalancerAttribute.LoadBalancerId
		if err := m.nlbMgr.Find(reqCtx, remote); err != nil {
			return fmt.Errorf("update remote model for lbId %s, error: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
		}
		// need update nlb security groups
		// or ipv6 address type
		if len(local.LoadBalancerAttribute.SecurityGroupIds) != 0 ||
			local.LoadBalancerAttribute.IPv6AddressType != "" {
			err := m.nlbMgr.Update(reqCtx, local, remote)
			if err != nil {
				return err
			}
		}
	}

	// check whether slb can be reused
	if !helper.NeedDeleteLoadBalancer(reqCtx.Service) && local.LoadBalancerAttribute.IsUserManaged {
		if ok, reason := isNLBReusable(reqCtx.Service, remote.LoadBalancerAttribute.Tags, remote.LoadBalancerAttribute.DNSName); !ok {
			return fmt.Errorf("the loadbalancer %s can not be reused, %s",
				remote.LoadBalancerAttribute.LoadBalancerId, reason)
		}
	}

	return nil
}

func (m *ParallelizeModelApplier) deleteLoadBalancer(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	if !local.LoadBalancerAttribute.IsUserManaged {
		err := m.sgMgr.BuildRemoteModel(reqCtx, remote)
		if err != nil {
			return fmt.Errorf("build remote server group model error: %s", err.Error())
		}
		if local.LoadBalancerAttribute.PreserveOnDelete {
			err := m.nlbMgr.SetProtectionsOff(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("set loadbalancer [%s] protections off error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}

			err = m.nlbMgr.CleanupLoadBalancerTags(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("cleanup loadbalancer [%s] tags error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}
			reqCtx.Log.Info(fmt.Sprintf("successfully cleanup preserved nlb %s", remote.LoadBalancerAttribute.LoadBalancerId))
		} else {
			err := m.nlbMgr.Delete(reqCtx, remote)
			if err != nil {
				return fmt.Errorf("delete nlb [%s] error: %s",
					remote.LoadBalancerAttribute.LoadBalancerId, err.Error())
			}
			reqCtx.Log.Info(fmt.Sprintf("successfully delete nlb %s", remote.LoadBalancerAttribute.LoadBalancerId))
			remote.LoadBalancerAttribute.LoadBalancerId = ""
			remote.LoadBalancerAttribute.DNSName = ""
		}

		return nil
	}
	reqCtx.Log.Info(fmt.Sprintf("nlb %s is reused, skip delete it", remote.LoadBalancerAttribute.LoadBalancerId))
	return nil
}

func (m *ParallelizeModelApplier) buildServerGroupCreateAndUpdateActions(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) ([]serverGroupAction, error) {
	var actions []serverGroupAction
	updatedServerGroups := map[string]bool{}

	for i := range local.ServerGroups {
		found := false
		var old nlbmodel.ServerGroup
		sgKey := fmt.Sprintf("name-%s", local.ServerGroups[i].ServerGroupName)
		if local.ServerGroups[i].ServerGroupId != "" {
			sgKey = fmt.Sprintf("id-%s", local.ServerGroups[i].ServerGroupId)
		}
		for _, rv := range remote.ServerGroups {
			// for reuse vgroup case, find by vgroup id first
			if local.ServerGroups[i].ServerGroupId != "" &&
				local.ServerGroups[i].ServerGroupId == rv.ServerGroupId {
				found = true
				old = *rv
				break
			}
			// find by vgroup name
			if local.ServerGroups[i].ServerGroupId == "" &&
				local.ServerGroups[i].ServerGroupName == rv.ServerGroupName {
				found = true
				local.ServerGroups[i].ServerGroupId = rv.ServerGroupId
				old = *rv
				break
			}
		}

		if updatedServerGroups[sgKey] {
			reqCtx.Log.Info("already updated server group, skip",
				"sgID", local.ServerGroups[i].ServerGroupId, "sgName", local.ServerGroups[i].ServerGroupId)
			continue
		}

		// update
		if found {
			// if server group type changed, need to recreate
			if local.ServerGroups[i].ServerGroupType != "" &&
				local.ServerGroups[i].ServerGroupType != old.ServerGroupType {
				if local.ServerGroups[i].IsUserManaged {
					return nil, fmt.Errorf("ServerGroupType of user managed server group %s should be [%s], but [%s]",
						local.ServerGroups[i].ServerGroupId, local.ServerGroups[i].ServerGroupType, old.ServerGroupType)
				}
				reqCtx.Log.Info(fmt.Sprintf("ServerGroupType changed [%s] - [%s], need to recreate server group",
					old.ServerGroupType, local.ServerGroups[i].ServerGroupType),
					"sgId", old.ServerGroupId, "sgName", old.ServerGroupName)
				found = false
			} else {
				actions = append(actions, serverGroupAction{
					Action: serverGroupActionUpdate,
					Local:  local.ServerGroups[i],
					Remote: &old,
				})
				updatedServerGroups[sgKey] = true
			}
		}

		// create
		if !found {
			reqCtx.Log.Info(fmt.Sprintf("create server group %s", local.ServerGroups[i].ServerGroupName))
			if remote.LoadBalancerAttribute.VpcId != "" {
				local.ServerGroups[i].VPCId = remote.LoadBalancerAttribute.VpcId
			}
			actions = append(actions, serverGroupAction{
				Action: serverGroupActionCreateAndAddBackendServers,
				Local:  local.ServerGroups[i],
				Remote: local.ServerGroups[i],
			})
			remote.ServerGroups = append(remote.ServerGroups, local.ServerGroups[i])
			updatedServerGroups[sgKey] = true
		}
	}

	return actions, nil
}

func (m *ParallelizeModelApplier) applyServerGroups(reqCtx *svcCtx.RequestContext, actions []serverGroupAction, serverGroupChannel chan serverGroupCreateResult) error {
	if serverGroupChannel != nil {
		defer close(serverGroupChannel)
	}

	errs := m.sgMgr.ParallelUpdateServerGroups(reqCtx, actions, serverGroupChannel)
	return utilerrors.NewAggregate(errs)
}

func (m *ParallelizeModelApplier) applyListeners(reqCtx *svcCtx.RequestContext, actions []listenerAction, serverGroupChannel chan serverGroupCreateResult) error {
	var nowActions []listenerAction
	var waitActions []listenerAction
	listenerMap := map[string][]listenerAction{}
	for _, act := range actions {
		switch act.Action {
		case listenerActionDelete:
			nowActions = append(nowActions, act)
		default:
			if act.Local.ServerGroupId == "" {
				waitActions = append(waitActions, act)
				listenerMap[act.Local.ServerGroupName] = append(listenerMap[act.Local.ServerGroupName], act)
			} else {
				nowActions = append(nowActions, act)
			}
		}
	}

	var errs []error
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		retErrs := m.lisMgr.ParallelUpdateListeners(reqCtx, nowActions)
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
				"vgroupName", n.ServerGroupName)
			continue
		}
		actions := listenerMap[n.ServerGroupName]
		if len(actions) == 0 {
			continue
		}
		for i := range actions {
			actions[i].Local.ServerGroupId = n.ServerGroupID
		}
		delete(listenerMap, n.ServerGroupName)
		wg.Add(1)
		go func() {
			defer wg.Done()
			retErrs := m.lisMgr.ParallelUpdateListeners(reqCtx, actions)
			if len(retErrs) != 0 {
				mtx.Lock()
				errs = append(errs, retErrs...)
				mtx.Unlock()
			}
		}()
	}

	wg.Wait()
	if len(listenerMap) != 0 {
		errs = append(errs, fmt.Errorf("there are unprocessed listeners in reconciles: %+v", listenerMap))
	}
	return utilerrors.NewAggregate(errs)
}

func buildActionsForListeners(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) ([]listenerAction, error) {
	var actions []listenerAction

	// associate listener and vGroup
	for i := range local.Listeners {
		if local.Listeners[i].ServerGroupId != "" {
			continue
		}
		if err := findServerGroup(local.ServerGroups, local.Listeners[i]); err != nil {
			return nil, fmt.Errorf("find servergroup error: %s", err.Error())
		}
	}

	// delete
	for _, r := range remote.Listeners {
		found := false
		for i, l := range local.Listeners {
			if isListenerPortMatch(l, r) && r.ListenerProtocol == l.ListenerProtocol {
				found = true
				local.Listeners[i].ListenerId = r.ListenerId
			}
		}

		if !found {
			if local.LoadBalancerAttribute.IsUserManaged {
				if r.NamedKey == nil || !r.NamedKey.IsManagedByService(reqCtx.Service, base.CLUSTER_ID) {
					reqCtx.Log.V(5).Info(fmt.Sprintf("listener %s is managed by user, skip delete", r.ListenerId))
					continue
				}
			}

			reqCtx.Log.Info(fmt.Sprintf("delete listener: %s [%s]", r.ListenerProtocol, r.PortString()))
			actions = append(actions, listenerAction{
				Action: listenerActionDelete,
				Remote: r,
			})
		}
	}

	for i := range local.Listeners {
		found := false
		for j := range remote.Listeners {
			if local.Listeners[i].ListenerId == remote.Listeners[j].ListenerId {
				found = true
				actions = append(actions, listenerAction{
					Action: listenerActionUpdate,
					Local:  local.Listeners[i],
					Remote: remote.Listeners[j],
				})
			}
		}

		// create
		if !found {
			reqCtx.Log.Info(fmt.Sprintf("create listener: %s [%s]", local.Listeners[i].ListenerProtocol, local.Listeners[i].PortString()))
			actions = append(actions, listenerAction{
				Action: listenerActionCreate,
				Local:  local.Listeners[i],
				LBId:   remote.LoadBalancerAttribute.LoadBalancerId,
			})
		}
	}

	return actions, nil
}

func (m *ParallelizeModelApplier) cleanup(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	var actions []serverGroupAction
	// delete server groups
	for _, r := range remote.ServerGroups {
		found := false
		for _, l := range local.ServerGroups {
			if l.ServerGroupId == r.ServerGroupId {
				found = true
				break
			}
		}

		// delete unused vgroup
		if !found {
			// if the loadbalancer is preserved, and the service is deleting,
			// remove the server group tag instead of deleting the server group.
			if local.LoadBalancerAttribute.PreserveOnDelete {
				if err := m.sgMgr.CleanupServerGroupTags(reqCtx, r); err != nil {
					return err
				}
				continue
			}

			// do not delete user managed server group, but need to clean the backends
			if r.NamedKey == nil || r.IsUserManaged || !r.NamedKey.IsManagedByService(reqCtx.Service, base.CLUSTER_ID) {
				reqCtx.Log.Info(fmt.Sprintf("try to delete vgroup: [%s] description [%s] is managed by user, skip delete",
					r.ServerGroupName, r.ServerGroupId))
				var del []nlbmodel.ServerGroupServer
				for _, remote := range r.Servers {
					if !remote.IsUserManaged {
						del = append(del, remote)
					}
				}
				if len(del) > 0 {
					if err := m.sgMgr.BatchRemoveServers(reqCtx, r, del); err != nil {
						return err
					}
				}
				continue
			}

			reqCtx.Log.Info(fmt.Sprintf("delete server group [%s], %s", r.ServerGroupName, r.ServerGroupId))
			actions = append(actions, serverGroupAction{
				Action: serverGroupActionDelete,
				Remote: r,
			})
		}
	}

	errs := m.sgMgr.ParallelUpdateServerGroups(reqCtx, actions, nil)
	return utilerrors.NewAggregate(errs)
}

func isNLBReusable(service *v1.Service, tags []tag.Tag, dnsName string) (bool, string) {
	for _, t := range tags {
		// the tag of the apiserver slb is "ack.aliyun.com": "${clusterid}",
		// so can not reuse slbs which have ack.aliyun.com tag key.
		if t.Key == helper.TAGKEY || t.Key == util.ClusterTagKey {
			return false, "can not reuse loadbalancer created by kubernetes."
		}
	}

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		found := false
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.Hostname == dnsName {
				found = true
			}
		}
		if !found {
			return false, fmt.Sprintf("service has been associated with dnsname [%v], cannot be bound to dnsname [%s]",
				service.Status.LoadBalancer.Ingress[0].Hostname, dnsName)
		}
	}

	return true, ""
}

func findServerGroup(sgs []*nlbmodel.ServerGroup, lis *nlbmodel.ListenerAttribute) error {
	for _, sg := range sgs {
		if sg.ServerGroupName == lis.ServerGroupName {
			lis.ServerGroupId = sg.ServerGroupId
			return nil
		}
	}
	return fmt.Errorf("can not find server group by name %s", lis.ServerGroupName)

}

func isListenerPortMatch(l, r *nlbmodel.ListenerAttribute) bool {
	if l.ListenerPort != 0 {
		return l.ListenerPort == r.ListenerPort
	}
	return l.StartPort == r.StartPort && l.EndPort == r.EndPort
}
