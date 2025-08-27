// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListListenersResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetListeners(v []*ListListenersResponseBodyListeners) *ListListenersResponseBody
	GetListeners() []*ListListenersResponseBodyListeners
	SetMaxResults(v int32) *ListListenersResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *ListListenersResponseBody
	GetNextToken() *string
	SetRequestId(v string) *ListListenersResponseBody
	GetRequestId() *string
	SetTotalCount(v int32) *ListListenersResponseBody
	GetTotalCount() *int32
}

type ListListenersResponseBody struct {
	// The listeners.
	Listeners []*ListListenersResponseBodyListeners `json:"Listeners,omitempty" xml:"Listeners,omitempty" type:"Repeated"`
	// The number of entries returned per page.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The token that is used for the next query. Valid values:
	//
	// 	- If **NextToken*	- is empty, it indicates that no next query is to be sent.
	//
	// 	- If a value of **NextToken*	- is returned, the value is the token used for the next query.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// CEF72CEB-54B6-4AE8-B225-F876FF7BA984
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The number of entries returned.
	//
	// example:
	//
	// 4
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListListenersResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListListenersResponseBody) GoString() string {
	return s.String()
}

func (s *ListListenersResponseBody) GetListeners() []*ListListenersResponseBodyListeners {
	return s.Listeners
}

func (s *ListListenersResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListListenersResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *ListListenersResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListListenersResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *ListListenersResponseBody) SetListeners(v []*ListListenersResponseBodyListeners) *ListListenersResponseBody {
	s.Listeners = v
	return s
}

func (s *ListListenersResponseBody) SetMaxResults(v int32) *ListListenersResponseBody {
	s.MaxResults = &v
	return s
}

func (s *ListListenersResponseBody) SetNextToken(v string) *ListListenersResponseBody {
	s.NextToken = &v
	return s
}

func (s *ListListenersResponseBody) SetRequestId(v string) *ListListenersResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListListenersResponseBody) SetTotalCount(v int32) *ListListenersResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListListenersResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListListenersResponseBodyListeners struct {
	// Indicates whether Application-Layer Protocol Negotiation (ALPN) is enabled. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
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
	// A list of CA certificates.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	CaCertificateIds []*string `json:"CaCertificateIds,omitempty" xml:"CaCertificateIds,omitempty" type:"Repeated"`
	// Indicates whether mutual authentication is enabled. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	CaEnabled *bool `json:"CaEnabled,omitempty" xml:"CaEnabled,omitempty"`
	// The server certificate.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	CertificateIds []*string `json:"CertificateIds,omitempty" xml:"CertificateIds,omitempty" type:"Repeated"`
	// The maximum number of new connections per second supported by the listener in each zone (virtual IP address). Valid values: **0*	- to **1000000**. **0*	- indicates that the number of connections is unlimited.
	//
	// example:
	//
	// 1000
	Cps *int32 `json:"Cps,omitempty" xml:"Cps,omitempty"`
	// The last port in the listener port range.
	//
	// example:
	//
	// 455
	EndPort *string `json:"EndPort,omitempty" xml:"EndPort,omitempty"`
	// The timeout period of idle connections. Unit: seconds. Valid values: **1*	- to **900**. Default value: **900**.
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
	// The listener ID.
	//
	// example:
	//
	// lsn-ga6sjjcll6ou34l1et****
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The information about the listener port of your server.
	//
	// example:
	//
	// 443
	ListenerPort *int32 `json:"ListenerPort,omitempty" xml:"ListenerPort,omitempty"`
	// The listener protocol. Valid values: **TCP**, **UDP**, and **TCPSSL**.
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
	// The CLB instance ID.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The size of the largest TCP packet segment. Unit: bytes. Valid values: **0*	- to **1500**. **0*	- indicates that the Mss value of TCP packets remains unchanged.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	//
	// example:
	//
	// 200
	Mss *int32 `json:"Mss,omitempty" xml:"Mss,omitempty"`
	// Indicates whether the Proxy protocol passes source client IP addresses to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	ProxyProtocolEnabled *bool `json:"ProxyProtocolEnabled,omitempty" xml:"ProxyProtocolEnabled,omitempty"`
	// Indicates whether the Proxy protocol passes the VpcId, PrivateLinkEpId, and PrivateLinkEpsId parameters to backend servers.
	ProxyProtocolV2Config *ListListenersResponseBodyListenersProxyProtocolV2Config `json:"ProxyProtocolV2Config,omitempty" xml:"ProxyProtocolV2Config,omitempty" type:"Struct"`
	// The region ID of the NLB instance.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// Indicates whether fine-grained monitoring is enabled. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	SecSensorEnabled *bool `json:"SecSensorEnabled,omitempty" xml:"SecSensorEnabled,omitempty"`
	// The ID of the security policy.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	//
	// example:
	//
	// tls_cipher_policy_1_1
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
	// The server group ID.
	//
	// example:
	//
	// sgp-ppdpc14gdm3x4o****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The first port in the listener port range.
	//
	// example:
	//
	// 233
	StartPort *string `json:"StartPort,omitempty" xml:"StartPort,omitempty"`
	// A list of tags.
	Tags []*ListListenersResponseBodyListenersTags `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
}

func (s ListListenersResponseBodyListeners) String() string {
	return dara.Prettify(s)
}

func (s ListListenersResponseBodyListeners) GoString() string {
	return s.String()
}

func (s *ListListenersResponseBodyListeners) GetAlpnEnabled() *bool {
	return s.AlpnEnabled
}

func (s *ListListenersResponseBodyListeners) GetAlpnPolicy() *string {
	return s.AlpnPolicy
}

func (s *ListListenersResponseBodyListeners) GetCaCertificateIds() []*string {
	return s.CaCertificateIds
}

func (s *ListListenersResponseBodyListeners) GetCaEnabled() *bool {
	return s.CaEnabled
}

func (s *ListListenersResponseBodyListeners) GetCertificateIds() []*string {
	return s.CertificateIds
}

func (s *ListListenersResponseBodyListeners) GetCps() *int32 {
	return s.Cps
}

func (s *ListListenersResponseBodyListeners) GetEndPort() *string {
	return s.EndPort
}

func (s *ListListenersResponseBodyListeners) GetIdleTimeout() *int32 {
	return s.IdleTimeout
}

func (s *ListListenersResponseBodyListeners) GetListenerDescription() *string {
	return s.ListenerDescription
}

func (s *ListListenersResponseBodyListeners) GetListenerId() *string {
	return s.ListenerId
}

func (s *ListListenersResponseBodyListeners) GetListenerPort() *int32 {
	return s.ListenerPort
}

func (s *ListListenersResponseBodyListeners) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *ListListenersResponseBodyListeners) GetListenerStatus() *string {
	return s.ListenerStatus
}

func (s *ListListenersResponseBodyListeners) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *ListListenersResponseBodyListeners) GetMss() *int32 {
	return s.Mss
}

func (s *ListListenersResponseBodyListeners) GetProxyProtocolEnabled() *bool {
	return s.ProxyProtocolEnabled
}

func (s *ListListenersResponseBodyListeners) GetProxyProtocolV2Config() *ListListenersResponseBodyListenersProxyProtocolV2Config {
	return s.ProxyProtocolV2Config
}

func (s *ListListenersResponseBodyListeners) GetRegionId() *string {
	return s.RegionId
}

func (s *ListListenersResponseBodyListeners) GetSecSensorEnabled() *bool {
	return s.SecSensorEnabled
}

func (s *ListListenersResponseBodyListeners) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *ListListenersResponseBodyListeners) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *ListListenersResponseBodyListeners) GetStartPort() *string {
	return s.StartPort
}

func (s *ListListenersResponseBodyListeners) GetTags() []*ListListenersResponseBodyListenersTags {
	return s.Tags
}

func (s *ListListenersResponseBodyListeners) SetAlpnEnabled(v bool) *ListListenersResponseBodyListeners {
	s.AlpnEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetAlpnPolicy(v string) *ListListenersResponseBodyListeners {
	s.AlpnPolicy = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetCaCertificateIds(v []*string) *ListListenersResponseBodyListeners {
	s.CaCertificateIds = v
	return s
}

func (s *ListListenersResponseBodyListeners) SetCaEnabled(v bool) *ListListenersResponseBodyListeners {
	s.CaEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetCertificateIds(v []*string) *ListListenersResponseBodyListeners {
	s.CertificateIds = v
	return s
}

func (s *ListListenersResponseBodyListeners) SetCps(v int32) *ListListenersResponseBodyListeners {
	s.Cps = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetEndPort(v string) *ListListenersResponseBodyListeners {
	s.EndPort = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetIdleTimeout(v int32) *ListListenersResponseBodyListeners {
	s.IdleTimeout = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetListenerDescription(v string) *ListListenersResponseBodyListeners {
	s.ListenerDescription = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetListenerId(v string) *ListListenersResponseBodyListeners {
	s.ListenerId = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetListenerPort(v int32) *ListListenersResponseBodyListeners {
	s.ListenerPort = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetListenerProtocol(v string) *ListListenersResponseBodyListeners {
	s.ListenerProtocol = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetListenerStatus(v string) *ListListenersResponseBodyListeners {
	s.ListenerStatus = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetLoadBalancerId(v string) *ListListenersResponseBodyListeners {
	s.LoadBalancerId = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetMss(v int32) *ListListenersResponseBodyListeners {
	s.Mss = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetProxyProtocolEnabled(v bool) *ListListenersResponseBodyListeners {
	s.ProxyProtocolEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetProxyProtocolV2Config(v *ListListenersResponseBodyListenersProxyProtocolV2Config) *ListListenersResponseBodyListeners {
	s.ProxyProtocolV2Config = v
	return s
}

func (s *ListListenersResponseBodyListeners) SetRegionId(v string) *ListListenersResponseBodyListeners {
	s.RegionId = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetSecSensorEnabled(v bool) *ListListenersResponseBodyListeners {
	s.SecSensorEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetSecurityPolicyId(v string) *ListListenersResponseBodyListeners {
	s.SecurityPolicyId = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetServerGroupId(v string) *ListListenersResponseBodyListeners {
	s.ServerGroupId = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetStartPort(v string) *ListListenersResponseBodyListeners {
	s.StartPort = &v
	return s
}

func (s *ListListenersResponseBodyListeners) SetTags(v []*ListListenersResponseBodyListenersTags) *ListListenersResponseBodyListeners {
	s.Tags = v
	return s
}

func (s *ListListenersResponseBodyListeners) Validate() error {
	return dara.Validate(s)
}

type ListListenersResponseBodyListenersProxyProtocolV2Config struct {
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

func (s ListListenersResponseBodyListenersProxyProtocolV2Config) String() string {
	return dara.Prettify(s)
}

func (s ListListenersResponseBodyListenersProxyProtocolV2Config) GoString() string {
	return s.String()
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) GetPpv2PrivateLinkEpIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpIdEnabled
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) GetPpv2PrivateLinkEpsIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpsIdEnabled
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) GetPpv2VpcIdEnabled() *bool {
	return s.Ppv2VpcIdEnabled
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) SetPpv2PrivateLinkEpIdEnabled(v bool) *ListListenersResponseBodyListenersProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpIdEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) SetPpv2PrivateLinkEpsIdEnabled(v bool) *ListListenersResponseBodyListenersProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpsIdEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) SetPpv2VpcIdEnabled(v bool) *ListListenersResponseBodyListenersProxyProtocolV2Config {
	s.Ppv2VpcIdEnabled = &v
	return s
}

func (s *ListListenersResponseBodyListenersProxyProtocolV2Config) Validate() error {
	return dara.Validate(s)
}

type ListListenersResponseBodyListenersTags struct {
	// The tag key.
	//
	// example:
	//
	// Created
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The tag value.
	//
	// example:
	//
	// TF
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListListenersResponseBodyListenersTags) String() string {
	return dara.Prettify(s)
}

func (s ListListenersResponseBodyListenersTags) GoString() string {
	return s.String()
}

func (s *ListListenersResponseBodyListenersTags) GetKey() *string {
	return s.Key
}

func (s *ListListenersResponseBodyListenersTags) GetValue() *string {
	return s.Value
}

func (s *ListListenersResponseBodyListenersTags) SetKey(v string) *ListListenersResponseBodyListenersTags {
	s.Key = &v
	return s
}

func (s *ListListenersResponseBodyListenersTags) SetValue(v string) *ListListenersResponseBodyListenersTags {
	s.Value = &v
	return s
}

func (s *ListListenersResponseBodyListenersTags) Validate() error {
	return dara.Validate(s)
}
