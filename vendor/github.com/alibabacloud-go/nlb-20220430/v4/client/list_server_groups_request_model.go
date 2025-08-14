// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListServerGroupsRequest interface {
	dara.Model
	String() string
	GoString() string
	SetMaxResults(v int32) *ListServerGroupsRequest
	GetMaxResults() *int32
	SetNextToken(v string) *ListServerGroupsRequest
	GetNextToken() *string
	SetRegionId(v string) *ListServerGroupsRequest
	GetRegionId() *string
	SetResourceGroupId(v string) *ListServerGroupsRequest
	GetResourceGroupId() *string
	SetServerGroupIds(v []*string) *ListServerGroupsRequest
	GetServerGroupIds() []*string
	SetServerGroupNames(v []*string) *ListServerGroupsRequest
	GetServerGroupNames() []*string
	SetServerGroupType(v string) *ListServerGroupsRequest
	GetServerGroupType() *string
	SetTag(v []*ListServerGroupsRequestTag) *ListServerGroupsRequest
	GetTag() []*ListServerGroupsRequestTag
	SetVpcId(v string) *ListServerGroupsRequest
	GetVpcId() *string
}

type ListServerGroupsRequest struct {
	// The number of entries per page. Valid values: **1*	- to **100**. Default value: **20**.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The pagination token used in the next request to retrieve a new page of results. Valid values:
	//
	// 	- For the first request and last request, you do not need to specify this parameter.
	//
	// 	- You must specify the token obtained from the previous query as the value of NextToken.
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
	// The ID of the resource group to which the server group belongs.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The server group IDs. You can specify up to 20 server group IDs in each call.
	ServerGroupIds []*string `json:"ServerGroupIds,omitempty" xml:"ServerGroupIds,omitempty" type:"Repeated"`
	// The names of the server groups to be queried. You can specify up to 20 names in each call.
	ServerGroupNames []*string `json:"ServerGroupNames,omitempty" xml:"ServerGroupNames,omitempty" type:"Repeated"`
	// The type of server group. Valid values:
	//
	// 	- **Instance**: allows you to add servers of the **Ecs**, **Ens**, and **Eci*	- types.
	//
	// 	- **Ip**: allows you to add servers by specifying IP addresses.
	//
	// example:
	//
	// Instance
	ServerGroupType *string `json:"ServerGroupType,omitempty" xml:"ServerGroupType,omitempty"`
	// The tags.
	Tag []*ListServerGroupsRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
	// The ID of the virtual private cloud (VPC) in which the server group is deployed.
	//
	// example:
	//
	// vpc-bp15zckdt37pq72zv****
	VpcId *string `json:"VpcId,omitempty" xml:"VpcId,omitempty"`
}

func (s ListServerGroupsRequest) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsRequest) GoString() string {
	return s.String()
}

func (s *ListServerGroupsRequest) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListServerGroupsRequest) GetNextToken() *string {
	return s.NextToken
}

func (s *ListServerGroupsRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *ListServerGroupsRequest) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *ListServerGroupsRequest) GetServerGroupIds() []*string {
	return s.ServerGroupIds
}

func (s *ListServerGroupsRequest) GetServerGroupNames() []*string {
	return s.ServerGroupNames
}

func (s *ListServerGroupsRequest) GetServerGroupType() *string {
	return s.ServerGroupType
}

func (s *ListServerGroupsRequest) GetTag() []*ListServerGroupsRequestTag {
	return s.Tag
}

func (s *ListServerGroupsRequest) GetVpcId() *string {
	return s.VpcId
}

func (s *ListServerGroupsRequest) SetMaxResults(v int32) *ListServerGroupsRequest {
	s.MaxResults = &v
	return s
}

func (s *ListServerGroupsRequest) SetNextToken(v string) *ListServerGroupsRequest {
	s.NextToken = &v
	return s
}

func (s *ListServerGroupsRequest) SetRegionId(v string) *ListServerGroupsRequest {
	s.RegionId = &v
	return s
}

func (s *ListServerGroupsRequest) SetResourceGroupId(v string) *ListServerGroupsRequest {
	s.ResourceGroupId = &v
	return s
}

func (s *ListServerGroupsRequest) SetServerGroupIds(v []*string) *ListServerGroupsRequest {
	s.ServerGroupIds = v
	return s
}

func (s *ListServerGroupsRequest) SetServerGroupNames(v []*string) *ListServerGroupsRequest {
	s.ServerGroupNames = v
	return s
}

func (s *ListServerGroupsRequest) SetServerGroupType(v string) *ListServerGroupsRequest {
	s.ServerGroupType = &v
	return s
}

func (s *ListServerGroupsRequest) SetTag(v []*ListServerGroupsRequestTag) *ListServerGroupsRequest {
	s.Tag = v
	return s
}

func (s *ListServerGroupsRequest) SetVpcId(v string) *ListServerGroupsRequest {
	s.VpcId = &v
	return s
}

func (s *ListServerGroupsRequest) Validate() error {
	return dara.Validate(s)
}

type ListServerGroupsRequestTag struct {
	// The key of the tag. You can specify up to 10 tag keys.
	//
	// The tag key can be up to 64 characters in length. It cannot start with `aliyun` or `acs:` and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// Test
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. You can specify up to 10 tag values.
	//
	// The tag value can be up to 128 characters in length. It cannot start with `aliyun` and `acs:`, and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// Test
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListServerGroupsRequestTag) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsRequestTag) GoString() string {
	return s.String()
}

func (s *ListServerGroupsRequestTag) GetKey() *string {
	return s.Key
}

func (s *ListServerGroupsRequestTag) GetValue() *string {
	return s.Value
}

func (s *ListServerGroupsRequestTag) SetKey(v string) *ListServerGroupsRequestTag {
	s.Key = &v
	return s
}

func (s *ListServerGroupsRequestTag) SetValue(v string) *ListServerGroupsRequestTag {
	s.Value = &v
	return s
}

func (s *ListServerGroupsRequestTag) Validate() error {
	return dara.Validate(s)
}
