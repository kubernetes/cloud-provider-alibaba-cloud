package vmock

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
)

func NewMockPVTZ(
	auth *alibaba.ClientAuth,
) *MockPVTZ {
	return &MockPVTZ{auth: auth}
}

type MockPVTZ struct {
	auth *alibaba.ClientAuth
}

func (p *MockPVTZ) ListPVTZ(ctx context.Context) ([]*model.PvtzEndpoint, error) {
	panic("implement me")
}

func (p *MockPVTZ) SearchPVTZ(ctx context.Context, ep *model.PvtzEndpoint, exact bool) ([]*model.PvtzEndpoint, error) {
	panic("implement me")
}

func (p *MockPVTZ) UpdatePVTZ(ctx context.Context, ep *model.PvtzEndpoint) error {
	panic("implement me")
}

func (p *MockPVTZ) DeletePVTZ(ctx context.Context, ep *model.PvtzEndpoint) error {
	panic("implement me")
}
