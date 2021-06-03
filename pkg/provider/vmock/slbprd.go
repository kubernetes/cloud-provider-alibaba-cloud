package vmock

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
)

func NewMockCLB(
	auth *alibaba.ClientAuth,
) *MockCLB {
	return &MockCLB{auth: auth}
}

type MockCLB struct {
	auth *alibaba.ClientAuth
}

func (m *MockCLB) FindLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return nil
}
func (m *MockCLB) DescribeLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return nil
}
func (m *MockCLB) DeleteLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return nil
}
func (m *MockCLB) ModifyLoadBalancerInstanceSpec(ctx context.Context, lbId string, spec string) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerDeleteProtection(ctx context.Context, lbId string, flag string) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerName(ctx context.Context, lbId string, name string) error {
	return nil
}
func (m *MockCLB) ModifyLoadBalancerInternetSpec(ctx context.Context, lbId string, chargeType string, bandwidth int) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerModificationProtection(ctx context.Context, lbId string, flag string) error {
	return nil
}
func (m *MockCLB) AddTags(ctx context.Context, lbId string, tags string) error {
	return nil
}

func (m *MockCLB) DescribeTags(ctx context.Context, lbId string) ([]model.Tag, error) {
	return nil, nil
}

// Listener
func (m *MockCLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	return nil, nil
}
func (m *MockCLB) StartLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	return nil
}
func (m *MockCLB) StopLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	return nil
}
func (m *MockCLB) DeleteLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerTCPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerTCPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerUDPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerUDPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerHTTPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerHTTPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerHTTPSListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerHTTPSListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}

// VServerGroup
func (m *MockCLB) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	return nil, nil
}
func (m *MockCLB) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
	return nil
}
func (m *MockCLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	return model.VServerGroup{}, nil
}
func (m *MockCLB) DeleteVServerGroup(ctx context.Context, vGroupId string) error {
	return nil
}
func (m *MockCLB) AddVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	return nil
}
func (m *MockCLB) RemoveVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	return nil
}
func (m *MockCLB) SetVServerGroupAttribute(ctx context.Context, vGroupId string, backends string) error {
	return nil
}
func (m *MockCLB) ModifyVServerGroupBackendServers(ctx context.Context, vGroupId string, old string, new string) error {
	return nil
}
