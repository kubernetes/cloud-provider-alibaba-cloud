// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateLoadBalancerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *CreateLoadBalancerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *CreateLoadBalancerResponse
	GetStatusCode() *int32
	SetBody(v *CreateLoadBalancerResponseBody) *CreateLoadBalancerResponse
	GetBody() *CreateLoadBalancerResponseBody
}

type CreateLoadBalancerResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *CreateLoadBalancerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s CreateLoadBalancerResponse) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerResponse) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *CreateLoadBalancerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *CreateLoadBalancerResponse) GetBody() *CreateLoadBalancerResponseBody {
	return s.Body
}

func (s *CreateLoadBalancerResponse) SetHeaders(v map[string]*string) *CreateLoadBalancerResponse {
	s.Headers = v
	return s
}

func (s *CreateLoadBalancerResponse) SetStatusCode(v int32) *CreateLoadBalancerResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateLoadBalancerResponse) SetBody(v *CreateLoadBalancerResponseBody) *CreateLoadBalancerResponse {
	s.Body = v
	return s
}

func (s *CreateLoadBalancerResponse) Validate() error {
	return dara.Validate(s)
}
