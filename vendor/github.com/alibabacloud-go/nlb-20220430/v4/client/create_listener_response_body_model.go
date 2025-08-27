// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateListenerResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *CreateListenerResponseBody
	GetJobId() *string
	SetListenerId(v string) *CreateListenerResponseBody
	GetListenerId() *string
	SetRequestId(v string) *CreateListenerResponseBody
	GetRequestId() *string
}

type CreateListenerResponseBody struct {
	// The asynchronous task ID.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
	// The listener ID.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@80
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The request ID.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateListenerResponseBody) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerResponseBody) GoString() string {
	return s.String()
}

func (s *CreateListenerResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *CreateListenerResponseBody) GetListenerId() *string {
	return s.ListenerId
}

func (s *CreateListenerResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *CreateListenerResponseBody) SetJobId(v string) *CreateListenerResponseBody {
	s.JobId = &v
	return s
}

func (s *CreateListenerResponseBody) SetListenerId(v string) *CreateListenerResponseBody {
	s.ListenerId = &v
	return s
}

func (s *CreateListenerResponseBody) SetRequestId(v string) *CreateListenerResponseBody {
	s.RequestId = &v
	return s
}

func (s *CreateListenerResponseBody) Validate() error {
	return dara.Validate(s)
}
