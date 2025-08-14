// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListServerGroupsResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListServerGroupsResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListServerGroupsResponse
	GetStatusCode() *int32
	SetBody(v *ListServerGroupsResponseBody) *ListServerGroupsResponse
	GetBody() *ListServerGroupsResponseBody
}

type ListServerGroupsResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListServerGroupsResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListServerGroupsResponse) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsResponse) GoString() string {
	return s.String()
}

func (s *ListServerGroupsResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListServerGroupsResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListServerGroupsResponse) GetBody() *ListServerGroupsResponseBody {
	return s.Body
}

func (s *ListServerGroupsResponse) SetHeaders(v map[string]*string) *ListServerGroupsResponse {
	s.Headers = v
	return s
}

func (s *ListServerGroupsResponse) SetStatusCode(v int32) *ListServerGroupsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListServerGroupsResponse) SetBody(v *ListServerGroupsResponseBody) *ListServerGroupsResponse {
	s.Body = v
	return s
}

func (s *ListServerGroupsResponse) Validate() error {
	return dara.Validate(s)
}
