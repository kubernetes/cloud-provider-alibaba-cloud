// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListAsynJobsRequest interface {
	dara.Model
	String() string
	GoString() string
	SetJobIds(v []*string) *ListAsynJobsRequest
	GetJobIds() []*string
}

type ListAsynJobsRequest struct {
	// The task IDs.
	JobIds []*string `json:"JobIds,omitempty" xml:"JobIds,omitempty" type:"Repeated"`
}

func (s ListAsynJobsRequest) String() string {
	return dara.Prettify(s)
}

func (s ListAsynJobsRequest) GoString() string {
	return s.String()
}

func (s *ListAsynJobsRequest) GetJobIds() []*string {
	return s.JobIds
}

func (s *ListAsynJobsRequest) SetJobIds(v []*string) *ListAsynJobsRequest {
	s.JobIds = v
	return s
}

func (s *ListAsynJobsRequest) Validate() error {
	return dara.Validate(s)
}
