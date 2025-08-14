// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDisableLoadBalancerIpv6InternetRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *DisableLoadBalancerIpv6InternetRequest
	GetClientToken() *string
	SetDryRun(v bool) *DisableLoadBalancerIpv6InternetRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *DisableLoadBalancerIpv6InternetRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *DisableLoadBalancerIpv6InternetRequest
	GetRegionId() *string
}

type DisableLoadBalancerIpv6InternetRequest struct {
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
}

func (s DisableLoadBalancerIpv6InternetRequest) String() string {
	return dara.Prettify(s)
}

func (s DisableLoadBalancerIpv6InternetRequest) GoString() string {
	return s.String()
}

func (s *DisableLoadBalancerIpv6InternetRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *DisableLoadBalancerIpv6InternetRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *DisableLoadBalancerIpv6InternetRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *DisableLoadBalancerIpv6InternetRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *DisableLoadBalancerIpv6InternetRequest) SetClientToken(v string) *DisableLoadBalancerIpv6InternetRequest {
	s.ClientToken = &v
	return s
}

func (s *DisableLoadBalancerIpv6InternetRequest) SetDryRun(v bool) *DisableLoadBalancerIpv6InternetRequest {
	s.DryRun = &v
	return s
}

func (s *DisableLoadBalancerIpv6InternetRequest) SetLoadBalancerId(v string) *DisableLoadBalancerIpv6InternetRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *DisableLoadBalancerIpv6InternetRequest) SetRegionId(v string) *DisableLoadBalancerIpv6InternetRequest {
	s.RegionId = &v
	return s
}

func (s *DisableLoadBalancerIpv6InternetRequest) Validate() error {
	return dara.Validate(s)
}
