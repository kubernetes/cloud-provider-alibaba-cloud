// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetListenerAttributeResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetAlpnEnabled(v bool) *GetListenerAttributeResponseBody
	GetAlpnEnabled() *bool
	SetAlpnPolicy(v string) *GetListenerAttributeResponseBody
	GetAlpnPolicy() *string
	SetCaCertificateIds(v []*string) *GetListenerAttributeResponseBody
	GetCaCertificateIds() []*string
	SetCaEnabled(v bool) *GetListenerAttributeResponseBody
	GetCaEnabled() *bool
	SetCertificateIds(v []*string) *GetListenerAttributeResponseBody
	GetCertificateIds() []*string
	SetCps(v int32) *GetListenerAttributeResponseBody
	GetCps() *int32
	SetEndPort(v string) *GetListenerAttributeResponseBody
	GetEndPort() *string
	SetIdleTimeout(v int32) *GetListenerAttributeResponseBody
	GetIdleTimeout() *int32
	SetListenerDescription(v string) *GetListenerAttributeResponseBody
	GetListenerDescription() *string
	SetListenerId(v string) *GetListenerAttributeResponseBody
	GetListenerId() *string
	SetListenerPort(v int32) *GetListenerAttributeResponseBody
	GetListenerPort() *int32
	SetListenerProtocol(v string) *GetListenerAttributeResponseBody
	GetListenerProtocol() *string
	SetListenerStatus(v string) *GetListenerAttributeResponseBody
	GetListenerStatus() *string
	SetLoadBalancerId(v string) *GetListenerAttributeResponseBody
	GetLoadBalancerId() *string
	SetMss(v int32) *GetListenerAttributeResponseBody
	GetMss() *int32
	SetProxyProtocolEnabled(v bool) *GetListenerAttributeResponseBody
	GetProxyProtocolEnabled() *bool
	SetProxyProtocolV2Config(v *GetListenerAttributeResponseBodyProxyProtocolV2Config) *GetListenerAttributeResponseBody
	GetProxyProtocolV2Config() *GetListenerAttributeResponseBodyProxyProtocolV2Config
	SetRegionId(v string) *GetListenerAttributeResponseBody
	GetRegionId() *string
	SetRequestId(v string) *GetListenerAttributeResponseBody
	GetRequestId() *string
	SetSecSensorEnabled(v bool) *GetListenerAttributeResponseBody
	GetSecSensorEnabled() *bool
	SetSecurityPolicyId(v string) *GetListenerAttributeResponseBody
	GetSecurityPolicyId() *string
	SetServerGroupId(v string) *GetListenerAttributeResponseBody
	GetServerGroupId() *string
	SetStartPort(v string) *GetListenerAttributeResponseBody
	GetStartPort() *string
	SetTags(v []*GetListenerAttributeResponseBodyTags) *GetListenerAttributeResponseBody
	GetTags() []*GetListenerAttributeResponseBodyTags
}

type GetListenerAttributeResponseBody struct {
	// Indicates whether Application-Layer Protocol Negotiation (ALPN) is enabled. Valid values:
	//
	// 	- **true**: yes
	//
	// 	- **false**: no
	//
	// example:
	//
	// false
	AlpnEnabled *bool `json:"AlpnEnabled,omitempty" xml:"AlpnEnabled,omitempty"`
	// The ALPN policy. Valid values:
	//
	// 	- **HTTP1Only**
	//
	// 	- **HTTP2Only**
	//
	// 	- **HTTP2Preferred**
	//
	// 	- **HTTP2Optional**
	//
	// example:
	//
	// HTTP1Only
	AlpnPolicy *string `json:"AlpnPolicy,omitempty" xml:"AlpnPolicy,omitempty"`
	// The CA certificates. Only one CA certificate is supported.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	CaCertificateIds []*string `json:"CaCertificateIds,omitempty" xml:"CaCertificateIds,omitempty" type:"Repeated"`
	// Indicates whether mutual authentication is enabled. Valid values:
	//
	// 	- **true**: yes
	//
	// 	- **false**: no
	//
	// example:
	//
	// false
	CaEnabled *bool `json:"CaEnabled,omitempty" xml:"CaEnabled,omitempty"`
	// The server certificates. Only one server certificate is supported.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	CertificateIds []*string `json:"CertificateIds,omitempty" xml:"CertificateIds,omitempty" type:"Repeated"`
	// The maximum number of new connections per second supported by the listener in each zone (virtual IP address). Valid values: **0*	- to **1000000**. **0*	- indicates that the number of connections is unlimited.
	//
	// example:
	//
	// 1000
	Cps *int32 `json:"Cps,omitempty" xml:"Cps,omitempty"`
	// The last port in the listening port range. Valid values: **0*	- to **65535**. The number of the last port must be smaller than that of the first port.
	//
	// example:
	//
	// 455
	EndPort *string `json:"EndPort,omitempty" xml:"EndPort,omitempty"`
	// The timeout period of an idle connection. Unit: seconds. Valid values: **1*	- to **900**.
	//
	// example:
	//
	// 900
	IdleTimeout *int32 `json:"IdleTimeout,omitempty" xml:"IdleTimeout,omitempty"`
	// The name of the listener.
	//
	// The name must be 2 to 256 characters in length, and can contain letters, digits, commas (,), periods (.), semicolons (;), forward slashes (/), at signs (@), underscores (_), and hyphens (-).
	//
	// example:
	//
	// tcpssl_443
	ListenerDescription *string `json:"ListenerDescription,omitempty" xml:"ListenerDescription,omitempty"`
	// The ID of the listener.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@233
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The listening port. Valid values: **0*	- to **65535**. A value of **0*	- specifies all ports. If you set the value to **0**, you must also set the **StartPort*	- and **EndPort*	- parameters.
	//
	// example:
	//
	// 233
	ListenerPort *int32 `json:"ListenerPort,omitempty" xml:"ListenerPort,omitempty"`
	// The listening protocol. Valid values: **TCP**, **UDP**, and **TCPSSL**.
	//
	// example:
	//
	// TCPSSL
	ListenerProtocol *string `json:"ListenerProtocol,omitempty" xml:"ListenerProtocol,omitempty"`
	// The status of the listener. Valid values:
	//
	// 	- **Provisioning**: The listener is being created.
	//
	// 	- **Running**: The listener is running.
	//
	// 	- **Configuring**: The listener is being configured.
	//
	// 	- **Stopping**: The listener is being stopped.
	//
	// 	- **Stopped**: The listener is stopped.
	//
	// 	- **Starting**: The listener is being started.
	//
	// 	- **Deleting**: The listener is being deleted.
	//
	// 	- **Deleted**: The listener is deleted.
	//
	// example:
	//
	// Running
	ListenerStatus *string `json:"ListenerStatus,omitempty" xml:"ListenerStatus,omitempty"`
	// The ID of the NLB instance.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The size of the largest TCP segment. Unit: bytes. Valid values: **0*	- to **1500**. **0*	- specifies that the maximum segment size remains unchanged.
	//
	// >  This parameter is supported only by listeners that use SSL over TCP.
	//
	// example:
	//
	// 166
	Mss *int32 `json:"Mss,omitempty" xml:"Mss,omitempty"`
	// Indicates whether the Proxy protocol is used to pass client IP addresses to backend servers. Valid values:
	//
	// 	- **true**: yes
	//
	// 	- **false**: no
	//
	// example:
	//
	// false
	ProxyProtocolEnabled *bool `json:"ProxyProtocolEnabled,omitempty" xml:"ProxyProtocolEnabled,omitempty"`
	// Indicates whether the Proxy protocol passes the VpcId, PrivateLinkEpId, and PrivateLinkEpsId parameters to backend servers.
	ProxyProtocolV2Config *GetListenerAttributeResponseBodyProxyProtocolV2Config `json:"ProxyProtocolV2Config,omitempty" xml:"ProxyProtocolV2Config,omitempty" type:"Struct"`
	// The ID of the region where the NLB instance is deployed.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// Indicates whether fine-grained monitoring is enabled. Valid values:
	//
	// 	- **true**: yes
	//
	// 	- **false**: no
	//
	// example:
	//
	// false
	SecSensorEnabled *bool `json:"SecSensorEnabled,omitempty" xml:"SecSensorEnabled,omitempty"`
	// The ID of the security policy. System security policies and custom security policies are supported.
	//
	// - Valid values: **tls_cipher_policy_1_0**, **tls_cipher_policy_1_1**, **tls_cipher_policy_1_2**, **tls_cipher_policy_1_2_strict**, and **tls_cipher_policy_1_2_strict_with_1_3**.
	//
	// - Custom security policy: the ID of the custom security policy.
	//
	//     - For more information about how to create a custom security policy, see [CreateSecurityPolicy](https://help.aliyun.com/document_detail/2399231.html) .
	//
	//     - For more information about how to query security policies, see [ListSecurityPolicy](https://help.aliyun.com/document_detail/2399234.html) .
	//
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	//
	// example:
	//
	// tls_cipher_policy_1_0
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
	// The ID of the server group.
	//
	// example:
	//
	// sgp-ppdpc14gdm3x4o****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The first port in the listening port range. Valid values: **0*	- to **65535**.
	//
	// example:
	//
	// 233
	StartPort *string `json:"StartPort,omitempty" xml:"StartPort,omitempty"`
	// The tags.
	Tags []*GetListenerAttributeResponseBodyTags `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
}

func (s GetListenerAttributeResponseBody) String() string {
	return dara.Prettify(s)
}

func (s GetListenerAttributeResponseBody) GoString() string {
	return s.String()
}

func (s *GetListenerAttributeResponseBody) GetAlpnEnabled() *bool {
	return s.AlpnEnabled
}

func (s *GetListenerAttributeResponseBody) GetAlpnPolicy() *string {
	return s.AlpnPolicy
}

func (s *GetListenerAttributeResponseBody) GetCaCertificateIds() []*string {
	return s.CaCertificateIds
}

func (s *GetListenerAttributeResponseBody) GetCaEnabled() *bool {
	return s.CaEnabled
}

func (s *GetListenerAttributeResponseBody) GetCertificateIds() []*string {
	return s.CertificateIds
}

func (s *GetListenerAttributeResponseBody) GetCps() *int32 {
	return s.Cps
}

func (s *GetListenerAttributeResponseBody) GetEndPort() *string {
	return s.EndPort
}

func (s *GetListenerAttributeResponseBody) GetIdleTimeout() *int32 {
	return s.IdleTimeout
}

func (s *GetListenerAttributeResponseBody) GetListenerDescription() *string {
	return s.ListenerDescription
}

func (s *GetListenerAttributeResponseBody) GetListenerId() *string {
	return s.ListenerId
}

func (s *GetListenerAttributeResponseBody) GetListenerPort() *int32 {
	return s.ListenerPort
}

func (s *GetListenerAttributeResponseBody) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *GetListenerAttributeResponseBody) GetListenerStatus() *string {
	return s.ListenerStatus
}

func (s *GetListenerAttributeResponseBody) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *GetListenerAttributeResponseBody) GetMss() *int32 {
	return s.Mss
}

func (s *GetListenerAttributeResponseBody) GetProxyProtocolEnabled() *bool {
	return s.ProxyProtocolEnabled
}

func (s *GetListenerAttributeResponseBody) GetProxyProtocolV2Config() *GetListenerAttributeResponseBodyProxyProtocolV2Config {
	return s.ProxyProtocolV2Config
}

func (s *GetListenerAttributeResponseBody) GetRegionId() *string {
	return s.RegionId
}

func (s *GetListenerAttributeResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *GetListenerAttributeResponseBody) GetSecSensorEnabled() *bool {
	return s.SecSensorEnabled
}

func (s *GetListenerAttributeResponseBody) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *GetListenerAttributeResponseBody) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *GetListenerAttributeResponseBody) GetStartPort() *string {
	return s.StartPort
}

func (s *GetListenerAttributeResponseBody) GetTags() []*GetListenerAttributeResponseBodyTags {
	return s.Tags
}

func (s *GetListenerAttributeResponseBody) SetAlpnEnabled(v bool) *GetListenerAttributeResponseBody {
	s.AlpnEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetAlpnPolicy(v string) *GetListenerAttributeResponseBody {
	s.AlpnPolicy = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetCaCertificateIds(v []*string) *GetListenerAttributeResponseBody {
	s.CaCertificateIds = v
	return s
}

func (s *GetListenerAttributeResponseBody) SetCaEnabled(v bool) *GetListenerAttributeResponseBody {
	s.CaEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetCertificateIds(v []*string) *GetListenerAttributeResponseBody {
	s.CertificateIds = v
	return s
}

func (s *GetListenerAttributeResponseBody) SetCps(v int32) *GetListenerAttributeResponseBody {
	s.Cps = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetEndPort(v string) *GetListenerAttributeResponseBody {
	s.EndPort = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetIdleTimeout(v int32) *GetListenerAttributeResponseBody {
	s.IdleTimeout = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetListenerDescription(v string) *GetListenerAttributeResponseBody {
	s.ListenerDescription = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetListenerId(v string) *GetListenerAttributeResponseBody {
	s.ListenerId = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetListenerPort(v int32) *GetListenerAttributeResponseBody {
	s.ListenerPort = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetListenerProtocol(v string) *GetListenerAttributeResponseBody {
	s.ListenerProtocol = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetListenerStatus(v string) *GetListenerAttributeResponseBody {
	s.ListenerStatus = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetLoadBalancerId(v string) *GetListenerAttributeResponseBody {
	s.LoadBalancerId = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetMss(v int32) *GetListenerAttributeResponseBody {
	s.Mss = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetProxyProtocolEnabled(v bool) *GetListenerAttributeResponseBody {
	s.ProxyProtocolEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetProxyProtocolV2Config(v *GetListenerAttributeResponseBodyProxyProtocolV2Config) *GetListenerAttributeResponseBody {
	s.ProxyProtocolV2Config = v
	return s
}

func (s *GetListenerAttributeResponseBody) SetRegionId(v string) *GetListenerAttributeResponseBody {
	s.RegionId = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetRequestId(v string) *GetListenerAttributeResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetSecSensorEnabled(v bool) *GetListenerAttributeResponseBody {
	s.SecSensorEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetSecurityPolicyId(v string) *GetListenerAttributeResponseBody {
	s.SecurityPolicyId = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetServerGroupId(v string) *GetListenerAttributeResponseBody {
	s.ServerGroupId = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetStartPort(v string) *GetListenerAttributeResponseBody {
	s.StartPort = &v
	return s
}

func (s *GetListenerAttributeResponseBody) SetTags(v []*GetListenerAttributeResponseBodyTags) *GetListenerAttributeResponseBody {
	s.Tags = v
	return s
}

func (s *GetListenerAttributeResponseBody) Validate() error {
	return dara.Validate(s)
}

type GetListenerAttributeResponseBodyProxyProtocolV2Config struct {
	// Indicates whether the Proxy protocol passes the PrivateLinkEpId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	Ppv2PrivateLinkEpIdEnabled *bool `json:"Ppv2PrivateLinkEpIdEnabled,omitempty" xml:"Ppv2PrivateLinkEpIdEnabled,omitempty"`
	// Indicates whether the Proxy protocol passes the PrivateLinkEpsId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	Ppv2PrivateLinkEpsIdEnabled *bool `json:"Ppv2PrivateLinkEpsIdEnabled,omitempty" xml:"Ppv2PrivateLinkEpsIdEnabled,omitempty"`
	// Indicates whether the Proxy protocol passes the VpcId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	Ppv2VpcIdEnabled *bool `json:"Ppv2VpcIdEnabled,omitempty" xml:"Ppv2VpcIdEnabled,omitempty"`
}

func (s GetListenerAttributeResponseBodyProxyProtocolV2Config) String() string {
	return dara.Prettify(s)
}

func (s GetListenerAttributeResponseBodyProxyProtocolV2Config) GoString() string {
	return s.String()
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) GetPpv2PrivateLinkEpIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpIdEnabled
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) GetPpv2PrivateLinkEpsIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpsIdEnabled
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) GetPpv2VpcIdEnabled() *bool {
	return s.Ppv2VpcIdEnabled
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) SetPpv2PrivateLinkEpIdEnabled(v bool) *GetListenerAttributeResponseBodyProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpIdEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) SetPpv2PrivateLinkEpsIdEnabled(v bool) *GetListenerAttributeResponseBodyProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpsIdEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) SetPpv2VpcIdEnabled(v bool) *GetListenerAttributeResponseBodyProxyProtocolV2Config {
	s.Ppv2VpcIdEnabled = &v
	return s
}

func (s *GetListenerAttributeResponseBodyProxyProtocolV2Config) Validate() error {
	return dara.Validate(s)
}

type GetListenerAttributeResponseBodyTags struct {
	// The tag key.
	//
	// example:
	//
	// ac-cus-tag-4
	TagKey *string `json:"TagKey,omitempty" xml:"TagKey,omitempty"`
	// The tag value.
	//
	// example:
	//
	// ON
	TagValue *string `json:"TagValue,omitempty" xml:"TagValue,omitempty"`
}

func (s GetListenerAttributeResponseBodyTags) String() string {
	return dara.Prettify(s)
}

func (s GetListenerAttributeResponseBodyTags) GoString() string {
	return s.String()
}

func (s *GetListenerAttributeResponseBodyTags) GetTagKey() *string {
	return s.TagKey
}

func (s *GetListenerAttributeResponseBodyTags) GetTagValue() *string {
	return s.TagValue
}

func (s *GetListenerAttributeResponseBodyTags) SetTagKey(v string) *GetListenerAttributeResponseBodyTags {
	s.TagKey = &v
	return s
}

func (s *GetListenerAttributeResponseBodyTags) SetTagValue(v string) *GetListenerAttributeResponseBodyTags {
	s.TagValue = &v
	return s
}

func (s *GetListenerAttributeResponseBodyTags) Validate() error {
	return dara.Validate(s)
}
