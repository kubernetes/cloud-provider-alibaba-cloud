// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetListenerAttributeResponse interface {
	dara.Model
	String() string
	GoString() string
	SetHeaders(v map[string]*string) *GetListenerAttributeResponse
	GetHeaders() map[string]*string
	SetStatusCode(v int32) *GetListenerAttributeResponse
	GetStatusCode() *int32
	SetBody(v *GetListenerAttributeResponseBody) *GetListenerAttributeResponse
	GetBody() *GetListenerAttributeResponseBody
}

type GetListenerAttributeResponse struct {
	Headers    map[string]*string                `json:"headers,omitempty" xml:"headers,omitempty"`
	StatusCode *int32                            `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Body       *GetListenerAttributeResponseBody `json:"body,omitempty" xml:"body,omitempty"`
}

func (s GetListenerAttributeResponse) String() string {
	return dara.Prettify(s)
}

func (s GetListenerAttributeResponse) GoString() string {
	return s.String()
}

func (s *GetListenerAttributeResponse) GetHeaders() map[string]*string {
	return s.Headers
}

func (s *GetListenerAttributeResponse) GetStatusCode() *int32 {
	return s.StatusCode
}

func (s *GetListenerAttributeResponse) GetBody() *GetListenerAttributeResponseBody {
	return s.Body
}

func (s *GetListenerAttributeResponse) SetHeaders(v map[string]*string) *GetListenerAttributeResponse {
	s.Headers = v
	return s
}

func (s *GetListenerAttributeResponse) SetStatusCode(v int32) *GetListenerAttributeResponse {
	s.StatusCode = &v
	return s
}

func (s *GetListenerAttributeResponse) SetBody(v *GetListenerAttributeResponseBody) *GetListenerAttributeResponse {
	s.Body = v
	return s
}

func (s *GetListenerAttributeResponse) Validate() error {
	return dara.Validate(s)
}
