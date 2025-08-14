// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateServerGroupServersAttributeResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *UpdateServerGroupServersAttributeResponseBody
	GetJobId() *string
	SetRequestId(v string) *UpdateServerGroupServersAttributeResponseBody
	GetRequestId() *string
	SetServerGroupId(v string) *UpdateServerGroupServersAttributeResponseBody
	GetServerGroupId() *string
}

type UpdateServerGroupServersAttributeResponseBody struct {
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

func (s UpdateServerGroupServersAttributeResponseBody) String() string {
	return dara.Prettify(s)
}

func (s UpdateServerGroupServersAttributeResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateServerGroupServersAttributeResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *UpdateServerGroupServersAttributeResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *UpdateServerGroupServersAttributeResponseBody) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *UpdateServerGroupServersAttributeResponseBody) SetJobId(v string) *UpdateServerGroupServersAttributeResponseBody {
	s.JobId = &v
	return s
}

func (s *UpdateServerGroupServersAttributeResponseBody) SetRequestId(v string) *UpdateServerGroupServersAttributeResponseBody {
	s.RequestId = &v
	return s
}

func (s *UpdateServerGroupServersAttributeResponseBody) SetServerGroupId(v string) *UpdateServerGroupServersAttributeResponseBody {
	s.ServerGroupId = &v
	return s
}

func (s *UpdateServerGroupServersAttributeResponseBody) Validate() error {
	return dara.Validate(s)
}
