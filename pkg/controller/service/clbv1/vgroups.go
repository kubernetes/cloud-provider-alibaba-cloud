package clbv1

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/parallel"
	"k8s.io/klog/v2"
	"net"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DefaultServerWeight = 100

func NewVGroupManager(kubeClient client.Client, cloud prvd.Provider) (*VGroupManager, error) {
	vpcId, err := cloud.VpcID()
	if err != nil {
		return nil, err
	}

	return &VGroupManager{
		kubeClient: kubeClient,
		cloud:      cloud,
		vpcId:      vpcId,
	}, nil
}

type VGroupManager struct {
	kubeClient client.Client
	cloud      prvd.Provider
	vpcId      string
}

type vGroupActionType string

const (
	vGroupActionCreateAndAddBackendServers vGroupActionType = "create_and_add"
	vGroupActionDelete                     vGroupActionType = "delete"
	vGroupActionUpdate                     vGroupActionType = "update"
)

type vGroupAction struct {
	Action vGroupActionType
	Local  *model.VServerGroup
	Remote *model.VServerGroup
	LBId   string
}

func (mgr *VGroupManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, m *model.LoadBalancer) error {
	candidates, err := backend.NewEndpointWithENI(reqCtx, mgr.kubeClient)
	if err != nil {
		return err
	}

	var vgs []model.VServerGroup
	errs := make([]error, len(reqCtx.Service.Spec.Ports))
	containsPotentialReadyEndpoints := false
	for _, port := range reqCtx.Service.Spec.Ports {
		vg, cpr, err := mgr.buildVGroupForServicePort(reqCtx, port, candidates, m.LoadBalancerAttribute.IsUserManaged)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		vgs = append(vgs, vg)
		containsPotentialReadyEndpoints = containsPotentialReadyEndpoints || cpr
	}

	err = mgr.updateVServerGroupENIBackendID(reqCtx, vgs, candidates.AddressIPVersion)
	if err != nil {
		return err
	}

	m.VServerGroups = vgs
	m.ContainsPotentialReadyEndpoints = containsPotentialReadyEndpoints

	return nil
}

func (mgr *VGroupManager) updateVServerGroupENIBackendID(reqCtx *svcCtx.RequestContext, vgs []model.VServerGroup, ipVersion model.AddressIPVersionType) error {
	eniIPs := sets.Set[string]{}
	for _, vg := range vgs {
		for _, b := range vg.Backends {
			if b.Type == model.ENIBackendType {
				eniIPs.Insert(b.ServerIp)
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

	var vpcCIDRs []*net.IPNet
	for i := range vgs {
		var filteredBackends []model.BackendAttribute
		var skipIPs []string
		for _, b := range vgs[i].Backends {
			if b.Type == model.ENIBackendType {
				eniid, ok := result[b.ServerIp]
				if !ok {
					if vpcCIDRs == nil {
						vpcCIDRs, err = mgr.cloud.DescribeVpcCIDRBlock(reqCtx.Ctx, mgr.vpcId, ipVersion)
						if err != nil {
							return fmt.Errorf("call DescribeVpcCIDRBlock: %s", err.Error())
						}
					}
					if containsIP(vpcCIDRs, b.ServerIp) {
						return fmt.Errorf("can not find eniid for ip %s in vpc %s", b.ServerIp, mgr.vpcId)
					} else {
						skipIPs = append(skipIPs, b.ServerIp)
						continue
					}
				}
				b.ServerId = eniid
			}
			filteredBackends = append(filteredBackends, b)
		}
		vgs[i].Backends = filteredBackends
		if len(skipIPs) > 0 {
			reqCtx.Log.Info(fmt.Sprintf("warning: filter pods by vpc cidr %+v, podIPs=%+v", vpcCIDRs, skipIPs))
			reqCtx.Recorder.Event(
				reqCtx.Service,
				v1.EventTypeNormal,
				helper.SkipSyncBackends,
				fmt.Sprintf("Not sync pods [%s] whose ip is not in vpc cidrs", strings.Join(skipIPs, ",")),
			)
		}
	}

	return nil
}

func (mgr *VGroupManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, m *model.LoadBalancer) error {
	vgs, err := mgr.DescribeVServerGroups(reqCtx, m.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeVServerGroups error: %s", err.Error())
	}
	m.VServerGroups = vgs
	return nil
}

func (mgr *VGroupManager) ParallelUpdateVServerGroup(reqCtx *svcCtx.RequestContext, actions []vGroupAction, serverGroupChannel chan serverGroupApplyResult) []error {
	if len(actions) == 0 {
		reqCtx.Log.Info("no action to do for vgroup")
		return nil
	}
	errs := make([]error, len(actions))
	reqCtx.Log.V(5).Info("update vgroups parallelly", "actionsCount", len(actions))
	parallel.DoPiece(reqCtx.Ctx, ctrlCfg.ControllerCFG.MaxConcurrentActions, len(actions), func(i int) {
		var err error
		act := actions[i]
		switch act.Action {
		case vGroupActionCreateAndAddBackendServers:
			err = mgr.CreateVServerGroupAndAddBackendServers(reqCtx, act.Local, act.LBId, serverGroupChannel)
			if serverGroupChannel != nil {
				serverGroupChannel <- serverGroupApplyResult{
					VGroupName: act.Local.VGroupName,
					VGroupID:   act.Local.VGroupId,
					Err:        err,
				}
			}
		case vGroupActionUpdate:
			err = mgr.UpdateVServerGroup(reqCtx, *act.Local, *act.Remote)
		case vGroupActionDelete:
			err = mgr.DeleteVServerGroup(reqCtx, act.Remote.VGroupId)
		}
		errs[i] = err
	})
	return errs
}

func (mgr *VGroupManager) CreateVServerGroupAndAddBackendServers(reqCtx *svcCtx.RequestContext, local *model.VServerGroup, lbId string, sgChannel chan serverGroupApplyResult) error {
	err := mgr.CreateVServerGroup(reqCtx, local, lbId)
	if err != nil {
		return err
	}
	if len(local.Backends) > 0 {
		err = mgr.BatchAddVServerGroupBackendServers(reqCtx, *local, local.Backends)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mgr *VGroupManager) CreateVServerGroup(reqCtx *svcCtx.RequestContext, vg *model.VServerGroup, lbId string) error {
	reqCtx.Log.Info(fmt.Sprintf("create vgroup %s", vg.VGroupName))
	return mgr.cloud.CreateVServerGroup(reqCtx.Ctx, vg, lbId)
}

func (mgr *VGroupManager) DeleteVServerGroup(reqCtx *svcCtx.RequestContext, vGroupId string) error {
	reqCtx.Log.Info(fmt.Sprintf("delete vgroup %s", vGroupId))
	return mgr.cloud.DeleteVServerGroup(reqCtx.Ctx, vGroupId)
}

func (mgr *VGroupManager) UpdateVServerGroup(reqCtx *svcCtx.RequestContext, local, remote model.VServerGroup) error {
	add, del, update := diff(remote, local)
	if len(add) == 0 && len(del) == 0 && len(update) == 0 {
		reqCtx.Log.Info(fmt.Sprintf("reconcile vgroup: [%s] not change, skip reconcile", remote.VGroupId),
			"vgroupName", remote.VGroupName)
		return nil
	}

	reqCtx.Log.Info(fmt.Sprintf("reconcile vgroup: [%s] changed, local: [%s], remote: [%s]",
		remote.VGroupId, local.BackendInfo(), remote.BackendInfo()), "vgroupName", remote.VGroupName)

	var errs []error
	retryAdd := false

	if len(add) > 0 {
		if err := mgr.BatchAddVServerGroupBackendServers(reqCtx, local, add); err != nil {
			if strings.Contains(err.Error(), "InstanceBackendServerNumberOverLimit") {
				retryAdd = true
			} else {
				errs = append(errs, err)
			}
		}
	}

	if len(del) > 0 {
		if err := mgr.BatchRemoveVServerGroupBackendServers(reqCtx, remote, del); err != nil {
			errs = append(errs, err)
		}
	}

	if len(update) > 0 {
		if err := mgr.BatchUpdateVServerGroupBackendServers(reqCtx, remote, update); err != nil {
			errs = append(errs, err)
		}
	}

	if retryAdd {
		if err := mgr.BatchAddVServerGroupBackendServers(reqCtx, local, add); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (mgr *VGroupManager) DescribeVServerGroups(reqCtx *svcCtx.RequestContext, lbId string) ([]model.VServerGroup, error) {
	vgs, err := mgr.cloud.DescribeVServerGroups(reqCtx.Ctx, lbId)
	if err != nil {
		return vgs, err
	}

	reusedVgIDs, err := getVGroupIDs(reqCtx.Anno.Get(annotation.VGroupPort))
	if err != nil {
		return vgs, err
	}

	errs := make([]error, len(vgs))
	parallel.DoPiece(reqCtx.Ctx, ctrlCfg.ControllerCFG.MaxConcurrentActions, len(vgs), func(i int) {
		if isVGroupManagedByMyService(vgs[i], reqCtx.Service) || isReusedVGroup(reusedVgIDs, vgs[i].VGroupId) {
			vs, err := mgr.cloud.DescribeVServerGroupAttribute(reqCtx.Ctx, vgs[i].VGroupId)
			if err != nil {
				errs[i] = err
				return
			}
			for idx, backend := range vs.Backends {
				if !isBackendManagedByMyService(reqCtx, backend) {
					vs.Backends[idx].IsUserManaged = true
					// if backend is managed by user, vServerGroup is also managed by user.
					vgs[i].IsUserManaged = true
					reqCtx.Log.Info(fmt.Sprintf("vgroup %s backend is managed by user: [%+v]",
						vgs[i].VGroupName, vs.Backends[idx]))
				}
			}
			vgs[i].Backends = vs.Backends
		}
	})

	return vgs, nil
}

func (mgr *VGroupManager) BatchAddVServerGroupBackendServers(reqCtx *svcCtx.RequestContext, vGroup model.VServerGroup, add interface{}) error {
	batch := 0
	return backend.Batch(add, ctrlCfg.ControllerCFG.ServerGroupBatchSize,
		func(list []interface{}) error {
			batch += 1
			additions, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
			}
			reqCtx.Log.Info(fmt.Sprintf("reconcile vgroup: [%s] backend add [%s]", vGroup.VGroupId, string(additions)), "batch", batch)
			return mgr.cloud.AddVServerGroupBackendServers(reqCtx.Ctx, vGroup.VGroupId, string(additions))
		})
}

func (mgr *VGroupManager) BatchRemoveVServerGroupBackendServers(reqCtx *svcCtx.RequestContext, vGroup model.VServerGroup, del interface{}) error {
	batch := 0
	return backend.Batch(del, ctrlCfg.ControllerCFG.ServerGroupBatchSize,
		func(list []interface{}) error {
			batch += 1
			deletions, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
			}
			reqCtx.Log.Info(fmt.Sprintf("reconcile vgroup: [%s] backend del [%s]", vGroup.VGroupId, string(deletions)), "batch", batch)
			return mgr.cloud.RemoveVServerGroupBackendServers(reqCtx.Ctx, vGroup.VGroupId, string(deletions))
		})
}

func (mgr *VGroupManager) BatchUpdateVServerGroupBackendServers(reqCtx *svcCtx.RequestContext, vGroup model.VServerGroup, update interface{}) error {
	batch := 0
	return backend.Batch(update, ctrlCfg.ControllerCFG.ServerGroupBatchSize,
		func(list []interface{}) error {
			batch += 1
			updateJson, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
			}
			reqCtx.Log.Info(fmt.Sprintf("reconcile vgroup: [%s] backend update [%s]", vGroup.VGroupId, string(updateJson)), "batch", batch)
			return mgr.cloud.SetVServerGroupAttribute(reqCtx.Ctx, vGroup.VGroupId, string(updateJson))
		})
}

func diff(remote, local model.VServerGroup) (
	[]model.BackendAttribute, []model.BackendAttribute, []model.BackendAttribute) {

	var (
		addition  []model.BackendAttribute
		deletions []model.BackendAttribute
		updates   []model.BackendAttribute
	)

	for _, r := range remote.Backends {
		if r.IsUserManaged {
			continue
		}
		found := false
		for _, l := range local.Backends {
			if l.Type == "eni" {
				if r.ServerId == l.ServerId &&
					r.ServerIp == l.ServerIp &&
					r.Port == l.Port {
					found = true
					break
				}
			} else {
				if r.ServerId == l.ServerId &&
					r.Port == l.Port {
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
					r.ServerIp == l.ServerIp &&
					r.Port == l.Port {
					found = true
					break
				}
			} else {
				if r.ServerId == l.ServerId &&
					r.Port == l.Port {
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
			if l.Type == "eni" {
				if l.ServerId == r.ServerId &&
					l.ServerIp == r.ServerIp &&
					(l.Port != r.Port || (!local.IgnoreWeightUpdate && l.Weight != r.Weight) || l.Description != r.Description) {
					updates = append(updates, l)
					break
				}
			} else {
				if l.ServerId == r.ServerId &&
					(l.Port != r.Port || (!local.IgnoreWeightUpdate && l.Weight != r.Weight) || l.Description != r.Description) {
					updates = append(updates, l)
					break
				}
			}
		}
	}
	return addition, deletions, updates
}

func isBackendManagedByMyService(reqCtx *svcCtx.RequestContext, remoteBackend model.BackendAttribute) bool {
	namedKey, err := model.LoadVGroupNamedKey(remoteBackend.Description)
	if err != nil {
		return false
	}

	return namedKey.ServiceName == reqCtx.Service.Name &&
		namedKey.Namespace == reqCtx.Service.Namespace &&
		namedKey.CID == base.CLUSTER_ID
}

func isVGroupManagedByMyService(remote model.VServerGroup, service *v1.Service) bool {
	if remote.IsUserManaged || remote.NamedKey == nil {
		return false
	}

	return remote.NamedKey.ServiceName == service.Name &&
		remote.NamedKey.Namespace == service.Namespace &&
		remote.NamedKey.CID == base.CLUSTER_ID
}

func isReusedVGroup(reusedVgIDs []string, vGroupId string) bool {
	for _, vgID := range reusedVgIDs {
		if vGroupId == vgID {
			return true
		}
	}
	return false
}

func (mgr *VGroupManager) buildVGroupForServicePort(reqCtx *svcCtx.RequestContext,
	port v1.ServicePort, candidates *backend.EndpointWithENI, isUserManagedLB bool) (model.VServerGroup, bool, error) {

	vg := model.VServerGroup{
		NamedKey:    getVGroupNamedKey(reqCtx.Service, port),
		ServicePort: port,
	}
	vg.VGroupName = vg.NamedKey.Key()

	if isUserManagedLB && reqCtx.Anno.Get(annotation.VGroupPort) != "" {
		vgroupId, err := vgroup(reqCtx.Anno.Get(annotation.VGroupPort), port)
		if err != nil {
			return vg, false, fmt.Errorf("vgroupid parse error: %s", err.Error())
		}
		if vgroupId != "" {
			v, err := getVgroupById(mgr, reqCtx, vgroupId)
			if err != nil {
				return vg, false, fmt.Errorf("find vgroupId %s of lb %s error: %s", vgroupId,
					reqCtx.Anno.Get(annotation.LoadBalancerId), err.Error())
			}
			if v == nil {
				return vg, false, fmt.Errorf("can not find vgroupId %s in lb %s",
					vgroupId, reqCtx.Anno.Get(annotation.LoadBalancerId))
			}
			if !v.IsUserManaged {
				return vg, false, fmt.Errorf("cannot reuse a ccm created vgroup %s for slb %s", vgroupId, reqCtx.Anno.Get(annotation.LoadBalancerId))
			}
			reqCtx.Log.Info(fmt.Sprintf("user managed vgroupId %s for port %d", vgroupId, port.Port))
			vg.VGroupId = vgroupId
			vg.IsUserManaged = true
		}

		if reqCtx.Anno.Get(annotation.VGroupWeight) != "" {
			w, err := strconv.Atoi(reqCtx.Anno.Get(annotation.VGroupWeight))
			if err != nil || w < 0 || w > 100 {
				return vg, false, fmt.Errorf("weight must be integer in range [0,100] , got [%s]",
					reqCtx.Anno.Get(annotation.VGroupWeight))
			}
			vg.VGroupWeight = &w
		}
	}

	if strings.EqualFold(reqCtx.Anno.Get(annotation.IgnoreWeightUpdate), string(model.OnFlag)) {
		vg.IgnoreWeightUpdate = true
	}

	// build backends
	var (
		backends []model.BackendAttribute
		err      error
	)

	initialBackends, containsPotentialReadyEndpoints, err := mgr.setGenericBackendAttribute(reqCtx, candidates, vg)
	if err != nil {
		return vg, false, err
	}

	switch candidates.TrafficPolicy {
	case helper.ENITrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("eni mode, build backends for %s", vg.NamedKey))
		backends, err = mgr.buildENIBackends(reqCtx, candidates, initialBackends, vg)
		if err != nil {
			return vg, false, fmt.Errorf("build eni backends error: %s", err.Error())
		}
	case helper.LocalTrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("local mode, build backends for %s", vg.NamedKey))
		backends, err = mgr.buildLocalBackends(reqCtx, candidates, initialBackends, vg)
		if err != nil {
			return vg, false, fmt.Errorf("build local backends error: %s", err.Error())
		}
	case helper.ClusterTrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("cluster mode, build backends for %s", vg.NamedKey))
		backends, err = mgr.buildClusterBackends(reqCtx, candidates, initialBackends, vg)
		if err != nil {
			return vg, false, fmt.Errorf("build cluster backends error: %s", err.Error())
		}
	default:
		return vg, false, fmt.Errorf("not supported traffic policy [%s]", candidates.TrafficPolicy)
	}

	if len(backends) == 0 {
		reqCtx.Recorder.Event(
			reqCtx.Service,
			v1.EventTypeNormal,
			helper.UnAvailableBackends,
			"There are no available nodes for LoadBalancer",
		)
	}

	vg.Backends = backends
	vg.InitialBackends = initialBackends
	return vg, containsPotentialReadyEndpoints, nil
}

func getVgroupById(mgr *VGroupManager, reqCtx *svcCtx.RequestContext, vgroupId string) (*model.VServerGroup, error) {
	// check vgroup id is existed
	vgroups, err := mgr.cloud.DescribeVServerGroups(reqCtx.Ctx, reqCtx.Anno.Get(annotation.LoadBalancerId))
	if err != nil {
		return nil, fmt.Errorf("cannot find vgroup by vgroupId %s error: %s", vgroupId, err.Error())
	}
	for _, v := range vgroups {
		if v.VGroupId == vgroupId {
			return &v, nil
		}
	}
	return nil, nil
}

func (mgr *VGroupManager) setGenericBackendAttribute(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI, vgroup model.VServerGroup) ([]model.BackendAttribute, bool, error) {
	if utilfeature.DefaultFeatureGate.Enabled(ctrlCfg.EndpointSlice) {
		return mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
	}
	return mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
}

func (mgr *VGroupManager) setBackendsFromEndpoints(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI, vgroup model.VServerGroup) ([]model.BackendAttribute, bool, error) {
	var backends []model.BackendAttribute

	readinessGateName := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, reqCtx.Service.Name)
	containsPotentialReadyEndpoints := false

	if len(candidates.Endpoints.Subsets) == 0 {
		return nil, false, nil
	}
	for _, ep := range candidates.Endpoints.Subsets {
		var backendPort int
		if vgroup.ServicePort.TargetPort.Type == intstr.Int {
			backendPort = vgroup.ServicePort.TargetPort.IntValue()
		} else {
			for _, p := range ep.Ports {
				if p.Name == vgroup.ServicePort.Name {
					backendPort = int(p.Port)
					break
				}
			}
			if backendPort == 0 {
				klog.Warningf("%s cannot find port according port name: %s", vgroup.VGroupName, vgroup.ServicePort.Name)
				continue
			}
		}

		for _, addr := range ep.Addresses {
			backends = append(backends, model.BackendAttribute{
				NodeName: addr.NodeName,
				ServerIp: addr.IP,
				// set backend port to targetPort by default
				// if backend type is ecs, update backend port to nodePort
				Port:        backendPort,
				Description: vgroup.VGroupName,
			})
		}

		for _, addr := range ep.NotReadyAddresses {
			// ignore endpoint that not referenced to a pod
			if addr.TargetRef == nil || addr.TargetRef.Kind != "Pod" {
				continue
			}

			pod := &v1.Pod{}
			podKey := types.NamespacedName{Namespace: addr.TargetRef.Namespace, Name: addr.TargetRef.Name}
			err := mgr.kubeClient.Get(context.TODO(), podKey, pod)
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

			backends = append(backends, model.BackendAttribute{
				NodeName: addr.NodeName,
				ServerIp: addr.IP,
				// set backend port to targetPort by default
				// if backend type is ecs, update backend port to nodePort
				Port:        backendPort,
				Description: vgroup.VGroupName,
			})

		}
	}

	return backends, containsPotentialReadyEndpoints, nil
}

func (mgr *VGroupManager) setBackendsFromEndpointSlices(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI, vgroup model.VServerGroup) ([]model.BackendAttribute, bool, error) {
	// used for deduplicate when endpointslice is enabled
	// https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/#duplicate-endpoints
	endpointMap := make(map[string]bool)
	var backends []model.BackendAttribute

	containsPotentialReadyEndpoints := false
	readinessGateName := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, reqCtx.Service.Name)

	if len(candidates.EndpointSlices) == 0 {
		return nil, false, nil
	}

	for _, es := range candidates.EndpointSlices {
		var backendPort int
		if vgroup.ServicePort.TargetPort.Type == intstr.Int {
			backendPort = vgroup.ServicePort.TargetPort.IntValue()
		} else {
			for _, p := range es.Ports {
				// be compatible with IntOrString type target port
				if p.Name != nil && *p.Name == vgroup.ServicePort.Name {
					if p.Port != nil {
						backendPort = int(*p.Port)
					}
					break
				}
			}
			if backendPort == 0 {
				klog.Warningf("%s cannot find port according port name: %s", vgroup.VGroupName, vgroup.ServicePort.Name)
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
					backends = append(backends, model.BackendAttribute{
						NodeName: ep.NodeName,
						ServerIp: addr,
						// set backend port to targetPort by default
						// if backend type is ecs, update backend port to nodePort
						Port:        backendPort,
						Description: vgroup.VGroupName,
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
				err := mgr.kubeClient.Get(context.TODO(), podKey, pod)
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

				// TODO: may need to check node status
				endpointMap[addr] = true
				backends = append(backends, model.BackendAttribute{
					NodeName:    ep.NodeName,
					ServerIp:    addr,
					Port:        backendPort,
					Description: vgroup.VGroupName,
					TargetRef:   ep.TargetRef,
				})
			}
		}
	}
	return backends, containsPotentialReadyEndpoints, nil
}

func (mgr *VGroupManager) buildENIBackends(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI,
	backends []model.BackendAttribute, vgroup model.VServerGroup) ([]model.BackendAttribute, error) {
	if len(backends) == 0 {
		return nil, nil
	}

	backends, err := updateENIBackends(reqCtx, mgr, backends, candidates.AddressIPVersion)
	if err != nil {
		return backends, err
	}

	return setWeightBackends(helper.ENITrafficPolicy, backends, vgroup.VGroupWeight), nil
}

func (mgr *VGroupManager) buildLocalBackends(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI,
	initBackends []model.BackendAttribute, vgroup model.VServerGroup) ([]model.BackendAttribute, error) {
	var (
		ecsBackends, eciBackends []model.BackendAttribute
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

		_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}
		backend.ServerId = id
		backend.ServerIp = ""
		backend.Type = model.ECSBackendType
		// for ECS backend type, port should be set to NodePort
		backend.Port = int(vgroup.ServicePort.NodePort)
		ecsBackends = append(ecsBackends, backend)
	}

	// 2. add eci backends
	if len(eciBackends) != 0 {
		reqCtx.Log.Info("add eciBackends")
		eciBackends, err = updateENIBackends(reqCtx, mgr, eciBackends, candidates.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("update eci backends error: %s", err.Error())
		}
	}

	backends := append(ecsBackends, eciBackends...)

	// 3. set weight
	backends = setWeightBackends(helper.LocalTrafficPolicy, backends, vgroup.VGroupWeight)

	// 4. remove duplicated ecs
	return removeDuplicatedECS(backends), nil
}

func removeDuplicatedECS(backends []model.BackendAttribute) []model.BackendAttribute {
	nodeMap := make(map[string]bool)
	var uniqBackends []model.BackendAttribute
	for _, backend := range backends {
		if _, ok := nodeMap[backend.ServerId]; ok {
			continue
		}
		nodeMap[backend.ServerId] = true
		uniqBackends = append(uniqBackends, backend)
	}
	return uniqBackends

}

func (mgr *VGroupManager) buildClusterBackends(reqCtx *svcCtx.RequestContext, candidates *backend.EndpointWithENI,
	initBackends []model.BackendAttribute, vgroup model.VServerGroup) ([]model.BackendAttribute, error) {
	var (
		ecsBackends, eciBackends []model.BackendAttribute
		err                      error
	)

	// 1. add ecs backends. add all cluster nodes.
	for _, node := range candidates.Nodes {
		if helper.IsNodeExcludeFromLoadBalancer(&node) {
			reqCtx.Log.Info("node has exclude label or has type which cannot be added to lb backend", "node", node.Name)
			continue
		}
		_, id, err := helper.NodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("normal parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}

		ecsBackends = append(
			ecsBackends,
			model.BackendAttribute{
				ServerId:    id,
				Weight:      DefaultServerWeight,
				Port:        int(vgroup.ServicePort.NodePort),
				Type:        model.ECSBackendType,
				Description: vgroup.VGroupName,
			},
		)
	}

	// 2. add eci backends
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

		// check if the node is VK
		if node.Labels["type"] == helper.LabelNodeTypeVK {
			eciBackends = append(eciBackends, backend)
			continue
		}
	}

	if len(eciBackends) != 0 {
		eciBackends, err = updateENIBackends(reqCtx, mgr, eciBackends, candidates.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("update eci backends error: %s", err.Error())
		}
	}

	backends := append(ecsBackends, eciBackends...)

	return setWeightBackends(helper.ClusterTrafficPolicy, backends, vgroup.VGroupWeight), nil
}

func updateENIBackends(reqCtx *svcCtx.RequestContext, mgr *VGroupManager, backends []model.BackendAttribute, ipVersion model.AddressIPVersionType) ([]model.BackendAttribute, error) {
	for i := range backends {
		// for ENI backend type, port should be set to targetPort (default value), no need to update
		backends[i].Type = model.ENIBackendType
	}

	return backends, nil
}

func setWeightBackends(mode helper.TrafficPolicy, backends []model.BackendAttribute, weight *int) []model.BackendAttribute {
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
func podNumberAlgorithm(mode helper.TrafficPolicy, backends []model.BackendAttribute) []model.BackendAttribute {
	if mode == helper.ENITrafficPolicy || mode == helper.ClusterTrafficPolicy {
		for i := range backends {
			backends[i].Weight = DefaultServerWeight
		}
		return backends
	}

	// LocalTrafficPolicy
	ecsPods := make(map[string]int)
	for _, b := range backends {
		ecsPods[b.ServerId] += 1
	}
	for i := range backends {
		backends[i].Weight = ecsPods[backends[i].ServerId]
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
func podPercentAlgorithm(mode helper.TrafficPolicy, backends []model.BackendAttribute, weight int) []model.BackendAttribute {
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
			backends[i].Weight = per
		}
		return backends
	}

	// LocalTrafficPolicy
	ecsPods := make(map[string]int)
	for _, b := range backends {
		ecsPods[b.ServerId] += 1
	}
	for i := range backends {
		backends[i].Weight = weight * ecsPods[backends[i].ServerId] / len(backends)
		if backends[i].Weight < 1 {
			backends[i].Weight = 1
		}
	}
	return backends
}

func getVGroupIDs(annotation string) ([]string, error) {
	if annotation == "" {
		return nil, nil
	}
	var ids []string
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return nil, fmt.Errorf("vgroupid and "+
				"protocol format must be like 'vsp-xxx:443' with colon separated. got=[%+v]", pp)
		}
		ids = append(ids, pp[0])
	}
	return ids, nil
}

func containsIP(cidrs []*net.IPNet, serverIp string) bool {
	ip := net.ParseIP(serverIp)
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
