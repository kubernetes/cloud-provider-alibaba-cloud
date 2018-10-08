/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"strings"
	"sync"
)

type InstanceClient struct {
	c               ClientInstanceSDK
	lock            sync.RWMutex
	CurrentNodeName types.NodeName
}

type ClientInstanceSDK interface {
	DescribeInstances(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error)
}

// filterOutByRegion Used for multi-region or multi-vpc. works for single region or vpc too.
// SLB only support Backends within the same vpc in the same region. so we need to remove the other backends which not in
// the same region vpc with the SLB. Keep the most backends
func (s *InstanceClient) filterOutByRegion(nodes []*v1.Node, region common.Region) ([]*v1.Node, error) {
	result := []*v1.Node{}
	mvpc := make(map[string]int)
	for _, node := range nodes {
		v, err := s.findInstanceByNode(types.NodeName(node.Name))
		if err != nil {
			return []*v1.Node{}, err
		}
		if v != nil {
			mvpc[v.VpcAttributes.VpcId] = mvpc[v.VpcAttributes.VpcId] + 1
		}
	}
	max, key := 0, ""
	for k, v := range mvpc {
		if v > max {
			max = v
			key = k
		}
	}
	records := []string{}
	for _, node := range nodes {
		v, err := s.findInstanceByNode(types.NodeName(node.Name))
		if err != nil {
			glog.Errorf("alicloud: error find instance by node, retrieve nodes [%s]\n", err.Error())
			return []*v1.Node{}, err
		}
		if v != nil && v.VpcAttributes.VpcId == key {
			result = append(result, node)
			records = append(records, node.Name)
		}
	}
	glog.V(4).Infof("alicloud: accept nodes by region id=[%v], records=%v\n", region, records)
	return result, nil
}

func (s *InstanceClient) filterOutByLabel(nodes []*v1.Node, labels string) ([]*v1.Node, error) {
	if labels == "" {
		// skip filter when label is empty
		glog.V(2).Infof("alicloud: slb backend server label does not specified, skip filter nodes by label.")
		return nodes, nil
	}
	result := []*v1.Node{}
	lbl := strings.Split(labels, ",")
	records := []string{}
	for _, node := range nodes {
		found := true
		for _, v := range lbl {
			l := strings.Split(v, "=")
			if len(l) < 2 {
				msg := fmt.Sprintf("alicloud: error parse backend label with value [%s], must be key value like [k1=v1,k2=v2]\n", v)
				glog.Errorf(msg)
				return []*v1.Node{}, errors.New(msg)
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
	glog.V(4).Infof("alicloud: accept nodes by service backend labels[%s], %v\n", labels, records)
	return result, nil
}

// we use '.' separated nodeid which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7' to identify node
// This is the format of "REGION.NODEID"
func nodeinfo(nodename types.NodeName) (common.Region, types.NodeName, error) {
	name := strings.Split(string(nodename), ".")
	if len(name) < 2 {
		glog.Warningf("alicloud: unable to split instanceid and region from nodename, error unexpected nodename=%s \n", nodename)
		return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from nodename, error unexpected nodename=%s \n", nodename)
	}
	return common.Region(name[0]), types.NodeName(name[1]), nil
}

// we use '.' separated nodeid which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7' to identify node
// This is the format of "REGION.NODEID"
func nodeid(region, nodename string) string {
	return fmt.Sprintf("%s.%s", region, nodename)
}

// getAddressesByName return an instance address slice by it's name.
func (s *InstanceClient) findAddress(nodeName types.NodeName) ([]v1.NodeAddress, error) {

	instance, err := s.findInstanceByNode(nodeName)
	if err != nil {
		glog.Errorf("alicloud: error getting instance by instanceid. instanceid='%s', message=[%s]\n", nodeName, err.Error())
		return nil, err
	}

	addrs := []v1.NodeAddress{}

	if len(instance.PublicIpAddress.IpAddress) > 0 {
		for _, ipaddr := range instance.PublicIpAddress.IpAddress {
			addrs = append(addrs, v1.NodeAddress{Type: v1.NodeExternalIP, Address: ipaddr})
		}
	}

	if instance.EipAddress.IpAddress != "" {
		addrs = append(addrs, v1.NodeAddress{Type: v1.NodeExternalIP, Address: instance.EipAddress.IpAddress})
	}

	if len(instance.InnerIpAddress.IpAddress) > 0 {
		for _, ipaddr := range instance.InnerIpAddress.IpAddress {
			addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: ipaddr})
		}
	}

	if len(instance.VpcAttributes.PrivateIpAddress.IpAddress) > 0 {
		for _, ipaddr := range instance.VpcAttributes.PrivateIpAddress.IpAddress {
			addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: ipaddr})
		}
	}

	if instance.VpcAttributes.NatIpAddress != "" {
		addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: instance.VpcAttributes.NatIpAddress})
	}

	return addrs, nil
}

// nodeName must be a '.' separated Region and ndoeid string which is the instance identity
// Returns instance information
func (s *InstanceClient) findInstanceByNode(nodeName types.NodeName) (*ecs.InstanceAttributesType, error) {
	region, nodeid, err := nodeinfo(nodeName)
	if err != nil {
		return nil, err
	}
	return s.refreshInstance(types.NodeName(nodeid), common.Region(region))
}

func (s *InstanceClient) refreshInstance(nodeName types.NodeName, region common.Region) (*ecs.InstanceAttributesType, error) {
	args := ecs.DescribeInstancesArgs{
		RegionId:    region,
		InstanceIds: fmt.Sprintf("[\"%s\"]", string(nodeName)),
	}

	instances, _, err := s.c.DescribeInstances(&args)
	if err != nil {
		glog.Errorf("alicloud: calling DescribeInstances error. region=%s, instancename=%s, message=[%s]\n", args.RegionId, args.InstanceName, err.Error())
		return nil, err
	}

	if len(instances) == 0 {
		glog.Infof("alicloud: InstanceNotFound, instanceid=[%s.%s]. It is likely to be deleted.", region, nodeName)
		return nil, cloudprovider.InstanceNotFound
	}
	if len(instances) > 1 {
		glog.Warningf("alicloud: multipul instance found by nodename=[%s], the first one will be used, instanceid=[%s]", string(nodeName), instances[0].InstanceId)
	}
	return &instances[0], nil
}
