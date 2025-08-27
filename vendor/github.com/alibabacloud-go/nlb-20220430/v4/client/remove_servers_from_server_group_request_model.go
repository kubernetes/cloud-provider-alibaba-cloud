// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iRemoveServersFromServerGroupRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *RemoveServersFromServerGroupRequest
	GetClientToken() *string
	SetDryRun(v bool) *RemoveServersFromServerGroupRequest
	GetDryRun() *bool
	SetRegionId(v string) *RemoveServersFromServerGroupRequest
	GetRegionId() *string
	SetServerGroupId(v string) *RemoveServersFromServerGroupRequest
	GetServerGroupId() *string
	SetServers(v []*RemoveServersFromServerGroupRequestServers) *RemoveServersFromServerGroupRequest
	GetServers() []*RemoveServersFromServerGroupRequestServers
}

type RemoveServersFromServerGroupRequest struct {
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
	// The server group ID.
	//
	// This parameter is required.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The backend servers. You can specify up to 200 backend servers in each call.
	//
	// This parameter is required.
	Servers []*RemoveServersFromServerGroupRequestServers `json:"Servers,omitempty" xml:"Servers,omitempty" type:"Repeated"`
}

func (s RemoveServersFromServerGroupRequest) String() string {
	return dara.Prettify(s)
}

func (s RemoveServersFromServerGroupRequest) GoString() string {
	return s.String()
}

func (s *RemoveServersFromServerGroupRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *RemoveServersFromServerGroupRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *RemoveServersFromServerGroupRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *RemoveServersFromServerGroupRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *RemoveServersFromServerGroupRequest) GetServers() []*RemoveServersFromServerGroupRequestServers {
	return s.Servers
}

func (s *RemoveServersFromServerGroupRequest) SetClientToken(v string) *RemoveServersFromServerGroupRequest {
	s.ClientToken = &v
	return s
}

func (s *RemoveServersFromServerGroupRequest) SetDryRun(v bool) *RemoveServersFromServerGroupRequest {
	s.DryRun = &v
	return s
}

func (s *RemoveServersFromServerGroupRequest) SetRegionId(v string) *RemoveServersFromServerGroupRequest {
	s.RegionId = &v
	return s
}

func (s *RemoveServersFromServerGroupRequest) SetServerGroupId(v string) *RemoveServersFromServerGroupRequest {
	s.ServerGroupId = &v
	return s
}

func (s *RemoveServersFromServerGroupRequest) SetServers(v []*RemoveServersFromServerGroupRequestServers) *RemoveServersFromServerGroupRequest {
	s.Servers = v
	return s
}

func (s *RemoveServersFromServerGroupRequest) Validate() error {
	return dara.Validate(s)
}

type RemoveServersFromServerGroupRequestServers struct {
	// The port that is used by the backend server. Valid values: **0*	- to **65535**. If you do not set this parameter, the default value **0*	- is used.
	//
	// if can be null:
	// true
	//
	// example:
	//
	// 443
	Port *int32 `json:"Port,omitempty" xml:"Port,omitempty"`
	// The backend server ID.
	//
	// 	- If the server group is of the **Instance*	- type, set this parameter to the IDs of servers of the **Ecs**, **Eni**, or **Eci*	- type.
	//
	// 	- If the server group is of the **Ip*	- type, set ServerId to IP addresses.
	//
	// This parameter is required.
	//
	// example:
	//
	// ecs-bp67acfmxazb4p****
	ServerId *string `json:"ServerId,omitempty" xml:"ServerId,omitempty"`
	// The IP addresses of the server. If the server group type is **Ip**, set the ServerId parameter to IP addresses.
	//
	// example:
	//
	// 192.168.6.6
	ServerIp *string `json:"ServerIp,omitempty" xml:"ServerIp,omitempty"`
	// The type of the backend server. Valid values:
	//
	// 	- **Ecs**: the Elastic Compute Service (ECS) instance
	//
	// 	- **Eni**: the elastic network interface (ENI)
	//
	// 	- **Eci**: the elastic container instance
	//
	// 	- **Ip**: the IP address
	//
	// This parameter is required.
	//
	// example:
	//
	// Ecs
	ServerType *string `json:"ServerType,omitempty" xml:"ServerType,omitempty"`
}

func (s RemoveServersFromServerGroupRequestServers) String() string {
	return dara.Prettify(s)
}

func (s RemoveServersFromServerGroupRequestServers) GoString() string {
	return s.String()
}

func (s *RemoveServersFromServerGroupRequestServers) GetPort() *int32 {
	return s.Port
}

func (s *RemoveServersFromServerGroupRequestServers) GetServerId() *string {
	return s.ServerId
}

func (s *RemoveServersFromServerGroupRequestServers) GetServerIp() *string {
	return s.ServerIp
}

func (s *RemoveServersFromServerGroupRequestServers) GetServerType() *string {
	return s.ServerType
}

func (s *RemoveServersFromServerGroupRequestServers) SetPort(v int32) *RemoveServersFromServerGroupRequestServers {
	s.Port = &v
	return s
}

func (s *RemoveServersFromServerGroupRequestServers) SetServerId(v string) *RemoveServersFromServerGroupRequestServers {
	s.ServerId = &v
	return s
}

func (s *RemoveServersFromServerGroupRequestServers) SetServerIp(v string) *RemoveServersFromServerGroupRequestServers {
	s.ServerIp = &v
	return s
}

func (s *RemoveServersFromServerGroupRequestServers) SetServerType(v string) *RemoveServersFromServerGroupRequestServers {
	s.ServerType = &v
	return s
}

func (s *RemoveServersFromServerGroupRequestServers) Validate() error {
	return dara.Validate(s)
}
