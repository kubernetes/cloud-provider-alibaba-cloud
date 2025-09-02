// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStartShiftLoadBalancerZonesResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *StartShiftLoadBalancerZonesResponseBody
	GetRequestId() *string
}

type StartShiftLoadBalancerZonesResponseBody struct {
	// The request ID.
	//
	// example:
	//
	// 54B48E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s StartShiftLoadBalancerZonesResponseBody) String() string {
	return dara.Prettify(s)
}

func (s StartShiftLoadBalancerZonesResponseBody) GoString() string {
	return s.String()
}

func (s *StartShiftLoadBalancerZonesResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *StartShiftLoadBalancerZonesResponseBody) SetRequestId(v string) *StartShiftLoadBalancerZonesResponseBody {
	s.RequestId = &v
	return s
}

func (s *StartShiftLoadBalancerZonesResponseBody) Validate() error {
	return dara.Validate(s)
}
