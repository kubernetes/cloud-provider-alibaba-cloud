package ecs

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
)

const (
	SecurityGroupTypeEnterprise = "enterprise"

	SecurityGroupPolicyAccept = "accept"
	SecurityGroupPolicyDrop   = "drop"

	SecurityGroupIpProtocolAll = "All"
)

type SecurityGroupPermission struct {
	SourceGroupId           string
	Policy                  string
	Description             string
	DestPrefixListName      string
	Direction               string
	SourceCidrIp            string
	SourcePrefixListName    string
	DestCidrIp              string
	Ipv6DestCidrIp          string
	SourcePortRange         string
	Priority                string
	CreateTime              string
	Ipv6SourceCidrIp        string
	NicType                 string
	DestGroupId             string
	SourceGroupName         string
	PortRange               string
	DestGroupOwnerAccount   string
	DestPrefixListId        string
	IpProtocol              string
	SecurityGroupRuleId     string
	DestGroupName           string
	SourceGroupOwnerAccount string
	SourcePrefixListId      string
}

type SecurityGroup struct {
	ID                string
	Name              string
	Description       string
	Type              string
	ResourceGroupID   string
	VpcID             string
	InnerAccessPolicy string
	Permissions       []SecurityGroupPermission
	Tags              []tag.Tag
}
