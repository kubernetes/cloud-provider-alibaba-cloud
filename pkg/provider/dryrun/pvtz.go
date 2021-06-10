package dryrun

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/pvtz"
)

func NewDryRunPVTZ(
	auth *base.ClientAuth,
	pvtz *pvtz.PVTZProvider,
) *DryRunPVTZ {
	return &DryRunPVTZ{auth: auth, pvtz: pvtz}
}

type DryRunPVTZ struct {
	auth *base.ClientAuth
	pvtz *pvtz.PVTZProvider
}

func (p *DryRunPVTZ) ListPVTZ(ctx context.Context) ([]*model.PvtzEndpoint, error) {
	return p.pvtz.ListPVTZ(ctx)
}

func (p *DryRunPVTZ) SearchPVTZ(ctx context.Context, ep *model.PvtzEndpoint, exact bool) ([]*model.PvtzEndpoint, error) {
	return p.pvtz.SearchPVTZ(ctx, ep, exact)
}

func (p *DryRunPVTZ) UpdatePVTZ(ctx context.Context, ep *model.PvtzEndpoint) error {
	return p.pvtz.UpdatePVTZ(ctx, ep)
}

func (p *DryRunPVTZ) DeletePVTZ(ctx context.Context, ep *model.PvtzEndpoint) error {
	return p.pvtz.DeletePVTZ(ctx, ep)
}
