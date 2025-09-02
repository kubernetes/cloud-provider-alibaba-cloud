// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iStartListenerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *StartListenerResponseBody
	GetJobId() *string
	SetRequestId(v string) *StartListenerResponseBody
	GetRequestId() *string
}

type StartListenerResponseBody struct {
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

func (s StartListenerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s StartListenerResponseBody) GoString() string {
	return s.String()
}

func (s *StartListenerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *StartListenerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *StartListenerResponseBody) SetJobId(v string) *StartListenerResponseBody {
	s.JobId = &v
	return s
}

func (s *StartListenerResponseBody) SetRequestId(v string) *StartListenerResponseBody {
	s.RequestId = &v
	return s
}

func (s *StartListenerResponseBody) Validate() error {
	return dara.Validate(s)
}
