// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteServerGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *DeleteServerGroupResponseBody
	GetJobId() *string
	SetRequestId(v string) *DeleteServerGroupResponseBody
	GetRequestId() *string
}

type DeleteServerGroupResponseBody struct {
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
}

func (s DeleteServerGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DeleteServerGroupResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteServerGroupResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *DeleteServerGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DeleteServerGroupResponseBody) SetJobId(v string) *DeleteServerGroupResponseBody {
	s.JobId = &v
	return s
}

func (s *DeleteServerGroupResponseBody) SetRequestId(v string) *DeleteServerGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *DeleteServerGroupResponseBody) Validate() error {
	return dara.Validate(s)
}
