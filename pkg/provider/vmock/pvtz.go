package vmock

import (
	"context"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
)

func NewPVTZProvider(
	auth *metadata.ClientAuth,
) *MockPVTZ {
	return &MockPVTZ{auth: auth}
}

type MockPVTZ struct {
	auth *metadata.ClientAuth
}

func (p *MockPVTZ) ListPVTZ(ctx context.Context) ([]*prvd.PvtzEndpoint, error) {
	panic("implement me")
}

func (p *MockPVTZ) SearchPVTZ(ctx context.Context, ep *prvd.PvtzEndpoint, exact bool) ([]*prvd.PvtzEndpoint, error) {
	panic("implement me")
}

func (p *MockPVTZ) UpdatePVTZ(ctx context.Context, ep *prvd.PvtzEndpoint) error {
	panic("implement me")
}

func (p *MockPVTZ) DeletePVTZ(ctx context.Context, ep *prvd.PvtzEndpoint) error {
	panic("implement me")
}
