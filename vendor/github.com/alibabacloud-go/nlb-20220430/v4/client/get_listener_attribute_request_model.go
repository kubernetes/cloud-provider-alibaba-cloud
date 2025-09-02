// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetListenerAttributeRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *GetListenerAttributeRequest
	GetClientToken() *string
	SetDryRun(v bool) *GetListenerAttributeRequest
	GetDryRun() *bool
	SetListenerId(v string) *GetListenerAttributeRequest
	GetListenerId() *string
	SetRegionId(v string) *GetListenerAttributeRequest
	GetRegionId() *string
}

type GetListenerAttributeRequest struct {
	// The client token that is used to ensure the idempotence of the request.
	//
	// You can use the client to generate the value, but you must ensure that it is unique among all requests. ClientToken can contain only ASCII characters.
	//
	// >  If you do not set this parameter, **ClientToken*	- is set to the value of **RequestId**. The value of **RequestId*	- is different for each request.
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
	// The listener ID.
	//
	// This parameter is required.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@233
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s GetListenerAttributeRequest) String() string {
	return dara.Prettify(s)
}

func (s GetListenerAttributeRequest) GoString() string {
	return s.String()
}

func (s *GetListenerAttributeRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *GetListenerAttributeRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *GetListenerAttributeRequest) GetListenerId() *string {
	return s.ListenerId
}

func (s *GetListenerAttributeRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *GetListenerAttributeRequest) SetClientToken(v string) *GetListenerAttributeRequest {
	s.ClientToken = &v
	return s
}

func (s *GetListenerAttributeRequest) SetDryRun(v bool) *GetListenerAttributeRequest {
	s.DryRun = &v
	return s
}

func (s *GetListenerAttributeRequest) SetListenerId(v string) *GetListenerAttributeRequest {
	s.ListenerId = &v
	return s
}

func (s *GetListenerAttributeRequest) SetRegionId(v string) *GetListenerAttributeRequest {
	s.RegionId = &v
	return s
}

func (s *GetListenerAttributeRequest) Validate() error {
	return dara.Validate(s)
}
