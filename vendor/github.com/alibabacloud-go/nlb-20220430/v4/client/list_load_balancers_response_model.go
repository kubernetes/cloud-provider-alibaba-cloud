// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListLoadBalancersResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListLoadBalancersResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListLoadBalancersResponse
	GetStatusCode() *int32
	SetBody(v *ListLoadBalancersResponseBody) *ListLoadBalancersResponse
	GetBody() *ListLoadBalancersResponseBody
}

type ListLoadBalancersResponse struct {
	Headers    map[string]*string             `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                         `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListLoadBalancersResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListLoadBalancersResponse) String() string {
	return dara.Prettify(s)
}

func (s ListLoadBalancersResponse) GoString() string {
	return s.String()
}

func (s *ListLoadBalancersResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListLoadBalancersResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListLoadBalancersResponse) GetBody() *ListLoadBalancersResponseBody {
	return s.Body
}

func (s *ListLoadBalancersResponse) SetHeaders(v map[string]*string) *ListLoadBalancersResponse {
	s.Headers = v
	return s
}

func (s *ListLoadBalancersResponse) SetStatusCode(v int32) *ListLoadBalancersResponse {
	s.StatusCode = &v
	return s
}

func (s *ListLoadBalancersResponse) SetBody(v *ListLoadBalancersResponseBody) *ListLoadBalancersResponse {
	s.Body = v
	return s
}

func (s *ListLoadBalancersResponse) Validate() error {
	return dara.Validate(s)
}
