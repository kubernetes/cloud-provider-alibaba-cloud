// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetLoadBalancerAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *GetLoadBalancerAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *GetLoadBalancerAttributeResponse
	GetStatusCode() *int32
	SetBody(v *GetLoadBalancerAttributeResponseBody) *GetLoadBalancerAttributeResponse
	GetBody() *GetLoadBalancerAttributeResponseBody
}

type GetLoadBalancerAttributeResponse struct {
	Headers    map[string]*string                    `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *GetLoadBalancerAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s GetLoadBalancerAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s GetLoadBalancerAttributeResponse) GoString() string {
	return s.String()
}

func (s *GetLoadBalancerAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *GetLoadBalancerAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *GetLoadBalancerAttributeResponse) GetBody() *GetLoadBalancerAttributeResponseBody {
	return s.Body
}

func (s *GetLoadBalancerAttributeResponse) SetHeaders(v map[string]*string) *GetLoadBalancerAttributeResponse {
	s.Headers = v
	return s
}

func (s *GetLoadBalancerAttributeResponse) SetStatusCode(v int32) *GetLoadBalancerAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *GetLoadBalancerAttributeResponse) SetBody(v *GetLoadBalancerAttributeResponseBody) *GetLoadBalancerAttributeResponse {
	s.Body = v
	return s
}

func (s *GetLoadBalancerAttributeResponse) Validate() error {
	return dara.Validate(s)
}
