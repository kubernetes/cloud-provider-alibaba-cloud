package vmock

import (
	"context"
	"fmt"
	"net"

	servicesvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"
)

func NewMockVPC(
	auth *base.ClientMgr,
) *MockVPC {
	return &MockVPC{auth: auth}
}

var _ prvd.IVPC = &MockVPC{}

type MockVPC struct {
	auth *base.ClientMgr
}

func (m *MockVPC) CreateRoutes(ctx context.Context, table string, routes []*model.Route) ([]string, []prvd.RouteUpdateStatus, error) {
	var ids []string
	var status []prvd.RouteUpdateStatus
	for _, r := range routes {
		ids = append(ids, fmt.Sprintf("r-%s", r.ProviderId))
		s := prvd.RouteUpdateStatus{Route: r, Failed: false}
		if r.NodeReference != nil {
			switch r.NodeReference.Name {
			case "dup-cidr":
				s.Failed = true
				s.FailedCode = "VPC_ROUTE_ENTRY_CIDR_BLOCK_DUPLICATE"
				s.FailedMessage = "duplicate"
			case "status-err":
				s.Failed = true
				s.FailedCode = "VPC_ROUTE_ENTRY_STATUS_ERROR"
				s.FailedMessage = "middle status"
			case "create-fail":
				s.Failed = true
				s.FailedCode = "OTHER"
				s.FailedMessage = "create failed"
			}
		}
		status = append(status, s)
	}
	return ids, status, nil
}

func (m *MockVPC) CreateRoute(ctx context.Context, table string, providerID string, destinationCIDR string) (*model.Route, error) {
	if providerID == "i-duplicate" && destinationCIDR == "10.96.0.0/24" {
		return nil, fmt.Errorf("InvalidCIDRBlock.Duplicate: cidr already exists")
	}
	if providerID == "i-dup-find-fail" && destinationCIDR == "10.96.0.0/24" {
		return nil, fmt.Errorf("InvalidCIDRBlock.Duplicate: cidr already exists")
	}
	return &model.Route{
		DestinationCIDR: destinationCIDR,
		ProviderId:      providerID,
	}, nil
}

func (m *MockVPC) DeleteRoute(ctx context.Context, table, provideID, destinationCIDR string) error {
	return nil
}
func (m *MockVPC) DeleteRoutes(ctx context.Context, table string, routes []*model.Route) ([]prvd.RouteUpdateStatus, error) {
	var status []prvd.RouteUpdateStatus
	for _, r := range routes {
		s := prvd.RouteUpdateStatus{Route: r, Failed: false}
		if r.ProviderId == "i-delete-not-exist" {
			s.Failed = true
			s.FailedCode = "VPC_ROUTER_ENTRY_NOT_EXIST"
			s.FailedMessage = "route not found"
		} else if r.ProviderId == "i-delete-fail" {
			s.Failed = true
			s.FailedCode = "OTHER"
			s.FailedMessage = "delete failed"
		}
		status = append(status, s)
	}
	return status, nil
}

func (m *MockVPC) ListRoute(ctx context.Context, table string) ([]*model.Route, error) {
	if table == "route-table-list-err" {
		return nil, fmt.Errorf("list route error")
	}
	if table == "route-table-invalid-cidr" {
		return []*model.Route{
			{DestinationCIDR: "invalid", ProviderId: "i-other", Name: "bad-route"},
		}, nil
	}
	if table == "route-table-1" {
		return []*model.Route{
			{DestinationCIDR: "10.96.0.64/26", ProviderId: "i-other", Name: "conflict-route"},
		}, nil
	}
	return []*model.Route{}, nil
}

func (m *MockVPC) FindRoute(ctx context.Context, table, pvid, cidr string) (*model.Route, error) {
	// Simulate finding a route based on provider ID and CIDR
	switch {
	case pvid == "i-123" && cidr == "192.168.1.0/24":
		return &model.Route{
			Name:            "route-1",
			DestinationCIDR: "192.168.1.0/24",
			ProviderId:      "i-123",
		}, nil
	case pvid == "i-456" && cidr == "":
		return &model.Route{
			Name:            "route-2",
			DestinationCIDR: "192.168.2.0/24",
			ProviderId:      "i-456",
		}, nil
	case pvid == "" && cidr == "192.168.1.0/24":
		return &model.Route{
			Name:            "route-1",
			DestinationCIDR: "192.168.1.0/24",
			ProviderId:      "i-123",
		}, nil
	case pvid == "" && cidr == "":
		return nil, fmt.Errorf("empty query condition")
	case pvid == "i-duplicate" && cidr == "10.96.0.0/24":
		return &model.Route{
			Name:            "route-dup",
			DestinationCIDR: "10.96.0.0/24",
			ProviderId:      "i-duplicate",
		}, nil
	case pvid == "i-dup-find-fail" && cidr == "10.96.0.0/24":
		return nil, nil
	}

	return nil, nil
}

func (m *MockVPC) ListRouteTables(ctx context.Context, vpcID string) ([]string, error) {
	switch vpcID {
	case "vpc-no-route-table":
		return []string{}, nil
	case "vpc-single-route-table":
		return []string{"route-table-1"}, nil
	case "vpc-multi-route-table":
		return []string{"route-table-1", "route-table-2"}, nil
	}
	return []string{}, nil
}

func (m *MockVPC) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	if instanceType == string(vpc.SlbInstance) {
		return []string{"172.16.0.10"}, nil
	}
	return nil, nil
}

func (m *MockVPC) DescribeVSwitches(ctx context.Context, vpcID string) ([]servicesvpc.VSwitch, error) {
	panic("implement me")
}

func (m *MockVPC) DescribeVswitchByID(ctx context.Context, vswId string) (servicesvpc.VSwitch, error) {
	panic("implement me")
}

func (m *MockVPC) DescribeVpcCIDRBlock(ctx context.Context, vpcId string, ipVersion model.AddressIPVersionType) ([]*net.IPNet, error) {
	return []*net.IPNet{
		{
			IP:   net.ParseIP("10.96.0.0"),
			Mask: net.CIDRMask(16, 32),
		},
	}, nil
}
