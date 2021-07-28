package dryrun

import (
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"
)

func NewDryRunVPC(
	auth *base.ClientMgr,
	vpc *vpc.VPCProvider,
) *DryRunVPC {
	return &DryRunVPC{auth: auth, vpc: vpc}
}

type DryRunVPC struct {
	auth        *base.ClientMgr
	vpc         *vpc.VPCProvider
	routetables []string
}

func (m *DryRunVPC) CreateRoute(ctx context.Context, table string, provideID string, destinationCIDR string) (*model.Route, error) {
	return m.vpc.CreateRoute(ctx, table, provideID, destinationCIDR)
}

func (m *DryRunVPC) DeleteRoute(ctx context.Context, table, provideID, destinationCIDR string) error {
	return m.vpc.DeleteRoute(ctx, table, provideID, destinationCIDR)
}

func (m *DryRunVPC) ListRoute(ctx context.Context, table string) ([]*model.Route, error) {
	return m.vpc.ListRoute(ctx, table)
}

func (m *DryRunVPC) FindRoute(ctx context.Context, table, pvid, cidr string) (*model.Route, error) {
	return m.vpc.FindRoute(ctx, table, pvid, cidr)
}

func (m *DryRunVPC) ListRouteTables(ctx context.Context, vpcID string) ([]string, error) {
	if len(m.routetables) == 0 {
		return nil, fmt.Errorf("no route table")
	}
	return m.routetables, nil
}

func (m *DryRunVPC) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	return m.vpc.DescribeEipAddresses(ctx, instanceType, instanceId)
}
