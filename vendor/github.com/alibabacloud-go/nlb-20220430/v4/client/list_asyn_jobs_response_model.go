// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListAsynJobsResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *ListAsynJobsResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *ListAsynJobsResponse
	GetStatusCode() *int32
	SetBody(v *ListAsynJobsResponseBody) *ListAsynJobsResponse
	GetBody() *ListAsynJobsResponseBody
}

type ListAsynJobsResponse struct {
	Headers    map[string]*string        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *ListAsynJobsResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s ListAsynJobsResponse) String() string {
	return dara.Prettify(s)
}

func (s ListAsynJobsResponse) GoString() string {
	return s.String()
}

func (s *ListAsynJobsResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *ListAsynJobsResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *ListAsynJobsResponse) GetBody() *ListAsynJobsResponseBody {
	return s.Body
}

func (s *ListAsynJobsResponse) SetHeaders(v map[string]*string) *ListAsynJobsResponse {
	s.Headers = v
	return s
}

func (s *ListAsynJobsResponse) SetStatusCode(v int32) *ListAsynJobsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListAsynJobsResponse) SetBody(v *ListAsynJobsResponseBody) *ListAsynJobsResponse {
	s.Body = v
	return s
}

func (s *ListAsynJobsResponse) Validate() error {
	return dara.Validate(s)
}
