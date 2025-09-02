// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteListenerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DeleteListenerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DeleteListenerResponse
	GetStatusCode() *int32
	SetBody(v *DeleteListenerResponseBody) *DeleteListenerResponse
	GetBody() *DeleteListenerResponseBody
}

type DeleteListenerResponse struct {
	Headers    map[string]*string          `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                      `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DeleteListenerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DeleteListenerResponse) String() string {
	return dara.Prettify(s)
}

func (s DeleteListenerResponse) GoString() string {
	return s.String()
}

func (s *DeleteListenerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DeleteListenerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DeleteListenerResponse) GetBody() *DeleteListenerResponseBody {
	return s.Body
}

func (s *DeleteListenerResponse) SetHeaders(v map[string]*string) *DeleteListenerResponse {
	s.Headers = v
	return s
}

func (s *DeleteListenerResponse) SetStatusCode(v int32) *DeleteListenerResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteListenerResponse) SetBody(v *DeleteListenerResponseBody) *DeleteListenerResponse {
	s.Body = v
	return s
}

func (s *DeleteListenerResponse) Validate() error {
	return dara.Validate(s)
}
