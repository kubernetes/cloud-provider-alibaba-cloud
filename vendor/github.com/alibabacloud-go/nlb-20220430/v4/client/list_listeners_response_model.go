// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListListenersResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListListenersResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListListenersResponse
	GetStatusCode() *int32
	SetBody(v *ListListenersResponseBody) *ListListenersResponse
	GetBody() *ListListenersResponseBody
}

type ListListenersResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListListenersResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListListenersResponse) String() string {
	return dara.Prettify(s)
}

func (s ListListenersResponse) GoString() string {
	return s.String()
}

func (s *ListListenersResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListListenersResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListListenersResponse) GetBody() *ListListenersResponseBody {
	return s.Body
}

func (s *ListListenersResponse) SetHeaders(v map[string]*string) *ListListenersResponse {
	s.Headers = v
	return s
}

func (s *ListListenersResponse) SetStatusCode(v int32) *ListListenersResponse {
	s.StatusCode = &v
	return s
}

func (s *ListListenersResponse) SetBody(v *ListListenersResponseBody) *ListListenersResponse {
	s.Body = v
	return s
}

func (s *ListListenersResponse) Validate() error {
	return dara.Validate(s)
}
