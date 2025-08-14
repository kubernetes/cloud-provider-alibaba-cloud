// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDisassociateAdditionalCertificatesWithListenerRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAdditionalCertificateIds(v []*string) *DisassociateAdditionalCertificatesWithListenerRequest
	GetAdditionalCertificateIds() []*string
	SetClientToken(v string) *DisassociateAdditionalCertificatesWithListenerRequest
	GetClientToken() *string
	SetDryRun(v bool) *DisassociateAdditionalCertificatesWithListenerRequest
	GetDryRun() *bool
	SetListenerId(v string) *DisassociateAdditionalCertificatesWithListenerRequest
	GetListenerId() *string
	SetRegionId(v string) *DisassociateAdditionalCertificatesWithListenerRequest
	GetRegionId() *string
}

type DisassociateAdditionalCertificatesWithListenerRequest struct {
	// The additional certificates. You can disassociate up to 15 additional certificates in each call.
	//
	// This parameter is required.
	AdditionalCertificateIds []*string `json:"AdditionalCertificateIds,omitempty" xml:"AdditionalCertificateIds,omitempty" type:"Repeated"`
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate this value. Ensure that the value is unique among all requests. Only ASCII characters are allowed.
	//
	// >  If you do not specify this parameter, the value of **RequestId*	- is used.***	- **RequestId*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// Specifies whether to perform a dry run. Valid values:
	//
	// 	- **true**: Validates the request without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the validation, the corresponding error message is returned. If the request passes the validation, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): validates the request and performs the operation. If the request passes the validation, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// true
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The listener ID. Only TCP/SSL listener IDs are supported.
	//
	// This parameter is required.
	//
	// example:
	//
	// lsn-bpn0kn908w4nbw****@80
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

func (s DisassociateAdditionalCertificatesWithListenerRequest) String() string {
	return dara.Prettify(s)
}

func (s DisassociateAdditionalCertificatesWithListenerRequest) GoString() string {
	return s.String()
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) GetAdditionalCertificateIds() []*string {
	return s.AdditionalCertificateIds
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) GetListenerId() *string {
	return s.ListenerId
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) SetAdditionalCertificateIds(v []*string) *DisassociateAdditionalCertificatesWithListenerRequest {
	s.AdditionalCertificateIds = v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) SetClientToken(v string) *DisassociateAdditionalCertificatesWithListenerRequest {
	s.ClientToken = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) SetDryRun(v bool) *DisassociateAdditionalCertificatesWithListenerRequest {
	s.DryRun = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) SetListenerId(v string) *DisassociateAdditionalCertificatesWithListenerRequest {
	s.ListenerId = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) SetRegionId(v string) *DisassociateAdditionalCertificatesWithListenerRequest {
	s.RegionId = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerRequest) Validate() error {
	return dara.Validate(s)
}
