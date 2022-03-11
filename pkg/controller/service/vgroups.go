package service

import (
	"encoding/json"
	"fmt"
	discovery "k8s.io/api/discovery/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EndpointWithENI
// Currently, EndpointWithENI accept two kind of backend
// normal nodes of type []*v1.Node, and endpoints of type *v1.Endpoints
type TrafficPolicy string

const (
	// LocalTrafficPolicy externalTrafficPolicy=Local
	LocalTrafficPolicy = TrafficPolicy("Local")
	// ClusterTrafficPolicy externalTrafficPolicy=Cluster
	ClusterTrafficPolicy = TrafficPolicy("Cluster")
	// ENITrafficPolicy is forwarded to pod directly
	ENITrafficPolicy = TrafficPolicy("ENI")
)

func NewEndpointWithENI(reqCtx *RequestContext, kubeClient client.Client) (*EndpointWithENI, error) {
	endpointWithENI := &EndpointWithENI{}
	endpointWithENI.setTrafficPolicy(reqCtx)
	endpointWithENI.setAddressIpVersion(reqCtx)

	nodes, err := getNodes(reqCtx, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %s", err.Error())
	}
	endpointWithENI.Nodes = nodes

	if utilfeature.DefaultMutableFeatureGate.Enabled(helper.EndpointSlice) {
		esList, err := getEndpointByEndpointSlice(reqCtx, kubeClient, endpointWithENI.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("get endpointslice error: %s", err.Error())
		}
		endpointWithENI.EndpointSlices = esList
		reqCtx.Log.Info("backend details", "endpointslices", LogEndpointSliceList(esList))
	} else {
		eps, err := getEndpoints(reqCtx, kubeClient)
		if err != nil {
			return nil, fmt.Errorf("get endpoints error: %s", err.Error())
		}
		endpointWithENI.Endpoints = eps
		reqCtx.Log.Info("backend details", "endpoints", LogEndpoints(eps))
	}

	return endpointWithENI, nil
}

type EndpointWithENI struct {
	// TrafficPolicy
	// external traffic policy.
	TrafficPolicy TrafficPolicy
	// AddressIPVersion
	// it indicates the address ip version type of the backends attached to the LoadBalancer
	AddressIPVersion model.AddressIPVersionType
	// Nodes
	// contains all the candidate nodes consider of LoadBalance Backends.
	// Cloud implementation has the right to make any filter on it.
	Nodes []v1.Node
	// Endpoints
	// It is the direct pod location information which cloud implementation
	// may needed for some kind of filtering. eg. direct ENI attach.
	Endpoints *v1.Endpoints
	// EndpointSlices
	// contains all the endpointslices of a service
	EndpointSlices []discovery.EndpointSlice
}

func (e *EndpointWithENI) setTrafficPolicy(reqCtx *RequestContext) {
	if isENIBackendType(reqCtx.Service) {
		e.TrafficPolicy = ENITrafficPolicy
		return
	}
	if isLocalModeService(reqCtx.Service) {
		e.TrafficPolicy = LocalTrafficPolicy
		return
	}
	e.TrafficPolicy = ClusterTrafficPolicy
}

func (e *EndpointWithENI) setAddressIpVersion(reqCtx *RequestContext) {
	// Only EndpointSlice support dual stack.
	// Enable IPv6DualStack and EndpointSlice feature gates if you want to use ipv6 backends
	if utilfeature.DefaultMutableFeatureGate.Enabled(helper.IPv6DualStack) &&
		utilfeature.DefaultMutableFeatureGate.Enabled(helper.EndpointSlice) &&
		reqCtx.Anno.Get(IPVersion) == string(model.IPv6) &&
		reqCtx.Anno.Get(BackendIPVersion) == string(model.IPv6) {
		e.AddressIPVersion = model.IPv6
		reqCtx.Log.Info("backend address ip version is ipv6")
		return
	}
	e.AddressIPVersion = model.IPv4
}

func NewVGroupManager(kubeClient client.Client, cloud prvd.Provider) *VGroupManager {
	return &VGroupManager{
		kubeClient: kubeClient,
		cloud:      cloud,
	}
}

type VGroupManager struct {
	kubeClient client.Client
	cloud      prvd.Provider
}

func (mgr *VGroupManager) BuildLocalModel(reqCtx *RequestContext, m *model.LoadBalancer) error {
	var vgs []model.VServerGroup

	candidates, err := NewEndpointWithENI(reqCtx, mgr.kubeClient)
	if err != nil {
		return err
	}

	for _, port := range reqCtx.Service.Spec.Ports {
		vg, err := mgr.buildVGroupForServicePort(reqCtx, port, candidates, m.LoadBalancerAttribute.IsUserManaged)
		if err != nil {
			return fmt.Errorf("build vgroup for port %d error: %s", port.Port, err.Error())
		}
		vgs = append(vgs, vg)
	}
	m.VServerGroups = vgs
	return nil
}

func (mgr *VGroupManager) BuildRemoteModel(reqCtx *RequestContext, m *model.LoadBalancer) error {
	vgs, err := mgr.DescribeVServerGroups(reqCtx, m.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeVServerGroups error: %s", err.Error())
	}
	m.VServerGroups = vgs
	return nil
}

func (mgr *VGroupManager) CreateVServerGroup(reqCtx *RequestContext, vg *model.VServerGroup, lbId string) error {
	reqCtx.Log.Info(fmt.Sprintf("create vgroup %s", vg.VGroupName))
	return mgr.cloud.CreateVServerGroup(reqCtx.Ctx, vg, lbId)
}

func (mgr *VGroupManager) DeleteVServerGroup(reqCtx *RequestContext, vGroupId string) error {
	reqCtx.Log.Info(fmt.Sprintf("delete vgroup %s", vGroupId))
	return mgr.cloud.DeleteVServerGroup(reqCtx.Ctx, vGroupId)
}

func (mgr *VGroupManager) UpdateVServerGroup(reqCtx *RequestContext, local, remote model.VServerGroup) error {
	add, del, update := diff(reqCtx, remote, local)
	if len(add) == 0 && len(del) == 0 && len(update) == 0 {
		reqCtx.Log.Info(fmt.Sprintf("update vgroup [%s]: no change, skip reconcile", remote.VGroupId))
	} else {
		reqCtx.Log.Info(fmt.Sprintf("try to update vgroup [%s]: local: [%s], remote: [%s]",
			remote.VGroupId, local.BackendInfo(), remote.BackendInfo()))
	}
	if len(add) > 0 {
		if err := mgr.BatchAddVServerGroupBackendServers(reqCtx, local, add); err != nil {
			return err
		}
	}
	if len(del) > 0 {
		if err := mgr.BatchRemoveVServerGroupBackendServers(reqCtx, remote, del); err != nil {
			return err
		}
	}
	if len(update) > 0 {
		return mgr.BatchUpdateVServerGroupBackendServers(reqCtx, remote, update)
	}
	return nil
}

func (mgr *VGroupManager) DescribeVServerGroups(reqCtx *RequestContext, lbId string) ([]model.VServerGroup, error) {
	return mgr.cloud.DescribeVServerGroups(reqCtx.Ctx, lbId)
}

func (mgr *VGroupManager) BatchAddVServerGroupBackendServers(reqCtx *RequestContext, vGroup model.VServerGroup, add interface{}) error {
	return Batch(add, MaxBackendNum,
		func(list []interface{}) error {
			additions, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
			}
			reqCtx.Log.Info(fmt.Sprintf("update vgroup [%s]: backend add [%s]", vGroup.VGroupId, string(additions)))
			return mgr.cloud.AddVServerGroupBackendServers(reqCtx.Ctx, vGroup.VGroupId, string(additions))
		})
}

func (mgr *VGroupManager) BatchRemoveVServerGroupBackendServers(reqCtx *RequestContext, vGroup model.VServerGroup, del interface{}) error {
	return Batch(del, MaxBackendNum,
		func(list []interface{}) error {
			deletions, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
			}
			reqCtx.Log.Info(fmt.Sprintf("update vgroup [%s]: backend del [%s]", vGroup.VGroupId, string(deletions)))
			return mgr.cloud.RemoveVServerGroupBackendServers(reqCtx.Ctx, vGroup.VGroupId, string(deletions))
		})
}

func (mgr *VGroupManager) BatchUpdateVServerGroupBackendServers(reqCtx *RequestContext, vGroup model.VServerGroup, update interface{}) error {
	return Batch(update, MaxBackendNum,
		func(list []interface{}) error {
			updateJson, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
			}
			reqCtx.Log.Info(fmt.Sprintf("update vgroup [%s]: backend update [%s]", vGroup.VGroupId, string(updateJson)))
			return mgr.cloud.SetVServerGroupAttribute(reqCtx.Ctx, vGroup.VGroupId, string(updateJson))
		})
}

func diff(reqCtx *RequestContext, remote, local model.VServerGroup) (
	[]model.BackendAttribute, []model.BackendAttribute, []model.BackendAttribute) {

	var (
		addition  []model.BackendAttribute
		deletions []model.BackendAttribute
		updates   []model.BackendAttribute
	)

	for _, r := range remote.Backends {
		if !isBackendManagedByMyService(reqCtx, r, local.VGroupName) {
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
					(l.Port != r.Port || l.Weight != r.Weight || l.Description != r.Description) {
					updates = append(updates, l)
					break
				}
			} else {
				if l.ServerId == r.ServerId &&
					(l.Port != r.Port || l.Weight != r.Weight || l.Description != r.Description) {
					updates = append(updates, l)
					break
				}
			}
		}
	}
	return addition, deletions, updates
}

func isBackendManagedByMyService(reqCtx *RequestContext, remoteBackend model.BackendAttribute, localVGroupName string) bool {
	if localVGroupName != "" {
		return remoteBackend.Description == localVGroupName
	}

	namedKey, err := model.LoadVGroupNamedKey(remoteBackend.Description)
	if err != nil {
		reqCtx.Log.V(5).Info(fmt.Sprintf("warning: %s description %s which is managed by user, skip delete",
			remoteBackend.ServerId, remoteBackend.Description))
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

func getNodes(reqCtx *RequestContext, client client.Client) ([]v1.Node, error) {
	nodeList := v1.NodeList{}
	err := client.List(reqCtx.Ctx, &nodeList)
	if err != nil {
		return nil, fmt.Errorf("get nodes error: %s", err.Error())
	}

	// 1. filter by label
	items := nodeList.Items
	if reqCtx.Anno.Get(BackendLabel) != "" {
		items, err = filterOutByLabel(nodeList.Items, reqCtx.Anno.Get(BackendLabel))
		if err != nil {
			return nil, fmt.Errorf("filter nodes by label error: %s", err.Error())
		}
	}

	var nodes []v1.Node
	for _, n := range items {
		if needExcludeFromLB(reqCtx, &n) {
			continue
		}
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func filterOutByLabel(nodes []v1.Node, labels string) ([]v1.Node, error) {
	if labels == "" {
		return nodes, nil
	}
	var result []v1.Node
	lbl := strings.Split(labels, ",")
	var records []string
	for _, node := range nodes {
		found := true
		for _, v := range lbl {
			l := strings.Split(v, "=")
			if len(l) < 2 {
				return []v1.Node{}, fmt.Errorf("parse backend label: %s, [k1=v1,k2=v2]", v)
			}
			if nv, exist := node.Labels[l[0]]; !exist || nv != l[1] {
				found = false
				break
			}
		}
		if found {
			result = append(result, node)
			records = append(records, node.Name)
		}
	}
	klog.V(4).Infof("accept nodes backend labels[%s], %v", labels, records)
	return result, nil
}

func needExcludeFromLB(reqCtx *RequestContext, node *v1.Node) bool {
	// need to keep the node who has exclude label in order to be compatible with vk node
	// It's safe because these nodes will be filtered in build backends func

	if isMasterNode(node) {
		klog.V(5).Infof("[%s] node %s is master node, skip adding it to lb", util.Key(reqCtx.Service), node.Name)
		return true
	}

	// filter unscheduled node
	if node.Spec.Unschedulable && reqCtx.Anno.Get(RemoveUnscheduled) != "" {
		if reqCtx.Anno.Get(RemoveUnscheduled) == string(model.OnFlag) {
			reqCtx.Log.Info("node is unschedulable, skip add to lb", "node", node.Name)
			return true
		}
	}

	// ignore vk node condition check.
	// Even if the vk node is NotReady, it still can be added to lb. Because the eci pod that actually joins the lb, not a vk node
	if label, ok := node.Labels["type"]; ok && label == LabelNodeTypeVK {
		return false
	}

	// If we have no info, don't accept
	if len(node.Status.Conditions) == 0 {
		reqCtx.Log.Info("node condition is nil, skip add to lb", "node", node.Name)
		return true
	}

	for _, cond := range node.Status.Conditions {
		// We consider the node for load balancing only when its NodeReady
		// condition status is ConditionTrue
		if cond.Type == v1.NodeReady &&
			cond.Status != v1.ConditionTrue {
			reqCtx.Log.Info(fmt.Sprintf("node not ready with %v condition, status %v", cond.Type, cond.Status),
				"node", node.Name)
			return true
		}
	}

	return false
}

func getEndpoints(reqCtx *RequestContext, client client.Client) (*v1.Endpoints, error) {
	eps := &v1.Endpoints{}
	err := client.Get(reqCtx.Ctx, util.NamespacedName(reqCtx.Service), eps)
	if err != nil && apierrors.IsNotFound(err) {
		reqCtx.Log.Info("warning: endpoint not found")
		return eps, nil
	}
	return eps, err
}

func getEndpointByEndpointSlice(reqCtx *RequestContext, kubeClient client.Client, ipVersion model.AddressIPVersionType) ([]discovery.EndpointSlice, error) {
	epsList := &discovery.EndpointSliceList{}
	err := kubeClient.List(reqCtx.Ctx, epsList, client.MatchingLabels{
		discovery.LabelServiceName: reqCtx.Service.Name,
	}, client.InNamespace(reqCtx.Service.Namespace))
	if err != nil {
		return nil, err
	}

	addressType := discovery.AddressTypeIPv4
	if ipVersion == model.IPv6 {
		addressType = discovery.AddressTypeIPv6
	}

	var ret []discovery.EndpointSlice
	for _, es := range epsList.Items {
		if es.AddressType == addressType {
			ret = append(ret, es)
		}
	}

	return ret, nil
}

func (mgr *VGroupManager) buildVGroupForServicePort(reqCtx *RequestContext, port v1.ServicePort,
	candidates *EndpointWithENI, isUserManagedLB bool) (model.VServerGroup, error) {
	vg := model.VServerGroup{
		NamedKey:    getVGroupNamedKey(reqCtx.Service, port),
		ServicePort: port,
	}
	vg.VGroupName = vg.NamedKey.Key()

	if isUserManagedLB && reqCtx.Anno.Get(VGroupPort) != "" {
		vgroupId, err := vgroup(reqCtx.Anno.Get(VGroupPort), port)
		if err != nil {
			return vg, fmt.Errorf("vgroupid parse error: %s", err.Error())
		}
		if vgroupId != "" {
			exist, err := isVGroupIdExist(mgr, reqCtx, vgroupId)
			if err != nil {
				return vg, fmt.Errorf("find vgroupId %s of lb %s error: %s", vgroupId,
					reqCtx.Anno.Get(LoadBalancerId), err.Error())
			}
			if !exist {
				return vg, fmt.Errorf("can not find vgroupId %s in lb %s", vgroupId, reqCtx.Anno.Get(LoadBalancerId))
			}
			reqCtx.Log.Info(fmt.Sprintf("user managed vgroupId %s for port %d", vgroupId, port.Port))
			vg.VGroupId = vgroupId
			vg.IsUserManaged = true
		}

		if reqCtx.Anno.Get(VGroupWeight) != "" {
			w, err := strconv.Atoi(reqCtx.Anno.Get(VGroupWeight))
			if err != nil || w < 0 || w > 100 {
				return vg, fmt.Errorf("weight must be integer in range [0,100] , got [%s]", reqCtx.Anno.Get(VGroupWeight))
			}
			vg.VGroupWeight = &w
		}
	}

	// build backends
	var (
		backends []model.BackendAttribute
		err      error
	)
	switch candidates.TrafficPolicy {
	case ENITrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("eni mode, build backends for %s", vg.NamedKey))
		backends, err = mgr.buildENIBackends(candidates, vg)
		if err != nil {
			return vg, fmt.Errorf("build eni backends error: %s", err.Error())
		}
	case LocalTrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("local mode, build backends for %s", vg.NamedKey))
		backends, err = mgr.buildLocalBackends(reqCtx, candidates, vg)
		if err != nil {
			return vg, fmt.Errorf("build local backends error: %s", err.Error())
		}
	case ClusterTrafficPolicy:
		reqCtx.Log.Info(fmt.Sprintf("cluster mode, build backends for %s", vg.NamedKey))
		backends, err = mgr.buildClusterBackends(reqCtx, candidates, vg)
		if err != nil {
			return vg, fmt.Errorf("build cluster backends error: %s", err.Error())
		}
	default:
		return vg, fmt.Errorf("not supported traffic policy [%s]", candidates.TrafficPolicy)
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
	return vg, nil
}

func isVGroupIdExist(mgr *VGroupManager, reqCtx *RequestContext, vgroupId string) (bool, error) {
	// check vgroup id is existed
	vgroups, err := mgr.cloud.DescribeVServerGroups(reqCtx.Ctx, reqCtx.Anno.Get(LoadBalancerId))
	if err != nil {
		return false, fmt.Errorf("cannot find vgroup by vgroupId %s error: %s", vgroupId, err.Error())
	}
	for _, v := range vgroups {
		if v.VGroupId == vgroupId {
			return true, nil
		}
	}
	return false, nil
}

func setGenericBackendAttribute(candidates *EndpointWithENI, vgroup model.VServerGroup) []model.BackendAttribute {
	if utilfeature.DefaultMutableFeatureGate.Enabled(helper.EndpointSlice) {
		return setBackendsFromEndpointSlices(candidates, vgroup)
	}
	return setBackendsFromEndpoints(candidates, vgroup)
}

func setBackendsFromEndpoints(candidates *EndpointWithENI, vgroup model.VServerGroup) []model.BackendAttribute {
	var backends []model.BackendAttribute

	if len(candidates.Endpoints.Subsets) == 0 {
		return nil
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
	}
	return backends
}

func setBackendsFromEndpointSlices(candidates *EndpointWithENI, vgroup model.VServerGroup) []model.BackendAttribute {
	// used for deduplicate when endpointslice is enabled
	// https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/#duplicate-endpoints
	endpointMap := make(map[string]bool)
	var backends []model.BackendAttribute

	if len(candidates.EndpointSlices) == 0 {
		return nil
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
			}
		}

		for _, ep := range es.Endpoints {
			if ep.Conditions.Ready == nil {
				continue
			}
			if !*ep.Conditions.Ready {
				continue
			}

			for _, addr := range ep.Addresses {
				if _, ok := endpointMap[addr]; ok {
					continue
				}
				endpointMap[addr] = true
				// NodeName of endpoint is nil, use topology.hostname instead of NodeName
				hostName := ep.Topology[v1.LabelHostname]
				backends = append(backends, model.BackendAttribute{
					NodeName: &hostName,
					ServerIp: addr,
					// set backend port to targetPort by default
					// if backend type is ecs, update backend port to nodePort
					Port:        backendPort,
					Description: vgroup.VGroupName,
				})
			}
		}
	}

	return backends
}

func (mgr *VGroupManager) buildENIBackends(candidates *EndpointWithENI, vgroup model.VServerGroup) ([]model.BackendAttribute, error) {
	backends := setGenericBackendAttribute(candidates, vgroup)
	if len(backends) == 0 {
		return nil, nil
	}

	backends, err := updateENIBackends(mgr, backends, candidates.AddressIPVersion)
	if err != nil {
		return backends, err
	}

	return setWeightBackends(ENITrafficPolicy, backends, vgroup.VGroupWeight), nil
}

func (mgr *VGroupManager) buildLocalBackends(reqCtx *RequestContext, candidates *EndpointWithENI, vgroup model.VServerGroup) ([]model.BackendAttribute, error) {
	initBackends := setGenericBackendAttribute(candidates, vgroup)
	if len(initBackends) == 0 {
		return nil, nil
	}

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
		node := findNodeByNodeName(candidates.Nodes, *backend.NodeName)
		if node == nil {
			reqCtx.Log.Info(fmt.Sprintf("warning: can not find correspond node %s for endpoint %s", *backend.NodeName, backend.ServerIp))
			continue
		}

		// check if the node is virtual node, virtual node add as eci backend
		if node.Labels["type"] == LabelNodeTypeVK {
			eciBackends = append(eciBackends, backend)
			continue
		}

		if isNodeExcludeFromLoadBalancer(node) {
			reqCtx.Log.Info("node has exclude label which cannot be added to lb backend", "node", node.Name)
			continue
		}

		_, id, err := nodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}
		backend.ServerId = id
		backend.Type = model.ECSBackendType
		// for ECS backend type, port should be set to NodePort
		backend.Port = int(vgroup.ServicePort.NodePort)
		ecsBackends = append(ecsBackends, backend)
	}

	// 2. add eci backends
	if len(eciBackends) != 0 {
		reqCtx.Log.Info("add eciBackends")
		eciBackends, err = updateENIBackends(mgr, eciBackends, candidates.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("update eci backends error: %s", err.Error())
		}
	}

	backends := append(ecsBackends, eciBackends...)

	// 3. set weight
	backends = setWeightBackends(LocalTrafficPolicy, backends, vgroup.VGroupWeight)

	// 4. remove duplicated ecs
	return remoteDuplicatedECS(backends), nil
}

func remoteDuplicatedECS(backends []model.BackendAttribute) []model.BackendAttribute {
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

func (mgr *VGroupManager) buildClusterBackends(reqCtx *RequestContext, candidates *EndpointWithENI, vgroup model.VServerGroup) ([]model.BackendAttribute, error) {
	initBackends := setGenericBackendAttribute(candidates, vgroup)
	if len(initBackends) == 0 {
		return nil, nil
	}

	var (
		ecsBackends, eciBackends []model.BackendAttribute
		err                      error
	)

	// 1. add ecs backends. add all cluster nodes.
	for _, node := range candidates.Nodes {
		if isNodeExcludeFromLoadBalancer(&node) {
			reqCtx.Log.Info("node has exclude label which cannot be added to lb backend", "node", node.Name)
			continue
		}
		_, id, err := nodeFromProviderID(node.Spec.ProviderID)
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
		node := findNodeByNodeName(candidates.Nodes, *backend.NodeName)
		if node == nil {
			reqCtx.Log.Info(fmt.Sprintf("warning: can not find correspond node %s for endpoint %s", *backend.NodeName, backend.ServerIp))
			continue
		}

		// check if the node is VK
		if node.Labels["type"] == LabelNodeTypeVK {
			eciBackends = append(eciBackends, backend)
			continue
		}
	}

	if len(eciBackends) != 0 {
		eciBackends, err = updateENIBackends(mgr, eciBackends, candidates.AddressIPVersion)
		if err != nil {
			return nil, fmt.Errorf("update eci backends error: %s", err.Error())
		}
	}

	backends := append(ecsBackends, eciBackends...)

	return setWeightBackends(ClusterTrafficPolicy, backends, vgroup.VGroupWeight), nil
}

func updateENIBackends(mgr *VGroupManager, backends []model.BackendAttribute, ipVersion model.AddressIPVersionType) (
	[]model.BackendAttribute, error) {
	vpcId, err := mgr.cloud.VpcID()
	if err != nil {
		return nil, fmt.Errorf("get vpc id from metadata error:%s", err.Error())
	}

	var ips []string
	for _, b := range backends {
		ips = append(ips, b.ServerIp)
	}

	result, err := mgr.cloud.DescribeNetworkInterfaces(vpcId, ips, ipVersion)
	if err != nil {
		return nil, fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
	}

	for i := range backends {
		eniid, ok := result[backends[i].ServerIp]
		if !ok {
			return nil, fmt.Errorf("can not find eniid for ip %s in vpc %s", backends[i].ServerIp, vpcId)
		}
		// for ENI backend type, port should be set to targetPort (default value), no need to update
		backends[i].ServerId = eniid
		backends[i].Type = model.ENIBackendType
	}
	return backends, nil
}

func setWeightBackends(mode TrafficPolicy, backends []model.BackendAttribute, weight *int) []model.BackendAttribute {
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
func podNumberAlgorithm(mode TrafficPolicy, backends []model.BackendAttribute) []model.BackendAttribute {
	if mode == ENITrafficPolicy || mode == ClusterTrafficPolicy {
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
func podPercentAlgorithm(mode TrafficPolicy, backends []model.BackendAttribute, weight int) []model.BackendAttribute {
	if weight == 0 {
		for i := range backends {
			backends[i].Weight = 0
		}
		return backends
	}

	if mode == ENITrafficPolicy || mode == ClusterTrafficPolicy {
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
