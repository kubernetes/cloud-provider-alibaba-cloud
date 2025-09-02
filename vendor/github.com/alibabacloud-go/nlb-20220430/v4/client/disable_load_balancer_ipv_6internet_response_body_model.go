// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDisableLoadBalancerIpv6InternetResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *DisableLoadBalancerIpv6InternetResponseBody
	GetRequestId() *string
}

type DisableLoadBalancerIpv6InternetResponseBody struct {
	// The request ID.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DisableLoadBalancerIpv6InternetResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DisableLoadBalancerIpv6InternetResponseBody) GoString() string {
	return s.String()
}

func (s *DisableLoadBalancerIpv6InternetResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DisableLoadBalancerIpv6InternetResponseBody) SetRequestId(v string) *DisableLoadBalancerIpv6InternetResponseBody {
	s.RequestId = &v
	return s
}

func (s *DisableLoadBalancerIpv6InternetResponseBody) Validate() error {
	return dara.Validate(s)
}
