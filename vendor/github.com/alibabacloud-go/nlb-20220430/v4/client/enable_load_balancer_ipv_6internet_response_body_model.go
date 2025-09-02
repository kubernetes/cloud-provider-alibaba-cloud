// This file is auto-generated, don't edit it. Thanks.
package client

import (
  "github.com/alibabacloud-go/tea/dara"
)

type iEnableLoadBalancerIpv6InternetResponseBody interface {
  dara.Model
  String() string
  GoString() string
  SetRequestId(v string) *EnableLoadBalancerIpv6InternetResponseBody
  GetRequestId() *string 
}

type EnableLoadBalancerIpv6InternetResponseBody struct {
  // The ID of the request.
  // 
  // example:
  // 
  // CEF72CEB-54B6-4AE8-B225-F876FF7BA984
  RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s EnableLoadBalancerIpv6InternetResponseBody) String() string {
  return dara.Prettify(s)
}

func (s EnableLoadBalancerIpv6InternetResponseBody) GoString() string {
  return s.String()
}

func (s *EnableLoadBalancerIpv6InternetResponseBody) GetRequestId() *string  {
  return s.RequestId
}

func (s *EnableLoadBalancerIpv6InternetResponseBody) SetRequestId(v string) *EnableLoadBalancerIpv6InternetResponseBody {
  s.RequestId = &v
  return s
}

func (s *EnableLoadBalancerIpv6InternetResponseBody) Validate() error {
  return dara.Validate(s)
}

