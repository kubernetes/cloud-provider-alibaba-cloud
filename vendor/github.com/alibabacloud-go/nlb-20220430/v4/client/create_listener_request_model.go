// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateListenerRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAlpnEnabled(v bool) *CreateListenerRequest
	GetAlpnEnabled() *bool
	SetAlpnPolicy(v string) *CreateListenerRequest
	GetAlpnPolicy() *string
	SetCaCertificateIds(v []*string) *CreateListenerRequest
	GetCaCertificateIds() []*string
	SetCaEnabled(v bool) *CreateListenerRequest
	GetCaEnabled() *bool
	SetCertificateIds(v []*string) *CreateListenerRequest
	GetCertificateIds() []*string
	SetClientToken(v string) *CreateListenerRequest
	GetClientToken() *string
	SetCps(v int32) *CreateListenerRequest
	GetCps() *int32
	SetDryRun(v bool) *CreateListenerRequest
	GetDryRun() *bool
	SetEndPort(v int32) *CreateListenerRequest
	GetEndPort() *int32
	SetIdleTimeout(v int32) *CreateListenerRequest
	GetIdleTimeout() *int32
	SetListenerDescription(v string) *CreateListenerRequest
	GetListenerDescription() *string
	SetListenerPort(v int32) *CreateListenerRequest
	GetListenerPort() *int32
	SetListenerProtocol(v string) *CreateListenerRequest
	GetListenerProtocol() *string
	SetLoadBalancerId(v string) *CreateListenerRequest
	GetLoadBalancerId() *string
	SetMss(v int32) *CreateListenerRequest
	GetMss() *int32
	SetProxyProtocolEnabled(v bool) *CreateListenerRequest
	GetProxyProtocolEnabled() *bool
	SetProxyProtocolV2Config(v *CreateListenerRequestProxyProtocolV2Config) *CreateListenerRequest
	GetProxyProtocolV2Config() *CreateListenerRequestProxyProtocolV2Config
	SetRegionId(v string) *CreateListenerRequest
	GetRegionId() *string
	SetSecSensorEnabled(v bool) *CreateListenerRequest
	GetSecSensorEnabled() *bool
	SetSecurityPolicyId(v string) *CreateListenerRequest
	GetSecurityPolicyId() *string
	SetServerGroupId(v string) *CreateListenerRequest
	GetServerGroupId() *string
	SetStartPort(v int32) *CreateListenerRequest
	GetStartPort() *int32
	SetTag(v []*CreateListenerRequestTag) *CreateListenerRequest
	GetTag() []*CreateListenerRequestTag
}

type CreateListenerRequest struct {
	// Specifies whether to enable Application-Layer Protocol Negotiation (ALPN). Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	AlpnEnabled *bool `json:"AlpnEnabled,omitempty" xml:"AlpnEnabled,omitempty"`
	// The ALPN policy. Valid values:
	//
	// 	- **HTTP1Only**: uses only HTTP 1.x. The priority of HTTP 1.1 is higher than that of HTTP 1.0.
	//
	// 	- **HTTP2Only**: uses only HTTP 2.0.
	//
	// 	- **HTTP2Optional**: preferentially uses HTTP 1.x over HTTP 2.0. The priority of HTTP 1.1 is higher than that of HTTP 1.0, and the priority of HTTP 1.0 is higher than that of HTTP 2.0.
	//
	// 	- **HTTP2Preferred**: preferentially uses HTTP 2.0 over HTTP 1.x. The priority of HTTP 2.0 is higher than that of HTTP 1.1, and the priority of HTTP 1.1 is higher than that of HTTP 1.0.
	//
	// >  This parameter is required if **AlpnEnabled*	- is set to true.
	//
	// example:
	//
	// HTTP1Only
	AlpnPolicy *string `json:"AlpnPolicy,omitempty" xml:"AlpnPolicy,omitempty"`
	// The certificate authority (CA) certificate. This parameter is supported only by TCLSSL listeners.
	//
	// >  You can specify only one CA certificate.
	CaCertificateIds []*string `json:"CaCertificateIds,omitempty" xml:"CaCertificateIds,omitempty" type:"Repeated"`
	// Specifies whether to enable mutual authentication. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	CaEnabled *bool `json:"CaEnabled,omitempty" xml:"CaEnabled,omitempty"`
	// The server certificate. This parameter is supported only by TCLSSL listeners.
	//
	// >  You can specify only one server certificate.
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
	// 100
	Cps *int32 `json:"Cps,omitempty" xml:"Cps,omitempty"`
	// Specifies whether to perform a dry run, without sending the actual request. Valid values:
	//
	// 	- **true**: validates the request without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the validation, the corresponding error message is returned. If the request passes the validation, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): validates the request and performs the operation. If the request passes the validation, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// false
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The last port in the listener port range. Valid values: **0*	- to **65535**. The port number of the last port must be greater than the port number of the first port.
	//
	// >  This parameter is required when **ListenerPort*	- is set to **0**.
	//
	// example:
	//
	// 566
	EndPort *int32 `json:"EndPort,omitempty" xml:"EndPort,omitempty"`
	// The timeout period for idle connections. Unit: seconds.
	//
	// 	- If you set **ListenerProtocol*	- to **TCP*	- or **TCPSSL**, this parameter can be set to a value ranging from **10*	- to **900**. Default value: **900**.
	//
	// 	- If **ListenerProtocol*	- is set to **UDP**, this parameter can be set to a value ranging from **10*	- to **20**. Default value: **20**.
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
	// tcp_80
	ListenerDescription *string `json:"ListenerDescription,omitempty" xml:"ListenerDescription,omitempty"`
	// The listener port. Valid values: **0*	- to **65535**.
	//
	// If you set this parameter to **0**, the listener listens by port range. If you set this parameter to **0**, you must also set the **StartPort*	- and **EndPort*	- parameters.
	//
	// This parameter is required.
	//
	// example:
	//
	// 80
	ListenerPort *int32 `json:"ListenerPort,omitempty" xml:"ListenerPort,omitempty"`
	// The listener protocol. Valid values: **TCP**, **UDP**, and **TCPSSL**.
	//
	// This parameter is required.
	//
	// example:
	//
	// TCP
	ListenerProtocol *string `json:"ListenerProtocol,omitempty" xml:"ListenerProtocol,omitempty"`
	// The ID of the NLB instance.
	//
	// This parameter is required.
	//
	// example:
	//
	// nlb-83ckzc8d4xlp8o****
	LoadBalancerId *string `json:"LoadBalancerId,omitempty" xml:"LoadBalancerId,omitempty"`
	// The size of the largest TCP packet segment. Unit: bytes. Valid values: **0*	- to **1500**. **0*	- indicates that the maximum segment size (MSS) value of TCP packets remains unchanged.
	//
	// >  This parameter takes effect only for TCP and TCPSSL listeners.
	//
	// example:
	//
	// 43
	Mss *int32 `json:"Mss,omitempty" xml:"Mss,omitempty"`
	// Specifies whether to use the Proxy protocol to pass client IP addresses to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	ProxyProtocolEnabled *bool `json:"ProxyProtocolEnabled,omitempty" xml:"ProxyProtocolEnabled,omitempty"`
	// Specifies whether to use the Proxy protocol to pass the VpcId, PrivateLinkEpId, and PrivateLinkEpsId parameters to backend servers.
	ProxyProtocolV2Config *CreateListenerRequestProxyProtocolV2Config `json:"ProxyProtocolV2Config,omitempty" xml:"ProxyProtocolV2Config,omitempty" type:"Struct"`
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
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	SecSensorEnabled *bool `json:"SecSensorEnabled,omitempty" xml:"SecSensorEnabled,omitempty"`
	// The ID of the security policy. System security policies and custom security policies are supported.
	//
	// 	- Valid values for system security policies: **tls_cipher_policy_1_0*	- (default), **tls_cipher_policy_1_1**, **tls_cipher_policy_1_2**, **tls_cipher_policy_1_2_strict**, and **tls_cipher_policy_1_2_strict_with_1_3**.
	//
	// 	- For a custom security policy, enter the policy ID.
	//
	//     	- For information about creating a custom security policy, see [CreateSecurityPolicy](https://help.aliyun.com/document_detail/445901.html).
	//
	//     	- For information about querying security policies, see [ListSecurityPolicy](https://help.aliyun.com/document_detail/445900.html).
	//
	// >  This parameter takes effect only for TCPSSL listeners.
	//
	// example:
	//
	// tls_cipher_policy_1_0
	SecurityPolicyId *string `json:"SecurityPolicyId,omitempty" xml:"SecurityPolicyId,omitempty"`
	// The server group ID.
	//
	// >  	- If you set **ListenerProtocol*	- to **TCP**, you can associate the listener with server groups whose backend protocol is **TCP*	- or **TCP_UDP**. You cannot associate the listener with server groups whose backend protocol is **UDP**.
	//
	// >  	- If you set **ListenerProtocol*	- to **UDP**, you can associate the listener with server groups whose backend protocol is **UDP*	- or **TCP_UDP**. You cannot associate the listener with server groups whose backend protocol is **TCP**.
	//
	// >  	- If you set **ListenerProtocol*	- to **TCPSSL**, you can associate the listener with server groups whose backend protocol is **TCP*	- and have **client IP preservation disabled**. You cannot associate the listener with server groups whose backend protocol is **TCP*	- and have **client IP preservation enabled*	- or server groups whose backend protocol is **UDP*	- or **TCP_UDP**.
	//
	// This parameter is required.
	//
	// example:
	//
	// sgp-ppdpc14gdm3x4o****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The first port in the listener port range. Valid values: **0*	- to **65535**.
	//
	// >  This parameter is required when **ListenerPort*	- is set to **0**.
	//
	// example:
	//
	// 244
	StartPort *int32 `json:"StartPort,omitempty" xml:"StartPort,omitempty"`
	// The tags.
	//
	// if can be null:
	// true
	Tag []*CreateListenerRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
}

func (s CreateListenerRequest) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerRequest) GoString() string {
	return s.String()
}

func (s *CreateListenerRequest) GetAlpnEnabled() *bool {
	return s.AlpnEnabled
}

func (s *CreateListenerRequest) GetAlpnPolicy() *string {
	return s.AlpnPolicy
}

func (s *CreateListenerRequest) GetCaCertificateIds() []*string {
	return s.CaCertificateIds
}

func (s *CreateListenerRequest) GetCaEnabled() *bool {
	return s.CaEnabled
}

func (s *CreateListenerRequest) GetCertificateIds() []*string {
	return s.CertificateIds
}

func (s *CreateListenerRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *CreateListenerRequest) GetCps() *int32 {
	return s.Cps
}

func (s *CreateListenerRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *CreateListenerRequest) GetEndPort() *int32 {
	return s.EndPort
}

func (s *CreateListenerRequest) GetIdleTimeout() *int32 {
	return s.IdleTimeout
}

func (s *CreateListenerRequest) GetListenerDescription() *string {
	return s.ListenerDescription
}

func (s *CreateListenerRequest) GetListenerPort() *int32 {
	return s.ListenerPort
}

func (s *CreateListenerRequest) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *CreateListenerRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *CreateListenerRequest) GetMss() *int32 {
	return s.Mss
}

func (s *CreateListenerRequest) GetProxyProtocolEnabled() *bool {
	return s.ProxyProtocolEnabled
}

func (s *CreateListenerRequest) GetProxyProtocolV2Config() *CreateListenerRequestProxyProtocolV2Config {
	return s.ProxyProtocolV2Config
}

func (s *CreateListenerRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *CreateListenerRequest) GetSecSensorEnabled() *bool {
	return s.SecSensorEnabled
}

func (s *CreateListenerRequest) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *CreateListenerRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *CreateListenerRequest) GetStartPort() *int32 {
	return s.StartPort
}

func (s *CreateListenerRequest) GetTag() []*CreateListenerRequestTag {
	return s.Tag
}

func (s *CreateListenerRequest) SetAlpnEnabled(v bool) *CreateListenerRequest {
	s.AlpnEnabled = &v
	return s
}

func (s *CreateListenerRequest) SetAlpnPolicy(v string) *CreateListenerRequest {
	s.AlpnPolicy = &v
	return s
}

func (s *CreateListenerRequest) SetCaCertificateIds(v []*string) *CreateListenerRequest {
	s.CaCertificateIds = v
	return s
}

func (s *CreateListenerRequest) SetCaEnabled(v bool) *CreateListenerRequest {
	s.CaEnabled = &v
	return s
}

func (s *CreateListenerRequest) SetCertificateIds(v []*string) *CreateListenerRequest {
	s.CertificateIds = v
	return s
}

func (s *CreateListenerRequest) SetClientToken(v string) *CreateListenerRequest {
	s.ClientToken = &v
	return s
}

func (s *CreateListenerRequest) SetCps(v int32) *CreateListenerRequest {
	s.Cps = &v
	return s
}

func (s *CreateListenerRequest) SetDryRun(v bool) *CreateListenerRequest {
	s.DryRun = &v
	return s
}

func (s *CreateListenerRequest) SetEndPort(v int32) *CreateListenerRequest {
	s.EndPort = &v
	return s
}

func (s *CreateListenerRequest) SetIdleTimeout(v int32) *CreateListenerRequest {
	s.IdleTimeout = &v
	return s
}

func (s *CreateListenerRequest) SetListenerDescription(v string) *CreateListenerRequest {
	s.ListenerDescription = &v
	return s
}

func (s *CreateListenerRequest) SetListenerPort(v int32) *CreateListenerRequest {
	s.ListenerPort = &v
	return s
}

func (s *CreateListenerRequest) SetListenerProtocol(v string) *CreateListenerRequest {
	s.ListenerProtocol = &v
	return s
}

func (s *CreateListenerRequest) SetLoadBalancerId(v string) *CreateListenerRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *CreateListenerRequest) SetMss(v int32) *CreateListenerRequest {
	s.Mss = &v
	return s
}

func (s *CreateListenerRequest) SetProxyProtocolEnabled(v bool) *CreateListenerRequest {
	s.ProxyProtocolEnabled = &v
	return s
}

func (s *CreateListenerRequest) SetProxyProtocolV2Config(v *CreateListenerRequestProxyProtocolV2Config) *CreateListenerRequest {
	s.ProxyProtocolV2Config = v
	return s
}

func (s *CreateListenerRequest) SetRegionId(v string) *CreateListenerRequest {
	s.RegionId = &v
	return s
}

func (s *CreateListenerRequest) SetSecSensorEnabled(v bool) *CreateListenerRequest {
	s.SecSensorEnabled = &v
	return s
}

func (s *CreateListenerRequest) SetSecurityPolicyId(v string) *CreateListenerRequest {
	s.SecurityPolicyId = &v
	return s
}

func (s *CreateListenerRequest) SetServerGroupId(v string) *CreateListenerRequest {
	s.ServerGroupId = &v
	return s
}

func (s *CreateListenerRequest) SetStartPort(v int32) *CreateListenerRequest {
	s.StartPort = &v
	return s
}

func (s *CreateListenerRequest) SetTag(v []*CreateListenerRequestTag) *CreateListenerRequest {
	s.Tag = v
	return s
}

func (s *CreateListenerRequest) Validate() error {
	return dara.Validate(s)
}

type CreateListenerRequestProxyProtocolV2Config struct {
	// Specifies whether to use the Proxy protocol to pass the Ppv2PrivateLinkEpId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	Ppv2PrivateLinkEpIdEnabled *bool `json:"Ppv2PrivateLinkEpIdEnabled,omitempty" xml:"Ppv2PrivateLinkEpIdEnabled,omitempty"`
	// Specifies whether to use the Proxy protocol to pass the PrivateLinkEpsId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	Ppv2PrivateLinkEpsIdEnabled *bool `json:"Ppv2PrivateLinkEpsIdEnabled,omitempty" xml:"Ppv2PrivateLinkEpsIdEnabled,omitempty"`
	// Specifies whether to use the Proxy protocol to pass the VpcId parameter to backend servers. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	Ppv2VpcIdEnabled *bool `json:"Ppv2VpcIdEnabled,omitempty" xml:"Ppv2VpcIdEnabled,omitempty"`
}

func (s CreateListenerRequestProxyProtocolV2Config) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerRequestProxyProtocolV2Config) GoString() string {
	return s.String()
}

func (s *CreateListenerRequestProxyProtocolV2Config) GetPpv2PrivateLinkEpIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpIdEnabled
}

func (s *CreateListenerRequestProxyProtocolV2Config) GetPpv2PrivateLinkEpsIdEnabled() *bool {
	return s.Ppv2PrivateLinkEpsIdEnabled
}

func (s *CreateListenerRequestProxyProtocolV2Config) GetPpv2VpcIdEnabled() *bool {
	return s.Ppv2VpcIdEnabled
}

func (s *CreateListenerRequestProxyProtocolV2Config) SetPpv2PrivateLinkEpIdEnabled(v bool) *CreateListenerRequestProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpIdEnabled = &v
	return s
}

func (s *CreateListenerRequestProxyProtocolV2Config) SetPpv2PrivateLinkEpsIdEnabled(v bool) *CreateListenerRequestProxyProtocolV2Config {
	s.Ppv2PrivateLinkEpsIdEnabled = &v
	return s
}

func (s *CreateListenerRequestProxyProtocolV2Config) SetPpv2VpcIdEnabled(v bool) *CreateListenerRequestProxyProtocolV2Config {
	s.Ppv2VpcIdEnabled = &v
	return s
}

func (s *CreateListenerRequestProxyProtocolV2Config) Validate() error {
	return dara.Validate(s)
}

type CreateListenerRequestTag struct {
	// The key of the tag. The tag key can be up to 64 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`. The tag value can contain letters, digits, and the following special characters: _ . : / = + - @
	//
	// You can specify up to 20 tags in each call.
	//
	// example:
	//
	// KeyTest
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. The tag value can be up to 128 characters in length, cannot start with `acs:` or `aliyun`, and cannot contain `http://` or `https://`. The tag value can contain letters, digits, and the following special characters: _ . : / = + - @
	//
	// You can specify up to 20 tags in each call.
	//
	// example:
	//
	// Test
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s CreateListenerRequestTag) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerRequestTag) GoString() string {
	return s.String()
}

func (s *CreateListenerRequestTag) GetKey() *string {
	return s.Key
}

func (s *CreateListenerRequestTag) GetValue() *string {
	return s.Value
}

func (s *CreateListenerRequestTag) SetKey(v string) *CreateListenerRequestTag {
	s.Key = &v
	return s
}

func (s *CreateListenerRequestTag) SetValue(v string) *CreateListenerRequestTag {
	s.Value = &v
	return s
}

func (s *CreateListenerRequestTag) Validate() error {
	return dara.Validate(s)
}
