// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetJobStatusResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *GetJobStatusResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *GetJobStatusResponse
	GetStatusCode() *int32
	SetBody(v *GetJobStatusResponseBody) *GetJobStatusResponse
	GetBody() *GetJobStatusResponseBody
}

type GetJobStatusResponse struct {
	Headers    map[string]*string        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *GetJobStatusResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s GetJobStatusResponse) String() string {
	return dara.Prettify(s)
}

func (s GetJobStatusResponse) GoString() string {
	return s.String()
}

func (s *GetJobStatusResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *GetJobStatusResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *GetJobStatusResponse) GetBody() *GetJobStatusResponseBody {
	return s.Body
}

func (s *GetJobStatusResponse) SetHeaders(v map[string]*string) *GetJobStatusResponse {
	s.Headers = v
	return s
}

func (s *GetJobStatusResponse) SetStatusCode(v int32) *GetJobStatusResponse {
	s.StatusCode = &v
	return s
}

func (s *GetJobStatusResponse) SetBody(v *GetJobStatusResponseBody) *GetJobStatusResponse {
	s.Body = v
	return s
}

func (s *GetJobStatusResponse) Validate() error {
	return dara.Validate(s)
}
