package dryrun

import (
	"context"
	"fmt"
	"net"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"

	servicesvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
)

func NewDryRunVPC(
	auth *base.ClientMgr,
	vpc *vpc.VPCProvider,
) *DryRunVPC {
	return &DryRunVPC{auth: auth, vpc: vpc}
}

var _ prvd.IVPC = &DryRunVPC{}

type DryRunVPC struct {
	auth        *base.ClientMgr
	vpc         *vpc.VPCProvider
	routetables []string
}

func (m *DryRunVPC) CreateRoutes(ctx context.Context, table string, routes []*model.Route) ([]string, []prvd.RouteUpdateStatus, error) {
	return m.vpc.CreateRoutes(ctx, table, routes)
}

func (m *DryRunVPC) CreateRoute(ctx context.Context, table string, provideID string, destinationCIDR string) (*model.Route, error) {
	return m.vpc.CreateRoute(ctx, table, provideID, destinationCIDR)
}

func (m *DryRunVPC) DeleteRoute(ctx context.Context, table, provideID, destinationCIDR string) error {
	return m.vpc.DeleteRoute(ctx, table, provideID, destinationCIDR)
}

func (m *DryRunVPC) DeleteRoutes(ctx context.Context, table string, routes []*model.Route) ([]prvd.RouteUpdateStatus, error) {
	return m.vpc.DeleteRoutes(ctx, table, routes)
}

func (m *DryRunVPC) ListRoute(ctx context.Context, table string) ([]*model.Route, error) {
	return m.vpc.ListRoute(ctx, table)
}

func (m *DryRunVPC) FindRoute(ctx context.Context, table, pvid, cidr string) (*model.Route, error) {
	return m.vpc.FindRoute(ctx, table, pvid, cidr)
}

func (m *DryRunVPC) ListRouteTables(ctx context.Context, vpcID string) ([]string, error) {
	if len(m.routetables) == 0 {
		return nil, fmt.Errorf("no route table found in %s", vpcID)
	}
	return m.routetables, nil
}

func (m *DryRunVPC) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	return m.vpc.DescribeEipAddresses(ctx, instanceType, instanceId)
}

func (m *DryRunVPC) DescribeVswitchByID(ctx context.Context, vswId string) (servicesvpc.VSwitch, error) {
	return m.vpc.DescribeVswitchByID(ctx, vswId)
}

func (m *DryRunVPC) DescribeVSwitches(ctx context.Context, vpcID string) ([]servicesvpc.VSwitch, error) {
	return m.vpc.DescribeVSwitches(ctx, vpcID)
}

func (m *DryRunVPC) DescribeVpcCIDRBlock(ctx context.Context, vpcId string, ipVersion model.AddressIPVersionType) ([]*net.IPNet, error) {
	return m.vpc.DescribeVpcCIDRBlock(ctx, vpcId, ipVersion)
}
