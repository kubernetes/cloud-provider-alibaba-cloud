// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iUpdateServerGroupAttributeRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *UpdateServerGroupAttributeRequest
	GetClientToken() *string
	SetConnectionDrainEnabled(v bool) *UpdateServerGroupAttributeRequest
	GetConnectionDrainEnabled() *bool
	SetConnectionDrainTimeout(v int32) *UpdateServerGroupAttributeRequest
	GetConnectionDrainTimeout() *int32
	SetDryRun(v bool) *UpdateServerGroupAttributeRequest
	GetDryRun() *bool
	SetHealthCheckConfig(v *UpdateServerGroupAttributeRequestHealthCheckConfig) *UpdateServerGroupAttributeRequest
	GetHealthCheckConfig() *UpdateServerGroupAttributeRequestHealthCheckConfig
	SetPreserveClientIpEnabled(v bool) *UpdateServerGroupAttributeRequest
	GetPreserveClientIpEnabled() *bool
	SetRegionId(v string) *UpdateServerGroupAttributeRequest
	GetRegionId() *string
	SetScheduler(v string) *UpdateServerGroupAttributeRequest
	GetScheduler() *string
	SetServerGroupId(v string) *UpdateServerGroupAttributeRequest
	GetServerGroupId() *string
	SetServerGroupName(v string) *UpdateServerGroupAttributeRequest
	GetServerGroupName() *string
}

type UpdateServerGroupAttributeRequest struct {
	// The client token used to ensure the idempotence of the request.
	//
	// You can use the client to generate the token. Ensure that the token is unique among different requests. Only ASCII characters are allowed.
	//
	// >  If you do not specify this parameter, the value of **RequestId*	- is used.***	- The value of **RequestId*	- is different for each request.
	//
	// example:
	//
	// 123e4567-e89b-12d3-a456-426655440000
	ClientToken *string `json:"ClientToken,omitempty" xml:"ClientToken,omitempty"`
	// Specifies whether to enable connection draining. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	ConnectionDrainEnabled *bool `json:"ConnectionDrainEnabled,omitempty" xml:"ConnectionDrainEnabled,omitempty"`
	// Specify a timeout period for connection draining. Unit: seconds. Valid values: **0*	- to **900**.
	//
	// example:
	//
	// 10
	ConnectionDrainTimeout *int32 `json:"ConnectionDrainTimeout,omitempty" xml:"ConnectionDrainTimeout,omitempty"`
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
	// Health check configurations.
	HealthCheckConfig *UpdateServerGroupAttributeRequestHealthCheckConfig `json:"HealthCheckConfig,omitempty" xml:"HealthCheckConfig,omitempty" type:"Struct"`
	// Specifies whether to enable client IP preservation. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// >  You cannot set this parameter to **true*	- if **PreserveClientIpEnabled*	- is set to **false*	- and the server group is associated with a listener that uses **SSL over TCP**.
	//
	// example:
	//
	// false
	PreserveClientIpEnabled *bool `json:"PreserveClientIpEnabled,omitempty" xml:"PreserveClientIpEnabled,omitempty"`
	// The ID of the region where the NLB instance is deployed.
	//
	// You can call the [DescribeRegions](https://help.aliyun.com/document_detail/443657.html) operation to query the most recent region list.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The scheduling algorithm. Valid values:
	//
	// 	- **Wrr**: weighted round-robin. Backend servers with higher weights receive more requests.
	//
	// 	- **Wlc**: weighted least connections. Requests are distributed based on the weight and load of each backend server. The load refers to the number of connections on a backend server. If multiple backend servers have the same weight, requests are forwarded to the backend server with the least connections.
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
	// The server group ID.
	//
	// This parameter is required.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The new name of the server group.
	//
	// The name must be 2 to 128 characters in length, can contain digits, periods (.), underscores (_), and hyphens (-), and must start with a letter.
	//
	// example:
	//
	// NLB_ServerGroup1
	ServerGroupName *string `json:"ServerGroupName,omitempty" xml:"ServerGroupName,omitempty"`
}

func (s UpdateServerGroupAttributeRequest) String() string {
	return dara.Prettify(s)
}

func (s UpdateServerGroupAttributeRequest) GoString() string {
	return s.String()
}

func (s *UpdateServerGroupAttributeRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *UpdateServerGroupAttributeRequest) GetConnectionDrainEnabled() *bool {
	return s.ConnectionDrainEnabled
}

func (s *UpdateServerGroupAttributeRequest) GetConnectionDrainTimeout() *int32 {
	return s.ConnectionDrainTimeout
}

func (s *UpdateServerGroupAttributeRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *UpdateServerGroupAttributeRequest) GetHealthCheckConfig() *UpdateServerGroupAttributeRequestHealthCheckConfig {
	return s.HealthCheckConfig
}

func (s *UpdateServerGroupAttributeRequest) GetPreserveClientIpEnabled() *bool {
	return s.PreserveClientIpEnabled
}

func (s *UpdateServerGroupAttributeRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *UpdateServerGroupAttributeRequest) GetScheduler() *string {
	return s.Scheduler
}

func (s *UpdateServerGroupAttributeRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *UpdateServerGroupAttributeRequest) GetServerGroupName() *string {
	return s.ServerGroupName
}

func (s *UpdateServerGroupAttributeRequest) SetClientToken(v string) *UpdateServerGroupAttributeRequest {
	s.ClientToken = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetConnectionDrainEnabled(v bool) *UpdateServerGroupAttributeRequest {
	s.ConnectionDrainEnabled = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetConnectionDrainTimeout(v int32) *UpdateServerGroupAttributeRequest {
	s.ConnectionDrainTimeout = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetDryRun(v bool) *UpdateServerGroupAttributeRequest {
	s.DryRun = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetHealthCheckConfig(v *UpdateServerGroupAttributeRequestHealthCheckConfig) *UpdateServerGroupAttributeRequest {
	s.HealthCheckConfig = v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetPreserveClientIpEnabled(v bool) *UpdateServerGroupAttributeRequest {
	s.PreserveClientIpEnabled = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetRegionId(v string) *UpdateServerGroupAttributeRequest {
	s.RegionId = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetScheduler(v string) *UpdateServerGroupAttributeRequest {
	s.Scheduler = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetServerGroupId(v string) *UpdateServerGroupAttributeRequest {
	s.ServerGroupId = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) SetServerGroupName(v string) *UpdateServerGroupAttributeRequest {
	s.ServerGroupName = &v
	return s
}

func (s *UpdateServerGroupAttributeRequest) Validate() error {
	return dara.Validate(s)
}

type UpdateServerGroupAttributeRequestHealthCheckConfig struct {
	// The backend port that is used for health checks. Valid values: **0*	- to **65535**. If you set this parameter to **0**, the port that the backend server uses to provide services is also used for health checks.
	//
	// example:
	//
	// 0
	HealthCheckConnectPort *int32 `json:"HealthCheckConnectPort,omitempty" xml:"HealthCheckConnectPort,omitempty"`
	// The timeout period for a health check response. Unit: seconds. Valid values: **1 to 300**. Default value: **5**.
	//
	// example:
	//
	// 100
	HealthCheckConnectTimeout *int32 `json:"HealthCheckConnectTimeout,omitempty" xml:"HealthCheckConnectTimeout,omitempty"`
	// The domain name used for health checks. Valid values:
	//
	// 	- **$SERVER_IP**: the internal IP address of a backend server.
	//
	// 	- **domain**: the specified domain name. The domain name must be 1 to 80 characters in length, and can contain lowercase letters, digits, hyphens (-), and periods (.).
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// $SERVER_IP
	HealthCheckDomain *string `json:"HealthCheckDomain,omitempty" xml:"HealthCheckDomain,omitempty"`
	// Specifies whether to enable health checks. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	HealthCheckEnabled *bool `json:"HealthCheckEnabled,omitempty" xml:"HealthCheckEnabled,omitempty"`
	// The response string of UDP health checks. The string must be 1 to 512 characters in length, and can contain letters and digits.
	//
	// example:
	//
	// ok
	HealthCheckExp *string `json:"HealthCheckExp,omitempty" xml:"HealthCheckExp,omitempty"`
	// The HTTP status codes to return for health checks. Separate multiple HTTP status codes with commas (,). Valid values: **http_2xx*	- (default), **http_3xx**, **http_4xx**, and **http_5xx**.
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	HealthCheckHttpCode []*string `json:"HealthCheckHttpCode,omitempty" xml:"HealthCheckHttpCode,omitempty" type:"Repeated"`
	// The version of the HTTP protocol. Valid values: **HTTP1.0*	- and **HTTP1.1**.
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// HTTP1.0
	HealthCheckHttpVersion *string `json:"HealthCheckHttpVersion,omitempty" xml:"HealthCheckHttpVersion,omitempty"`
	// The interval at which health checks are performed. Unit: seconds. Default value: **5**.
	//
	// 	- If you set **HealthCheckType*	- to **TCP*	- or **HTTP**, valid values are **1 to 50**.
	//
	// 	- If you set **HealthCheckType*	- to **UDP**, valid values are **1 to 300**. Set the health check interval equal to or larger than the response timeout period to ensure that UDP response timeouts are not determined as no responses.
	//
	// example:
	//
	// 5
	HealthCheckInterval *int32 `json:"HealthCheckInterval,omitempty" xml:"HealthCheckInterval,omitempty"`
	// The request string of UDP health checks. The string must be 1 to 512 characters in length, and can contain letters and digits.
	//
	// example:
	//
	// hello
	HealthCheckReq *string `json:"HealthCheckReq,omitempty" xml:"HealthCheckReq,omitempty"`
	// The protocol that is used for health checks. Valid values:
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
	// The path to which health check probes are sent.
	//
	// The path must be 1 to 80 characters in length and can contain only letters, digits, and the following special characters: `- / . % ? # & =`. It can also contain the following extended characters: `_ ; ~ ! ( ) 	- [ ] @ $ ^ : \\" , +`. It must start with a forward slash (/).
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// /test/index.html
	HealthCheckUrl *string `json:"HealthCheckUrl,omitempty" xml:"HealthCheckUrl,omitempty"`
	// The number of times that an unhealthy backend server must consecutively pass health checks before it is declared healthy. In this case, the health status changes from **fail*	- to **success**. Valid values: **2*	- to **10**.
	//
	// example:
	//
	// 3
	HealthyThreshold *int32 `json:"HealthyThreshold,omitempty" xml:"HealthyThreshold,omitempty"`
	// The HTTP method used for health checks. Valid values: **GET*	- and **HEAD**.
	//
	// >  This parameter takes effect only if you set **HealthCheckType*	- to **HTTP**.
	//
	// example:
	//
	// GET
	HttpCheckMethod *string `json:"HttpCheckMethod,omitempty" xml:"HttpCheckMethod,omitempty"`
	// The number of times that a healthy backend server must consecutively fail health checks before it is declared unhealthy. In this case, the health status changes from **success*	- to **fail**. Valid values: **2*	- to **10**.
	//
	// example:
	//
	// 3
	UnhealthyThreshold *int32 `json:"UnhealthyThreshold,omitempty" xml:"UnhealthyThreshold,omitempty"`
}

func (s UpdateServerGroupAttributeRequestHealthCheckConfig) String() string {
	return dara.Prettify(s)
}

func (s UpdateServerGroupAttributeRequestHealthCheckConfig) GoString() string {
	return s.String()
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckConnectPort() *int32 {
	return s.HealthCheckConnectPort
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckConnectTimeout() *int32 {
	return s.HealthCheckConnectTimeout
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckDomain() *string {
	return s.HealthCheckDomain
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckEnabled() *bool {
	return s.HealthCheckEnabled
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckExp() *string {
	return s.HealthCheckExp
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckHttpCode() []*string {
	return s.HealthCheckHttpCode
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckHttpVersion() *string {
	return s.HealthCheckHttpVersion
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckInterval() *int32 {
	return s.HealthCheckInterval
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckReq() *string {
	return s.HealthCheckReq
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckType() *string {
	return s.HealthCheckType
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthCheckUrl() *string {
	return s.HealthCheckUrl
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHealthyThreshold() *int32 {
	return s.HealthyThreshold
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetHttpCheckMethod() *string {
	return s.HttpCheckMethod
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) GetUnhealthyThreshold() *int32 {
	return s.UnhealthyThreshold
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckConnectPort(v int32) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckConnectPort = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckConnectTimeout(v int32) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckConnectTimeout = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckDomain(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckDomain = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckEnabled(v bool) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckEnabled = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckExp(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckExp = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckHttpCode(v []*string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckHttpCode = v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckHttpVersion(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckHttpVersion = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckInterval(v int32) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckInterval = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckReq(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckReq = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckType(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckType = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthCheckUrl(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthCheckUrl = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHealthyThreshold(v int32) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HealthyThreshold = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetHttpCheckMethod(v string) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.HttpCheckMethod = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) SetUnhealthyThreshold(v int32) *UpdateServerGroupAttributeRequestHealthCheckConfig {
	s.UnhealthyThreshold = &v
	return s
}

func (s *UpdateServerGroupAttributeRequestHealthCheckConfig) Validate() error {
	return dara.Validate(s)
}
