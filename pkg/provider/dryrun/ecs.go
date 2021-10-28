package dryrun

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/ecs"
)

func NewDryRunECS(
	auth *base.ClientMgr,
	ecs *ecs.EcsProvider,
) *DryRunECS {
	return &DryRunECS{auth: auth, ecs: ecs}
}

type DryRunECS struct {
	auth *base.ClientMgr
	ecs  *ecs.EcsProvider
}

var _ prvd.IInstance = &DryRunECS{}

func (d *DryRunECS) ListInstances(ctx *node.NodeContext, ids []string) (map[string]*prvd.NodeAttribute, error) {
	return d.ecs.ListInstances(ctx, ids)
}

func (d *DryRunECS) SetInstanceTags(ctx *node.NodeContext, id string, tags map[string]string) error {
	return d.ecs.SetInstanceTags(ctx, id, tags)
}

func (d *DryRunECS) DescribeNetworkInterfaces(vpcId string, ips []string) (map[string]string, error) {
	return d.ecs.DescribeNetworkInterfaces(vpcId, ips)
}
