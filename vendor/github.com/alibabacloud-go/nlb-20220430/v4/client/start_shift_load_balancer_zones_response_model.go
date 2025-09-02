// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStartShiftLoadBalancerZonesResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *StartShiftLoadBalancerZonesResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *StartShiftLoadBalancerZonesResponse
	GetStatusCode() *int32
	SetBody(v *StartShiftLoadBalancerZonesResponseBody) *StartShiftLoadBalancerZonesResponse
	GetBody() *StartShiftLoadBalancerZonesResponseBody
}

type StartShiftLoadBalancerZonesResponse struct {
	Headers    map[string]*string                       `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                   `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *StartShiftLoadBalancerZonesResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s StartShiftLoadBalancerZonesResponse) String() string {
	return dara.Prettify(s)
}

func (s StartShiftLoadBalancerZonesResponse) GoString() string {
	return s.String()
}

func (s *StartShiftLoadBalancerZonesResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *StartShiftLoadBalancerZonesResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *StartShiftLoadBalancerZonesResponse) GetBody() *StartShiftLoadBalancerZonesResponseBody {
	return s.Body
}

func (s *StartShiftLoadBalancerZonesResponse) SetHeaders(v map[string]*string) *StartShiftLoadBalancerZonesResponse {
	s.Headers = v
	return s
}

func (s *StartShiftLoadBalancerZonesResponse) SetStatusCode(v int32) *StartShiftLoadBalancerZonesResponse {
	s.StatusCode = &v
	return s
}

func (s *StartShiftLoadBalancerZonesResponse) SetBody(v *StartShiftLoadBalancerZonesResponseBody) *StartShiftLoadBalancerZonesResponse {
	s.Body = v
	return s
}

func (s *StartShiftLoadBalancerZonesResponse) Validate() error {
	return dara.Validate(s)
}
