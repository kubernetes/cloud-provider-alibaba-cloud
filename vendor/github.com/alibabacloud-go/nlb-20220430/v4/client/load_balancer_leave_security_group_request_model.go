// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iLoadBalancerLeaveSecurityGroupRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *LoadBalancerLeaveSecurityGroupRequest
	GetClientToken() *string
	SetDryRun(v bool) *LoadBalancerLeaveSecurityGroupRequest
	GetDryRun() *bool
	SetLoadBalancerId(v string) *LoadBalancerLeaveSecurityGroupRequest
	GetLoadBalancerId() *string
	SetRegionId(v string) *LoadBalancerLeaveSecurityGroupRequest
	GetRegionId() *string
	SetSecurityGroupIds(v []*string) *LoadBalancerLeaveSecurityGroupRequest
	GetSecurityGroupIds() []*string
}

type LoadBalancerLeaveSecurityGroupRequest struct {
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
	// Specifies whether to perform a dry run, without sending the actual request. Valid values:
	//
	// 	- **true**: checks the request without performing the operation. The system checks the required parameters, request syntax, and limits. If the request fails the dry run, an error message is returned. If the request passes the dry run, the `DryRunOperation` error code is returned.
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
	// The ID of the security group to be disassociated.
	//
	// This parameter is required.
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" xml:"SecurityGroupIds,omitempty" type:"Repeated"`
}

func (s LoadBalancerLeaveSecurityGroupRequest) String() string {
	return dara.Prettify(s)
}

func (s LoadBalancerLeaveSecurityGroupRequest) GoString() string {
	return s.String()
}

func (s *LoadBalancerLeaveSecurityGroupRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *LoadBalancerLeaveSecurityGroupRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *LoadBalancerLeaveSecurityGroupRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *LoadBalancerLeaveSecurityGroupRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *LoadBalancerLeaveSecurityGroupRequest) GetSecurityGroupIds() []*string {
	return s.SecurityGroupIds
}

func (s *LoadBalancerLeaveSecurityGroupRequest) SetClientToken(v string) *LoadBalancerLeaveSecurityGroupRequest {
	s.ClientToken = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupRequest) SetDryRun(v bool) *LoadBalancerLeaveSecurityGroupRequest {
	s.DryRun = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupRequest) SetLoadBalancerId(v string) *LoadBalancerLeaveSecurityGroupRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupRequest) SetRegionId(v string) *LoadBalancerLeaveSecurityGroupRequest {
	s.RegionId = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupRequest) SetSecurityGroupIds(v []*string) *LoadBalancerLeaveSecurityGroupRequest {
	s.SecurityGroupIds = v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupRequest) Validate() error {
	return dara.Validate(s)
}
