package alibaba

import (
	"context"
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
)

func (p ProviderSLB) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupsRequest()
	req.LoadBalancerId = lbId
	resp, err := p.auth.SLB.DescribeVServerGroups(req)
	if err != nil {
		return nil, err
	}
	var vgs []model.VServerGroup
	for _, v := range resp.VServerGroups.VServerGroup {
		vg, err := p.DescribeVServerGroupAttribute(ctx, v.VServerGroupId)
		if err != nil {
			return vgs, err
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

func (p ProviderSLB) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
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
		return err
	}
	vg.VGroupId = resp.VServerGroupId
	return nil
}

func (p ProviderSLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	resp, err := p.auth.SLB.DescribeVServerGroupAttribute(req)
	if err != nil {
		return model.VServerGroup{}, err
	}
	vg := setVServerGroupFromResponse(resp)
	return vg, nil

}

func (p ProviderSLB) DeleteVServerGroup(ctx context.Context, vGroupId string) error {
	req := slb.CreateDeleteVServerGroupRequest()
	req.VServerGroupId = vGroupId
	_, err := p.auth.SLB.DeleteVServerGroup(req)
	return err
}

func (p ProviderSLB) AddVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateAddVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.AddVServerGroupBackendServers(req)
	return err

}

func (p ProviderSLB) RemoveVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateRemoveVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.RemoveVServerGroupBackendServers(req)
	return err
}

func (p ProviderSLB) SetVServerGroupAttribute(ctx context.Context, vGroupId string, backends string) error {
	req := slb.CreateSetVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	req.BackendServers = backends
	_, err := p.auth.SLB.SetVServerGroupAttribute(req)
	return err
}

func (p ProviderSLB) ModifyVServerGroupBackendServers(ctx context.Context, vGroupId string, old string, new string) error {
	req := slb.CreateModifyVServerGroupBackendServersRequest()
	req.VServerGroupId = vGroupId
	req.OldBackendServers = old
	req.NewBackendServers = new
	_, err := p.auth.SLB.ModifyVServerGroupBackendServers(req)
	return err
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
