package dryrun

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/dryrun"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	ecsmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/ecs"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/ecs"
)

func NewDryRunECS(
	auth *base.ClientMgr,
	ecs *ecs.ECSProvider,
) *DryRunECS {
	return &DryRunECS{auth: auth, ecs: ecs}
}

type DryRunECS struct {
	auth *base.ClientMgr
	ecs  *ecs.ECSProvider
}

const (
	MTypeAuthorizeSecurityGroup       = "AuthorizeSecurityGroup"
	MTypeDeleteSecurityGroup          = "DeleteSecurityGroup"
	MTypeRevokeSecurityGroup          = "RevokeSecurityGroup"
	MTypeModifySecurityGroupAttribute = "ModifySecurityGroupAttribute"
	MTypeCreateSecurityGroup          = "CreateSecurityGroup"
	MTypeModifySecurityGroupRule      = "ModifySecurityGroupRule"
)

var _ prvd.IInstance = &DryRunECS{}

func (d *DryRunECS) ListInstances(ctx context.Context, ids []string) (map[string]*prvd.NodeAttribute, error) {
	return d.ecs.ListInstances(ctx, ids)
}

func (d *DryRunECS) GetInstancesByIP(ctx context.Context, ips []string) (*prvd.NodeAttribute, error) {
	return d.ecs.GetInstancesByIP(ctx, ips)
}

func (d *DryRunECS) DescribeNetworkInterfaces(vpcId string, ips []string, ipVersionType model.AddressIPVersionType) (map[string]string, error) {
	return d.ecs.DescribeNetworkInterfaces(vpcId, ips, ipVersionType)
}

func (d *DryRunECS) DescribeNetworkInterfacesByIDs(ids []string) ([]*prvd.EniAttribute, error) {
	return d.ecs.DescribeNetworkInterfacesByIDs(ids)
}

func (d *DryRunECS) ModifyNetworkInterfaceSourceDestCheck(id string, enabled bool) error {
	klog.Infof("[DryRun] ModifyNetworkInterfaceSourceDestCheck: id=%s, enabled=%t", id, enabled)
	return nil
}

func (d *DryRunECS) AuthorizeSecurityGroup(ctx context.Context, sgId string, permissions []ecsmodel.SecurityGroupPermission) error {
	mtype := MTypeAuthorizeSecurityGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeAuthorizeSecurityGroup, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("need to authorize security group %s", sgId))
}

func (d *DryRunECS) DescribeSecurityGroupAttribute(ctx context.Context, sgId string) (ecsmodel.SecurityGroup, error) {
	return d.ecs.DescribeSecurityGroupAttribute(ctx, sgId)
}

func (d *DryRunECS) DeleteSecurityGroup(ctx context.Context, sgId string) error {
	mtype := MTypeDeleteSecurityGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeDeleteSecurityGroup, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("need to delete security group %s", sgId))
}

func (d *DryRunECS) DescribeSecurityGroups(ctx context.Context, tags []tag.Tag) ([]ecsmodel.SecurityGroup, error) {
	return d.ecs.DescribeSecurityGroups(ctx, tags)
}

func (d *DryRunECS) RevokeSecurityGroup(ctx context.Context, sgId string, permissions []ecsmodel.SecurityGroupPermission) error {
	mtype := MTypeRevokeSecurityGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeRevokeSecurityGroup, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("need to revoke security group %s", sgId))
}

func (d *DryRunECS) ModifySecurityGroupAttribute(ctx context.Context, sgId string, sg *ecsmodel.SecurityGroup) error {
	mtype := MTypeModifySecurityGroupAttribute
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeModifySecurityGroupAttribute, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("need to modify security group attribute %s", sgId))
}

func (d *DryRunECS) CreateSecurityGroup(ctx context.Context, sg ecsmodel.SecurityGroup) error {
	mtype := MTypeCreateSecurityGroup
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeCreateSecurityGroup, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("need to create security group %s", sg.Name))
}

func (d *DryRunECS) ModifySecurityGroupRule(ctx context.Context, sgId string, permission ecsmodel.SecurityGroupPermission) error {
	mtype := MTypeModifySecurityGroupRule
	svc := getService(ctx)
	dryrun.AddEvent(dryrun.NLB, util.Key(svc), "", CodeModifySecurityGroupRule, dryrun.ERROR, "")
	return hintError(mtype, fmt.Sprintf("need to modify security group rule %s", sgId))
}
