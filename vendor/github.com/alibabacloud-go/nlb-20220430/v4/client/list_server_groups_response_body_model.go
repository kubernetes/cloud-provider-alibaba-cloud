// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListServerGroupsResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetMaxResults(v int32) *ListServerGroupsResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *ListServerGroupsResponseBody
	GetNextToken() *string
	SetRequestId(v string) *ListServerGroupsResponseBody
	GetRequestId() *string
	SetServerGroups(v []*ListServerGroupsResponseBodyServerGroups) *ListServerGroupsResponseBody
	GetServerGroups() []*ListServerGroupsResponseBodyServerGroups
	SetTotalCount(v int32) *ListServerGroupsResponseBody
	GetTotalCount() *int32
}

type ListServerGroupsResponseBody struct {
	// The number of entries per page. Valid values: **1*	- to **100**.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// A pagination token. It can be used in the next request to retrieve a new page of results. Valid values:
	//
	// 	- If **NextToken*	- is empty, no next page exists.
	//
	// 	- If a value is returned for **NextToken**, the value is the token that determines the start point of the next query.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The request ID.
	//
	// example:
	//
	// 54B28E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// A list of server groups.
	ServerGroups []*ListServerGroupsResponseBodyServerGroups `json:"ServerGroups,omitempty" xml:"ServerGroups,omitempty" type:"Repeated"`
	// The total number of entries returned.
	//
	// example:
	//
	// 1
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListServerGroupsResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsResponseBody) GoString() string {
	return s.String()
}

func (s *ListServerGroupsResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListServerGroupsResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *ListServerGroupsResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListServerGroupsResponseBody) GetServerGroups() []*ListServerGroupsResponseBodyServerGroups {
	return s.ServerGroups
}

func (s *ListServerGroupsResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *ListServerGroupsResponseBody) SetMaxResults(v int32) *ListServerGroupsResponseBody {
	s.MaxResults = &v
	return s
}

func (s *ListServerGroupsResponseBody) SetNextToken(v string) *ListServerGroupsResponseBody {
	s.NextToken = &v
	return s
}

func (s *ListServerGroupsResponseBody) SetRequestId(v string) *ListServerGroupsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListServerGroupsResponseBody) SetServerGroups(v []*ListServerGroupsResponseBodyServerGroups) *ListServerGroupsResponseBody {
	s.ServerGroups = v
	return s
}

func (s *ListServerGroupsResponseBody) SetTotalCount(v int32) *ListServerGroupsResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListServerGroupsResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListServerGroupsResponseBodyServerGroups struct {
	// The IP version. Valid values:
	//
	// 	- **ipv4**
	//
	// 	- **DualStack**
	//
	// example:
	//
	// ipv4
	AddressIPVersion *string `json:"AddressIPVersion,omitempty" xml:"AddressIPVersion,omitempty"`
	// The ID of the Alibaba Cloud account.
	//
	// example:
	//
	// 165820696622****
	AliUid *int64 `json:"AliUid,omitempty" xml:"AliUid,omitempty"`
	// Indicates whether the feature of forwarding requests to all ports is enabled. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	AnyPortEnabled *bool `json:"AnyPortEnabled,omitempty" xml:"AnyPortEnabled,omitempty"`
	// Indicates whether connection draining is enabled. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// example:
	//
	// false
	ConnectionDrainEnabled *bool `json:"ConnectionDrainEnabled,omitempty" xml:"ConnectionDrainEnabled,omitempty"`
	// The timeout period of connection draining. Unit: seconds. Valid values: **10*	- to **900**.
	//
	// example:
	//
	// 200
	ConnectionDrainTimeout *int32 `json:"ConnectionDrainTimeout,omitempty" xml:"ConnectionDrainTimeout,omitempty"`
	// The configurations of health checks.
	HealthCheck *ListServerGroupsResponseBodyServerGroupsHealthCheck `json:"HealthCheck,omitempty" xml:"HealthCheck,omitempty" type:"Struct"`
	// Indicates whether client IP preservation is enabled. Valid values:
	//
	// 	- **true**
	//
	// 	- **false**
	//
	// > This parameter is set to **true*	- by default when **AddressIPVersion*	- is set to **ipv4**. This parameter is set to **false*	- when **AddressIPVersion*	- is set to **ipv6**. **true*	- will be supported by later versions.
	//
	// example:
	//
	// true
	PreserveClientIpEnabled *bool `json:"PreserveClientIpEnabled,omitempty" xml:"PreserveClientIpEnabled,omitempty"`
	// The backend protocol. Valid values: **TCP*	- and **UDP**.
	//
	// example:
	//
	// TCP
	Protocol *string `json:"Protocol,omitempty" xml:"Protocol,omitempty"`
	// The region ID of the NLB instance.
	//
	// example:
	//
	// cn-hangzhou
	RegionId *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	// The NLB instances.
	RelatedLoadBalancerIds []*string `json:"RelatedLoadBalancerIds,omitempty" xml:"RelatedLoadBalancerIds,omitempty" type:"Repeated"`
	// The ID of the resource group to which the server group belongs.
	//
	// example:
	//
	// rg-atstuj3rtop****
	ResourceGroupId *string `json:"ResourceGroupId,omitempty" xml:"ResourceGroupId,omitempty"`
	// The routing algorithm. Valid values:
	//
	// 	- **Wrr**: Backend servers with higher weights receive more requests than backend servers with lower weights.
	//
	// 	- **rr**: Requests are forwarded to the backend servers in sequence. sch: Requests are forwarded to the backend servers based on source IP address hashing.
	//
	// 	- **sch**: Requests from the same source IP address are forwarded to the same backend server.
	//
	// 	- **tch**: Four-element hashing, which specifies consistent hashing that is based on four factors: source IP address, destination IP address, source port, and destination port. Requests that contain the same information based on the four factors are forwarded to the same backend server.
	//
	// 	- **qch**: QUIC ID hashing. Requests that contain the same QUIC ID are forwarded to the same backend server.
	//
	// example:
	//
	// Wrr
	Scheduler *string `json:"Scheduler,omitempty" xml:"Scheduler,omitempty"`
	// The number of server groups associated with the NLB instances.
	//
	// example:
	//
	// 2
	ServerCount *int32 `json:"ServerCount,omitempty" xml:"ServerCount,omitempty"`
	// The server group ID.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The server group name.
	//
	// example:
	//
	// NLB_ServerGroup
	ServerGroupName *string `json:"ServerGroupName,omitempty" xml:"ServerGroupName,omitempty"`
	// The status of the server group. Valid values:
	//
	// 	- **Creating**
	//
	// 	- **Available**
	//
	// 	- **Configuring**
	//
	// example:
	//
	// Available
	ServerGroupStatus *string `json:"ServerGroupStatus,omitempty" xml:"ServerGroupStatus,omitempty"`
	// The type of server group. Valid values:
	//
	// 	- **Instance*	- : contains servers of the **Ecs**, **Ens**, and **Eci*	- types.
	//
	// 	- **Ip**: contains servers specified by IP addresses.
	//
	// example:
	//
	// Instance
	ServerGroupType *string `json:"ServerGroupType,omitempty" xml:"ServerGroupType,omitempty"`
	// The tag.
	Tags []*ListServerGroupsResponseBodyServerGroupsTags `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
	// The ID of the VPC to which the server group belongs.
	//
	// example:
	//
	// vpc-bp15zckdt37pq72zv****
	VpcId *string `json:"VpcId,omitempty" xml:"VpcId,omitempty"`
}

func (s ListServerGroupsResponseBodyServerGroups) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsResponseBodyServerGroups) GoString() string {
	return s.String()
}

func (s *ListServerGroupsResponseBodyServerGroups) GetAddressIPVersion() *string {
	return s.AddressIPVersion
}

func (s *ListServerGroupsResponseBodyServerGroups) GetAliUid() *int64 {
	return s.AliUid
}

func (s *ListServerGroupsResponseBodyServerGroups) GetAnyPortEnabled() *bool {
	return s.AnyPortEnabled
}

func (s *ListServerGroupsResponseBodyServerGroups) GetConnectionDrainEnabled() *bool {
	return s.ConnectionDrainEnabled
}

func (s *ListServerGroupsResponseBodyServerGroups) GetConnectionDrainTimeout() *int32 {
	return s.ConnectionDrainTimeout
}

func (s *ListServerGroupsResponseBodyServerGroups) GetHealthCheck() *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	return s.HealthCheck
}

func (s *ListServerGroupsResponseBodyServerGroups) GetPreserveClientIpEnabled() *bool {
	return s.PreserveClientIpEnabled
}

func (s *ListServerGroupsResponseBodyServerGroups) GetProtocol() *string {
	return s.Protocol
}

func (s *ListServerGroupsResponseBodyServerGroups) GetRegionId() *string {
	return s.RegionId
}

func (s *ListServerGroupsResponseBodyServerGroups) GetRelatedLoadBalancerIds() []*string {
	return s.RelatedLoadBalancerIds
}

func (s *ListServerGroupsResponseBodyServerGroups) GetResourceGroupId() *string {
	return s.ResourceGroupId
}

func (s *ListServerGroupsResponseBodyServerGroups) GetScheduler() *string {
	return s.Scheduler
}

func (s *ListServerGroupsResponseBodyServerGroups) GetServerCount() *int32 {
	return s.ServerCount
}

func (s *ListServerGroupsResponseBodyServerGroups) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *ListServerGroupsResponseBodyServerGroups) GetServerGroupName() *string {
	return s.ServerGroupName
}

func (s *ListServerGroupsResponseBodyServerGroups) GetServerGroupStatus() *string {
	return s.ServerGroupStatus
}

func (s *ListServerGroupsResponseBodyServerGroups) GetServerGroupType() *string {
	return s.ServerGroupType
}

func (s *ListServerGroupsResponseBodyServerGroups) GetTags() []*ListServerGroupsResponseBodyServerGroupsTags {
	return s.Tags
}

func (s *ListServerGroupsResponseBodyServerGroups) GetVpcId() *string {
	return s.VpcId
}

func (s *ListServerGroupsResponseBodyServerGroups) SetAddressIPVersion(v string) *ListServerGroupsResponseBodyServerGroups {
	s.AddressIPVersion = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetAliUid(v int64) *ListServerGroupsResponseBodyServerGroups {
	s.AliUid = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetAnyPortEnabled(v bool) *ListServerGroupsResponseBodyServerGroups {
	s.AnyPortEnabled = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetConnectionDrainEnabled(v bool) *ListServerGroupsResponseBodyServerGroups {
	s.ConnectionDrainEnabled = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetConnectionDrainTimeout(v int32) *ListServerGroupsResponseBodyServerGroups {
	s.ConnectionDrainTimeout = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetHealthCheck(v *ListServerGroupsResponseBodyServerGroupsHealthCheck) *ListServerGroupsResponseBodyServerGroups {
	s.HealthCheck = v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetPreserveClientIpEnabled(v bool) *ListServerGroupsResponseBodyServerGroups {
	s.PreserveClientIpEnabled = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetProtocol(v string) *ListServerGroupsResponseBodyServerGroups {
	s.Protocol = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetRegionId(v string) *ListServerGroupsResponseBodyServerGroups {
	s.RegionId = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetRelatedLoadBalancerIds(v []*string) *ListServerGroupsResponseBodyServerGroups {
	s.RelatedLoadBalancerIds = v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetResourceGroupId(v string) *ListServerGroupsResponseBodyServerGroups {
	s.ResourceGroupId = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetScheduler(v string) *ListServerGroupsResponseBodyServerGroups {
	s.Scheduler = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetServerCount(v int32) *ListServerGroupsResponseBodyServerGroups {
	s.ServerCount = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetServerGroupId(v string) *ListServerGroupsResponseBodyServerGroups {
	s.ServerGroupId = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetServerGroupName(v string) *ListServerGroupsResponseBodyServerGroups {
	s.ServerGroupName = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetServerGroupStatus(v string) *ListServerGroupsResponseBodyServerGroups {
	s.ServerGroupStatus = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetServerGroupType(v string) *ListServerGroupsResponseBodyServerGroups {
	s.ServerGroupType = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetTags(v []*ListServerGroupsResponseBodyServerGroupsTags) *ListServerGroupsResponseBodyServerGroups {
	s.Tags = v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) SetVpcId(v string) *ListServerGroupsResponseBodyServerGroups {
	s.VpcId = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroups) Validate() error {
	return dara.Validate(s)
}

type ListServerGroupsResponseBodyServerGroupsHealthCheck struct {
	// The backend port that is used for health checks.
	//
	// Valid values: **0*	- to **65535**.
	//
	// A value of **0*	- indicates that the port on a backend server is used for health checks.
	//
	// example:
	//
	// 200
	HealthCheckConnectPort *int32 `json:"HealthCheckConnectPort,omitempty" xml:"HealthCheckConnectPort,omitempty"`
	// The maximum timeout period of a health check response. Unit: seconds. Default value: **5**.
	//
	// Valid values: **1*	- to **300**
	//
	// example:
	//
	// 200
	HealthCheckConnectTimeout *int32 `json:"HealthCheckConnectTimeout,omitempty" xml:"HealthCheckConnectTimeout,omitempty"`
	// The domain name that you want to use for health checks. Valid values:
	//
	// 	- **$SERVER_IP**: the private IP address of a backend server.
	//
	// 	- **domain**: a specified domain name. The domain name must be 1 to 80 characters in length, and can contain lowercase letters, digits, hyphens (-), and periods (.).
	//
	// > This parameter takes effect only when **HealthCheckType*	- is set to **HTTP**.
	//
	// example:
	//
	// $SERVER_IP
	HealthCheckDomain *string `json:"HealthCheckDomain,omitempty" xml:"HealthCheckDomain,omitempty"`
	// Indicates whether the health check feature is enabled. Valid values:
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
	// The HTTP status codes returned for health checks. Multiple HTTP status codes are separated by commas (,). Valid values: **http_2xx**, **http_3xx**, **http_4xx**, and **http_5xx**.
	//
	// > This parameter takes effect only when **HealthCheckType*	- is set to **HTTP**.
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
	// 200
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
	// > This parameter takes effect only when **HealthCheckType*	- is set to **HTTP**.
	//
	// example:
	//
	// /test/index.html
	HealthCheckUrl *string `json:"HealthCheckUrl,omitempty" xml:"HealthCheckUrl,omitempty"`
	// The number of times that an unhealthy backend server must consecutively pass health checks before it is declared healthy. In this case, the health status changes from **fail*	- to **success**.
	//
	// Valid values: **2*	- to **10**.
	//
	// example:
	//
	// 2
	HealthyThreshold *int32 `json:"HealthyThreshold,omitempty" xml:"HealthyThreshold,omitempty"`
	// The HTTP method that is used for health checks. Valid values: **GET*	- and **HEAD**.
	//
	// > This parameter takes effect only when **HealthCheckType*	- is set to **HTTP**.
	//
	// example:
	//
	// GET
	HttpCheckMethod *string `json:"HttpCheckMethod,omitempty" xml:"HttpCheckMethod,omitempty"`
	// The number of times that a healthy backend server must consecutively fail health checks before it is declared unhealthy. In this case, the health status changes from **success*	- to **fail**.
	//
	// Valid values: **2*	- to **10**.
	//
	// example:
	//
	// 3
	UnhealthyThreshold *int32 `json:"UnhealthyThreshold,omitempty" xml:"UnhealthyThreshold,omitempty"`
}

func (s ListServerGroupsResponseBodyServerGroupsHealthCheck) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsResponseBodyServerGroupsHealthCheck) GoString() string {
	return s.String()
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckConnectPort() *int32 {
	return s.HealthCheckConnectPort
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckConnectTimeout() *int32 {
	return s.HealthCheckConnectTimeout
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckDomain() *string {
	return s.HealthCheckDomain
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckEnabled() *bool {
	return s.HealthCheckEnabled
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckExp() *string {
	return s.HealthCheckExp
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckHttpCode() []*string {
	return s.HealthCheckHttpCode
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckHttpVersion() *string {
	return s.HealthCheckHttpVersion
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckInterval() *int32 {
	return s.HealthCheckInterval
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckReq() *string {
	return s.HealthCheckReq
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckType() *string {
	return s.HealthCheckType
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthCheckUrl() *string {
	return s.HealthCheckUrl
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHealthyThreshold() *int32 {
	return s.HealthyThreshold
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetHttpCheckMethod() *string {
	return s.HttpCheckMethod
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) GetUnhealthyThreshold() *int32 {
	return s.UnhealthyThreshold
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckConnectPort(v int32) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckConnectPort = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckConnectTimeout(v int32) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckConnectTimeout = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckDomain(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckDomain = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckEnabled(v bool) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckEnabled = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckExp(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckExp = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckHttpCode(v []*string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckHttpCode = v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckHttpVersion(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckHttpVersion = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckInterval(v int32) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckInterval = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckReq(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckReq = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckType(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckType = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthCheckUrl(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthCheckUrl = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHealthyThreshold(v int32) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HealthyThreshold = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetHttpCheckMethod(v string) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.HttpCheckMethod = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) SetUnhealthyThreshold(v int32) *ListServerGroupsResponseBodyServerGroupsHealthCheck {
	s.UnhealthyThreshold = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsHealthCheck) Validate() error {
	return dara.Validate(s)
}

type ListServerGroupsResponseBodyServerGroupsTags struct {
	// The tag key. At most 10 tag keys are returned.
	//
	// The tag key can be up to 64 characters in length, and cannot contain `http://` or `https://`. It cannot start with `aliyun` or `acs:`.
	//
	// example:
	//
	// Test
	Key *string `json:"Key,omitempty" xml:"Key,omitempty"`
	// The tag value. At most 10 tag values are returned.
	//
	// The tag value can be up to 128 characters in length, and cannot contain `http://` or `https://`. It cannot start with `aliyun` or `acs:`.
	//
	// example:
	//
	// Test
	Value *string `json:"Value,omitempty" xml:"Value,omitempty"`
}

func (s ListServerGroupsResponseBodyServerGroupsTags) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupsResponseBodyServerGroupsTags) GoString() string {
	return s.String()
}

func (s *ListServerGroupsResponseBodyServerGroupsTags) GetKey() *string {
	return s.Key
}

func (s *ListServerGroupsResponseBodyServerGroupsTags) GetValue() *string {
	return s.Value
}

func (s *ListServerGroupsResponseBodyServerGroupsTags) SetKey(v string) *ListServerGroupsResponseBodyServerGroupsTags {
	s.Key = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsTags) SetValue(v string) *ListServerGroupsResponseBodyServerGroupsTags {
	s.Value = &v
	return s
}

func (s *ListServerGroupsResponseBodyServerGroupsTags) Validate() error {
	return dara.Validate(s)
}
