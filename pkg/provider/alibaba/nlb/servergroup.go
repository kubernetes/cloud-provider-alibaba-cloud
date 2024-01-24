package nlb

import (
	"context"
	"fmt"
	nlb "github.com/alibabacloud-go/nlb-20220430/client"
	"github.com/alibabacloud-go/tea/tea"
	"k8s.io/apimachinery/pkg/util/wait"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"time"
)

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
		resp, err := p.auth.NLB.ListServerGroups(req)
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

	var (
		sgs []*nlbmodel.ServerGroup
		err error
	)
	for _, ret := range remoteServerGroups {
		sg := &nlbmodel.ServerGroup{
			ServerGroupId:           tea.StringValue(ret.ServerGroupId),
			ServerGroupType:         nlbmodel.ServerGroupType(tea.StringValue(ret.ServerGroupType)),
			ServerGroupName:         tea.StringValue(ret.ServerGroupName),
			AddressIPVersion:        tea.StringValue(ret.AddressIPVersion),
			Scheduler:               tea.StringValue(ret.Scheduler),
			Protocol:                tea.StringValue(ret.Protocol),
			ConnectionDrainEnabled:  ret.ConnectionDrainEnabled,
			ConnectionDrainTimeout:  tea.Int32Value(ret.ConnectionDrainTimeout),
			ResourceGroupId:         tea.StringValue(ret.ResourceGroupId),
			PreserveClientIpEnabled: ret.PreserveClientIpEnabled,
		}
		if ret.HealthCheck != nil {
			sg.HealthCheckConfig = &nlbmodel.HealthCheckConfig{
				HealthCheckEnabled:        ret.HealthCheck.HealthCheckEnabled,
				HealthCheckType:           tea.StringValue(ret.HealthCheck.HealthCheckType),
				HealthCheckConnectPort:    tea.Int32Value(ret.HealthCheck.HealthCheckConnectPort),
				HealthyThreshold:          tea.Int32Value(ret.HealthCheck.HealthyThreshold),
				UnhealthyThreshold:        tea.Int32Value(ret.HealthCheck.UnhealthyThreshold),
				HealthCheckConnectTimeout: tea.Int32Value(ret.HealthCheck.HealthCheckConnectTimeout),
				HealthCheckInterval:       tea.Int32Value(ret.HealthCheck.HealthCheckInterval),
				HealthCheckDomain:         tea.StringValue(ret.HealthCheck.HealthCheckDomain),
				HealthCheckUrl:            tea.StringValue(ret.HealthCheck.HealthCheckUrl),
				HttpCheckMethod:           tea.StringValue(ret.HealthCheck.HttpCheckMethod),
			}
			if len(ret.HealthCheck.HealthCheckHttpCode) != 0 {
				for _, code := range ret.HealthCheck.HealthCheckHttpCode {
					sg.HealthCheckConfig.HealthCheckHttpCode = append(sg.HealthCheckConfig.HealthCheckHttpCode,
						tea.StringValue(code))
				}
			}
		}
		sg.NamedKey, err = nlbmodel.LoadNLBSGNamedKey(sg.ServerGroupName)
		if err != nil {
			sg.IsUserManaged = true
		}

		servers, err := p.ListNLBServers(ctx, sg.ServerGroupId)
		if err != nil {
			return nil, fmt.Errorf("[%s] [%s] %s", sg.ServerGroupId, sg.ServerGroupName,
				util.SDKError("ListNLBServers", err).Error())
		}
		sg.Servers = servers

		sgs = append(sgs, sg)
	}

	return sgs, nil
}

func (p *NLBProvider) CreateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	req := &nlb.CreateServerGroupRequest{}
	req.ServerGroupName = tea.String(sg.ServerGroupName)
	req.VpcId = tea.String(sg.VPCId)
	req.Protocol = tea.String(sg.Protocol)
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

	resp, err := p.auth.NLB.CreateServerGroup(req)
	if err != nil {
		return util.SDKError("CreateServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI CreateServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "CreateServerGroup")

	sg.ServerGroupId = tea.StringValue(resp.Body.ServerGroupId)

	var (
		getResp *nlb.ListServerGroupsResponse
		retErr  error
	)

	_ = wait.PollImmediate(3*time.Second, 10*time.Second, func() (bool, error) {
		getReq := &nlb.ListServerGroupsRequest{}
		getReq.ServerGroupIds = []*string{tea.String(sg.ServerGroupId)}

		getResp, retErr = p.auth.NLB.ListServerGroups(getReq)
		if retErr != nil {
			retErr = util.SDKError("ListServerGroups", retErr)
			return false, retErr
		}
		if getResp == nil || getResp.Body == nil {
			retErr = fmt.Errorf("OpenAPI ListServerGroups resp is nil, req: %+v", getReq)
			return false, nil
		}
		if tea.StringValue(getResp.Body.ServerGroups[0].ServerGroupStatus) == "Creating" {
			klog.V(5).Infof("%s is still in creating status, wait next time", sg.ServerGroupId)
			return false, nil
		}
		retErr = nil
		return true, retErr
	})
	return retErr
}

func (p *NLBProvider) DeleteNLBServerGroup(ctx context.Context, sgId string) error {
	req := &nlb.DeleteServerGroupRequest{}
	req.ServerGroupId = tea.String(sgId)
	resp, err := p.auth.NLB.DeleteServerGroup(req)
	if err != nil {
		return util.SDKError("DeleteServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI DeleteServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DeleteServerGroup")
	return nil

}

func (p *NLBProvider) UpdateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
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

	resp, err := p.auth.NLB.UpdateServerGroupAttribute(req)
	if err != nil {
		return util.SDKError("UpdateServerGroupAttribute", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI UpdateServerGroupAttribute resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateServerGroupAttribute")
	return nil
}

// ServerGroupServer
func (p *NLBProvider) AddNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer,
) error {
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

	resp, err := p.auth.NLB.AddServersToServerGroup(req)
	if err != nil {
		return util.SDKError("AddServersToServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI AddServersToServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "AddServersToServerGroup")
	return p.waitJobFinish("AddServersToServerGroup", tea.StringValue(resp.Body.JobId))
}

func (p *NLBProvider) RemoveNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer,
) error {

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

	resp, err := p.auth.NLB.RemoveServersFromServerGroup(req)
	if err != nil {
		return util.SDKError("RemoveServersFromServerGroup", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI RemoveServersFromServerGroup resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "RemoveServersFromServerGroup")
	return p.waitJobFinish("RemoveServersFromServerGroup", tea.StringValue(resp.Body.JobId))
}

func (p *NLBProvider) UpdateNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer,
) error {
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

	resp, err := p.auth.NLB.UpdateServerGroupServersAttribute(req)
	if err != nil {
		return util.SDKError("UpdateServerGroupServersAttribute", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI UpdateServerGroupServersAttribute resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateServerGroupServersAttribute")
	return p.waitJobFinish("UpdateServerGroupServersAttribute", tea.StringValue(resp.Body.JobId))
}

func (p *NLBProvider) ListNLBServers(ctx context.Context, sgId string) ([]nlbmodel.ServerGroupServer, error) {
	var ret []nlbmodel.ServerGroupServer
	var nextToken = ""
	for {
		req := &nlb.ListServerGroupServersRequest{}
		req.ServerGroupId = tea.String(sgId)
		req.MaxResults = tea.Int32(100)
		req.NextToken = tea.String(nextToken)

		resp, err := p.auth.NLB.ListServerGroupServers(req)
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
