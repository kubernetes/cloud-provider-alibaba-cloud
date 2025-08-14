package nlb

import (
	"context"
	"fmt"
	nlb "github.com/alibabacloud-go/nlb-20220430/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/parallel"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"time"
)

func (p *NLBProvider) toLocalServerGroup(ctx context.Context, remote *nlb.ListServerGroupsResponseBodyServerGroups, servers []nlbmodel.ServerGroupServer) (*nlbmodel.ServerGroup, error) {
	var err error
	sg := &nlbmodel.ServerGroup{
		ServerGroupId:           tea.StringValue(remote.ServerGroupId),
		ServerGroupType:         nlbmodel.ServerGroupType(tea.StringValue(remote.ServerGroupType)),
		ServerGroupName:         tea.StringValue(remote.ServerGroupName),
		AddressIPVersion:        tea.StringValue(remote.AddressIPVersion),
		Scheduler:               tea.StringValue(remote.Scheduler),
		Protocol:                tea.StringValue(remote.Protocol),
		ConnectionDrainEnabled:  remote.ConnectionDrainEnabled,
		ConnectionDrainTimeout:  tea.Int32Value(remote.ConnectionDrainTimeout),
		ResourceGroupId:         tea.StringValue(remote.ResourceGroupId),
		PreserveClientIpEnabled: remote.PreserveClientIpEnabled,
		AnyPortEnabled:          tea.BoolValue(remote.AnyPortEnabled),
	}
	if remote.HealthCheck != nil {
		sg.HealthCheckConfig = &nlbmodel.HealthCheckConfig{
			HealthCheckEnabled:        remote.HealthCheck.HealthCheckEnabled,
			HealthCheckType:           tea.StringValue(remote.HealthCheck.HealthCheckType),
			HealthCheckConnectPort:    tea.Int32Value(remote.HealthCheck.HealthCheckConnectPort),
			HealthyThreshold:          tea.Int32Value(remote.HealthCheck.HealthyThreshold),
			UnhealthyThreshold:        tea.Int32Value(remote.HealthCheck.UnhealthyThreshold),
			HealthCheckConnectTimeout: tea.Int32Value(remote.HealthCheck.HealthCheckConnectTimeout),
			HealthCheckInterval:       tea.Int32Value(remote.HealthCheck.HealthCheckInterval),
			HealthCheckDomain:         tea.StringValue(remote.HealthCheck.HealthCheckDomain),
			HealthCheckUrl:            tea.StringValue(remote.HealthCheck.HealthCheckUrl),
			HttpCheckMethod:           tea.StringValue(remote.HealthCheck.HttpCheckMethod),
		}
		if len(remote.HealthCheck.HealthCheckHttpCode) != 0 {
			for _, code := range remote.HealthCheck.HealthCheckHttpCode {
				sg.HealthCheckConfig.HealthCheckHttpCode = append(sg.HealthCheckConfig.HealthCheckHttpCode,
					tea.StringValue(code))
			}
		}
	}
	sg.NamedKey, err = nlbmodel.LoadNLBSGNamedKey(sg.ServerGroupName)
	if err != nil {
		sg.IsUserManaged = true
	}

	sg.Servers = servers
	for _, t := range remote.Tags {
		tag := tag.Tag{
			Key:   tea.StringValue(t.Key),
			Value: tea.StringValue(t.Value),
		}
		sg.Tags = append(sg.Tags, tag)
	}
	return sg, nil
}

// ServerGroup
func (p *NLBProvider) ListNLBServerGroups(ctx context.Context, tags []tag.Tag) ([]*nlbmodel.ServerGroup, error) {
	var remoteServerGroups []*nlb.ListServerGroupsResponseBodyServerGroups
	var nextToken = ""
	for {
		req := &nlb.ListServerGroupsRequest{}
		req.MaxResults = tea.Int32(100)
		req.NextToken = tea.String(nextToken)
		for _, t := range tags {
			req.Tag = append(req.Tag, &nlb.ListServerGroupsRequestTag{
				Key:   tea.String(t.Key),
				Value: tea.String(t.Value),
			})
		}
		var resp *nlb.ListServerGroupsResponse
		err := retryOnThrottling("ListServerGroups", func() error {
			var err error
			resp, err = p.auth.NLB.ListServerGroups(req)
			return err
		})
		if err != nil {
			return nil, util.SDKError("ListServerGroups", err)
		}
		if resp == nil || resp.Body == nil {
			return nil, fmt.Errorf("OpenAPI ListServerGroups resp is nil")
		}
		klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "ListServerGroups")

		remoteServerGroups = append(remoteServerGroups, resp.Body.ServerGroups...)

		nextToken = tea.StringValue(resp.Body.NextToken)
		if nextToken == "" {
			break
		}
	}

	var sgIds []string
	for _, r := range remoteServerGroups {
		if tea.Int32Value(r.ServerCount) != 0 {
			sgIds = append(sgIds, tea.StringValue(r.ServerGroupId))
		}
	}
	serversMap, err := p.parallelListNLBServers(ctx, sgIds)
	if err != nil {
		return nil, err
	}

	var (
		sgs []*nlbmodel.ServerGroup
	)
	for _, ret := range remoteServerGroups {
		sg, err := p.toLocalServerGroup(ctx, ret, serversMap[tea.StringValue(ret.ServerGroupId)])
		if err != nil {
			return nil, err
		}

		sgs = append(sgs, sg)
	}

	return sgs, nil
}

func (p *NLBProvider) parallelListNLBServers(ctx context.Context, sgId []string) (map[string][]nlbmodel.ServerGroupServer, error) {
	servers := make([][]nlbmodel.ServerGroupServer, len(sgId))
	errs := make([]error, len(sgId))
	parallel.Parallelize(ctx, ctrlCfg.ControllerCFG.MaxConcurrentActions, len(sgId), func(i int) {
		var s []nlbmodel.ServerGroupServer
		var err error
		s, err = p.ListNLBServers(ctx, sgId[i])
		if err != nil {
			errs[i] = err
		}
		servers[i] = s
	})
	ret := map[string][]nlbmodel.ServerGroupServer{}
	for i, id := range sgId {
		ret[id] = servers[i]
	}
	return ret, utilerrors.NewAggregate(errs)
}

func (p *NLBProvider) GetNLBServerGroup(ctx context.Context, sgId string) (*nlbmodel.ServerGroup, error) {
	req := &nlb.ListServerGroupsRequest{}
	req.ServerGroupIds = []*string{tea.String(sgId)}
	var resp *nlb.ListServerGroupsResponse
	err := retryOnThrottling("ListServerGroups", func() error {
		var err error
		resp, err = p.auth.NLB.ListServerGroups(req)
		return err
	})
	if err != nil {
		return nil, util.SDKError("ListServerGroups", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("OpenAPI ListServerGroups resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s, ServerGroupId: %s", tea.StringValue(resp.Body.RequestId), "ListServerGroups", sgId)
	remoteServerGroups := resp.Body.ServerGroups
	if len(remoteServerGroups) == 0 {
		return nil, nil
	}
	servers, err := p.ListNLBServers(ctx, sgId)
	if err != nil {
		return nil, err
	}
	return p.toLocalServerGroup(ctx, remoteServerGroups[0], servers)
}

func (p *NLBProvider) CreateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	var jobId string
	jobId, err := p.CreateNLBServerGroupAsync(ctx, sg)
	if err != nil {
		return err
	}
	return p.waitJobFinish("CreateServerGroup", jobId, 1200*time.Millisecond, DefaultRetryTimeout)
}

func (p *NLBProvider) DeleteNLBServerGroup(ctx context.Context, sgId string) error {
	_, err := p.DeleteNLBServerGroupAsync(ctx, sgId)
	return err

}

func (p *NLBProvider) UpdateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	_, err := p.UpdateNLBServerGroupAsync(ctx, sg)
	return err
}

// ServerGroupServer
func (p *NLBProvider) AddNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer,
) error {
	jobId, err := p.AddNLBServersAsync(ctx, sgId, backends)
	if err != nil {
		return err
	}
	return p.waitJobFinish("AddServersToServerGroup", jobId, 1200*time.Millisecond, DefaultRetryTimeout)
}

func (p *NLBProvider) RemoveNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer,
) error {
	jobId, err := p.RemoveNLBServersAsync(ctx, sgId, backends)
	if err != nil {
		return err
	}
	return p.waitJobFinish("RemoveServersFromServerGroup", jobId, 1200*time.Millisecond, DefaultRetryTimeout)
}

func (p *NLBProvider) UpdateNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer,
) error {
	jobId, err := p.UpdateNLBServersAsync(ctx, sgId, backends)
	if err != nil {
		return err
	}
	return p.waitJobFinish("UpdateServerGroupServersAttribute", jobId, 1200*time.Millisecond, DefaultRetryTimeout)
}

func (p *NLBProvider) ListNLBServers(ctx context.Context, sgId string) ([]nlbmodel.ServerGroupServer, error) {
	var ret []nlbmodel.ServerGroupServer
	var nextToken = ""
	for {
		req := &nlb.ListServerGroupServersRequest{}
		req.ServerGroupId = tea.String(sgId)
		req.MaxResults = tea.Int32(100)
		req.NextToken = tea.String(nextToken)

		var resp *nlb.ListServerGroupServersResponse
		err := retryOnThrottling("ListServerGroupServers", func() error {
			var err error
			resp, err = p.auth.NLB.ListServerGroupServers(req)
			return err
		})
		if err != nil {
			return nil, util.SDKError("ListServerGroupServers", err)
		}
		if resp == nil || resp.Body == nil {
			return nil, fmt.Errorf("OpenAPI ListServerGroupServers resp is nil")
		}
		klog.V(5).Infof("RequestId: %s, API: %s, ServerGroupId: %s",
			tea.StringValue(resp.Body.RequestId), "ListServerGroupServers", tea.StringValue(req.ServerGroupId))

		for _, s := range resp.Body.Servers {
			ret = append(ret, nlbmodel.ServerGroupServer{
				ServerGroupId: tea.StringValue(s.ServerGroupId),
				Description:   tea.StringValue(s.Description),
				ServerId:      tea.StringValue(s.ServerId),
				ServerIp:      tea.StringValue(s.ServerIp),
				ServerType:    nlbmodel.ServerType(tea.StringValue(s.ServerType)),
				Port:          tea.Int32Value(s.Port),
				Weight:        tea.Int32Value(s.Weight),
			})
		}

		nextToken = tea.StringValue(resp.Body.NextToken)
		if nextToken == "" {
			return ret, nil
		}
	}
}

func (p *NLBProvider) CreateNLBServerGroupAsync(ctx context.Context, sg *nlbmodel.ServerGroup) (string, error) {
	req := &nlb.CreateServerGroupRequest{}
	req.ServerGroupName = tea.String(sg.ServerGroupName)
	req.VpcId = tea.String(sg.VPCId)
	req.Protocol = tea.String(sg.Protocol)
	req.AnyPortEnabled = tea.Bool(sg.AnyPortEnabled)
	for _, t := range sg.Tags {
		req.Tag = append(req.Tag, &nlb.CreateServerGroupRequestTag{
			Key:   tea.String(t.Key),
			Value: tea.String(t.Value),
		})
	}

	if sg.ServerGroupType != "" {
		req.ServerGroupType = tea.String(string(sg.ServerGroupType))
	}
	if sg.AddressIPVersion != "" {
		req.AddressIPVersion = tea.String(sg.AddressIPVersion)
	}
	if sg.ConnectionDrainEnabled != nil {
		req.ConnectionDrainEnabled = sg.ConnectionDrainEnabled
		req.ConnectionDrainTimeout = tea.Int32(sg.ConnectionDrainTimeout)
	}
	if sg.Scheduler != "" {
		req.Scheduler = tea.String(sg.Scheduler)
	}
	if sg.PreserveClientIpEnabled != nil {
		req.PreserveClientIpEnabled = sg.PreserveClientIpEnabled
	}
	if sg.ServerGroupId != "" {
		req.ResourceGroupId = tea.String(sg.ResourceGroupId)
	}
	if sg.ResourceGroupId != "" {
		req.ResourceGroupId = tea.String(sg.ResourceGroupId)
	}
	// health check
	if sg.HealthCheckConfig != nil {
		req.HealthCheckConfig = &nlb.CreateServerGroupRequestHealthCheckConfig{
			HealthCheckEnabled: sg.HealthCheckConfig.HealthCheckEnabled,
		}
		if sg.HealthCheckConfig.HealthCheckType != "" {
			req.HealthCheckConfig.HealthCheckType = tea.String(sg.HealthCheckConfig.HealthCheckType)
		}
		if sg.HealthCheckConfig.HealthCheckConnectPort != 0 {
			req.HealthCheckConfig.HealthCheckConnectPort = tea.Int32(sg.HealthCheckConfig.HealthCheckConnectPort)
		}
		if sg.HealthCheckConfig.HealthyThreshold != 0 {
			req.HealthCheckConfig.HealthyThreshold = tea.Int32(sg.HealthCheckConfig.HealthyThreshold)
		}
		if sg.HealthCheckConfig.UnhealthyThreshold != 0 {
			req.HealthCheckConfig.UnhealthyThreshold = tea.Int32(sg.HealthCheckConfig.UnhealthyThreshold)
		}
		if sg.HealthCheckConfig.HealthCheckConnectTimeout != 0 {
			req.HealthCheckConfig.HealthCheckConnectTimeout = tea.Int32(sg.HealthCheckConfig.HealthCheckConnectTimeout)
		}
		if sg.HealthCheckConfig.HealthCheckInterval != 0 {
			req.HealthCheckConfig.HealthCheckInterval = tea.Int32(sg.HealthCheckConfig.HealthCheckInterval)
		}
		if sg.HealthCheckConfig.HealthCheckDomain != "" {
			req.HealthCheckConfig.HealthCheckDomain = tea.String(sg.HealthCheckConfig.HealthCheckDomain)
		}
		if sg.HealthCheckConfig.HealthCheckUrl != "" {
			req.HealthCheckConfig.HealthCheckUrl = tea.String(sg.HealthCheckConfig.HealthCheckUrl)
		}
		if sg.HealthCheckConfig.HttpCheckMethod != "" {
			req.HealthCheckConfig.HttpCheckMethod = tea.String(sg.HealthCheckConfig.HttpCheckMethod)
		}
		if len(sg.HealthCheckConfig.HealthCheckHttpCode) != 0 {
			for _, code := range sg.HealthCheckConfig.HealthCheckHttpCode {
				req.HealthCheckConfig.HealthCheckHttpCode = append(req.HealthCheckConfig.HealthCheckHttpCode,
					tea.String(code))
			}
		}
	}

	var resp *nlb.CreateServerGroupResponse
	err := retryOnThrottling("CreateServerGroup", func() error {
		var err error
		resp, err = p.auth.NLB.CreateServerGroup(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("CreateServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI CreateServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "CreateServerGroup")

	sg.ServerGroupId = tea.StringValue(resp.Body.ServerGroupId)
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) DeleteNLBServerGroupAsync(ctx context.Context, sgId string) (string, error) {
	req := &nlb.DeleteServerGroupRequest{}
	req.ServerGroupId = tea.String(sgId)
	var resp *nlb.DeleteServerGroupResponse
	err := retryOnThrottling("DeleteServerGroup", func() error {
		var err error
		resp, err = p.auth.NLB.DeleteServerGroup(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("DeleteServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI DeleteServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DeleteServerGroup")
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) UpdateNLBServerGroupAsync(ctx context.Context, sg *nlbmodel.ServerGroup) (string, error) {
	req := &nlb.UpdateServerGroupAttributeRequest{}
	// required
	req.ServerGroupId = tea.String(sg.ServerGroupId)
	// options
	if sg.ServerGroupName != "" {
		req.ServerGroupName = tea.String(sg.ServerGroupName)
	}
	if sg.ConnectionDrainEnabled != nil {
		req.ConnectionDrainEnabled = sg.ConnectionDrainEnabled
		req.ConnectionDrainTimeout = tea.Int32(sg.ConnectionDrainTimeout)
	}
	if sg.Scheduler != "" {
		req.Scheduler = tea.String(sg.Scheduler)
	}
	req.PreserveClientIpEnabled = sg.PreserveClientIpEnabled
	if sg.HealthCheckConfig != nil {
		req.HealthCheckConfig = &nlb.UpdateServerGroupAttributeRequestHealthCheckConfig{
			HealthCheckEnabled: sg.HealthCheckConfig.HealthCheckEnabled,
		}
		if sg.HealthCheckConfig.HealthCheckType != "" {
			req.HealthCheckConfig.HealthCheckType = tea.String(sg.HealthCheckConfig.HealthCheckType)
		}
		if sg.HealthCheckConfig.HealthCheckConnectPort != 0 {
			req.HealthCheckConfig.HealthCheckConnectPort = tea.Int32(sg.HealthCheckConfig.HealthCheckConnectPort)
		}
		if sg.HealthCheckConfig.HealthyThreshold != 0 {
			req.HealthCheckConfig.HealthyThreshold = tea.Int32(sg.HealthCheckConfig.HealthyThreshold)
		}
		if sg.HealthCheckConfig.UnhealthyThreshold != 0 {
			req.HealthCheckConfig.UnhealthyThreshold = tea.Int32(sg.HealthCheckConfig.UnhealthyThreshold)
		}
		if sg.HealthCheckConfig.HealthCheckConnectTimeout != 0 {
			req.HealthCheckConfig.HealthCheckConnectTimeout = tea.Int32(sg.HealthCheckConfig.HealthCheckConnectTimeout)
		}
		if sg.HealthCheckConfig.HealthCheckInterval != 0 {
			req.HealthCheckConfig.HealthCheckInterval = tea.Int32(sg.HealthCheckConfig.HealthCheckInterval)
		}
		if sg.HealthCheckConfig.HealthCheckDomain != "" {
			req.HealthCheckConfig.HealthCheckDomain = tea.String(sg.HealthCheckConfig.HealthCheckDomain)
		}
		if sg.HealthCheckConfig.HealthCheckUrl != "" {
			req.HealthCheckConfig.HealthCheckUrl = tea.String(sg.HealthCheckConfig.HealthCheckUrl)
		}
		if sg.HealthCheckConfig.HttpCheckMethod != "" {
			req.HealthCheckConfig.HttpCheckMethod = tea.String(sg.HealthCheckConfig.HttpCheckMethod)
		}
		if len(sg.HealthCheckConfig.HealthCheckHttpCode) != 0 {
			for _, code := range sg.HealthCheckConfig.HealthCheckHttpCode {
				req.HealthCheckConfig.HealthCheckHttpCode = append(req.HealthCheckConfig.HealthCheckHttpCode,
					tea.String(code))
			}
		}
	}

	var resp *nlb.UpdateServerGroupAttributeResponse
	err := retryOnThrottling("UpdateServerGroupAttribute", func() error {
		var err error
		resp, err = p.auth.NLB.UpdateServerGroupAttribute(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("UpdateServerGroupAttribute", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI UpdateServerGroupAttribute resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateServerGroupAttribute")
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) AddNLBServersAsync(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) (string, error) {
	req := &nlb.AddServersToServerGroupRequest{}
	req.ServerGroupId = tea.String(sgId)
	for _, b := range backends {
		reqServer := &nlb.AddServersToServerGroupRequestServers{}
		reqServer.ServerId = tea.String(b.ServerId)
		reqServer.ServerType = tea.String(string(b.ServerType))
		reqServer.Port = tea.Int32(b.Port)
		reqServer.Description = tea.String(b.Description)
		reqServer.Weight = tea.Int32(b.Weight)
		if b.ServerIp != "" {
			reqServer.ServerIp = tea.String(b.ServerIp)
		}

		req.Servers = append(req.Servers, reqServer)
	}

	var resp *nlb.AddServersToServerGroupResponse
	err := retryOnThrottling("AddServersToServerGroup", func() error {
		var err error
		resp, err = p.auth.NLB.AddServersToServerGroup(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("AddServersToServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI AddServersToServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "AddServersToServerGroup")
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) RemoveNLBServersAsync(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) (string, error) {
	req := &nlb.RemoveServersFromServerGroupRequest{}
	req.ServerGroupId = tea.String(sgId)

	for _, b := range backends {
		reqServer := &nlb.RemoveServersFromServerGroupRequestServers{}
		reqServer.ServerId = tea.String(b.ServerId)
		reqServer.ServerType = tea.String(string(b.ServerType))
		reqServer.Port = tea.Int32(b.Port)
		if b.ServerIp != "" {
			reqServer.ServerIp = tea.String(b.ServerIp)
		}

		req.Servers = append(req.Servers, reqServer)
	}

	var resp *nlb.RemoveServersFromServerGroupResponse
	err := retryOnThrottling("RemoveServersFromServerGroup", func() error {
		var err error
		resp, err = p.auth.NLB.RemoveServersFromServerGroup(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("RemoveServersFromServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI RemoveServersFromServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "RemoveServersFromServerGroup")
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) UpdateNLBServersAsync(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) (string, error) {
	req := &nlb.UpdateServerGroupServersAttributeRequest{}
	req.ServerGroupId = tea.String(sgId)

	for _, b := range backends {
		reqServer := &nlb.UpdateServerGroupServersAttributeRequestServers{}

		reqServer.ServerId = tea.String(b.ServerId)
		reqServer.ServerType = tea.String(string(b.ServerType))
		reqServer.Description = tea.String(b.Description)
		reqServer.Port = tea.Int32(b.Port)
		reqServer.Weight = tea.Int32(b.Weight)
		if b.ServerIp != "" {
			reqServer.ServerIp = tea.String(b.ServerIp)
		}

		req.Servers = append(req.Servers, reqServer)
	}

	var resp *nlb.UpdateServerGroupServersAttributeResponse
	err := retryOnThrottling("UpdateServerGroupServersAttribute", func() error {
		var err error
		resp, err = p.auth.NLB.UpdateServerGroupServersAttribute(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("UpdateServerGroupServersAttribute", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI UpdateServerGroupServersAttribute resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateServerGroupServersAttribute")
	return tea.StringValue(resp.Body.JobId), nil
}
