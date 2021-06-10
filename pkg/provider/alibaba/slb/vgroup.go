package slb

import (
	"context"
	"encoding/json"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

func (p SLBProvider) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupsRequest()
	req.LoadBalancerId = lbId
	resp, err := p.auth.SLB.DescribeVServerGroups(req)
	if err != nil {
		return nil, util.FormatErrorMessage(err)
	}
	var vgs []model.VServerGroup
	for _, v := range resp.VServerGroups.VServerGroup {
		vg, err := p.DescribeVServerGroupAttribute(ctx, v.VServerGroupId)
		if err != nil {
			return vgs, util.FormatErrorMessage(err)
		}

		namedKey, err := model.LoadVGroupNamedKey(vg.VGroupName)
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
	backendJson, err := json.Marshal(vg.Backends)
	if err != nil {
		return err
	}
	req.BackendServers = string(backendJson)
	resp, err := p.auth.SLB.CreateVServerGroup(req)
	if err != nil {
		return util.FormatErrorMessage(err)
	}
	vg.VGroupId = resp.VServerGroupId
	return nil
}

func (p SLBProvider) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	resp, err := p.auth.SLB.DescribeVServerGroupAttribute(req)
	if err != nil {
		return model.VServerGroup{}, util.FormatErrorMessage(err)
	}
	vg := setVServerGroupFromResponse(resp)
	return vg, nil

}

func (p SLBProvider) DeleteVServerGroup(ctx context.Context, vGroupId string) error {
	req := slb.CreateDeleteVServerGroupRequest()
	req.VServerGroupId = vGroupId
	_, err := p.auth.SLB.DeleteVServerGroup(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) AddVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateAddVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.AddVServerGroupBackendServers(req)
	return util.FormatErrorMessage(err)

}

func (p SLBProvider) RemoveVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateRemoveVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.RemoveVServerGroupBackendServers(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) SetVServerGroupAttribute(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateSetVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.SetVServerGroupAttribute(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) ModifyVServerGroupBackendServers(ctx context.Context, vGroupId string, old string, new string) error {
	req := slb.CreateModifyVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.OldBackendServers = old
	req.NewBackendServers = new
	_, err := p.auth.SLB.ModifyVServerGroupBackendServers(req)
	return util.FormatErrorMessage(err)
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
