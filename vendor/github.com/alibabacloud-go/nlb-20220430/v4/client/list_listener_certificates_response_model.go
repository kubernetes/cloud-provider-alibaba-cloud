// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListListenerCertificatesResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListListenerCertificatesResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListListenerCertificatesResponse
	GetStatusCode() *int32
	SetBody(v *ListListenerCertificatesResponseBody) *ListListenerCertificatesResponse
	GetBody() *ListListenerCertificatesResponseBody
}

type ListListenerCertificatesResponse struct {
	Headers    map[string]*string                    `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListListenerCertificatesResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListListenerCertificatesResponse) String() string {
	return dara.Prettify(s)
}

func (s ListListenerCertificatesResponse) GoString() string {
	return s.String()
}

func (s *ListListenerCertificatesResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListListenerCertificatesResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListListenerCertificatesResponse) GetBody() *ListListenerCertificatesResponseBody {
	return s.Body
}

func (s *ListListenerCertificatesResponse) SetHeaders(v map[string]*string) *ListListenerCertificatesResponse {
	s.Headers = v
	return s
}

func (s *ListListenerCertificatesResponse) SetStatusCode(v int32) *ListListenerCertificatesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListListenerCertificatesResponse) SetBody(v *ListListenerCertificatesResponseBody) *ListListenerCertificatesResponse {
	s.Body = v
	return s
}

func (s *ListListenerCertificatesResponse) Validate() error {
	return dara.Validate(s)
}
