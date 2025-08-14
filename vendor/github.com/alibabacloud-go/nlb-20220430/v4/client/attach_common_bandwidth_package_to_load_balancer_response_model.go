// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAttachCommonBandwidthPackageToLoadBalancerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *AttachCommonBandwidthPackageToLoadBalancerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *AttachCommonBandwidthPackageToLoadBalancerResponse
	GetStatusCode() *int32
	SetBody(v *AttachCommonBandwidthPackageToLoadBalancerResponseBody) *AttachCommonBandwidthPackageToLoadBalancerResponse
	GetBody() *AttachCommonBandwidthPackageToLoadBalancerResponseBody
}

type AttachCommonBandwidthPackageToLoadBalancerResponse struct {
	Headers    map[string]*string                                      `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                                  `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *AttachCommonBandwidthPackageToLoadBalancerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s AttachCommonBandwidthPackageToLoadBalancerResponse) String() string {
	return dara.Prettify(s)
}

func (s AttachCommonBandwidthPackageToLoadBalancerResponse) GoString() string {
	return s.String()
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) GetBody() *AttachCommonBandwidthPackageToLoadBalancerResponseBody {
	return s.Body
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) SetHeaders(v map[string]*string) *AttachCommonBandwidthPackageToLoadBalancerResponse {
	s.Headers = v
	return s
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) SetStatusCode(v int32) *AttachCommonBandwidthPackageToLoadBalancerResponse {
	s.StatusCode = &v
	return s
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) SetBody(v *AttachCommonBandwidthPackageToLoadBalancerResponseBody) *AttachCommonBandwidthPackageToLoadBalancerResponse {
	s.Body = v
	return s
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponse) Validate() error {
	return dara.Validate(s)
}
