package ecs

import (
	"context"
	"encoding/json"
	"fmt"

	"strings"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ecsmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/ecs"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/klog/v2"

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
	MaxResult              = 100
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

func (e *ECSProvider) AuthorizeSecurityGroup(ctx context.Context, sgId string, permissions []ecsmodel.SecurityGroupPermission) error {
	if len(permissions) == 0 {
		return nil
	}

	req := ecs.CreateAuthorizeSecurityGroupRequest()
	req.SecurityGroupId = sgId
	ecstPerms := make([]ecs.AuthorizeSecurityGroupPermissions, 0, len(permissions))
	for _, perm := range permissions {
		ecstPerms = append(ecstPerms, ecs.AuthorizeSecurityGroupPermissions{
			Policy:                  perm.Policy,
			Priority:                perm.Priority,
			IpProtocol:              perm.IpProtocol,
			SourceCidrIp:            perm.SourceCidrIp,
			Ipv6SourceCidrIp:        perm.Ipv6SourceCidrIp,
			SourceGroupId:           perm.SourceGroupId,
			SourcePrefixListId:      perm.SourcePrefixListId,
			PortRange:               perm.PortRange,
			DestCidrIp:              perm.DestCidrIp,
			Ipv6DestCidrIp:          perm.Ipv6DestCidrIp,
			SourcePortRange:         perm.SourcePortRange,
			SourceGroupOwnerAccount: perm.SourceGroupOwnerAccount,
			NicType:                 perm.NicType,
			Description:             perm.Description,
		})
	}
	req.Permissions = &ecstPerms

	resp, err := e.auth.ECS.AuthorizeSecurityGroup(req)
	if err != nil {
		klog.Errorf("calling AuthorizeSecurityGroup: securityGroupId=%s, message=[%s].", req.SecurityGroupId, err.Error())
		return err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "AuthorizeSecurityGroup", req.SecurityGroupId)
	return nil
}

func (e *ECSProvider) DescribeSecurityGroupAttribute(ctx context.Context, sgId string) (ecsmodel.SecurityGroup, error) {
	req := ecs.CreateDescribeSecurityGroupAttributeRequest()
	req.SecurityGroupId = sgId

	resp, err := e.auth.ECS.DescribeSecurityGroupAttribute(req)
	if err != nil {
		klog.Errorf("calling DescribeSecurityGroupAttribute: securityGroupId=%s, message=[%s].", sgId, err.Error())
		return ecsmodel.SecurityGroup{}, err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "DescribeSecurityGroupAttribute", sgId)

	var permissions []ecsmodel.SecurityGroupPermission
	for _, perm := range resp.Permissions.Permission {
		permissions = append(permissions, ecsmodel.SecurityGroupPermission{
			SourceGroupId:           perm.SourceGroupId,
			Policy:                  perm.Policy,
			Description:             perm.Description,
			DestPrefixListName:      perm.DestPrefixListName,
			Direction:               perm.Direction,
			SourceCidrIp:            perm.SourceCidrIp,
			SourcePrefixListName:    perm.SourcePrefixListName,
			DestCidrIp:              perm.DestCidrIp,
			Ipv6DestCidrIp:          perm.Ipv6DestCidrIp,
			SourcePortRange:         perm.SourcePortRange,
			Priority:                perm.Priority,
			CreateTime:              perm.CreateTime,
			Ipv6SourceCidrIp:        perm.Ipv6SourceCidrIp,
			NicType:                 perm.NicType,
			DestGroupId:             perm.DestGroupId,
			SourceGroupName:         perm.SourceGroupName,
			PortRange:               perm.PortRange,
			DestGroupOwnerAccount:   perm.DestGroupOwnerAccount,
			DestPrefixListId:        perm.DestPrefixListId,
			IpProtocol:              perm.IpProtocol,
			SecurityGroupRuleId:     perm.SecurityGroupRuleId,
			DestGroupName:           perm.DestGroupName,
			SourceGroupOwnerAccount: perm.SourceGroupOwnerAccount,
			SourcePrefixListId:      perm.SourcePrefixListId,
		})
	}

	return ecsmodel.SecurityGroup{
		Name:              resp.SecurityGroupName,
		ID:                resp.SecurityGroupId,
		InnerAccessPolicy: resp.InnerAccessPolicy,
		Description:       resp.Description,
		Permissions:       permissions,
	}, nil
}

func (e *ECSProvider) DeleteSecurityGroup(ctx context.Context, sgId string) error {
	req := ecs.CreateDeleteSecurityGroupRequest()
	req.SecurityGroupId = sgId

	resp, err := e.auth.ECS.DeleteSecurityGroup(req)
	if err != nil {
		klog.Errorf("calling DeleteSecurityGroup: securityGroupId=%s, message=[%s].", sgId, err.Error())
		return err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "DeleteSecurityGroup", sgId)
	return nil
}

func (e *ECSProvider) DescribeSecurityGroups(ctx context.Context, tags []tag.Tag) ([]ecsmodel.SecurityGroup, error) {
	req := ecs.CreateDescribeSecurityGroupsRequest()

	if len(tags) > 0 {
		ecstags := make([]ecs.DescribeSecurityGroupsTag, 0, len(tags))
		for _, t := range tags {
			ecstags = append(ecstags, ecs.DescribeSecurityGroupsTag{
				Key:   t.Key,
				Value: t.Value,
			})
		}
		req.Tag = &ecstags
	}

	var result []ecsmodel.SecurityGroup
	for {
		resp, err := e.auth.ECS.DescribeSecurityGroups(req)
		if err != nil {
			klog.Errorf("calling DescribeSecurityGroups: message=[%s].", err.Error())
			return nil, err
		}
		klog.V(5).Infof("RequestId: %s, API: %s", resp.RequestId, "DescribeSecurityGroups")

		for _, sg := range resp.SecurityGroups.SecurityGroup {
			result = append(result, ecsmodel.SecurityGroup{
				ID:   sg.SecurityGroupId,
				Name: sg.SecurityGroupName,
				Type: sg.SecurityGroupType,
				Tags: convertSecurityGroupTags(sg.Tags),
			})
		}

		if resp.NextToken == "" {
			break
		}
		req.NextToken = resp.NextToken
	}

	return result, nil
}

func convertSecurityGroupTags(tags ecs.TagsInDescribeSecurityGroups) []tag.Tag {
	var result []tag.Tag
	for _, t := range tags.Tag {
		result = append(result, tag.Tag{
			Key:   t.Key,
			Value: t.Value,
		})
	}
	return result
}

func (e *ECSProvider) CreateSecurityGroup(ctx context.Context, sg ecsmodel.SecurityGroup) error {
	req := ecs.CreateCreateSecurityGroupRequest()
	req.VpcId = sg.VpcID
	req.SecurityGroupName = sg.Name
	req.SecurityGroupType = sg.Type
	req.Description = sg.Description
	req.ResourceGroupId = sg.ResourceGroupID

	if len(sg.Tags) > 0 {
		ecstags := make([]ecs.CreateSecurityGroupTag, 0, len(sg.Tags))
		for _, t := range sg.Tags {
			ecstags = append(ecstags, ecs.CreateSecurityGroupTag{
				Key:   t.Key,
				Value: t.Value,
			})
		}
		req.Tag = &ecstags
	}

	resp, err := e.auth.ECS.CreateSecurityGroup(req)
	if err != nil {
		klog.Errorf("calling CreateSecurityGroup: securityGroupName=%s, message=[%s].", sg.Name, err.Error())
		return err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "CreateSecurityGroup", resp.SecurityGroupId)
	return nil
}

func (e *ECSProvider) RevokeSecurityGroup(ctx context.Context, sgId string, permissions []ecsmodel.SecurityGroupPermission) error {
	if len(permissions) == 0 {
		return nil
	}

	req := ecs.CreateRevokeSecurityGroupRequest()
	req.SecurityGroupId = sgId

	ecstPerms := make([]ecs.RevokeSecurityGroupPermissions, 0, len(permissions))
	for _, perm := range permissions {
		ecstPerms = append(ecstPerms, ecs.RevokeSecurityGroupPermissions{
			Policy:                  perm.Policy,
			Priority:                perm.Priority,
			IpProtocol:              perm.IpProtocol,
			SourceCidrIp:            perm.SourceCidrIp,
			Ipv6SourceCidrIp:        perm.Ipv6SourceCidrIp,
			SourceGroupId:           perm.SourceGroupId,
			SourcePrefixListId:      perm.SourcePrefixListId,
			PortRange:               perm.PortRange,
			DestCidrIp:              perm.DestCidrIp,
			Ipv6DestCidrIp:          perm.Ipv6DestCidrIp,
			SourcePortRange:         perm.SourcePortRange,
			SourceGroupOwnerAccount: perm.SourceGroupOwnerAccount,
			NicType:                 perm.NicType,
			Description:             perm.Description,
		})
	}
	req.Permissions = &ecstPerms

	resp, err := e.auth.ECS.RevokeSecurityGroup(req)
	if err != nil {
		klog.Errorf("calling RevokeSecurityGroup: securityGroupId=%s, message=[%s].", req.SecurityGroupId, err.Error())
		return err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "RevokeSecurityGroup", req.SecurityGroupId)
	return nil
}

func (e *ECSProvider) ModifySecurityGroupAttribute(ctx context.Context, sgId string, sg *ecsmodel.SecurityGroup) error {
	req := ecs.CreateModifySecurityGroupAttributeRequest()
	req.SecurityGroupId = sgId
	req.SecurityGroupName = sg.Name
	req.Description = sg.Description

	resp, err := e.auth.ECS.ModifySecurityGroupAttribute(req)
	if err != nil {
		klog.Errorf("calling ModifySecurityGroupAttribute: securityGroupId=%s, message=[%s].", sgId, err.Error())
		return err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "ModifySecurityGroupAttribute", sgId)
	return nil
}

func (e *ECSProvider) ModifySecurityGroupRule(ctx context.Context, sgId string, permission ecsmodel.SecurityGroupPermission) error {
	req := ecs.CreateModifySecurityGroupRuleRequest()
	req.SecurityGroupId = sgId
	req.SecurityGroupRuleId = permission.SecurityGroupRuleId
	req.Policy = permission.Policy
	req.Description = permission.Description
	req.Priority = permission.Priority
	req.IpProtocol = permission.IpProtocol
	req.PortRange = permission.PortRange
	req.SourcePortRange = permission.SourcePortRange
	req.SourceCidrIp = permission.SourceCidrIp
	req.Ipv6SourceCidrIp = permission.Ipv6SourceCidrIp
	req.DestCidrIp = permission.DestCidrIp
	req.Ipv6DestCidrIp = permission.Ipv6DestCidrIp
	req.SourceGroupId = permission.SourceGroupId
	req.SourceGroupOwnerAccount = permission.SourceGroupOwnerAccount
	req.NicType = permission.NicType
	req.SourcePrefixListId = permission.SourcePrefixListId

	resp, err := e.auth.ECS.ModifySecurityGroupRule(req)
	if err != nil {
		klog.Errorf("calling ModifySecurityGroupRule: securityGroupId=%s, message=[%s].", sgId, err.Error())
		return err
	}
	klog.V(5).Infof("RequestId: %s, API: %s, securityGroupId: %s", resp.RequestId, "ModifySecurityGroupRule", sgId)
	return nil
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

				primaryNetworkInterface := ""
				for _, i := range n.NetworkInterfaces.NetworkInterface {
					if i.Type == "Primary" {
						primaryNetworkInterface = i.NetworkInterfaceId
						break
					}
				}

				mins[id] = &prvd.NodeAttribute{
					InstanceID:                n.InstanceId,
					InstanceType:              n.InstanceType,
					Addresses:                 findAddress(&n),
					Zone:                      n.ZoneId,
					Region:                    n.RegionId,
					InstanceChargeType:        n.InstanceChargeType,
					SpotStrategy:              n.SpotStrategy,
					PrimaryNetworkInterfaceID: primaryNetworkInterface,
					Tags:                      tags,
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
		SpotStrategy:       ins.SpotStrategy,
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
	req.MaxResults = requests.NewInteger(MaxResult)
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
		if resp.NextToken == "" || resp.PageSize < MaxResult {
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
		req.MaxResults = requests.NewInteger(MaxResult)

		for {
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

			if resp.NextToken == "" || resp.PageSize < MaxResult {
				break
			}
			req.NextToken = resp.NextToken
		}

	}
	return result, nil
}

func (e *ECSProvider) DescribeNetworkInterfacesByIDs(ids []string) ([]*prvd.EniAttribute, error) {
	var result []*prvd.EniAttribute
	for begin := 0; begin < len(ids); begin += MaxNetworkInterfaceNum {
		last := len(ids)
		if begin+MaxNetworkInterfaceNum < last {
			last = begin + MaxNetworkInterfaceNum
		}
		networkInterfaceId := ids[begin:last]

		req := ecs.CreateDescribeNetworkInterfacesRequest()
		req.NetworkInterfaceId = &networkInterfaceId

		nextToken := ""
		for {
			req.NextToken = nextToken
			resp, err := e.auth.ECS.DescribeNetworkInterfaces(req)
			if err != nil {
				return result, err
			}
			klog.V(5).Infof("RequestId: %s, API: %s, ips: %s, networkInterfaceIds[%d:%d]",
				resp.RequestId, "DescribeNetworkInterfaces", networkInterfaceId, begin, last)

			for _, eni := range resp.NetworkInterfaceSets.NetworkInterfaceSet {
				eni := &prvd.EniAttribute{
					NetworkInterfaceID: eni.NetworkInterfaceId,
					Status:             eni.Status,
					PrivateIPAddress:   eni.PrivateIpAddress,
					SourceDestCheck:    eni.SourceDestCheck,
				}
				result = append(result, eni)
			}

			if resp.NextToken == "" {
				break
			}
			nextToken = resp.NextToken
		}
	}
	return result, nil
}

func (e *ECSProvider) ModifyNetworkInterfaceSourceDestCheck(id string, enabled bool) error {
	req := ecs.CreateModifyNetworkInterfaceAttributeRequest()
	req.NetworkInterfaceId = id
	req.SourceDestCheck = requests.NewBoolean(enabled)

	resp, err := e.auth.ECS.ModifyNetworkInterfaceAttribute(req)
	if err != nil {
		return err
	}

	klog.V(5).Infof("RequestId: %s, API: %s, sourceDestCheck: %t",
		resp.RequestId, "ModifyNetworkInterfaceAttribute", enabled)

	return nil
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
