// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateServerGroupServersAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateServerGroupServersAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateServerGroupServersAttributeResponse
	GetStatusCode() *int32
	SetBody(v *UpdateServerGroupServersAttributeResponseBody) *UpdateServerGroupServersAttributeResponse
	GetBody() *UpdateServerGroupServersAttributeResponseBody
}

type UpdateServerGroupServersAttributeResponse struct {
	Headers    map[string]*string                             `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                                         `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateServerGroupServersAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateServerGroupServersAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateServerGroupServersAttributeResponse) GoString() string {
	return s.String()
}

func (s *UpdateServerGroupServersAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateServerGroupServersAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateServerGroupServersAttributeResponse) GetBody() *UpdateServerGroupServersAttributeResponseBody {
	return s.Body
}

func (s *UpdateServerGroupServersAttributeResponse) SetHeaders(v map[string]*string) *UpdateServerGroupServersAttributeResponse {
	s.Headers = v
	return s
}

func (s *UpdateServerGroupServersAttributeResponse) SetStatusCode(v int32) *UpdateServerGroupServersAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateServerGroupServersAttributeResponse) SetBody(v *UpdateServerGroupServersAttributeResponseBody) *UpdateServerGroupServersAttributeResponse {
	s.Body = v
	return s
}

func (s *UpdateServerGroupServersAttributeResponse) Validate() error {
	return dara.Validate(s)
}
