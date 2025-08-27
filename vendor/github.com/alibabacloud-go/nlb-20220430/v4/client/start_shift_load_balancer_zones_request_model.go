// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStartShiftLoadBalancerZonesRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *StartShiftLoadBalancerZonesRequest
	GetClientToken() *string
	SetDryRun(v bool) *StartShiftLoadBalancerZonesRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *StartShiftLoadBalancerZonesRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *StartShiftLoadBalancerZonesRequest
	GetRegionId() *string
	SetZoneMappings(v []*StartShiftLoadBalancerZonesRequestZoneMappings) *StartShiftLoadBalancerZonesRequest
	GetZoneMappings() []*StartShiftLoadBalancerZonesRequestZoneMappings
}

type StartShiftLoadBalancerZonesRequest struct {
	// The client token that is used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. The client token can contain only ASCII characters.
	//
	// >  If you do not specify this parameter, the system automatically uses the **request ID*	- as the **client token**. The **request ID*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// Specifies whether to perform a dry run, without sending the actual request. Valid values:
	//
	// 	- **true**: performs only a dry run. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the dry run, an error message is returned. If the request passes the dry run, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): performs a dry run and sends the actual request. If the request passes the dry run, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// true
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The NLB instance ID.
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
	// The mappings between zones and vSwitches.
	//
	// >  You can remove only one IP address (or zone) in each call.
	//
	// This parameter is required.
	ZoneMappings []*StartShiftLoadBalancerZonesRequestZoneMappings `json:"ZoneMappings,omitempty" xml:"ZoneMappings,omitempty" type:"Repeated"`
}

func (s StartShiftLoadBalancerZonesRequest) String() string {
	return dara.Prettify(s)
}

func (s StartShiftLoadBalancerZonesRequest) GoString() string {
	return s.String()
}

func (s *StartShiftLoadBalancerZonesRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *StartShiftLoadBalancerZonesRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *StartShiftLoadBalancerZonesRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *StartShiftLoadBalancerZonesRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *StartShiftLoadBalancerZonesRequest) GetZoneMappings() []*StartShiftLoadBalancerZonesRequestZoneMappings {
	return s.ZoneMappings
}

func (s *StartShiftLoadBalancerZonesRequest) SetClientToken(v string) *StartShiftLoadBalancerZonesRequest {
	s.ClientToken = &v
	return s
}

func (s *StartShiftLoadBalancerZonesRequest) SetDryRun(v bool) *StartShiftLoadBalancerZonesRequest {
	s.DryRun = &v
	return s
}

func (s *StartShiftLoadBalancerZonesRequest) SetLoadBalancerId(v string) *StartShiftLoadBalancerZonesRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *StartShiftLoadBalancerZonesRequest) SetRegionId(v string) *StartShiftLoadBalancerZonesRequest {
	s.RegionId = &v
	return s
}

func (s *StartShiftLoadBalancerZonesRequest) SetZoneMappings(v []*StartShiftLoadBalancerZonesRequestZoneMappings) *StartShiftLoadBalancerZonesRequest {
	s.ZoneMappings = v
	return s
}

func (s *StartShiftLoadBalancerZonesRequest) Validate() error {
	return dara.Validate(s)
}

type StartShiftLoadBalancerZonesRequestZoneMappings struct {
	// The ID of the vSwitch in the zone. By default, each zone uses one vSwitch and one subnet.
	//
	// This parameter is required.
	//
	// example:
	//
	// vsw-bp1rmcrwg3erh1fh8****
	VSwitchId *string `json:"VSwitchId,omitempty" xml:"VSwitchId,omitempty"`
	// The zone ID of the NLB instance.
	//
	// >  You can remove only one IP address (or zone) in each call.
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

func (s StartShiftLoadBalancerZonesRequestZoneMappings) String() string {
	return dara.Prettify(s)
}

func (s StartShiftLoadBalancerZonesRequestZoneMappings) GoString() string {
	return s.String()
}

func (s *StartShiftLoadBalancerZonesRequestZoneMappings) GetVSwitchId() *string {
	return s.VSwitchId
}

func (s *StartShiftLoadBalancerZonesRequestZoneMappings) GetZoneId() *string {
	return s.ZoneId
}

func (s *StartShiftLoadBalancerZonesRequestZoneMappings) SetVSwitchId(v string) *StartShiftLoadBalancerZonesRequestZoneMappings {
	s.VSwitchId = &v
	return s
}

func (s *StartShiftLoadBalancerZonesRequestZoneMappings) SetZoneId(v string) *StartShiftLoadBalancerZonesRequestZoneMappings {
	s.ZoneId = &v
	return s
}

func (s *StartShiftLoadBalancerZonesRequestZoneMappings) Validate() error {
	return dara.Validate(s)
}
