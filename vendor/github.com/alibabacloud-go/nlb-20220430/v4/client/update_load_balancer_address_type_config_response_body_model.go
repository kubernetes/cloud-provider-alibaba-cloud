// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateLoadBalancerAddressTypeConfigResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *UpdateLoadBalancerAddressTypeConfigResponseBody
	GetJobId() *string
	SetRequestId(v string) *UpdateLoadBalancerAddressTypeConfigResponseBody
	GetRequestId() *string
}

type UpdateLoadBalancerAddressTypeConfigResponseBody struct {
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

func (s UpdateLoadBalancerAddressTypeConfigResponseBody) String() string {
	return dara.Prettify(s)
}

func (s UpdateLoadBalancerAddressTypeConfigResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateLoadBalancerAddressTypeConfigResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *UpdateLoadBalancerAddressTypeConfigResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *UpdateLoadBalancerAddressTypeConfigResponseBody) SetJobId(v string) *UpdateLoadBalancerAddressTypeConfigResponseBody {
	s.JobId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigResponseBody) SetRequestId(v string) *UpdateLoadBalancerAddressTypeConfigResponseBody {
	s.RequestId = &v
	return s
}

func (s *UpdateLoadBalancerAddressTypeConfigResponseBody) Validate() error {
	return dara.Validate(s)
}
