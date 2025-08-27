// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateSecurityPolicyAttributeRequest interface {
	dara.Model
	String() string
	GoString() string
	SetCiphers(v []*string) *UpdateSecurityPolicyAttributeRequest
	GetCiphers() []*string
	SetClientToken(v string) *UpdateSecurityPolicyAttributeRequest
	GetClientToken() *string
	SetDryRun(v bool) *UpdateSecurityPolicyAttributeRequest
	GetDryRun() *bool
	SetRegionId(v string) *UpdateSecurityPolicyAttributeRequest
	GetRegionId() *string
	SetSecurityPolicyId(v string) *UpdateSecurityPolicyAttributeRequest
	GetSecurityPolicyId() *string
	SetSecurityPolicyName(v string) *UpdateSecurityPolicyAttributeRequest
	GetSecurityPolicyName() *string
	SetTlsVersions(v []*string) *UpdateSecurityPolicyAttributeRequest
	GetTlsVersions() []*string
}

type UpdateSecurityPolicyAttributeRequest struct {
	// The cipher suites supported by the security policy. Valid values of this parameter vary based on the value of TlsVersions. You can specify up to 32 cipher suites.
	//
	// TLSv1.0 and TLSv1.1 support the following cipher suites:
	//
	// 	- **ECDHE-ECDSA-AES128-SHA**
	//
	// 	- **ECDHE-ECDSA-AES256-SHA**
	//
	// 	- **ECDHE-RSA-AES128-SHA**
	//
	// 	- **ECDHE-RSA-AES256-SHA**
	//
	// 	- **AES128-SHA**
	//
	// 	- **AES256-SHA**
	//
	// 	- **DES-CBC3-SHA**
	//
	// TLSv1.2 supports the following cipher suites:
	//
	// 	- **ECDHE-ECDSA-AES128-SHA**
	//
	// 	- **ECDHE-ECDSA-AES256-SHA**
	//
	// 	- **ECDHE-RSA-AES128-SHA**
	//
	// 	- **ECDHE-RSA-AES256-SHA**
	//
	// 	- **AES128-SHA**
	//
	// 	- **AES256-SHA**
	//
	// 	- **DES-CBC3-SHA**
	//
	// 	- **ECDHE-ECDSA-AES128-GCM-SHA256**
	//
	// 	- **ECDHE-ECDSA-AES256-GCM-SHA384**
	//
	// 	- **ECDHE-ECDSA-AES128-SHA256**
	//
	// 	- **ECDHE-ECDSA-AES256-SHA384**
	//
	// 	- **ECDHE-RSA-AES128-GCM-SHA256**
	//
	// 	- **ECDHE-RSA-AES256-GCM-SHA384**
	//
	// 	- **ECDHE-RSA-AES128-SHA256**
	//
	// 	- **ECDHE-RSA-AES256-SHA384**
	//
	// 	- **AES128-GCM-SHA256**
	//
	// 	- **AES256-GCM-SHA384**
	//
	// 	- **AES128-SHA256**
	//
	// 	- **AES256-SHA256**
	//
	// TLSv1.3 supports the following cipher suites:
	//
	// 	- **TLS_AES_128_GCM_SHA256**
	//
	// 	- **TLS_AES_256_GCM_SHA384**
	//
	// 	- **TLS_CHACHA20_POLY1305_SHA256**
	//
	// 	- **TLS_AES_128_CCM_SHA256**
	//
	// 	- **TLS_AES_128_CCM_8_SHA256**
	Ciphers []*string `json:"Ciphers,omitempty" xml:"Ciphers,omitempty" type:"Repeated"`
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
	// Specifies whether to perform a dry run. Valid values:
	//
	// 	- **true**: validates the request without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the validation, the corresponding error message is returned. If the request passes the validation, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): validates the request and performs the operation. If the request passes the validation, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// false
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The ID of the TLS security policy.
	//
	// This parameter is required.
	//
	// example:
	//
	// tls-bp14bb1e7dll4f****
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
	// The name of the security policy.
	//
	// The name must be 1 to 200 characters in length, and can contain letters, digits, periods (.), underscores (_), and hyphens (-).
	//
	// example:
	//
	// TLSCipherPolicy
	SecurityPolicyName *string `json:"SecurityPolicyName,omitempty" xml:"SecurityPolicyName,omitempty"`
	// The supported TLS versions. Valid values: **TLSv1.0**, **TLSv1.1**, **TLSv1.2**, and **TLSv1.3**. You can specify up to four TLS versions.
	TlsVersions []*string `json:"TlsVersions,omitempty" xml:"TlsVersions,omitempty" type:"Repeated"`
}

func (s UpdateSecurityPolicyAttributeRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateSecurityPolicyAttributeRequest) GoString() string {
	return s.String()
}

func (s *UpdateSecurityPolicyAttributeRequest) GetCiphers() []*string {
	return s.Ciphers
}

func (s *UpdateSecurityPolicyAttributeRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateSecurityPolicyAttributeRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateSecurityPolicyAttributeRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateSecurityPolicyAttributeRequest) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *UpdateSecurityPolicyAttributeRequest) GetSecurityPolicyName() *string {
	return s.SecurityPolicyName
}

func (s *UpdateSecurityPolicyAttributeRequest) GetTlsVersions() []*string {
	return s.TlsVersions
}

func (s *UpdateSecurityPolicyAttributeRequest) SetCiphers(v []*string) *UpdateSecurityPolicyAttributeRequest {
	s.Ciphers = v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) SetClientToken(v string) *UpdateSecurityPolicyAttributeRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) SetDryRun(v bool) *UpdateSecurityPolicyAttributeRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) SetRegionId(v string) *UpdateSecurityPolicyAttributeRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) SetSecurityPolicyId(v string) *UpdateSecurityPolicyAttributeRequest {
	s.SecurityPolicyId = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) SetSecurityPolicyName(v string) *UpdateSecurityPolicyAttributeRequest {
	s.SecurityPolicyName = &v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) SetTlsVersions(v []*string) *UpdateSecurityPolicyAttributeRequest {
	s.TlsVersions = v
	return s
}

func (s *UpdateSecurityPolicyAttributeRequest) Validate() error {
	return dara.Validate(s)
}
