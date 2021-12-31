package vmock

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockALB(
	auth *base.ClientMgr,
) *MockALB {
	return &MockALB{auth: auth}
}

type MockALB struct {
	auth *base.ClientMgr
}

func (p MockALB) TagALBResources(request *albsdk.TagResourcesRequest) (response *albsdk.TagResourcesResponse, err error) {
	return nil, nil
}

func (p MockALB) DescribeALBZones(request *albsdk.DescribeZonesRequest) (response *albsdk.DescribeZonesResponse, err error) {
	return nil, nil
}
func (p MockALB) CreateALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, trackingProvider tracking.TrackingProvider) (albmodel.LoadBalancerStatus, error) {
	return albmodel.LoadBalancerStatus{}, nil
}
func (p MockALB) ReuseALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, lbID string, trackingProvider tracking.TrackingProvider) (albmodel.LoadBalancerStatus, error) {
	return albmodel.LoadBalancerStatus{}, nil
}

func (p MockALB) UpdateALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, sdkLB albsdk.LoadBalancer) (albmodel.LoadBalancerStatus, error) {
	return albmodel.LoadBalancerStatus{}, nil
}
func (p MockALB) DeleteALB(ctx context.Context, lbID string) error {
	return nil
}

// ALB Listener
func (p MockALB) CreateALBListener(ctx context.Context, resLS *albmodel.Listener) (albmodel.ListenerStatus, error) {
	return albmodel.ListenerStatus{}, nil
}
func (p MockALB) UpdateALBListener(ctx context.Context, resLS *albmodel.Listener, sdkLB *albsdk.Listener) (albmodel.ListenerStatus, error) {
	return albmodel.ListenerStatus{}, nil
}
func (p MockALB) DeleteALBListener(ctx context.Context, lsID string) error {
	return nil
}
func (p MockALB) ListALBListeners(ctx context.Context, lbID string) ([]albsdk.Listener, error) {
	return nil, nil
}

// ALB Listener Rule
func (p MockALB) CreateALBListenerRule(ctx context.Context, resLR *albmodel.ListenerRule) (albmodel.ListenerRuleStatus, error) {
	return albmodel.ListenerRuleStatus{}, nil
}
func (p MockALB) CreateALBListenerRules(ctx context.Context, resLR []*albmodel.ListenerRule) (map[int]albmodel.ListenerRuleStatus, error) {
	return nil, nil
}
func (p MockALB) UpdateALBListenerRule(ctx context.Context, resLR *albmodel.ListenerRule, sdkLR *albsdk.Rule) (albmodel.ListenerRuleStatus, error) {
	return albmodel.ListenerRuleStatus{}, nil
}
func (p MockALB) UpdateALBListenerRules(ctx context.Context, matches []albmodel.ResAndSDKListenerRulePair) error {
	return nil
}
func (p MockALB) DeleteALBListenerRule(ctx context.Context, sdkLRId string) error {
	return nil
}
func (p MockALB) DeleteALBListenerRules(ctx context.Context, sdkLRIds []string) error {
	return nil
}
func (p MockALB) ListALBListenerRules(ctx context.Context, lsID string) ([]albsdk.Rule, error) {
	return nil, nil
}

// ALB Server
func (p MockALB) RegisterALBServers(ctx context.Context, serverGroupID string, resServers []albmodel.BackendItem) error {
	return nil
}
func (p MockALB) DeregisterALBServers(ctx context.Context, serverGroupID string, sdkServers []albsdk.BackendServer) error {
	return nil
}
func (p MockALB) ReplaceALBServers(ctx context.Context, serverGroupID string, resServers []albmodel.BackendItem, sdkServers []albsdk.BackendServer) error {
	return nil
}
func (p MockALB) ListALBServers(ctx context.Context, serverGroupID string) ([]albsdk.BackendServer, error) {
	return nil, nil
}

// ALB ServerGroup
func (p MockALB) CreateALBServerGroup(ctx context.Context, resSGP *albmodel.ServerGroup, trackingProvider tracking.TrackingProvider) (albmodel.ServerGroupStatus, error) {
	return albmodel.ServerGroupStatus{}, nil
}
func (p MockALB) UpdateALBServerGroup(ctx context.Context, resSGP *albmodel.ServerGroup, sdkSGP albmodel.ServerGroupWithTags) (albmodel.ServerGroupStatus, error) {
	return albmodel.ServerGroupStatus{}, nil
}
func (p MockALB) DeleteALBServerGroup(ctx context.Context, serverGroupID string) error {
	return nil
}

// ALB Tags
func (p MockALB) ListALBServerGroupsWithTags(ctx context.Context, tagFilters map[string]string) ([]albmodel.ServerGroupWithTags, error) {
	return nil, nil
}
func (p MockALB) ListALBsWithTags(ctx context.Context, tagFilters map[string]string) ([]albmodel.AlbLoadBalancerWithTags, error) {
	return nil, nil
}
