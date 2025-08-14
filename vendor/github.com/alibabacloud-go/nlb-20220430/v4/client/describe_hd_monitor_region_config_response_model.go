// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDescribeHdMonitorRegionConfigResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DescribeHdMonitorRegionConfigResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DescribeHdMonitorRegionConfigResponse
	GetStatusCode() *int32
	SetBody(v *DescribeHdMonitorRegionConfigResponseBody) *DescribeHdMonitorRegionConfigResponse
	GetBody() *DescribeHdMonitorRegionConfigResponseBody
}

type DescribeHdMonitorRegionConfigResponse struct {
	Headers    map[string]*string                         `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                     `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DescribeHdMonitorRegionConfigResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DescribeHdMonitorRegionConfigResponse) String() string {
	return dara.Prettify(s)
}

func (s DescribeHdMonitorRegionConfigResponse) GoString() string {
	return s.String()
}

func (s *DescribeHdMonitorRegionConfigResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DescribeHdMonitorRegionConfigResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DescribeHdMonitorRegionConfigResponse) GetBody() *DescribeHdMonitorRegionConfigResponseBody {
	return s.Body
}

func (s *DescribeHdMonitorRegionConfigResponse) SetHeaders(v map[string]*string) *DescribeHdMonitorRegionConfigResponse {
	s.Headers = v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponse) SetStatusCode(v int32) *DescribeHdMonitorRegionConfigResponse {
	s.StatusCode = &v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponse) SetBody(v *DescribeHdMonitorRegionConfigResponseBody) *DescribeHdMonitorRegionConfigResponse {
	s.Body = v
	return s
}

func (s *DescribeHdMonitorRegionConfigResponse) Validate() error {
	return dara.Validate(s)
}
