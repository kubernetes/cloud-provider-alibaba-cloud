// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iCreateServerGroupRequest interface {
	dara.Model
	String() string
	GoString() string
	SetAddressIPVersion(v string) *CreateServerGroupRequest
	GetAddressIPVersion() *string
	SetAnyPortEnabled(v bool) *CreateServerGroupRequest
	GetAnyPortEnabled() *bool
	SetClientToken(v string) *CreateServerGroupRequest
	GetClientToken() *string
	SetConnectionDrainEnabled(v bool) *CreateServerGroupRequest
	GetConnectionDrainEnabled() *bool
	SetConnectionDrainTimeout(v int32) *CreateServerGroupRequest
	GetConnectionDrainTimeout() *int32
	SetDryRun(v bool) *CreateServerGroupRequest
	GetDryRun() *bool
	SetHealthCheckConfig(v *CreateServerGroupRequestHealthCheckConfig) *CreateServerGroupRequest
	GetHealthCheckConfig() *CreateServerGroupRequestHealthCheckConfig
	SetPreserveClientIpEnabled(v bool) *CreateServerGroupRequest
	GetPreserveClientIpEnabled() *bool
	SetProtocol(v string) *CreateServerGroupRequest
	GetProtocol() *string
	SetRegionId(v string) *CreateServerGroupRequest
	GetRegionId() *string
	SetResourceGroupId(v string) *CreateServerGroupRequest
	GetResourceGroupId() *string
	SetScheduler(v string) *CreateServerGroupRequest
	GetScheduler() *string
	SetServerGroupName(v string) *CreateServerGroupRequest
	GetServerGroupName() *string
	SetServerGroupType(v string) *CreateServerGroupRequest
	GetServerGroupType() *string
	SetTag(v []*CreateServerGroupRequestTag) *CreateServerGroupRequest
	GetTag() []*CreateServerGroupRequestTag
	SetVpcId(v string) *CreateServerGroupRequest
	GetVpcId() *string
}

type CreateServerGroupRequest struct {
	// The IP version. Valid values:
	//
	// 	- **ipv4*	- (default): IPv4
	//
	// 	- **DualStack**: dual-stack
	//
	// example:
	//
	// ipv4
	AddressIPVersion *string `json:"AddressIPVersion,omitempty" xml:"AddressIPVersion,omitempty"`
	// Specifies whether to enable multi-port forwarding. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	AnyPortEnabled *bool `json:"AnyPortEnabled,omitempty" xml:"AnyPortEnabled,omitempty"`
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
	// Specifies whether to enable connection draining. Valid values:
	//
	// 	- **true**
	//
	// 	- **false*	- (default)
	//
	// example:
	//
	// false
	ConnectionDrainEnabled *bool `json:"ConnectionDrainEnabled,omitempty" xml:"ConnectionDrainEnabled,omitempty"`
	// Specifies a timeout period for connection draining. Unit: seconds. Valid values: **0*	- to **900**.
	//
	// example:
	//
	// 10
	ConnectionDrainTimeout *int32 `json:"ConnectionDrainTimeout,omitempty" xml:"ConnectionDrainTimeout,omitempty"`
	// Specifies whether to perform a dry run. Valid values:
	//
	// 	- **true:**: validates the request without performing the operation. The system checks the request for potential issues, including missing parameter values, incorrect request syntax, and service limits. If the request fails the validation, the corresponding error message is returned. If the request passes the validation, the `DryRunOperation` error code is returned.
	//
	// 	- **false*	- (default): validates the request and performs the operation. If the request passes the dry run, a 2xx HTTP status code is returned and the operation is performed.
	//
	// example:
	//
	// true
	DryRun *bool `json:"DryRun,omitempty" xml:"DryRun,omitempty"`
	// The configurations of health checks.
	HealthCheckConfig *CreateServerGroupRequestHealthCheckConfig `json:"HealthCheckConfig,omitempty" xml:"HealthCheckConfig,omitempty" type:"Struct"`
	// Specifies whether to enable client IP preservation. Valid values:
	//
	// 	- **true*	- (default)
	//
	// 	- **false**
	//
	// >  If you set this parameter to **true*	- and **Protocol*	- to **TCP**, the server group cannot be associated with **TCPSSL*	- listeners.
	//
	// if can be null:
	// false
	//
	// example:
	//
	// true
	PreserveClientIpEnabled *bool `json:"PreserveClientIpEnabled,omitempty" xml:"PreserveClientIpEnabled,omitempty"`
	// The protocol between the NLB instance and backend servers. Valid values:
	//
	// 	- **TCP*	- (default)
	//
	// 	- **UDP**
	//
	// 	- **TCP_UDP**
	//
	// > 	- If you set this parameter to **UDP**, you can associate the server group only with **UDP*	- listeners.
	//
	// > 	- If you set this parameter to **TCP*	- and **PreserveClientIpEnabled*	- to **true**, you can associate the server group only with **TCP*	- listeners.
	//
	// > 	- If you set this parameter to **TCP*	- and **PreserveClientIpEnabled*	- to **false**, you can associate the server group with **TCPSSL*	- and **TCP*	- listeners.
	//
	// > 	- If you set this parameter to **TCP_UDP**, you can associate the server group with **TCP*	- and **UDP*	- listeners.
	//
	// example:
	//
	// TCP
	Protocol *string `json:"Protocol,omitempty" xml:"Protocol,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The ID of the resource group to which the server group belongs.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The scheduling algorithm. Valid values:
	//
	// 	- **Wrr*	- (default): weighted round-robin. Backend servers with higher weights receive more requests.
	//
	// 	- **Wlc**: weighted least connections. Requests are distributed based on the weights and the number of connections to backend servers. If multiple backend servers have the same weight, requests are forwarded to the backend server with the least connections.
	//
	// 	- **rr**: Requests are forwarded to backend servers in sequence.
	//
	// 	- **sch**: source IP hash. Requests from the same source IP address are forwarded to the same backend server.
	//
	// 	- **tch**: consistent hashing based on four factors: source IP address, destination IP address, source port, and destination port. Requests that contain the same four factors are forwarded to the same backend server.
	//
	// 	- **qch**: QUIC ID hash. Requests that contain the same QUIC ID are forwarded to the same backend server.
	//
	// >  QUIC ID hash is supported only when the backend protocol is set to UDP.
	//
	// example:
	//
	// Wrr
	Scheduler *string `json:"Scheduler,omitempty" xml:"Scheduler,omitempty"`
	// The server group name.
	//
	// The name must be 2 to 128 characters in length, can contain digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// This parameter is required.
	//
	// example:
	//
	// NLB_ServerGroup
	ServerGroupName *string `json:"ServerGroupName,omitempty" xml:"ServerGroupName,omitempty"`
	// The type of the server group. Valid values:
	//
	// 	- **Instance*	- (default): allows you to specify servers of the **Ecs**, **Eni**, or **Eci*	- type.
	//
	// 	- **Ip**: allows you to specify IP addresses.
	//
	// example:
	//
	// Instance
	ServerGroupType *string `json:"ServerGroupType,omitempty" xml:"ServerGroupType,omitempty"`
	// The tags.
	//
	// if can be null:
	// true
	Tag []*CreateServerGroupRequestTag `json:"Tag,omitempty" xml:"Tag,omitempty" type:"Repeated"`
	// The ID of the virtual private cloud (VPC) where the server group is deployed.
	//
	// >  If **ServerGroupType*	- is set to **Instance**, only servers in the specified VPC can be added to the server group.
	//
	// This parameter is required.
	//
	// example:
	//
	// vpc-bp15zckdt37pq72zv****
	VpcId *string `json:"VpcId,omitempty" xml:"VpcId,omitempty"`
}

func (s CreateServerGroupRequest) String() string {
	return dara.Prettify(s)
}

func (s CreateServerGroupRequest) GoString() string {
	return s.String()
}

func (s *CreateServerGroupRequest) GetAddressIPVersion() *string {
	return s.AddressIPVersion
}

func (s *CreateServerGroupRequest) GetAnyPortEnabled() *bool {
	return s.AnyPortEnabled
}

func (s *CreateServerGroupRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *CreateServerGroupRequest) GetConnectionDrainEnabled() *bool {
	return s.ConnectionDrainEnabled
}

func (s *CreateServerGroupRequest) GetConnectionDrainTimeout() *int32 {
	return s.ConnectionDrainTimeout
}

func (s *CreateServerGroupRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *CreateServerGroupRequest) GetHealthCheckConfig() *CreateServerGroupRequestHealthCheckConfig {
	return s.HealthCheckConfig
}

func (s *CreateServerGroupRequest) GetPreserveClientIpEnabled() *bool {
	return s.PreserveClientIpEnabled
}

func (s *CreateServerGroupRequest) GetProtocol() *string {
	return s.Protocol
}

func (s *CreateServerGroupRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *CreateServerGroupRequest) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *CreateServerGroupRequest) GetScheduler() *string {
	return s.Scheduler
}

func (s *CreateServerGroupRequest) GetServerGroupName() *string {
	return s.ServerGroupName
}

func (s *CreateServerGroupRequest) GetServerGroupType() *string {
	return s.ServerGroupType
}

func (s *CreateServerGroupRequest) GetTag() []*CreateServerGroupRequestTag {
	return s.Tag
}

func (s *CreateServerGroupRequest) GetVpcId() *string {
	return s.VpcId
}

func (s *CreateServerGroupRequest) SetAddressIPVersion(v string) *CreateServerGroupRequest {
	s.AddressIPVersion = &v
	return s
}

func (s *CreateServerGroupRequest) SetAnyPortEnabled(v bool) *CreateServerGroupRequest {
	s.AnyPortEnabled = &v
	return s
}

func (s *CreateServerGroupRequest) SetClientToken(v string) *CreateServerGroupRequest {
	s.ClientToken = &v
	return s
}

func (s *CreateServerGroupRequest) SetConnectionDrainEnabled(v bool) *CreateServerGroupRequest {
	s.ConnectionDrainEnabled = &v
	return s
}

func (s *CreateServerGroupRequest) SetConnectionDrainTimeout(v int32) *CreateServerGroupRequest {
	s.ConnectionDrainTimeout = &v
	return s
}

func (s *CreateServerGroupRequest) SetDryRun(v bool) *CreateServerGroupRequest {
	s.DryRun = &v
	return s
}

func (s *CreateServerGroupRequest) SetHealthCheckConfig(v *CreateServerGroupRequestHealthCheckConfig) *CreateServerGroupRequest {
	s.HealthCheckConfig = v
	return s
}

func (s *CreateServerGroupRequest) SetPreserveClientIpEnabled(v bool) *CreateServerGroupRequest {
	s.PreserveClientIpEnabled = &v
	return s
}

func (s *CreateServerGroupRequest) SetProtocol(v string) *CreateServerGroupRequest {
	s.Protocol = &v
	return s
}

func (s *CreateServerGroupRequest) SetRegionId(v string) *CreateServerGroupRequest {
	s.RegionId = &v
	return s
}

func (s *CreateServerGroupRequest) SetResourceGroupId(v string) *CreateServerGroupRequest {
	s.ResourceGroupId = &v
	return s
}

func (s *CreateServerGroupRequest) SetScheduler(v string) *CreateServerGroupRequest {
	s.Scheduler = &v
	return s
}

func (s *CreateServerGroupRequest) SetServerGroupName(v string) *CreateServerGroupRequest {
	s.ServerGroupName = &v
	return s
}

func (s *CreateServerGroupRequest) SetServerGroupType(v string) *CreateServerGroupRequest {
	s.ServerGroupType = &v
	return s
}

func (s *CreateServerGroupRequest) SetTag(v []*CreateServerGroupRequestTag) *CreateServerGroupRequest {
	s.Tag = v
	return s
}

func (s *CreateServerGroupRequest) SetVpcId(v string) *CreateServerGroupRequest {
	s.VpcId = &v
	return s
}

func (s *CreateServerGroupRequest) Validate() error {
	return dara.Validate(s)
}

type CreateServerGroupRequestHealthCheckConfig struct {
	// The port that you want to use for health checks on backend servers.
	//
	// Valid values: **0*	- to **65535**.
	//
	// Default value: **0**. If you set this parameter to 0, the port that the backend server uses to provide services is also used for health checks.
	//
	// example:
	//
	// 0
	HealthCheckConnectPort *int32 `json:"HealthCheckConnectPort,omitempty" xml:"HealthCheckConnectPort,omitempty"`
	// The timeout period for a health check response. Unit: seconds. Valid values: **1*	- to **300**. Default value: **5**.
	//
	// example:
	//
	// 5
	HealthCheckConnectTimeout *int32 `json:"HealthCheckConnectTimeout,omitempty" xml:"HealthCheckConnectTimeout,omitempty"`
	// The domain name that is used for health checks. Valid values:
	//
	// 	- **$SERVER_IP**: the internal IP address of a backend server.
	//
	// 	- **domain**: a domain name. The domain name must be 1 to 80 characters in length, and can contain letters, digits, hyphens (-), and periods (.).
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// $SERVER_IP
	HealthCheckDomain *string `json:"HealthCheckDomain,omitempty" xml:"HealthCheckDomain,omitempty"`
	// Specifies whether to enable health checks. Valid values:
	//
	// 	- **true*	- (default)
	//
	// 	- **false**
	//
	// example:
	//
	// true
	HealthCheckEnabled *bool `json:"HealthCheckEnabled,omitempty" xml:"HealthCheckEnabled,omitempty"`
	// The response string that backend servers return to UDP listeners for health checks. The string must be 1 to 512 characters in length and can contain only letters and digits.
	//
	// example:
	//
	// ok
	HealthCheckExp *string `json:"HealthCheckExp,omitempty" xml:"HealthCheckExp,omitempty"`
	// The HTTP status codes to return for health checks. Separate multiple HTTP status codes with commas (,). Valid values: **http_2xx*	- (default), **http_3xx**, **http_4xx**, and **http_5xx**.
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	HealthCheckHttpCode []*string `json:"HealthCheckHttpCode,omitempty" xml:"HealthCheckHttpCode,omitempty" type:"Repeated"`
	// The HTTP version used for health checks. Valid values: **HTTP1.0*	- (default) and **HTTP1.1**.
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// HTTP1.0
	HealthCheckHttpVersion *string `json:"HealthCheckHttpVersion,omitempty" xml:"HealthCheckHttpVersion,omitempty"`
	// The interval at which health checks are performed. Unit: seconds. Default value: **5**.
	//
	// 	- If you set **HealthCheckType*	- to **TCP*	- or **HTTP**, valid values are **1*	- to **50**.
	//
	// 	- If you set **HealthCheckType*	- to **UDP**, valid values are **1*	- to **300**. Set the health check interval equal to or larger than the response timeout period to ensure that UDP response timeouts are not determined as no responses.
	//
	// example:
	//
	// 5
	HealthCheckInterval *int32 `json:"HealthCheckInterval,omitempty" xml:"HealthCheckInterval,omitempty"`
	// The request string that UDP listeners send to backend servers for health checks. The string must be 1 to 512 characters in length and can contain only letters and digits.
	//
	// example:
	//
	// hello
	HealthCheckReq *string `json:"HealthCheckReq,omitempty" xml:"HealthCheckReq,omitempty"`
	// The protocol that you want to use for health checks. Valid values:
	//
	// 	- **TCP**
	//
	// 	- **HTTP**
	//
	// 	- **UDP**
	//
	// example:
	//
	// TCP
	HealthCheckType *string `json:"HealthCheckType,omitempty" xml:"HealthCheckType,omitempty"`
	// The URL path to which health check probes are sent.
	//
	// The URL path must be 1 to 80 characters in length, and can contain letters, digits, and the following special characters: ` - / . % ? # &  `. It must start with a forward slash (/).
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// /test/index.html
	HealthCheckUrl *string `json:"HealthCheckUrl,omitempty" xml:"HealthCheckUrl,omitempty"`
	// The number of times that an unhealthy backend server must consecutively pass health checks before it is declared healthy. In this case, the health status changes from **fail*	- to **success**.
	//
	// Valid values: **2*	- to **10**.
	//
	// Default value: **2**.
	//
	// example:
	//
	// 2
	HealthyThreshold *int32 `json:"HealthyThreshold,omitempty" xml:"HealthyThreshold,omitempty"`
	// The HTTP method that is used for health checks. Valid values: **GET*	- (default) and **HEAD**.
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// GET
	HttpCheckMethod *string `json:"HttpCheckMethod,omitempty" xml:"HttpCheckMethod,omitempty"`
	// The number of times that a healthy backend server must consecutively fail health checks before it is declared unhealthy. In this case, the health status changes from **success*	- to **fail**.
	//
	// Valid values: **2*	- to **10**.
	//
	// Default value: **2**.
	//
	// example:
	//
	// 2
	UnhealthyThreshold *int32 `json:"UnhealthyThreshold,omitempty" xml:"UnhealthyThreshold,omitempty"`
}

func (s CreateServerGroupRequestHealthCheckConfig) String() string {
	return dara.Prettify(s)
}

func (s CreateServerGroupRequestHealthCheckConfig) GoString() string {
	return s.String()
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckConnectPort() *int32 {
	return s.HealthCheckConnectPort
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckConnectTimeout() *int32 {
	return s.HealthCheckConnectTimeout
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckDomain() *string {
	return s.HealthCheckDomain
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckEnabled() *bool {
	return s.HealthCheckEnabled
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckExp() *string {
	return s.HealthCheckExp
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckHttpCode() []*string {
	return s.HealthCheckHttpCode
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckHttpVersion() *string {
	return s.HealthCheckHttpVersion
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckInterval() *int32 {
	return s.HealthCheckInterval
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckReq() *string {
	return s.HealthCheckReq
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckType() *string {
	return s.HealthCheckType
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthCheckUrl() *string {
	return s.HealthCheckUrl
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHealthyThreshold() *int32 {
	return s.HealthyThreshold
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetHttpCheckMethod() *string {
	return s.HttpCheckMethod
}

func (s *CreateServerGroupRequestHealthCheckConfig) GetUnhealthyThreshold() *int32 {
	return s.UnhealthyThreshold
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckConnectPort(v int32) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckConnectPort = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckConnectTimeout(v int32) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckConnectTimeout = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckDomain(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckDomain = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckEnabled(v bool) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckEnabled = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckExp(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckExp = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckHttpCode(v []*string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckHttpCode = v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckHttpVersion(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckHttpVersion = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckInterval(v int32) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckInterval = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckReq(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckReq = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckType(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckType = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthCheckUrl(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthCheckUrl = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHealthyThreshold(v int32) *CreateServerGroupRequestHealthCheckConfig {
	s.HealthyThreshold = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetHttpCheckMethod(v string) *CreateServerGroupRequestHealthCheckConfig {
	s.HttpCheckMethod = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) SetUnhealthyThreshold(v int32) *CreateServerGroupRequestHealthCheckConfig {
	s.UnhealthyThreshold = &v
	return s
}

func (s *CreateServerGroupRequestHealthCheckConfig) Validate() error {
	return dara.Validate(s)
}

type CreateServerGroupRequestTag struct {
	// The key of the tag. The tag key can be up to 64 characters in length, cannot start with `aliyun` or `acs:`, and cannot contain `http://` or `https://`. The tag key can contain letters, digits, and the following special characters: _ . : / = + - @
	//
	// You can specify up to 20 tags in each call.
	//
	// example:
	//
	// env
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The value of the tag. The tag value can be up to 128 characters in length, cannot start with `acs:` or `aliyun`, and cannot contain `http://` or `https://`. The tag value can contain letters, digits, and the following special characters: _ . : / = + - @
	//
	// You can specify up to 20 tags in each call.
	//
	// example:
	//
	// product
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s CreateServerGroupRequestTag) String() string {
	return dara.Prettify(s)
}

func (s CreateServerGroupRequestTag) GoString() string {
	return s.String()
}

func (s *CreateServerGroupRequestTag) GetKey() *string {
	return s.Key
}

func (s *CreateServerGroupRequestTag) GetValue() *string {
	return s.Value
}

func (s *CreateServerGroupRequestTag) SetKey(v string) *CreateServerGroupRequestTag {
	s.Key = &v
	return s
}

func (s *CreateServerGroupRequestTag) SetValue(v string) *CreateServerGroupRequestTag {
	s.Value = &v
	return s
}

func (s *CreateServerGroupRequestTag) Validate() error {
	return dara.Validate(s)
}
