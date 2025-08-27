// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStartListenerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *StartListenerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *StartListenerResponse
	GetStatusCode() *int32
	SetBody(v *StartListenerResponseBody) *StartListenerResponse
	GetBody() *StartListenerResponseBody
}

type StartListenerResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *StartListenerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s StartListenerResponse) String() string {
	return dara.Prettify(s)
}

func (s StartListenerResponse) GoString() string {
	return s.String()
}

func (s *StartListenerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *StartListenerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *StartListenerResponse) GetBody() *StartListenerResponseBody {
	return s.Body
}

func (s *StartListenerResponse) SetHeaders(v map[string]*string) *StartListenerResponse {
	s.Headers = v
	return s
}

func (s *StartListenerResponse) SetStatusCode(v int32) *StartListenerResponse {
	s.StatusCode = &v
	return s
}

func (s *StartListenerResponse) SetBody(v *StartListenerResponseBody) *StartListenerResponse {
	s.Body = v
	return s
}

func (s *StartListenerResponse) Validate() error {
	return dara.Validate(s)
}
