package nlbv2

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/parallel"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/mohae/deepcopy"
	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	reconbackend "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DefaultServerWeight = 100

func NewServerGroupManager(kubeClient client.Client, cloud prvd.Provider) (*ServerGroupManager, error) {
	manager := &ServerGroupManager{
		kubeClient: kubeClient,
		cloud:      cloud,
	}

	vpcId, err := manager.cloud.VpcID()
	if err != nil {
		return nil, err
	}

	manager.vpcId = vpcId
	return manager, nil
}

type ServerGroupManager struct {
	kubeClient client.Client
	cloud      prvd.Provider
	vpcId      string
}

type serverGroupActionType string

const (
	serverGroupActionCreateAndAddBackendServers serverGroupActionType = "createAndAddBackendServers"
	serverGroupActionDelete                     serverGroupActionType = "delete"
	serverGroupActionUpdate                     serverGroupActionType = "update"
)

type serverGroupAction struct {
	Action serverGroupActionType
	Local  *nlbmodel.ServerGroup
	Remote *nlbmodel.ServerGroup
}

func (mgr *ServerGroupManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	var sgs []*nlbmodel.ServerGroup

	candidates, err := reconbackend.NewEndpointWithENI(reqCtx, mgr.kubeClient)
	if err != nil {
		return err
	}

	containsPotentialReadyBackends := false
	for _, lis := range mdl.Listeners {
		sg := &nlbmodel.ServerGroup{
			VPCId:       mgr.vpcId,
			ServicePort: lis.ServicePort,
			Tags:        getServerGroupTag(reqCtx),
			Protocol:    nlbmodel.GetListenerProtocolType(lis.ListenerProtocol),
		}
		if candidates.AddressIPVersion == model.IPv6 {
			sg.AddressIPVersion = string(model.DualStack)
		}

		if lis.ListenerPort != 0 {
			sg.NamedKey = getServerGroupNamedKey(reqCtx.Service, sg.Protocol, lis.ServicePort)
		} else {
			sg.AnyPortEnabled = true
			sg.HealthCheckConfig = &nlbmodel.HealthCheckConfig{
				HealthCheckEnabled:     tea.Bool(true),
				HealthCheckConnectPort: sg.ServicePort.TargetPort.IntVal,
			}
			sg.NamedKey = getAnyPortServerGroupNamedKey(reqCtx.Service, sg.Protocol, lis.StartPort, lis.EndPort)
		}
		sg.ServerGroupName = sg.NamedKey.Key()
		if err := setServerGroupAttributeFromAnno(sg, reqCtx.Anno); err != nil {
			return err
		}
		cpr, err := mgr.setServerGroupServers(reqCtx, sg, candidates, mdl.LoadBalancerAttribute.IsUserManaged)
		if err != nil {
			return fmt.Errorf("set ServerGroup for port %d error: %s", lis.ServicePort.Port, err.Error())
		}
		sgs = append(sgs, sg)
		containsPotentialReadyBackends = containsPotentialReadyBackends || cpr
	}

	err = mgr.updateServerGroupENIBackendID(reqCtx, sgs, candidates.AddressIPVersion)
	if err != nil {
		return err
	}

	mdl.ServerGroups = sgs
	mdl.ContainsPotentialReadyEndpoints = containsPotentialReadyBackends
	return nil
}

func (mgr *ServerGroupManager) updateServerGroupENIBackendID(reqCtx *svcCtx.RequestContext, sgs []*nlbmodel.ServerGroup, ipVersion model.AddressIPVersionType) error {
	eniIPs := sets.Set[string]{}
	for _, sg := range sgs {
		for _, s := range sg.Servers {
			if s.ServerType == nlbmodel.EniServerType {
				eniIPs.Insert(s.ServerIp)
			}
		}
	}

	ips := eniIPs.UnsortedList()
	if len(ips) == 0 {
		return nil
	}
	result, err := mgr.cloud.DescribeNetworkInterfaces(mgr.vpcId, ips, ipVersion)
	if err != nil {
		return fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
	}

	for _, sg := range sgs {
		for i := range sg.Servers {
			if sg.Servers[i].ServerType == nlbmodel.EniServerType {
				eniid, ok := result[sg.Servers[i].ServerIp]
				if !ok {
					return fmt.Errorf("can not find eniid for ip %s in vpc %s", sg.Servers[i].ServerIp, mgr.vpcId)
				}
				sg.Servers[i].ServerId = eniid
			}
		}
	}

	return nil
}

func (mgr *ServerGroupManager) ListNLBServerGroups(reqCtx *svcCtx.RequestContext) ([]*nlbmodel.ServerGroup, error) {
	sgs, err := mgr.cloud.ListNLBServerGroups(reqCtx.Ctx, getServerGroupTag(reqCtx))
	if err != nil {
		return sgs, err
	}

	reusedSgIDs, err := getServerGroupIDs(reqCtx.Anno.Get(annotation.VGroupPort))
	if err != nil {
		return sgs, err
	}

	for _, id := range reusedSgIDs {
		sg, err := mgr.cloud.GetNLBServerGroup(reqCtx.Ctx, id)
		if err != nil {
			return sgs, err
		}
		if sg == nil {
			return sgs, fmt.Errorf("reused server group id %s not found", id)
		}
		sg.IsUserManaged = true
		sgs = append(sgs, sg)
	}

	for i := range sgs {
		for j := range sgs[i].Servers {
			if !isServerManagedByMyService(reqCtx, sgs[i].Servers[j]) {
				sgs[i].Servers[j].IsUserManaged = true
				// if backend is managed by user, server group is also managed by user.
				sgs[i].IsUserManaged = true
				reqCtx.Log.Info(fmt.Sprintf("server group %s backend is managed by user: [%+v]",
					sgs[i].ServerGroupName, sgs[i].Servers[j]))
			}
		}
	}
	return sgs, nil
}

func (mgr *ServerGroupManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	sgs, err := mgr.ListNLBServerGroups(reqCtx)
	if err != nil {
		return fmt.Errorf("DescribeVServerGroups error: %s", err.Error())
	}
	mdl.ServerGroups = sgs
	return nil
}

func (mgr *ServerGroupManager) ParallelUpdateServerGroups(reqCtx *svcCtx.RequestContext, actions []serverGroupAction, sgChannel chan serverGroupApplyResult) []error {
	if len(actions) == 0 {
		reqCtx.Log.Info("no action to do for server group")
		return nil
	}
	errs := make([]error, len(actions))
	reqCtx.Log.V(5).Info("update server groups parallelly", "actionsCount", len(actions))
	parallel.DoPiece(reqCtx.Ctx, ctrlCfg.ControllerCFG.MaxConcurrentActions, len(actions), func(i int) {
		var err error
		act := actions[i]
		switch act.Action {
		case serverGroupActionCreateAndAddBackendServers:
			err = mgr.CreateServerGroupAndAddServers(reqCtx, act.Local, sgChannel)
		case serverGroupActionUpdate:
			err = mgr.UpdateServerGroup(reqCtx, act.Local, act.Remote)
		case serverGroupActionDelete:
			err = mgr.DeleteServerGroup(reqCtx, act.Remote.ServerGroupId)
		}
		errs[i] = err
	})
	return errs
}

func (mgr *ServerGroupManager) CreateServerGroupAndAddServers(reqCtx *svcCtx.RequestContext, local *nlbmodel.ServerGroup, sgChannel chan<- serverGroupApplyResult) error {
	err := mgr.CreateServerGroup(reqCtx, local)
	if sgChannel != nil {
		sgChannel <- serverGroupApplyResult{
			Err:             err,
			ServerGroupID:   local.ServerGroupId,
			ServerGroupName: local.ServerGroupName,
		}
	}
	if err != nil {
		return fmt.Errorf("EnsureServerGroupCreated error: %w", err)
	}
	if len(local.Servers) > 0 {
		err = mgr.BatchAddServers(reqCtx, local, local.Servers)
		if err != nil {
			return fmt.Errorf("BatchAddServers error: %w", err)
		}
	}

	return nil
}
func (mgr *ServerGroupManager) CreateServerGroup(reqCtx *svcCtx.RequestContext, sg *nlbmodel.ServerGroup) error {
	err := setDefaultValueForServerGroup(sg)
	if err != nil {
		return err
	}
	return mgr.cloud.CreateNLBServerGroup(reqCtx.Ctx, sg)
}

func (mgr *ServerGroupManager) DeleteServerGroup(reqCtx *svcCtx.RequestContext, sgId string) error {
	return mgr.cloud.DeleteNLBServerGroup(reqCtx.Ctx, sgId)
}

func (mgr *ServerGroupManager) UpdateServerGroup(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.ServerGroup) error {
	var errs []error
	// skip if server group is managed by user using "vgroup-port"
	if !local.IsUserManaged {
		update := deepcopy.Copy(remote).(*nlbmodel.ServerGroup)
		needUpdate := false
		updateDetail := ""

		if local.ServerGroupName != remote.ServerGroupName {
			needUpdate = true
			update.ServerGroupName = local.ServerGroupName
			updateDetail += fmt.Sprintf("ServerGroupName %v should be changed to %v;",
				remote.ServerGroupName, local.ServerGroupName)
		}
		if local.Scheduler != "" &&
			!strings.EqualFold(local.Scheduler, remote.Scheduler) {
			needUpdate = true
			update.Scheduler = local.Scheduler
			updateDetail += fmt.Sprintf("Scheduler %v should be changed to %v;",
				remote.Scheduler, local.Scheduler)
		}
		if local.ConnectionDrainEnabled != nil &&
			tea.BoolValue(local.ConnectionDrainEnabled) != tea.BoolValue(remote.ConnectionDrainEnabled) {
			needUpdate = true
			update.ConnectionDrainEnabled = local.ConnectionDrainEnabled
			updateDetail += fmt.Sprintf("ConnectionDrainEnabled %v should be changed to %v;",
				*remote.ConnectionDrainEnabled, *local.ConnectionDrainEnabled)
		}
		if local.ConnectionDrainTimeout != 0 &&
			local.ConnectionDrainTimeout != remote.ConnectionDrainTimeout {
			needUpdate = true
			update.ConnectionDrainTimeout = local.ConnectionDrainTimeout
			updateDetail += fmt.Sprintf("ConnectionDrainTimeout %v should be changed to %v;",
				remote.ConnectionDrainTimeout, local.ConnectionDrainTimeout)
		}
		if local.PreserveClientIpEnabled != nil &&
			tea.BoolValue(local.PreserveClientIpEnabled) != tea.BoolValue(remote.PreserveClientIpEnabled) {
			needUpdate = true
			update.PreserveClientIpEnabled = local.PreserveClientIpEnabled
			updateDetail += fmt.Sprintf("PreserveClientIpEnabled %v should be changed to %v;",
				tea.BoolValue(remote.PreserveClientIpEnabled), tea.BoolValue(local.PreserveClientIpEnabled))
		}
		// health check
		if local.HealthCheckConfig != nil {
			if remote.HealthCheckConfig == nil {
				needUpdate = true
				update.HealthCheckConfig = local.HealthCheckConfig
				updateDetail += fmt.Sprintf("HealthCheckConfig nil should be changed to %v;",
					*local.HealthCheckConfig)
				goto update
			}

			localHC, remoteHC := local.HealthCheckConfig, remote.HealthCheckConfig
			if localHC.HealthCheckEnabled != nil &&
				tea.BoolValue(localHC.HealthCheckEnabled) != tea.BoolValue(remoteHC.HealthCheckEnabled) {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckEnabled = localHC.HealthCheckEnabled
				updateDetail += fmt.Sprintf("HealthCheckEnabled %v should be changed to %v;",
					*remoteHC.HealthCheckEnabled, *localHC.HealthCheckEnabled)
			}
			if localHC.HealthCheckType != "" &&
				!strings.EqualFold(localHC.HealthCheckType, remoteHC.HealthCheckType) {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckType = localHC.HealthCheckType
				updateDetail += fmt.Sprintf("HealthCheckType %v should be changed to %v;",
					remoteHC.HealthCheckType, localHC.HealthCheckType)
			}
			if localHC.HealthCheckConnectPort != 0 &&
				localHC.HealthCheckConnectPort != remoteHC.HealthCheckConnectPort {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckConnectPort = localHC.HealthCheckConnectPort
				updateDetail += fmt.Sprintf("HealthCheckConnectPort %v should be changed to %v;",
					remoteHC.HealthCheckConnectPort, localHC.HealthCheckConnectPort)
			}
			if localHC.HealthyThreshold != 0 &&
				localHC.HealthyThreshold != remoteHC.HealthyThreshold {
				needUpdate = true
				update.HealthCheckConfig.HealthyThreshold = localHC.HealthyThreshold
				updateDetail += fmt.Sprintf("HealthyThreshold %v should be changed to %v;",
					remoteHC.HealthyThreshold, localHC.HealthyThreshold)
			}
			if localHC.UnhealthyThreshold != 0 &&
				localHC.UnhealthyThreshold != remoteHC.UnhealthyThreshold {
				needUpdate = true
				update.HealthCheckConfig.UnhealthyThreshold = localHC.UnhealthyThreshold
				updateDetail += fmt.Sprintf("UnhealthyThreshold %v should be changed to %v;",
					remoteHC.UnhealthyThreshold, localHC.UnhealthyThreshold)
			}
			if localHC.HealthCheckConnectTimeout != 0 &&
				localHC.HealthCheckConnectTimeout != remoteHC.HealthCheckConnectTimeout {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckConnectTimeout = localHC.HealthCheckConnectTimeout
				updateDetail += fmt.Sprintf("HealthCheckConnectTimeout %v should be changed to %v;",
					remoteHC.HealthCheckConnectTimeout, localHC.HealthCheckConnectTimeout)
			}
			if localHC.HealthCheckInterval != 0 &&
				localHC.HealthCheckInterval != remoteHC.HealthCheckInterval {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckInterval = localHC.HealthCheckInterval
				updateDetail += fmt.Sprintf("HealthCheckConnectTimeout %v should be changed to %v;",
					remoteHC.HealthCheckInterval, localHC.HealthCheckInterval)
			}
			if localHC.HealthCheckDomain != "" &&
				localHC.HealthCheckDomain != remoteHC.HealthCheckDomain {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckDomain = localHC.HealthCheckDomain
				updateDetail += fmt.Sprintf("HealthCheckDomain %v should be changed to %v;",
					remoteHC.HealthCheckDomain, localHC.HealthCheckDomain)
			}
			if localHC.HealthCheckUrl != "" &&
				localHC.HealthCheckUrl != remoteHC.HealthCheckUrl {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckUrl = localHC.HealthCheckUrl
				updateDetail += fmt.Sprintf("HealthCheckUrl %v should be changed to %v;",
					remoteHC.HealthCheckUrl, localHC.HealthCheckUrl)
			}
			if localHC.HttpCheckMethod != "" &&
				!strings.EqualFold(localHC.HttpCheckMethod, remoteHC.HttpCheckMethod) {
				needUpdate = true
				update.HealthCheckConfig.HttpCheckMethod = localHC.HttpCheckMethod
				updateDetail += fmt.Sprintf("HttpCheckMethod %v should be changed to %v;",
					remoteHC.HttpCheckMethod, localHC.HttpCheckMethod)
			}
			if len(localHC.HealthCheckHttpCode) != 0 &&
				!util.IsStringSliceEqual(localHC.HealthCheckHttpCode, remoteHC.HealthCheckHttpCode) {
				needUpdate = true
				update.HealthCheckConfig.HealthCheckHttpCode = localHC.HealthCheckHttpCode
				updateDetail += fmt.Sprintf("HealthCheckHttpCode %v should be changed to %v;",
					remoteHC.HealthCheckHttpCode, localHC.HealthCheckHttpCode)
			}

		}
	update:
		if needUpdate {
			reqCtx.Log.Info(fmt.Sprintf("update server group: %s [%s] changed, detail %s",
				local.ServerGroupId, local.ServerGroupName, updateDetail))
			if err := mgr.cloud.UpdateNLBServerGroup(reqCtx.Ctx, update); err != nil {
				errs = append(errs, fmt.Errorf("UpdateNLBServerGroup error: %s", err.Error()))
			}
		}
	} else {
		reqCtx.Log.Info(fmt.Sprintf("server group %s[%s] is user managed, skip update attribute.",
			remote.ServerGroupId, remote.ServerGroupName))
	}

	if err := mgr.updateServerGroupServers(reqCtx, local, remote); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

func (mgr *ServerGroupManager) updateServerGroupServers(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.ServerGroup) error {
	add, del, update := diff(remote, local)
	if len(add) == 0 && len(del) == 0 && len(update) == 0 {
		reqCtx.Log.Info(fmt.Sprintf("reconcile sg: [%s] not change, skip reconcile", remote.ServerGroupId),
			"sgName", remote.ServerGroupName)
		return nil
	}

	reqCtx.Log.Info(fmt.Sprintf("reconcile sg: [%s] changed, local: [%s], remote: [%s]",
		remote.ServerGroupId, local.BackendInfo(), remote.BackendInfo()), "sgName", remote.ServerGroupName)

	var errs []error

	if len(add) > 0 {
		if err := mgr.BatchAddServers(reqCtx, local, add); err != nil {
			errs = append(errs, err)
		}
	}
	if len(del) > 0 {
		if err := mgr.BatchRemoveServers(reqCtx, remote, del); err != nil {
			errs = append(errs, err)
		}
	}
	if len(update) > 0 {
		if err := mgr.BatchUpdateServers(reqCtx, remote, update); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (mgr *ServerGroupManager) BatchAddServers(reqCtx *svcCtx.RequestContext, sg *nlbmodel.ServerGroup,
	add []nlbmodel.ServerGroupServer) error {
	reqCtx.Log.Info(fmt.Sprintf("reconcile server group: [%s] backend add [%+v]", sg.ServerGroupId, add))
	return reconbackend.Batch(add, ctrlCfg.ControllerCFG.ServerGroupBatchSize,
		func(list []interface{}) error {
			var batchAdds []nlbmodel.ServerGroupServer
			for _, item := range list {
				item, _ := item.(nlbmodel.ServerGroupServer)
				batchAdds = append(batchAdds, item)
			}
			return mgr.cloud.AddNLBServers(reqCtx.Ctx, sg.ServerGroupId, batchAdds)
		})

}

func (mgr *ServerGroupManager) BatchRemoveServers(reqCtx *svcCtx.RequestContext, sg *nlbmodel.ServerGroup,
	del []nlbmodel.ServerGroupServer) error {
	reqCtx.Log.Info(fmt.Sprintf("reconcile server group: [%s] backend del [%+v]", sg.ServerGroupId, del))
	return reconbackend.Batch(del, ctrlCfg.ControllerCFG.ServerGroupBatchSize,
		func(list []interface{}) error {
			var batchDels []nlbmodel.ServerGroupServer
			for _, item := range list {
				item, _ := item.(nlbmodel.ServerGroupServer)
				batchDels = append(batchDels, item)
			}
			return mgr.cloud.RemoveNLBServers(reqCtx.Ctx, sg.ServerGroupId, batchDels)
		})

}

func (mgr *ServerGroupManager) BatchUpdateServers(reqCtx *svcCtx.RequestContext, sg *nlbmodel.ServerGroup,
	update []nlbmodel.ServerGroupServer) error {
	reqCtx.Log.Info(fmt.Sprintf("reconcile server group: [%s] backend update [%+v]", sg.ServerGroupId, update))
	return reconbackend.Batch(update, ctrlCfg.ControllerCFG.ServerGroupBatchSize,
		func(list []interface{}) error {
			var batchUpdates []nlbmodel.ServerGroupServer
			for _, item := range list {
				item, _ := item.(nlbmodel.ServerGroupServer)
				batchUpdates = append(batchUpdates, item)
			}
			return mgr.cloud.UpdateNLBServers(reqCtx.Ctx, sg.ServerGroupId, batchUpdates)
		})
}

func (mgr *ServerGroupManager) setServerGroupServers(reqCtx *svcCtx.RequestContext, sg *nlbmodel.ServerGroup, candidates *reconbackend.EndpointWithENI, isUserManagedLB bool) (bool, error) {
	var (
		backends []nlbmodel.ServerGroupServer
		err      error
	)

	if isUserManagedLB && reqCtx.Anno.Get(annotation.VGroupPort) != "" {
		sgId, err := serverGroup(reqCtx.Anno.Get(annotation.VGroupPort), *sg.ServicePort)

		if err != nil {
			return false, fmt.Errorf("server group id parse error: %s", err.Error())
		}
		if sgId != "" {
			remoteSg, err := mgr.cloud.GetNLBServerGroup(reqCtx.Ctx, sgId)
			if err != nil {
				return false, fmt.Errorf("find server group id %s for nlb %s error: %s", sgId, reqCtx.Anno.Get(annotation.LoadBalancerId), err.Error())
			}
			if remoteSg == nil {
				return false, fmt.Errorf("cannot find server group id %s for nlb %s", sgId, reqCtx.Anno.Get(annotation.LoadBalancerId))
			}
			if !remoteSg.IsUserManaged {
				return false, fmt.Errorf("cannot reuse a ccm created server group %s for nlb %s", sgId, reqCtx.Anno.Get(annotation.LoadBalancerId))
			}
			reqCtx.Log.Info(fmt.Sprintf("user managed server group id %s for port %d", sgId, sg.ServicePort.Port))
			sg.ServerGroupId = sgId
			sg.IsUserManaged = true
		}

		if reqCtx.Anno.Get(annotation.VGroupWeight) != "" {
			w, err := strconv.Atoi(reqCtx.Anno.Get(annotation.VGroupWeight))
			if err != nil || w < 0 || w > 100 {
				return false, fmt.Errorf("weight must be integer in range [0, 100], got [%s]",
					reqCtx.Anno.Get(annotation.VGroupWeight))
			}
			sg.Weight = &w
		}
	}

	initialServers, containsPotentialReadyEndpoints, err := mgr.setGenericBackendAttribute(reqCtx, candidates, *sg)
	if err != nil {
		return false, err
	}

	switch candidates.TrafficPolicy {
	case helper.ENITrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("eni mode, build backends for %s", sg.NamedKey))
		backends, err = mgr.buildENIBackends(reqCtx, candidates, initialServers, *sg)
		if err != nil {
			return false, fmt.Errorf("build eni backends error: %s", err.Error())
		}
	case helper.LocalTrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("local mode, build backends for %s", sg.NamedKey))
		backends, err = mgr.buildLocalBackends(reqCtx, candidates, initialServers, *sg)
		if err != nil {
			return false, fmt.Errorf("build local backends error: %s", err.Error())
		}
	case helper.ClusterTrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("cluster mode, build backends for %s", sg.NamedKey))
		backends, err = mgr.buildClusterBackends(reqCtx, candidates, initialServers, *sg)
		if err != nil {
			return false, fmt.Errorf("build cluster backends error: %s", err.Error())
		}
	default:
		return false, fmt.Errorf("not supported traffic policy [%s]", candidates.TrafficPolicy)
	}

	if len(backends) == 0 {
		reqCtx.Recorder.Event(
			reqCtx.Service,
			v1.EventTypeNormal,
			helper.UnAvailableBackends,
			"There are no available nodes for NetworkLoadBalancer",
		)
	}

	sg.Servers = backends
	sg.InitialServers = initialServers
	return containsPotentialReadyEndpoints, nil
}

func (mgr *ServerGroupManager) setGenericBackendAttribute(reqCtx *svcCtx.RequestContext, candidates *reconbackend.EndpointWithENI, sg nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, bool, error) {
	if utilfeature.DefaultMutableFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		return mgr.setBackendsFromEndpointSlices(reqCtx, candidates, sg)
	}
	return mgr.setBackendsFromEndpoints(reqCtx, candidates, sg)
}

func (mgr *ServerGroupManager) setBackendsFromEndpoints(reqCtx *svcCtx.RequestContext, candidates *reconbackend.EndpointWithENI, sg nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, bool, error) {
	var backends []nlbmodel.ServerGroupServer

	readinessGateName := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, reqCtx.Service.Name)
	containsPotentialReadyEndpoints := false

	if len(candidates.Endpoints.Subsets) == 0 {
		return nil, false, nil
	}
	for _, ep := range candidates.Endpoints.Subsets {
		var backendPort int32
		if sg.ServicePort.TargetPort.Type == intstr.Int {
			backendPort = int32(sg.ServicePort.TargetPort.IntValue())
		} else {
			for _, p := range ep.Ports {
				if p.Name == sg.ServicePort.Name {
					backendPort = p.Port
					break
				}
			}
			if backendPort == 0 {
				klog.Warningf("%s cannot find port according port name: %s", sg.ServerGroupName, sg.ServicePort.Name)
				continue
			}
		}

		for _, addr := range ep.Addresses {
			backends = append(backends, nlbmodel.ServerGroupServer{
				NodeName: addr.NodeName,
				ServerIp: addr.IP,
				// set backend port to targetPort by default
				// if backend type is ecs, update backend port to nodePort
				// if server group type is anyport, set port to 0
				Port:        getBackendPort(backendPort, sg.AnyPortEnabled),
				Description: sg.ServerGroupName,
				TargetRef:   addr.TargetRef,
			})
		}

		for _, addr := range ep.NotReadyAddresses {
			// ignore endpoint that not referenced to a pod
			if addr.TargetRef == nil || addr.TargetRef.Kind != "Pod" {
				continue
			}

			pod := &v1.Pod{}
			podKey := types.NamespacedName{Namespace: addr.TargetRef.Namespace, Name: addr.TargetRef.Name}
			err := mgr.kubeClient.Get(reqCtx.Ctx, podKey, pod)
			if err != nil {
				if errors.IsNotFound(err) {
					// pod may be not synced to the controller cache, which caused a not found error.
					// ignore it and set containsPotentialReadyEndpoints to true
					reqCtx.Log.Info("pod not found, will retry it in next reconcile", "pod", podKey.String())
					containsPotentialReadyEndpoints = true
					continue
				}
				return nil, false, err
			}

			if !helper.IsPodHasReadinessGate(pod, string(readinessGateName)) {
				continue
			}

			if !helper.IsPodContainersReady(pod) {
				containsPotentialReadyEndpoints = true
				continue
			}

			backends = append(backends, nlbmodel.ServerGroupServer{
				NodeName: addr.NodeName,
				ServerIp: addr.IP,
				// set backend port to targetPort by default
				// if backend type is ecs, update backend port to nodePort
				// if server group type is anyport, set port to 0
				Port:        getBackendPort(backendPort, sg.AnyPortEnabled),
				Description: sg.ServerGroupName,
				TargetRef:   addr.TargetRef,
			})
		}
	}
	return backends, containsPotentialReadyEndpoints, nil
}

func (mgr *ServerGroupManager) setBackendsFromEndpointSlices(reqCtx *svcCtx.RequestContext, candidates *reconbackend.EndpointWithENI, sg nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, bool, error) {
	// used for deduplicate when endpointslice is enabled
	// https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/#duplicate-endpoints
	endpointMap := make(map[string]bool)
	var backends []nlbmodel.ServerGroupServer

	containsPotentialReadyEndpoints := false
	readinessGateName := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, reqCtx.Service.Name)

	if len(candidates.EndpointSlices) == 0 {
		return nil, false, nil
	}

	for _, es := range candidates.EndpointSlices {
		var backendPort int32
		if sg.ServicePort.TargetPort.Type == intstr.Int {
			backendPort = int32(sg.ServicePort.TargetPort.IntValue())
		} else {
			for _, p := range es.Ports {
				// be compatible with IntOrString type target port
				if p.Name != nil && *p.Name == sg.ServicePort.Name {
					if p.Port != nil {
						backendPort = *p.Port
					}
					break
				}
			}
			if backendPort == 0 {
				klog.Warningf("%s cannot find port according port name: %s", sg.ServerGroupName, sg.ServicePort.Name)
				continue
			}
		}

		for _, ep := range es.Endpoints {
			// ignore terminating pods
			if ep.Conditions.Terminating != nil && *ep.Conditions.Terminating {
				continue
			}

			for _, addr := range ep.Addresses {
				if _, ok := endpointMap[addr]; ok {
					continue
				}

				if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
					endpointMap[addr] = true
					backends = append(backends, nlbmodel.ServerGroupServer{
						NodeName: ep.NodeName,
						ServerIp: addr,
						// set backend port to targetPort by default
						// if backend type is ecs, update backend port to nodePort
						// if server group type is anyport, set port to 0
						Port:        getBackendPort(backendPort, sg.AnyPortEnabled),
						Description: sg.ServerGroupName,
						TargetRef:   ep.TargetRef,
					})
					continue
				}

				// ignore endpoint that not referenced to a pod
				if ep.TargetRef == nil || ep.TargetRef.Kind != "Pod" {
					continue
				}

				pod := &v1.Pod{}
				podKey := types.NamespacedName{Namespace: ep.TargetRef.Namespace, Name: ep.TargetRef.Name}
				err := mgr.kubeClient.Get(reqCtx.Ctx, podKey, pod)
				if err != nil {
					if errors.IsNotFound(err) {
						// pod may be not synced to the controller cache, which caused a not found error.
						// ignore it and set containsPotentialReadyEndpoints to true
						reqCtx.Log.Info("pod not found, will retry it in next reconcile", "pod", podKey.String())
						containsPotentialReadyEndpoints = true
						continue
					}
					return nil, false, err
				}

				klog.Errorf("readiness gate name: %s, try find it on %s", readinessGateName, pod.Name)

				if !helper.IsPodHasReadinessGate(pod, string(readinessGateName)) {
					continue
				}

				if !helper.IsPodContainersReady(pod) {
					containsPotentialReadyEndpoints = true
					continue
				}

				// TODO: may need to check node status
				endpointMap[addr] = true
				backends = append(backends, nlbmodel.ServerGroupServer{
					NodeName: ep.NodeName,
					ServerIp: addr,
					// set backend port to targetPort by default
					// if backend type is ecs, update backend port to nodePort
					// if server group type is anyport, set port to 0
					Port:        getBackendPort(backendPort, sg.AnyPortEnabled),
					Description: sg.ServerGroupName,
					TargetRef:   ep.TargetRef,
				})
			}
		}
	}

	return backends, containsPotentialReadyEndpoints, nil
}

func getBackendPort(port int32, anyPort bool) int32 {
	if anyPort {
		return 0
	}
	return port
}

func (mgr *ServerGroupManager) buildENIBackends(reqCtx *svcCtx.RequestContext, candidates *reconbackend.EndpointWithENI, backends []nlbmodel.ServerGroupServer, sg nlbmodel.ServerGroup,
) ([]nlbmodel.ServerGroupServer, error) {
	if len(backends) == 0 {
		return nil, nil
	}

	backends, err := updateENIBackends(mgr, backends, candidates.AddressIPVersion, sg.ServerGroupType)
	if err != nil {
		return backends, err
	}

	return setWeightBackends(helper.ENITrafficPolicy, backends, sg.Weight), nil
}

func (mgr *ServerGroupManager) buildLocalBackends(reqCtx *svcCtx.RequestContext, candidates *reconbackend.EndpointWithENI, initBackends []nlbmodel.ServerGroupServer,
	sg nlbmodel.ServerGroup) ([]nlbmodel.ServerGroupServer, error) {
	if len(initBackends) == 0 {
		return nil, nil
	}

	var (
		ecsBackends, eciBackends []nlbmodel.ServerGroupServer
		err                      error
	)

	// filter ecs backends and eci backends
	// 1. add ecs backends. add pod located nodes.
	// Attention: will add duplicated ecs backends.
	for _, backend := range initBackends {
		if backend.NodeName == nil {
			return nil, fmt.Errorf("add ecs backends for service[%s] error, NodeName is nil for ip %s ",
				util.Key(reqCtx.Service), backend.ServerIp)
		}
		node := helper.FindNodeByNodeName(candidates.Nodes, *backend.NodeName)
		if node == nil {
			reqCtx.Log.Info(fmt.Sprintf("warning: can not find correspond node %s for endpoint %s", *backend.NodeName, backend.ServerIp))
			continue
		}

		// check if the node is virtual node, virtual node add as eci backend
		if node.Labels["type"] == helper.LabelNodeTypeVK {
			eciBackends = append(eciBackends, backend)
			continue
		}

		if helper.IsNodeExcludeFromLoadBalancer(node) {
			reqCtx.Log.Info("node has exclude label or has type which cannot be added to lb backend", "node", node.Name)
			continue
		}

		if sg.ServerGroupType == nlbmodel.IpServerGroupType {
			ip, err := helper.GetNodeInternalIP(node)
			if err != nil {
				return nil, fmt.Errorf("get node address err: %s", err.Error())
			}
			backend.ServerId = ip
			backend.ServerIp = ip
			backend.ServerType = nlbmodel.IpServerType
		} else {
			_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
			if err != nil {
				return nil, fmt.Errorf("parse providerid: %s. "+
					"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
			}
			backend.ServerId = id
			backend.ServerIp = ""
			backend.ServerType = nlbmodel.EcsServerType
		}

		// for ECS backend type, port should be set to NodePort
		backend.Port = sg.ServicePort.NodePort
		ecsBackends = append(ecsBackends, backend)
	}

	// 2. add eci backends
	if len(eciBackends) != 0 {
		reqCtx.Log.Info("add eciBackends")
		eciBackends, err = updateENIBackends(mgr, eciBackends, candidates.AddressIPVersion, sg.ServerGroupType)
		if err != nil {
			return nil, fmt.Errorf("update eci backends error: %s", err.Error())
		}
	}

	backends := append(ecsBackends, eciBackends...)

	// 3. set weight
	backends = setWeightBackends(helper.LocalTrafficPolicy, backends, sg.Weight)

	// 4. remove duplicated ecs
	return removeDuplicatedECS(backends), nil
}

func removeDuplicatedECS(backends []nlbmodel.ServerGroupServer) []nlbmodel.ServerGroupServer {
	nodeMap := make(map[string]bool)
	var uniqBackends []nlbmodel.ServerGroupServer
	for _, backend := range backends {
		if backend.ServerType != nlbmodel.EcsServerType {
			uniqBackends = append(uniqBackends, backend)
			continue
		}
		if _, ok := nodeMap[backend.ServerId]; ok {
			continue
		}
		nodeMap[backend.ServerId] = true
		uniqBackends = append(uniqBackends, backend)
	}
	return uniqBackends

}

func (mgr *ServerGroupManager) buildClusterBackends(
	reqCtx *svcCtx.RequestContext, candidates *reconbackend.EndpointWithENI, initBackends []nlbmodel.ServerGroupServer, sg nlbmodel.ServerGroup,
) ([]nlbmodel.ServerGroupServer, error) {
	var (
		ecsBackends, eciBackends []nlbmodel.ServerGroupServer
		err                      error
	)

	// 1. add ecs backends. add all cluster nodes.
	for _, node := range candidates.Nodes {
		if helper.IsNodeExcludeFromLoadBalancer(&node) {
			reqCtx.Log.Info("node has exclude label or has type which cannot be added to lb backend", "node", node.Name)
			continue
		}

		backend := nlbmodel.ServerGroupServer{
			Weight:      DefaultServerWeight,
			Port:        sg.ServicePort.NodePort,
			Description: sg.ServerGroupName,
		}

		if sg.ServerGroupType == nlbmodel.IpServerGroupType {
			ip, err := helper.GetNodeInternalIP(&node)
			if err != nil {
				return nil, fmt.Errorf("get node address err: %s", err.Error())
			}
			backend.ServerId = ip
			backend.ServerIp = ip
			backend.ServerType = nlbmodel.IpServerType
		} else {
			_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
			if err != nil {
				return nil, fmt.Errorf("normal parse providerid: %s. "+
					"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
			}
			backend.ServerId = id
			backend.ServerIp = ""
			backend.ServerType = nlbmodel.EcsServerType
		}

		ecsBackends = append(ecsBackends, backend)
	}

	// 2. add eci backends
	for _, b := range initBackends {
		if b.NodeName == nil {
			return nil, fmt.Errorf("add ecs backends for service[%s] error, NodeName is nil for ip %s ",
				util.Key(reqCtx.Service), b.ServerIp)
		}
		node := helper.FindNodeByNodeName(candidates.Nodes, *b.NodeName)
		if node == nil {
			reqCtx.Log.Info(fmt.Sprintf("warning: can not find correspond node %s for endpoint %s",
				*b.NodeName, b.ServerIp))
			continue
		}

		// check if the node is VK
		if node.Labels["type"] == helper.LabelNodeTypeVK {
			eciBackends = append(eciBackends, b)
			continue
		}
	}

	if len(eciBackends) != 0 {
		eciBackends, err = updateENIBackends(mgr, eciBackends, candidates.AddressIPVersion, sg.ServerGroupType)
		if err != nil {
			return nil, fmt.Errorf("update eci backends error: %s", err.Error())
		}
	}

	backends := append(ecsBackends, eciBackends...)

	return setWeightBackends(helper.ClusterTrafficPolicy, backends, sg.Weight), nil
}

func (mgr *ServerGroupManager) CleanupServerGroupTags(reqCtx *svcCtx.RequestContext, r *nlbmodel.ServerGroup) error {
	defaultTags := reqCtx.Anno.GetDefaultTags()

	var deletedTags []*string
	for _, r := range r.Tags {
		for _, l := range defaultTags {
			if l.Key == r.Key && l.Value == r.Value {
				deletedTags = append(deletedTags, tea.String(r.Key))
				break
			}
		}
	}

	if len(deletedTags) == 0 {
		return nil
	}

	return mgr.cloud.UntagNLBResources(reqCtx.Ctx, r.ServerGroupId, nlbmodel.ServerGroupTagType, deletedTags)
}

func updateENIBackends(mgr *ServerGroupManager, backends []nlbmodel.ServerGroupServer,
	ipVersion model.AddressIPVersionType, serverGroupType nlbmodel.ServerGroupType,
) ([]nlbmodel.ServerGroupServer, error) {
	if serverGroupType == nlbmodel.IpServerGroupType {
		for i := range backends {
			backends[i].ServerId = backends[i].ServerIp
			backends[i].ServerType = nlbmodel.IpServerType
		}
		return backends, nil
	}

	for i := range backends {
		// for ENI backend type, port should be set to targetPort (default value), no need to update
		backends[i].ServerType = nlbmodel.EniServerType
	}
	return backends, nil
}

func setWeightBackends(mode helper.TrafficPolicy, backends []nlbmodel.ServerGroupServer, weight *int) []nlbmodel.ServerGroupServer {
	// use default
	if weight == nil {
		return podNumberAlgorithm(mode, backends)
	}

	return podPercentAlgorithm(mode, backends, *weight)

}

// weight algorithm
// podNumberAlgorithm (default algorithm)
/*
	Calculate node weight by pod.
	ClusterMode:  nodeWeight = 1
	ENIMode:      podWeight = 1
	LocalMode:    node_weight = nodePodNum
*/
func podNumberAlgorithm(mode helper.TrafficPolicy, backends []nlbmodel.ServerGroupServer) []nlbmodel.ServerGroupServer {
	if mode == helper.ENITrafficPolicy || mode == helper.ClusterTrafficPolicy {
		for i := range backends {
			backends[i].Weight = DefaultServerWeight
		}
		return backends
	}

	// LocalTrafficPolicy
	ecsPods := make(map[string]int32)
	for _, b := range backends {
		if b.ServerId != "" {
			ecsPods[b.ServerId] += 1
		}
	}
	for i := range backends {
		if backends[i].ServerType != nlbmodel.EcsServerType {
			backends[i].Weight = 1
		} else {
			backends[i].Weight = ecsPods[backends[i].ServerId]
		}
	}
	return backends
}

// podPercentAlgorithm
/*
	Calculate node weight by percent.
	ClusterMode:  node_weight = weightSum/nodesNum
	ENIMode:      pod_weight = weightSum/podsNum
	LocalMode:    node_weight = node_pod_num/pods_num *weightSum
*/
func podPercentAlgorithm(mode helper.TrafficPolicy, backends []nlbmodel.ServerGroupServer, weight int,
) []nlbmodel.ServerGroupServer {
	if len(backends) == 0 {
		return backends
	}

	if weight == 0 {
		for i := range backends {
			backends[i].Weight = 0
		}
		return backends
	}

	if mode == helper.ENITrafficPolicy || mode == helper.ClusterTrafficPolicy {
		per := weight / len(backends)
		if per < 1 {
			per = 1
		}

		for i := range backends {
			backends[i].Weight = int32(per)
		}
		return backends
	}

	// LocalTrafficPolicy
	ecsPods := make(map[string]int)
	for _, b := range backends {
		if b.ServerId != "" {
			ecsPods[b.ServerId] += 1
		}
	}
	for i := range backends {
		if backends[i].ServerType != nlbmodel.EcsServerType {
			backends[i].Weight = int32(weight * 1 / len(backends))
		} else {
			backends[i].Weight = int32(weight * ecsPods[backends[i].ServerId] / len(backends))
		}
		if backends[i].Weight < 1 {
			backends[i].Weight = 1
		}
	}
	return backends
}

func getServerGroupNamedKey(svc *v1.Service, protocol string, servicePort *v1.ServicePort) *nlbmodel.SGNamedKey {
	sgPort := ""
	if helper.IsENIBackendType(svc) {
		switch servicePort.TargetPort.Type {
		case intstr.Int:
			sgPort = fmt.Sprintf("%d", servicePort.TargetPort.IntValue())
		case intstr.String:
			sgPort = servicePort.TargetPort.StrVal
		}
	} else {
		sgPort = fmt.Sprintf("%d", servicePort.NodePort)
	}
	return &nlbmodel.SGNamedKey{
		NamedKey: nlbmodel.NamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			Namespace:   svc.Namespace,
			CID:         base.CLUSTER_ID,
			ServiceName: svc.Name,
		},
		Protocol:    protocol,
		SGGroupPort: sgPort}
}

func getAnyPortServerGroupNamedKey(svc *v1.Service, protocol string, startPort, endPort int32) *nlbmodel.SGNamedKey {
	return &nlbmodel.SGNamedKey{
		NamedKey: nlbmodel.NamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			Namespace:   svc.Namespace,
			CID:         base.CLUSTER_ID,
			ServiceName: svc.Name,
		},
		Protocol:    protocol,
		SGGroupPort: fmt.Sprintf("%d_%d", startPort, endPort),
	}
}

func getServerGroupTag(reqCtx *svcCtx.RequestContext) []tag.Tag {
	return []tag.Tag{
		{
			Key:   helper.TAGKEY,
			Value: reqCtx.Anno.GetDefaultLoadBalancerName(),
		},
		{
			Key:   util.ClusterTagKey,
			Value: base.CLUSTER_ID,
		},
	}
}

func setServerGroupAttributeFromAnno(sg *nlbmodel.ServerGroup, anno *annotation.AnnotationRequest) error {
	if anno.Get(annotation.ServerGroupType) != "" {
		if strings.EqualFold(anno.Get(annotation.ServerGroupType), string(nlbmodel.IpServerGroupType)) {
			sg.ServerGroupType = nlbmodel.IpServerGroupType
		} else if strings.EqualFold(anno.Get(annotation.ServerGroupType), string(nlbmodel.InstanceServerGroupType)) {
			sg.ServerGroupType = nlbmodel.InstanceServerGroupType
		} else {
			return fmt.Errorf("unsupport server ServerGroupType [%s]", sg.ServerGroupType)
		}
	}

	if anno.Get(annotation.ConnectionDrain) != "" {
		sg.ConnectionDrainEnabled = tea.Bool(strings.EqualFold(anno.Get(annotation.ConnectionDrain), string(model.OnFlag)))
	}

	if anno.Get(annotation.ConnectionDrainTimeout) != "" {
		timeout, err := strconv.Atoi(anno.Get(annotation.ConnectionDrainTimeout))
		if err != nil {
			return fmt.Errorf("ConnectionDrainTimeout parse error: %s", err.Error())
		}
		sg.ConnectionDrainTimeout = int32(timeout)
	}

	sg.Scheduler = anno.Get(annotation.Scheduler)

	if anno.Get(annotation.PreserveClientIp) != "" {
		sg.PreserveClientIpEnabled = tea.Bool(
			strings.EqualFold(anno.Get(annotation.PreserveClientIp), string(model.OnFlag)))
	}

	if rgID := anno.Get(annotation.ResourceGroupId); rgID != "" {
		sg.ResourceGroupId = rgID
	}

	// healthcheck
	if anno.Get(annotation.HealthCheckFlag) != "" {
		var healthCheckConfig *nlbmodel.HealthCheckConfig
		if sg.HealthCheckConfig != nil {
			healthCheckConfig = sg.HealthCheckConfig
		} else {
			healthCheckConfig = &nlbmodel.HealthCheckConfig{}
		}
		healthCheckConfig.HealthCheckEnabled = tea.Bool(strings.EqualFold(anno.Get(annotation.HealthCheckFlag), string(model.OnFlag)))

		if *healthCheckConfig.HealthCheckEnabled {
			if anno.Get(annotation.HealthCheckType) != "" {
				healthCheckConfig.HealthCheckType = anno.Get(annotation.HealthCheckType)
			}
			if anno.Get(annotation.HealthCheckConnectPort) != "" {
				checkPort, err := strconv.Atoi(anno.Get(annotation.HealthCheckConnectPort))
				if err != nil {
					return fmt.Errorf("HealthCheckConnectPort parse error: [%s]", err.Error())
				}
				healthCheckConfig.HealthCheckConnectPort = int32(checkPort)
			}
			if anno.Get(annotation.HealthyThreshold) != "" {
				healthyThreshold, err := strconv.Atoi(anno.Get(annotation.HealthyThreshold))
				if err != nil {
					return fmt.Errorf("HealthyThreshold parse error: [%s]", err.Error())
				}
				healthCheckConfig.HealthyThreshold = int32(healthyThreshold)
			}
			if anno.Get(annotation.UnhealthyThreshold) != "" {
				unhealthyThreshold, err := strconv.Atoi(anno.Get(annotation.UnhealthyThreshold))
				if err != nil {
					return fmt.Errorf("UnhealthyThreshold parse error: [%s]", err.Error())
				}
				healthCheckConfig.UnhealthyThreshold = int32(unhealthyThreshold)
			}
			if anno.Get(annotation.HealthCheckConnectTimeout) != "" {
				healthCheckConnectTimeout, err := strconv.Atoi(anno.Get(annotation.HealthCheckConnectTimeout))
				if err != nil {
					return fmt.Errorf("HealthCheckConnectTimeout parse error: [%s]", err.Error())
				}
				healthCheckConfig.HealthCheckConnectTimeout = int32(healthCheckConnectTimeout)
			}
			if anno.Get(annotation.HealthCheckInterval) != "" {
				healthCheckInterval, err := strconv.Atoi(anno.Get(annotation.HealthCheckInterval))
				if err != nil {
					return fmt.Errorf("HealthCheckInterval parse error: [%s]", err.Error())
				}
				healthCheckConfig.HealthCheckInterval = int32(healthCheckInterval)
			}
			healthCheckConfig.HealthCheckDomain = anno.Get(annotation.HealthCheckDomain)
			healthCheckConfig.HealthCheckUrl = anno.Get(annotation.HealthCheckURI)
			healthCheckConfig.HttpCheckMethod = anno.Get(annotation.HealthCheckMethod)
			if anno.Get(annotation.HealthCheckHTTPCode) != "" {
				healthCheckConfig.HealthCheckHttpCode = strings.Split(anno.Get(annotation.HealthCheckHTTPCode), ",")
			}
		}
		sg.HealthCheckConfig = healthCheckConfig
	}

	if strings.EqualFold(anno.Get(annotation.IgnoreWeightUpdate), string(model.OnFlag)) {
		sg.IgnoreWeightUpdate = true
	}
	return nil
}

func diff(remote, local *nlbmodel.ServerGroup) (
	[]nlbmodel.ServerGroupServer, []nlbmodel.ServerGroupServer, []nlbmodel.ServerGroupServer) {

	var (
		additions []nlbmodel.ServerGroupServer
		deletions []nlbmodel.ServerGroupServer
		updates   []nlbmodel.ServerGroupServer
	)

	for _, r := range remote.Servers {
		if r.IsUserManaged {
			continue
		}
		found := false
		for _, l := range local.Servers {
			if isServerEqual(r, l) {
				found = true
			}
		}
		if !found {
			deletions = append(deletions, r)
		}
	}

	for _, l := range local.Servers {
		found := false
		for _, r := range remote.Servers {
			if isServerEqual(l, r) {
				found = true
			}
		}
		if !found {
			additions = append(additions, l)
		}
	}

	for _, l := range local.Servers {
		for _, r := range remote.Servers {
			if isServerEqual(l, r) {
				if l.Port != r.Port || (!local.IgnoreWeightUpdate && l.Weight != r.Weight) || l.Description != r.Description {
					updates = append(updates, l)
				}
			}
		}
	}

	return additions, deletions, updates

}

func isServerEqual(a, b nlbmodel.ServerGroupServer) bool {
	if a.ServerType != b.ServerType {
		return false
	}

	switch a.ServerType {
	case nlbmodel.EniServerType:
		return a.ServerId == b.ServerId && a.ServerIp == b.ServerIp
	case nlbmodel.EcsServerType:
		return a.ServerId == b.ServerId
	case nlbmodel.IpServerType:
		return a.ServerId == b.ServerId && a.ServerIp == b.ServerIp
	default:
		klog.Errorf("%s is not supported, skip", a.ServerType)
		return false
	}
}

func setDefaultValueForServerGroup(sg *nlbmodel.ServerGroup) error {
	if sg.ResourceGroupId == "" {
		sg.ResourceGroupId = ctrlCfg.CloudCFG.Global.ResourceGroupID
	}
	return nil
}

func getServerGroupIDs(annotation string) ([]string, error) {
	if annotation == "" {
		return nil, nil
	}
	var ids []string
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return nil, fmt.Errorf("server group id and protocol format must be like"+
				"'sgp-xxx:443' with colon separated. got=[%+v]", pp)
		}
		ids = append(ids, pp[0])
	}
	return ids, nil
}

func isServerManagedByMyService(reqCtx *svcCtx.RequestContext, remote nlbmodel.ServerGroupServer) bool {
	namedKey, err := nlbmodel.LoadNLBSGNamedKey(remote.Description)
	if err != nil {
		return false
	}
	return namedKey.ServiceName == reqCtx.Service.Name &&
		namedKey.Namespace == reqCtx.Service.Namespace &&
		namedKey.CID == base.CLUSTER_ID
}
