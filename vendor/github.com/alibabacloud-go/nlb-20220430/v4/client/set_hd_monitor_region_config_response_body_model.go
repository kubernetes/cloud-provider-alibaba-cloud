// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iSetHdMonitorRegionConfigResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetLogProject(v string) *SetHdMonitorRegionConfigResponseBody
	GetLogProject() *string
	SetMetricStore(v string) *SetHdMonitorRegionConfigResponseBody
	GetMetricStore() *string
	SetRegionId(v string) *SetHdMonitorRegionConfigResponseBody
	GetRegionId() *string
	SetRequestId(v string) *SetHdMonitorRegionConfigResponseBody
	GetRequestId() *string
}

type SetHdMonitorRegionConfigResponseBody struct {
	// The name of the Log Service project.
	//
	// example:
	//
	// hdmonitor-cn-hangzhou-223794579283657556
	LogProject *string `json:"LogProject,omitempty" xml:"LogProject,omitempty"`
	// The name of the MetricStore in Simple Log Service.
	//
	// example:
	//
	// hdmonitor-cn-hangzhou-metricStore-1
	MetricStore *string `json:"MetricStore,omitempty" xml:"MetricStore,omitempty"`
	// The region ID of the NLB instance.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/2399192.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The request ID.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s SetHdMonitorRegionConfigResponseBody) String() string {
	return dara.Prettify(s)
}

func (s SetHdMonitorRegionConfigResponseBody) GoString() string {
	return s.String()
}

func (s *SetHdMonitorRegionConfigResponseBody) GetLogProject() *string {
	return s.LogProject
}

func (s *SetHdMonitorRegionConfigResponseBody) GetMetricStore() *string {
	return s.MetricStore
}

func (s *SetHdMonitorRegionConfigResponseBody) GetRegionId() *string {
	return s.RegionId
}

func (s *SetHdMonitorRegionConfigResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *SetHdMonitorRegionConfigResponseBody) SetLogProject(v string) *SetHdMonitorRegionConfigResponseBody {
	s.LogProject = &v
	return s
}

func (s *SetHdMonitorRegionConfigResponseBody) SetMetricStore(v string) *SetHdMonitorRegionConfigResponseBody {
	s.MetricStore = &v
	return s
}

func (s *SetHdMonitorRegionConfigResponseBody) SetRegionId(v string) *SetHdMonitorRegionConfigResponseBody {
	s.RegionId = &v
	return s
}

func (s *SetHdMonitorRegionConfigResponseBody) SetRequestId(v string) *SetHdMonitorRegionConfigResponseBody {
	s.RequestId = &v
	return s
}

func (s *SetHdMonitorRegionConfigResponseBody) Validate() error {
	return dara.Validate(s)
}
