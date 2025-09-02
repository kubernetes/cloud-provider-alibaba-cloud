// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iRemoveServersFromServerGroupResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *RemoveServersFromServerGroupResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *RemoveServersFromServerGroupResponse
	GetStatusCode() *int32
	SetBody(v *RemoveServersFromServerGroupResponseBody) *RemoveServersFromServerGroupResponse
	GetBody() *RemoveServersFromServerGroupResponseBody
}

type RemoveServersFromServerGroupResponse struct {
	Headers    map[string]*string                        `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                    `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *RemoveServersFromServerGroupResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s RemoveServersFromServerGroupResponse) String() string {
	return dara.Prettify(s)
}

func (s RemoveServersFromServerGroupResponse) GoString() string {
	return s.String()
}

func (s *RemoveServersFromServerGroupResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *RemoveServersFromServerGroupResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *RemoveServersFromServerGroupResponse) GetBody() *RemoveServersFromServerGroupResponseBody {
	return s.Body
}

func (s *RemoveServersFromServerGroupResponse) SetHeaders(v map[string]*string) *RemoveServersFromServerGroupResponse {
	s.Headers = v
	return s
}

func (s *RemoveServersFromServerGroupResponse) SetStatusCode(v int32) *RemoveServersFromServerGroupResponse {
	s.StatusCode = &v
	return s
}

func (s *RemoveServersFromServerGroupResponse) SetBody(v *RemoveServersFromServerGroupResponseBody) *RemoveServersFromServerGroupResponse {
	s.Body = v
	return s
}

func (s *RemoveServersFromServerGroupResponse) Validate() error {
	return dara.Validate(s)
}
