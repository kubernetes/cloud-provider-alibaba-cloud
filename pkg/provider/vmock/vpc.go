package vmock

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
)

func NewMockVPC(
	auth *alibaba.ClientAuth,
) *MockVPC {
	return &MockVPC{auth: auth}
}

type MockVPC struct {
	auth *alibaba.ClientAuth
}

func (m *MockVPC) CreateRoute() {
	panic("implement me")
}

func (m *MockVPC) DeleteRoute() {
	panic("implement me")
}

func (m *MockVPC) ListRoute() {
	panic("implement me")
}

func (m *MockVPC) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	panic("implement me")
}
