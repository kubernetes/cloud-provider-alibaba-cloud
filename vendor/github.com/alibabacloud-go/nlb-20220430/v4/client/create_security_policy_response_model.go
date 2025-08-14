// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateSecurityPolicyResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *CreateSecurityPolicyResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *CreateSecurityPolicyResponse
	GetStatusCode() *int32
	SetBody(v *CreateSecurityPolicyResponseBody) *CreateSecurityPolicyResponse
	GetBody() *CreateSecurityPolicyResponseBody
}

type CreateSecurityPolicyResponse struct {
	Headers    map[string]*string                `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                            `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *CreateSecurityPolicyResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s CreateSecurityPolicyResponse) String() string {
	return dara.Prettify(s)
}

func (s CreateSecurityPolicyResponse) GoString() string {
	return s.String()
}

func (s *CreateSecurityPolicyResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *CreateSecurityPolicyResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *CreateSecurityPolicyResponse) GetBody() *CreateSecurityPolicyResponseBody {
	return s.Body
}

func (s *CreateSecurityPolicyResponse) SetHeaders(v map[string]*string) *CreateSecurityPolicyResponse {
	s.Headers = v
	return s
}

func (s *CreateSecurityPolicyResponse) SetStatusCode(v int32) *CreateSecurityPolicyResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateSecurityPolicyResponse) SetBody(v *CreateSecurityPolicyResponseBody) *CreateSecurityPolicyResponse {
	s.Body = v
	return s
}

func (s *CreateSecurityPolicyResponse) Validate() error {
	return dara.Validate(s)
}
