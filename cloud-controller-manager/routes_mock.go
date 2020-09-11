package alicloud

import (
	"context"
	"fmt"
	"github.com/denverdino/aliyungo/cen"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"reflect"
	"sync"
)

type RouteStore struct {
	// ecs.VpcSetType
	vpcs sync.Map

	// ecs.VRouterSetType
	routers sync.Map

	// ecs.RouteTableSetType
	tables sync.Map
}

func key(region, id string) string {
	return fmt.Sprintf("%s/%s", region, id)
}

var ROUTES = RouteStore{}

type mockRouteSDK struct {
	describeVpcs                    func(args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error)
	describeVRouters                func(args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error)
	describeRouteTables             func(args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error)
	deleteRouteEntry                func(args *ecs.DeleteRouteEntryArgs) error
	createRouteEntry                func(args *ecs.CreateRouteEntryArgs) error
	waitForAllRouteEntriesAvailable func(vrouterId string, routeTableId string, timeout int) error
	describeRouteEntryList          func(args *ecs.DescribeRouteEntryListArgs) (response *ecs.DescribeRouteEntryListResponse, err error)
}

func WithNewRouteStore() CloudDataMock {
	return func() {
		ROUTES = RouteStore{}
	}
}

func WithVpcs() CloudDataMock {
	return func() {
		ROUTES.vpcs.Store(
			key(string(REGION), VPCID),
			ecs.VpcSetType{
				VpcId:    VPCID,
				RegionId: REGION,
				VSwitchIds: struct {
					VSwitchId []string
				}{
					VSwitchId: []string{VSWITCH_ID},
				},
				CidrBlock: "192.168.0.0/16",
				VRouterId: VROUTER_ID,
				RouterTableIds: struct {
					RouterTableIds []string
				}{
					RouterTableIds: []string{ROUTE_TABLE_ID},
				},
			},
		)
	}
}

func WithVRouter() CloudDataMock {
	return func() {
		ROUTES.routers.Store(
			key(string(REGION), VROUTER_ID),
			ecs.VRouterSetType{
				VRouterId: VROUTER_ID,
				RegionId:  REGION,
				VpcId:     VPCID,
				RouteTableIds: struct {
					RouteTableId []string
				}{
					RouteTableId: []string{ROUTE_TABLE_ID},
				},
			},
		)
	}
}

func WithRouteTableEntrySet() CloudDataMock {
	return func() {
		ROUTES.tables.Store(
			ROUTE_TABLE_ID,
			ecs.RouteTableSetType{
				VRouterId:    VROUTER_ID,
				RouteTableId: ROUTE_TABLE_ID,
				RouteEntrys: struct {
					RouteEntry []ecs.RouteEntrySetType
				}{
					RouteEntry: ROUTE_ENTRIES,
				},
				RouteTableType: "System",
			},
		)
	}
}

func (m *mockRouteSDK) DescribeVpcs(ctx context.Context, args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error) {
	if m.describeVpcs != nil {
		return m.describeVpcs(args)
	}
	if args.VpcId == "" {
		return nil, nil, fmt.Errorf("no VPCID specified")
	}
	vpc, ok := ROUTES.vpcs.Load(key(string(args.RegionId), args.VpcId))
	if !ok {
		return []ecs.VpcSetType{}, nil, nil
	}
	result, ok := vpc.(ecs.VpcSetType)
	if !ok {
		return nil, nil, fmt.Errorf("not type ecs.VpcSetType %s", reflect.TypeOf(vpc))
	}
	return []ecs.VpcSetType{result}, nil, nil
}

func (m *mockRouteSDK) DescribeVRouters(ctx context.Context, args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error) {
	if m.describeVRouters != nil {
		return m.describeVRouters(args)
	}
	if args.VRouterId == "" {
		return nil, nil, fmt.Errorf("no vrouteid specified")
	}
	vrouter, ok := ROUTES.routers.Load(key(string(args.RegionId), args.VRouterId))
	if !ok {
		return []ecs.VRouterSetType{}, nil, nil
	}
	result, ok := vrouter.(ecs.VRouterSetType)
	if !ok {
		return nil, nil, fmt.Errorf("not type ecs.VRouterSetType %s", reflect.TypeOf(vrouter))
	}
	return []ecs.VRouterSetType{result}, nil, nil
}

func (m *mockRouteSDK) DescribeRouteTables(ctx context.Context, args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error) {
	if m.describeRouteTables != nil {
		return m.describeRouteTables(args)
	}
	if args.RouteTableId == "" {
		return nil, nil, fmt.Errorf("no routetalbeid specified")
	}
	vrouter, ok := ROUTES.tables.Load(args.RouteTableId)
	if !ok {
		return []ecs.RouteTableSetType{}, nil, nil
	}
	result, ok := vrouter.(ecs.RouteTableSetType)
	if !ok {
		return nil, nil, fmt.Errorf("not type ecs.RouteTableSetType %s", reflect.TypeOf(vrouter))
	}
	return []ecs.RouteTableSetType{result}, nil, nil
}

func (m *mockRouteSDK) DeleteRouteEntry(ctx context.Context, args *ecs.DeleteRouteEntryArgs) error {
	if m.deleteRouteEntry != nil {
		return m.deleteRouteEntry(args)
	}
	vrouter, ok := ROUTES.tables.Load(args.RouteTableId)
	if !ok {
		return fmt.Errorf("not found %s", args.RouteTableId)
	}
	result, ok := vrouter.(ecs.RouteTableSetType)
	if !ok {
		return fmt.Errorf("not type ecs.RouteTableSetType %s, DeleteRouteEntry", reflect.TypeOf(vrouter))
	}
	var entries []ecs.RouteEntrySetType
	for _, v := range result.RouteEntrys.RouteEntry {
		if v.RouteTableId == args.RouteTableId &&
			v.NextHopId == args.NextHopId &&
			v.DestinationCidrBlock == args.DestinationCidrBlock {
			// delete
			continue
		}
		entries = append(entries, v)
	}
	result.RouteEntrys.RouteEntry = entries
	ROUTES.tables.Store(args.RouteTableId, result)
	return nil
}
func (m *mockRouteSDK) CreateRouteEntry(ctx context.Context, args *ecs.CreateRouteEntryArgs) error {
	if m.createRouteEntry != nil {
		return m.createRouteEntry(args)
	}
	vrouter, ok := ROUTES.tables.Load(args.RouteTableId)
	if !ok {
		return fmt.Errorf("not found %s", args.RouteTableId)
	}
	result, ok := vrouter.(ecs.RouteTableSetType)
	if !ok {
		return fmt.Errorf("not type ecs.RouteTableSetType %s, DeleteRouteEntry", reflect.TypeOf(vrouter))
	}
	found := false

	for _, v := range result.RouteEntrys.RouteEntry {
		if v.RouteTableId == args.RouteTableId &&
			v.NextHopId == args.NextHopId &&
			v.DestinationCidrBlock == args.DestinationCidrBlock {
			// delete
			found = true
			break
		}
	}
	if !found {
		route := ecs.RouteEntrySetType{
			RouteTableId:         args.RouteTableId,
			DestinationCidrBlock: args.DestinationCidrBlock,
			NextHopId:            args.NextHopId,
			NextHopType:          string(args.NextHopType),
		}
		result.RouteEntrys.RouteEntry = append(result.RouteEntrys.RouteEntry, route)
	}

	ROUTES.tables.Store(args.RouteTableId, result)
	return nil
}

func (m *mockRouteSDK) PublishRouteEntry(ctx context.Context, args *cen.PublishRouteEntriesArgs) error {
	return nil
}

func (m *mockRouteSDK) WaitForAllRouteEntriesAvailable(ctx context.Context, vrouterId string, routeTableId string, timeout int) error {
	if m.waitForAllRouteEntriesAvailable != nil {
		return m.waitForAllRouteEntriesAvailable(vrouterId, routeTableId, timeout)
	}
	return nil
}

func (m *mockRouteSDK) DescribeRouteEntryList(ctx context.Context, args *ecs.DescribeRouteEntryListArgs) (response *ecs.DescribeRouteEntryListResponse, err error) {
	if m.describeRouteEntryList != nil {
		return m.describeRouteEntryList(args)
	}
	response = &ecs.DescribeRouteEntryListResponse{}
	if args.RouteTableId == "" {
		return nil, fmt.Errorf("no routetalbeid specified")
	}
	vrouter, ok := ROUTES.tables.Load(args.RouteTableId)
	if !ok {
		return response, nil
	}
	result, ok := vrouter.(ecs.RouteTableSetType)
	if !ok {
		return response, fmt.Errorf("not type ecs.RouteTableSetType %s", reflect.TypeOf(vrouter))
	}

	for _, e := range result.RouteEntrys.RouteEntry {
		routeEntry := ecs.RouteEntry{
			DestinationCidrBlock: e.DestinationCidrBlock,
			IpVersion:            "",
			RouteEntryId:         "",
			RouteEntryName:       "",
			RouteTableId:         e.RouteTableId,
			Status:               string(e.Status),
			Type:                 string(e.Type),
			NextHops: struct {
				NextHop []ecs.NextHop
			}{
				NextHop: []ecs.NextHop{
					{
						NextHopId:   e.InstanceId,
						NextHopType: e.NextHopType,
					},
				},
			},
		}
		response.RouteEntrys.RouteEntry = append(response.RouteEntrys.RouteEntry, routeEntry)
	}
	return response, nil
}
