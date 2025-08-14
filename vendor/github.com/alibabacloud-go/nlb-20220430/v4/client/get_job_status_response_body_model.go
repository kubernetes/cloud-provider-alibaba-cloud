// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetJobStatusResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *GetJobStatusResponseBody
	GetRequestId() *string
	SetStatus(v string) *GetJobStatusResponseBody
	GetStatus() *string
}

type GetJobStatusResponseBody struct {
	// The ID of the request.
	//
	// example:
	//
	// 365F4154-92F6-4AE4-92F8-7FF34B540710
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The state of the task. Valid values:
	//
	// 	- **Succeeded**: The task is successful.
	//
	// 	- **processing**: The ticket is being executed.
	//
	// example:
	//
	// Succeeded
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s GetJobStatusResponseBody) String() string {
	return dara.Prettify(s)
}

func (s GetJobStatusResponseBody) GoString() string {
	return s.String()
}

func (s *GetJobStatusResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *GetJobStatusResponseBody) GetStatus() *string {
	return s.Status
}

func (s *GetJobStatusResponseBody) SetRequestId(v string) *GetJobStatusResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetJobStatusResponseBody) SetStatus(v string) *GetJobStatusResponseBody {
	s.Status = &v
	return s
}

func (s *GetJobStatusResponseBody) Validate() error {
	return dara.Validate(s)
}
