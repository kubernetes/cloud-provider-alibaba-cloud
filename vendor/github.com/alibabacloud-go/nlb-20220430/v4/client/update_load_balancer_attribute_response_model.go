// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateLoadBalancerAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateLoadBalancerAttributeResponse
	GetStatusCode() *int32
	SetBody(v *UpdateLoadBalancerAttributeResponseBody) *UpdateLoadBalancerAttributeResponse
	GetBody() *UpdateLoadBalancerAttributeResponseBody
}

type UpdateLoadBalancerAttributeResponse struct {
	Headers    map[string]*string                       `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                   `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateLoadBalancerAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateLoadBalancerAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAttributeResponse) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateLoadBalancerAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateLoadBalancerAttributeResponse) GetBody() *UpdateLoadBalancerAttributeResponseBody {
	return s.Body
}

func (s *UpdateLoadBalancerAttributeResponse) SetHeaders(v map[string]*string) *UpdateLoadBalancerAttributeResponse {
	s.Headers = v
	return s
}

func (s *UpdateLoadBalancerAttributeResponse) SetStatusCode(v int32) *UpdateLoadBalancerAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateLoadBalancerAttributeResponse) SetBody(v *UpdateLoadBalancerAttributeResponseBody) *UpdateLoadBalancerAttributeResponse {
	s.Body = v
	return s
}

func (s *UpdateLoadBalancerAttributeResponse) Validate() error {
	return dara.Validate(s)
}
