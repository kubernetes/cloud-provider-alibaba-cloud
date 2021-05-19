package vmock

import (
	"context"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/auth"
)

func NewMockCLB(
	auth *auth.ClientAuth,
) *MockCLB {
	return &MockCLB{auth: auth}
}

type MockCLB struct {
	auth *auth.ClientAuth
}

func (MockCLB) FindSLB(ctx context.Context, id string) ([]prvd.SLB, error) {
	panic("implement me")
}

func (MockCLB) ListSLB(ctx context.Context, slb prvd.SLB) ([]prvd.SLB, error) {
	panic("implement me")
}

func (MockCLB) CreateSLB(ctx context.Context, slb prvd.SLB) error {
	panic("implement me")
}

func (MockCLB) DeleteSLB(ctx context.Context, slb prvd.SLB) error {
	panic("implement me")
}
