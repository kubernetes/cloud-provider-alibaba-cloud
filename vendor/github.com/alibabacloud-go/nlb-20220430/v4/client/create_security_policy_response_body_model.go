// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateSecurityPolicyResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *CreateSecurityPolicyResponseBody
	GetJobId() *string
	SetRequestId(v string) *CreateSecurityPolicyResponseBody
	GetRequestId() *string
	SetSecurityPolicyId(v string) *CreateSecurityPolicyResponseBody
	GetSecurityPolicyId() *string
}

type CreateSecurityPolicyResponseBody struct {
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
	// D7A8875F-373A-5F48-8484-25B07A61F2AF
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The ID of the TLS security policy.
	//
	// example:
	//
	// tls-bp14bb1e7dll4f****
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
}

func (s CreateSecurityPolicyResponseBody) String() string {
	return dara.Prettify(s)
}

func (s CreateSecurityPolicyResponseBody) GoString() string {
	return s.String()
}

func (s *CreateSecurityPolicyResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *CreateSecurityPolicyResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *CreateSecurityPolicyResponseBody) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *CreateSecurityPolicyResponseBody) SetJobId(v string) *CreateSecurityPolicyResponseBody {
	s.JobId = &v
	return s
}

func (s *CreateSecurityPolicyResponseBody) SetRequestId(v string) *CreateSecurityPolicyResponseBody {
	s.RequestId = &v
	return s
}

func (s *CreateSecurityPolicyResponseBody) SetSecurityPolicyId(v string) *CreateSecurityPolicyResponseBody {
	s.SecurityPolicyId = &v
	return s
}

func (s *CreateSecurityPolicyResponseBody) Validate() error {
	return dara.Validate(s)
}
