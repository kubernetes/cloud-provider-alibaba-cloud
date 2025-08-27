// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListServerGroupServersRequest interface {
	dara.Model
	String() string
	GoString() string
	SetMaxResults(v int32) *ListServerGroupServersRequest
	GetMaxResults() *int32
	SetNextToken(v string) *ListServerGroupServersRequest
	GetNextToken() *string
	SetRegionId(v string) *ListServerGroupServersRequest
	GetRegionId() *string
	SetServerGroupId(v string) *ListServerGroupServersRequest
	GetServerGroupId() *string
	SetServerIds(v []*string) *ListServerGroupServersRequest
	GetServerIds() []*string
	SetServerIps(v []*string) *ListServerGroupServersRequest
	GetServerIps() []*string
}

type ListServerGroupServersRequest struct {
	// The number of entries to return in each call. Valid values: **1*	- to **100**. Default value: **20**.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The pagination token used to specify a particular page of results. Valid values:
	//
	// 	- Left this parameter empty if this is the first query or the only query.
	//
	// 	- Set this parameter to the value of NextToken obtained from the previous query.
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
	// The ID of the server group.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The IDs of the backend servers. You can specify up to 40 backend servers in each call.
	ServerIds []*string `json:"ServerIds,omitempty" xml:"ServerIds,omitempty" type:"Repeated"`
	// The IP addresses of the backend servers. You can specify up to 40 backend servers in each call.
	ServerIps []*string `json:"ServerIps,omitempty" xml:"ServerIps,omitempty" type:"Repeated"`
}

func (s ListServerGroupServersRequest) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupServersRequest) GoString() string {
	return s.String()
}

func (s *ListServerGroupServersRequest) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListServerGroupServersRequest) GetNextToken() *string {
	return s.NextToken
}

func (s *ListServerGroupServersRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *ListServerGroupServersRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *ListServerGroupServersRequest) GetServerIds() []*string {
	return s.ServerIds
}

func (s *ListServerGroupServersRequest) GetServerIps() []*string {
	return s.ServerIps
}

func (s *ListServerGroupServersRequest) SetMaxResults(v int32) *ListServerGroupServersRequest {
	s.MaxResults = &v
	return s
}

func (s *ListServerGroupServersRequest) SetNextToken(v string) *ListServerGroupServersRequest {
	s.NextToken = &v
	return s
}

func (s *ListServerGroupServersRequest) SetRegionId(v string) *ListServerGroupServersRequest {
	s.RegionId = &v
	return s
}

func (s *ListServerGroupServersRequest) SetServerGroupId(v string) *ListServerGroupServersRequest {
	s.ServerGroupId = &v
	return s
}

func (s *ListServerGroupServersRequest) SetServerIds(v []*string) *ListServerGroupServersRequest {
	s.ServerIds = v
	return s
}

func (s *ListServerGroupServersRequest) SetServerIps(v []*string) *ListServerGroupServersRequest {
	s.ServerIps = v
	return s
}

func (s *ListServerGroupServersRequest) Validate() error {
	return dara.Validate(s)
}
