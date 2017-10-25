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

// we use '.' separated nodeid which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7' to identify node
// This is the format of "REGION.NODEID"
func nodeid(nodename types.NodeName) (common.Region, types.NodeName, error) {
	name := strings.Split(string(nodename), ".")
	if len(name) < 2 {
		glog.Warningf("Unexpected nodename: %s \n", nodename)
		return "", "", errors.New(fmt.Sprintf("Alicloud: nodeid Unexpected, nodename=%s \n", nodename))
	}
	return common.Region(name[0]), types.NodeName(name[1]), nil
}

// getAddressesByName return an instance address slice by it's name.
func (s *SDKClientINS) findAddress(nodeName types.NodeName) ([]v1.NodeAddress, error) {

	instance, err := s.findInstanceByNode(nodeName)
	if err != nil {
		glog.Errorf("Error getting instance by InstanceId '%s': %v", nodeName, err)
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
	glog.V(2).Infof("Alicloud.findInstanceByNode(\"%s\")", nodeName)
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
		glog.Errorf("refreshInstance: DescribeInstances error=%s, region=%s, instanceName=%s", err.Error(), args.RegionId, args.InstanceName)
		//s.Set(fmt.Sprintf("%s/%s",nodeName,region),nil)
		return nil, err
	}

	if len(instances) == 0 {
		glog.Infof("refreshInstance: InstanceNotFound, [%s.%s]", nodeName, region)
		return nil, cloudprovider.InstanceNotFound
	}
	if len(instances) > 1 {
		glog.Errorf("Warning: Multipul instance found by nodename [%s], the first one will be used. Instance: [%+v]", string(nodeName), instances)
	}
	glog.V(2).Infof("Alicloud.refreshInstance(\"%s\") finished. [ %+v ]\n", string(nodeName), instances[0])
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
