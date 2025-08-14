// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iLoadBalancerJoinSecurityGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *LoadBalancerJoinSecurityGroupResponseBody
	GetJobId() *string
	SetRequestId(v string) *LoadBalancerJoinSecurityGroupResponseBody
	GetRequestId() *string
}

type LoadBalancerJoinSecurityGroupResponseBody struct {
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

func (s LoadBalancerJoinSecurityGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s LoadBalancerJoinSecurityGroupResponseBody) GoString() string {
	return s.String()
}

func (s *LoadBalancerJoinSecurityGroupResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *LoadBalancerJoinSecurityGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *LoadBalancerJoinSecurityGroupResponseBody) SetJobId(v string) *LoadBalancerJoinSecurityGroupResponseBody {
	s.JobId = &v
	return s
}

func (s *LoadBalancerJoinSecurityGroupResponseBody) SetRequestId(v string) *LoadBalancerJoinSecurityGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *LoadBalancerJoinSecurityGroupResponseBody) Validate() error {
	return dara.Validate(s)
}
