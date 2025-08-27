// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListSystemSecurityPolicyResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListSystemSecurityPolicyResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListSystemSecurityPolicyResponse
	GetStatusCode() *int32
	SetBody(v *ListSystemSecurityPolicyResponseBody) *ListSystemSecurityPolicyResponse
	GetBody() *ListSystemSecurityPolicyResponseBody
}

type ListSystemSecurityPolicyResponse struct {
	Headers    map[string]*string                    `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListSystemSecurityPolicyResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListSystemSecurityPolicyResponse) String() string {
	return dara.Prettify(s)
}

func (s ListSystemSecurityPolicyResponse) GoString() string {
	return s.String()
}

func (s *ListSystemSecurityPolicyResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListSystemSecurityPolicyResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListSystemSecurityPolicyResponse) GetBody() *ListSystemSecurityPolicyResponseBody {
	return s.Body
}

func (s *ListSystemSecurityPolicyResponse) SetHeaders(v map[string]*string) *ListSystemSecurityPolicyResponse {
	s.Headers = v
	return s
}

func (s *ListSystemSecurityPolicyResponse) SetStatusCode(v int32) *ListSystemSecurityPolicyResponse {
	s.StatusCode = &v
	return s
}

func (s *ListSystemSecurityPolicyResponse) SetBody(v *ListSystemSecurityPolicyResponseBody) *ListSystemSecurityPolicyResponse {
	s.Body = v
	return s
}

func (s *ListSystemSecurityPolicyResponse) Validate() error {
	return dara.Validate(s)
}
