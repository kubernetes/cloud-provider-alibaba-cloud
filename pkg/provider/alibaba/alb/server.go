package alb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
)

var registerServersFunc = func(ctx context.Context, serverMgr *ALBProvider, sgpID string, servers []albsdk.AddServersToServerGroupServers) error {
	if len(servers) == 0 {
		return nil
	}

	traceID := ctx.Value(util.TraceID)

	addServerToSgpReq := albsdk.CreateAddServersToServerGroupRequest()
	addServerToSgpReq.ServerGroupId = sgpID
	addServerToSgpReq.Servers = &servers

	startTime := time.Now()
	serverMgr.logger.V(util.MgrLogLevel).Info("adding server to server group",
		"serverGroupID", sgpID,
		"servers", servers,
		"traceID", traceID,
		"startTime", startTime,
		util.Action, util.AddALBServersToServerGroup)
	addServerToSgpResp, err := serverMgr.auth.ALB.AddServersToServerGroup(addServerToSgpReq)
	if err != nil {
		return err
	}
	serverMgr.logger.V(util.MgrLogLevel).Info("added server to server group",
		"serverGroupID", sgpID,
		"traceID", traceID,
		"requestID", addServerToSgpResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.AddALBServersToServerGroup)

	if util.IsWaitServersAsynchronousComplete {
		asynchronousStartTime := time.Now()
		serverMgr.logger.V(util.MgrLogLevel).Info("adding server to server group asynchronous",
			"serverGroupID", sgpID,
			"servers", servers,
			"traceID", traceID,
			"startTime", asynchronousStartTime,
			util.Action, util.AddALBServersToServerGroupAsynchronous)
		for i := 0; i < util.AddALBServersToServerGroupWaitAvailableMaxRetryTimes; i++ {
			time.Sleep(util.AddALBServersToServerGroupWaitAvailableRetryInterval)

			isCompleted, err := isRegisterServersCompleted(ctx, serverMgr, sgpID, servers)
			if err != nil {
				serverMgr.logger.V(util.MgrLogLevel).Error(err, "failed to add server to server group asynchronous",
					"serverGroupID", sgpID,
					"servers", servers,
					"traceID", traceID,
					util.Action, util.AddALBServersToServerGroupAsynchronous)
				return err
			}
			if isCompleted {
				break
			}
		}
		serverMgr.logger.V(util.MgrLogLevel).Info("added server to server group asynchronous",
			"serverGroupID", sgpID,
			"traceID", traceID,
			"requestID", addServerToSgpResp.RequestId,
			"elapsedTime", time.Since(asynchronousStartTime).Milliseconds(),
			util.Action, util.AddALBServersToServerGroupAsynchronous)
	}

	return nil
}

type BatchRegisterServersFunc func(context.Context, *ALBProvider, string, []albsdk.AddServersToServerGroupServers) error

func BatchRegisterServers(ctx context.Context, serverMgr *ALBProvider, sgpID string, servers []albsdk.AddServersToServerGroupServers, cnt int, batch BatchRegisterServersFunc) error {
	if cnt <= 0 || cnt >= util.BatchRegisterDeregisterServersMaxNum {
		cnt = util.BatchRegisterServersDefaultNum
	}

	for len(servers) > cnt {
		if err := batch(ctx, serverMgr, sgpID, servers[0:cnt]); err != nil {
			return err
		}
		servers = servers[cnt:]
	}
	if len(servers) <= 0 {
		return nil
	}

	return batch(ctx, serverMgr, sgpID, servers)
}

func (m *ALBProvider) RegisterALBServers(ctx context.Context, serverGroupID string, resServers []alb.BackendItem) error {
	if len(serverGroupID) == 0 {
		return fmt.Errorf("empty server group id when register servers error")
	}

	if len(resServers) == 0 {
		return nil
	}

	serversToAdd, err := transModelBackendsToSDKAddServersToServerGroupServers(resServers)
	if err != nil {
		return err
	}

	return BatchRegisterServers(ctx, m, serverGroupID, serversToAdd, util.BatchRegisterServersDefaultNum, registerServersFunc)
}

var deregisterServersFunc = func(ctx context.Context, serverMgr *ALBProvider, sgpID string, servers []albsdk.RemoveServersFromServerGroupServers) error {
	if len(servers) == 0 {
		return nil
	}

	traceID := ctx.Value(util.TraceID)

	removeServerFromSgpReq := albsdk.CreateRemoveServersFromServerGroupRequest()
	removeServerFromSgpReq.ServerGroupId = sgpID
	removeServerFromSgpReq.Servers = &servers

	startTime := time.Now()
	serverMgr.logger.V(util.MgrLogLevel).Info("removing server from server group",
		"serverGroupID", sgpID,
		"traceID", traceID,
		"servers", servers,
		"startTime", startTime,
		util.Action, util.RemoveALBServersFromServerGroup)
	removeServerFromSgpResp, err := serverMgr.auth.ALB.RemoveServersFromServerGroup(removeServerFromSgpReq)
	if err != nil {
		return err
	}
	serverMgr.logger.V(util.MgrLogLevel).Info("removed server from server group",
		"serverGroupID", sgpID,
		"traceID", traceID,
		"requestID", removeServerFromSgpResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.RemoveALBServersFromServerGroup)

	if util.IsWaitServersAsynchronousComplete {
		asynchronousStartTime := time.Now()
		serverMgr.logger.V(util.MgrLogLevel).Info("removing server from server group asynchronous",
			"serverGroupID", sgpID,
			"traceID", traceID,
			"servers", servers,
			"startTime", startTime,
			util.Action, util.RemoveALBServersFromServerGroupAsynchronous)
		for i := 0; i < util.RemoveALBServersFromServerGroupMaxRetryTimes; i++ {
			time.Sleep(util.RemoveALBServersFromServerGroupRetryInterval)

			isCompleted, err := isDeregisterServersCompleted(ctx, serverMgr, sgpID, servers)
			if err != nil {
				serverMgr.logger.V(util.MgrLogLevel).Error(err, "failed to remove server from server group asynchronous",
					"serverGroupID", sgpID,
					"traceID", traceID,
					"requestID", removeServerFromSgpResp.RequestId,
					"elapsedTime", time.Since(asynchronousStartTime).Milliseconds(),
					util.Action, util.RemoveALBServersFromServerGroupAsynchronous)
				return err
			}
			if isCompleted {
				break
			}
		}
		serverMgr.logger.V(util.MgrLogLevel).Info("removed server from server group asynchronous",
			"serverGroupID", sgpID,
			"traceID", traceID,
			"requestID", removeServerFromSgpResp.RequestId,
			"elapsedTime", time.Since(asynchronousStartTime).Milliseconds(),
			util.Action, util.RemoveALBServersFromServerGroupAsynchronous)
	}

	return nil
}

type DeregisterServersFunc func(context.Context, *ALBProvider, string, []albsdk.RemoveServersFromServerGroupServers) error

func BatchDeregisterServers(ctx context.Context, serverMgr *ALBProvider, sgpID string, servers []albsdk.RemoveServersFromServerGroupServers, cnt int, batch DeregisterServersFunc) error {
	if cnt <= 0 || cnt >= util.BatchRegisterDeregisterServersMaxNum {
		cnt = util.BatchRegisterServersDefaultNum
	}

	for len(servers) > cnt {
		if err := batch(ctx, serverMgr, sgpID, servers[0:cnt]); err != nil {
			return err
		}
		servers = servers[cnt:]
	}
	if len(servers) <= 0 {
		return nil
	}

	return batch(ctx, serverMgr, sgpID, servers)
}

func (m *ALBProvider) DeregisterALBServers(ctx context.Context, serverGroupID string, sdkServers []albsdk.BackendServer) error {
	if len(serverGroupID) == 0 {
		return fmt.Errorf("empty server group id when deregister servers error")
	}

	if len(sdkServers) == 0 {
		return nil
	}

	serversToRemove := make([]albsdk.RemoveServersFromServerGroupServers, 0)
	for _, sdkServer := range sdkServers {
		if isServerStatusRemoving(sdkServer.Status) {
			continue
		}
		serverToRemove, err := transSDKBackendServerToRemoveServersFromServerGroupServer(sdkServer)
		if err != nil {
			return err
		}
		serversToRemove = append(serversToRemove, *serverToRemove)
	}

	if len(serversToRemove) == 0 {
		return nil
	}

	return BatchDeregisterServers(ctx, m, serverGroupID, serversToRemove, util.BatchDeregisterServersDefaultNum, deregisterServersFunc)
}

func (m *ALBProvider) ReplaceALBServers(ctx context.Context, serverGroupID string, resServers []alb.BackendItem, sdkServers []albsdk.BackendServer) error {
	if len(serverGroupID) == 0 {
		return fmt.Errorf("empty server group id when replace servers error")
	}

	traceID := ctx.Value(util.TraceID)

	if len(resServers) == 0 && len(sdkServers) == 0 {
		return nil
	}

	addedServers, err := transModelBackendsToSDKReplaceServersInServerGroupAddedServers(resServers)
	if err != nil {
		return err
	}

	removedServers := make([]albsdk.ReplaceServersInServerGroupRemovedServers, 0)
	for _, sdkServer := range sdkServers {
		if isServerStatusRemoving(sdkServer.Status) {
			continue
		}
		serverToRemove, err := transSDKBackendServerToReplaceServersInServerGroupRemovedServer(sdkServer)
		if err != nil {
			return err
		}
		removedServers = append(removedServers, *serverToRemove)
	}

	replaceServerFromSgpReq := albsdk.CreateReplaceServersInServerGroupRequest()
	replaceServerFromSgpReq.ServerGroupId = serverGroupID
	replaceServerFromSgpReq.AddedServers = &addedServers
	replaceServerFromSgpReq.RemovedServers = &removedServers

	startTime := time.Now()
	m.logger.V(util.MgrLogLevel).Info("replacing server in server group",
		"serverGroupID", serverGroupID,
		"traceID", traceID,
		"addedServers", addedServers,
		"removedServers", removedServers,
		"startTime", startTime,
		util.Action, util.ReplaceALBServersInServerGroup)
	replaceServerFromSgpResp, err := m.auth.ALB.ReplaceServersInServerGroup(replaceServerFromSgpReq)
	if err != nil {
		return err
	}
	m.logger.V(util.MgrLogLevel).Info("replaced server in server group",
		"serverGroupID", serverGroupID,
		"traceID", traceID,
		"requestID", replaceServerFromSgpResp.RequestId,
		"elapsedTime", time.Since(startTime).Milliseconds(),
		util.Action, util.ReplaceALBServersInServerGroup)

	if util.IsWaitServersAsynchronousComplete {
		asynchronousStartTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("replacing server in server group asynchronous",
			"serverGroupID", serverGroupID,
			"traceID", traceID,
			"addedServers", addedServers,
			"removedServers", removedServers,
			"startTime", startTime,
			util.Action, util.ReplaceALBServersInServerGroupAsynchronous)
		for i := 0; i < util.ReplaceALBServersInServerGroupMaxRetryTimes; i++ {
			time.Sleep(util.ReplaceALBServersInServerGroupRetryInterval)

			isCompleted, err := isReplaceServersCompleted(ctx, m, serverGroupID, addedServers, removedServers)
			if err != nil {
				m.logger.V(util.MgrLogLevel).Error(err, "failed to replace server in server group asynchronous",
					"serverGroupID", serverGroupID,
					"traceID", traceID,
					"requestID", replaceServerFromSgpResp.RequestId,
					util.Action, util.ReplaceALBServersInServerGroupAsynchronous)
				return err
			}
			if isCompleted {
				break
			}
		}
		m.logger.V(util.MgrLogLevel).Info("replaced server in server group asynchronous",
			"serverGroupID", serverGroupID,
			"traceID", traceID,
			"requestID", replaceServerFromSgpResp.RequestId,
			"elapsedTime", time.Since(asynchronousStartTime).Milliseconds(),
			util.Action, util.ReplaceALBServersInServerGroupAsynchronous)
	}

	return nil
}

func (m *ALBProvider) ListALBServers(ctx context.Context, serverGroupID string) ([]albsdk.BackendServer, error) {
	if len(serverGroupID) == 0 {
		return nil, fmt.Errorf("empty server group id when list servers error")
	}

	traceID := ctx.Value(util.TraceID)

	var (
		nextToken string
		servers   []albsdk.BackendServer
	)

	listSgpServersReq := albsdk.CreateListServerGroupServersRequest()
	listSgpServersReq.ServerGroupId = serverGroupID

	for {
		listSgpServersReq.NextToken = nextToken

		startTime := time.Now()
		m.logger.V(util.MgrLogLevel).Info("listing servers",
			"serverGroupID", serverGroupID,
			"traceID", traceID,
			"startTime", startTime,
			util.Action, util.ListALBServerGroupServers)
		listSgpServersResp, err := m.auth.ALB.ListServerGroupServers(listSgpServersReq)
		if err != nil {
			return nil, err
		}
		m.logger.V(util.MgrLogLevel).Info("listed servers",
			"serverGroupID", serverGroupID,
			"traceID", traceID,
			"servers", listSgpServersResp.Servers,
			"requestID", listSgpServersResp.RequestId,
			"elapsedTime", time.Since(startTime).Milliseconds(),
			util.Action, util.ListALBServerGroupServers)

		servers = append(servers, listSgpServersResp.Servers...)

		if listSgpServersResp.NextToken == "" {
			break
		} else {
			nextToken = listSgpServersResp.NextToken
		}
	}

	return servers, nil
}

func isServerStatusRemoving(status string) bool {
	return strings.EqualFold(status, util.ServerStatusRemoving)
}

func transSDKBackendServerToRemoveServersFromServerGroupServer(server albsdk.BackendServer) (*albsdk.RemoveServersFromServerGroupServers, error) {
	serverToRemove := new(albsdk.RemoveServersFromServerGroupServers)

	serverToRemove.ServerIp = server.ServerIp
	if len(server.ServerId) == 0 {
		return nil, fmt.Errorf("invalid server id for server: %v", server)
	}
	serverToRemove.ServerId = server.ServerId

	if !isServerPortValid(server.Port) {
		return nil, fmt.Errorf("invalid server port for server: %v", server)
	}
	serverToRemove.Port = strconv.Itoa(server.Port)

	if !isServerTypeValid(server.ServerType) {
		return nil, fmt.Errorf("invalid server type for server: %v", server)
	}
	serverToRemove.ServerType = server.ServerType

	return serverToRemove, nil
}

func transSDKBackendServerToReplaceServersInServerGroupRemovedServer(server albsdk.BackendServer) (*albsdk.ReplaceServersInServerGroupRemovedServers, error) {
	serverToRemove := new(albsdk.ReplaceServersInServerGroupRemovedServers)

	serverToRemove.ServerIp = server.ServerIp
	if len(server.ServerId) == 0 {
		return nil, fmt.Errorf("invalid server id for server: %v", server)
	}
	serverToRemove.ServerId = server.ServerId

	if !isServerPortValid(server.Port) {
		return nil, fmt.Errorf("invalid server port for server: %v", server)
	}
	serverToRemove.Port = strconv.Itoa(server.Port)

	if !isServerTypeValid(server.ServerType) {
		return nil, fmt.Errorf("invalid server type for server: %v", server)
	}
	serverToRemove.ServerType = server.ServerType

	return serverToRemove, nil
}

func transModelBackendToSDKAddServersToServerGroupServer(server alb.BackendItem) (*albsdk.AddServersToServerGroupServers, error) {
	serverToAdd := new(albsdk.AddServersToServerGroupServers)

	serverToAdd.ServerIp = server.ServerIp

	if len(server.ServerId) == 0 {
		return nil, fmt.Errorf("invalid server id for server: %v", server)
	}
	serverToAdd.ServerId = server.ServerId

	if !isServerPortValid(server.Port) {
		return nil, fmt.Errorf("invalid server port for server: %v", server)
	}
	serverToAdd.Port = strconv.Itoa(server.Port)

	if !isServerTypeValid(server.Type) {
		return nil, fmt.Errorf("invalid server type for server: %v", server)
	}
	serverToAdd.ServerType = server.Type

	if !isServerWeightValid(server.Weight) {
		return nil, fmt.Errorf("invalid server weight for server: %v", server)
	}
	serverToAdd.Weight = strconv.Itoa(server.Weight)

	return serverToAdd, nil
}

func transModelBackendToSDKReplaceServersInServerGroupAddedServer(server alb.BackendItem) (*albsdk.ReplaceServersInServerGroupAddedServers, error) {
	serverToAdd := new(albsdk.ReplaceServersInServerGroupAddedServers)

	serverToAdd.ServerIp = server.ServerIp

	if len(server.ServerId) == 0 {
		return nil, fmt.Errorf("invalid server id for server: %v", server)
	}
	serverToAdd.ServerId = server.ServerId

	if !isServerPortValid(server.Port) {
		return nil, fmt.Errorf("invalid server port for server: %v", server)
	}
	serverToAdd.Port = strconv.Itoa(server.Port)

	if !isServerTypeValid(server.Type) {
		return nil, fmt.Errorf("invalid server type for server: %v", server)
	}
	serverToAdd.ServerType = server.Type

	if !isServerWeightValid(server.Weight) {
		return nil, fmt.Errorf("invalid server weight for server: %v", server)
	}
	serverToAdd.Weight = strconv.Itoa(server.Weight)

	return serverToAdd, nil
}

func transModelBackendsToSDKAddServersToServerGroupServers(servers []alb.BackendItem) ([]albsdk.AddServersToServerGroupServers, error) {
	serversToAdd := make([]albsdk.AddServersToServerGroupServers, 0)
	for _, resServer := range servers {
		serverToAdd, err := transModelBackendToSDKAddServersToServerGroupServer(resServer)
		if err != nil {
			return nil, err
		}
		serversToAdd = append(serversToAdd, *serverToAdd)
	}
	return serversToAdd, nil
}

func transModelBackendsToSDKReplaceServersInServerGroupAddedServers(servers []alb.BackendItem) ([]albsdk.ReplaceServersInServerGroupAddedServers, error) {
	serversToAdd := make([]albsdk.ReplaceServersInServerGroupAddedServers, 0)
	for _, resServer := range servers {
		serverToAdd, err := transModelBackendToSDKReplaceServersInServerGroupAddedServer(resServer)
		if err != nil {
			return nil, err
		}
		serversToAdd = append(serversToAdd, *serverToAdd)
	}
	return serversToAdd, nil
}

func isServerPortValid(port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	return true
}

func isServerTypeValid(serverType string) bool {
	if !strings.EqualFold(serverType, util.ServerTypeEcs) &&
		!strings.EqualFold(serverType, util.ServerTypeEni) &&
		!strings.EqualFold(serverType, util.ServerTypeEci) {
		return false
	}

	return true
}

func isServerWeightValid(weight int) bool {
	if weight < 0 || weight > 100 {
		return false
	}
	return true
}

func isRegisterServersCompleted(ctx context.Context, serverMgr *ALBProvider, sgpID string, servers []albsdk.AddServersToServerGroupServers) (bool, error) {
	sdkServers, err := serverMgr.ListALBServers(ctx, sgpID)
	if err != nil {
		return false, err
	}

	var isCompleted = true
	for _, server := range servers {
		var serverUID string
		if len(server.ServerIp) == 0 {
			serverUID = fmt.Sprintf("%v:%v", server.ServerId, server.Port)
		} else {
			serverUID = fmt.Sprintf("%v:%v:%v", server.ServerId, server.ServerIp, server.Port)
		}

		isExist := false
		var backendServer albsdk.BackendServer
		for _, sdkServer := range sdkServers {
			var sdkServerUID string
			if len(server.ServerIp) == 0 {
				sdkServerUID = fmt.Sprintf("%v:%v", sdkServer.ServerId, sdkServer.Port)
			} else {
				sdkServerUID = fmt.Sprintf("%v:%v:%v", sdkServer.ServerId, sdkServer.ServerIp, sdkServer.Port)
			}
			if strings.EqualFold(serverUID, sdkServerUID) {
				isExist = true
				backendServer = sdkServer
				break
			}
		}

		if isExist && strings.EqualFold(backendServer.Status, util.ServerStatusAvailable) {
			continue
		}

		isCompleted = false
		break
	}

	if isCompleted {
		return true, nil
	}

	return false, nil
}

func isDeregisterServersCompleted(ctx context.Context, serverMgr *ALBProvider, sgpID string, servers []albsdk.RemoveServersFromServerGroupServers) (bool, error) {
	sdkServers, err := serverMgr.ListALBServers(ctx, sgpID)
	if err != nil {
		return false, err
	}

	var isCompleted = true
	for _, server := range servers {
		var serverUID string
		if len(server.ServerIp) == 0 {
			serverUID = fmt.Sprintf("%v:%v", server.ServerId, server.Port)
		} else {
			serverUID = fmt.Sprintf("%v:%v:%v", server.ServerId, server.ServerIp, server.Port)
		}

		isExist := false
		for _, sdkServer := range sdkServers {
			var sdkServerUID string
			if len(server.ServerIp) == 0 {
				sdkServerUID = fmt.Sprintf("%v:%v", sdkServer.ServerId, sdkServer.Port)
			} else {
				sdkServerUID = fmt.Sprintf("%v:%v:%v", sdkServer.ServerId, sdkServer.ServerIp, sdkServer.Port)
			}
			if strings.EqualFold(serverUID, sdkServerUID) {
				isExist = true
				break
			}
		}

		if isExist {
			isCompleted = false
			break
		}
	}

	if isCompleted {
		return true, nil
	}

	return false, nil
}

func isRegisterServersForReplaceCompleted(sdkServers []albsdk.BackendServer, servers []albsdk.ReplaceServersInServerGroupAddedServers) (bool, error) {
	var isCompleted = true
	for _, server := range servers {
		var serverUID string
		if len(server.ServerIp) == 0 {
			serverUID = fmt.Sprintf("%v:%v", server.ServerId, server.Port)
		} else {
			serverUID = fmt.Sprintf("%v:%v:%v", server.ServerId, server.ServerIp, server.Port)
		}

		isExist := false
		var backendServer albsdk.BackendServer
		for _, sdkServer := range sdkServers {
			var sdkServerUID string
			if len(server.ServerIp) == 0 {
				sdkServerUID = fmt.Sprintf("%v:%v", sdkServer.ServerId, sdkServer.Port)
			} else {
				sdkServerUID = fmt.Sprintf("%v:%v:%v", sdkServer.ServerId, sdkServer.ServerIp, sdkServer.Port)
			}
			if strings.EqualFold(serverUID, sdkServerUID) {
				isExist = true
				backendServer = sdkServer
				break
			}
		}

		if isExist && strings.EqualFold(backendServer.Status, util.ServerStatusAvailable) {
			continue
		}

		isCompleted = false
		break
	}

	if isCompleted {
		return true, nil
	}

	return false, nil
}

func isDeregisterServersForReplaceCompleted(sdkServers []albsdk.BackendServer, servers []albsdk.ReplaceServersInServerGroupRemovedServers) (bool, error) {
	var isCompleted = true
	for _, server := range servers {
		var serverUID string
		if len(server.ServerIp) == 0 {
			serverUID = fmt.Sprintf("%v:%v", server.ServerId, server.Port)
		} else {
			serverUID = fmt.Sprintf("%v:%v:%v", server.ServerId, server.ServerIp, server.Port)
		}

		isExist := false
		for _, sdkServer := range sdkServers {
			var sdkServerUID string
			if len(server.ServerIp) == 0 {
				sdkServerUID = fmt.Sprintf("%v:%v", sdkServer.ServerId, sdkServer.Port)
			} else {
				sdkServerUID = fmt.Sprintf("%v:%v:%v", sdkServer.ServerId, sdkServer.ServerIp, sdkServer.Port)
			}
			if strings.EqualFold(serverUID, sdkServerUID) {
				isExist = true
				break
			}
		}

		if isExist {
			isCompleted = false
			break
		}
	}

	if isCompleted {
		return true, nil
	}

	return false, nil
}

func isReplaceServersCompleted(ctx context.Context, m *ALBProvider, serverGroupID string, registerServers []albsdk.ReplaceServersInServerGroupAddedServers, deregisterServers []albsdk.ReplaceServersInServerGroupRemovedServers) (bool, error) {
	sdkServers, err := m.ListALBServers(ctx, serverGroupID)
	if err != nil {
		return false, err
	}

	isRegisterComplete, err := isRegisterServersForReplaceCompleted(sdkServers, registerServers)
	if err != nil {
		return false, err
	}
	isDeregisterComplete, err := isDeregisterServersForReplaceCompleted(sdkServers, deregisterServers)
	if err != nil {
		return false, err
	}

	if isRegisterComplete && isDeregisterComplete {
		return true, nil
	}

	return false, nil

}
