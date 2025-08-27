// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAttachCommonBandwidthPackageToLoadBalancerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *AttachCommonBandwidthPackageToLoadBalancerResponseBody
	GetJobId() *string
	SetRequestId(v string) *AttachCommonBandwidthPackageToLoadBalancerResponseBody
	GetRequestId() *string
}

type AttachCommonBandwidthPackageToLoadBalancerResponseBody struct {
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

func (s AttachCommonBandwidthPackageToLoadBalancerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s AttachCommonBandwidthPackageToLoadBalancerResponseBody) GoString() string {
	return s.String()
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponseBody) SetJobId(v string) *AttachCommonBandwidthPackageToLoadBalancerResponseBody {
	s.JobId = &v
	return s
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponseBody) SetRequestId(v string) *AttachCommonBandwidthPackageToLoadBalancerResponseBody {
	s.RequestId = &v
	return s
}

func (s *AttachCommonBandwidthPackageToLoadBalancerResponseBody) Validate() error {
	return dara.Validate(s)
}
