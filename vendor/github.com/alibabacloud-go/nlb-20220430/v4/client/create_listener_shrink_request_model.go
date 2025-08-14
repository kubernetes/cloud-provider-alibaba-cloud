// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateListenerShrinkRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAlpnEnabled(v bool) *CreateListenerShrinkRequest
	GetAlpnEnabled() *bool
	SetAlpnPolicy(v string) *CreateListenerShrinkRequest
	GetAlpnPolicy() *string
	SetCaCertificateIds(v []*string) *CreateListenerShrinkRequest
	GetCaCertificateIds() []*string
	SetCaEnabled(v bool) *CreateListenerShrinkRequest
	GetCaEnabled() *bool
	SetCertificateIds(v []*string) *CreateListenerShrinkRequest
	GetCertificateIds() []*string
	SetClientToken(v string) *CreateListenerShrinkRequest
	GetClientToken() *string
	SetCps(v int32) *CreateListenerShrinkRequest
	GetCps() *int32
	SetDryRun(v bool) *CreateListenerShrinkRequest
	GetDryRun() *bool
	SetEndPort(v int32) *CreateListenerShrinkRequest
	GetEndPort() *int32
	SetIdleTimeout(v int32) *CreateListenerShrinkRequest
	GetIdleTimeout() *int32
	SetListenerDescription(v string) *CreateListenerShrinkRequest
	GetListenerDescription() *string
	SetListenerPort(v int32) *CreateListenerShrinkRequest
	GetListenerPort() *int32
	SetListenerProtocol(v string) *CreateListenerShrinkRequest
	GetListenerProtocol() *string
	SetLoadBalancerId(v string) *CreateListenerShrinkRequest
	GetLoadBalancerId() *string
	SetMss(v int32) *CreateListenerShrinkRequest
	GetMss() *int32
	SetProxyProtocolEnabled(v bool) *CreateListenerShrinkRequest
	GetProxyProtocolEnabled() *bool
	SetProxyProtocolV2ConfigShrink(v string) *CreateListenerShrinkRequest
	GetProxyProtocolV2ConfigShrink() *string
	SetRegionId(v string) *CreateListenerShrinkRequest
	GetRegionId() *string
	SetSecSensorEnabled(v bool) *CreateListenerShrinkRequest
	GetSecSensorEnabled() *bool
	SetSecurityPolicyId(v string) *CreateListenerShrinkRequest
	GetSecurityPolicyId() *string
	SetServerGroupId(v string) *CreateListenerShrinkRequest
	GetServerGroupId() *string
	SetStartPort(v int32) *CreateListenerShrinkRequest
	GetStartPort() *int32
	SetTag(v []*CreateListenerShrinkRequestTag) *CreateListenerShrinkRequest
	GetTag() []*CreateListenerShrinkRequestTag
}

type CreateListenerShrinkRequest struct {
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
	ProxyProtocolV2ConfigShrink *string `json:"ProxyProtocolV2Config,omitempty" xml:"ProxyProtocolV2Config,omitempty"`
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
	Tag []*CreateListenerShrinkRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
}

func (s CreateListenerShrinkRequest) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerShrinkRequest) GoString() string {
	return s.String()
}

func (s *CreateListenerShrinkRequest) GetAlpnEnabled() *bool {
	return s.AlpnEnabled
}

func (s *CreateListenerShrinkRequest) GetAlpnPolicy() *string {
	return s.AlpnPolicy
}

func (s *CreateListenerShrinkRequest) GetCaCertificateIds() []*string {
	return s.CaCertificateIds
}

func (s *CreateListenerShrinkRequest) GetCaEnabled() *bool {
	return s.CaEnabled
}

func (s *CreateListenerShrinkRequest) GetCertificateIds() []*string {
	return s.CertificateIds
}

func (s *CreateListenerShrinkRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *CreateListenerShrinkRequest) GetCps() *int32 {
	return s.Cps
}

func (s *CreateListenerShrinkRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *CreateListenerShrinkRequest) GetEndPort() *int32 {
	return s.EndPort
}

func (s *CreateListenerShrinkRequest) GetIdleTimeout() *int32 {
	return s.IdleTimeout
}

func (s *CreateListenerShrinkRequest) GetListenerDescription() *string {
	return s.ListenerDescription
}

func (s *CreateListenerShrinkRequest) GetListenerPort() *int32 {
	return s.ListenerPort
}

func (s *CreateListenerShrinkRequest) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *CreateListenerShrinkRequest) GetLoadBalancerId() *string {
	return s.LoadBalancerId
}

func (s *CreateListenerShrinkRequest) GetMss() *int32 {
	return s.Mss
}

func (s *CreateListenerShrinkRequest) GetProxyProtocolEnabled() *bool {
	return s.ProxyProtocolEnabled
}

func (s *CreateListenerShrinkRequest) GetProxyProtocolV2ConfigShrink() *string {
	return s.ProxyProtocolV2ConfigShrink
}

func (s *CreateListenerShrinkRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *CreateListenerShrinkRequest) GetSecSensorEnabled() *bool {
	return s.SecSensorEnabled
}

func (s *CreateListenerShrinkRequest) GetSecurityPolicyId() *string {
	return s.SecurityPolicyId
}

func (s *CreateListenerShrinkRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *CreateListenerShrinkRequest) GetStartPort() *int32 {
	return s.StartPort
}

func (s *CreateListenerShrinkRequest) GetTag() []*CreateListenerShrinkRequestTag {
	return s.Tag
}

func (s *CreateListenerShrinkRequest) SetAlpnEnabled(v bool) *CreateListenerShrinkRequest {
	s.AlpnEnabled = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetAlpnPolicy(v string) *CreateListenerShrinkRequest {
	s.AlpnPolicy = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetCaCertificateIds(v []*string) *CreateListenerShrinkRequest {
	s.CaCertificateIds = v
	return s
}

func (s *CreateListenerShrinkRequest) SetCaEnabled(v bool) *CreateListenerShrinkRequest {
	s.CaEnabled = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetCertificateIds(v []*string) *CreateListenerShrinkRequest {
	s.CertificateIds = v
	return s
}

func (s *CreateListenerShrinkRequest) SetClientToken(v string) *CreateListenerShrinkRequest {
	s.ClientToken = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetCps(v int32) *CreateListenerShrinkRequest {
	s.Cps = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetDryRun(v bool) *CreateListenerShrinkRequest {
	s.DryRun = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetEndPort(v int32) *CreateListenerShrinkRequest {
	s.EndPort = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetIdleTimeout(v int32) *CreateListenerShrinkRequest {
	s.IdleTimeout = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetListenerDescription(v string) *CreateListenerShrinkRequest {
	s.ListenerDescription = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetListenerPort(v int32) *CreateListenerShrinkRequest {
	s.ListenerPort = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetListenerProtocol(v string) *CreateListenerShrinkRequest {
	s.ListenerProtocol = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetLoadBalancerId(v string) *CreateListenerShrinkRequest {
	s.LoadBalancerId = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetMss(v int32) *CreateListenerShrinkRequest {
	s.Mss = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetProxyProtocolEnabled(v bool) *CreateListenerShrinkRequest {
	s.ProxyProtocolEnabled = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetProxyProtocolV2ConfigShrink(v string) *CreateListenerShrinkRequest {
	s.ProxyProtocolV2ConfigShrink = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetRegionId(v string) *CreateListenerShrinkRequest {
	s.RegionId = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetSecSensorEnabled(v bool) *CreateListenerShrinkRequest {
	s.SecSensorEnabled = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetSecurityPolicyId(v string) *CreateListenerShrinkRequest {
	s.SecurityPolicyId = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetServerGroupId(v string) *CreateListenerShrinkRequest {
	s.ServerGroupId = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetStartPort(v int32) *CreateListenerShrinkRequest {
	s.StartPort = &v
	return s
}

func (s *CreateListenerShrinkRequest) SetTag(v []*CreateListenerShrinkRequestTag) *CreateListenerShrinkRequest {
	s.Tag = v
	return s
}

func (s *CreateListenerShrinkRequest) Validate() error {
	return dara.Validate(s)
}

type CreateListenerShrinkRequestTag struct {
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

func (s CreateListenerShrinkRequestTag) String() string {
	return dara.Prettify(s)
}

func (s CreateListenerShrinkRequestTag) GoString() string {
	return s.String()
}

func (s *CreateListenerShrinkRequestTag) GetKey() *string {
	return s.Key
}

func (s *CreateListenerShrinkRequestTag) GetValue() *string {
	return s.Value
}

func (s *CreateListenerShrinkRequestTag) SetKey(v string) *CreateListenerShrinkRequestTag {
	s.Key = &v
	return s
}

func (s *CreateListenerShrinkRequestTag) SetValue(v string) *CreateListenerShrinkRequestTag {
	s.Value = &v
	return s
}

func (s *CreateListenerShrinkRequestTag) Validate() error {
	return dara.Validate(s)
}
