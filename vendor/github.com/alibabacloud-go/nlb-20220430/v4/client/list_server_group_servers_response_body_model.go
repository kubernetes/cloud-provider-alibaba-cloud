// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iListServerGroupServersResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetMaxResults(v int32) *ListServerGroupServersResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *ListServerGroupServersResponseBody
	GetNextToken() *string
	SetRequestId(v string) *ListServerGroupServersResponseBody
	GetRequestId() *string
	SetServers(v []*ListServerGroupServersResponseBodyServers) *ListServerGroupServersResponseBody
	GetServers() []*ListServerGroupServersResponseBodyServers
	SetTotalCount(v int32) *ListServerGroupServersResponseBody
	GetTotalCount() *int32
}

type ListServerGroupServersResponseBody struct {
	// The number of entries returned per page.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The token that is used for the next query. Valid values:
	//
	// 	- If this is your first query or no next query is to be sent, ignore this parameter.
	//
	// 	- If a next query is to be sent, set the parameter to the value of NextToken that is returned from the last call.
	//
	// example:
	//
	// FFmyTO70tTpLG6I3FmYAXGKPd****
	NextToken *string `json:"NextToken,omitempty" xml:"NextToken,omitempty"`
	// The ID of the request.
	//
	// example:
	//
	// 54B48E3D-DF70-471B-AA93-08E683A1B45
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	// The backend servers.
	Servers []*ListServerGroupServersResponseBodyServers `json:"Servers,omitempty" xml:"Servers,omitempty" type:"Repeated"`
	// The number of entries returned.
	//
	// example:
	//
	// 10
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListServerGroupServersResponseBody) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupServersResponseBody) GoString() string {
	return s.String()
}

func (s *ListServerGroupServersResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *ListServerGroupServersResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *ListServerGroupServersResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *ListServerGroupServersResponseBody) GetServers() []*ListServerGroupServersResponseBodyServers {
	return s.Servers
}

func (s *ListServerGroupServersResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *ListServerGroupServersResponseBody) SetMaxResults(v int32) *ListServerGroupServersResponseBody {
	s.MaxResults = &v
	return s
}

func (s *ListServerGroupServersResponseBody) SetNextToken(v string) *ListServerGroupServersResponseBody {
	s.NextToken = &v
	return s
}

func (s *ListServerGroupServersResponseBody) SetRequestId(v string) *ListServerGroupServersResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListServerGroupServersResponseBody) SetServers(v []*ListServerGroupServersResponseBodyServers) *ListServerGroupServersResponseBody {
	s.Servers = v
	return s
}

func (s *ListServerGroupServersResponseBody) SetTotalCount(v int32) *ListServerGroupServersResponseBody {
	s.TotalCount = &v
	return s
}

func (s *ListServerGroupServersResponseBody) Validate() error {
	return dara.Validate(s)
}

type ListServerGroupServersResponseBodyServers struct {
	// The description of the backend server.
	//
	// example:
	//
	// ECS
	Description *string `json:"Description,omitempty" xml:"Description,omitempty"`
	// The port that is used by the backend server. Valid values: **1*	- to **65535**.
	//
	// example:
	//
	// 80
	Port *int32 `json:"Port,omitempty" xml:"Port,omitempty"`
	// The ID of the server group.
	//
	// example:
	//
	// sgp-atstuj3rtoptyui****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
	// The ID of the server group.
	//
	// example:
	//
	// i-bp67acfmxazb4p****
	ServerId *string `json:"ServerId,omitempty" xml:"ServerId,omitempty"`
	// The IP address of the backend server.
	//
	// example:
	//
	// 192.168.2.1
	ServerIp *string `json:"ServerIp,omitempty" xml:"ServerIp,omitempty"`
	// The type of backend server. Valid values:
	//
	// 	- **Ecs**: Elastic Compute Service (ECS) instance
	//
	// 	- **Eni**: elastic network interface (ENI)
	//
	// 	- **Eci**: elastic container instance
	//
	// 	- **Ip**: IP address
	//
	// example:
	//
	// Ecs
	ServerType *string `json:"ServerType,omitempty" xml:"ServerType,omitempty"`
	// The status of the backend server. Valid values:
	//
	// 	- **Adding**
	//
	// 	- **Available**
	//
	// 	- **Configuring**
	//
	// 	- **Removing**
	//
	// example:
	//
	// Available
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
	// The weight of the backend server.
	//
	// example:
	//
	// 100
	Weight *int32 `json:"Weight,omitempty" xml:"Weight,omitempty"`
	// The zone ID of the server.
	//
	// example:
	//
	// cn-hangzhou-a
	ZoneId *string `json:"ZoneId,omitempty" xml:"ZoneId,omitempty"`
}

func (s ListServerGroupServersResponseBodyServers) String() string {
	return dara.Prettify(s)
}

func (s ListServerGroupServersResponseBodyServers) GoString() string {
	return s.String()
}

func (s *ListServerGroupServersResponseBodyServers) GetDescription() *string {
	return s.Description
}

func (s *ListServerGroupServersResponseBodyServers) GetPort() *int32 {
	return s.Port
}

func (s *ListServerGroupServersResponseBodyServers) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *ListServerGroupServersResponseBodyServers) GetServerId() *string {
	return s.ServerId
}

func (s *ListServerGroupServersResponseBodyServers) GetServerIp() *string {
	return s.ServerIp
}

func (s *ListServerGroupServersResponseBodyServers) GetServerType() *string {
	return s.ServerType
}

func (s *ListServerGroupServersResponseBodyServers) GetStatus() *string {
	return s.Status
}

func (s *ListServerGroupServersResponseBodyServers) GetWeight() *int32 {
	return s.Weight
}

func (s *ListServerGroupServersResponseBodyServers) GetZoneId() *string {
	return s.ZoneId
}

func (s *ListServerGroupServersResponseBodyServers) SetDescription(v string) *ListServerGroupServersResponseBodyServers {
	s.Description = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetPort(v int32) *ListServerGroupServersResponseBodyServers {
	s.Port = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetServerGroupId(v string) *ListServerGroupServersResponseBodyServers {
	s.ServerGroupId = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetServerId(v string) *ListServerGroupServersResponseBodyServers {
	s.ServerId = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetServerIp(v string) *ListServerGroupServersResponseBodyServers {
	s.ServerIp = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetServerType(v string) *ListServerGroupServersResponseBodyServers {
	s.ServerType = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetStatus(v string) *ListServerGroupServersResponseBodyServers {
	s.Status = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetWeight(v int32) *ListServerGroupServersResponseBodyServers {
	s.Weight = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) SetZoneId(v string) *ListServerGroupServersResponseBodyServers {
	s.ZoneId = &v
	return s
}

func (s *ListServerGroupServersResponseBodyServers) Validate() error {
	return dara.Validate(s)
}
