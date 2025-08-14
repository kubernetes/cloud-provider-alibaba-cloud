// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"github.com/alibabacloud-go/tea/dara"
)

type iGetListenerHealthStatusResponseBody interface {
	dara.Model
	String() string
	GoString() string
	SetListenerHealthStatus(v []*GetListenerHealthStatusResponseBodyListenerHealthStatus) *GetListenerHealthStatusResponseBody
	GetListenerHealthStatus() []*GetListenerHealthStatusResponseBodyListenerHealthStatus
	SetMaxResults(v int32) *GetListenerHealthStatusResponseBody
	GetMaxResults() *int32
	SetNextToken(v string) *GetListenerHealthStatusResponseBody
	GetNextToken() *string
	SetRequestId(v string) *GetListenerHealthStatusResponseBody
	GetRequestId() *string
	SetTotalCount(v int32) *GetListenerHealthStatusResponseBody
	GetTotalCount() *int32
}

type GetListenerHealthStatusResponseBody struct {
	// The health check status of the server group of the listener.
	ListenerHealthStatus []*GetListenerHealthStatusResponseBodyListenerHealthStatus `json:"ListenerHealthStatus,omitempty" xml:"ListenerHealthStatus,omitempty" type:"Repeated"`
	// The number of entries returned per page.
	//
	// example:
	//
	// 20
	MaxResults *int32 `json:"MaxResults,omitempty" xml:"MaxResults,omitempty"`
	// The token that is used for the next query. Valid values:
	//
	// - If **NextToken*	- is empty, it indicates that no next query is to be sent.
	//
	// - If a value of **NextToken*	- is returned, the value is the token used for the next query.
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
	// The number of entries returned.
	//
	// example:
	//
	// 10
	TotalCount *int32 `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s GetListenerHealthStatusResponseBody) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusResponseBody) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusResponseBody) GetListenerHealthStatus() []*GetListenerHealthStatusResponseBodyListenerHealthStatus {
	return s.ListenerHealthStatus
}

func (s *GetListenerHealthStatusResponseBody) GetMaxResults() *int32 {
	return s.MaxResults
}

func (s *GetListenerHealthStatusResponseBody) GetNextToken() *string {
	return s.NextToken
}

func (s *GetListenerHealthStatusResponseBody) GetRequestId() *string {
	return s.RequestId
}

func (s *GetListenerHealthStatusResponseBody) GetTotalCount() *int32 {
	return s.TotalCount
}

func (s *GetListenerHealthStatusResponseBody) SetListenerHealthStatus(v []*GetListenerHealthStatusResponseBodyListenerHealthStatus) *GetListenerHealthStatusResponseBody {
	s.ListenerHealthStatus = v
	return s
}

func (s *GetListenerHealthStatusResponseBody) SetMaxResults(v int32) *GetListenerHealthStatusResponseBody {
	s.MaxResults = &v
	return s
}

func (s *GetListenerHealthStatusResponseBody) SetNextToken(v string) *GetListenerHealthStatusResponseBody {
	s.NextToken = &v
	return s
}

func (s *GetListenerHealthStatusResponseBody) SetRequestId(v string) *GetListenerHealthStatusResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetListenerHealthStatusResponseBody) SetTotalCount(v int32) *GetListenerHealthStatusResponseBody {
	s.TotalCount = &v
	return s
}

func (s *GetListenerHealthStatusResponseBody) Validate() error {
	return dara.Validate(s)
}

type GetListenerHealthStatusResponseBodyListenerHealthStatus struct {
	// The ID of the listener of the NLB instance.
	//
	// example:
	//
	// lsn-bp1bpn0kn908w4nbw****@80
	ListenerId *string `json:"ListenerId,omitempty" xml:"ListenerId,omitempty"`
	// The listening port.
	//
	// example:
	//
	// 80
	ListenerPort *int32 `json:"ListenerPort,omitempty" xml:"ListenerPort,omitempty"`
	// The listening protocol. Valid values: **TCP**, **UDP**, and **TCPSSL**.
	//
	// example:
	//
	// TCPSSL
	ListenerProtocol *string `json:"ListenerProtocol,omitempty" xml:"ListenerProtocol,omitempty"`
	// The information about the server groups.
	ServerGroupInfos []*GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos `json:"ServerGroupInfos,omitempty" xml:"ServerGroupInfos,omitempty" type:"Repeated"`
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatus) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatus) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) GetListenerId() *string {
	return s.ListenerId
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) GetListenerPort() *int32 {
	return s.ListenerPort
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) GetListenerProtocol() *string {
	return s.ListenerProtocol
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) GetServerGroupInfos() []*GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos {
	return s.ServerGroupInfos
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) SetListenerId(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatus {
	s.ListenerId = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) SetListenerPort(v int32) *GetListenerHealthStatusResponseBodyListenerHealthStatus {
	s.ListenerPort = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) SetListenerProtocol(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatus {
	s.ListenerProtocol = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) SetServerGroupInfos(v []*GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) *GetListenerHealthStatusResponseBodyListenerHealthStatus {
	s.ServerGroupInfos = v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatus) Validate() error {
	return dara.Validate(s)
}

type GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos struct {
	// Indicates whether the health check feature is enabled. Valid values:
	//
	// 	- **true**: enabled
	//
	// 	- **false**: disabled
	//
	// example:
	//
	// true
	HeathCheckEnabled *bool `json:"HeathCheckEnabled,omitempty" xml:"HeathCheckEnabled,omitempty"`
	// A list of unhealthy backend servers.
	NonNormalServers []*GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers `json:"NonNormalServers,omitempty" xml:"NonNormalServers,omitempty" type:"Repeated"`
	// The ID of the server group.
	//
	// example:
	//
	// sgp-ppdpc14gdm3x4o****
	ServerGroupId *string `json:"ServerGroupId,omitempty" xml:"ServerGroupId,omitempty"`
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) GetHeathCheckEnabled() *bool {
	return s.HeathCheckEnabled
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) GetNonNormalServers() []*GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers {
	return s.NonNormalServers
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) GetServerGroupId() *string {
	return s.ServerGroupId
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) SetHeathCheckEnabled(v bool) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos {
	s.HeathCheckEnabled = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) SetNonNormalServers(v []*GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos {
	s.NonNormalServers = v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) SetServerGroupId(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos {
	s.ServerGroupId = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfos) Validate() error {
	return dara.Validate(s)
}

type GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers struct {
	// The backend port.
	//
	// example:
	//
	// 80
	Port *int32 `json:"Port,omitempty" xml:"Port,omitempty"`
	// The cause of the health check failure.
	Reason *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason `json:"Reason,omitempty" xml:"Reason,omitempty" type:"Struct"`
	// The ID of the backend server.
	//
	// example:
	//
	// i-bp1bt75jaujl7tjl****
	ServerId *string `json:"ServerId,omitempty" xml:"ServerId,omitempty"`
	// The IP address of the backend server.
	//
	// example:
	//
	// 192.168.8.10
	ServerIp *string `json:"ServerIp,omitempty" xml:"ServerIp,omitempty"`
	// The health check status. Valid values:
	//
	// 	- **Initial**: indicates that health checks are configured for the NLB instance, but no data was found.
	//
	// 	- **Unhealthy**: indicates that the backend server consecutively fails health checks.
	//
	// 	- **Unavailable**: indicates that health checks are disabled.
	//
	// example:
	//
	// Initial
	Status *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) GetPort() *int32 {
	return s.Port
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) GetReason() *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason {
	return s.Reason
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) GetServerId() *string {
	return s.ServerId
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) GetServerIp() *string {
	return s.ServerIp
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) GetStatus() *string {
	return s.Status
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) SetPort(v int32) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers {
	s.Port = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) SetReason(v *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers {
	s.Reason = v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) SetServerId(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers {
	s.ServerId = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) SetServerIp(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers {
	s.ServerIp = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) SetStatus(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers {
	s.Status = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServers) Validate() error {
	return dara.Validate(s)
}

type GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason struct {
	// The reason why the **status*	- is abnormal. Valid values:
	//
	// 	- **CONNECT_TIMEOUT**: The NLB instance failed to connect to the backend server within the specified period of time.
	//
	// 	- **CONNECT_FAILED**: The NLB instance failed to connect to the backend server.
	//
	// 	- **RECV_RESPONSE_TIMEOUT**: The NLB instance failed to receive a response from the backend server within the specified period of time.
	//
	// 	- **CONNECT_INTERRUPT**: The connection between the health check and the backend servers was interrupted.
	//
	// 	- **HTTP_CODE_NOT_MATCH**: The HTTP status code from the backend servers was not the expected one.
	//
	// 	- **HTTP_INVALID_HEADER**: The format of the response from the backend servers is invalid.
	//
	// example:
	//
	// CONNECT_TIMEOUT
	ReasonCode *string `json:"ReasonCode,omitempty" xml:"ReasonCode,omitempty"`
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason) String() string {
	return dara.Prettify(s)
}

func (s GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason) GoString() string {
	return s.String()
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason) GetReasonCode() *string {
	return s.ReasonCode
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason) SetReasonCode(v string) *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason {
	s.ReasonCode = &v
	return s
}

func (s *GetListenerHealthStatusResponseBodyListenerHealthStatusServerGroupInfosNonNormalServersReason) Validate() error {
	return dara.Validate(s)
}
