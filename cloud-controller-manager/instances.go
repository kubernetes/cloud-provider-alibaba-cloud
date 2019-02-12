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
	"strings"
	"sync"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
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
		v, err := s.findInstanceByProviderID(node.Spec.ProviderID)
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
		v, err := s.findInstanceByProviderID(node.Spec.ProviderID)
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

// Use '.' to separate providerID which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7'. The format of "REGION.NODEID"
func nodeFromProviderID(providerID string) (common.Region, string, error) {
	name := strings.Split(providerID, ".")
	if len(name) < 2 {
		return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
	}
	return common.Region(name[0]), name[1], nil
}

// we use '.' separated nodeid which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7' to identify node
// This is the format of "REGION.NODEID"
func nodeid(region, nodename string) string {
	return fmt.Sprintf("%s.%s", region, nodename)
}

// findAddressByNodeName returns an address slice by it's host name.
func (s *InstanceClient) findAddressByNodeName(nodeName types.NodeName) ([]v1.NodeAddress, error) {
	instance, err := s.findInstanceByNodeName(nodeName)
	if err != nil {
		glog.Errorf("alicloud: error getting instance by nodeName. providerID='%s', message=[%s]\n", nodeName, err.Error())
		return nil, err
	}

	return s.findAddressByInstance(instance)
}

func (s *InstanceClient) findAddressByInstance(instance *ecs.InstanceAttributesType) ([]v1.NodeAddress, error) {
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

// findAddressByProviderID returns an address slice by it's providerID.
func (s *InstanceClient) findAddressByProviderID(providerID string) ([]v1.NodeAddress, error) {

	instance, err := s.findInstanceByProviderID(providerID)
	if err != nil {
		glog.Errorf("alicloud: error getting instance by providerID. providerID='%s', message=[%s]\n", providerID, err.Error())
		return nil, err
	}

	return s.findAddressByInstance(instance)
}

// Returns instance information. Currently, we had a constraint that the name should has the same as provider id for compatibility concern.
func (s *InstanceClient) findInstanceByNodeName(nodeName types.NodeName) (*ecs.InstanceAttributesType, error) {
	return s.findInstanceByProviderID(string(nodeName))
}

func (s *InstanceClient) findInstanceByProviderID(providerID string) (*ecs.InstanceAttributesType, error) {
	region, nodeid, err := nodeFromProviderID(providerID)
	if err != nil {
		return nil, err
	}
	return s.refreshInstance(nodeid, common.Region(region))
}

func (s *InstanceClient) refreshInstance(nodeID string, region common.Region) (*ecs.InstanceAttributesType, error) {
	args := ecs.DescribeInstancesArgs{
		RegionId:    region,
		InstanceIds: fmt.Sprintf("[\"%s\"]", nodeID),
	}

	instances, _, err := s.c.DescribeInstances(&args)
	if err != nil {
		glog.Errorf("alicloud: calling DescribeInstances error. region=%s, instancename=%s, message=[%s].\n", args.RegionId, args.InstanceName, err.Error())
		return nil, err
	}

	if len(instances) == 0 {
		glog.Infof("alicloud: InstanceNotFound, instanceid=[%s.%s]. It is likely to be deleted.\n", region, nodeID)
		return nil, cloudprovider.InstanceNotFound
	}
	if len(instances) > 1 {
		glog.Warningf("alicloud: multipul instance found by nodename=[%s], the first one will be used, instanceid=[%s]\n", string(nodeID), instances[0].InstanceId)
	}
	return &instances[0], nil
}
