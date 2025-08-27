// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateServerGroupResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *CreateServerGroupResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *CreateServerGroupResponse
	GetStatusCode() *int32
	SetBody(v *CreateServerGroupResponseBody) *CreateServerGroupResponse
	GetBody() *CreateServerGroupResponseBody
}

type CreateServerGroupResponse struct {
	Headers    map[string]*string             `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                         `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *CreateServerGroupResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s CreateServerGroupResponse) String() string {
	return dara.Prettify(s)
}

func (s CreateServerGroupResponse) GoString() string {
	return s.String()
}

func (s *CreateServerGroupResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *CreateServerGroupResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *CreateServerGroupResponse) GetBody() *CreateServerGroupResponseBody {
	return s.Body
}

func (s *CreateServerGroupResponse) SetHeaders(v map[string]*string) *CreateServerGroupResponse {
	s.Headers = v
	return s
}

func (s *CreateServerGroupResponse) SetStatusCode(v int32) *CreateServerGroupResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateServerGroupResponse) SetBody(v *CreateServerGroupResponseBody) *CreateServerGroupResponse {
	s.Body = v
	return s
}

func (s *CreateServerGroupResponse) Validate() error {
	return dara.Validate(s)
}
