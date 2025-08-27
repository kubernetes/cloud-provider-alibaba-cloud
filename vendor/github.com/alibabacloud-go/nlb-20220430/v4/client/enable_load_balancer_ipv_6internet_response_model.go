// This file is auto-generated, don't edit it. Thanks.
package client

import (
  "github.com/alibabacloud-go/tea/dara"
)

type iEnableLoadBalancerIpv6InternetResponse interface {
  dara.Model
  String() string
  GoString() string
  SetHeaders(v map[string]*string) *EnableLoadBalancerIpv6InternetResponse
  GetHeaders() map[string]*string 
  SetStatusCode(v int32) *EnableLoadBalancerIpv6InternetResponse
  GetStatusCode() *int32 
  SetBody(v *EnableLoadBalancerIpv6InternetResponseBody) *EnableLoadBalancerIpv6InternetResponse
  GetBody() *EnableLoadBalancerIpv6InternetResponseBody 
}

type EnableLoadBalancerIpv6InternetResponse struct {
  Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
  StatusCode *int32 `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
  Body *EnableLoadBalancerIpv6InternetResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s EnableLoadBalancerIpv6InternetResponse) String() string {
  return dara.Prettify(s)
}

func (s EnableLoadBalancerIpv6InternetResponse) GoString() string {
  return s.String()
}

func (s *EnableLoadBalancerIpv6InternetResponse) GetHeaders() map[string]*string  {
  return s.Headers
}

func (s *EnableLoadBalancerIpv6InternetResponse) GetStatusCode() *int32  {
  return s.StatusCode
}

func (s *EnableLoadBalancerIpv6InternetResponse) GetBody() *EnableLoadBalancerIpv6InternetResponseBody  {
  return s.Body
}

func (s *EnableLoadBalancerIpv6InternetResponse) SetHeaders(v map[string]*string) *EnableLoadBalancerIpv6InternetResponse {
  s.Headers = v
  return s
}

func (s *EnableLoadBalancerIpv6InternetResponse) SetStatusCode(v int32) *EnableLoadBalancerIpv6InternetResponse {
  s.StatusCode = &v
  return s
}

func (s *EnableLoadBalancerIpv6InternetResponse) SetBody(v *EnableLoadBalancerIpv6InternetResponseBody) *EnableLoadBalancerIpv6InternetResponse {
  s.Body = v
  return s
}

func (s *EnableLoadBalancerIpv6InternetResponse) Validate() error {
  return dara.Validate(s)
}

