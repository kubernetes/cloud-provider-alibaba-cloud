// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCancelShiftLoadBalancerZonesResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *CancelShiftLoadBalancerZonesResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *CancelShiftLoadBalancerZonesResponse
	GetStatusCode() *int32
	SetBody(v *CancelShiftLoadBalancerZonesResponseBody) *CancelShiftLoadBalancerZonesResponse
	GetBody() *CancelShiftLoadBalancerZonesResponseBody
}

type CancelShiftLoadBalancerZonesResponse struct {
	Headers    map[string]*string                        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *CancelShiftLoadBalancerZonesResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s CancelShiftLoadBalancerZonesResponse) String() string {
	return dara.Prettify(s)
}

func (s CancelShiftLoadBalancerZonesResponse) GoString() string {
	return s.String()
}

func (s *CancelShiftLoadBalancerZonesResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *CancelShiftLoadBalancerZonesResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *CancelShiftLoadBalancerZonesResponse) GetBody() *CancelShiftLoadBalancerZonesResponseBody {
	return s.Body
}

func (s *CancelShiftLoadBalancerZonesResponse) SetHeaders(v map[string]*string) *CancelShiftLoadBalancerZonesResponse {
	s.Headers = v
	return s
}

func (s *CancelShiftLoadBalancerZonesResponse) SetStatusCode(v int32) *CancelShiftLoadBalancerZonesResponse {
	s.StatusCode = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesResponse) SetBody(v *CancelShiftLoadBalancerZonesResponseBody) *CancelShiftLoadBalancerZonesResponse {
	s.Body = v
	return s
}

func (s *CancelShiftLoadBalancerZonesResponse) Validate() error {
	return dara.Validate(s)
}
