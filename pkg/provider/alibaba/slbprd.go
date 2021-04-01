package alibaba

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
)

func NewLBProvider(
	auth *metadata.ClientAuth,
) *ProviderSLB {
	return &ProviderSLB{auth: auth}
}

type ProviderSLB struct{
	auth *metadata.ClientAuth
}

func (ProviderSLB) FindSLB(ctx context.Context, id string) ([]prvd.SLB, error) {
	panic("implement me")
}

func (ProviderSLB) ListSLB(ctx context.Context, slb prvd.SLB) ([]prvd.SLB, error) {
	panic("implement me")
}

func (ProviderSLB) CreateSLB(ctx context.Context, slb prvd.SLB) error {
	panic("implement me")
}

func (ProviderSLB) DeleteSLB(ctx context.Context, slb prvd.SLB) error {
	panic("implement me")
}