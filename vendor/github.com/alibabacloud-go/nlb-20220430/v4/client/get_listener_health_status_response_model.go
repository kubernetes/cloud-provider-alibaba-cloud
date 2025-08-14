// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetListenerHealthStatusResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *GetListenerHealthStatusResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *GetListenerHealthStatusResponse
	GetStatusCode() *int32
	SetBody(v *GetListenerHealthStatusResponseBody) *GetListenerHealthStatusResponse
	GetBody() *GetListenerHealthStatusResponseBody
}

type GetListenerHealthStatusResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *GetListenerHealthStatusResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s GetListenerHealthStatusResponse) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusResponse) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *GetListenerHealthStatusResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *GetListenerHealthStatusResponse) GetBody() *GetListenerHealthStatusResponseBody {
	return s.Body
}

func (s *GetListenerHealthStatusResponse) SetHeaders(v map[string]*string) *GetListenerHealthStatusResponse {
	s.Headers = v
	return s
}

func (s *GetListenerHealthStatusResponse) SetStatusCode(v int32) *GetListenerHealthStatusResponse {
	s.StatusCode = &v
	return s
}

func (s *GetListenerHealthStatusResponse) SetBody(v *GetListenerHealthStatusResponseBody) *GetListenerHealthStatusResponse {
	s.Body = v
	return s
}

func (s *GetListenerHealthStatusResponse) Validate() error {
	return dara.Validate(s)
}
