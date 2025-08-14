// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateLoadBalancerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetLoadbalancerId(v string) *CreateLoadBalancerResponseBody
	GetLoadbalancerId() *string
	SetOrderId(v int64) *CreateLoadBalancerResponseBody
	GetOrderId() *int64
	SetRequestId(v string) *CreateLoadBalancerResponseBody
	GetRequestId() *string
}

type CreateLoadBalancerResponseBody struct {
	// The ID of the NLB instance.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadbalancerId *string `json:"LoadbalancerId,omitempty" xml:"LoadbalancerId,omitempty"`
	// The ID of the order for the NLB instance.
	//
	// example:
	//
	// 20230000
	OrderId *int64 `json:"OrderId,omitempty" xml:"OrderId,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateLoadBalancerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s CreateLoadBalancerResponseBody) GoString() string {
	return s.String()
}

func (s *CreateLoadBalancerResponseBody) GetLoadbalancerId() *string {
	return s.LoadbalancerId
}

func (s *CreateLoadBalancerResponseBody) GetOrderId() *int64 {
	return s.OrderId
}

func (s *CreateLoadBalancerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *CreateLoadBalancerResponseBody) SetLoadbalancerId(v string) *CreateLoadBalancerResponseBody {
	s.LoadbalancerId = &v
	return s
}

func (s *CreateLoadBalancerResponseBody) SetOrderId(v int64) *CreateLoadBalancerResponseBody {
	s.OrderId = &v
	return s
}

func (s *CreateLoadBalancerResponseBody) SetRequestId(v string) *CreateLoadBalancerResponseBody {
	s.RequestId = &v
	return s
}

func (s *CreateLoadBalancerResponseBody) Validate() error {
	return dara.Validate(s)
}
