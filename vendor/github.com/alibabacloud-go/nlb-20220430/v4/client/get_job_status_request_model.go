// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetJobStatusRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *GetJobStatusRequest
	GetClientToken() *string
	SetJobId(v string) *GetJobStatusRequest
	GetJobId() *string
}

type GetJobStatusRequest struct {
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. Only ASCII characters are allowed.
	//
	// >  If you do not set this parameter, the value of **RequestId*	- is used.***	- The value of **RequestId*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// The ID of the asynchronous job.
	//
	// This parameter is required.
	//
	// example:
	//
	// 72dcd26b-f12d-4c27-b3af-18f6aed5****
	JobId *string `json:"JobId,omitempty" xml:"JobId,omitempty"`
}

func (s GetJobStatusRequest) String() string {
	return dara.Prettify(s)
}

func (s GetJobStatusRequest) GoString() string {
	return s.String()
}

func (s *GetJobStatusRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *GetJobStatusRequest) GetJobId() *string {
	return s.JobId
}

func (s *GetJobStatusRequest) SetClientToken(v string) *GetJobStatusRequest {
	s.ClientToken = &v
	return s
}

func (s *GetJobStatusRequest) SetJobId(v string) *GetJobStatusRequest {
	s.JobId = &v
	return s
}

func (s *GetJobStatusRequest) Validate() error {
	return dara.Validate(s)
}
