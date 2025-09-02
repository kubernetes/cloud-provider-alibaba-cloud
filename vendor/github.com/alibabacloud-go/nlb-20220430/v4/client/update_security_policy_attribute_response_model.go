// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateSecurityPolicyAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateSecurityPolicyAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateSecurityPolicyAttributeResponse
	GetStatusCode() *int32
	SetBody(v *UpdateSecurityPolicyAttributeResponseBody) *UpdateSecurityPolicyAttributeResponse
	GetBody() *UpdateSecurityPolicyAttributeResponseBody
}

type UpdateSecurityPolicyAttributeResponse struct {
	Headers    map[string]*string                         `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                     `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateSecurityPolicyAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateSecurityPolicyAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateSecurityPolicyAttributeResponse) GoString() string {
	return s.String()
}

func (s *UpdateSecurityPolicyAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateSecurityPolicyAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateSecurityPolicyAttributeResponse) GetBody() *UpdateSecurityPolicyAttributeResponseBody {
	return s.Body
}

func (s *UpdateSecurityPolicyAttributeResponse) SetHeaders(v map[string]*string) *UpdateSecurityPolicyAttributeResponse {
	s.Headers = v
	return s
}

func (s *UpdateSecurityPolicyAttributeResponse) SetStatusCode(v int32) *UpdateSecurityPolicyAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeResponse) SetBody(v *UpdateSecurityPolicyAttributeResponseBody) *UpdateSecurityPolicyAttributeResponse {
	s.Body = v
	return s
}

func (s *UpdateSecurityPolicyAttributeResponse) Validate() error {
	return dara.Validate(s)
}
