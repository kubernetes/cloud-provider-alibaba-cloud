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
	regions         map[string][]string
	c               ClientInstanceSDK
	lock            sync.RWMutex
	CurrentNodeName types.NodeName
}

type ClientInstanceSDK interface {
	DescribeInstances(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error)
}

// filterOutByRegion Used for multi-region or multi-vpc. works for single region or vpc too.
// SLB only support Backends within the same vpc in the same region. so we need to remove the other backends which not in
// the same region vpc with teh SLB. Keep the most backends
func (s *InstanceClient) filterOutByRegion(nodes []*v1.Node, region common.Region) []*v1.Node {
	result := []*v1.Node{}
	mvpc := make(map[string]int)
	for _, node := range nodes {
		v, err := s.findInstanceByNode(types.NodeName(node.Name))
		if err != nil {
			glog.Errorf("alicloud: error execute c.ins.doFindInstance(). message=[%s]\n", err.Error())
			continue
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
	for _, node := range nodes {
		v, err := s.findInstanceByNode(types.NodeName(node.Name))
		if err != nil {
			glog.Errorf("alicloud: error execute findInstanceByNode(). message=[%s]\n", err.Error())
			continue
		}
		if v != nil && v.VpcAttributes.VpcId == key {
			result = append(result, node)
			glog.V(2).Infof("alicloud: accept node=%v\n", node.Name)
		}
	}
	return result
}

func (s *InstanceClient) filterOutByLabel(nodes []*v1.Node, labels string) []*v1.Node {
	if labels == "" {
		// skip filter when label is empty
		glog.V(2).Infof("alicloud: slb backend server label doesnot specified, skip filter nodes by label.")
		return nodes
	}
	result := []*v1.Node{}
	lbl := strings.Split(labels, ",")
	for _, node := range nodes {
		found := true
		for _, v := range lbl {
			l := strings.Split(v, "=")
			if len(l) < 2 {
				glog.Errorf("alicloud: error parse backend label with value [%s], must be key value like [k=v]\n", v)
				return []*v1.Node{}
			}
			if nv, exist := node.Labels[l[0]]; !exist || nv != l[1] {
				found = false
				break
			}
		}
		if found {
			result = append(result, node)
		}
	}
	return result
}

// we use '.' separated nodeid which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7' to identify node
// This is the format of "REGION.NODEID"
func nodeinfo(nodename types.NodeName) (common.Region, types.NodeName, error) {
	name := strings.Split(string(nodename), ".")
	if len(name) < 2 {
		glog.Warningf("alicloud: unable to split instanceid and region from nodename, error unexpected nodename=%s \n", nodename)
		return "", "", errors.New(fmt.Sprintf("alicloud: unable to split instanceid and region from nodename, error unexpected nodename=%s \n", nodename))
	}
	return common.Region(name[0]), types.NodeName(name[1]), nil
}

// we use '.' separated nodeid which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7' to identify node
// This is the format of "REGION.NODEID"
func nodeid(region, nodename string) string {
	return fmt.Sprintf("%s.%s",region,nodename)
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

func (s *InstanceClient) Regions() map[string][]string {
	return s.regions
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

	s.storeVpcid(&instances[0])
	return &instances[0], nil
}

func (s *InstanceClient) storeVpcid(i *ecs.InstanceAttributesType) {
	if s.regions == nil {
		s.regions = make(map[string][]string)
	}
	if v, e := s.regions[string(i.RegionId)]; !e {
		s.regions[string(i.RegionId)] = []string{i.VpcAttributes.VpcId}
	} else {
		found := false
		for _, n := range v {
			if n == i.VpcAttributes.VpcId {
				found = true
				break
			}
		}
		if !found {
			vpcs := s.regions[string(i.RegionId)]
			s.regions[string(i.RegionId)] = append(vpcs, i.VpcAttributes.VpcId)
		}
	}
}
