// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListSecurityPolicyRequest interface {
	dara.Model
	String() string
	GoString() string
	SetMaxResults(v int32) *ListSecurityPolicyRequest
	GetMaxResults() *int32
	SetNextToken(v string) *ListSecurityPolicyRequest
	GetNextToken() *string
	SetRegionId(v string) *ListSecurityPolicyRequest
	GetRegionId() *string
	SetResourceGroupId(v string) *ListSecurityPolicyRequest
	GetResourceGroupId() *string
	SetSecurityPolicyIds(v []*string) *ListSecurityPolicyRequest
	GetSecurityPolicyIds() []*string
	SetSecurityPolicyNames(v []*string) *ListSecurityPolicyRequest
	GetSecurityPolicyNames() []*string
	SetTag(v []*ListSecurityPolicyRequestTag) *ListSecurityPolicyRequest
	GetTag() []*ListSecurityPolicyRequestTag
}

type ListSecurityPolicyRequest struct {
	// The number of entries to return per page. Valid values: **1*	- to **100**. Default value: **20**.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The pagination token that is used in the next request to retrieve a new page of results. Valid values:
	//
	// 	- You do not need to specify this parameter for the first request.
	//
	// 	- You must specify the token that is obtained from the previous query as the value of NextToken.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The resource group ID.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The IDs of the TLS security policies. You can specify at most 20 policy IDs in each call.
	SecurityPolicyIds []*string `json:"SecurityPolicyIds,omitempty" xml:"SecurityPolicyIds,omitempty" type:"Repeated"`
	// The names of the TLS security policies. You can specify at most 20 policy names.
	SecurityPolicyNames []*string `json:"SecurityPolicyNames,omitempty" xml:"SecurityPolicyNames,omitempty" type:"Repeated"`
	// The tags.
	Tag []*ListSecurityPolicyRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
}

func (s ListSecurityPolicyRequest) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyRequest) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyRequest) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListSecurityPolicyRequest) GetNextToken() *string {
	return s.NextToken
}

func (s *ListSecurityPolicyRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *ListSecurityPolicyRequest) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *ListSecurityPolicyRequest) GetSecurityPolicyIds() []*string {
	return s.SecurityPolicyIds
}

func (s *ListSecurityPolicyRequest) GetSecurityPolicyNames() []*string {
	return s.SecurityPolicyNames
}

func (s *ListSecurityPolicyRequest) GetTag() []*ListSecurityPolicyRequestTag {
	return s.Tag
}

func (s *ListSecurityPolicyRequest) SetMaxResults(v int32) *ListSecurityPolicyRequest {
	s.MaxResults = &v
	return s
}

func (s *ListSecurityPolicyRequest) SetNextToken(v string) *ListSecurityPolicyRequest {
	s.NextToken = &v
	return s
}

func (s *ListSecurityPolicyRequest) SetRegionId(v string) *ListSecurityPolicyRequest {
	s.RegionId = &v
	return s
}

func (s *ListSecurityPolicyRequest) SetResourceGroupId(v string) *ListSecurityPolicyRequest {
	s.ResourceGroupId = &v
	return s
}

func (s *ListSecurityPolicyRequest) SetSecurityPolicyIds(v []*string) *ListSecurityPolicyRequest {
	s.SecurityPolicyIds = v
	return s
}

func (s *ListSecurityPolicyRequest) SetSecurityPolicyNames(v []*string) *ListSecurityPolicyRequest {
	s.SecurityPolicyNames = v
	return s
}

func (s *ListSecurityPolicyRequest) SetTag(v []*ListSecurityPolicyRequestTag) *ListSecurityPolicyRequest {
	s.Tag = v
	return s
}

func (s *ListSecurityPolicyRequest) Validate() error {
	return dara.Validate(s)
}

type ListSecurityPolicyRequestTag struct {
	// The tag key. You can specify up to 10 tag keys.
	//
	// The tag key can be up to 64 characters in length. It cannot start with `aliyun` or `acs:` and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// Test
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The tag value. You can specify up to 10 tag values.
	//
	// The tag value can be up to 128 characters in length. It cannot start with `aliyun` or `acs:` and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// Test
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListSecurityPolicyRequestTag) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyRequestTag) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyRequestTag) GetKey() *string {
	return s.Key
}

func (s *ListSecurityPolicyRequestTag) GetValue() *string {
	return s.Value
}

func (s *ListSecurityPolicyRequestTag) SetKey(v string) *ListSecurityPolicyRequestTag {
	s.Key = &v
	return s
}

func (s *ListSecurityPolicyRequestTag) SetValue(v string) *ListSecurityPolicyRequestTag {
	s.Value = &v
	return s
}

func (s *ListSecurityPolicyRequestTag) Validate() error {
	return dara.Validate(s)
}
