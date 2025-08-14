// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateListenerAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *UpdateListenerAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *UpdateListenerAttributeResponse
	GetStatusCode() *int32
	SetBody(v *UpdateListenerAttributeResponseBody) *UpdateListenerAttributeResponse
	GetBody() *UpdateListenerAttributeResponseBody
}

type UpdateListenerAttributeResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *UpdateListenerAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s UpdateListenerAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s UpdateListenerAttributeResponse) GoString() string {
	return s.String()
}

func (s *UpdateListenerAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *UpdateListenerAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *UpdateListenerAttributeResponse) GetBody() *UpdateListenerAttributeResponseBody {
	return s.Body
}

func (s *UpdateListenerAttributeResponse) SetHeaders(v map[string]*string) *UpdateListenerAttributeResponse {
	s.Headers = v
	return s
}

func (s *UpdateListenerAttributeResponse) SetStatusCode(v int32) *UpdateListenerAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateListenerAttributeResponse) SetBody(v *UpdateListenerAttributeResponseBody) *UpdateListenerAttributeResponse {
	s.Body = v
	return s
}

func (s *UpdateListenerAttributeResponse) Validate() error {
	return dara.Validate(s)
}
