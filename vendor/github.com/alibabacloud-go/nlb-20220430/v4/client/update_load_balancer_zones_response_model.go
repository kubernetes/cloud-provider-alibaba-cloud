// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerZonesResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateLoadBalancerZonesResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateLoadBalancerZonesResponse
	GetStatusCode() *int32
	SetBody(v *UpdateLoadBalancerZonesResponseBody) *UpdateLoadBalancerZonesResponse
	GetBody() *UpdateLoadBalancerZonesResponseBody
}

type UpdateLoadBalancerZonesResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateLoadBalancerZonesResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateLoadBalancerZonesResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerZonesResponse) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerZonesResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateLoadBalancerZonesResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateLoadBalancerZonesResponse) GetBody() *UpdateLoadBalancerZonesResponseBody {
	return s.Body
}

func (s *UpdateLoadBalancerZonesResponse) SetHeaders(v map[string]*string) *UpdateLoadBalancerZonesResponse {
	s.Headers = v
	return s
}

func (s *UpdateLoadBalancerZonesResponse) SetStatusCode(v int32) *UpdateLoadBalancerZonesResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateLoadBalancerZonesResponse) SetBody(v *UpdateLoadBalancerZonesResponseBody) *UpdateLoadBalancerZonesResponse {
	s.Body = v
	return s
}

func (s *UpdateLoadBalancerZonesResponse) Validate() error {
	return dara.Validate(s)
}
