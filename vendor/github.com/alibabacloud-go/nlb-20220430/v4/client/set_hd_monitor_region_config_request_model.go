// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iSetHdMonitorRegionConfigRequest interface {
	dara.Model
	String() string
	GoString() string
	SetLogProject(v string) *SetHdMonitorRegionConfigRequest
	GetLogProject() *string
	SetMetricStore(v string) *SetHdMonitorRegionConfigRequest
	GetMetricStore() *string
	SetRegionId(v string) *SetHdMonitorRegionConfigRequest
	GetRegionId() *string
}

type SetHdMonitorRegionConfigRequest struct {
	// The name of the Log Service project.
	//
	// This parameter is required.
	//
	// example:
	//
	// hdmonitor-cn-hangzhou-223794579283657556
	LogProject *string `json:"LogProject,omitempty" xml:"LogProject,omitempty"`
	// The name of the MetricStore in Simple Log Service.
	//
	// This parameter is required.
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
}

func (s SetHdMonitorRegionConfigRequest) String() string {
	return dara.Prettify(s)
}

func (s SetHdMonitorRegionConfigRequest) GoString() string {
	return s.String()
}

func (s *SetHdMonitorRegionConfigRequest) GetLogProject() *string {
	return s.LogProject
}

func (s *SetHdMonitorRegionConfigRequest) GetMetricStore() *string {
	return s.MetricStore
}

func (s *SetHdMonitorRegionConfigRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *SetHdMonitorRegionConfigRequest) SetLogProject(v string) *SetHdMonitorRegionConfigRequest {
	s.LogProject = &v
	return s
}

func (s *SetHdMonitorRegionConfigRequest) SetMetricStore(v string) *SetHdMonitorRegionConfigRequest {
	s.MetricStore = &v
	return s
}

func (s *SetHdMonitorRegionConfigRequest) SetRegionId(v string) *SetHdMonitorRegionConfigRequest {
	s.RegionId = &v
	return s
}

func (s *SetHdMonitorRegionConfigRequest) Validate() error {
	return dara.Validate(s)
}
