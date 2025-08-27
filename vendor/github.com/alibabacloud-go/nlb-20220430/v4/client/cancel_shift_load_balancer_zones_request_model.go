// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCancelShiftLoadBalancerZonesRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *CancelShiftLoadBalancerZonesRequest
	GetClientToken() *string
	SetDryRun(v bool) *CancelShiftLoadBalancerZonesRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *CancelShiftLoadBalancerZonesRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *CancelShiftLoadBalancerZonesRequest
	GetRegionId() *string
	SetZoneMappings(v []*CancelShiftLoadBalancerZonesRequestZoneMappings) *CancelShiftLoadBalancerZonesRequest
	GetZoneMappings() []*CancelShiftLoadBalancerZonesRequestZoneMappings
}

type CancelShiftLoadBalancerZonesRequest struct {
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
	// 	- **false*	- (default): validates the request and performs the operation. If the request passes the validation, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// true
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The ID of the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The mapping between the zone and the vSwitch.
	//
	// >  You can specify only one zone ID in each call.
	//
	// This parameter is required.
	ZoneMappings []*CancelShiftLoadBalancerZonesRequestZoneMappings `json:"ZoneMappings,omitempty" xml:"ZoneMappings,omitempty" type:"Repeated"`
}

func (s CancelShiftLoadBalancerZonesRequest) String() string {
	return dara.Prettify(s)
}

func (s CancelShiftLoadBalancerZonesRequest) GoString() string {
	return s.String()
}

func (s *CancelShiftLoadBalancerZonesRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *CancelShiftLoadBalancerZonesRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *CancelShiftLoadBalancerZonesRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *CancelShiftLoadBalancerZonesRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *CancelShiftLoadBalancerZonesRequest) GetZoneMappings() []*CancelShiftLoadBalancerZonesRequestZoneMappings {
	return s.ZoneMappings
}

func (s *CancelShiftLoadBalancerZonesRequest) SetClientToken(v string) *CancelShiftLoadBalancerZonesRequest {
	s.ClientToken = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequest) SetDryRun(v bool) *CancelShiftLoadBalancerZonesRequest {
	s.DryRun = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequest) SetLoadBalancerId(v string) *CancelShiftLoadBalancerZonesRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequest) SetRegionId(v string) *CancelShiftLoadBalancerZonesRequest {
	s.RegionId = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequest) SetZoneMappings(v []*CancelShiftLoadBalancerZonesRequestZoneMappings) *CancelShiftLoadBalancerZonesRequest {
	s.ZoneMappings = v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequest) Validate() error {
	return dara.Validate(s)
}

type CancelShiftLoadBalancerZonesRequestZoneMappings struct {
	// The ID of the vSwitch in the zone. By default, each zone uses one vSwitch and one subnet.
	//
	// This parameter is required.
	//
	// example:
	//
	// vsw-sersdf****
	VSwitchId *string `json:"VSwitchId,omitempty" xml:"VSwitchId,omitempty"`
	// The zone ID of the NLB instance.
	//
	// >  You can specify only one zone ID in each call.
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

func (s CancelShiftLoadBalancerZonesRequestZoneMappings) String() string {
	return dara.Prettify(s)
}

func (s CancelShiftLoadBalancerZonesRequestZoneMappings) GoString() string {
	return s.String()
}

func (s *CancelShiftLoadBalancerZonesRequestZoneMappings) GetVSwitchId() *string {
	return s.VSwitchId
}

func (s *CancelShiftLoadBalancerZonesRequestZoneMappings) GetZoneId() *string {
	return s.ZoneId
}

func (s *CancelShiftLoadBalancerZonesRequestZoneMappings) SetVSwitchId(v string) *CancelShiftLoadBalancerZonesRequestZoneMappings {
	s.VSwitchId = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequestZoneMappings) SetZoneId(v string) *CancelShiftLoadBalancerZonesRequestZoneMappings {
	s.ZoneId = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesRequestZoneMappings) Validate() error {
	return dara.Validate(s)
}
