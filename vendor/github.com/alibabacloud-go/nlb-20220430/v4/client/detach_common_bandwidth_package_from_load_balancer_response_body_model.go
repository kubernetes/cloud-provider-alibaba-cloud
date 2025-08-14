// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDetachCommonBandwidthPackageFromLoadBalancerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *DetachCommonBandwidthPackageFromLoadBalancerResponseBody
	GetJobId() *string
	SetRequestId(v string) *DetachCommonBandwidthPackageFromLoadBalancerResponseBody
	GetRequestId() *string
}

type DetachCommonBandwidthPackageFromLoadBalancerResponseBody struct {
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
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DetachCommonBandwidthPackageFromLoadBalancerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DetachCommonBandwidthPackageFromLoadBalancerResponseBody) GoString() string {
	return s.String()
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) SetJobId(v string) *DetachCommonBandwidthPackageFromLoadBalancerResponseBody {
	s.JobId = &v
	return s
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) SetRequestId(v string) *DetachCommonBandwidthPackageFromLoadBalancerResponseBody {
	s.RequestId = &v
	return s
}

func (s *DetachCommonBandwidthPackageFromLoadBalancerResponseBody) Validate() error {
	return dara.Validate(s)
}
