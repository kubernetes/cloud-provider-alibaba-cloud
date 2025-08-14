// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetListenerHealthStatusRequest interface {
	dara.Model
	String() string
	GoString() string
	SetListenerId(v string) *GetListenerHealthStatusRequest
	GetListenerId() *string
	SetRegionId(v string) *GetListenerHealthStatusRequest
	GetRegionId() *string
}

type GetListenerHealthStatusRequest struct {
	// The ID of the listener on the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@80
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s GetListenerHealthStatusRequest) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusRequest) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusRequest) GetListenerId() *string {
	return s.ListenerId
}

func (s *GetListenerHealthStatusRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *GetListenerHealthStatusRequest) SetListenerId(v string) *GetListenerHealthStatusRequest {
	s.ListenerId = &v
	return s
}

func (s *GetListenerHealthStatusRequest) SetRegionId(v string) *GetListenerHealthStatusRequest {
	s.RegionId = &v
	return s
}

func (s *GetListenerHealthStatusRequest) Validate() error {
	return dara.Validate(s)
}
