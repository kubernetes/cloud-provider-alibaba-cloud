// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateListenerResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *CreateListenerResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *CreateListenerResponse
	GetStatusCode() *int32
	SetBody(v *CreateListenerResponseBody) *CreateListenerResponse
	GetBody() *CreateListenerResponseBody
}

type CreateListenerResponse struct {
	Headers    map[string]*string          `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                      `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *CreateListenerResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s CreateListenerResponse) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerResponse) GoString() string {
	return s.String()
}

func (s *CreateListenerResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *CreateListenerResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *CreateListenerResponse) GetBody() *CreateListenerResponseBody {
	return s.Body
}

func (s *CreateListenerResponse) SetHeaders(v map[string]*string) *CreateListenerResponse {
	s.Headers = v
	return s
}

func (s *CreateListenerResponse) SetStatusCode(v int32) *CreateListenerResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateListenerResponse) SetBody(v *CreateListenerResponseBody) *CreateListenerResponse {
	s.Body = v
	return s
}

func (s *CreateListenerResponse) Validate() error {
	return dara.Validate(s)
}
