package alibaba

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

type AssociatedInstanceType string

const SlbInstance = AssociatedInstanceType("SlbInstance")

func NewVPCProvider(
	auth *ClientAuth,
) *VPCProvider {
	return &VPCProvider{auth: auth}
}

var _ prvd.IVPC = &VPCProvider{}

type VPCProvider struct {
	auth *ClientAuth
}

func (r *VPCProvider) ListRouteTables(ctx context.Context, vpcID string) ([]string, error) {
	tableListRequest := vpc.CreateDescribeRouteTableListRequest()
	tableListRequest.VpcId = vpcID
	resp, err := r.auth.VPC.DescribeRouteTableList(tableListRequest)
	if err != nil {
		return nil, fmt.Errorf("error describe vpc: %v route tables, error: %v", vpcID, err)
	}
	var tableIds []string
	for _, table := range resp.RouterTableList.RouterTableListType {
		tableIds = append(tableIds, table.RouteTableId)
	}
	return tableIds, nil
}

func (r *VPCProvider) FindRoute(ctx context.Context, table, provID, cidr string) (*model.Route, error) {
	describeRouteEntryRequest := vpc.CreateDescribeRouteEntryListRequest()
	describeRouteEntryRequest.RouteTableId = table
	describeRouteEntryRequest.MaxResult = requests.NewInteger(model.RouteMaxQueryRouteEntry)
	describeRouteEntryRequest.RouteEntryType = model.RouteEntryTypeCustom
	if provID != "" {
		_, instance, err := nodeFromProviderID(provID)
		if err != nil {
			return nil, fmt.Errorf("invalid provide id: %v, err: %v", provID, err)
		}
		describeRouteEntryRequest.NextHopId = instance
		describeRouteEntryRequest.NextHopType = model.RouteNextHopTypeInstance
	}

	if cidr != "" {
		describeRouteEntryRequest.DestinationCidrBlock = cidr
	}
	resp, err := r.auth.VPC.DescribeRouteEntryList(describeRouteEntryRequest)
	if err != nil {
		return nil, fmt.Errorf("error describe route entry list: %v", err)
	}
	if len(resp.RouteEntrys.RouteEntry) >= 1 {
		route := &model.Route{
			DestinationCIDR: resp.RouteEntrys.RouteEntry[0].DestinationCidrBlock,
		}
		if len(resp.RouteEntrys.RouteEntry[0].NextHops.NextHop) > 0 &&
			resp.RouteEntrys.RouteEntry[0].NextHops.NextHop[0].NextHopType == model.RouteNextHopTypeInstance {
			region, err := r.auth.Meta.Region()
			if err != nil {
				return nil, fmt.Errorf("error get region id for route entry: %v", err)
			}
			route.ProviderId = providerIDFromInstance(region, resp.RouteEntrys.RouteEntry[0].NextHops.NextHop[0].NextHopId)
		}
		return route, nil
	}
	return nil, nil
}

func (r *VPCProvider) CreateRoute(ctx context.Context, table string, provideID string, destinationCIDR string) (*model.Route, error) {
	createRouteEntryRequest := vpc.CreateCreateRouteEntryRequest()
	createRouteEntryRequest.RouteTableId = table
	createRouteEntryRequest.DestinationCidrBlock = destinationCIDR
	createRouteEntryRequest.NextHopType = model.RouteNextHopTypeInstance
	_, instance, err := nodeFromProviderID(provideID)
	if err != nil {
		return nil, fmt.Errorf("invalid provide id: %v, err: %v", provideID, err)
	}
	createRouteEntryRequest.NextHopId = instance
	_, err = r.auth.VPC.CreateRouteEntry(createRouteEntryRequest)
	if err != nil {
		return nil, fmt.Errorf("error create route entry for %s, %s, error: %v", provideID, destinationCIDR, err)
	}
	return &model.Route{
		Name:            fmt.Sprintf("%-%", provideID, destinationCIDR),
		DestinationCIDR: destinationCIDR,
		ProviderId:      provideID,
	}, nil
}

func (r *VPCProvider) DeleteRoute(ctx context.Context, table, provideID, destinationCIDR string) error {
	deleteRouteEntryRequest := vpc.CreateDeleteRouteEntryRequest()
	deleteRouteEntryRequest.RouteTableId = table
	deleteRouteEntryRequest.DestinationCidrBlock = destinationCIDR
	_, instance, err := nodeFromProviderID(provideID)
	if err != nil {
		return fmt.Errorf("invalid provide id: %v, err: %v", provideID, err)
	}
	deleteRouteEntryRequest.NextHopId = instance
	_, err = r.auth.VPC.DeleteRouteEntry(deleteRouteEntryRequest)
	if err != nil {
		return fmt.Errorf("error delete route entry for %s, %s, error: %v", provideID, destinationCIDR, err)
	}
	return nil
}

func (r *VPCProvider) ListRoute(ctx context.Context, table string) ([]*model.Route, error) {
	panic("implement me")
}

func (p *VPCProvider) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) ([]string, error) {
	req := vpc.CreateDescribeEipAddressesRequest()
	req.AssociatedInstanceType = instanceType
	req.AssociatedInstanceId = instanceId
	var ips []string
	next := &Pagination{
		PageNumber: 1,
		PageSize:   10,
	}

	for {
		req.PageSize = requests.Integer(next.PageSize)
		req.PageNumber = requests.Integer(next.PageNumber)
		resp, err := p.auth.VPC.DescribeEipAddresses(req)
		if err != nil {
			return nil, err
		}

		for _, eip := range resp.EipAddresses.EipAddress {
			ips = append(ips, eip.IpAddress)
		}

		pageResult := &PaginationResult{
			PageNumber: resp.PageNumber,
			PageSize:   resp.PageSize,
			TotalCount: resp.TotalCount,
		}
		next := pageResult.NextPage()
		if next == nil {
			break
		}
	}
	return ips, nil
}
