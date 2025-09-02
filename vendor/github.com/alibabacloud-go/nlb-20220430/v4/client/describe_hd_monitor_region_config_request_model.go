// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDescribeHdMonitorRegionConfigRequest interface {
	dara.Model
	String() string
	GoString() string
	SetRegionId(v string) *DescribeHdMonitorRegionConfigRequest
	GetRegionId() *string
}

type DescribeHdMonitorRegionConfigRequest struct {
	// The ID of the region where the resources are deployed.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
}

func (s DescribeHdMonitorRegionConfigRequest) String() string {
	return dara.Prettify(s)
}

func (s DescribeHdMonitorRegionConfigRequest) GoString() string {
	return s.String()
}

func (s *DescribeHdMonitorRegionConfigRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *DescribeHdMonitorRegionConfigRequest) SetRegionId(v string) *DescribeHdMonitorRegionConfigRequest {
	s.RegionId = &v
	return s
}

func (s *DescribeHdMonitorRegionConfigRequest) Validate() error {
	return dara.Validate(s)
}
