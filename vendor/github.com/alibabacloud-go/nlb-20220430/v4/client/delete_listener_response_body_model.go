// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteListenerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *DeleteListenerResponseBody
	GetJobId() *string
	SetRequestId(v string) *DeleteListenerResponseBody
	GetRequestId() *string
}

type DeleteListenerResponseBody struct {
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

func (s DeleteListenerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DeleteListenerResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteListenerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *DeleteListenerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DeleteListenerResponseBody) SetJobId(v string) *DeleteListenerResponseBody {
	s.JobId = &v
	return s
}

func (s *DeleteListenerResponseBody) SetRequestId(v string) *DeleteListenerResponseBody {
	s.RequestId = &v
	return s
}

func (s *DeleteListenerResponseBody) Validate() error {
	return dara.Validate(s)
}
