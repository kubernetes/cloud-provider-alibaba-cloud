// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateListenerAttributeRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAlpnEnabled(v bool) *UpdateListenerAttributeRequest
	GetAlpnEnabled() *bool
	SetAlpnPolicy(v string) *UpdateListenerAttributeRequest
	GetAlpnPolicy() *string
	SetCaCertificateIds(v []*string) *UpdateListenerAttributeRequest
	GetCaCertificateIds() []*string
	SetCaEnabled(v bool) *UpdateListenerAttributeRequest
	GetCaEnabled() *bool
	SetCertificateIds(v []*string) *UpdateListenerAttributeRequest
	GetCertificateIds() []*string
	SetClientToken(v string) *UpdateListenerAttributeRequest
	GetClientToken() *string
	SetCps(v int32) *UpdateListenerAttributeRequest
	GetCps() *int32
	SetDryRun(v bool) *UpdateListenerAttributeRequest
	GetDryRun() *bool
	SetIdleTimeout(v int32) *UpdateListenerAttributeRequest
	GetIdleTimeout() *int32
	SetListenerDescription(v string) *UpdateListenerAttributeRequest
	GetListenerDescription() *string
	SetListenerId(v string) *UpdateListenerAttributeRequest
	GetListenerId() *string
	SetMss(v int32) *UpdateListenerAttributeRequest
	GetMss() *int32
	SetProxyProtocolEnabled(v bool) *UpdateListenerAttributeRequest
	GetProxyProtocolEnabled() *bool
	SetProxyProtocolV2Config(v *UpdateListenerAttributeRequestProxyProtocolV2Config) *UpdateListenerAttributeRequest
	GetProxyProtocolV2Config() *UpdateListenerAttributeRequestProxyProtocolV2Config
	SetRegionId(v string) *UpdateListenerAttributeRequest
	GetRegionId() *string
	SetSecSensorEnabled(v bool) *UpdateListenerAttributeRequest
	GetSecSensorEnabled() *bool
	SetSecurityPolicyId(v string) *UpdateListenerAttributeRequest
	GetSecurityPolicyId() *string
	SetServerGroupId(v string) *UpdateListenerAttributeRequest
	GetServerGroupId() *string
}

type UpdateListenerAttributeRequest struct {
	// Specifies whether to enable Application-Layer Protocol Negotiation (ALPN). Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	AlpnEnabled *bool `json:"AlpnEnabled,omitempty" xml:"AlpnEnabled,omitempty"`
	// The name of the ALPN policy. The following are the possible values:
	//
	// 	- **HTTP1Only**: Negotiate only HTTP/1.\\*. The ALPN preference list is HTTP/1.1, HTTP/1.0.
	//
	// 	- **HTTP2Only**: Negotiate only HTTP/2. The ALPN preference list is HTTP/2.
	//
	// 	- **HTTP2Optional**: Prefer HTTP/1.\\	- over HTTP/2. The ALPN preference list is HTTP/1.1, HTTP/1.0, HTTP/2.
	//
	// 	- **HTTP2Preferred**: Prefer HTTP/2 over HTTP/1.\\*. The ALPN preference list is HTTP/2, HTTP/1.1, HTTP/1.0.
	//
	// >  This parameter is required if AlpnEnabled is set to true.
	//
	// if can be null:
	// true
	//
	// example:
	//
	// HTTP1Only
	AlpnPolicy *string `json:"AlpnPolicy,omitempty" xml:"AlpnPolicy,omitempty"`
	// The CA certificate. You can specify only one CA certificate.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	CaCertificateIds []*string `json:"CaCertificateIds,omitempty" xml:"CaCertificateIds,omitempty" type:"Repeated"`
	// Specifies whether to enable mutual authentication. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	CaEnabled *bool `json:"CaEnabled,omitempty" xml:"CaEnabled,omitempty"`
	// The server certificate. Only one server certificate is supported.
	//
	// >  This parameter takes effect only for listeners that use SSL over TCP.
	CertificateIds []*string `json:"CertificateIds,omitempty" xml:"CertificateIds,omitempty" type:"Repeated"`
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
	// The maximum number of new connections per second supported by the listener in each zone (virtual IP address). Valid values: **0*	- to **1000000**. **0*	- indicates that the number of connections is unlimited.
	//
	// example:
	//
	// 10000
	Cps *int32 `json:"Cps,omitempty" xml:"Cps,omitempty"`
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
	// The timeout period for idle connections. Unit: seconds
	//
	// 	- If the listener uses **TCP*	- or **TCPSSL**, you can set this parameter to a value ranging from **10*	- to **900**. Default value: **900**
	//
	// 	- If the listener uses **UDP**, you can set this parameter to a value ranging from **10*	- to **20**. Default value: **20**
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
	// This parameter is required.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@443
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The size of the largest TCP packet segment. Unit: bytes. Valid values: **0*	- to **1500**. **0*	- indicates that the maximum segment size (MSS) remains unchanged. This parameter is supported only by TCP listeners and listeners that use SSL over TCP.
	//
	// example:
	//
	// 344
	Mss *int32 `json:"Mss,omitempty" xml:"Mss,omitempty"`
	// Specifies whether to use the Proxy protocol to pass the client IP address to the backend server. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	ProxyProtocolEnabled *bool `json:"ProxyProtocolEnabled,omitempty" xml:"ProxyProtocolEnabled,omitempty"`
	// Specifies that the Proxy protocol passes the VpcId, PrivateLinkEpId, and PrivateLinkEpsId parameters to backend servers.
	ProxyProtocolV2Config *UpdateListenerAttributeRequestProxyProtocolV2Config `json:"ProxyProtocolV2Config,omitempty" xml:"ProxyProtocolV2Config,omitempty" type:"Struct"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// Specifies whether to enable fine-grained monitoring. Valid values:
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
	// >
	//
	// 	- If the listener uses **TCP**, you can specify server groups whose protocol is **TCP*	- or **TCP_UDP**. **UDP*	- server groups are not supported.
	//
	// 	- If the listener uses **UDP**, you can specify server groups whose protocol is **UDP*	- or **TCP_UDP**. **TCP*	- server groups are not supported.
	//
	// 	- If the listener uses **TCPSSL**, you can specify server groups whose protocol is **TCP*	- and whose **client IP preservation is disabled**. **TCP*	- server groups for which **client IP preservation is enabled*	- and server groups whose protocol is **UDP*	- or **TCP_UDP*	- are not supported.
	//
	// example:
	//
	// sgp-ppdpc14gdm3x4o****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
}

func (s UpdateListenerAttributeRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateListenerAttributeRequest) GoString() string {
	return s.String()
}

func (s *UpdateListenerAttributeRequest) GetAlpnEnabled() *bool {
	return s.AlpnEnabled
}

func (s *UpdateListenerAttributeRequest) GetAlpnPolicy() *string {
	return s.AlpnPolicy
}

func (s *UpdateListenerAttributeRequest) GetCaCertificateIds() []*string {
	return s.CaCertificateIds
}

func (s *UpdateListenerAttributeRequest) GetCaEnabled() *bool {
	return s.CaEnabled
}

func (s *UpdateListenerAttributeRequest) GetCertificateIds() []*string {
	return s.CertificateIds
}

func (s *UpdateListenerAttributeRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateListenerAttributeRequest) GetCps() *int32 {
	return s.Cps
}

func (s *UpdateListenerAttributeRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateListenerAttributeRequest) GetIdleTimeout() *int32 {
	return s.IdleTimeout
}

func (s *UpdateListenerAttributeRequest) GetListenerDescription() *string {
	return s.ListenerDescription
}

func (s *UpdateListenerAttributeRequest) GetListenerId() *string {
	return s.ListenerId
}

func (s *UpdateListenerAttributeRequest) GetMss() *int32 {
	return s.Mss
}

func (s *UpdateListenerAttributeRequest) GetProxyProtocolEnabled() *bool {
	return s.ProxyProtocolEnabled
}

func (s *UpdateListenerAttributeRequest) GetProxyProtocolV2Config() *UpdateListenerAttributeRequestProxyProtocolV2Config {
	return s.ProxyProtocolV2Config
}

func (s *UpdateListenerAttributeRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateListenerAttributeRequest) GetSecSensorEnabled() *bool {
	return s.SecSensorEnabled
}

func (s *UpdateListenerAttributeRequest) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *UpdateListenerAttributeRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *UpdateListenerAttributeRequest) SetAlpnEnabled(v bool) *UpdateListenerAttributeRequest {
	s.AlpnEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetAlpnPolicy(v string) *UpdateListenerAttributeRequest {
	s.AlpnPolicy = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetCaCertificateIds(v []*string) *UpdateListenerAttributeRequest {
	s.CaCertificateIds = v
	return s
}

func (s *UpdateListenerAttributeRequest) SetCaEnabled(v bool) *UpdateListenerAttributeRequest {
	s.CaEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetCertificateIds(v []*string) *UpdateListenerAttributeRequest {
	s.CertificateIds = v
	return s
}

func (s *UpdateListenerAttributeRequest) SetClientToken(v string) *UpdateListenerAttributeRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetCps(v int32) *UpdateListenerAttributeRequest {
	s.Cps = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetDryRun(v bool) *UpdateListenerAttributeRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetIdleTimeout(v int32) *UpdateListenerAttributeRequest {
	s.IdleTimeout = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetListenerDescription(v string) *UpdateListenerAttributeRequest {
	s.ListenerDescription = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetListenerId(v string) *UpdateListenerAttributeRequest {
	s.ListenerId = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetMss(v int32) *UpdateListenerAttributeRequest {
	s.Mss = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetProxyProtocolEnabled(v bool) *UpdateListenerAttributeRequest {
	s.ProxyProtocolEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetProxyProtocolV2Config(v *UpdateListenerAttributeRequestProxyProtocolV2Config) *UpdateListenerAttributeRequest {
	s.ProxyProtocolV2Config = v
	return s
}

func (s *UpdateListenerAttributeRequest) SetRegionId(v string) *UpdateListenerAttributeRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetSecSensorEnabled(v bool) *UpdateListenerAttributeRequest {
	s.SecSensorEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetSecurityPolicyId(v string) *UpdateListenerAttributeRequest {
	s.SecurityPolicyId = &v
	return s
}

func (s *UpdateListenerAttributeRequest) SetServerGroupId(v string) *UpdateListenerAttributeRequest {
	s.ServerGroupId = &v
	return s
}

func (s *UpdateListenerAttributeRequest) Validate() error {
	return dara.Validate(s)
}

type UpdateListenerAttributeRequestProxyProtocolV2Config struct {
	// Specifies whether to use the Proxy protocol to pass the PrivateLinkEpId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	Ppv2PrivateLinkEpIdEnabled *bool `json:"Ppv2PrivateLinkEpIdEnabled,omitempty" xml:"Ppv2PrivateLinkEpIdEnabled,omitempty"`
	// Specifies whether to use the Proxy protocol to pass the PrivateLinkEpsId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	Ppv2PrivateLinkEpsIdEnabled *bool `json:"Ppv2PrivateLinkEpsIdEnabled,omitempty" xml:"Ppv2PrivateLinkEpsIdEnabled,omitempty"`
	// Specifies whether to use the Proxy protocol to pass the VpcId parameter to backend servers. Valid values:
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

func (s UpdateListenerAttributeRequestProxyProtocolV2Config) String() string {
	return dara.Prettify(s)
}

func (s UpdateListenerAttributeRequestProxyProtocolV2Config) GoString() string {
	return s.String()
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) GetPpv2PrivateLinkEpIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpIdEnabled
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) GetPpv2PrivateLinkEpsIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpsIdEnabled
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) GetPpv2VpcIdEnabled() *bool {
	return s.Ppv2VpcIdEnabled
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) SetPpv2PrivateLinkEpIdEnabled(v bool) *UpdateListenerAttributeRequestProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpIdEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) SetPpv2PrivateLinkEpsIdEnabled(v bool) *UpdateListenerAttributeRequestProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpsIdEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) SetPpv2VpcIdEnabled(v bool) *UpdateListenerAttributeRequestProxyProtocolV2Config {
	s.Ppv2VpcIdEnabled = &v
	return s
}

func (s *UpdateListenerAttributeRequestProxyProtocolV2Config) Validate() error {
	return dara.Validate(s)
}
