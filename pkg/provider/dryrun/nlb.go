package dryrun

import (
	"context"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/nlb"
)

func NewDryRunNLB(
	auth *base.ClientMgr, nlb *nlb.NLBProvider,
) *DryRunNLB {
	return &DryRunNLB{auth: auth, nlb: nlb}
}

var _ prvd.INLB = &DryRunNLB{}

type DryRunNLB struct {
	auth *base.ClientMgr
	nlb  *nlb.NLBProvider
}

func (d DryRunNLB) UpdateNLBSecurityGroupIds(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer, added, removed []string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) TagNLBResource(ctx context.Context, resourceId string, resourceType nlbmodel.TagResourceType, tags []tag.Tag) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) ListNLBTagResources(ctx context.Context, lbId string) ([]tag.Tag, error) {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) FindNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) DescribeNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) CreateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) DeleteNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateNLBAddressType(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateNLBZones(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateLoadBalancerProtection(ctx context.Context, lbId string,
	delCfg *nlbmodel.DeletionProtectionConfig, modCfg *nlbmodel.ModificationProtectionConfig) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) AttachCommonBandwidthPackageToLoadBalancer(ctx context.Context, lbId string, bandwidthPackageId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) DetachCommonBandwidthPackageFromLoadBalancer(ctx context.Context, lbId string, bandwidthPackageId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) ListNLBServerGroups(ctx context.Context, tags []tag.Tag) ([]*nlbmodel.ServerGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) CreateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) DeleteNLBServerGroup(ctx context.Context, sgId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) AddNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) RemoveNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) ListNLBListeners(ctx context.Context, lbId string) ([]*nlbmodel.ListenerAttribute, error) {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) CreateNLBListener(ctx context.Context, lbId string, lis *nlbmodel.ListenerAttribute) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) UpdateNLBListener(ctx context.Context, lis *nlbmodel.ListenerAttribute) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) DeleteNLBListener(ctx context.Context, listenerId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) StartNLBListener(ctx context.Context, listenerId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) StopNLBListener(ctx context.Context, listenerId string) error {
	//TODO implement me
	panic("implement me")
}
