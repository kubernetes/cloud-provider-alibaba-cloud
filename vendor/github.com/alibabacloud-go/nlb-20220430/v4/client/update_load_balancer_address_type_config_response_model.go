// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerAddressTypeConfigResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateLoadBalancerAddressTypeConfigResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateLoadBalancerAddressTypeConfigResponse
	GetStatusCode() *int32
	SetBody(v *UpdateLoadBalancerAddressTypeConfigResponseBody) *UpdateLoadBalancerAddressTypeConfigResponse
	GetBody() *UpdateLoadBalancerAddressTypeConfigResponseBody
}

type UpdateLoadBalancerAddressTypeConfigResponse struct {
	Headers    map[string]*string                               `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                           `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateLoadBalancerAddressTypeConfigResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateLoadBalancerAddressTypeConfigResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAddressTypeConfigResponse) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) GetBody() *UpdateLoadBalancerAddressTypeConfigResponseBody {
	return s.Body
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) SetHeaders(v map[string]*string) *UpdateLoadBalancerAddressTypeConfigResponse {
	s.Headers = v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) SetStatusCode(v int32) *UpdateLoadBalancerAddressTypeConfigResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) SetBody(v *UpdateLoadBalancerAddressTypeConfigResponseBody) *UpdateLoadBalancerAddressTypeConfigResponse {
	s.Body = v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigResponse) Validate() error {
	return dara.Validate(s)
}
