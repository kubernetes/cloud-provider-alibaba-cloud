// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListLoadBalancersResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetLoadBalancers(v []*ListLoadBalancersResponseBodyLoadBalancers) *ListLoadBalancersResponseBody
	GetLoadBalancers() []*ListLoadBalancersResponseBodyLoadBalancers
	SetMaxResults(v int32) *ListLoadBalancersResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *ListLoadBalancersResponseBody
	GetNextToken() *string
	SetRequestId(v string) *ListLoadBalancersResponseBody
	GetRequestId() *string
	SetTotalCount(v int32) *ListLoadBalancersResponseBody
	GetTotalCount() *int32
}

type ListLoadBalancersResponseBody struct {
	// The NLB instances.
	LoadBalancers []*ListLoadBalancersResponseBodyLoadBalancers `json:"LoadBalancers,omitempty" xml:"LoadBalancers,omitempty" type:"Repeated"`
	// The number of entries returned per page.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The token that determines the start point of the next query. Valid values:
	//
	// 	- If this is your first query and no subsequent queries are to be sent, ignore this parameter.
	//
	// 	- If a subsequent query is to be sent, set the parameter to the value of NextToken that is returned from the last call.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The total number of entries returned.
	//
	// example:
	//
	// 10
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListLoadBalancersResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBody) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBody) GetLoadBalancers() []*ListLoadBalancersResponseBodyLoadBalancers {
	return s.LoadBalancers
}

func (s *ListLoadBalancersResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListLoadBalancersResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *ListLoadBalancersResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListLoadBalancersResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *ListLoadBalancersResponseBody) SetLoadBalancers(v []*ListLoadBalancersResponseBodyLoadBalancers) *ListLoadBalancersResponseBody {
	s.LoadBalancers = v
	return s
}

func (s *ListLoadBalancersResponseBody) SetMaxResults(v int32) *ListLoadBalancersResponseBody {
	s.MaxResults = &v
	return s
}

func (s *ListLoadBalancersResponseBody) SetNextToken(v string) *ListLoadBalancersResponseBody {
	s.NextToken = &v
	return s
}

func (s *ListLoadBalancersResponseBody) SetRequestId(v string) *ListLoadBalancersResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListLoadBalancersResponseBody) SetTotalCount(v int32) *ListLoadBalancersResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListLoadBalancersResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancers struct {
	// The IP version. Valid values:
	//
	// 	- **ipv4**: IPv4
	//
	// 	- **DualStack**: dual stack
	//
	// example:
	//
	// ipv4
	AddressIpVersion *string `json:"AddressIpVersion,omitempty" xml:"AddressIpVersion,omitempty"`
	// The type of IPv4 address used by the NLB instance. Valid values:
	//
	// 	- **Internet**: The NLB instance uses a public IP address. The domain name of the NLB instance is resolved to the public IP address. Therefore, the NLB instance can be accessed over the Internet.
	//
	// 	- **Intranet**: The NLB instance uses a private IP address. The domain name of the NLB instance is resolved to the private IP address. Therefore, the NLB instance can be accessed over the VPC where the NLB instance is deployed.
	//
	// example:
	//
	// Internet
	AddressType *string `json:"AddressType,omitempty" xml:"AddressType,omitempty"`
	// The ID of the EIP bandwidth plan that is associated with the NLB instance if the NLB instance uses a public IP address.
	//
	// example:
	//
	// cbwp-bp1vevu8h3ieh****
	BandwidthPackageId *string `json:"BandwidthPackageId,omitempty" xml:"BandwidthPackageId,omitempty"`
	// The time when the resource was created. The time is displayed in UTC in the `yyyy-MM-ddTHH:mm:ssZ` format.
	//
	// example:
	//
	// 2022-07-18T17:22Z
	CreateTime *string `json:"CreateTime,omitempty" xml:"CreateTime,omitempty"`
	// Indicates whether cross-zone load balancing is enabled for the NLB instance. Valid values:
	//
	// 	- **true**: enabled
	//
	// 	- **false**: disabled
	//
	// example:
	//
	// true
	CrossZoneEnabled *bool `json:"CrossZoneEnabled,omitempty" xml:"CrossZoneEnabled,omitempty"`
	// The domain name of the NLB instance.
	//
	// example:
	//
	// nlb-wb7r6dlwetvt5j****.cn-hangzhou.nlb.aliyuncs.com
	DNSName *string `json:"DNSName,omitempty" xml:"DNSName,omitempty"`
	// The configuration of the deletion protection feature.
	DeletionProtectionConfig *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig `json:"DeletionProtectionConfig,omitempty" xml:"DeletionProtectionConfig,omitempty" type:"Struct"`
	// The type of IPv6 address used by the NLB instance. Valid values:
	//
	// 	- **Internet**: The NLB instance uses a public IP address. The domain name of the NLB instance is resolved to the public IP address. Therefore, the NLB instance can be accessed over the Internet.
	//
	// 	- **Intranet**: The NLB instance uses a private IP address. The domain name of the NLB instance is resolved to the private IP address. Therefore, the NLB instance can be accessed over the VPC where the NLB instance is deployed.
	//
	// example:
	//
	// Internet
	Ipv6AddressType *string `json:"Ipv6AddressType,omitempty" xml:"Ipv6AddressType,omitempty"`
	// The billing settings of the NLB instance.
	LoadBalancerBillingConfig *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig `json:"LoadBalancerBillingConfig,omitempty" xml:"LoadBalancerBillingConfig,omitempty" type:"Struct"`
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
	// The ID of the NLB instance.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The name of the NLB instance.
	//
	// example:
	//
	// NLB1
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" xml:"LoadBalancerName,omitempty"`
	// The status of the NLB instance. Valid values:
	//
	// 	- **Inactive**: The NLB instance is disabled. Listeners of NLB instances in the Inactive state do not forward traffic.
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
	// The type of the SLB instance. Only **Network*	- is returned, which indicates NLB.
	//
	// example:
	//
	// Network
	LoadBalancerType *string `json:"LoadBalancerType,omitempty" xml:"LoadBalancerType,omitempty"`
	// The configuration of the configuration read-only mode.
	ModificationProtectionConfig *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig `json:"ModificationProtectionConfig,omitempty" xml:"ModificationProtectionConfig,omitempty" type:"Struct"`
	// The configuration of the operation lock. This parameter takes effect if the value of `LoadBalancerBussinessStatus` is **Abnormal**.
	OperationLocks []*ListLoadBalancersResponseBodyLoadBalancersOperationLocks `json:"OperationLocks,omitempty" xml:"OperationLocks,omitempty" type:"Repeated"`
	// The ID of the region where the NLB instance is deployed.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The ID of the resource group.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The security group to which the NLB instance is added.
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" xml:"SecurityGroupIds,omitempty" type:"Repeated"`
	// A list of tags.
	Tags []*ListLoadBalancersResponseBodyLoadBalancersTags `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
	// The ID of the VPC where the NLB instance is deployed.
	//
	// example:
	//
	// vpc-bp1b49rqrybk45nio****
	VpcId *string `json:"VpcId,omitempty" xml:"VpcId,omitempty"`
	// The mappings between zones and vSwitches.
	ZoneMappings []*ListLoadBalancersResponseBodyLoadBalancersZoneMappings `json:"ZoneMappings,omitempty" xml:"ZoneMappings,omitempty" type:"Repeated"`
}

func (s ListLoadBalancersResponseBodyLoadBalancers) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancers) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetAddressIpVersion() *string {
	return s.AddressIpVersion
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetAddressType() *string {
	return s.AddressType
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetBandwidthPackageId() *string {
	return s.BandwidthPackageId
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetCreateTime() *string {
	return s.CreateTime
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetCrossZoneEnabled() *bool {
	return s.CrossZoneEnabled
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetDNSName() *string {
	return s.DNSName
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetDeletionProtectionConfig() *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig {
	return s.DeletionProtectionConfig
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetIpv6AddressType() *string {
	return s.Ipv6AddressType
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetLoadBalancerBillingConfig() *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig {
	return s.LoadBalancerBillingConfig
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetLoadBalancerBusinessStatus() *string {
	return s.LoadBalancerBusinessStatus
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetLoadBalancerName() *string {
	return s.LoadBalancerName
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetLoadBalancerStatus() *string {
	return s.LoadBalancerStatus
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetLoadBalancerType() *string {
	return s.LoadBalancerType
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetModificationProtectionConfig() *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig {
	return s.ModificationProtectionConfig
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetOperationLocks() []*ListLoadBalancersResponseBodyLoadBalancersOperationLocks {
	return s.OperationLocks
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetRegionId() *string {
	return s.RegionId
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetSecurityGroupIds() []*string {
	return s.SecurityGroupIds
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetTags() []*ListLoadBalancersResponseBodyLoadBalancersTags {
	return s.Tags
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetVpcId() *string {
	return s.VpcId
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) GetZoneMappings() []*ListLoadBalancersResponseBodyLoadBalancersZoneMappings {
	return s.ZoneMappings
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetAddressIpVersion(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.AddressIpVersion = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetAddressType(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.AddressType = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetBandwidthPackageId(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.BandwidthPackageId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetCreateTime(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.CreateTime = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetCrossZoneEnabled(v bool) *ListLoadBalancersResponseBodyLoadBalancers {
	s.CrossZoneEnabled = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetDNSName(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.DNSName = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetDeletionProtectionConfig(v *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) *ListLoadBalancersResponseBodyLoadBalancers {
	s.DeletionProtectionConfig = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetIpv6AddressType(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.Ipv6AddressType = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetLoadBalancerBillingConfig(v *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig) *ListLoadBalancersResponseBodyLoadBalancers {
	s.LoadBalancerBillingConfig = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetLoadBalancerBusinessStatus(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.LoadBalancerBusinessStatus = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetLoadBalancerId(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.LoadBalancerId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetLoadBalancerName(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.LoadBalancerName = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetLoadBalancerStatus(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.LoadBalancerStatus = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetLoadBalancerType(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.LoadBalancerType = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetModificationProtectionConfig(v *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) *ListLoadBalancersResponseBodyLoadBalancers {
	s.ModificationProtectionConfig = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetOperationLocks(v []*ListLoadBalancersResponseBodyLoadBalancersOperationLocks) *ListLoadBalancersResponseBodyLoadBalancers {
	s.OperationLocks = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetRegionId(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.RegionId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetResourceGroupId(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.ResourceGroupId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetSecurityGroupIds(v []*string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.SecurityGroupIds = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetTags(v []*ListLoadBalancersResponseBodyLoadBalancersTags) *ListLoadBalancersResponseBodyLoadBalancers {
	s.Tags = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetVpcId(v string) *ListLoadBalancersResponseBodyLoadBalancers {
	s.VpcId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) SetZoneMappings(v []*ListLoadBalancersResponseBodyLoadBalancersZoneMappings) *ListLoadBalancersResponseBodyLoadBalancers {
	s.ZoneMappings = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancers) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig struct {
	// Indicates whether deletion protection is enabled. Valid values:
	//
	// 	- **true**: enabled
	//
	// 	- **false**: disabled
	//
	// example:
	//
	// true
	Enabled *bool `json:"Enabled,omitempty" xml:"Enabled,omitempty"`
	// The time when deletion protection was enabled. The time is displayed in UTC in `yyyy-MM-ddTHH:mm:ssZ` format.
	//
	// example:
	//
	// 2022-12-01T17:22Z
	EnabledTime *string `json:"EnabledTime,omitempty" xml:"EnabledTime,omitempty"`
	// The reason why the deletion protection feature is enabled or disabled. The reason must be 2 to 128 characters in length and can contain letters, digits, periods (.), underscores (_), and hyphens (-). The reason must start with a letter.
	//
	// example:
	//
	// The instance is running
	Reason *string `json:"Reason,omitempty" xml:"Reason,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) GetEnabled() *bool {
	return s.Enabled
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) GetEnabledTime() *string {
	return s.EnabledTime
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) GetReason() *string {
	return s.Reason
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) SetEnabled(v bool) *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig {
	s.Enabled = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) SetEnabledTime(v string) *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig {
	s.EnabledTime = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) SetReason(v string) *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig {
	s.Reason = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersDeletionProtectionConfig) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig struct {
	// The billing method of the NLB instance. Only **PostPay*	- is supported, which indicates the pay-as-you-go billing method.
	//
	// example:
	//
	// PostPay
	PayType *string `json:"PayType,omitempty" xml:"PayType,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig) GetPayType() *string {
	return s.PayType
}

func (s *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig) SetPayType(v string) *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig {
	s.PayType = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersLoadBalancerBillingConfig) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig struct {
	// The time when the configuration read-only mode was enabled. The time is displayed in UTC in `yyyy-MM-ddTHH:mm:ssZ` format.
	//
	// example:
	//
	// 2022-12-01T17:22Z
	EnabledTime *string `json:"EnabledTime,omitempty" xml:"EnabledTime,omitempty"`
	// The reason why the configuration read-only mode is enabled. The reason must be 2 to 128 characters in length and can contain letters, digits, periods (.), underscores (_), and hyphens (-). The reason must start with a letter.
	//
	// This parameter takes effect only if **Status*	- is set to **ConsoleProtection**.
	//
	// example:
	//
	// Service guarantee period
	Reason *string `json:"Reason,omitempty" xml:"Reason,omitempty"`
	// Indicates whether the configuration read-only mode is enabled. Valid values:
	//
	// 	- **NonProtection**: disabled. In this case, **Reason*	- is not returned. If **Reason*	- is set, the value is cleared.
	//
	// 	- **ConsoleProtection**: enabled. In this case, **Reason*	- is returned.
	//
	// >  If you set this parameter to **ConsoleProtection**, you cannot use the NLB console to modify instance configurations. However, you can call API operations to modify instance configurations.
	//
	// example:
	//
	// ConsoleProtection
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) GetEnabledTime() *string {
	return s.EnabledTime
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) GetReason() *string {
	return s.Reason
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) GetStatus() *string {
	return s.Status
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) SetEnabledTime(v string) *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig {
	s.EnabledTime = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) SetReason(v string) *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig {
	s.Reason = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) SetStatus(v string) *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig {
	s.Status = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersModificationProtectionConfig) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersOperationLocks struct {
	// The reason why the NLB instance is locked.
	//
	// example:
	//
	// Service exception
	LockReason *string `json:"LockReason,omitempty" xml:"LockReason,omitempty"`
	// The type of lock. Valid values:
	//
	// 	- **SecurityLocked**: The NLB instance is locked due to security reasons.
	//
	// 	- **RelatedResourceLocked**: The NLB instance is locked due to association issues.
	//
	// 	- **FinancialLocked**: The NLB instance is locked due to overdue payments.
	//
	// 	- **ResidualLocked**: The NLB instance is locked because the payments of the associated resources are overdue and the resources are released.
	//
	// example:
	//
	// SecurityLocked
	LockType *string `json:"LockType,omitempty" xml:"LockType,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersOperationLocks) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersOperationLocks) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersOperationLocks) GetLockReason() *string {
	return s.LockReason
}

func (s *ListLoadBalancersResponseBodyLoadBalancersOperationLocks) GetLockType() *string {
	return s.LockType
}

func (s *ListLoadBalancersResponseBodyLoadBalancersOperationLocks) SetLockReason(v string) *ListLoadBalancersResponseBodyLoadBalancersOperationLocks {
	s.LockReason = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersOperationLocks) SetLockType(v string) *ListLoadBalancersResponseBodyLoadBalancersOperationLocks {
	s.LockType = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersOperationLocks) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersTags struct {
	// The tag key.
	//
	// example:
	//
	// KeyTest
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The tag value.
	//
	// example:
	//
	// ValueTest
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersTags) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersTags) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersTags) GetKey() *string {
	return s.Key
}

func (s *ListLoadBalancersResponseBodyLoadBalancersTags) GetValue() *string {
	return s.Value
}

func (s *ListLoadBalancersResponseBodyLoadBalancersTags) SetKey(v string) *ListLoadBalancersResponseBodyLoadBalancersTags {
	s.Key = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersTags) SetValue(v string) *ListLoadBalancersResponseBodyLoadBalancersTags {
	s.Value = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersTags) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersZoneMappings struct {
	// The IP addresses that are used by the NLB instance.
	LoadBalancerAddresses []*ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses `json:"LoadBalancerAddresses,omitempty" xml:"LoadBalancerAddresses,omitempty" type:"Repeated"`
	// The zone status. Valid values:
	//
	// - **Active**: The zone is available.
	//
	// - **Stopped**: The zone is disabled. You can set the zone to this status only by using Cloud Architect Design Tools (CADT).
	//
	// - **Shifted**: The DNS record is removed.
	//
	// - **Starting**: The zone is being enabled. You can set the zone to this status only by using CADT.
	//
	// - **Stopping*	- You can set the zone to this status only by using CADT.
	//
	// example:
	//
	// Active
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
	// The ID of the vSwitch in the zone. By default, each zone contains one vSwitch and one subnet.
	//
	// example:
	//
	// vsw-bp1rmcrwg3erh1fh8****
	VSwitchId *string `json:"VSwitchId,omitempty" xml:"VSwitchId,omitempty"`
	// The name of the zone. You can call the [DescribeZones](https://help.aliyun.com/document_detail/443890.html) operation to query the zones.
	//
	// example:
	//
	// cn-hangzhou-a
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersZoneMappings) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersZoneMappings) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) GetLoadBalancerAddresses() []*ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	return s.LoadBalancerAddresses
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) GetStatus() *string {
	return s.Status
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) GetVSwitchId() *string {
	return s.VSwitchId
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) GetZoneId() *string {
	return s.ZoneId
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) SetLoadBalancerAddresses(v []*ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) *ListLoadBalancersResponseBodyLoadBalancersZoneMappings {
	s.LoadBalancerAddresses = v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) SetStatus(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappings {
	s.Status = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) SetVSwitchId(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappings {
	s.VSwitchId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) SetZoneId(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappings {
	s.ZoneId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappings) Validate() error {
	return dara.Validate(s)
}

type ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses struct {
	// The ID of the elastic IP address (EIP).
	//
	// example:
	//
	// eip-bp1aedxso6u80u0qf****
	AllocationId *string `json:"AllocationId,omitempty" xml:"AllocationId,omitempty"`
	// The ID of the elastic network interface (ENI) attached to the NLB instance.
	//
	// example:
	//
	// eni-bp12f1xhs5yal61a****
	EniId *string `json:"EniId,omitempty" xml:"EniId,omitempty"`
	// The IPv6 address used by the NLB instance.
	//
	// example:
	//
	// 2001:db8:1:1:1:1:1:1
	Ipv6Address *string `json:"Ipv6Address,omitempty" xml:"Ipv6Address,omitempty"`
	// The private IPv4 address of the NLB instance.
	//
	// example:
	//
	// 192.168.3.32
	PrivateIPv4Address *string `json:"PrivateIPv4Address,omitempty" xml:"PrivateIPv4Address,omitempty"`
	// The health status of the private IPv4 address of the NLB instance. Valid values:
	//
	// - **Healthy**
	//
	// - **Unhealthy**
	//
	// > This parameter is returned only when the Status of the zone is Active.
	//
	// example:
	//
	// Healthy
	PrivateIPv4HcStatus *string `json:"PrivateIPv4HcStatus,omitempty" xml:"PrivateIPv4HcStatus,omitempty"`
	// The health status of the IPv6 address of the NLB instance. Valid values:
	//
	// - **Healthy**
	//
	// - **Unhealthy**
	//
	// > This parameter is returned only when the Status of the zone is Active.
	//
	// example:
	//
	// Healthy
	PrivateIPv6HcStatus *string `json:"PrivateIPv6HcStatus,omitempty" xml:"PrivateIPv6HcStatus,omitempty"`
	// The public IPv4 address of the NLB instance.
	//
	// example:
	//
	// 120.XX.XX.69
	PublicIPv4Address *string `json:"PublicIPv4Address,omitempty" xml:"PublicIPv4Address,omitempty"`
}

func (s ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetAllocationId() *string {
	return s.AllocationId
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetEniId() *string {
	return s.EniId
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetIpv6Address() *string {
	return s.Ipv6Address
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetPrivateIPv4Address() *string {
	return s.PrivateIPv4Address
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetPrivateIPv4HcStatus() *string {
	return s.PrivateIPv4HcStatus
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetPrivateIPv6HcStatus() *string {
	return s.PrivateIPv6HcStatus
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) GetPublicIPv4Address() *string {
	return s.PublicIPv4Address
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetAllocationId(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.AllocationId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetEniId(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.EniId = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetIpv6Address(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.Ipv6Address = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetPrivateIPv4Address(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.PrivateIPv4Address = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetPrivateIPv4HcStatus(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.PrivateIPv4HcStatus = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetPrivateIPv6HcStatus(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.PrivateIPv6HcStatus = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) SetPublicIPv4Address(v string) *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses {
	s.PublicIPv4Address = &v
	return s
}

func (s *ListLoadBalancersResponseBodyLoadBalancersZoneMappingsLoadBalancerAddresses) Validate() error {
	return dara.Validate(s)
}
