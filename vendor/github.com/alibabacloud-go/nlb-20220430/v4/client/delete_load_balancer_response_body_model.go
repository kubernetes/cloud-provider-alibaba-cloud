// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteLoadBalancerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *DeleteLoadBalancerResponseBody
	GetJobId() *string
	SetRequestId(v string) *DeleteLoadBalancerResponseBody
	GetRequestId() *string
}

type DeleteLoadBalancerResponseBody struct {
	// The ID of the asynchronous task.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// 365F4154-92F6-4AE4-92F8-7FF34B540710
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteLoadBalancerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DeleteLoadBalancerResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteLoadBalancerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *DeleteLoadBalancerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DeleteLoadBalancerResponseBody) SetJobId(v string) *DeleteLoadBalancerResponseBody {
	s.JobId = &v
	return s
}

func (s *DeleteLoadBalancerResponseBody) SetRequestId(v string) *DeleteLoadBalancerResponseBody {
	s.RequestId = &v
	return s
}

func (s *DeleteLoadBalancerResponseBody) Validate() error {
	return dara.Validate(s)
}
