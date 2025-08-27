// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateServerGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *CreateServerGroupResponseBody
	GetJobId() *string
	SetRequestId(v string) *CreateServerGroupResponseBody
	GetRequestId() *string
	SetServerGroupId(v string) *CreateServerGroupResponseBody
	GetServerGroupId() *string
}

type CreateServerGroupResponseBody struct {
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
	// 54B48E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The server group ID.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
}

func (s CreateServerGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s CreateServerGroupResponseBody) GoString() string {
	return s.String()
}

func (s *CreateServerGroupResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *CreateServerGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *CreateServerGroupResponseBody) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *CreateServerGroupResponseBody) SetJobId(v string) *CreateServerGroupResponseBody {
	s.JobId = &v
	return s
}

func (s *CreateServerGroupResponseBody) SetRequestId(v string) *CreateServerGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *CreateServerGroupResponseBody) SetServerGroupId(v string) *CreateServerGroupResponseBody {
	s.ServerGroupId = &v
	return s
}

func (s *CreateServerGroupResponseBody) Validate() error {
	return dara.Validate(s)
}
