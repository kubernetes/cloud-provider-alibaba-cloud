// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAddServersToServerGroupResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *AddServersToServerGroupResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *AddServersToServerGroupResponse
	GetStatusCode() *int32
	SetBody(v *AddServersToServerGroupResponseBody) *AddServersToServerGroupResponse
	GetBody() *AddServersToServerGroupResponseBody
}

type AddServersToServerGroupResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *AddServersToServerGroupResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s AddServersToServerGroupResponse) String() string {
	return dara.Prettify(s)
}

func (s AddServersToServerGroupResponse) GoString() string {
	return s.String()
}

func (s *AddServersToServerGroupResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *AddServersToServerGroupResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *AddServersToServerGroupResponse) GetBody() *AddServersToServerGroupResponseBody {
	return s.Body
}

func (s *AddServersToServerGroupResponse) SetHeaders(v map[string]*string) *AddServersToServerGroupResponse {
	s.Headers = v
	return s
}

func (s *AddServersToServerGroupResponse) SetStatusCode(v int32) *AddServersToServerGroupResponse {
	s.StatusCode = &v
	return s
}

func (s *AddServersToServerGroupResponse) SetBody(v *AddServersToServerGroupResponseBody) *AddServersToServerGroupResponse {
	s.Body = v
	return s
}

func (s *AddServersToServerGroupResponse) Validate() error {
	return dara.Validate(s)
}
