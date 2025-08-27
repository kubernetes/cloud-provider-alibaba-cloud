// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCancelShiftLoadBalancerZonesResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *CancelShiftLoadBalancerZonesResponseBody
	GetRequestId() *string
}

type CancelShiftLoadBalancerZonesResponseBody struct {
	// The request ID.
	//
	// example:
	//
	// 54B48E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CancelShiftLoadBalancerZonesResponseBody) String() string {
	return dara.Prettify(s)
}

func (s CancelShiftLoadBalancerZonesResponseBody) GoString() string {
	return s.String()
}

func (s *CancelShiftLoadBalancerZonesResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *CancelShiftLoadBalancerZonesResponseBody) SetRequestId(v string) *CancelShiftLoadBalancerZonesResponseBody {
	s.RequestId = &v
	return s
}

func (s *CancelShiftLoadBalancerZonesResponseBody) Validate() error {
	return dara.Validate(s)
}
