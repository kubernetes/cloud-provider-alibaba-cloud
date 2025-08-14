// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAddServersToServerGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *AddServersToServerGroupResponseBody
	GetJobId() *string
	SetRequestId(v string) *AddServersToServerGroupResponseBody
	GetRequestId() *string
	SetServerGroupId(v string) *AddServersToServerGroupResponseBody
	GetServerGroupId() *string
}

type AddServersToServerGroupResponseBody struct {
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
	// 54B48E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The ID of the server group.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
}

func (s AddServersToServerGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s AddServersToServerGroupResponseBody) GoString() string {
	return s.String()
}

func (s *AddServersToServerGroupResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *AddServersToServerGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *AddServersToServerGroupResponseBody) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *AddServersToServerGroupResponseBody) SetJobId(v string) *AddServersToServerGroupResponseBody {
	s.JobId = &v
	return s
}

func (s *AddServersToServerGroupResponseBody) SetRequestId(v string) *AddServersToServerGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *AddServersToServerGroupResponseBody) SetServerGroupId(v string) *AddServersToServerGroupResponseBody {
	s.ServerGroupId = &v
	return s
}

func (s *AddServersToServerGroupResponseBody) Validate() error {
	return dara.Validate(s)
}
