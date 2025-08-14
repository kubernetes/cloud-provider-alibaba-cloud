// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteServerGroupResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *DeleteServerGroupResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *DeleteServerGroupResponse
	GetStatusCode() *int32
	SetBody(v *DeleteServerGroupResponseBody) *DeleteServerGroupResponse
	GetBody() *DeleteServerGroupResponseBody
}

type DeleteServerGroupResponse struct {
	Headers    map[string]*string             `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                         `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *DeleteServerGroupResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s DeleteServerGroupResponse) String() string {
	return dara.Prettify(s)
}

func (s DeleteServerGroupResponse) GoString() string {
	return s.String()
}

func (s *DeleteServerGroupResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *DeleteServerGroupResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *DeleteServerGroupResponse) GetBody() *DeleteServerGroupResponseBody {
	return s.Body
}

func (s *DeleteServerGroupResponse) SetHeaders(v map[string]*string) *DeleteServerGroupResponse {
	s.Headers = v
	return s
}

func (s *DeleteServerGroupResponse) SetStatusCode(v int32) *DeleteServerGroupResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteServerGroupResponse) SetBody(v *DeleteServerGroupResponseBody) *DeleteServerGroupResponse {
	s.Body = v
	return s
}

func (s *DeleteServerGroupResponse) Validate() error {
	return dara.Validate(s)
}
