// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iAddServersToServerGroupRequest interface {
	dara.Model
	String() string
	GoString() string
	SetClientToken(v string) *AddServersToServerGroupRequest
	GetClientToken() *string
	SetDryRun(v bool) *AddServersToServerGroupRequest
	GetDryRun() *bool
	SetRegionId(v string) *AddServersToServerGroupRequest
	GetRegionId() *string
	SetServerGroupId(v string) *AddServersToServerGroupRequest
	GetServerGroupId() *string
	SetServers(v []*AddServersToServerGroupRequestServers) *AddServersToServerGroupRequest
	GetServers() []*AddServersToServerGroupRequestServers
}

type AddServersToServerGroupRequest struct {
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
	// The backend servers that you want to add.
	//
	// >  You can add up to 200 backend servers in each call.
	//
	// This parameter is required.
	Servers []*AddServersToServerGroupRequestServers `json:"Servers,omitempty" xml:"Servers,omitempty" type:"Repeated"`
}

func (s AddServersToServerGroupRequest) String() string {
	return dara.Prettify(s)
}

func (s AddServersToServerGroupRequest) GoString() string {
	return s.String()
}

func (s *AddServersToServerGroupRequest) GetClientToken() *string {
	return s.ClientToken
}

func (s *AddServersToServerGroupRequest) GetDryRun() *bool {
	return s.DryRun
}

func (s *AddServersToServerGroupRequest) GetRegionId() *string {
	return s.RegionId
}

func (s *AddServersToServerGroupRequest) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *AddServersToServerGroupRequest) GetServers() []*AddServersToServerGroupRequestServers {
	return s.Servers
}

func (s *AddServersToServerGroupRequest) SetClientToken(v string) *AddServersToServerGroupRequest {
	s.ClientToken = &v
	return s
}

func (s *AddServersToServerGroupRequest) SetDryRun(v bool) *AddServersToServerGroupRequest {
	s.DryRun = &v
	return s
}

func (s *AddServersToServerGroupRequest) SetRegionId(v string) *AddServersToServerGroupRequest {
	s.RegionId = &v
	return s
}

func (s *AddServersToServerGroupRequest) SetServerGroupId(v string) *AddServersToServerGroupRequest {
	s.ServerGroupId = &v
	return s
}

func (s *AddServersToServerGroupRequest) SetServers(v []*AddServersToServerGroupRequestServers) *AddServersToServerGroupRequest {
	s.Servers = v
	return s
}

func (s *AddServersToServerGroupRequest) Validate() error {
	return dara.Validate(s)
}

type AddServersToServerGroupRequestServers struct {
	// The description of the backend server.
	//
	// The description must be 2 to 256 characters in length, and can contain letters, digits, commas (,), periods (.), semicolons (;), forward slashes (/), at sings (@), underscores (_), and hyphens (-).
	//
	// example:
	//
	// ECS
	Description *string `json:"Description,omitempty" xml:"Description,omitempty"`
	// The port that is used by the backend server to provide services. Valid values: **0 to 65535**. If you do not set this parameter, the default value **0*	- is used.
	//
	// If multi-port forwarding is enabled, you do not need to set this parameter. The default value 0 is used. NLB forwards requests to the requested ports. To determine whether multi-port forwarding is enabled, call the [ListServerGroups](https://help.aliyun.com/document_detail/445895.html) operation and check the value of the **AnyPortEnabled*	- parameter.
	//
	// example:
	//
	// 443
	Port *int32 `json:"Port,omitempty" xml:"Port,omitempty"`
	// The backend server ID.
	//
	// 	- If the server group is of the **Instance*	- type, set this parameter to the IDs of **Elastic Compute Service (ECS) instances**, **elastic network interfaces (ENIs)**, or **elastic container instances**.
	//
	// 	- If the server group is of the **Ip*	- type, set ServerId to IP addresses.
	//
	// This parameter is required.
	//
	// example:
	//
	// i-bp67acfmxazb4p****
	ServerId *string `json:"ServerId,omitempty" xml:"ServerId,omitempty"`
	// The IP address of the backend server. If the server group type is **Ip**, set the ServerId parameter to IP addresses.
	//
	// example:
	//
	// 192.168.6.6
	ServerIp *string `json:"ServerIp,omitempty" xml:"ServerIp,omitempty"`
	// The type of the backend server. Valid values:
	//
	// 	- **Ecs**: the ECS instance
	//
	// 	- **Eni**: the ENI
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
	// The weight of the backend server. Valid values: **0*	- to **100**. Default value: **100**. If this parameter is set to **0**, no requests are forwarded to the server.
	//
	// example:
	//
	// 100
	Weight *int32 `json:"Weight,omitempty" xml:"Weight,omitempty"`
}

func (s AddServersToServerGroupRequestServers) String() string {
	return dara.Prettify(s)
}

func (s AddServersToServerGroupRequestServers) GoString() string {
	return s.String()
}

func (s *AddServersToServerGroupRequestServers) GetDescription() *string {
	return s.Description
}

func (s *AddServersToServerGroupRequestServers) GetPort() *int32 {
	return s.Port
}

func (s *AddServersToServerGroupRequestServers) GetServerId() *string {
	return s.ServerId
}

func (s *AddServersToServerGroupRequestServers) GetServerIp() *string {
	return s.ServerIp
}

func (s *AddServersToServerGroupRequestServers) GetServerType() *string {
	return s.ServerType
}

func (s *AddServersToServerGroupRequestServers) GetWeight() *int32 {
	return s.Weight
}

func (s *AddServersToServerGroupRequestServers) SetDescription(v string) *AddServersToServerGroupRequestServers {
	s.Description = &v
	return s
}

func (s *AddServersToServerGroupRequestServers) SetPort(v int32) *AddServersToServerGroupRequestServers {
	s.Port = &v
	return s
}

func (s *AddServersToServerGroupRequestServers) SetServerId(v string) *AddServersToServerGroupRequestServers {
	s.ServerId = &v
	return s
}

func (s *AddServersToServerGroupRequestServers) SetServerIp(v string) *AddServersToServerGroupRequestServers {
	s.ServerIp = &v
	return s
}

func (s *AddServersToServerGroupRequestServers) SetServerType(v string) *AddServersToServerGroupRequestServers {
	s.ServerType = &v
	return s
}

func (s *AddServersToServerGroupRequestServers) SetWeight(v int32) *AddServersToServerGroupRequestServers {
	s.Weight = &v
	return s
}

func (s *AddServersToServerGroupRequestServers) Validate() error {
	return dara.Validate(s)
}
