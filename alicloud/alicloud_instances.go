package alicloud

import (
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"strings"
)

type SDKClientINS struct {
	regions         map[string][]string
	c               *ecs.Client
	CurrentNodeName types.NodeName
}

func NewSDKClientINS(access_key_id string, access_key_secret string) *SDKClientINS {
	ins := &SDKClientINS{
		c: ecs.NewClient(access_key_id, access_key_secret),
	}
	ins.c.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)
	return ins
}

// filterOutByRegion Used for multi-region or multi-vpc. works for single region or vpc too.
// SLB only support Backends within the same vpc in the same region. so we need to remove the other backends which not in
// the same region vpc with teh SLB. Keep the most backends
func (s *SDKClientINS) filterOutByRegion(nodes []*v1.Node, region common.Region) []*v1.Node {
	result := []*v1.Node{}
	mvpc := make(map[string]int)
	for _, node := range nodes {
		glog.V(2).Infof("Alicloud.filterOutByRegion(): for each node=%v\n", node.Name)
		v, err := s.findInstanceByNode(types.NodeName(node.Name))
		if err != nil {
			glog.Errorf("Alicloud.filterOutByRegion(): error execute c.ins.doFindInstance(). message=[%s]\n", err.Error())
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
			glog.Errorf("Alicloud.filterOutByRegion(): error execute findInstanceByNode(). message=[%s]\n", err.Error())
			continue
		}
		if v != nil && v.VpcAttributes.VpcId == key {
			result = append(result, node)
			glog.V(2).Infof("Alicloud.filterOutByRegion(): accept node=%v\n", node.Name)
		}
	}
	return result
}

func (s *SDKClientINS) filterOutByLabel(nodes []*v1.Node, labels string) []*v1.Node {
	result := []*v1.Node{}
	lbl := strings.Split(labels, ",")
	for _, node := range nodes{
		found := true
		for _,v := range lbl{
			if _,exist := node.Labels[v]; !exist{
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
func nodeid(nodename types.NodeName) (common.Region, types.NodeName, error) {
	name := strings.Split(string(nodename), ".")
	if len(name) < 2 {
		glog.Warningf("Alicloud: unable to split instanceid and region from nodename, error unexpected nodename=%s \n", nodename)
		return "", "", errors.New(fmt.Sprintf("Alicloud: unable to split instanceid and region from nodename, error unexpected nodename=%s \n", nodename))
	}
	return common.Region(name[0]), types.NodeName(name[1]), nil
}

// getAddressesByName return an instance address slice by it's name.
func (s *SDKClientINS) findAddress(nodeName types.NodeName) ([]v1.NodeAddress, error) {

	instance, err := s.findInstanceByNode(nodeName)
	if err != nil {
		glog.Errorf("Alicloud: error getting instance by instanceid. instanceid='%s', message=[%s]\n", nodeName, err.Error())
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

// nodeName must be a ':' separated Region and ndoeid string which is the instance identity
// Returns instance information
func (s *SDKClientINS) findInstanceByNode(nodeName types.NodeName) (*ecs.InstanceAttributesType, error) {
	region, nodeid, err := nodeid(nodeName)
	if err != nil {
		return nil, err
	}
	return s.refreshInstance(types.NodeName(nodeid), common.Region(region))
}

func (s *SDKClientINS) Regions() map[string][]string {
	return s.regions
}

func (s *SDKClientINS) refreshInstance(nodeName types.NodeName, region common.Region) (*ecs.InstanceAttributesType, error) {
	args := ecs.DescribeInstancesArgs{
		RegionId:    region,
		InstanceIds: fmt.Sprintf("[\"%s\"]", string(nodeName)),
	}

	instances, _, err := s.c.DescribeInstances(&args)
	if err != nil {
		glog.Errorf("Alicloud: calling DescribeInstances error. region=%s, instancename=%s, message=[%s]\n", args.RegionId, args.InstanceName, err.Error())
		return nil, err
	}

	if len(instances) == 0 {
		glog.Infof("Alicloud: InstanceNotFound, instanceid=[%s.%s]. It is likely to be deleted.", nodeName, region)
		return nil, cloudprovider.InstanceNotFound
	}
	if len(instances) > 1 {
		glog.Warningf("Alicloud: multipul instance found by nodename=[%s], the first one will be used, instanceid=[%s]", string(nodeName), instances[0].InstanceId)
	}

	s.storeVpcid(&instances[0])
	return &instances[0], nil
}

func (s *SDKClientINS) storeVpcid(i *ecs.InstanceAttributesType) {
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
