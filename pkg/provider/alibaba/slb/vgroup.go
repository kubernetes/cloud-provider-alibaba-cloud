package slb

import (
	"context"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

func (p SLBProvider) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupsRequest()
	req.LoadBalancerId = lbId
	resp, err := p.auth.SLB.DescribeVServerGroups(req)
	if err != nil {
		return nil, util.SDKError("DescribeVServerGroups", err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s, lbId: %s", resp.RequestId, "DescribeVServerGroups", lbId)
	var vgs []model.VServerGroup
	for _, v := range resp.VServerGroups.VServerGroup {
		vg := model.VServerGroup{
			VGroupId:   v.VServerGroupId,
			VGroupName: v.VServerGroupName,
		}
		namedKey, err := model.LoadVGroupNamedKey(v.VServerGroupName)
		if err != nil {
			// add to vgs, for reusing vGroupId
			vg.IsUserManaged = true
		}
		vg.NamedKey = namedKey
		vgs = append(vgs, vg)
	}
	return vgs, nil
}

func (p SLBProvider) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
	req := slb.CreateCreateVServerGroupRequest()
	req.LoadBalancerId = lbId
	req.VServerGroupName = vg.VGroupName
	// create vserver group with empty backends to avoid reach the limit of backends per action
	resp, err := p.auth.SLB.CreateVServerGroup(req)
	if err != nil {
		return util.SDKError("CreateVServerGroup", err)
	}
	vg.VGroupId = resp.VServerGroupId
	return nil
}

func (p SLBProvider) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	resp, err := p.auth.SLB.DescribeVServerGroupAttribute(req)
	if err != nil {
		return model.VServerGroup{}, util.SDKError("DescribeVServerGroupAttribute", err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s, vGroupId: %s", resp.RequestId, "DescribeVServerGroupAttribute", vGroupId)
	vg := setVServerGroupFromResponse(resp)
	return vg, nil

}

func (p SLBProvider) DeleteVServerGroup(ctx context.Context, vGroupId string) error {
	req := slb.CreateDeleteVServerGroupRequest()
	req.VServerGroupId = vGroupId
	_, err := p.auth.SLB.DeleteVServerGroup(req)
	return util.SDKError("DeleteVServerGroup", err)
}

func (p SLBProvider) AddVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateAddVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.AddVServerGroupBackendServers(req)
	return util.SDKError("AddVServerGroupBackendServers", err)

}

func (p SLBProvider) RemoveVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateRemoveVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.RemoveVServerGroupBackendServers(req)
	return util.SDKError("RemoveVServerGroupBackendServers", err)
}

func (p SLBProvider) SetVServerGroupAttribute(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateSetVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.SetVServerGroupAttribute(req)
	return util.SDKError("SetVServerGroupAttribute", err)
}

func (p SLBProvider) ModifyVServerGroupBackendServers(ctx context.Context, vGroupId string, old string, new string) error {
	req := slb.CreateModifyVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.OldBackendServers = old
	req.NewBackendServers = new
	_, err := p.auth.SLB.ModifyVServerGroupBackendServers(req)
	return util.SDKError("ModifyVServerGroupBackendServers", err)
}

func setVServerGroupFromResponse(resp *slb.DescribeVServerGroupAttributeResponse) model.VServerGroup {
	vg := model.VServerGroup{
		VGroupId:   resp.VServerGroupId,
		VGroupName: resp.VServerGroupName,
		Backends:   nil,
	}
	var backends []model.BackendAttribute
	for _, backend := range resp.BackendServers.BackendServer {
		b := model.BackendAttribute{
			Description: backend.Description,
			ServerId:    backend.ServerId,
			ServerIp:    backend.ServerIp,
			Weight:      backend.Weight,
			Port:        backend.Port,
			Type:        backend.Type,
		}
		backends = append(backends, b)
	}
	vg.Backends = backends
	return vg

}
