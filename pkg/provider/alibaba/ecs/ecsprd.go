package ecs

import (
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func NewEcsProvider(
	auth *base.ClientMgr,
) *EcsProvider {
	return &EcsProvider{auth: auth}
}

var _ prvd.IInstance = &EcsProvider{}

type EcsProvider struct {
	auth *base.ClientMgr
}

func (e *EcsProvider) ListInstances(ctx *node.NodeContext, ids []string) (map[string]*prvd.NodeAttribute, error) {
	nodeRegionMap := make(map[string][]string)
	for _, id := range ids {
		regionID, nodeID, err := util.NodeFromProviderID(id)
		if err != nil {
			return nil, err
		}
		nodeRegionMap[regionID] = append(nodeRegionMap[regionID], nodeID)
	}

	var insList []ecs.Instance
	for region, nodes := range nodeRegionMap {
		ins, err := e.getInstances(nodes, region)
		if err != nil {
			return nil, err
		}
		insList = append(insList, ins...)
	}
	mins := make(map[string]*prvd.NodeAttribute)
	for _, id := range ids {
		mins[id] = nil
		for _, n := range insList {
			if strings.Contains(id, n.InstanceId) {
				mins[id] = &prvd.NodeAttribute{
					InstanceID:   n.InstanceId,
					InstanceType: n.InstanceType,
					Addresses:    findAddress(&n),
					Zone:         n.ZoneId,
					Region:       n.RegionId,
				}
				break
			}
		}
	}
	return mins, nil
}

func (e *EcsProvider) getInstances(ids []string, region string) ([]ecs.Instance, error) {
	bids, err := json.Marshal(ids)
	if err != nil {
		return nil, fmt.Errorf("get instances error: %s", err.Error())
	}
	req := ecs.CreateDescribeInstancesRequest()
	req.RegionId = region
	req.InstanceIds = string(bids)
	req.NextToken = ""
	req.MaxResults = requests.NewInteger(50)

	var ecsInstances []ecs.Instance
	for {
		resp, err := e.auth.ECS.DescribeInstances(req)
		if err != nil {
			klog.Errorf("calling DescribeInstances: region=%s, "+
				"instancename=%s, message=[%s].", req.RegionId, req.InstanceName, err.Error())
			return nil, err
		}
		ecsInstances = append(ecsInstances, resp.Instances.Instance...)
		if resp.NextToken == "" {
			break
		}

		req.NextToken = resp.NextToken
	}

	return ecsInstances, nil
}

func (e *EcsProvider) SetInstanceTags(
	ctx *node.NodeContext, id string, tags map[string]string,
) error {
	var mtag []ecs.AddTagsTag
	for k, v := range tags {
		mtag = append(mtag, ecs.AddTagsTag{Key: k, Value: v})
	}
	req := ecs.CreateAddTagsRequest()
	req.ResourceId = id
	req.Tag = &mtag
	req.ResourceType = "instance"

	_, err := e.auth.ECS.AddTags(req)
	return err
}

func (e *EcsProvider) DescribeNetworkInterfaces(vpcId string, ips *[]string, nextToken string) (*ecs.DescribeNetworkInterfacesResponse, error) {
	req := ecs.CreateDescribeNetworkInterfacesRequest()
	req.VpcId = vpcId
	req.PrivateIpAddress = ips
	req.NextToken = nextToken
	req.MaxResults = requests.NewInteger(100)
	return e.auth.ECS.DescribeNetworkInterfaces(req)
}

const (
	DefaultWaitForInterval = 5
)

func findAddress(instance *ecs.Instance) []v1.NodeAddress {
	var addrs []v1.NodeAddress

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
