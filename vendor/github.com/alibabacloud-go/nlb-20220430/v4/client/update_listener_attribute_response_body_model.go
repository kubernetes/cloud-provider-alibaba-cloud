// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateListenerAttributeResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *UpdateListenerAttributeResponseBody
	GetJobId() *string
	SetRequestId(v string) *UpdateListenerAttributeResponseBody
	GetRequestId() *string
}

type UpdateListenerAttributeResponseBody struct {
	// The ID of the asynchronous task.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s UpdateListenerAttributeResponseBody) String() string {
	return dara.Prettify(s)
}

func (s UpdateListenerAttributeResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateListenerAttributeResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *UpdateListenerAttributeResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *UpdateListenerAttributeResponseBody) SetJobId(v string) *UpdateListenerAttributeResponseBody {
	s.JobId = &v
	return s
}

func (s *UpdateListenerAttributeResponseBody) SetRequestId(v string) *UpdateListenerAttributeResponseBody {
	s.RequestId = &v
	return s
}

func (s *UpdateListenerAttributeResponseBody) Validate() error {
	return dara.Validate(s)
}
