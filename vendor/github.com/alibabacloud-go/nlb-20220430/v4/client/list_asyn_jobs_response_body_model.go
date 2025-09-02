// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListAsynJobsResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobs(v []*ListAsynJobsResponseBodyJobs) *ListAsynJobsResponseBody
	GetJobs() []*ListAsynJobsResponseBodyJobs
	SetRequestId(v string) *ListAsynJobsResponseBody
	GetRequestId() *string
	SetTotalCount(v string) *ListAsynJobsResponseBody
	GetTotalCount() *string
}

type ListAsynJobsResponseBody struct {
	// The queried tasks.
	Jobs []*ListAsynJobsResponseBodyJobs `json:"Jobs,omitempty" xml:"Jobs,omitempty" type:"Repeated"`
	// Id of the request
	//
	// example:
	//
	// 365F4154-92F6-4AE4-92F8-7FF3******
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The number of entries returned.
	//
	// example:
	//
	// 1000
	TotalCount *string `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListAsynJobsResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListAsynJobsResponseBody) GoString() string {
	return s.String()
}

func (s *ListAsynJobsResponseBody) GetJobs() []*ListAsynJobsResponseBodyJobs {
	return s.Jobs
}

func (s *ListAsynJobsResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListAsynJobsResponseBody) GetTotalCount() *string {
	return s.TotalCount
}

func (s *ListAsynJobsResponseBody) SetJobs(v []*ListAsynJobsResponseBodyJobs) *ListAsynJobsResponseBody {
	s.Jobs = v
	return s
}

func (s *ListAsynJobsResponseBody) SetRequestId(v string) *ListAsynJobsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListAsynJobsResponseBody) SetTotalCount(v string) *ListAsynJobsResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListAsynJobsResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListAsynJobsResponseBodyJobs struct {
	// The task ID.
	//
	// example:
	//
	// 365F4154-92F6-4AE4-92F8-7FF34B5****
	Id *string `json:"Id,omitempty" xml:"Id,omitempty"`
	// The status of the task. Valid values:
	//
	// 	- Succeeded: The task is successful.
	//
	// 	- Failed: The task fails.
	//
	// 	- Processing: The task is being processed.
	//
	// example:
	//
	// Succeeded
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s ListAsynJobsResponseBodyJobs) String() string {
	return dara.Prettify(s)
}

func (s ListAsynJobsResponseBodyJobs) GoString() string {
	return s.String()
}

func (s *ListAsynJobsResponseBodyJobs) GetId() *string {
	return s.Id
}

func (s *ListAsynJobsResponseBodyJobs) GetStatus() *string {
	return s.Status
}

func (s *ListAsynJobsResponseBodyJobs) SetId(v string) *ListAsynJobsResponseBodyJobs {
	s.Id = &v
	return s
}

func (s *ListAsynJobsResponseBodyJobs) SetStatus(v string) *ListAsynJobsResponseBodyJobs {
	s.Status = &v
	return s
}

func (s *ListAsynJobsResponseBodyJobs) Validate() error {
	return dara.Validate(s)
}
