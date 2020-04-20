package alicloud

import (
	"context"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/klog"
	"reflect"
	"strings"
)

type vgroup struct {
	NamedKey       *NamedKey
	LoadBalancerId string
	RegionId       common.Region
	VpcID          string
	VGroupId       string
	Client         ClientSLBSDK
	InsClient      ClientInstanceSDK
	BackendServers []slb.VBackendServerType
}

func (v *vgroup) Logf(format string, args ...interface{}) {
	prefix := ""
	if v.NamedKey != nil {
		prefix = fmt.Sprintf("[%s/%s]", v.NamedKey.Namespace, v.NamedKey.ServiceName)
	}
	klog.Infof(prefix+format, args...)
}

func (v *vgroup) Describe(ctx context.Context) error {
	if v.NamedKey == nil {
		return fmt.Errorf("describe: format error of vgroup name")
	}
	vargs := slb.DescribeVServerGroupsArgs{
		RegionId:       v.RegionId,
		LoadBalancerId: v.LoadBalancerId,
	}
	vgrp, err := v.Client.DescribeVServerGroups(ctx, &vargs)
	if err != nil {
		return fmt.Errorf("describe: vgroup error, %s", err.Error())
	}
	if vgrp != nil {
		for _, val := range vgrp.VServerGroups.VServerGroup {
			if val.VServerGroupName ==
				v.NamedKey.Key() {
				v.VGroupId = val.VServerGroupId
				return nil
			}
		}
	}
	return fmt.Errorf("vgroup not found, %s", v.NamedKey.Key())
}
func (v *vgroup) Add(ctx context.Context) error {
	if v.VGroupId != "" {
		return fmt.Errorf("vgroupid already exists")
	}
	if v.NamedKey == nil {
		return fmt.Errorf("format error of vgroup name")
	}
	vgp := slb.CreateVServerGroupArgs{
		LoadBalancerId:   v.LoadBalancerId,
		VServerGroupName: v.NamedKey.Key(),
		RegionId:         v.RegionId,
	}

	if len(v.BackendServers) >= 1 {
		// work around for vserver group old version,it needs backend on creating.
		backend, err := json.Marshal(v.BackendServers[0:1])
		if err != nil {
			return fmt.Errorf("add new vserver group: %s", err.Error())
		}
		vgp.BackendServers = string(backend)
	}
	gp, err := v.Client.CreateVServerGroup(ctx, &vgp)
	if err != nil {
		return fmt.Errorf("CreateVServerGroup. %s", err.Error())
	}
	v.Logf("create new vserver group[%s]"+
		" for loadbalancer[%s] with empty backend list", v.NamedKey.Key(), v.LoadBalancerId)
	v.VGroupId = gp.VServerGroupId
	return nil
}
func (v *vgroup) Remove(ctx context.Context) error {
	if v.LoadBalancerId == "" || v.VGroupId == "" {
		return fmt.Errorf("can not delete vserver group. LoadBalancerId or vgroup id should not be empty")
	}
	vdel := slb.DeleteVServerGroupArgs{
		VServerGroupId: v.VGroupId,
		RegionId:       v.RegionId,
	}
	_, err := v.Client.DeleteVServerGroup(ctx, &vdel)
	return err
}
func (v *vgroup) Update(ctx context.Context) error {
	if v.VGroupId == "" {
		err := v.Describe(ctx)
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("update: vserver group error, %s", err.Error())
			}
			if err := v.Add(ctx); err != nil {
				return err
			}
		}
	}

	v.Logf("update: backend vgroupid [%s]", v.VGroupId)
	dsc := &slb.DescribeVServerGroupAttributeArgs{
		VServerGroupId: v.VGroupId,
		RegionId:       v.RegionId,
	}
	att, err := v.Client.DescribeVServerGroupAttribute(ctx, dsc)
	if err != nil {
		return fmt.Errorf("update: describe vserver group attribute error. %s", err.Error())
	}
	v.Logf("update: apis[%v], node[%v]", att.BackendServers.BackendServer, v.BackendServers)
	add, del, update := v.diff(att.BackendServers.BackendServer, v.BackendServers)
	if len(add) == 0 && len(del) == 0 && len(update) == 0 {
		v.Logf("update: no backend need to be added for vgroupid [%s]", v.VGroupId)
		return nil
	}

	if len(add) > 0 {
		if err := Batch(add, MAX_BACKEND_NUM,
			func(list []interface{}) error {
				additions, err := json.Marshal(list)
				if err != nil {
					return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
				}
				v.Logf("update: try to update vserver group[%s],"+
					" backend add[%s]", v.NamedKey.Key(), string(additions))
				_, err = v.Client.AddVServerGroupBackendServers(
					ctx,
					&slb.AddVServerGroupBackendServersArgs{
						VServerGroupId: v.VGroupId,
						RegionId:       v.RegionId,
						BackendServers: string(additions),
					})
				return err
			}); err != nil {
			return err
		}
	}
	if len(del) > 0 {
		if err := Batch(del, MAX_BACKEND_NUM,
			func(list []interface{}) error {
				deletions, err := json.Marshal(list)
				if err != nil {
					return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
				}
				v.Logf("update: try to update vserver group[%s],"+
					" backend del[%s]", v.NamedKey.Key(), string(deletions))
				_, err = v.Client.RemoveVServerGroupBackendServers(
					ctx,
					&slb.RemoveVServerGroupBackendServersArgs{
						VServerGroupId: v.VGroupId,
						RegionId:       v.RegionId,
						BackendServers: string(deletions),
					})
				return err
			}); err != nil {
			return err
		}
	}
	if len(update) > 0 {
		return Batch(update, MAX_BACKEND_NUM,
			func(list []interface{}) error {
				updateJson, err := json.Marshal(list)
				if err != nil {
					return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
				}
				v.Logf("update: try to update vserver group[%s],"+
					" backend del[%s]", v.NamedKey.Key(), string(updateJson))
				_, err = v.Client.SetVServerGroupAttribute(
					ctx,
					&slb.SetVServerGroupAttributeArgs{
						VServerGroupId: v.VGroupId,
						RegionId:       v.RegionId,
						BackendServers: string(updateJson),
					})
				return err
			})
	}
	return nil
}

// MAX_BACKEND_NUM max batch backend num
const MAX_BACKEND_NUM = 19

type Func func([]interface{}) error

// Batch batch process `object` m with func `func`
// for general purpose
func Batch(m interface{}, cnt int, batch Func) error {
	if cnt <= 0 {
		cnt = MAX_BACKEND_NUM
	}
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("non-slice type for %v", m)
	}

	// need to convert interface to []interface
	// see https://github.com/golang/go/wiki/InterfaceSlice
	target := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		target[i] = v.Index(i).Interface()
	}
	klog.Infof("batch process ,total length %d", len(target))
	for len(target) > cnt {
		if err := batch(target[0:cnt]); err != nil {

			return err
		}
		target = target[cnt:]
	}
	if len(target) <= 0 {
		return nil
	}

	klog.Infof("batch process ,total length %d last section", len(target))
	return batch(target)
}

func (v *vgroup) diff(apis, nodes []slb.VBackendServerType) (
	[]slb.VBackendServerType, []slb.VBackendServerType, []slb.VBackendServerType) {

	var (
		addition  []slb.VBackendServerType
		deletions []slb.VBackendServerType
		updates   []slb.VBackendServerType
	)

	for _, api := range apis {
		found := false
		for _, node := range nodes {
			if node.Type == "eni" {
				if api.ServerId == node.ServerId &&
					api.ServerIp == node.ServerIp {
					found = true
					break
				}
			} else {
				if api.ServerId == node.ServerId {
					found = true
					break
				}
			}

		}
		if !found {
			deletions = append(deletions, api)
		}
	}
	for _, node := range nodes {
		found := false
		for _, api := range apis {
			if node.Type == "eni" {
				if api.ServerId == node.ServerId &&
					api.ServerIp == node.ServerIp {
					found = true
					break
				}
			} else {
				if api.ServerId == node.ServerId {
					found = true
					break
				}
			}

		}
		if !found {
			addition = append(addition, node)
		}
	}
	for _, node := range nodes {
		for _, api := range apis {
			if node.Type == "eni" {
				if node.ServerId == api.ServerId &&
					node.ServerIp == api.ServerIp &&
					(node.Weight != api.Weight ||
						api.Description != v.NamedKey.Key()) {
					updates = append(updates, node)
					break
				}
			} else {
				if node.ServerId == api.ServerId &&
					(node.Weight != api.Weight ||
						api.Description != v.NamedKey.Key()) {
					updates = append(updates, node)
					break
				}
			}
		}
	}
	return addition, deletions, updates
}

func Ensure(ctx context.Context, v *vgroup, nodes *EndpointWithENI) error {
	backend, err := nodes.BuildBackend(ctx, v)
	if err != nil {
		return fmt.Errorf("build backend: %s, %s", err.Error(), v.NamedKey)
	}
	v.BackendServers = backend
	return v.Update(ctx)
}

type vgroups []*vgroup

func EnsureVirtualGroups(ctx context.Context, vgrps *vgroups, nodes *EndpointWithENI) error {
	klog.Infof("ensure vserver group: %d vgroup need to be processed.", len(*vgrps))
	for _, v := range *vgrps {
		if v == nil {
			return fmt.Errorf("unexpected nil vgroup ")
		}
		if err := Ensure(ctx, v, nodes); err != nil {
			return fmt.Errorf("ensure vgroup: %s. %s", err.Error(), v.NamedKey.Key())
		}
		v.Logf("EnsureGroup: id=[%s], Name:[%s], LoadBalancerId:[%s]", v.VGroupId, v.NamedKey.Key(), v.LoadBalancerId)
	}
	return nil
}

//CleanUPVGroupMerged Merge with service port and do clean vserver group
func CleanUPVGroupMerged(
	ctx context.Context,
	slbins *LoadBalancerClient,
	service *v1.Service,
	lb *slb.LoadBalancerType,
	local *vgroups,
) error {

	remote, err := BuildVirtualGroupFromRemoteAPI(ctx, lb, slbins)
	if err != nil {
		return fmt.Errorf("build vserver group from remote: %s", err.Error())
	}
	for _, rem := range remote {
		if rem.NamedKey.ServiceName != service.Name ||
			rem.NamedKey.Namespace != service.Namespace ||
			rem.NamedKey.CID != CLUSTER_ID {
			// skip those which does not belong to this service
			continue
		}
		found := false
		for _, svc := range *local {
			if rem.NamedKey.Port == svc.NamedKey.Port {
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
	}
	return nil
}

//CleanUPVGroupDirect do clean vserver group
func CleanUPVGroupDirect(ctx context.Context, local *vgroups) error {
	for _, vg := range *local {
		if vg.VGroupId == "" {
			err := vg.Describe(ctx)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					// skip none exist vgroup
					vg.Logf("skip none exist vgroup. %s", vg.LoadBalancerId)
					continue
				}
				return err
			}
		}
		err := vg.Remove(ctx)
		if err != nil {
			vg.Logf("Error: cleanup vgroup warining: "+
				"failed to remove vgroup[%s] directly. wait for next try. %s", vg.NamedKey.Key(), err.Error())
			return err
		}
	}
	return nil
}

func BuildVirturalGroupFromService(
	client *LoadBalancerClient,
	service *v1.Service,
	slbins *slb.LoadBalancerType,
) *vgroups {
	vgrps := vgroups{}
	for _, port := range service.Spec.Ports {
		vg := &vgroup{
			NamedKey: &NamedKey{
				CID:         CLUSTER_ID,
				Port:        port.NodePort,
				TargetPort:  port.TargetPort.IntVal,
				Namespace:   service.Namespace,
				ServiceName: service.Name,
				Prefix:      DEFAULT_PREFIX,
			},
			LoadBalancerId: slbins.LoadBalancerId,
			Client:         client.c,
			RegionId:       common.Region(client.region),
			InsClient:      client.ins,
			VpcID:          client.vpcid,
		}
		if utils.IsENIBackendType(service) {
			vg.NamedKey.Port = port.TargetPort.IntVal
		}
		vgrps = append(vgrps, vg)
	}
	// there is no need to delete vserver group.
	return &vgrps
}

func BuildVirtualGroupFromRemoteAPI(
	ctx context.Context,
	lb *slb.LoadBalancerType,
	slbins *LoadBalancerClient,
) (vgroups, error) {
	vgrps := vgroups{}
	vargs := slb.DescribeVServerGroupsArgs{
		RegionId:       common.Region(slbins.region),
		LoadBalancerId: lb.LoadBalancerId,
	}
	vgrp, err := slbins.c.DescribeVServerGroups(ctx, &vargs)
	if err != nil {
		return vgrps, fmt.Errorf("list: vgroup error, %s", err.Error())
	}
	for _, val := range vgrp.VServerGroups.VServerGroup {
		key, err := LoadNamedKey(val.VServerGroupName)
		if err != nil {
			klog.Warningf("we just en-counted an "+
				"unexpected vserver group name: [%s]. Assume user managed "+
				"vserver group, It is ok to skip this vgroup.", val.VServerGroupName)
			continue
		}
		vgrps = append(
			vgrps,
			&vgroup{
				NamedKey:       key,
				LoadBalancerId: lb.LoadBalancerId,
				VpcID:          slbins.vpcid,
				InsClient:      slbins.ins,
				Client:         slbins.c,
				RegionId:       common.Region(slbins.region),
				VGroupId:       val.VServerGroupId,
			},
		)
	}
	return vgrps, nil
}

func findENIbyAddrIP(resp *ecs.DescribeNetworkInterfacesResponse, addrIP string) (string, error) {
	for _, eni := range resp.NetworkInterfaceSets.NetworkInterfaceSet {
		for _, privateIpType := range eni.PrivateIpSets.PrivateIpSet {
			if addrIP == privateIpType.PrivateIpAddress {
				return eni.NetworkInterfaceId, nil
			}
		}
	}
	return "", fmt.Errorf("private ip address not found in openapi %s", addrIP)
}

func findNodeByNodeName(nodes []*v1.Node, nodeName string) *v1.Node {
	for _, n := range nodes {
		if n.Name == nodeName {
			return n
		}
	}
	klog.Infof("node %s not found ", nodeName)
	return nil
}

// EndpointWithENI
// Currently, EndpointWithENI accept two kind of backend
// normal nodes of type []*v1.Node, and endpoints of type *v1.Endpoints
type EndpointWithENI struct {
	// LocalMode externalTraffic=Local
	LocalMode bool
	// BackendTypeENI
	// whether it is an eni backend
	BackendTypeENI bool
	// Nodes
	// contains all the candidate nodes consider of LoadBalance Backends.
	// Cloud implementation has the right to make any filter on it.
	Nodes []*v1.Node

	// Endpoints
	// It is the direct pod location information which cloud implementation
	// may needed for some kind of filtering. eg. direct ENI attach.
	Endpoints *v1.Endpoints
}

// build backend function
func (v *EndpointWithENI) buildFunc(
	ctx context.Context,
	backend *[]slb.VBackendServerType,
	g *vgroup,
) func(o []interface{}) error {

	// backend build function
	return func(o []interface{}) error {
		var ips []string
		for _, i := range o {
			ip, ok := i.(string)
			if !ok {
				return fmt.Errorf("not string: %v", i)
			}
			ips = append(ips, ip)
		}
		targs := &ecs.DescribeNetworkInterfacesArgs{
			VpcID:            g.VpcID,
			RegionId:         g.RegionId,
			PrivateIpAddress: ips,
			PageSize:         50,
		}
		resp, err := g.InsClient.DescribeNetworkInterfaces(ctx, targs)
		if err != nil {
			return fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
		}
		for _, ip := range ips {
			eniid, err := findENIbyAddrIP(resp, ip)
			if err != nil {
				return err
			}
			*backend = append(
				*backend,
				slb.VBackendServerType{
					ServerId:    eniid,
					Weight:      DEFAULT_SERVER_WEIGHT,
					Type:        "eni",
					Port:        int(g.NamedKey.TargetPort),
					ServerIp:    ip,
					Description: g.NamedKey.Key(),
				},
			)
		}
		return nil
	}
}

func (v *EndpointWithENI) BuildBackend(ctx context.Context, g *vgroup) ([]slb.VBackendServerType, error) {
	backend, err := v.doBackendBuild(ctx, g)
	if err != nil {
		return backend, fmt.Errorf("build backend: %s", err.Error())
	}
	return v.nodeWeightWithMerge(backend)
}

func (v *EndpointWithENI) doBackendBuild(ctx context.Context, g *vgroup) ([]slb.VBackendServerType, error) {
	// backend would be modified by buildFunc
	var backend []slb.VBackendServerType

	// ENI Mode
	if v.BackendTypeENI {
		if v.Endpoints == nil {
			klog.Warningf("%s endpoint is nil in eni mode", g.NamedKey)
			return backend, nil
		}
		klog.Infof("[ENI] mode service: %s, endpoint subsets length: %d", g.NamedKey, len(v.Endpoints.Subsets))
		var privateIpAddress []string
		for _, ep := range v.Endpoints.Subsets {
			for _, addr := range ep.Addresses {
				privateIpAddress = append(privateIpAddress, addr.IP)
			}
		}
		err := Batch(privateIpAddress, 40, v.buildFunc(ctx, &backend, g))
		if err != nil {
			return backend, fmt.Errorf("batch process eni fail: %s", err.Error())
		}
		return backend, nil
	}

	// Local Mode
	// When ecs and eci are deployed in a cluster, add ecs first and then add eci
	if v.LocalMode {
		if v.Endpoints == nil {
			klog.Warningf("%s endpoint is nil in local mode", g.NamedKey)
			return backend, nil
		}
		klog.Infof("[Local] mode service: %s,  endpoint subsets length: %d", g.NamedKey, len(v.Endpoints.Subsets))
		// 1. add duplicate ecs backends
		for _, sub := range v.Endpoints.Subsets {
			for _, add := range sub.Addresses {
				if add.NodeName == nil {
					return nil, fmt.Errorf("add ecs backends for service[%s] error, NodeName is nil for ip %s ", g.NamedKey.String(), add.IP)
				}
				node := findNodeByNodeName(v.Nodes, *add.NodeName)
				if node == nil {
					klog.Warningf("can not find correspond node %s for endpoint %s", *add.NodeName, add.IP)
					continue
				}
				if isExcludeNode(node) {
					// filter vk node
					continue
				}
				_, id, err := nodeFromProviderID(node.Spec.ProviderID)
				if err != nil {
					return backend, fmt.Errorf("parse providerid: %s. "+
						"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
				}
				backend = append(
					backend,
					slb.VBackendServerType{
						ServerId:    string(id),
						Weight:      DEFAULT_SERVER_WEIGHT,
						Port:        int(g.NamedKey.Port),
						Type:        "ecs",
						Description: g.NamedKey.Key(),
					},
				)
			}
		}
		// 2. add eci backends
		return v.addECIBackends(ctx, backend, g)
	}

	//Cluster Mode
	// When ecs and eci are deployed in a cluster, add ecs first and then add eci
	klog.Infof("[Cluster] mode service: %s", g.NamedKey)
	// 1. add ecs backends
	for _, node := range v.Nodes {
		if isExcludeNode(node) {
			continue
		}
		_, id, err := nodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return backend, fmt.Errorf("normal parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}

		backend = append(
			backend,
			slb.VBackendServerType{
				ServerId:    string(id),
				Weight:      DEFAULT_SERVER_WEIGHT,
				Port:        int(g.NamedKey.Port),
				Type:        "ecs",
				Description: g.NamedKey.Key(),
			},
		)
	}
	// 2. add eci backends
	return v.addECIBackends(ctx, backend, g)
}

func (v *EndpointWithENI) nodeWeightWithMerge(backends []slb.VBackendServerType) ([]slb.VBackendServerType, error) {

	if !v.LocalMode || v.BackendTypeENI {
		// only local mode should be merged
		return backends, nil
	}

	if len(backends) == 0 {
		return backends, nil
	}

	mergedNode := make(map[string]slb.VBackendServerType)
	for _, b := range backends {
		if _, exist := mergedNode[b.ServerId]; exist {
			updateBackend := mergedNode[b.ServerId]
			updateBackend.Weight += 1
			mergedNode[b.ServerId] = updateBackend
		} else {
			b.Weight = 1
			mergedNode[b.ServerId] = b
		}
	}

	var mergedBackends []slb.VBackendServerType
	for _, v := range mergedNode {
		v.Weight = int(float64(v.Weight) / float64(len(backends)) * DEFAULT_SERVER_WEIGHT)
		if v.Weight < 1 {
			v.Weight = 1
		}
		mergedBackends = append(mergedBackends, v)
	}
	return mergedBackends, nil
}

func (v *EndpointWithENI) addECIBackends(
	ctx context.Context,
	backend []slb.VBackendServerType,
	g *vgroup,
) ([]slb.VBackendServerType, error) {
	var privateIpAddress []string
	// filter ECI nodes
	if v.Endpoints == nil {
		klog.Warningf("%s endpoint is nil in eci mode", g.NamedKey)
		return backend, nil
	}
	for _, sub := range v.Endpoints.Subsets {
		for _, add := range sub.Addresses {
			if add.NodeName == nil {
				return nil, fmt.Errorf("add eci backends for service[%s] error, NodeName is nil for ip %s ", g.NamedKey.String(), add.IP)
			}
			node := findNodeByNodeName(v.Nodes, *add.NodeName)
			if node == nil {
				continue
			}
			// check if the node is ECI
			if node.Labels["type"] == utils.ECINodeLabel {
				klog.Infof("hybrid: %s not an ecs, use eni object as backend", add.IP)
				privateIpAddress = append(privateIpAddress, add.IP)
			}
		}
	}

	// add ENI backends
	if len(privateIpAddress) > 0 {
		err := Batch(privateIpAddress, 40, v.buildFunc(ctx, &backend, g))
		if err != nil {
			return backend, fmt.Errorf("batch process eni fail: %s", err.Error())
		}
	}

	return backend, nil
}

func isExcludeNode(node *v1.Node) bool {
	if _, exclude := node.Labels[utils.LabelNodeRoleExcludeNode]; exclude {
		klog.Infof("ignore node with exclude node label %s", node.Name)
		return true
	}
	if _, exclude := node.Labels[utils.LabelNodeRoleExcludeBalancer]; exclude {
		klog.Infof("ignore node with exclude balancer label %s", node.Name)
		return true
	}
	return false
}
