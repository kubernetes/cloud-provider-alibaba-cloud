// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iSetHdMonitorRegionConfigResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *SetHdMonitorRegionConfigResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *SetHdMonitorRegionConfigResponse
	GetStatusCode() *int32
	SetBody(v *SetHdMonitorRegionConfigResponseBody) *SetHdMonitorRegionConfigResponse
	GetBody() *SetHdMonitorRegionConfigResponseBody
}

type SetHdMonitorRegionConfigResponse struct {
	Headers    map[string]*string                    `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *SetHdMonitorRegionConfigResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s SetHdMonitorRegionConfigResponse) String() string {
	return dara.Prettify(s)
}

func (s SetHdMonitorRegionConfigResponse) GoString() string {
	return s.String()
}

func (s *SetHdMonitorRegionConfigResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *SetHdMonitorRegionConfigResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *SetHdMonitorRegionConfigResponse) GetBody() *SetHdMonitorRegionConfigResponseBody {
	return s.Body
}

func (s *SetHdMonitorRegionConfigResponse) SetHeaders(v map[string]*string) *SetHdMonitorRegionConfigResponse {
	s.Headers = v
	return s
}

func (s *SetHdMonitorRegionConfigResponse) SetStatusCode(v int32) *SetHdMonitorRegionConfigResponse {
	s.StatusCode = &v
	return s
}

func (s *SetHdMonitorRegionConfigResponse) SetBody(v *SetHdMonitorRegionConfigResponseBody) *SetHdMonitorRegionConfigResponse {
	s.Body = v
	return s
}

func (s *SetHdMonitorRegionConfigResponse) Validate() error {
	return dara.Validate(s)
}
