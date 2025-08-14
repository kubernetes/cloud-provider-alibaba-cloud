// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateSecurityPolicyAttributeResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetJobId(v string) *UpdateSecurityPolicyAttributeResponseBody
	GetJobId() *string
	SetRequestId(v string) *UpdateSecurityPolicyAttributeResponseBody
	GetRequestId() *string
	SetSecurityPolicyId(v string) *UpdateSecurityPolicyAttributeResponseBody
	GetSecurityPolicyId() *string
}

type UpdateSecurityPolicyAttributeResponseBody struct {
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
	// D7A8875F-373A-5F48-8484-25B07A61F2AF
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The ID of the TLS security policy.
	//
	// example:
	//
	// tls-bp14bb1e7dll4f****
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
}

func (s UpdateSecurityPolicyAttributeResponseBody) String() string {
	return dara.Prettify(s)
}

func (s UpdateSecurityPolicyAttributeResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateSecurityPolicyAttributeResponseBody) GetJobId() *string {
	return s.JobId
}

func (s *UpdateSecurityPolicyAttributeResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *UpdateSecurityPolicyAttributeResponseBody) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *UpdateSecurityPolicyAttributeResponseBody) SetJobId(v string) *UpdateSecurityPolicyAttributeResponseBody {
	s.JobId = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeResponseBody) SetRequestId(v string) *UpdateSecurityPolicyAttributeResponseBody {
	s.RequestId = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeResponseBody) SetSecurityPolicyId(v string) *UpdateSecurityPolicyAttributeResponseBody {
	s.SecurityPolicyId = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeResponseBody) Validate() error {
	return dara.Validate(s)
}
