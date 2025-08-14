// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteLoadBalancerRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *DeleteLoadBalancerRequest
	GetClientToken() *string
	SetDryRun(v bool) *DeleteLoadBalancerRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *DeleteLoadBalancerRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *DeleteLoadBalancerRequest
	GetRegionId() *string
}

type DeleteLoadBalancerRequest struct {
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
	// false
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The ID of the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// nlb-wb7r6dlwetvt5j****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s DeleteLoadBalancerRequest) String() string {
	return dara.Prettify(s)
}

func (s DeleteLoadBalancerRequest) GoString() string {
	return s.String()
}

func (s *DeleteLoadBalancerRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *DeleteLoadBalancerRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *DeleteLoadBalancerRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *DeleteLoadBalancerRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *DeleteLoadBalancerRequest) SetClientToken(v string) *DeleteLoadBalancerRequest {
	s.ClientToken = &v
	return s
}

func (s *DeleteLoadBalancerRequest) SetDryRun(v bool) *DeleteLoadBalancerRequest {
	s.DryRun = &v
	return s
}

func (s *DeleteLoadBalancerRequest) SetLoadBalancerId(v string) *DeleteLoadBalancerRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *DeleteLoadBalancerRequest) SetRegionId(v string) *DeleteLoadBalancerRequest {
	s.RegionId = &v
	return s
}

func (s *DeleteLoadBalancerRequest) Validate() error {
	return dara.Validate(s)
}
