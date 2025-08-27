// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iLoadBalancerLeaveSecurityGroupResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *LoadBalancerLeaveSecurityGroupResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *LoadBalancerLeaveSecurityGroupResponse
	GetStatusCode() *int32
	SetBody(v *LoadBalancerLeaveSecurityGroupResponseBody) *LoadBalancerLeaveSecurityGroupResponse
	GetBody() *LoadBalancerLeaveSecurityGroupResponseBody
}

type LoadBalancerLeaveSecurityGroupResponse struct {
	Headers    map[string]*string                          `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                      `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *LoadBalancerLeaveSecurityGroupResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s LoadBalancerLeaveSecurityGroupResponse) String() string {
	return dara.Prettify(s)
}

func (s LoadBalancerLeaveSecurityGroupResponse) GoString() string {
	return s.String()
}

func (s *LoadBalancerLeaveSecurityGroupResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *LoadBalancerLeaveSecurityGroupResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *LoadBalancerLeaveSecurityGroupResponse) GetBody() *LoadBalancerLeaveSecurityGroupResponseBody {
	return s.Body
}

func (s *LoadBalancerLeaveSecurityGroupResponse) SetHeaders(v map[string]*string) *LoadBalancerLeaveSecurityGroupResponse {
	s.Headers = v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupResponse) SetStatusCode(v int32) *LoadBalancerLeaveSecurityGroupResponse {
	s.StatusCode = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupResponse) SetBody(v *LoadBalancerLeaveSecurityGroupResponseBody) *LoadBalancerLeaveSecurityGroupResponse {
	s.Body = v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupResponse) Validate() error {
	return dara.Validate(s)
}
