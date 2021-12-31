package vmock

import (
	"context"
	servicesvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockVPC(
	auth *base.ClientMgr,
) *MockVPC {
	return &MockVPC{auth: auth}
}

type MockVPC struct {
	auth *base.ClientMgr
}

func (m *MockVPC) CreateRoute(ctx context.Context, table string, provideID string, destinationCIDR string) (*model.Route, error) {
	panic("implement me")
}

func (m *MockVPC) DeleteRoute(ctx context.Context, table, provideID, destinationCIDR string) error {
	panic("implement me")
}

func (m *MockVPC) ListRoute(ctx context.Context, table string) ([]*model.Route, error) {
	panic("implement me")
}

func (m *MockVPC) FindRoute(ctx context.Context, table, pvid, cidr string) (*model.Route, error) {
	panic("implement me")
}

func (m *MockVPC) ListRouteTables(ctx context.Context, vpcID string) ([]string, error) {
	panic("implement me")
}

func (m *MockVPC) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	panic("implement me")
}
func (m *MockVPC) DescribeVSwitches(ctx context.Context, vpcID string) ([]servicesvpc.VSwitch, error) {
	panic("implement me")
}
