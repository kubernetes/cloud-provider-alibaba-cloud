// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListSecurityPolicyResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListSecurityPolicyResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListSecurityPolicyResponse
	GetStatusCode() *int32
	SetBody(v *ListSecurityPolicyResponseBody) *ListSecurityPolicyResponse
	GetBody() *ListSecurityPolicyResponseBody
}

type ListSecurityPolicyResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListSecurityPolicyResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListSecurityPolicyResponse) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyResponse) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListSecurityPolicyResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListSecurityPolicyResponse) GetBody() *ListSecurityPolicyResponseBody {
	return s.Body
}

func (s *ListSecurityPolicyResponse) SetHeaders(v map[string]*string) *ListSecurityPolicyResponse {
	s.Headers = v
	return s
}

func (s *ListSecurityPolicyResponse) SetStatusCode(v int32) *ListSecurityPolicyResponse {
	s.StatusCode = &v
	return s
}

func (s *ListSecurityPolicyResponse) SetBody(v *ListSecurityPolicyResponseBody) *ListSecurityPolicyResponse {
	s.Body = v
	return s
}

func (s *ListSecurityPolicyResponse) Validate() error {
	return dara.Validate(s)
}
