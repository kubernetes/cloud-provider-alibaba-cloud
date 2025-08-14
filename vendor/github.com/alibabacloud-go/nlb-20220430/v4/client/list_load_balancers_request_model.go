// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListLoadBalancersRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAddressIpVersion(v string) *ListLoadBalancersRequest
	GetAddressIpVersion() *string
	SetAddressType(v string) *ListLoadBalancersRequest
	GetAddressType() *string
	SetDNSName(v string) *ListLoadBalancersRequest
	GetDNSName() *string
	SetIpv6AddressType(v string) *ListLoadBalancersRequest
	GetIpv6AddressType() *string
	SetLoadBalancerBusinessStatus(v string) *ListLoadBalancersRequest
	GetLoadBalancerBusinessStatus() *string
	SetLoadBalancerIds(v []*string) *ListLoadBalancersRequest
	GetLoadBalancerIds() []*string
	SetLoadBalancerNames(v []*string) *ListLoadBalancersRequest
	GetLoadBalancerNames() []*string
	SetLoadBalancerStatus(v string) *ListLoadBalancersRequest
	GetLoadBalancerStatus() *string
	SetLoadBalancerType(v string) *ListLoadBalancersRequest
	GetLoadBalancerType() *string
	SetMaxResults(v int32) *ListLoadBalancersRequest
	GetMaxResults() *int32
	SetNextToken(v string) *ListLoadBalancersRequest
	GetNextToken() *string
	SetRegionId(v string) *ListLoadBalancersRequest
	GetRegionId() *string
	SetResourceGroupId(v string) *ListLoadBalancersRequest
	GetResourceGroupId() *string
	SetTag(v []*ListLoadBalancersRequestTag) *ListLoadBalancersRequest
	GetTag() []*ListLoadBalancersRequestTag
	SetVpcIds(v []*string) *ListLoadBalancersRequest
	GetVpcIds() []*string
	SetZoneId(v string) *ListLoadBalancersRequest
	GetZoneId() *string
}

type ListLoadBalancersRequest struct {
	// The IP version of the NLB instance. Valid values:
	//
	// 	- **ipv4**: IPv4
	//
	// 	- **DualStack**: dual-stack
	//
	// example:
	//
	// ipv4
	AddressIpVersion *string `json:"AddressIpVersion,omitempty" xml:"AddressIpVersion,omitempty"`
	// The type of IPv4 address used by the NLB instance. Valid values:
	//
	// 	- **Internet**: The NLB instance uses a public IP address. The domain name of the NLB instance is resolved to the public IP address. The NLB instance can be accessed over the Internet.
	//
	// 	- **Intranet**: The NLB instance uses a private IP address. The domain name of the NLB instance is resolved to the private IP address. The NLB instance can be accessed over the VPC where the NLB instance is deployed.
	//
	// example:
	//
	// Internet
	AddressType *string `json:"AddressType,omitempty" xml:"AddressType,omitempty"`
	// The domain name of the NLB instance.
	//
	// example:
	//
	// nlb-wb7r6dlwetvt5j****.cn-hangzhou.nlb.aliyuncs.com
	DNSName *string `json:"DNSName,omitempty" xml:"DNSName,omitempty"`
	// The type of IPv6 address used by the NLB instance. Valid values:
	//
	// 	- **Internet**: The NLB instance uses a public IP address. The domain name of the NLB instance is resolved to the public IP address. The NLB instance can be accessed over the Internet.
	//
	// 	- **Intranet**: The NLB instance uses a private IP address. The domain name of the NLB instance is resolved to the private IP address. The NLB instance can be accessed over the VPC where the NLB instance is deployed.
	//
	// example:
	//
	// Internet
	Ipv6AddressType *string `json:"Ipv6AddressType,omitempty" xml:"Ipv6AddressType,omitempty"`
	// The business status of the NLB instance. Valid values:
	//
	// 	- **Abnormal**: The NLB instance is not working as expected.
	//
	// 	- **Normal**: The NLB instance is working as expected.
	//
	// example:
	//
	// Normal
	LoadBalancerBusinessStatus *string `json:"LoadBalancerBusinessStatus,omitempty" xml:"LoadBalancerBusinessStatus,omitempty"`
	// The NLB instance IDs. You can specify up to 20 IDs in each call.
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" xml:"LoadBalancerIds,omitempty" type:"Repeated"`
	// The names of the NLB instances. You can specify up to 20 names in each call.
	LoadBalancerNames []*string `json:"LoadBalancerNames,omitempty" xml:"LoadBalancerNames,omitempty" type:"Repeated"`
	// The status of the NLB instance. Valid values:
	//
	// 	- **Inactive**: The NLB instance is disabled. Listeners of an NLB instance in the Inactive state do not forward traffic.
	//
	// 	- **Active**: The NLB instance is running.
	//
	// 	- **Provisioning**: The NLB instance is being created.
	//
	// 	- **Configuring**: The NLB instance is being modified.
	//
	// 	- **Deleting**: The NLB instance is being deleted.
	//
	// 	- **Deleted**: The NLB instance is deleted.
	//
	// example:
	//
	// Active
	LoadBalancerStatus *string `json:"LoadBalancerStatus,omitempty" xml:"LoadBalancerStatus,omitempty"`
	// The type of the Server Load Balancer (SLB) instances. Set the value to **network**, which specifies NLB.
	//
	// example:
	//
	// network
	LoadBalancerType *string `json:"LoadBalancerType,omitempty" xml:"LoadBalancerType,omitempty"`
	// The number of entries to return in each call. Valid values: **1*	- to **100**. Default value: **20**.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The pagination token used to specify a particular page of results. Valid values:
	//
	// 	- Leave this parameter empty for the first query or the only query.
	//
	// 	- Set this parameter to the value of NextToken obtained from the previous query.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The ID of the resource group to which the instance belongs.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The tags of the NLB instance.
	Tag []*ListLoadBalancersRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
	// The IDs of the virtual private clouds (VPCs) where the NLB instances are deployed. You can specify up to 10 VPC IDs in each call.
	VpcIds []*string `json:"VpcIds,omitempty" xml:"VpcIds,omitempty" type:"Repeated"`
	// The ID of the zone. You can call the [DescribeZones](https://help.aliyun.com/document_detail/443890.html) operation to query the most recent zone list.
	//
	// example:
	//
	// cn-hangzhou-a
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s ListLoadBalancersRequest) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersRequest) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersRequest) GetAddressIpVersion() *string {
	return s.AddressIpVersion
}

func (s *ListLoadBalancersRequest) GetAddressType() *string {
	return s.AddressType
}

func (s *ListLoadBalancersRequest) GetDNSName() *string {
	return s.DNSName
}

func (s *ListLoadBalancersRequest) GetIpv6AddressType() *string {
	return s.Ipv6AddressType
}

func (s *ListLoadBalancersRequest) GetLoadBalancerBusinessStatus() *string {
	return s.LoadBalancerBusinessStatus
}

func (s *ListLoadBalancersRequest) GetLoadBalancerIds() []*string {
	return s.LoadBalancerIds
}

func (s *ListLoadBalancersRequest) GetLoadBalancerNames() []*string {
	return s.LoadBalancerNames
}

func (s *ListLoadBalancersRequest) GetLoadBalancerStatus() *string {
	return s.LoadBalancerStatus
}

func (s *ListLoadBalancersRequest) GetLoadBalancerType() *string {
	return s.LoadBalancerType
}

func (s *ListLoadBalancersRequest) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListLoadBalancersRequest) GetNextToken() *string {
	return s.NextToken
}

func (s *ListLoadBalancersRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *ListLoadBalancersRequest) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *ListLoadBalancersRequest) GetTag() []*ListLoadBalancersRequestTag {
	return s.Tag
}

func (s *ListLoadBalancersRequest) GetVpcIds() []*string {
	return s.VpcIds
}

func (s *ListLoadBalancersRequest) GetZoneId() *string {
	return s.ZoneId
}

func (s *ListLoadBalancersRequest) SetAddressIpVersion(v string) *ListLoadBalancersRequest {
	s.AddressIpVersion = &v
	return s
}

func (s *ListLoadBalancersRequest) SetAddressType(v string) *ListLoadBalancersRequest {
	s.AddressType = &v
	return s
}

func (s *ListLoadBalancersRequest) SetDNSName(v string) *ListLoadBalancersRequest {
	s.DNSName = &v
	return s
}

func (s *ListLoadBalancersRequest) SetIpv6AddressType(v string) *ListLoadBalancersRequest {
	s.Ipv6AddressType = &v
	return s
}

func (s *ListLoadBalancersRequest) SetLoadBalancerBusinessStatus(v string) *ListLoadBalancersRequest {
	s.LoadBalancerBusinessStatus = &v
	return s
}

func (s *ListLoadBalancersRequest) SetLoadBalancerIds(v []*string) *ListLoadBalancersRequest {
	s.LoadBalancerIds = v
	return s
}

func (s *ListLoadBalancersRequest) SetLoadBalancerNames(v []*string) *ListLoadBalancersRequest {
	s.LoadBalancerNames = v
	return s
}

func (s *ListLoadBalancersRequest) SetLoadBalancerStatus(v string) *ListLoadBalancersRequest {
	s.LoadBalancerStatus = &v
	return s
}

func (s *ListLoadBalancersRequest) SetLoadBalancerType(v string) *ListLoadBalancersRequest {
	s.LoadBalancerType = &v
	return s
}

func (s *ListLoadBalancersRequest) SetMaxResults(v int32) *ListLoadBalancersRequest {
	s.MaxResults = &v
	return s
}

func (s *ListLoadBalancersRequest) SetNextToken(v string) *ListLoadBalancersRequest {
	s.NextToken = &v
	return s
}

func (s *ListLoadBalancersRequest) SetRegionId(v string) *ListLoadBalancersRequest {
	s.RegionId = &v
	return s
}

func (s *ListLoadBalancersRequest) SetResourceGroupId(v string) *ListLoadBalancersRequest {
	s.ResourceGroupId = &v
	return s
}

func (s *ListLoadBalancersRequest) SetTag(v []*ListLoadBalancersRequestTag) *ListLoadBalancersRequest {
	s.Tag = v
	return s
}

func (s *ListLoadBalancersRequest) SetVpcIds(v []*string) *ListLoadBalancersRequest {
	s.VpcIds = v
	return s
}

func (s *ListLoadBalancersRequest) SetZoneId(v string) *ListLoadBalancersRequest {
	s.ZoneId = &v
	return s
}

func (s *ListLoadBalancersRequest) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersRequestTag struct {
	// The key of the tag. You can specify up to 20 tags. The tag key cannot be an empty string.
	//
	// It must be 1 to 64 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// KeyTest
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. You can specify up to 20 tags. The tag value can be an empty string.
	//
	// The tag value can be up to 128 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// ValueTest
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListLoadBalancersRequestTag) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersRequestTag) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersRequestTag) GetKey() *string {
	return s.Key
}

func (s *ListLoadBalancersRequestTag) GetValue() *string {
	return s.Value
}

func (s *ListLoadBalancersRequestTag) SetKey(v string) *ListLoadBalancersRequestTag {
	s.Key = &v
	return s
}

func (s *ListLoadBalancersRequestTag) SetValue(v string) *ListLoadBalancersRequestTag {
	s.Value = &v
	return s
}

func (s *ListLoadBalancersRequestTag) Validate() error {
	return dara.Validate(s)
}
