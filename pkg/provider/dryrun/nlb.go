package dryrun

import (
	"context"
	"fmt"
	"time"

	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/dryrun"
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

const (
	MTypeDeleteListener                                 = "DeleteListener"
	MTypeUpdateListener                                 = "UpdateListener"
	MTypeCreateListener                                 = "CreateListener"
	MTypeUpdateNLBSecurityGroupIds                      = "UpdateSecurityGroupIds"
	MTypeTagResource                                    = "TagResource"
	MTypeCreateNLB                                      = "CreateNLB"
	MTypeDeleteNLB                                      = "DeleteNLB"
	MTypeUpdateNLB                                      = "UpdateNLB"
	MTypeUpdateAddressType                              = "UpdateAddressType"
	MTypeUpdateNLBZones                                 = "UpdateZones"
	MTypeUpdateLoadBalancerProtection                   = "UpdateLoadBalancerProtection"
	MTypeAttachCommonBandwidthPackage                   = "AttachCommonBandwidthPackage"
	MTypeDetachCommonBandwidthPackage                   = "DetachCommonBandwidthPackage"
	MTypeCreateServerGroup                              = "CreateServerGroup"
	MTypeDeleteServerGroup                              = "DeleteServerGroup"
	MTypeUpdateNLBServerGroup                           = "UpdateServerGroup"
	MTypeAddServers                                     = "AddServers"
	MTypeRemoveNLBServers                               = "RemoveServers"
	MTypeUpdateNLBServers                               = "UpdateServers"
	MTypeStartListener                                  = "StartListener"
	MTypeStopListener                                   = "StopListener"
	MTypeUpdateIPv6AddressType                          = "UpdateIPv6AddressType"
	MTypeBatchWaitJobsFinish                            = "BatchWaitJobsFinish"
	MTypeAssociateAdditionalCertificatesWithListener    = "AssociateAdditionalCertificatesWithListener"
	MTypeDisassociateAdditionalCertificatesWithListener = "DisassociateAdditionalCertificatesWithListener"
	MTypeWaitJobFinish                                  = "WaitJobFinish"
)

func (d DryRunNLB) DeleteNLBListenerAsync(ctx context.Context, listenerId string) (string, error) {
	mtype := MTypeDeleteListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeDeleteListener, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("listener %s should be deleted", listenerId))
}

func (d DryRunNLB) UpdateNLBListenerAsync(ctx context.Context, lis *nlbmodel.ListenerAttribute) (string, error) {
	mtype := MTypeUpdateListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), lis.ListenerId, CodeUpdateListener, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("listener %s should be updated", lis.ListenerId))
}

func (d DryRunNLB) CreateNLBListenerAsync(ctx context.Context, lbId string, lis *nlbmodel.ListenerAttribute) (string, error) {
	mtype := MTypeCreateListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), lbId, CodeCreateListener, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("listener for lb %s should be created", lbId))
}

func (d DryRunNLB) UpdateNLBSecurityGroupIds(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer, added, removed []string) error {
	mtype := MTypeUpdateNLBSecurityGroupIds
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, CodeUpdateSecurityGroupIds, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s security groups should be updated, added: %v, removed: %v", mdl.LoadBalancerAttribute.LoadBalancerId, added, removed))
}

func (d DryRunNLB) TagNLBResource(ctx context.Context, resourceId string, resourceType nlbmodel.TagResourceType, tags []tag.Tag) error {
	mtype := MTypeTagResource
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), resourceId, CodeTagResource, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("resource %s type %s should be tagged with %v", resourceId, resourceType, tags))
}

func (d DryRunNLB) UntagNLBResources(ctx context.Context, resourceId string, resourceType nlbmodel.TagResourceType, tagKey []*string) error {
	mtype := MTypeUntagResources
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), resourceId, CodeUntagResources, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("resource %s type %s should be untagged with %v", resourceId, resourceType, tagKey))
}

func (d DryRunNLB) NLBJoinSecurityGroup(ctx context.Context, lbId string, sgIds []string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) NLBLeaveSecurityGroup(ctx context.Context, lbId string, sgIds []string) error {
	//TODO implement me
	panic("implement me")
}

func (d DryRunNLB) ListNLBTagResources(ctx context.Context, lbId string) ([]tag.Tag, error) {
	return d.nlb.ListNLBTagResources(ctx, lbId)
}

func (d DryRunNLB) FindNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return d.nlb.FindNLB(ctx, mdl)
}

func (d DryRunNLB) DescribeNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return d.nlb.DescribeNLB(ctx, mdl)
}

func (d DryRunNLB) CreateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer, clientToken string) error {
	mtype := MTypeCreateNLB
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeCreateNLB, dryrun.ERROR, "")
	return hintError(mtype, "nlb should be created")
}

func (d DryRunNLB) DeleteNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mtype := MTypeDeleteNLB
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, CodeDeleteNLB, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s should be deleted", mdl.LoadBalancerAttribute.LoadBalancerId))
}

func (d DryRunNLB) UpdateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mtype := MTypeUpdateNLB
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, CodeUpdateNLB, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s should be updated", mdl.LoadBalancerAttribute.LoadBalancerId))
}

func (d DryRunNLB) UpdateNLBAddressType(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mtype := MTypeUpdateAddressType
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, CodeUpdateAddressType, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s address type should be updated to %s", mdl.LoadBalancerAttribute.LoadBalancerId, mdl.LoadBalancerAttribute.AddressType))
}

func (d DryRunNLB) UpdateNLBZones(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mtype := MTypeUpdateNLBZones
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, CodeUpdateZones, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s zones should be updated", mdl.LoadBalancerAttribute.LoadBalancerId))
}

func (d DryRunNLB) UpdateLoadBalancerProtection(ctx context.Context, lbId string,
	delCfg *nlbmodel.DeletionProtectionConfig, modCfg *nlbmodel.ModificationProtectionConfig) error {
	mtype := MTypeUpdateLoadBalancerProtection
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), lbId, CodeUpdateLoadBalancerProtection, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s protection configs should be updated", lbId))
}

func (d DryRunNLB) AttachCommonBandwidthPackageToLoadBalancer(ctx context.Context, lbId string, bandwidthPackageId string) error {
	mtype := MTypeAttachCommonBandwidthPackage
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), lbId, CodeAttachCommonBandwidthPackage, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s should attach bandwidth package %s", lbId, bandwidthPackageId))
}

func (d DryRunNLB) DetachCommonBandwidthPackageFromLoadBalancer(ctx context.Context, lbId string, bandwidthPackageId string) error {
	mtype := MTypeDetachCommonBandwidthPackage
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), lbId, CodeDetachCommonBandwidthPackage, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s should detach bandwidth package %s", lbId, bandwidthPackageId))
}

func (d DryRunNLB) ListNLBServerGroups(ctx context.Context, tags []tag.Tag) ([]*nlbmodel.ServerGroup, error) {
	return d.nlb.ListNLBServerGroups(ctx, tags)
}

func (d DryRunNLB) CreateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	mtype := MTypeCreateServerGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), sg.ServerGroupName), "", CodeCreateServerGroup, dryrun.ERROR, "")
	return hintError(mtype, "server group should be created")
}

func (d DryRunNLB) DeleteNLBServerGroup(ctx context.Context, sgId string) error {
	mtype := MTypeDeleteServerGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeDeleteServerGroup, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("server group %s should be deleted", sgId))
}

func (d DryRunNLB) UpdateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	mtype := MTypeUpdateNLBServerGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sg.ServerGroupId, CodeUpdateServerGroup, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("server group %s should be updated", sg.ServerGroupId))
}

func (d DryRunNLB) AddNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	mtype := MTypeAddServers
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeAddServers, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("servers should be added to server group %s", sgId))
}

func (d DryRunNLB) RemoveNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	mtype := MTypeRemoveNLBServers
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeRemoveServers, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("servers should be removed from server group %s", sgId))
}

func (d DryRunNLB) UpdateNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	mtype := MTypeUpdateNLBServers
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeUpdateServers, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("servers should be updated in server group %s", sgId))
}

func (d DryRunNLB) ListNLBListeners(ctx context.Context, lbId string) ([]*nlbmodel.ListenerAttribute, error) {
	return d.nlb.ListNLBListeners(ctx, lbId)
}

func (d DryRunNLB) CreateNLBListener(ctx context.Context, lbId string, lis *nlbmodel.ListenerAttribute) error {
	mtype := MTypeCreateListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%d", util.Key(svc), lis.ListenerPort), lbId, CodeCreateListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("listener for nlb %s should be created", lbId))
}

func (d DryRunNLB) UpdateNLBListener(ctx context.Context, lis *nlbmodel.ListenerAttribute) error {
	mtype := MTypeUpdateListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%d", util.Key(svc), lis.ListenerPort), lis.ListenerId, CodeUpdateListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("listener %s should be updated", lis.ListenerId))
}

func (d DryRunNLB) DeleteNLBListener(ctx context.Context, listenerId string) error {
	mtype := MTypeDeleteListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), listenerId), listenerId, CodeDeleteListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("listener %s should be deleted", listenerId))
}

func (d DryRunNLB) StartNLBListener(ctx context.Context, listenerId string) error {
	mtype := MTypeStartListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), listenerId), listenerId, CodeStartListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("listener %s should be started", listenerId))
}

func (d DryRunNLB) StopNLBListener(ctx context.Context, listenerId string) error {
	mtype := MTypeStopListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), listenerId), listenerId, CodeStopListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("listener %s should be stopped", listenerId))
}

func (d DryRunNLB) UpdateNLBIPv6AddressType(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mtype := MTypeUpdateIPv6AddressType
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, CodeUpdateIPv6AddressType, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("nlb %s IPv6 address type should be updated", mdl.LoadBalancerAttribute.LoadBalancerId))
}

func (d DryRunNLB) GetNLBServerGroup(ctx context.Context, sgId string) (*nlbmodel.ServerGroup, error) {
	return d.nlb.GetNLBServerGroup(ctx, sgId)
}

func (d DryRunNLB) CreateNLBServerGroupAsync(ctx context.Context, sg *nlbmodel.ServerGroup) (string, error) {
	mtype := MTypeCreateServerGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), sg.ServerGroupName), "", CodeCreateServerGroup, dryrun.ERROR, "")
	return "", hintError(mtype, "server group should be created")
}

func (d DryRunNLB) DeleteNLBServerGroupAsync(ctx context.Context, sgId string) (string, error) {
	mtype := MTypeDeleteServerGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeDeleteServerGroup, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("server group %s should be deleted", sgId))
}

func (d DryRunNLB) UpdateNLBServerGroupAsync(ctx context.Context, sg *nlbmodel.ServerGroup) (string, error) {
	mtype := MTypeUpdateNLBServerGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sg.ServerGroupId, CodeUpdateServerGroup, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("server group %s should be updated", sg.ServerGroupId))
}

func (d DryRunNLB) AddNLBServersAsync(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) (string, error) {
	mtype := MTypeAddServers
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeAddServers, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("servers should be added to server group %s", sgId))
}

func (d DryRunNLB) RemoveNLBServersAsync(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) (string, error) {
	mtype := MTypeRemoveNLBServers
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeRemoveServers, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("servers should be removed from server group %s", sgId))
}

func (d DryRunNLB) UpdateNLBServersAsync(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) (string, error) {
	mtype := MTypeUpdateNLBServers
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), sgId, CodeUpdateServers, dryrun.ERROR, "")
	return "", hintError(mtype, fmt.Sprintf("servers should be updated in server group %s", sgId))
}

func (d DryRunNLB) BatchWaitJobsFinish(ctx context.Context, api string, jobIds []string, args ...time.Duration) error {
	mtype := MTypeBatchWaitJobsFinish
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeBatchWaitJobsFinish, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("jobs %v for api %s should finish", jobIds, api))
}

func (d DryRunNLB) ListNLBListenerCertificates(ctx context.Context, listenerId string) ([]nlbmodel.ListenerCertificate, error) {
	return d.nlb.ListNLBListenerCertificates(ctx, listenerId)
}

func (d DryRunNLB) AssociateAdditionalCertificatesWithNLBListener(ctx context.Context, listenerId string, certIds []string) error {
	mtype := MTypeAssociateAdditionalCertificatesWithListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), listenerId), listenerId, CodeAssociateAdditionalCertificatesWithListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("additional certificates should be associated with listener %s", listenerId))
}

func (d DryRunNLB) DisassociateAdditionalCertificatesWithNLBListener(ctx context.Context, listenerId string, certIds []string) error {
	mtype := MTypeDisassociateAdditionalCertificatesWithListener
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, fmt.Sprintf("%s/%s", util.Key(svc), listenerId), listenerId, CodeDisassociateAdditionalCertificatesWithListener, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("additional certificates should be disassociated with listener %s", listenerId))
}

func (d DryRunNLB) WaitJobFinish(ctx context.Context, api, jobId string, args ...time.Duration) error {
	mtype := MTypeWaitJobFinish
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), jobId, CodeWaitJobFinish, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("job %s for api %s should finish", jobId, api))
}
