// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStopListenerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *StopListenerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *StopListenerResponse
	GetStatusCode() *int32
	SetBody(v *StopListenerResponseBody) *StopListenerResponse
	GetBody() *StopListenerResponseBody
}

type StopListenerResponse struct {
	Headers    map[string]*string        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *StopListenerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s StopListenerResponse) String() string {
	return dara.Prettify(s)
}

func (s StopListenerResponse) GoString() string {
	return s.String()
}

func (s *StopListenerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *StopListenerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *StopListenerResponse) GetBody() *StopListenerResponseBody {
	return s.Body
}

func (s *StopListenerResponse) SetHeaders(v map[string]*string) *StopListenerResponse {
	s.Headers = v
	return s
}

func (s *StopListenerResponse) SetStatusCode(v int32) *StopListenerResponse {
	s.StatusCode = &v
	return s
}

func (s *StopListenerResponse) SetBody(v *StopListenerResponseBody) *StopListenerResponse {
	s.Body = v
	return s
}

func (s *StopListenerResponse) Validate() error {
	return dara.Validate(s)
}
