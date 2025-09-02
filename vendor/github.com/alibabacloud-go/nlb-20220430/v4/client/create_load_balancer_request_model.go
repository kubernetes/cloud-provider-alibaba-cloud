// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateLoadBalancerRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAddressIpVersion(v string) *CreateLoadBalancerRequest
	GetAddressIpVersion() *string
	SetAddressType(v string) *CreateLoadBalancerRequest
	GetAddressType() *string
	SetBandwidthPackageId(v string) *CreateLoadBalancerRequest
	GetBandwidthPackageId() *string
	SetClientToken(v string) *CreateLoadBalancerRequest
	GetClientToken() *string
	SetDeletionProtectionConfig(v *CreateLoadBalancerRequestDeletionProtectionConfig) *CreateLoadBalancerRequest
	GetDeletionProtectionConfig() *CreateLoadBalancerRequestDeletionProtectionConfig
	SetDryRun(v bool) *CreateLoadBalancerRequest
	GetDryRun() *bool
	SetLoadBalancerBillingConfig(v *CreateLoadBalancerRequestLoadBalancerBillingConfig) *CreateLoadBalancerRequest
	GetLoadBalancerBillingConfig() *CreateLoadBalancerRequestLoadBalancerBillingConfig
	SetLoadBalancerName(v string) *CreateLoadBalancerRequest
	GetLoadBalancerName() *string
	SetLoadBalancerType(v string) *CreateLoadBalancerRequest
	GetLoadBalancerType() *string
	SetModificationProtectionConfig(v *CreateLoadBalancerRequestModificationProtectionConfig) *CreateLoadBalancerRequest
	GetModificationProtectionConfig() *CreateLoadBalancerRequestModificationProtectionConfig
	SetRegionId(v string) *CreateLoadBalancerRequest
	GetRegionId() *string
	SetResourceGroupId(v string) *CreateLoadBalancerRequest
	GetResourceGroupId() *string
	SetTag(v []*CreateLoadBalancerRequestTag) *CreateLoadBalancerRequest
	GetTag() []*CreateLoadBalancerRequestTag
	SetVpcId(v string) *CreateLoadBalancerRequest
	GetVpcId() *string
	SetZoneMappings(v []*CreateLoadBalancerRequestZoneMappings) *CreateLoadBalancerRequest
	GetZoneMappings() []*CreateLoadBalancerRequestZoneMappings
}

type CreateLoadBalancerRequest struct {
	// The version of IP addresses used for the NLB instance. Valid values:
	//
	// 	- **ipv4*	- (default)
	//
	// 	- **DualStack**
	//
	// example:
	//
	// ipv4
	AddressIpVersion *string `json:"AddressIpVersion,omitempty" xml:"AddressIpVersion,omitempty"`
	// The type of IPv4 addresses used for the NLB instance. Valid values are:
	//
	// 	- **Internet**: The nodes of an Internet-facing NLB instance have public IP addresses. The DNS name of an Internet-facing NLB instance is publicly resolvable to the public IP addresses of the nodes. Therefore, Internet-facing NLB instances can route requests from clients over the Internet.
	//
	// 	- **Intranet**: The nodes of an internal-facing NLB instance have only private IP addresses. The DNS name of an internal-facing NLB instance is publicly resolvable to the private IP addresses of the nodes. Therefore, internal-facing NLB instances can route requests only from clients with access to the virtual private cloud (VPC) for the NLB instance.
	//
	// >  To enable a public IPv6 address for a dual-stack NLB instance, call the [EnableLoadBalancerIpv6Internet](https://help.aliyun.com/document_detail/445878.html) operation.
	//
	// This parameter is required.
	//
	// example:
	//
	// Internet
	AddressType *string `json:"AddressType,omitempty" xml:"AddressType,omitempty"`
	// The ID of the Internet Shared Bandwidth instance that is associated with the Internet-facing NLB instance.
	//
	// example:
	//
	// cbwp-bp1vevu8h3ieh****
	BandwidthPackageId *string `json:"BandwidthPackageId,omitempty" xml:"BandwidthPackageId,omitempty"`
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. Only ASCII characters are allowed.
	//
	// >  If you do not specify this parameter, the value of **RequestId*	- is used.***	- The value of **RequestId*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// The configuration of the deletion protection feature.
	DeletionProtectionConfig *CreateLoadBalancerRequestDeletionProtectionConfig `json:"DeletionProtectionConfig,omitempty" xml:"DeletionProtectionConfig,omitempty" type:"Struct"`
	// Perform a dry run without actually making the request. Valid values are:
	//
	// 	- **true**: Perform only a dry run. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the check, an error message specifying the issue is returned. If the request passes, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): Check the request and perform the operation. If the request passes the check, a 2xx HTTP status code is returned, and the operation is performed.
	//
	// example:
	//
	// false
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The billing settings of the NLB instance.
	LoadBalancerBillingConfig *CreateLoadBalancerRequestLoadBalancerBillingConfig `json:"LoadBalancerBillingConfig,omitempty" xml:"LoadBalancerBillingConfig,omitempty" type:"Struct"`
	// The name of the NLB instance.
	//
	// It must be 2 to 128 characters in length, can contain letters, digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// example:
	//
	// NLB1
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" xml:"LoadBalancerName,omitempty"`
	// The type of the Server Load Balancer (SLB) instance. Set the value to **network**, which indicates an NLB instance.
	//
	// example:
	//
	// network
	LoadBalancerType *string `json:"LoadBalancerType,omitempty" xml:"LoadBalancerType,omitempty"`
	// The configuration of the configuration read-only mode.
	ModificationProtectionConfig *CreateLoadBalancerRequestModificationProtectionConfig `json:"ModificationProtectionConfig,omitempty" xml:"ModificationProtectionConfig,omitempty" type:"Struct"`
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
	// The tags.
	//
	// if can be null:
	// true
	Tag []*CreateLoadBalancerRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
	// The ID of the VPC where you want to create the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// vpc-bp1b49rqrybk45nio****
	VpcId *string `json:"VpcId,omitempty" xml:"VpcId,omitempty"`
	// The mappings between zones and vSwitches. An NLB instance can be deployed in up to 10 zones. If the region supports two or more zones, you must specify at least two zones.
	//
	// This parameter is required.
	ZoneMappings []*CreateLoadBalancerRequestZoneMappings `json:"ZoneMappings,omitempty" xml:"ZoneMappings,omitempty" type:"Repeated"`
}

func (s CreateLoadBalancerRequest) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerRequest) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerRequest) GetAddressIpVersion() *string {
	return s.AddressIpVersion
}

func (s *CreateLoadBalancerRequest) GetAddressType() *string {
	return s.AddressType
}

func (s *CreateLoadBalancerRequest) GetBandwidthPackageId() *string {
	return s.BandwidthPackageId
}

func (s *CreateLoadBalancerRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *CreateLoadBalancerRequest) GetDeletionProtectionConfig() *CreateLoadBalancerRequestDeletionProtectionConfig {
	return s.DeletionProtectionConfig
}

func (s *CreateLoadBalancerRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *CreateLoadBalancerRequest) GetLoadBalancerBillingConfig() *CreateLoadBalancerRequestLoadBalancerBillingConfig {
	return s.LoadBalancerBillingConfig
}

func (s *CreateLoadBalancerRequest) GetLoadBalancerName() *string {
	return s.LoadBalancerName
}

func (s *CreateLoadBalancerRequest) GetLoadBalancerType() *string {
	return s.LoadBalancerType
}

func (s *CreateLoadBalancerRequest) GetModificationProtectionConfig() *CreateLoadBalancerRequestModificationProtectionConfig {
	return s.ModificationProtectionConfig
}

func (s *CreateLoadBalancerRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *CreateLoadBalancerRequest) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *CreateLoadBalancerRequest) GetTag() []*CreateLoadBalancerRequestTag {
	return s.Tag
}

func (s *CreateLoadBalancerRequest) GetVpcId() *string {
	return s.VpcId
}

func (s *CreateLoadBalancerRequest) GetZoneMappings() []*CreateLoadBalancerRequestZoneMappings {
	return s.ZoneMappings
}

func (s *CreateLoadBalancerRequest) SetAddressIpVersion(v string) *CreateLoadBalancerRequest {
	s.AddressIpVersion = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetAddressType(v string) *CreateLoadBalancerRequest {
	s.AddressType = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetBandwidthPackageId(v string) *CreateLoadBalancerRequest {
	s.BandwidthPackageId = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetClientToken(v string) *CreateLoadBalancerRequest {
	s.ClientToken = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetDeletionProtectionConfig(v *CreateLoadBalancerRequestDeletionProtectionConfig) *CreateLoadBalancerRequest {
	s.DeletionProtectionConfig = v
	return s
}

func (s *CreateLoadBalancerRequest) SetDryRun(v bool) *CreateLoadBalancerRequest {
	s.DryRun = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetLoadBalancerBillingConfig(v *CreateLoadBalancerRequestLoadBalancerBillingConfig) *CreateLoadBalancerRequest {
	s.LoadBalancerBillingConfig = v
	return s
}

func (s *CreateLoadBalancerRequest) SetLoadBalancerName(v string) *CreateLoadBalancerRequest {
	s.LoadBalancerName = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetLoadBalancerType(v string) *CreateLoadBalancerRequest {
	s.LoadBalancerType = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetModificationProtectionConfig(v *CreateLoadBalancerRequestModificationProtectionConfig) *CreateLoadBalancerRequest {
	s.ModificationProtectionConfig = v
	return s
}

func (s *CreateLoadBalancerRequest) SetRegionId(v string) *CreateLoadBalancerRequest {
	s.RegionId = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetResourceGroupId(v string) *CreateLoadBalancerRequest {
	s.ResourceGroupId = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetTag(v []*CreateLoadBalancerRequestTag) *CreateLoadBalancerRequest {
	s.Tag = v
	return s
}

func (s *CreateLoadBalancerRequest) SetVpcId(v string) *CreateLoadBalancerRequest {
	s.VpcId = &v
	return s
}

func (s *CreateLoadBalancerRequest) SetZoneMappings(v []*CreateLoadBalancerRequestZoneMappings) *CreateLoadBalancerRequest {
	s.ZoneMappings = v
	return s
}

func (s *CreateLoadBalancerRequest) Validate() error {
	return dara.Validate(s)
}

type CreateLoadBalancerRequestDeletionProtectionConfig struct {
	// Specifies whether to enable the deletion protection feature. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	Enabled *bool `json:"Enabled,omitempty" xml:"Enabled,omitempty"`
	// The reason why the deletion protection feature is enabled or disabled. The reason must be 2 to 128 characters in length, can contain letters, digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// example:
	//
	// The instance is running
	Reason *string `json:"Reason,omitempty" xml:"Reason,omitempty"`
}

func (s CreateLoadBalancerRequestDeletionProtectionConfig) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerRequestDeletionProtectionConfig) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerRequestDeletionProtectionConfig) GetEnabled() *bool {
	return s.Enabled
}

func (s *CreateLoadBalancerRequestDeletionProtectionConfig) GetReason() *string {
	return s.Reason
}

func (s *CreateLoadBalancerRequestDeletionProtectionConfig) SetEnabled(v bool) *CreateLoadBalancerRequestDeletionProtectionConfig {
	s.Enabled = &v
	return s
}

func (s *CreateLoadBalancerRequestDeletionProtectionConfig) SetReason(v string) *CreateLoadBalancerRequestDeletionProtectionConfig {
	s.Reason = &v
	return s
}

func (s *CreateLoadBalancerRequestDeletionProtectionConfig) Validate() error {
	return dara.Validate(s)
}

type CreateLoadBalancerRequestLoadBalancerBillingConfig struct {
	// The billing method of the NLB instance.
	//
	// Set the value to **PostPay**, which specifies the pay-as-you-go billing method.
	//
	// example:
	//
	// PostPay
	PayType *string `json:"PayType,omitempty" xml:"PayType,omitempty"`
}

func (s CreateLoadBalancerRequestLoadBalancerBillingConfig) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerRequestLoadBalancerBillingConfig) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerRequestLoadBalancerBillingConfig) GetPayType() *string {
	return s.PayType
}

func (s *CreateLoadBalancerRequestLoadBalancerBillingConfig) SetPayType(v string) *CreateLoadBalancerRequestLoadBalancerBillingConfig {
	s.PayType = &v
	return s
}

func (s *CreateLoadBalancerRequestLoadBalancerBillingConfig) Validate() error {
	return dara.Validate(s)
}

type CreateLoadBalancerRequestModificationProtectionConfig struct {
	// The reason for enabling the configuration read-only mode. The reason must be 2 to 128 characters in length, can contain letters, digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// >  This parameter takes effect only when **Status*	- is set to **ConsoleProtection**.
	//
	// example:
	//
	// Service guarantee period
	Reason *string `json:"Reason,omitempty" xml:"Reason,omitempty"`
	// Specifies whether to enable the configuration read-only mode. Valid values:
	//
	// 	- **NonProtection**: does not enable the configuration read-only mode. You cannot set the **Reason*	- parameter. If the **Reason*	- parameter is set, the value is cleared.
	//
	// 	- **ConsoleProtection**: enables the configuration read-only mode. You can set the **Reason*	- parameter.
	//
	// >  If the parameter is set to **ConsoleProtection**, the configuration read-only mode is enabled. You cannot modify the configurations of the NLB instance in the NLB console. However, you can call API operations to modify the configurations of the NLB instance.
	//
	// example:
	//
	// ConsoleProtection
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s CreateLoadBalancerRequestModificationProtectionConfig) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerRequestModificationProtectionConfig) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerRequestModificationProtectionConfig) GetReason() *string {
	return s.Reason
}

func (s *CreateLoadBalancerRequestModificationProtectionConfig) GetStatus() *string {
	return s.Status
}

func (s *CreateLoadBalancerRequestModificationProtectionConfig) SetReason(v string) *CreateLoadBalancerRequestModificationProtectionConfig {
	s.Reason = &v
	return s
}

func (s *CreateLoadBalancerRequestModificationProtectionConfig) SetStatus(v string) *CreateLoadBalancerRequestModificationProtectionConfig {
	s.Status = &v
	return s
}

func (s *CreateLoadBalancerRequestModificationProtectionConfig) Validate() error {
	return dara.Validate(s)
}

type CreateLoadBalancerRequestTag struct {
	// The key of the tag. The tag key can be up to 64 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`. The tag key can contain letters, digits, and the following special characters: _ . : / = + - @
	//
	// You can specify up to 20 tags in each call.
	//
	// example:
	//
	// env
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. The tag value can be up to 128 characters in length, cannot start with `acs:` or `aliyun`, and cannot contain `http://` or `https://`. The tag value can contain letters, digits, and the following special characters: _ . : / = + - @
	//
	// You can specify up to 20 tags in each call.
	//
	// example:
	//
	// product
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s CreateLoadBalancerRequestTag) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerRequestTag) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerRequestTag) GetKey() *string {
	return s.Key
}

func (s *CreateLoadBalancerRequestTag) GetValue() *string {
	return s.Value
}

func (s *CreateLoadBalancerRequestTag) SetKey(v string) *CreateLoadBalancerRequestTag {
	s.Key = &v
	return s
}

func (s *CreateLoadBalancerRequestTag) SetValue(v string) *CreateLoadBalancerRequestTag {
	s.Value = &v
	return s
}

func (s *CreateLoadBalancerRequestTag) Validate() error {
	return dara.Validate(s)
}

type CreateLoadBalancerRequestZoneMappings struct {
	// The ID of the elastic IP address (EIP) that is associated with the Internet-facing NLB instance. Each zone is assigned one EIP. An NLB instance can be deployed in up to 10 zones. If the region supports two or more zones, specify at least two zones.
	//
	// example:
	//
	// eip-bp1aedxso6u80u0qf****
	AllocationId *string `json:"AllocationId,omitempty" xml:"AllocationId,omitempty"`
	// The local IPv4 addresses. The IP addresses that the NLB instance uses to communicate with the backend servers. The number of IP addresses must be an even number, which must be at least 2 and at most 8.
	Ipv4LocalAddresses []*string `json:"Ipv4LocalAddresses,omitempty" xml:"Ipv4LocalAddresses,omitempty" type:"Repeated"`
	// The VIP of the IPv6 version. The IPv6 address that the NLB instance uses to provide external services.
	//
	// example:
	//
	// 2408:400a:d5:3080:b409:840a:ca:e8e5
	Ipv6Address *string `json:"Ipv6Address,omitempty" xml:"Ipv6Address,omitempty"`
	// The local IPv6 addresses. The IP addresses that the NLB instance uses to communicate with the backend servers. The number of IP addresses must be an even number, which must be at least 2 and at most 8.
	Ipv6LocalAddresses []*string `json:"Ipv6LocalAddresses,omitempty" xml:"Ipv6LocalAddresses,omitempty" type:"Repeated"`
	// The private virtual IP address (VIP) of the IPv4 version. The private IPv4 address that the NLB instance uses to provide external services.
	//
	// example:
	//
	// 192.168.10.1
	PrivateIPv4Address *string `json:"PrivateIPv4Address,omitempty" xml:"PrivateIPv4Address,omitempty"`
	// The ID of the vSwitch in the zone. You can specify only one vSwitch (subnet) in each zone of an NLB instance. An NLB instance can be deployed in up to 10 zones. If the region supports two or more zones, you must specify at least two zones.
	//
	// This parameter is required.
	//
	// example:
	//
	// vsw-sersdf****
	VSwitchId *string `json:"VSwitchId,omitempty" xml:"VSwitchId,omitempty"`
	// The ID of the zone where the NLB instance is deployed. An NLB instance can be deployed in up to 10 zones. If the region supports two or more zones, specify at least two zones.
	//
	// You can call the [DescribeZones](https://help.aliyun.com/document_detail/443890.html) operation to query the most recent zone list.
	//
	// This parameter is required.
	//
	// example:
	//
	// cn-hangzhou-a
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s CreateLoadBalancerRequestZoneMappings) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerRequestZoneMappings) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerRequestZoneMappings) GetAllocationId() *string {
	return s.AllocationId
}

func (s *CreateLoadBalancerRequestZoneMappings) GetIpv4LocalAddresses() []*string {
	return s.Ipv4LocalAddresses
}

func (s *CreateLoadBalancerRequestZoneMappings) GetIpv6Address() *string {
	return s.Ipv6Address
}

func (s *CreateLoadBalancerRequestZoneMappings) GetIpv6LocalAddresses() []*string {
	return s.Ipv6LocalAddresses
}

func (s *CreateLoadBalancerRequestZoneMappings) GetPrivateIPv4Address() *string {
	return s.PrivateIPv4Address
}

func (s *CreateLoadBalancerRequestZoneMappings) GetVSwitchId() *string {
	return s.VSwitchId
}

func (s *CreateLoadBalancerRequestZoneMappings) GetZoneId() *string {
	return s.ZoneId
}

func (s *CreateLoadBalancerRequestZoneMappings) SetAllocationId(v string) *CreateLoadBalancerRequestZoneMappings {
	s.AllocationId = &v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) SetIpv4LocalAddresses(v []*string) *CreateLoadBalancerRequestZoneMappings {
	s.Ipv4LocalAddresses = v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) SetIpv6Address(v string) *CreateLoadBalancerRequestZoneMappings {
	s.Ipv6Address = &v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) SetIpv6LocalAddresses(v []*string) *CreateLoadBalancerRequestZoneMappings {
	s.Ipv6LocalAddresses = v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) SetPrivateIPv4Address(v string) *CreateLoadBalancerRequestZoneMappings {
	s.PrivateIPv4Address = &v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) SetVSwitchId(v string) *CreateLoadBalancerRequestZoneMappings {
	s.VSwitchId = &v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) SetZoneId(v string) *CreateLoadBalancerRequestZoneMappings {
	s.ZoneId = &v
	return s
}

func (s *CreateLoadBalancerRequestZoneMappings) Validate() error {
	return dara.Validate(s)
}
