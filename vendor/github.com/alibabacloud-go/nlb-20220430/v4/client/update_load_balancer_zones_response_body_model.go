// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerZonesResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *UpdateLoadBalancerZonesResponseBody
	GetJobId() *string
	SetRequestId(v string) *UpdateLoadBalancerZonesResponseBody
	GetRequestId() *string
}

type UpdateLoadBalancerZonesResponseBody struct {
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

func (s UpdateLoadBalancerZonesResponseBody) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerZonesResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerZonesResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *UpdateLoadBalancerZonesResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *UpdateLoadBalancerZonesResponseBody) SetJobId(v string) *UpdateLoadBalancerZonesResponseBody {
	s.JobId = &v
	return s
}

func (s *UpdateLoadBalancerZonesResponseBody) SetRequestId(v string) *UpdateLoadBalancerZonesResponseBody {
	s.RequestId = &v
	return s
}

func (s *UpdateLoadBalancerZonesResponseBody) Validate() error {
	return dara.Validate(s)
}
