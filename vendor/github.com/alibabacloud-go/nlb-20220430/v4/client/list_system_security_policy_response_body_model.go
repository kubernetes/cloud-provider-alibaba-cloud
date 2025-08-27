// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListSystemSecurityPolicyResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetRequestId(v string) *ListSystemSecurityPolicyResponseBody
	GetRequestId() *string
	SetSecurityPolicies(v []*ListSystemSecurityPolicyResponseBodySecurityPolicies) *ListSystemSecurityPolicyResponseBody
	GetSecurityPolicies() []*ListSystemSecurityPolicyResponseBodySecurityPolicies
}

type ListSystemSecurityPolicyResponseBody struct {
	// The request ID.
	//
	// example:
	//
	// 5C057647-284B-5C67-A07E-4B8F3DABA9F9
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// A list of TLS security policies.
	SecurityPolicies []*ListSystemSecurityPolicyResponseBodySecurityPolicies `json:"SecurityPolicies,omitempty" xml:"SecurityPolicies,omitempty" type:"Repeated"`
}

func (s ListSystemSecurityPolicyResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListSystemSecurityPolicyResponseBody) GoString() string {
	return s.String()
}

func (s *ListSystemSecurityPolicyResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListSystemSecurityPolicyResponseBody) GetSecurityPolicies() []*ListSystemSecurityPolicyResponseBodySecurityPolicies {
	return s.SecurityPolicies
}

func (s *ListSystemSecurityPolicyResponseBody) SetRequestId(v string) *ListSystemSecurityPolicyResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListSystemSecurityPolicyResponseBody) SetSecurityPolicies(v []*ListSystemSecurityPolicyResponseBodySecurityPolicies) *ListSystemSecurityPolicyResponseBody {
	s.SecurityPolicies = v
	return s
}

func (s *ListSystemSecurityPolicyResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListSystemSecurityPolicyResponseBodySecurityPolicies struct {
	// The cipher suite.
	//
	// example:
	//
	// ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-ECDSA-AES128-SHA256,ECDHE-ECDSA-AES256-SHA384,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-RSA-AES128-SHA256,ECDHE-RSA-AES256-SHA384,AES128-GCM-SHA256,AES256-GCM-SHA384,AES128-SHA256,AES256-SHA256,ECDHE-ECDSA-AES128-SHA,ECDHE-ECDSA-AES256-SHA,ECDHE-RSA-AES128-SHA,ECDHE-RSA-AES256-SHA,AES128-SHA,AES256-SHA,DES-CBC3-SHA
	Ciphers *string `json:"Ciphers,omitempty" xml:"Ciphers,omitempty"`
	// The ID of the TLS security policy.
	//
	// example:
	//
	// sp-3fdab6dkkke10s****
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
	// The name of the TLS security policy.
	//
	// example:
	//
	// test
	SecurityPolicyName *string `json:"SecurityPolicyName,omitempty" xml:"SecurityPolicyName,omitempty"`
	// The TLS version.
	//
	// example:
	//
	// TLSv1.0
	TlsVersion *string `json:"TlsVersion,omitempty" xml:"TlsVersion,omitempty"`
}

func (s ListSystemSecurityPolicyResponseBodySecurityPolicies) String() string {
	return dara.Prettify(s)
}

func (s ListSystemSecurityPolicyResponseBodySecurityPolicies) GoString() string {
	return s.String()
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) GetCiphers() *string {
	return s.Ciphers
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) GetSecurityPolicyName() *string {
	return s.SecurityPolicyName
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) GetTlsVersion() *string {
	return s.TlsVersion
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) SetCiphers(v string) *ListSystemSecurityPolicyResponseBodySecurityPolicies {
	s.Ciphers = &v
	return s
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) SetSecurityPolicyId(v string) *ListSystemSecurityPolicyResponseBodySecurityPolicies {
	s.SecurityPolicyId = &v
	return s
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) SetSecurityPolicyName(v string) *ListSystemSecurityPolicyResponseBodySecurityPolicies {
	s.SecurityPolicyName = &v
	return s
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) SetTlsVersion(v string) *ListSystemSecurityPolicyResponseBodySecurityPolicies {
	s.TlsVersion = &v
	return s
}

func (s *ListSystemSecurityPolicyResponseBodySecurityPolicies) Validate() error {
	return dara.Validate(s)
}
