// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerAttributeResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *UpdateLoadBalancerAttributeResponseBody
	GetJobId() *string
	SetRequestId(v string) *UpdateLoadBalancerAttributeResponseBody
	GetRequestId() *string
}

type UpdateLoadBalancerAttributeResponseBody struct {
	// The ID of the asynchronous task.
	//
	// example:
	//
	// aab74cfa-3bc4-48fc-80fc-0101da5a****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The request ID.
	//
	// example:
	//
	// 7294679F-08DE-16D4-8E5D-1625685DC10B
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s UpdateLoadBalancerAttributeResponseBody) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAttributeResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAttributeResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *UpdateLoadBalancerAttributeResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *UpdateLoadBalancerAttributeResponseBody) SetJobId(v string) *UpdateLoadBalancerAttributeResponseBody {
	s.JobId = &v
	return s
}

func (s *UpdateLoadBalancerAttributeResponseBody) SetRequestId(v string) *UpdateLoadBalancerAttributeResponseBody {
	s.RequestId = &v
	return s
}

func (s *UpdateLoadBalancerAttributeResponseBody) Validate() error {
	return dara.Validate(s)
}
