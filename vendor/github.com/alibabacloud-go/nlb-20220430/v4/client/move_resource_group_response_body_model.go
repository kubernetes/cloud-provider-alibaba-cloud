// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iMoveResourceGroupResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetData(v *MoveResourceGroupResponseBodyData) *MoveResourceGroupResponseBody
	GetData() *MoveResourceGroupResponseBodyData
	SetHttpStatusCode(v int32) *MoveResourceGroupResponseBody
	GetHttpStatusCode() *int32
	SetRequestId(v string) *MoveResourceGroupResponseBody
	GetRequestId() *string
	SetSuccess(v bool) *MoveResourceGroupResponseBody
	GetSuccess() *bool
}

type MoveResourceGroupResponseBody struct {
	// The data returned.
	Data *MoveResourceGroupResponseBodyData `json:"Data,omitempty" xml:"Data,omitempty" type:"Struct"`
	// The HTTP status code returned.
	//
	// example:
	//
	// 200
	HttpStatusCode *int32 `json:"HttpStatusCode,omitempty" xml:"HttpStatusCode,omitempty"`
	// The request ID.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// Indicates whether the request was successful. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// true
	Success *bool `json:"Success,omitempty" xml:"Success,omitempty"`
}

func (s MoveResourceGroupResponseBody) String() string {
	return dara.Prettify(s)
}

func (s MoveResourceGroupResponseBody) GoString() string {
	return s.String()
}

func (s *MoveResourceGroupResponseBody) GetData() *MoveResourceGroupResponseBodyData {
	return s.Data
}

func (s *MoveResourceGroupResponseBody) GetHttpStatusCode() *int32 {
	return s.HttpStatusCode
}

func (s *MoveResourceGroupResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *MoveResourceGroupResponseBody) GetSuccess() *bool {
	return s.Success
}

func (s *MoveResourceGroupResponseBody) SetData(v *MoveResourceGroupResponseBodyData) *MoveResourceGroupResponseBody {
	s.Data = v
	return s
}

func (s *MoveResourceGroupResponseBody) SetHttpStatusCode(v int32) *MoveResourceGroupResponseBody {
	s.HttpStatusCode = &v
	return s
}

func (s *MoveResourceGroupResponseBody) SetRequestId(v string) *MoveResourceGroupResponseBody {
	s.RequestId = &v
	return s
}

func (s *MoveResourceGroupResponseBody) SetSuccess(v bool) *MoveResourceGroupResponseBody {
	s.Success = &v
	return s
}

func (s *MoveResourceGroupResponseBody) Validate() error {
	return dara.Validate(s)
}

type MoveResourceGroupResponseBodyData struct {
	// The ID of the resource. You can specify up to 50 resource IDs in each call.
	//
	// example:
	//
	// nlb-nrnrxwd15en27r****
	ResourceId *string `json:"ResourceId,omitempty" xml:"ResourceId,omitempty"`
}

func (s MoveResourceGroupResponseBodyData) String() string {
	return dara.Prettify(s)
}

func (s MoveResourceGroupResponseBodyData) GoString() string {
	return s.String()
}

func (s *MoveResourceGroupResponseBodyData) GetResourceId() *string {
	return s.ResourceId
}

func (s *MoveResourceGroupResponseBodyData) SetResourceId(v string) *MoveResourceGroupResponseBodyData {
	s.ResourceId = &v
	return s
}

func (s *MoveResourceGroupResponseBodyData) Validate() error {
	return dara.Validate(s)
}
