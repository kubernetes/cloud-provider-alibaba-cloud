package dryrun

import (
	"context"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/elb"
)

func NewDryRunELB(auth *base.ClientMgr, elb *elb.ELBProvider) *DryRunELB {
	return &DryRunELB{auth: auth, elb: elb}
}

var _ prvd.IELB = &DryRunELB{}

type DryRunELB struct {
	auth *base.ClientMgr
	elb  *elb.ELBProvider
}

func (d DryRunELB) FindEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) CreateEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) DeleteEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) DescribeEdgeLoadBalancerById(ctx context.Context, lbId string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) DescribeEdgeLoadBalancerByName(ctx context.Context, lbName string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) SetEdgeLoadBalancerStatus(ctx context.Context, status string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) CreateEip(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) ReleaseEip(ctx context.Context, eipId string) error {
	return nil
}

func (d DryRunELB) AssociateElbEipAddress(ctx context.Context, eipId, lbId string) error {
	return nil
}

func (d DryRunELB) UnAssociateElbEipAddress(ctx context.Context, eipId string) error {
	return nil
}

func (d DryRunELB) ModifyEipAttribute(ctx context.Context, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) DescribeEnsEipAddressesById(ctx context.Context, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) DescribeEnsEipAddressesByName(ctx context.Context, eipName string, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) FindAssociatedInstance(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}

func (d DryRunELB) FindEdgeLoadBalancerListener(ctx context.Context, lbId string, listeners *elbmodel.EdgeListeners) error {
	return nil
}

func (d DryRunELB) DescribeEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) DescribeEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) DescribeEdgeLoadBalancerHTTPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) DescribeEdgeLoadBalancerHTTPSListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) StartEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	return nil
}

func (d DryRunELB) StopEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	return nil
}

func (d DryRunELB) CreateEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) CreateEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) ModifyEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) ModifyEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	return nil
}

func (d DryRunELB) DeleteEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	return nil
}

func (d DryRunELB) AddBackendToEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (d DryRunELB) UpdateEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (d DryRunELB) RemoveBackendFromEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (d DryRunELB) FindBackendFromLoadBalancer(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	return nil
}

func (d DryRunELB) GetEnsRegionIdByNetwork(ctx context.Context, networkId string) (string, error) {
	return "", nil
}

func (d DryRunELB) FindNetWorkAndVSwitchByLoadBalancerId(ctx context.Context, lbId string) ([]string, error) {
	return []string{}, nil
}

func (d DryRunELB) FindEnsInstancesByNetwork(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) (map[string]string, error) {
	return map[string]string{}, nil
}

func (d DryRunELB) DescribeNetwork(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	return nil
}
