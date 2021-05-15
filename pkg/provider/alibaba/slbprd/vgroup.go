package slbprd

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
		vg := model.VServerGroup{
			VGroupId:   v.VServerGroupId,
			VGroupName: v.VServerGroupName,
		}
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

func (p ProviderSLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (*model.VServerGroup, error) {
	req := slb.CreateDescribeVServerGroupAttributeRequest()
	req.VServerGroupId = vGroupId
	resp, err := p.auth.SLB.DescribeVServerGroupAttribute(req)
	if err != nil {
		return nil, err
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

func setVServerGroupFromResponse(resp *slb.DescribeVServerGroupAttributeResponse) *model.VServerGroup {
	vg := model.VServerGroup{
		VGroupId:   resp.VServerGroupId,
		VGroupName: resp.VServerGroupName,
		Backends:   nil,
	}
	// TODO
	return &vg

}
