// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListServerGroupServersResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListServerGroupServersResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListServerGroupServersResponse
	GetStatusCode() *int32
	SetBody(v *ListServerGroupServersResponseBody) *ListServerGroupServersResponse
	GetBody() *ListServerGroupServersResponseBody
}

type ListServerGroupServersResponse struct {
	Headers    map[string]*string                  `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                              `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListServerGroupServersResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListServerGroupServersResponse) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupServersResponse) GoString() string {
	return s.String()
}

func (s *ListServerGroupServersResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListServerGroupServersResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListServerGroupServersResponse) GetBody() *ListServerGroupServersResponseBody {
	return s.Body
}

func (s *ListServerGroupServersResponse) SetHeaders(v map[string]*string) *ListServerGroupServersResponse {
	s.Headers = v
	return s
}

func (s *ListServerGroupServersResponse) SetStatusCode(v int32) *ListServerGroupServersResponse {
	s.StatusCode = &v
	return s
}

func (s *ListServerGroupServersResponse) SetBody(v *ListServerGroupServersResponseBody) *ListServerGroupServersResponse {
	s.Body = v
	return s
}

func (s *ListServerGroupServersResponse) Validate() error {
	return dara.Validate(s)
}
