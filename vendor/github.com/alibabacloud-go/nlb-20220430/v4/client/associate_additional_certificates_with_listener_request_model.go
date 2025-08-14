// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAssociateAdditionalCertificatesWithListenerRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAdditionalCertificateIds(v []*string) *AssociateAdditionalCertificatesWithListenerRequest
	GetAdditionalCertificateIds() []*string
	SetClientToken(v string) *AssociateAdditionalCertificatesWithListenerRequest
	GetClientToken() *string
	SetDryRun(v bool) *AssociateAdditionalCertificatesWithListenerRequest
	GetDryRun() *bool
	SetListenerId(v string) *AssociateAdditionalCertificatesWithListenerRequest
	GetListenerId() *string
	SetRegionId(v string) *AssociateAdditionalCertificatesWithListenerRequest
	GetRegionId() *string
}

type AssociateAdditionalCertificatesWithListenerRequest struct {
	// The additional certificates. You can associate up to 15 additional certificates with a listener in each call.
	//
	// This parameter is required.
	AdditionalCertificateIds []*string `json:"AdditionalCertificateIds,omitempty" xml:"AdditionalCertificateIds,omitempty" type:"Repeated"`
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. Only ASCII characters are allowed.
	//
	// >  If you do not set this parameter, the value of **RequestId*	- is used.***	- The value of **RequestId*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// Specifies whether to perform a dry run. Valid values:
	//
	// 	- **true**: validates the request without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the validation, the corresponding error message is returned. If the request passes the validation, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): validates the request and performs the operation. If the request passes the validation, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// true
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The listener ID. Only TCPSSL listener IDs are supported.
	//
	// This parameter is required.
	//
	// example:
	//
	// lsn-bpn0kn908w4nbw****@80
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

func (s AssociateAdditionalCertificatesWithListenerRequest) String() string {
	return dara.Prettify(s)
}

func (s AssociateAdditionalCertificatesWithListenerRequest) GoString() string {
	return s.String()
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) GetAdditionalCertificateIds() []*string {
	return s.AdditionalCertificateIds
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) GetListenerId() *string {
	return s.ListenerId
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) SetAdditionalCertificateIds(v []*string) *AssociateAdditionalCertificatesWithListenerRequest {
	s.AdditionalCertificateIds = v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) SetClientToken(v string) *AssociateAdditionalCertificatesWithListenerRequest {
	s.ClientToken = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) SetDryRun(v bool) *AssociateAdditionalCertificatesWithListenerRequest {
	s.DryRun = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) SetListenerId(v string) *AssociateAdditionalCertificatesWithListenerRequest {
	s.ListenerId = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) SetRegionId(v string) *AssociateAdditionalCertificatesWithListenerRequest {
	s.RegionId = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerRequest) Validate() error {
	return dara.Validate(s)
}
