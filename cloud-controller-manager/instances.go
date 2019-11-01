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

	"encoding/json"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/node"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

// InstanceClient wrap for instance sdk
type InstanceClient struct {
	c               ClientInstanceSDK
	lock            sync.RWMutex
	CurrentNodeName types.NodeName
}

// ClientInstanceSDK instance sdk
type ClientInstanceSDK interface {
	AddTags(args *ecs.AddTagsArgs) error
	DescribeInstances(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error)
	DescribeNetworkInterfaces(args *ecs.DescribeNetworkInterfacesArgs) (resp *ecs.DescribeNetworkInterfacesResponse, err error)
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

	return s.findAddressByInstance(instance), nil
}

func (s *InstanceClient) findAddressByInstance(instance *ecs.InstanceAttributesType) []v1.NodeAddress {
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

	return addrs
}

// findAddressByProviderID returns an address slice by it's providerID.
func (s *InstanceClient) findAddressByProviderID(providerID string) ([]v1.NodeAddress, error) {

	instance, err := s.findInstanceByProviderID(providerID)
	if err != nil {
		glog.Errorf("alicloud: error getting instance by providerID. providerID='%s', message=[%s]\n", providerID, err.Error())
		return nil, err
	}

	return s.findAddressByInstance(instance), nil
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
	ins, err := s.getInstances([]string{nodeid}, region)
	if err != nil {
		glog.Errorf("alicloud: InstanceInspectError, instanceid=[%s.%s]. message=[%s]\n", region, nodeid, err.Error())
		return nil, err
	}

	if len(ins) == 0 {
		glog.Infof("alicloud: InstanceNotFound, instanceid=[%s.%s]. It is likely to be deleted.\n", region, nodeid)
		return nil, cloudprovider.InstanceNotFound
	}
	if len(ins) > 1 {
		glog.Warningf("alicloud: multipul instance found by nodename=[%s], "+
			"the first one will be used, instanceid=[%s]\n", string(nodeid), ins[0].InstanceId)
	}
	return &ins[0], nil
}

func (s *InstanceClient) ListInstances(ids []string) (map[string]*node.CloudNodeAttribute, error) {
	var (
		idsp   []string
		region common.Region
	)
	for _, id := range ids {
		regionid, nodeid, err := nodeFromProviderID(id)
		if err != nil {
			return nil, err
		}
		idsp = append(idsp, nodeid)
		region = regionid
	}
	ins, err := s.getInstances(idsp, region)
	if err != nil {
		return nil, err
	}
	mins := make(map[string]*node.CloudNodeAttribute)
	for _, id := range ids {
		mins[id] = nil
		for _, n := range ins {
			if strings.Contains(id, n.InstanceId) {
				mins[id] = &node.CloudNodeAttribute{
					InstanceID:   n.InstanceId,
					InstanceType: n.InstanceType,
					Addresses:    s.findAddressByInstance(&n),
				}
				break
			}
		}
	}
	return mins, nil
}

func (s *InstanceClient) getInstances(ids []string, region common.Region) ([]ecs.InstanceAttributesType, error) {
	bids, err := json.Marshal(ids)
	if err != nil {
		return nil, fmt.Errorf("get instances error: %s", err.Error())
	}
	args := ecs.DescribeInstancesArgs{
		RegionId:    region,
		InstanceIds: string(bids),
		Pagination:  common.Pagination{PageSize: 50},
	}

	instances, _, err := s.c.DescribeInstances(&args)
	if err != nil {
		glog.Errorf("alicloud: calling DescribeInstances error. region=%s, "+
			"instancename=%s, message=[%s].\n", args.RegionId, args.InstanceName, err.Error())
		return nil, err
	}
	return instances, nil
}

func (s *InstanceClient) AddCloudTags(id string, tags map[string]string, region common.Region) error {
	args := &ecs.AddTagsArgs{
		ResourceId:   id,
		Tag:          tags,
		RegionId:     region,
		ResourceType: ecs.TagResourceInstance,
	}
	return s.c.AddTags(args)
}
