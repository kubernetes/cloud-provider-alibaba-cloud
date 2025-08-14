// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDetachCommonBandwidthPackageFromLoadBalancerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DetachCommonBandwidthPackageFromLoadBalancerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DetachCommonBandwidthPackageFromLoadBalancerResponse
	GetStatusCode() *int32
	SetBody(v *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) *DetachCommonBandwidthPackageFromLoadBalancerResponse
	GetBody() *DetachCommonBandwidthPackageFromLoadBalancerResponseBody
}

type DetachCommonBandwidthPackageFromLoadBalancerResponse struct {
	Headers    map[string]*string                                        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DetachCommonBandwidthPackageFromLoadBalancerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DetachCommonBandwidthPackageFromLoadBalancerResponse) String() string {
	return dara.Prettify(s)
}

func (s DetachCommonBandwidthPackageFromLoadBalancerResponse) GoString() string {
	return s.String()
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) GetBody() *DetachCommonBandwidthPackageFromLoadBalancerResponseBody {
	return s.Body
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) SetHeaders(v map[string]*string) *DetachCommonBandwidthPackageFromLoadBalancerResponse {
	s.Headers = v
	return s
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) SetStatusCode(v int32) *DetachCommonBandwidthPackageFromLoadBalancerResponse {
	s.StatusCode = &v
	return s
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) SetBody(v *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) *DetachCommonBandwidthPackageFromLoadBalancerResponse {
	s.Body = v
	return s
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponse) Validate() error {
	return dara.Validate(s)
}
