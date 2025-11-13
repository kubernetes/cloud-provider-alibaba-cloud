package dryrun

import (
	"context"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/eflo"
)

func NewDryRunEFLO(
	auth *base.ClientMgr, eflo *eflo.EFLOProvider) *DryRunEFLO {
	return &DryRunEFLO{auth: auth, eflo: eflo}
}

var _ prvd.IEFLO = &DryRunEFLO{}

type DryRunEFLO struct {
	auth *base.ClientMgr
	eflo *eflo.EFLOProvider
}

func (d *DryRunEFLO) DescribeLingJunNode(ctx context.Context, id string) (*prvd.EFLONodeAttribute, error) {
	return d.eflo.DescribeLingJunNode(ctx, id)
}
