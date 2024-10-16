package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/klog/v2"
	"strings"

	v1 "k8s.io/api/core/v1"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

const (
	MaxNetworkInterfaceNum = 100
)

func NewECSProvider(
	auth *base.ClientMgr,
) *ECSProvider {
	return &ECSProvider{auth: auth}
}

var _ prvd.IInstance = &ECSProvider{}

type ECSProvider struct {
	auth *base.ClientMgr
}

func (e *ECSProvider) ListInstances(ctx context.Context, ids []string) (map[string]*prvd.NodeAttribute, error) {
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
				tags := map[string]string{}
				for _, tag := range n.Tags.Tag {
					tags[tag.TagKey] = tag.TagValue
				}

				mins[id] = &prvd.NodeAttribute{
					InstanceID:         n.InstanceId,
					InstanceType:       n.InstanceType,
					Addresses:          findAddress(&n),
					Zone:               n.ZoneId,
					Region:             n.RegionId,
					InstanceChargeType: n.InstanceChargeType,
					Tags:               tags,
				}
				break
			}
		}
	}
	return mins, nil
}

func (e *ECSProvider) GetInstancesByIP(ctx context.Context, ips []string) (*prvd.NodeAttribute, error) {
	req := ecs.CreateDescribeInstancesRequest()
	req.InstanceNetworkType = "vpc"
	bips, err := json.Marshal(ips)
	if err != nil {
		return nil, fmt.Errorf("node ips %v marshal error: %s", ips, err.Error())
	}
	req.PrivateIpAddresses = string(bips)
	req.VpcId, err = e.auth.Meta.VpcID()
	if err != nil {
		return nil, fmt.Errorf("get vpc id error: %s", err.Error())
	}
	req.Tag = &[]ecs.DescribeInstancesTag{
		{
			Key: ctrlCfg.CloudCFG.GetKubernetesClusterTag(),
		},
	}
	resp, err := e.auth.ECS.DescribeInstances(req)
	if err != nil {
		klog.V(5).Infof("RequestId: %s, API: %s, ips: %s", resp.RequestId, "DescribeInstances", req.PrivateIpAddresses)
		return nil, fmt.Errorf("describe instances by ip %s error: %s", ips, err.Error())
	}

	if len(resp.Instances.Instance) != 1 {
		klog.V(5).Infof("RequestId: %s, API: %s, ips: %s", resp.RequestId, "DescribeInstances", req.PrivateIpAddresses)
		return nil, fmt.Errorf("find none or multiple instances by ip %s", ips)
	}

	ins := resp.Instances.Instance[0]

	tags := map[string]string{}
	for _, tag := range ins.Tags.Tag {
		tags[tag.TagKey] = tag.TagValue
	}

	return &prvd.NodeAttribute{
		InstanceID:         ins.InstanceId,
		InstanceType:       ins.InstanceType,
		Addresses:          findAddress(&ins),
		Zone:               ins.ZoneId,
		Region:             e.auth.Region,
		InstanceChargeType: ins.InstanceChargeType,
		Tags:               tags,
	}, nil
}

func (e *ECSProvider) getInstances(ids []string, region string) ([]ecs.Instance, error) {
	bids, err := json.Marshal(ids)
	if err != nil {
		return nil, fmt.Errorf("get instances error: %s", err.Error())
	}
	req := ecs.CreateDescribeInstancesRequest()
	req.RegionId = region
	req.InstanceIds = string(bids)
	req.NextToken = ""
	req.MaxResults = requests.NewInteger(50)
	req.AdditionalAttributes = &[]string{"NETWORK_PRIMARY_ENI_IP"}

	var ecsInstances []ecs.Instance
	for {
		resp, err := e.auth.ECS.DescribeInstances(req)
		if err != nil {
			klog.Errorf("calling DescribeInstances: region=%s, "+
				"instancename=%s, message=[%s].", req.RegionId, req.InstanceName, err.Error())
			return nil, err
		}
		klog.V(5).Infof("RequestId: %s, API: %s, ids: %s", resp.RequestId, "DescribeInstances", string(bids))
		ecsInstances = append(ecsInstances, resp.Instances.Instance...)
		if resp.NextToken == "" {
			break
		}

		req.NextToken = resp.NextToken
	}

	return ecsInstances, nil
}

func (e *ECSProvider) DescribeNetworkInterfaces(vpcId string, ips []string, ipVersionType model.AddressIPVersionType) (map[string]string, error) {
	result := make(map[string]string)

	for begin := 0; begin < len(ips); begin += MaxNetworkInterfaceNum {
		last := len(ips)
		if begin+MaxNetworkInterfaceNum < last {
			last = begin + MaxNetworkInterfaceNum
		}
		privateIpAddress := ips[begin:last]

		req := ecs.CreateDescribeNetworkInterfacesRequest()
		req.VpcId = vpcId
		req.Status = "InUse"
		if ipVersionType == model.IPv6 {
			req.Ipv6Address = &privateIpAddress
		} else {
			req.PrivateIpAddress = &privateIpAddress
		}
		next := &util.Pagination{
			PageNumber: 1,
			PageSize:   100,
		}

		for {
			req.PageSize = requests.NewInteger(next.PageSize)
			req.PageNumber = requests.NewInteger(next.PageNumber)
			resp, err := e.auth.ECS.DescribeNetworkInterfaces(req)
			if err != nil {
				return result, err
			}
			klog.V(5).Infof("RequestId: %s, API: %s, ips: %s, privateIpAddress[%d:%d], ipVersionType: %s",
				resp.RequestId, "DescribeNetworkInterfaces", privateIpAddress, begin, last, ipVersionType)

			for _, eni := range resp.NetworkInterfaceSets.NetworkInterfaceSet {

				if ipVersionType == model.IPv6 {
					for _, ipv6 := range eni.Ipv6Sets.Ipv6Set {
						result[ipv6.Ipv6Address] = eni.NetworkInterfaceId
					}
				} else {
					for _, privateIp := range eni.PrivateIpSets.PrivateIpSet {
						result[privateIp.PrivateIpAddress] = eni.NetworkInterfaceId
					}
				}
			}

			pageResult := &util.PaginationResult{
				PageNumber: resp.PageNumber,
				PageSize:   resp.PageSize,
				TotalCount: resp.TotalCount,
			}
			next = pageResult.NextPage()
			if next == nil {
				break
			}
		}

	}
	return result, nil
}

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

	if utilfeature.DefaultFeatureGate.Enabled(ctrlCfg.IPv6DualStack) {
		if len(instance.NetworkInterfaces.NetworkInterface) > 0 {
			var primary *ecs.NetworkInterface
			for i := range instance.NetworkInterfaces.NetworkInterface {
				if instance.NetworkInterfaces.NetworkInterface[i].Type == "Primary" {
					primary = &instance.NetworkInterfaces.NetworkInterface[i]
					break
				}
			}
			if primary != nil {
				// add all ipv6 address of primary network interface to node address
				for _, addr := range primary.Ipv6Sets.Ipv6Set {
					addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: addr.Ipv6Address})
				}
			}
		}
	}

	return addrs
}

func (e *ECSProvider) DeleteInstance(ctx context.Context, id string) error {
	req := ecs.CreateDeleteInstanceRequest()
	req.InstanceId = id
	_, err := e.auth.ECS.DeleteInstance(req)
	if err != nil {
		klog.Errorf("calling DeleteInstance: region=%s, instanceID=%s, message=[%s].", req.RegionId, id, err.Error())
		return err
	}
	return nil
}
