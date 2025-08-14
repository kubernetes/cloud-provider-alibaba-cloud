// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerProtectionResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateLoadBalancerProtectionResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateLoadBalancerProtectionResponse
	GetStatusCode() *int32
	SetBody(v *UpdateLoadBalancerProtectionResponseBody) *UpdateLoadBalancerProtectionResponse
	GetBody() *UpdateLoadBalancerProtectionResponseBody
}

type UpdateLoadBalancerProtectionResponse struct {
	Headers    map[string]*string                        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateLoadBalancerProtectionResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateLoadBalancerProtectionResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerProtectionResponse) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerProtectionResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateLoadBalancerProtectionResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateLoadBalancerProtectionResponse) GetBody() *UpdateLoadBalancerProtectionResponseBody {
	return s.Body
}

func (s *UpdateLoadBalancerProtectionResponse) SetHeaders(v map[string]*string) *UpdateLoadBalancerProtectionResponse {
	s.Headers = v
	return s
}

func (s *UpdateLoadBalancerProtectionResponse) SetStatusCode(v int32) *UpdateLoadBalancerProtectionResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateLoadBalancerProtectionResponse) SetBody(v *UpdateLoadBalancerProtectionResponseBody) *UpdateLoadBalancerProtectionResponse {
	s.Body = v
	return s
}

func (s *UpdateLoadBalancerProtectionResponse) Validate() error {
	return dara.Validate(s)
}
