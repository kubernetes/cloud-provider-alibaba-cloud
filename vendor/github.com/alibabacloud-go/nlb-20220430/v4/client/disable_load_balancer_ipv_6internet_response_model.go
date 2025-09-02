// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDisableLoadBalancerIpv6InternetResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DisableLoadBalancerIpv6InternetResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DisableLoadBalancerIpv6InternetResponse
	GetStatusCode() *int32
	SetBody(v *DisableLoadBalancerIpv6InternetResponseBody) *DisableLoadBalancerIpv6InternetResponse
	GetBody() *DisableLoadBalancerIpv6InternetResponseBody
}

type DisableLoadBalancerIpv6InternetResponse struct {
	Headers    map[string]*string                           `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                       `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DisableLoadBalancerIpv6InternetResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DisableLoadBalancerIpv6InternetResponse) String() string {
	return dara.Prettify(s)
}

func (s DisableLoadBalancerIpv6InternetResponse) GoString() string {
	return s.String()
}

func (s *DisableLoadBalancerIpv6InternetResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DisableLoadBalancerIpv6InternetResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DisableLoadBalancerIpv6InternetResponse) GetBody() *DisableLoadBalancerIpv6InternetResponseBody {
	return s.Body
}

func (s *DisableLoadBalancerIpv6InternetResponse) SetHeaders(v map[string]*string) *DisableLoadBalancerIpv6InternetResponse {
	s.Headers = v
	return s
}

func (s *DisableLoadBalancerIpv6InternetResponse) SetStatusCode(v int32) *DisableLoadBalancerIpv6InternetResponse {
	s.StatusCode = &v
	return s
}

func (s *DisableLoadBalancerIpv6InternetResponse) SetBody(v *DisableLoadBalancerIpv6InternetResponseBody) *DisableLoadBalancerIpv6InternetResponse {
	s.Body = v
	return s
}

func (s *DisableLoadBalancerIpv6InternetResponse) Validate() error {
	return dara.Validate(s)
}
