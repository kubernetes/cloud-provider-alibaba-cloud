// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerAddressTypeConfigRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAddressType(v string) *UpdateLoadBalancerAddressTypeConfigRequest
	GetAddressType() *string
	SetClientToken(v string) *UpdateLoadBalancerAddressTypeConfigRequest
	GetClientToken() *string
	SetDryRun(v bool) *UpdateLoadBalancerAddressTypeConfigRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *UpdateLoadBalancerAddressTypeConfigRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *UpdateLoadBalancerAddressTypeConfigRequest
	GetRegionId() *string
	SetZoneMappings(v []*UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) *UpdateLoadBalancerAddressTypeConfigRequest
	GetZoneMappings() []*UpdateLoadBalancerAddressTypeConfigRequestZoneMappings
}

type UpdateLoadBalancerAddressTypeConfigRequest struct {
	// The new network type. Valid values:
	//
	// 	- **Internet**: The nodes of an Internet-facing NLB instance have public IP addresses. The DNS name of an Internet-facing NLB instance is publicly resolvable to the public IP addresses of the nodes. Therefore, Internet-facing NLB instances can route requests from clients over the Internet.
	//
	// 	- **Intranet**: The nodes of an internal-facing NLB instance have only private IP addresses. The DNS name of an internal-facing NLB instance is publicly resolvable to the private IP addresses of the nodes. Therefore, internal-facing NLB instances can route requests only from clients with access to the virtual private cloud (VPC) for the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// Internet
	AddressType *string `json:"AddressType,omitempty" xml:"AddressType,omitempty"`
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate this value. Ensure that the value is unique among all requests. Only ASCII characters are allowed.
	//
	// >  If you do not specify this parameter, the value of **RequestId*	- is used.***	- **RequestId*	- of each request is different.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
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
	// The ID of the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The mappings between zones and vSwitches. You can specify up to 10 zones.
	ZoneMappings []*UpdateLoadBalancerAddressTypeConfigRequestZoneMappings `json:"ZoneMappings,omitempty" xml:"ZoneMappings,omitempty" type:"Repeated"`
}

func (s UpdateLoadBalancerAddressTypeConfigRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAddressTypeConfigRequest) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) GetAddressType() *string {
	return s.AddressType
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) GetZoneMappings() []*UpdateLoadBalancerAddressTypeConfigRequestZoneMappings {
	return s.ZoneMappings
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) SetAddressType(v string) *UpdateLoadBalancerAddressTypeConfigRequest {
	s.AddressType = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) SetClientToken(v string) *UpdateLoadBalancerAddressTypeConfigRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) SetDryRun(v bool) *UpdateLoadBalancerAddressTypeConfigRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) SetLoadBalancerId(v string) *UpdateLoadBalancerAddressTypeConfigRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) SetRegionId(v string) *UpdateLoadBalancerAddressTypeConfigRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) SetZoneMappings(v []*UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) *UpdateLoadBalancerAddressTypeConfigRequest {
	s.ZoneMappings = v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequest) Validate() error {
	return dara.Validate(s)
}

type UpdateLoadBalancerAddressTypeConfigRequestZoneMappings struct {
	// The ID of the elastic IP address (EIP).
	//
	// example:
	//
	// eip-bp1aedxso6u80u0qf****
	AllocationId *string `json:"AllocationId,omitempty" xml:"AllocationId,omitempty"`
	// The type of the EIP. Valid values:
	//
	// 	- **Common**: an EIP
	//
	// 	- **Anycast**: an Anycast EIP
	//
	// >  This parameter is required only if **AddressType*	- is set to **Internet**.
	//
	// example:
	//
	// Common
	EipType *string `json:"EipType,omitempty" xml:"EipType,omitempty"`
	// The ID of the vSwitch in the zone. You can specify only one vSwitch (subnet) in each zone of an NLB instance.
	//
	// example:
	//
	// vsw-bp10ttov87felojcn****
	VSwitchId *string `json:"VSwitchId,omitempty" xml:"VSwitchId,omitempty"`
	// The zone ID of the NLB instance.
	//
	// You can call the [DescribeZones](https://help.aliyun.com/document_detail/443890.html) operation to query the most recent zone list.
	//
	// example:
	//
	// cn-hangzhou-a
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) GetAllocationId() *string {
	return s.AllocationId
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) GetEipType() *string {
	return s.EipType
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) GetVSwitchId() *string {
	return s.VSwitchId
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) GetZoneId() *string {
	return s.ZoneId
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) SetAllocationId(v string) *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings {
	s.AllocationId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) SetEipType(v string) *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings {
	s.EipType = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) SetVSwitchId(v string) *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings {
	s.VSwitchId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) SetZoneId(v string) *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings {
	s.ZoneId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigRequestZoneMappings) Validate() error {
	return dara.Validate(s)
}
