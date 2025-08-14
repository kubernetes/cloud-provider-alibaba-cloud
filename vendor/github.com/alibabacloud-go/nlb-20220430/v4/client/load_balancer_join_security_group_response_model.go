// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iLoadBalancerJoinSecurityGroupResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *LoadBalancerJoinSecurityGroupResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *LoadBalancerJoinSecurityGroupResponse
	GetStatusCode() *int32
	SetBody(v *LoadBalancerJoinSecurityGroupResponseBody) *LoadBalancerJoinSecurityGroupResponse
	GetBody() *LoadBalancerJoinSecurityGroupResponseBody
}

type LoadBalancerJoinSecurityGroupResponse struct {
	Headers    map[string]*string                         `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                     `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *LoadBalancerJoinSecurityGroupResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s LoadBalancerJoinSecurityGroupResponse) String() string {
	return dara.Prettify(s)
}

func (s LoadBalancerJoinSecurityGroupResponse) GoString() string {
	return s.String()
}

func (s *LoadBalancerJoinSecurityGroupResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *LoadBalancerJoinSecurityGroupResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *LoadBalancerJoinSecurityGroupResponse) GetBody() *LoadBalancerJoinSecurityGroupResponseBody {
	return s.Body
}

func (s *LoadBalancerJoinSecurityGroupResponse) SetHeaders(v map[string]*string) *LoadBalancerJoinSecurityGroupResponse {
	s.Headers = v
	return s
}

func (s *LoadBalancerJoinSecurityGroupResponse) SetStatusCode(v int32) *LoadBalancerJoinSecurityGroupResponse {
	s.StatusCode = &v
	return s
}

func (s *LoadBalancerJoinSecurityGroupResponse) SetBody(v *LoadBalancerJoinSecurityGroupResponseBody) *LoadBalancerJoinSecurityGroupResponse {
	s.Body = v
	return s
}

func (s *LoadBalancerJoinSecurityGroupResponse) Validate() error {
	return dara.Validate(s)
}
