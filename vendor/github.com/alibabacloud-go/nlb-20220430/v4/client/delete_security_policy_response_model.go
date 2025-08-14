// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteSecurityPolicyResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DeleteSecurityPolicyResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DeleteSecurityPolicyResponse
	GetStatusCode() *int32
	SetBody(v *DeleteSecurityPolicyResponseBody) *DeleteSecurityPolicyResponse
	GetBody() *DeleteSecurityPolicyResponseBody
}

type DeleteSecurityPolicyResponse struct {
	Headers    map[string]*string                `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                            `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DeleteSecurityPolicyResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DeleteSecurityPolicyResponse) String() string {
	return dara.Prettify(s)
}

func (s DeleteSecurityPolicyResponse) GoString() string {
	return s.String()
}

func (s *DeleteSecurityPolicyResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DeleteSecurityPolicyResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DeleteSecurityPolicyResponse) GetBody() *DeleteSecurityPolicyResponseBody {
	return s.Body
}

func (s *DeleteSecurityPolicyResponse) SetHeaders(v map[string]*string) *DeleteSecurityPolicyResponse {
	s.Headers = v
	return s
}

func (s *DeleteSecurityPolicyResponse) SetStatusCode(v int32) *DeleteSecurityPolicyResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteSecurityPolicyResponse) SetBody(v *DeleteSecurityPolicyResponseBody) *DeleteSecurityPolicyResponse {
	s.Body = v
	return s
}

func (s *DeleteSecurityPolicyResponse) Validate() error {
	return dara.Validate(s)
}
