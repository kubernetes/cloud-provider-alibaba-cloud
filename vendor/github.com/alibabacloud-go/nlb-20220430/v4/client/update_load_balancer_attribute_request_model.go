// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerAttributeRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *UpdateLoadBalancerAttributeRequest
	GetClientToken() *string
	SetCps(v int32) *UpdateLoadBalancerAttributeRequest
	GetCps() *int32
	SetCrossZoneEnabled(v bool) *UpdateLoadBalancerAttributeRequest
	GetCrossZoneEnabled() *bool
	SetDryRun(v bool) *UpdateLoadBalancerAttributeRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *UpdateLoadBalancerAttributeRequest
	GetLoadBalancerId() *string
	SetLoadBalancerName(v string) *UpdateLoadBalancerAttributeRequest
	GetLoadBalancerName() *string
	SetRegionId(v string) *UpdateLoadBalancerAttributeRequest
	GetRegionId() *string
}

type UpdateLoadBalancerAttributeRequest struct {
	// The client token that is used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. The client token can contain only ASCII characters.
	//
	// >  If you do not specify this parameter, the system uses the **request ID*	- as the **client token**. The **request ID*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// The maximum number of new connections per second in each zone supported by the NLB instance (virtual IP address). Valid values: **1*	- to **1000000**.
	//
	// example:
	//
	// 1
	Cps *int32 `json:"Cps,omitempty" xml:"Cps,omitempty"`
	// Specifies whether to enable cross-zone load balancing for the NLB instance. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	CrossZoneEnabled *bool `json:"CrossZoneEnabled,omitempty" xml:"CrossZoneEnabled,omitempty"`
	// Specifies whether to perform a dry run, without sending the actual request. Valid values:
	//
	// 	- **true**: performs only a dry run. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the dry run, an error message is returned. If the request passes the dry run, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): performs a dry run and sends the actual request. If the request passes the dry run, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// false
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The NLB instance ID.
	//
	// This parameter is required.
	//
	// example:
	//
	// nlb-wb7r6dlwetvt5j****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The NLB instance name.
	//
	// The name must be 2 to 128 characters in length, and can contain letters, digits, periods (.), underscores (_), and hyphens (-). The name must start with a letter.
	//
	// example:
	//
	// NLB1
	LoadBalancerName *string `json:"LoadBalancerName,omitempty" xml:"LoadBalancerName,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-beijing
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s UpdateLoadBalancerAttributeRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAttributeRequest) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAttributeRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateLoadBalancerAttributeRequest) GetCps() *int32 {
	return s.Cps
}

func (s *UpdateLoadBalancerAttributeRequest) GetCrossZoneEnabled() *bool {
	return s.CrossZoneEnabled
}

func (s *UpdateLoadBalancerAttributeRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateLoadBalancerAttributeRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *UpdateLoadBalancerAttributeRequest) GetLoadBalancerName() *string {
	return s.LoadBalancerName
}

func (s *UpdateLoadBalancerAttributeRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateLoadBalancerAttributeRequest) SetClientToken(v string) *UpdateLoadBalancerAttributeRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) SetCps(v int32) *UpdateLoadBalancerAttributeRequest {
	s.Cps = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) SetCrossZoneEnabled(v bool) *UpdateLoadBalancerAttributeRequest {
	s.CrossZoneEnabled = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) SetDryRun(v bool) *UpdateLoadBalancerAttributeRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) SetLoadBalancerId(v string) *UpdateLoadBalancerAttributeRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) SetLoadBalancerName(v string) *UpdateLoadBalancerAttributeRequest {
	s.LoadBalancerName = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) SetRegionId(v string) *UpdateLoadBalancerAttributeRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateLoadBalancerAttributeRequest) Validate() error {
	return dara.Validate(s)
}
