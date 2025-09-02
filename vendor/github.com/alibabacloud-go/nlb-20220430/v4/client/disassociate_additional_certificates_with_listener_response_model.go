// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDisassociateAdditionalCertificatesWithListenerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DisassociateAdditionalCertificatesWithListenerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DisassociateAdditionalCertificatesWithListenerResponse
	GetStatusCode() *int32
	SetBody(v *DisassociateAdditionalCertificatesWithListenerResponseBody) *DisassociateAdditionalCertificatesWithListenerResponse
	GetBody() *DisassociateAdditionalCertificatesWithListenerResponseBody
}

type DisassociateAdditionalCertificatesWithListenerResponse struct {
	Headers    map[string]*string                                          `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                                      `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DisassociateAdditionalCertificatesWithListenerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DisassociateAdditionalCertificatesWithListenerResponse) String() string {
	return dara.Prettify(s)
}

func (s DisassociateAdditionalCertificatesWithListenerResponse) GoString() string {
	return s.String()
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) GetBody() *DisassociateAdditionalCertificatesWithListenerResponseBody {
	return s.Body
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) SetHeaders(v map[string]*string) *DisassociateAdditionalCertificatesWithListenerResponse {
	s.Headers = v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) SetStatusCode(v int32) *DisassociateAdditionalCertificatesWithListenerResponse {
	s.StatusCode = &v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) SetBody(v *DisassociateAdditionalCertificatesWithListenerResponseBody) *DisassociateAdditionalCertificatesWithListenerResponse {
	s.Body = v
	return s
}

func (s *DisassociateAdditionalCertificatesWithListenerResponse) Validate() error {
	return dara.Validate(s)
}
