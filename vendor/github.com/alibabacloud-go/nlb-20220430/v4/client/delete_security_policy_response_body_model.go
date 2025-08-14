// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iDeleteSecurityPolicyResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *DeleteSecurityPolicyResponseBody
	GetRequestId() *string
}

type DeleteSecurityPolicyResponseBody struct {
	// The ID of the request.
	//
	// example:
	//
	// D7A8875F-373A-5F48-8484-25B07A61F2AF
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteSecurityPolicyResponseBody) String() string {
	return dara.Prettify(s)
}

func (s DeleteSecurityPolicyResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteSecurityPolicyResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *DeleteSecurityPolicyResponseBody) SetRequestId(v string) *DeleteSecurityPolicyResponseBody {
	s.RequestId = &v
	return s
}

func (s *DeleteSecurityPolicyResponseBody) Validate() error {
	return dara.Validate(s)
}
