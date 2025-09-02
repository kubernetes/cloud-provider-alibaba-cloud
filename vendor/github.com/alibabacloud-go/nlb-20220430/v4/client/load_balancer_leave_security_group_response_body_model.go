// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iLoadBalancerLeaveSecurityGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *LoadBalancerLeaveSecurityGroupResponseBody
	GetJobId() *string
	SetRequestId(v string) *LoadBalancerLeaveSecurityGroupResponseBody
	GetRequestId() *string
}

type LoadBalancerLeaveSecurityGroupResponseBody struct {
	// The ID of the asynchronous task.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The request ID.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s LoadBalancerLeaveSecurityGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s LoadBalancerLeaveSecurityGroupResponseBody) GoString() string {
	return s.String()
}

func (s *LoadBalancerLeaveSecurityGroupResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *LoadBalancerLeaveSecurityGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *LoadBalancerLeaveSecurityGroupResponseBody) SetJobId(v string) *LoadBalancerLeaveSecurityGroupResponseBody {
	s.JobId = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupResponseBody) SetRequestId(v string) *LoadBalancerLeaveSecurityGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *LoadBalancerLeaveSecurityGroupResponseBody) Validate() error {
	return dara.Validate(s)
}
