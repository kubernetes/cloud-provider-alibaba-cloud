package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"fmt"
	"k8s.io/api/core/v1"
	"strings"
	"k8s.io/apimachinery/pkg/types"
	"github.com/golang/glog"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
)

type vgroup struct {
	NamedKey 			*NamedKey
	LoadBalancerId   	string
	RegionId         	common.Region
	VGroupId   			string
	Client 				ClientSLBSDK
	BackendServers   	[]slb.VBackendServerType
}

func (v *vgroup) Describe() error {
	if v.NamedKey == nil {
		return fmt.Errorf("describe: format error of vgroup name")
	}
	vargs := slb.DescribeVServerGroupsArgs{
		RegionId:  v.RegionId,
		LoadBalancerId:  v.LoadBalancerId,
	}
	vgrp , err := v.Client.DescribeVServerGroups(&vargs)
	if err != nil {
		return fmt.Errorf("decribe: vgroup error, %s",err.Error())
	}
	if vgrp != nil {
		for _,val :=range vgrp.VServerGroups.VServerGroup {
			if val.VServerGroupName ==
				v.NamedKey.Key() {
				v.VGroupId = val.VServerGroupId
				return nil
			}
		}
	}
	return fmt.Errorf("vgroup not found, %s",v.NamedKey.Key())
}
func (v *vgroup) Add() 	 	error{
	if v.VGroupId != "" {
		return fmt.Errorf("vgroupid already exists")
	}
	if v.NamedKey == nil {
		return fmt.Errorf("format error of vgroup name")
	}
	backends, err := json.Marshal(v.BackendServers)
	if err != nil {
		return fmt.Errorf("add: error marshal backends")
	}
	vgp := slb.CreateVServerGroupArgs{
		LoadBalancerId: v.LoadBalancerId,
		VServerGroupName: v.NamedKey.Key(),
		RegionId: 		  v.RegionId,
		BackendServers:   string(backends),
	}
	gp,err := v.Client.CreateVServerGroup(&vgp)
	if err != nil {
		return fmt.Errorf("CreateVServerGroup. %s",err.Error())
	}
	glog.Infof("create new vserver group[%s]" +
		" for loadbalancer[%s] with backend list count[%s]",v.NamedKey.Key(),v.LoadBalancerId,len(v.BackendServers))
	v.VGroupId = gp.VServerGroupId
	return nil
}
func (v *vgroup) Remove() 	error{
	if v.LoadBalancerId == "" || v.VGroupId == "" {
		return fmt.Errorf("can not delete vserver group. LoadBalancerId or vgroup id should not be empty")
	}
	vdel := slb.DeleteVServerGroupArgs{
		VServerGroupId:  v.VGroupId,
		RegionId:        v.RegionId,
	}
	_, err := v.Client.DeleteVServerGroup(&vdel)
	return err
}
func (v *vgroup) Update() 	error{
	if v.VGroupId == "" {
		err := v.Describe()
		if err != nil {
			if ! strings.Contains(err.Error(),"not found") {
				return fmt.Errorf("update: vserver group error, %s",err.Error())
			}
			return v.Add()
		}
	}

	glog.Infof("update: backend vgroupid [%s]",v.VGroupId)
	dsc := &slb.DescribeVServerGroupAttributeArgs{
		VServerGroupId: v.VGroupId,
		RegionId: 		v.RegionId,
	}
	att, err := v.Client.DescribeVServerGroupAttribute(dsc)
	if err != nil {
		return fmt.Errorf("update: describe vserver group attribute error. %s",err.Error())
	}
	glog.Infof("update: apis[%v], node[%v]",att.BackendServers.BackendServer, v.BackendServers)
	add, del := v.diff(att.BackendServers.BackendServer, v.BackendServers)
	if len(add) == 0 && len(del) == 0 {
		glog.Infof("update: no backend need to be added for vgroupid [%s]",v.VGroupId)
		return nil
	}
	additions, err := json.Marshal(add)
	if err != nil {
		return fmt.Errorf("error marshal backends: %s, %v",err.Error(),add)
	}
	deletions, err := json.Marshal(del)
	if err != nil {
		return fmt.Errorf("error marshal backends: %s, %v",err.Error(),del)
	}

	vgp := slb.ModifyVServerGroupBackendServersArgs{
		VServerGroupId: 		v.VGroupId,
		OldBackendServers:      string(deletions),
		NewBackendServers:      string(additions),
	}
	glog.Infof("update: try to update vserver group[%s]," +
		" backend new[%s], old[%s]",v.NamedKey.Key(),string(additions),string(deletions))
	_, err = v.Client.ModifyVServerGroupBackendServers(&vgp)
	return err
}

func (v *vgroup) diff(apis, nodes []slb.VBackendServerType) (
										[]slb.VBackendServerType, []slb.VBackendServerType) {

	addition, deletions := []slb.VBackendServerType{},[]slb.VBackendServerType{}
	for _, api := range apis {
		found := false
		for _, node := range nodes {
			if api.ServerId == node.ServerId {
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
			if api.ServerId == node.ServerId {
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

func (v *vgroup) Ensure(nodes []*v1.Node) error {

	var backend []slb.VBackendServerType
	for _, node := range nodes {
		_, id,err := nodeinfo(types.NodeName(node.Spec.ProviderID))
		if err != nil {
			return fmt.Errorf("ensure: error parse providerid. %s. expected: ${regionid}.${nodeid}",node.Spec.ProviderID)
		}
		backend = append(backend, slb.VBackendServerType{
			ServerId: string(id),
			Weight:   100,
			Port:  	  int(v.NamedKey.Port),
			Type: 	  "ecs",
		})
	}
	v.BackendServers = backend
	return v.Update()
}

type vgroups []*vgroup

func (vgrps *vgroups) EnsureVGroup(nodes []*v1.Node) error {
	glog.Infof("ensure vserver group: %d vgroup need to be processed.",len(*vgrps))
	for _,v := range *vgrps {
		if v == nil {
			return fmt.Errorf("unexpected nil vgroup ")
		}
		if err := v.Ensure(nodes);err != nil {
			return fmt.Errorf("ensure vgroup: %s. %s",err.Error(),v.NamedKey.Key())
		}
		glog.Infof("EnsureGroup: id=[%s], Name:[%s], LoadBalancerId:[%s]",v.VGroupId, v.NamedKey.Key(),v.LoadBalancerId)
	}
	return nil
}

// Merge with service port and do clean vserver group
func CleanUPVGroupMerged(service *v1.Service,
					lb *slb.LoadBalancerType,
					client ClientSLBSDK, local *vgroups) error{

	remote,err := buildVGroupFromRemoteAPI(lb, client,lb.RegionId)
	if err != nil {
		return fmt.Errorf("build vserver group from remote: %s",err.Error())
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
			glog.Infof("try to remove unused vserver group, [%s][%s]",rem.NamedKey.Key(), rem.VGroupId)
			err := rem.Remove()
			if err != nil {
				glog.Errorf("cleanup vgroup warining: " +
					"failed to remove vgroup[%s] is ok. wait for next try. %s",rem.NamedKey.Key(),err.Error())
				return err
			}
		}
	}
	return nil
}

// Merge with service port and do clean vserver group
func CleanUPVGroupDirect(local *vgroups) error{
	for _, vg := range *local {
		if vg.VGroupId == "" {
			err := vg.Describe()
			if err != nil {
				return err
			}
		}
		err := vg.Remove()
		if err != nil {
			glog.Errorf("cleanup vgroup warining: " +
				"failed to remove vgroup[%s] is ok. wait for next try. %s",vg.NamedKey.Key(),err.Error())
			return err
		}
	}
	return nil
}

func buildVGroupFromService(service *v1.Service,
				lb *slb.LoadBalancerType,
				client ClientSLBSDK,
				region common.Region) (*vgroups) {
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
			LoadBalancerId: lb.LoadBalancerId,
			Client:         client,
			RegionId:       region,
		}
		vgrps = append(vgrps, vg)
	}
	// there is no need to delete vserver group.
	return &vgrps
}


func buildVGroupFromRemoteAPI(lb *slb.LoadBalancerType,
						client ClientSLBSDK,
						region common.Region) (vgroups,error) {
	vgrps := vgroups{}
	vargs := slb.DescribeVServerGroupsArgs{
		RegionId:  region,
		LoadBalancerId:  lb.LoadBalancerId,
	}
	vgrp , err := client.DescribeVServerGroups(&vargs)
	if err != nil {
		return vgrps, fmt.Errorf("list: vgroup error, %s",err.Error())
	}
	for _,val :=range vgrp.VServerGroups.VServerGroup {
		key, err := LoadNamedKey(val.VServerGroupName)
		if err != nil {
			glog.Warningf("we just en-counted an " +
				"unexpected vserver group name: [%s]. Assume user managed vserver group, It is ok to skip this vgroup.",val.VServerGroupName)
			continue
		}
		vgrps = append(vgrps, &vgroup{
			NamedKey: 		key,
			LoadBalancerId: lb.LoadBalancerId,
			Client:         client,
			RegionId:       region,
			VGroupId:       val.VServerGroupId,
		})
	}
	return vgrps, nil
}
