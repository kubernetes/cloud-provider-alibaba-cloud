// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListListenersRequest interface {
	dara.Model
	String() string
	GoString() string
	SetListenerIds(v []*string) *ListListenersRequest
	GetListenerIds() []*string
	SetListenerProtocol(v string) *ListListenersRequest
	GetListenerProtocol() *string
	SetLoadBalancerIds(v []*string) *ListListenersRequest
	GetLoadBalancerIds() []*string
	SetMaxResults(v int32) *ListListenersRequest
	GetMaxResults() *int32
	SetNextToken(v string) *ListListenersRequest
	GetNextToken() *string
	SetRegionId(v string) *ListListenersRequest
	GetRegionId() *string
	SetSecSensorEnabled(v string) *ListListenersRequest
	GetSecSensorEnabled() *string
	SetTag(v []*ListListenersRequestTag) *ListListenersRequest
	GetTag() []*ListListenersRequestTag
}

type ListListenersRequest struct {
	// The listener IDs. You can specify up to 20 listeners.
	ListenerIds []*string `json:"ListenerIds,omitempty" xml:"ListenerIds,omitempty" type:"Repeated"`
	// The listener protocol. Valid values: **TCP**, **UDP**, and **TCPSSL**.
	//
	// example:
	//
	// TCPSSL
	ListenerProtocol *string `json:"ListenerProtocol,omitempty" xml:"ListenerProtocol,omitempty"`
	// The IDs of the NLB instances. You can specify up to 20 instances.
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" xml:"LoadBalancerIds,omitempty" type:"Repeated"`
	// The number of entries to return in each call. Valid values: **1*	- to **100**. Default value: **20**
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The pagination token used to specify a particular page of results. Valid values:
	//
	// 	- Leave this parameter empty for the first query or the only query.
	//
	// 	- Set this parameter to the value of NextToken obtained from the previous query.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// Specifies whether to enable fine-grained monitoring. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	SecSensorEnabled *string `json:"SecSensorEnabled,omitempty" xml:"SecSensorEnabled,omitempty"`
	// The tags.
	Tag []*ListListenersRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
}

func (s ListListenersRequest) String() string {
	return dara.Prettify(s)
}

func (s ListListenersRequest) GoString() string {
	return s.String()
}

func (s *ListListenersRequest) GetListenerIds() []*string {
	return s.ListenerIds
}

func (s *ListListenersRequest) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *ListListenersRequest) GetLoadBalancerIds() []*string {
	return s.LoadBalancerIds
}

func (s *ListListenersRequest) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListListenersRequest) GetNextToken() *string {
	return s.NextToken
}

func (s *ListListenersRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *ListListenersRequest) GetSecSensorEnabled() *string {
	return s.SecSensorEnabled
}

func (s *ListListenersRequest) GetTag() []*ListListenersRequestTag {
	return s.Tag
}

func (s *ListListenersRequest) SetListenerIds(v []*string) *ListListenersRequest {
	s.ListenerIds = v
	return s
}

func (s *ListListenersRequest) SetListenerProtocol(v string) *ListListenersRequest {
	s.ListenerProtocol = &v
	return s
}

func (s *ListListenersRequest) SetLoadBalancerIds(v []*string) *ListListenersRequest {
	s.LoadBalancerIds = v
	return s
}

func (s *ListListenersRequest) SetMaxResults(v int32) *ListListenersRequest {
	s.MaxResults = &v
	return s
}

func (s *ListListenersRequest) SetNextToken(v string) *ListListenersRequest {
	s.NextToken = &v
	return s
}

func (s *ListListenersRequest) SetRegionId(v string) *ListListenersRequest {
	s.RegionId = &v
	return s
}

func (s *ListListenersRequest) SetSecSensorEnabled(v string) *ListListenersRequest {
	s.SecSensorEnabled = &v
	return s
}

func (s *ListListenersRequest) SetTag(v []*ListListenersRequestTag) *ListListenersRequest {
	s.Tag = v
	return s
}

func (s *ListListenersRequest) Validate() error {
	return dara.Validate(s)
}

type ListListenersRequestTag struct {
	// The key of the tag. You can specify up to 20 tags. The tag key cannot be an empty string.
	//
	// It can be up to 64 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// env
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. You can specify up to 20 tags. The tag value can be an empty string.
	//
	// It can be up to 128 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`.
	//
	// example:
	//
	// product
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListListenersRequestTag) String() string {
	return dara.Prettify(s)
}

func (s ListListenersRequestTag) GoString() string {
	return s.String()
}

func (s *ListListenersRequestTag) GetKey() *string {
	return s.Key
}

func (s *ListListenersRequestTag) GetValue() *string {
	return s.Value
}

func (s *ListListenersRequestTag) SetKey(v string) *ListListenersRequestTag {
	s.Key = &v
	return s
}

func (s *ListListenersRequestTag) SetValue(v string) *ListListenersRequestTag {
	s.Value = &v
	return s
}

func (s *ListListenersRequestTag) Validate() error {
	return dara.Validate(s)
}
