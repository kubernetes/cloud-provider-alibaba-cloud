package dryrun

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"

	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
)

func NewDryRunALB(
	auth *base.ClientMgr, alb *alb.ALBProvider,
) *DryRunALB {
	return &DryRunALB{auth: auth, alb: alb}
}

var _ prvd.IALB = &DryRunALB{}

type DryRunALB struct {
	auth *base.ClientMgr
	alb  *alb.ALBProvider
}

func (p DryRunALB) DoAction(request requests.AcsRequest, response responses.AcsResponse) (err error) {
	return p.auth.ALB.Client.DoAction(request, response)
}

func (p DryRunALB) UnTagALBResources(request *albsdk.UnTagResourcesRequest) (response *albsdk.UnTagResourcesResponse, err error) {
	return p.auth.ALB.UnTagResources(request)
}

func (p DryRunALB) TagALBResources(request *albsdk.TagResourcesRequest) (response *albsdk.TagResourcesResponse, err error) {
	return p.auth.ALB.TagResources(request)
}

func (p DryRunALB) DescribeALBZones(request *albsdk.DescribeZonesRequest) (response *albsdk.DescribeZonesResponse, err error) {
	return nil, nil
}
func (p DryRunALB) CreateALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, trackingProvider tracking.TrackingProvider) (albmodel.LoadBalancerStatus, error) {
	return albmodel.LoadBalancerStatus{}, nil
}
func (p DryRunALB) ReuseALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, lbID string, trackingProvider tracking.TrackingProvider) (albmodel.LoadBalancerStatus, error) {
	return albmodel.LoadBalancerStatus{}, nil
}
func (p DryRunALB) UnReuseALB(ctx context.Context, lbID string, trackingProvider tracking.TrackingProvider) error {
	return nil
}
func (p DryRunALB) UpdateALB(ctx context.Context, resLB *albmodel.AlbLoadBalancer, sdkLB albsdk.LoadBalancer, trackingProvider tracking.TrackingProvider) (albmodel.LoadBalancerStatus, error) {
	return albmodel.LoadBalancerStatus{}, nil
}
func (p DryRunALB) DeleteALB(ctx context.Context, lbID string) error {
	return nil
}

// ALB Listener
func (p DryRunALB) CreateALBListener(ctx context.Context, resLS *albmodel.Listener) (albmodel.ListenerStatus, error) {
	return albmodel.ListenerStatus{}, nil
}
func (p DryRunALB) UpdateALBListener(ctx context.Context, resLS *albmodel.Listener, sdkLB *albsdk.Listener) (albmodel.ListenerStatus, error) {
	return albmodel.ListenerStatus{}, nil
}
func (p DryRunALB) DeleteALBListener(ctx context.Context, lsID string) error {
	return nil
}
func (p DryRunALB) ListALBListeners(ctx context.Context, lbID string) ([]albsdk.Listener, error) {
	return nil, nil
}

// ALB Listener Rule
func (p DryRunALB) CreateALBListenerRule(ctx context.Context, resLR *albmodel.ListenerRule) (albmodel.ListenerRuleStatus, error) {
	return albmodel.ListenerRuleStatus{}, nil
}
func (p DryRunALB) CreateALBListenerRules(ctx context.Context, resLR []*albmodel.ListenerRule) (map[int]albmodel.ListenerRuleStatus, error) {
	return nil, nil
}
func (p DryRunALB) UpdateALBListenerRule(ctx context.Context, resLR *albmodel.ListenerRule, sdkLR *albsdk.Rule) (albmodel.ListenerRuleStatus, error) {
	return albmodel.ListenerRuleStatus{}, nil
}
func (p DryRunALB) UpdateALBListenerRules(ctx context.Context, matches []albmodel.ResAndSDKListenerRulePair) error {
	return nil
}
func (p DryRunALB) DeleteALBListenerRule(ctx context.Context, sdkLRId string) error {
	return nil
}
func (p DryRunALB) DeleteALBListenerRules(ctx context.Context, sdkLRIds []string) error {
	return nil
}
func (p DryRunALB) ListALBListenerRules(ctx context.Context, lsID string) ([]albsdk.Rule, error) {
	return nil, nil
}
func (p DryRunALB) GetALBListenerAttribute(ctx context.Context, lsID string) (*albsdk.GetListenerAttributeResponse, error) {
	return nil, nil
}

// ALB Server
func (p DryRunALB) RegisterALBServers(ctx context.Context, serverGroupID string, resServers []albmodel.BackendItem) error {
	return nil
}
func (p DryRunALB) DeregisterALBServers(ctx context.Context, serverGroupID string, sdkServers []albsdk.BackendServer) error {
	return nil
}
func (p DryRunALB) ReplaceALBServers(ctx context.Context, serverGroupID string, resServers []albmodel.BackendItem, sdkServers []albsdk.BackendServer) error {
	return nil
}
func (p DryRunALB) ListALBServers(ctx context.Context, serverGroupID string) ([]albsdk.BackendServer, error) {
	return nil, nil
}

// ALB ServerGroup
func (p DryRunALB) CreateALBServerGroup(ctx context.Context, resSGP *albmodel.ServerGroup, trackingProvider tracking.TrackingProvider) (albmodel.ServerGroupStatus, error) {
	return albmodel.ServerGroupStatus{}, nil
}
func (p DryRunALB) UpdateALBServerGroup(ctx context.Context, resSGP *albmodel.ServerGroup, sdkSGP albmodel.ServerGroupWithTags) (albmodel.ServerGroupStatus, error) {
	return albmodel.ServerGroupStatus{}, nil
}
func (p DryRunALB) DeleteALBServerGroup(ctx context.Context, serverGroupID string) error {
	return nil
}

// ALB Tags
func (p DryRunALB) ListALBServerGroupsWithTags(ctx context.Context, tagFilters map[string]string) ([]albmodel.ServerGroupWithTags, error) {
	return nil, nil
}
func (p DryRunALB) ListALBsWithTags(ctx context.Context, tagFilters map[string]string) ([]albmodel.AlbLoadBalancerWithTags, error) {
	return nil, nil
}

func (p DryRunALB) CreateAcl(ctx context.Context, resAcl *albmodel.Acl) (albmodel.AclStatus, error) {
	return albmodel.AclStatus{}, nil
}
func (p DryRunALB) UpdateAcl(ctx context.Context, listenerID string, resAndSDKAclPair albmodel.ResAndSDKAclPair) (albmodel.AclStatus, error) {
	return albmodel.AclStatus{}, nil
}
func (p DryRunALB) DeleteAcl(ctx context.Context, listenerID, sdkAclID string) error {
	return nil
}
func (p DryRunALB) ListAcl(ctx context.Context, listener *albmodel.Listener, aclId string) ([]albsdk.Acl, error) {
	return nil, nil
}

func (p DryRunALB) ListAclEntriesByID(traceID interface{}, sdkAclID string) ([]albsdk.AclEntry, error) {
	return nil, nil
}
