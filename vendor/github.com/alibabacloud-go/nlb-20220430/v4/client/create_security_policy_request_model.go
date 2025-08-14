// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateSecurityPolicyRequest interface {
	dara.Model
	String() string
	GoString() string
	SetCiphers(v []*string) *CreateSecurityPolicyRequest
	GetCiphers() []*string
	SetClientToken(v string) *CreateSecurityPolicyRequest
	GetClientToken() *string
	SetDryRun(v bool) *CreateSecurityPolicyRequest
	GetDryRun() *bool
	SetRegionId(v string) *CreateSecurityPolicyRequest
	GetRegionId() *string
	SetResourceGroupId(v string) *CreateSecurityPolicyRequest
	GetResourceGroupId() *string
	SetSecurityPolicyName(v string) *CreateSecurityPolicyRequest
	GetSecurityPolicyName() *string
	SetTag(v []*CreateSecurityPolicyRequestTag) *CreateSecurityPolicyRequest
	GetTag() []*CreateSecurityPolicyRequestTag
	SetTlsVersions(v []*string) *CreateSecurityPolicyRequest
	GetTlsVersions() []*string
}

type CreateSecurityPolicyRequest struct {
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
	//
	// This parameter is required.
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
	// The ID of the resource group to which the security policy belongs.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The name of the security policy.
	//
	// It must be 1 to 200 characters in length, and can contain letters, digits, periods (.), underscores (_), and hyphens (-).
	//
	// example:
	//
	// TLSCipherPolicy
	SecurityPolicyName *string `json:"SecurityPolicyName,omitempty" xml:"SecurityPolicyName,omitempty"`
	// The tags.
	//
	// if can be null:
	// true
	Tag []*CreateSecurityPolicyRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
	// The Transport Layer Security (TLS) versions supported by the security policy. Valid values: **TLSv1.0**, **TLSv1.1**, **TLSv1.2**, and **TLSv1.3**.
	//
	// This parameter is required.
	TlsVersions []*string `json:"TlsVersions,omitempty" xml:"TlsVersions,omitempty" type:"Repeated"`
}

func (s CreateSecurityPolicyRequest) String() string {
	return dara.Prettify(s)
}

func (s CreateSecurityPolicyRequest) GoString() string {
	return s.String()
}

func (s *CreateSecurityPolicyRequest) GetCiphers() []*string {
	return s.Ciphers
}

func (s *CreateSecurityPolicyRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *CreateSecurityPolicyRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *CreateSecurityPolicyRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *CreateSecurityPolicyRequest) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *CreateSecurityPolicyRequest) GetSecurityPolicyName() *string {
	return s.SecurityPolicyName
}

func (s *CreateSecurityPolicyRequest) GetTag() []*CreateSecurityPolicyRequestTag {
	return s.Tag
}

func (s *CreateSecurityPolicyRequest) GetTlsVersions() []*string {
	return s.TlsVersions
}

func (s *CreateSecurityPolicyRequest) SetCiphers(v []*string) *CreateSecurityPolicyRequest {
	s.Ciphers = v
	return s
}

func (s *CreateSecurityPolicyRequest) SetClientToken(v string) *CreateSecurityPolicyRequest {
	s.ClientToken = &v
	return s
}

func (s *CreateSecurityPolicyRequest) SetDryRun(v bool) *CreateSecurityPolicyRequest {
	s.DryRun = &v
	return s
}

func (s *CreateSecurityPolicyRequest) SetRegionId(v string) *CreateSecurityPolicyRequest {
	s.RegionId = &v
	return s
}

func (s *CreateSecurityPolicyRequest) SetResourceGroupId(v string) *CreateSecurityPolicyRequest {
	s.ResourceGroupId = &v
	return s
}

func (s *CreateSecurityPolicyRequest) SetSecurityPolicyName(v string) *CreateSecurityPolicyRequest {
	s.SecurityPolicyName = &v
	return s
}

func (s *CreateSecurityPolicyRequest) SetTag(v []*CreateSecurityPolicyRequestTag) *CreateSecurityPolicyRequest {
	s.Tag = v
	return s
}

func (s *CreateSecurityPolicyRequest) SetTlsVersions(v []*string) *CreateSecurityPolicyRequest {
	s.TlsVersions = v
	return s
}

func (s *CreateSecurityPolicyRequest) Validate() error {
	return dara.Validate(s)
}

type CreateSecurityPolicyRequestTag struct {
	// The key of the tag. It must be 1 to 64 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`. It can contain letters, digits, underscores (_), periods (.), colons (:), forward slashes (/), equal signs (=), plus signs (+), minus signs (-), and at signs (@).
	//
	// You can add up to 20 tags for the security policy in each call.
	//
	// example:
	//
	// KeyTest
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. It must be 1 to 128 characters in length, cannot start with `acs:` or `aliyun`, and cannot contain `http://` or `https://`. It can contain letters, digits, underscores (_), periods (.), colons (:), forward slashes (/), equal signs (=), plus signs (+), minus signs (-), and at signs (@).
	//
	// You can add up to 20 tags for the security policy in each call.
	//
	// example:
	//
	// ValueTest
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s CreateSecurityPolicyRequestTag) String() string {
	return dara.Prettify(s)
}

func (s CreateSecurityPolicyRequestTag) GoString() string {
	return s.String()
}

func (s *CreateSecurityPolicyRequestTag) GetKey() *string {
	return s.Key
}

func (s *CreateSecurityPolicyRequestTag) GetValue() *string {
	return s.Value
}

func (s *CreateSecurityPolicyRequestTag) SetKey(v string) *CreateSecurityPolicyRequestTag {
	s.Key = &v
	return s
}

func (s *CreateSecurityPolicyRequestTag) SetValue(v string) *CreateSecurityPolicyRequestTag {
	s.Value = &v
	return s
}

func (s *CreateSecurityPolicyRequestTag) Validate() error {
	return dara.Validate(s)
}
