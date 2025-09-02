// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAssociateAdditionalCertificatesWithListenerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *AssociateAdditionalCertificatesWithListenerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *AssociateAdditionalCertificatesWithListenerResponse
	GetStatusCode() *int32
	SetBody(v *AssociateAdditionalCertificatesWithListenerResponseBody) *AssociateAdditionalCertificatesWithListenerResponse
	GetBody() *AssociateAdditionalCertificatesWithListenerResponseBody
}

type AssociateAdditionalCertificatesWithListenerResponse struct {
	Headers    map[string]*string                                       `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                                   `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *AssociateAdditionalCertificatesWithListenerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s AssociateAdditionalCertificatesWithListenerResponse) String() string {
	return dara.Prettify(s)
}

func (s AssociateAdditionalCertificatesWithListenerResponse) GoString() string {
	return s.String()
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) GetBody() *AssociateAdditionalCertificatesWithListenerResponseBody {
	return s.Body
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) SetHeaders(v map[string]*string) *AssociateAdditionalCertificatesWithListenerResponse {
	s.Headers = v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) SetStatusCode(v int32) *AssociateAdditionalCertificatesWithListenerResponse {
	s.StatusCode = &v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) SetBody(v *AssociateAdditionalCertificatesWithListenerResponseBody) *AssociateAdditionalCertificatesWithListenerResponse {
	s.Body = v
	return s
}

func (s *AssociateAdditionalCertificatesWithListenerResponse) Validate() error {
	return dara.Validate(s)
}
