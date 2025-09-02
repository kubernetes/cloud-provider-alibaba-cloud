// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerZonesRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *UpdateLoadBalancerZonesRequest
	GetClientToken() *string
	SetDryRun(v bool) *UpdateLoadBalancerZonesRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *UpdateLoadBalancerZonesRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *UpdateLoadBalancerZonesRequest
	GetRegionId() *string
	SetZoneMappings(v []*UpdateLoadBalancerZonesRequestZoneMappings) *UpdateLoadBalancerZonesRequest
	GetZoneMappings() []*UpdateLoadBalancerZonesRequestZoneMappings
}

type UpdateLoadBalancerZonesRequest struct {
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. Only ASCII characters are allowed.
	//
	// >  If you do not set this parameter, the value of **RequestId*	- is used.***	- The value of **RequestId*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// Specifies whether to perform a dry run. Valid values:
	//
	// 	- **true**: validates the request without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the validation, the corresponding error message is returned. If the request passes the validation, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): validates the request and performs the request. If the request passes the validation, an HTTP 2xx status code is returned and the operation is performed.
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
	// The ID of region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The mappings between the zones and the vSwitches. You can specify up to 10 zones.
	//
	// This parameter is required.
	ZoneMappings []*UpdateLoadBalancerZonesRequestZoneMappings `json:"ZoneMappings,omitempty" xml:"ZoneMappings,omitempty" type:"Repeated"`
}

func (s UpdateLoadBalancerZonesRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerZonesRequest) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerZonesRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateLoadBalancerZonesRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateLoadBalancerZonesRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *UpdateLoadBalancerZonesRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateLoadBalancerZonesRequest) GetZoneMappings() []*UpdateLoadBalancerZonesRequestZoneMappings {
	return s.ZoneMappings
}

func (s *UpdateLoadBalancerZonesRequest) SetClientToken(v string) *UpdateLoadBalancerZonesRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequest) SetDryRun(v bool) *UpdateLoadBalancerZonesRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequest) SetLoadBalancerId(v string) *UpdateLoadBalancerZonesRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequest) SetRegionId(v string) *UpdateLoadBalancerZonesRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequest) SetZoneMappings(v []*UpdateLoadBalancerZonesRequestZoneMappings) *UpdateLoadBalancerZonesRequest {
	s.ZoneMappings = v
	return s
}

func (s *UpdateLoadBalancerZonesRequest) Validate() error {
	return dara.Validate(s)
}

type UpdateLoadBalancerZonesRequestZoneMappings struct {
	// The ID of the elastic IP address (EIP) or Anycast EIP.
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
	// >  For regions that support Anycast EIPs, see [Limits](https://help.aliyun.com/document_detail/470000.html).This parameter is required if **AddressType*	- is set to **Internet**.
	//
	// example:
	//
	// Common
	EipType *string `json:"EipType,omitempty" xml:"EipType,omitempty"`
	// The private IP address.
	//
	// example:
	//
	// 192.168.36.16
	PrivateIPv4Address *string `json:"PrivateIPv4Address,omitempty" xml:"PrivateIPv4Address,omitempty"`
	// The ID of the vSwitch in the zone. By default, each zone uses one vSwitch and one subnet.
	//
	// This parameter is required.
	//
	// example:
	//
	// vsw-bp1rmcrwg3erh1fh8****
	VSwitchId *string `json:"VSwitchId,omitempty" xml:"VSwitchId,omitempty"`
	// The zone ID. You can call the [DescribeZones](https://help.aliyun.com/document_detail/443890.html) operation to query the most recent zone list.
	//
	// This parameter is required.
	//
	// example:
	//
	// cn-hangzhou-a
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s UpdateLoadBalancerZonesRequestZoneMappings) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerZonesRequestZoneMappings) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) GetAllocationId() *string {
	return s.AllocationId
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) GetEipType() *string {
	return s.EipType
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) GetPrivateIPv4Address() *string {
	return s.PrivateIPv4Address
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) GetVSwitchId() *string {
	return s.VSwitchId
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) GetZoneId() *string {
	return s.ZoneId
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) SetAllocationId(v string) *UpdateLoadBalancerZonesRequestZoneMappings {
	s.AllocationId = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) SetEipType(v string) *UpdateLoadBalancerZonesRequestZoneMappings {
	s.EipType = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) SetPrivateIPv4Address(v string) *UpdateLoadBalancerZonesRequestZoneMappings {
	s.PrivateIPv4Address = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) SetVSwitchId(v string) *UpdateLoadBalancerZonesRequestZoneMappings {
	s.VSwitchId = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) SetZoneId(v string) *UpdateLoadBalancerZonesRequestZoneMappings {
	s.ZoneId = &v
	return s
}

func (s *UpdateLoadBalancerZonesRequestZoneMappings) Validate() error {
	return dara.Validate(s)
}
