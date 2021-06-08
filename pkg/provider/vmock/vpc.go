package vmock

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockVPC(
	auth *base.ClientAuth,
	tables []string,
) *MockVPC {
	return &MockVPC{auth: auth, routetables: tables}
}

type MockVPC struct {
	auth        *base.ClientAuth
	routetables []string
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
	if len(m.routetables) == 0 {
		return nil, fmt.Errorf("no route table")
	}
	return m.routetables, nil
}

func (m *MockVPC) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	panic("implement me")
}
