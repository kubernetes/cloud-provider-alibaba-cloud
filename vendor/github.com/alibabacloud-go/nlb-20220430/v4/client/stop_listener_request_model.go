// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStopListenerRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *StopListenerRequest
	GetClientToken() *string
	SetDryRun(v bool) *StopListenerRequest
	GetDryRun() *bool
	SetListenerId(v string) *StopListenerRequest
	GetListenerId() *string
	SetRegionId(v string) *StopListenerRequest
	GetRegionId() *string
}

type StopListenerRequest struct {
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
	// 	- **true**: performs a dry run without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the dry run, an error message is returned. If the request passes the dry run, the `DryRunOperation` error code is returned.
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
	// lsn-bp1bpn0kn908w4nbw****@80
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

func (s StopListenerRequest) String() string {
	return dara.Prettify(s)
}

func (s StopListenerRequest) GoString() string {
	return s.String()
}

func (s *StopListenerRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *StopListenerRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *StopListenerRequest) GetListenerId() *string {
	return s.ListenerId
}

func (s *StopListenerRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *StopListenerRequest) SetClientToken(v string) *StopListenerRequest {
	s.ClientToken = &v
	return s
}

func (s *StopListenerRequest) SetDryRun(v bool) *StopListenerRequest {
	s.DryRun = &v
	return s
}

func (s *StopListenerRequest) SetListenerId(v string) *StopListenerRequest {
	s.ListenerId = &v
	return s
}

func (s *StopListenerRequest) SetRegionId(v string) *StopListenerRequest {
	s.RegionId = &v
	return s
}

func (s *StopListenerRequest) Validate() error {
	return dara.Validate(s)
}
