// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteLoadBalancerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DeleteLoadBalancerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DeleteLoadBalancerResponse
	GetStatusCode() *int32
	SetBody(v *DeleteLoadBalancerResponseBody) *DeleteLoadBalancerResponse
	GetBody() *DeleteLoadBalancerResponseBody
}

type DeleteLoadBalancerResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DeleteLoadBalancerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DeleteLoadBalancerResponse) String() string {
	return dara.Prettify(s)
}

func (s DeleteLoadBalancerResponse) GoString() string {
	return s.String()
}

func (s *DeleteLoadBalancerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DeleteLoadBalancerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DeleteLoadBalancerResponse) GetBody() *DeleteLoadBalancerResponseBody {
	return s.Body
}

func (s *DeleteLoadBalancerResponse) SetHeaders(v map[string]*string) *DeleteLoadBalancerResponse {
	s.Headers = v
	return s
}

func (s *DeleteLoadBalancerResponse) SetStatusCode(v int32) *DeleteLoadBalancerResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteLoadBalancerResponse) SetBody(v *DeleteLoadBalancerResponseBody) *DeleteLoadBalancerResponse {
	s.Body = v
	return s
}

func (s *DeleteLoadBalancerResponse) Validate() error {
	return dara.Validate(s)
}
