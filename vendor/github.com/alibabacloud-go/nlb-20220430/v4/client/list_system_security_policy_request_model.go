// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListSystemSecurityPolicyRequest interface {
	dara.Model
	String() string
	GoString() string
	SetRegionId(v string) *ListSystemSecurityPolicyRequest
	GetRegionId() *string
}

type ListSystemSecurityPolicyRequest struct {
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s ListSystemSecurityPolicyRequest) String() string {
	return dara.Prettify(s)
}

func (s ListSystemSecurityPolicyRequest) GoString() string {
	return s.String()
}

func (s *ListSystemSecurityPolicyRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *ListSystemSecurityPolicyRequest) SetRegionId(v string) *ListSystemSecurityPolicyRequest {
	s.RegionId = &v
	return s
}

func (s *ListSystemSecurityPolicyRequest) Validate() error {
	return dara.Validate(s)
}
