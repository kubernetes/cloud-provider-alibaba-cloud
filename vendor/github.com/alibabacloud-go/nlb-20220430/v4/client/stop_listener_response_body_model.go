// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStopListenerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *StopListenerResponseBody
	GetJobId() *string
	SetRequestId(v string) *StopListenerResponseBody
	GetRequestId() *string
}

type StopListenerResponseBody struct {
	// The ID of the asynchronous task.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The request ID.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s StopListenerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s StopListenerResponseBody) GoString() string {
	return s.String()
}

func (s *StopListenerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *StopListenerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *StopListenerResponseBody) SetJobId(v string) *StopListenerResponseBody {
	s.JobId = &v
	return s
}

func (s *StopListenerResponseBody) SetRequestId(v string) *StopListenerResponseBody {
	s.RequestId = &v
	return s
}

func (s *StopListenerResponseBody) Validate() error {
	return dara.Validate(s)
}
