// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListSecurityPolicyResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetMaxResults(v int32) *ListSecurityPolicyResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *ListSecurityPolicyResponseBody
	GetNextToken() *string
	SetRequestId(v string) *ListSecurityPolicyResponseBody
	GetRequestId() *string
	SetSecurityPolicies(v []*ListSecurityPolicyResponseBodySecurityPolicies) *ListSecurityPolicyResponseBody
	GetSecurityPolicies() []*ListSecurityPolicyResponseBodySecurityPolicies
	SetTotalCount(v int32) *ListSecurityPolicyResponseBody
	GetTotalCount() *int32
}

type ListSecurityPolicyResponseBody struct {
	// The number of entries per page.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// A pagination token. It can be used in the next request to retrieve a new page of results. Valid values:
	//
	// 	- If NextToken is empty, no next page exists.
	//
	// 	- If a value is returned for NextToken, specify the value in the next request to retrieve a new page of results.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The request ID.
	//
	// example:
	//
	// D7A8875F-373A-5F48-8484-25B07A61F2AF
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The TLS security policies.
	SecurityPolicies []*ListSecurityPolicyResponseBodySecurityPolicies `json:"SecurityPolicies,omitempty" xml:"SecurityPolicies,omitempty" type:"Repeated"`
	// The total number of entries returned.
	//
	// example:
	//
	// 10
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListSecurityPolicyResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyResponseBody) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListSecurityPolicyResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *ListSecurityPolicyResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListSecurityPolicyResponseBody) GetSecurityPolicies() []*ListSecurityPolicyResponseBodySecurityPolicies {
	return s.SecurityPolicies
}

func (s *ListSecurityPolicyResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *ListSecurityPolicyResponseBody) SetMaxResults(v int32) *ListSecurityPolicyResponseBody {
	s.MaxResults = &v
	return s
}

func (s *ListSecurityPolicyResponseBody) SetNextToken(v string) *ListSecurityPolicyResponseBody {
	s.NextToken = &v
	return s
}

func (s *ListSecurityPolicyResponseBody) SetRequestId(v string) *ListSecurityPolicyResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListSecurityPolicyResponseBody) SetSecurityPolicies(v []*ListSecurityPolicyResponseBodySecurityPolicies) *ListSecurityPolicyResponseBody {
	s.SecurityPolicies = v
	return s
}

func (s *ListSecurityPolicyResponseBody) SetTotalCount(v int32) *ListSecurityPolicyResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListSecurityPolicyResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListSecurityPolicyResponseBodySecurityPolicies struct {
	// The cipher suites supported by the security policy. Valid values of this parameter vary based on the value of TlsVersions. A security policy supports up to 32 cipher suites.
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
	// example:
	//
	// ECDHE-ECDSA-AES128-SHA
	Ciphers *string `json:"Ciphers,omitempty" xml:"Ciphers,omitempty"`
	// The region ID of the NLB instance.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The listeners that are associated with the NLB instance.
	RelatedListeners []*ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners `json:"RelatedListeners,omitempty" xml:"RelatedListeners,omitempty" type:"Repeated"`
	// The resource group ID.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The ID of the TLS security policy.
	//
	// example:
	//
	// tls-bp14bb1e7dll4f****
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
	// The name of the TLS security policy.
	//
	// example:
	//
	// TLSCipherPolicy
	SecurityPolicyName *string `json:"SecurityPolicyName,omitempty" xml:"SecurityPolicyName,omitempty"`
	// The status of the TLS security policy. Valid values:
	//
	// 	- **Configuring**
	//
	// 	- **Available**
	//
	// example:
	//
	// Available
	SecurityPolicyStatus *string `json:"SecurityPolicyStatus,omitempty" xml:"SecurityPolicyStatus,omitempty"`
	// The tags.
	Tags []*ListSecurityPolicyResponseBodySecurityPoliciesTags `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
	// The supported versions of the TLS protocol. Valid values: **TLSv1.0**, **TLSv1.1**, **TLSv1.2**, and **TLSv1.3**.
	//
	// example:
	//
	// TLSv1.0
	TlsVersion *string `json:"TlsVersion,omitempty" xml:"TlsVersion,omitempty"`
}

func (s ListSecurityPolicyResponseBodySecurityPolicies) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyResponseBodySecurityPolicies) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetCiphers() *string {
	return s.Ciphers
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetRegionId() *string {
	return s.RegionId
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetRelatedListeners() []*ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners {
	return s.RelatedListeners
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetSecurityPolicyName() *string {
	return s.SecurityPolicyName
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetSecurityPolicyStatus() *string {
	return s.SecurityPolicyStatus
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetTags() []*ListSecurityPolicyResponseBodySecurityPoliciesTags {
	return s.Tags
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) GetTlsVersion() *string {
	return s.TlsVersion
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetCiphers(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.Ciphers = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetRegionId(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.RegionId = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetRelatedListeners(v []*ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.RelatedListeners = v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetResourceGroupId(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.ResourceGroupId = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetSecurityPolicyId(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.SecurityPolicyId = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetSecurityPolicyName(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.SecurityPolicyName = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetSecurityPolicyStatus(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.SecurityPolicyStatus = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetTags(v []*ListSecurityPolicyResponseBodySecurityPoliciesTags) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.Tags = v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) SetTlsVersion(v string) *ListSecurityPolicyResponseBodySecurityPolicies {
	s.TlsVersion = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPolicies) Validate() error {
	return dara.Validate(s)
}

type ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners struct {
	// The listener ID.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@443
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The listener port.
	//
	// example:
	//
	// 443
	ListenerPort *int64 `json:"ListenerPort,omitempty" xml:"ListenerPort,omitempty"`
	// The listener protocol. Valid value: **TCPSSL**.
	//
	// example:
	//
	// TCPSSL
	ListenerProtocol *string `json:"ListenerProtocol,omitempty" xml:"ListenerProtocol,omitempty"`
	// The NLB instance ID.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
}

func (s ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) GetListenerId() *string {
	return s.ListenerId
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) GetListenerPort() *int64 {
	return s.ListenerPort
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) SetListenerId(v string) *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners {
	s.ListenerId = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) SetListenerPort(v int64) *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners {
	s.ListenerPort = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) SetListenerProtocol(v string) *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners {
	s.ListenerProtocol = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) SetLoadBalancerId(v string) *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners {
	s.LoadBalancerId = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesRelatedListeners) Validate() error {
	return dara.Validate(s)
}

type ListSecurityPolicyResponseBodySecurityPoliciesTags struct {
	// The tag key. You can specify up to 10 tag keys.
	//
	// The tag key can be up to 64 characters in length, and cannot contain `http://` or `https://`. It cannot start with `aliyun` or `acs:`.
	//
	// example:
	//
	// Test
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The tag value. You can specify up to 10 tag values.
	//
	// The tag value can be up to 128 characters in length, and cannot contain `http://` or `https://`. It cannot start with `aliyun` or `acs:`.
	//
	// example:
	//
	// Test
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListSecurityPolicyResponseBodySecurityPoliciesTags) String() string {
	return dara.Prettify(s)
}

func (s ListSecurityPolicyResponseBodySecurityPoliciesTags) GoString() string {
	return s.String()
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesTags) GetKey() *string {
	return s.Key
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesTags) GetValue() *string {
	return s.Value
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesTags) SetKey(v string) *ListSecurityPolicyResponseBodySecurityPoliciesTags {
	s.Key = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesTags) SetValue(v string) *ListSecurityPolicyResponseBodySecurityPoliciesTags {
	s.Value = &v
	return s
}

func (s *ListSecurityPolicyResponseBodySecurityPoliciesTags) Validate() error {
	return dara.Validate(s)
}
