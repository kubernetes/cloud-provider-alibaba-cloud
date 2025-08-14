// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDescribeHdMonitorRegionConfigResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetLogProject(v string) *DescribeHdMonitorRegionConfigResponseBody
	GetLogProject() *string
	SetMetricStore(v string) *DescribeHdMonitorRegionConfigResponseBody
	GetMetricStore() *string
	SetRegionId(v string) *DescribeHdMonitorRegionConfigResponseBody
	GetRegionId() *string
	SetRequestId(v string) *DescribeHdMonitorRegionConfigResponseBody
	GetRequestId() *string
}

type DescribeHdMonitorRegionConfigResponseBody struct {
	// The name of the Log Service project.
	//
	// example:
	//
	// hdmonitor-cn-hangzhou-223794579283657556
	LogProject *string `json:"LogProject,omitempty" xml:"LogProject,omitempty"`
	// The name of the Metricstore in Simple Log Service.
	//
	// example:
	//
	// hdmonitor-cn-hangzhou-metricStore-1
	MetricStore *string `json:"MetricStore,omitempty" xml:"MetricStore,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to obtain the region ID.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// 54B48E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DescribeHdMonitorRegionConfigResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DescribeHdMonitorRegionConfigResponseBody) GoString() string {
	return s.String()
}

func (s *DescribeHdMonitorRegionConfigResponseBody) GetLogProject() *string {
	return s.LogProject
}

func (s *DescribeHdMonitorRegionConfigResponseBody) GetMetricStore() *string {
	return s.MetricStore
}

func (s *DescribeHdMonitorRegionConfigResponseBody) GetRegionId() *string {
	return s.RegionId
}

func (s *DescribeHdMonitorRegionConfigResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DescribeHdMonitorRegionConfigResponseBody) SetLogProject(v string) *DescribeHdMonitorRegionConfigResponseBody {
	s.LogProject = &v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponseBody) SetMetricStore(v string) *DescribeHdMonitorRegionConfigResponseBody {
	s.MetricStore = &v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponseBody) SetRegionId(v string) *DescribeHdMonitorRegionConfigResponseBody {
	s.RegionId = &v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponseBody) SetRequestId(v string) *DescribeHdMonitorRegionConfigResponseBody {
	s.RequestId = &v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponseBody) Validate() error {
	return dara.Validate(s)
}
