// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iRemoveServersFromServerGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *RemoveServersFromServerGroupResponseBody
	GetJobId() *string
	SetRequestId(v string) *RemoveServersFromServerGroupResponseBody
	GetRequestId() *string
	SetServerGroupId(v string) *RemoveServersFromServerGroupResponseBody
	GetServerGroupId() *string
}

type RemoveServersFromServerGroupResponseBody struct {
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

func (s RemoveServersFromServerGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s RemoveServersFromServerGroupResponseBody) GoString() string {
	return s.String()
}

func (s *RemoveServersFromServerGroupResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *RemoveServersFromServerGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *RemoveServersFromServerGroupResponseBody) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *RemoveServersFromServerGroupResponseBody) SetJobId(v string) *RemoveServersFromServerGroupResponseBody {
	s.JobId = &v
	return s
}

func (s *RemoveServersFromServerGroupResponseBody) SetRequestId(v string) *RemoveServersFromServerGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *RemoveServersFromServerGroupResponseBody) SetServerGroupId(v string) *RemoveServersFromServerGroupResponseBody {
	s.ServerGroupId = &v
	return s
}

func (s *RemoveServersFromServerGroupResponseBody) Validate() error {
	return dara.Validate(s)
}
