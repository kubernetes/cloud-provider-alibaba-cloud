package alicloud

import (
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
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
		prefix = fmt.Sprintf("%s/%s", v.NamedKey.Namespace, v.NamedKey.ServiceName)
	}
	format = fmt.Sprintf("%s %s", prefix, format)
	glog.Infof(format, args)
}

func (v *vgroup) Describe() error {
	if v.NamedKey == nil {
		return fmt.Errorf("describe: format error of vgroup name")
	}
	vargs := slb.DescribeVServerGroupsArgs{
		RegionId:       v.RegionId,
		LoadBalancerId: v.LoadBalancerId,
	}
	vgrp, err := v.Client.DescribeVServerGroups(&vargs)
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
func (v *vgroup) Add() error {
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
	gp, err := v.Client.CreateVServerGroup(&vgp)
	if err != nil {
		return fmt.Errorf("CreateVServerGroup. %s", err.Error())
	}
	v.Logf("create new vserver group[%s]"+
		" for loadbalancer[%s] with empty backend list", v.NamedKey.Key(), v.LoadBalancerId)
	v.VGroupId = gp.VServerGroupId
	return nil
}
func (v *vgroup) Remove() error {
	if v.LoadBalancerId == "" || v.VGroupId == "" {
		return fmt.Errorf("can not delete vserver group. LoadBalancerId or vgroup id should not be empty")
	}
	vdel := slb.DeleteVServerGroupArgs{
		VServerGroupId: v.VGroupId,
		RegionId:       v.RegionId,
	}
	_, err := v.Client.DeleteVServerGroup(&vdel)
	return err
}
func (v *vgroup) Update() error {
	if v.VGroupId == "" {
		err := v.Describe()
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("update: vserver group error, %s", err.Error())
			}
			if err := v.Add(); err != nil {
				return err
			}
		}
	}

	v.Logf("update: backend vgroupid [%s]", v.VGroupId)
	dsc := &slb.DescribeVServerGroupAttributeArgs{
		VServerGroupId: v.VGroupId,
		RegionId:       v.RegionId,
	}
	att, err := v.Client.DescribeVServerGroupAttribute(dsc)
	if err != nil {
		return fmt.Errorf("update: describe vserver group attribute error. %s", err.Error())
	}
	v.Logf("update: apis[%v], node[%v]", att.BackendServers.BackendServer, v.BackendServers)
	add, del := v.diff(att.BackendServers.BackendServer, v.BackendServers)
	if len(add) == 0 && len(del) == 0 {
		v.Logf("update: no backend need to be added for vgroupid [%s]", v.VGroupId)
		return nil
	}

	if len(add) > 0 {
		if err := BatchProcess(add,
			func(list []slb.VBackendServerType) error {
				additions, err := json.Marshal(list)
				if err != nil {
					return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
				}
				v.Logf("update: try to update vserver group[%s],"+
					" backend add[%s]", v.NamedKey.Key(), string(additions))
				_, err = v.Client.AddVServerGroupBackendServers(
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
		return BatchProcess(del,
			func(list []slb.VBackendServerType) error {
				deletions, err := json.Marshal(list)
				if err != nil {
					return fmt.Errorf("error marshal backends: %s, %v", err.Error(), list)
				}
				v.Logf("update: try to update vserver group[%s],"+
					" backend del[%s]", v.NamedKey.Key(), string(deletions))
				_, err = v.Client.RemoveVServerGroupBackendServers(
					&slb.RemoveVServerGroupBackendServersArgs{
						VServerGroupId: v.VGroupId,
						RegionId:       v.RegionId,
						BackendServers: string(deletions),
					})
				return err
			})
	}
	return nil
}

// MAX_BACKEND_NUM max batch backend num
const MAX_BACKEND_NUM = 19

//BatchProcess batch update backend.
func BatchProcess(list []slb.VBackendServerType,
	batch func(list []slb.VBackendServerType) error) error {

	glog.Infof("batch process virtual server backend, length %d", len(list))
	for len(list) > MAX_BACKEND_NUM {
		if err := batch(list[0:MAX_BACKEND_NUM]); err != nil {

			return err
		}
		list = list[MAX_BACKEND_NUM:]
	}
	if len(list) <= 0 {
		return nil
	}
	return batch(list)
}

func (v *vgroup) diff(apis, nodes []slb.VBackendServerType) (
	[]slb.VBackendServerType, []slb.VBackendServerType) {

	addition, deletions := []slb.VBackendServerType{}, []slb.VBackendServerType{}
	for _, api := range apis {
		found := false
		for _, node := range nodes {
			if api.ServerId == node.ServerId &&
				api.ServerIp == node.ServerIp {
				found = true
				break
			}
		}
		if !found {
			deletions = append(deletions, api)
		}
	}
	for _, node := range nodes {
		found := false
		for _, api := range apis {
			if api.ServerId == node.ServerId &&
				api.ServerIp == node.ServerIp {
				found = true
				break
			}
		}
		if !found {
			addition = append(addition, node)
		}
	}
	return addition, deletions
}

func Ensure(v *vgroup, nodes interface{}) error {
	var backend []slb.VBackendServerType
	switch nodes.(type) {
	case []*v1.Node:
		for _, node := range nodes.([]*v1.Node) {
			_, id, err := nodeFromProviderID(node.Spec.ProviderID)
			if err != nil {
				return fmt.Errorf("ensure: error parse providerid. %s. expected: ${regionid}.${nodeid}", node.Spec.ProviderID)
			}
			backend = append(backend, slb.VBackendServerType{
				ServerId: string(id),
				Weight:   100,
				Port:     int(v.NamedKey.Port),
				Type:     "ecs",
			})
		}
	case *v1.Endpoints:
		targs := &ecs.DescribeNetworkInterfacesArgs{
			VpcId:    v.VpcID,
			RegionId: v.RegionId,
		}
		resp, err := v.InsClient.DescribeNetworkInterfaces(targs)
		if err != nil {
			return fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
		}
		for _, ep := range nodes.(*v1.Endpoints).Subsets {
			for _, addr := range ep.Addresses {
				found := false
				eniid := ""
				for _, eni := range resp.NetworkInterfaceSets.NetworkInterfaceSet {
					for _, ip := range eni.PrivateIpSets.PrivateIpSet {
						if addr.IP == ip.PrivateIpAddress {
							found = true
							eniid = eni.NetworkInterfaceId
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					return fmt.Errorf("private ip address not found in openapi %s", addr.IP)
				}
				if eniid == "" {
					return fmt.Errorf("unexpected empty eni id found %s", addr.IP)
				}
				backend = append(
					backend,
					slb.VBackendServerType{
						ServerId: eniid,
						Weight:   100,
						Type:     "eni",
						Port:     int(v.NamedKey.Port),
						ServerIp: addr.IP,
					},
				)
			}
		}
	default:
		return fmt.Errorf("unknown backend type, %s", reflect.TypeOf(nodes))
	}
	v.BackendServers = backend
	return v.Update()
}

type vgroups []*vgroup

func EnsureVirtualGroups(vgrps *vgroups, nodes interface{}) error {
	glog.Infof("ensure vserver group: %d vgroup need to be processed.", len(*vgrps))
	for _, v := range *vgrps {
		if v == nil {
			return fmt.Errorf("unexpected nil vgroup ")
		}
		if err := Ensure(v, nodes); err != nil {
			return fmt.Errorf("ensure vgroup: %s. %s", err.Error(), v.NamedKey.Key())
		}
		v.Logf("EnsureGroup: id=[%s], Name:[%s], LoadBalancerId:[%s]", v.VGroupId, v.NamedKey.Key(), v.LoadBalancerId)
	}
	return nil
}

//CleanUPVGroupMerged Merge with service port and do clean vserver group
func CleanUPVGroupMerged(
	slbins *LoadBalancerClient,
	service *v1.Service,
	lb *slb.LoadBalancerType,
	local *vgroups,
) error {

	remote, err := BuildVirtualGroupFromRemoteAPI(lb, slbins)
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
			err := rem.Remove()
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
func CleanUPVGroupDirect(local *vgroups) error {
	for _, vg := range *local {
		if vg.VGroupId == "" {
			err := vg.Describe()
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					// skip none exist vgroup
					vg.Logf("skip none exist vgroup. %s", vg.LoadBalancerId)
					continue
				}
				return err
			}
		}
		err := vg.Remove()
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
		if isEniBackend(service) {
			vg.NamedKey.Port = port.Port
		}
		vgrps = append(vgrps, vg)
	}
	// there is no need to delete vserver group.
	return &vgrps
}

func BuildVirtualGroupFromRemoteAPI(
	lb *slb.LoadBalancerType,
	slbins *LoadBalancerClient,
) (vgroups, error) {
	vgrps := vgroups{}
	vargs := slb.DescribeVServerGroupsArgs{
		RegionId:       common.Region(slbins.region),
		LoadBalancerId: lb.LoadBalancerId,
	}
	vgrp, err := slbins.c.DescribeVServerGroups(&vargs)
	if err != nil {
		return vgrps, fmt.Errorf("list: vgroup error, %s", err.Error())
	}
	for _, val := range vgrp.VServerGroups.VServerGroup {
		key, err := LoadNamedKey(val.VServerGroupName)
		if err != nil {
			glog.Warningf("we just en-counted an "+
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
