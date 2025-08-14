// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateServerGroupAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateServerGroupAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateServerGroupAttributeResponse
	GetStatusCode() *int32
	SetBody(v *UpdateServerGroupAttributeResponseBody) *UpdateServerGroupAttributeResponse
	GetBody() *UpdateServerGroupAttributeResponseBody
}

type UpdateServerGroupAttributeResponse struct {
	Headers    map[string]*string                      `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                  `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateServerGroupAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateServerGroupAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateServerGroupAttributeResponse) GoString() string {
	return s.String()
}

func (s *UpdateServerGroupAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateServerGroupAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateServerGroupAttributeResponse) GetBody() *UpdateServerGroupAttributeResponseBody {
	return s.Body
}

func (s *UpdateServerGroupAttributeResponse) SetHeaders(v map[string]*string) *UpdateServerGroupAttributeResponse {
	s.Headers = v
	return s
}

func (s *UpdateServerGroupAttributeResponse) SetStatusCode(v int32) *UpdateServerGroupAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateServerGroupAttributeResponse) SetBody(v *UpdateServerGroupAttributeResponseBody) *UpdateServerGroupAttributeResponse {
	s.Body = v
	return s
}

func (s *UpdateServerGroupAttributeResponse) Validate() error {
	return dara.Validate(s)
}
