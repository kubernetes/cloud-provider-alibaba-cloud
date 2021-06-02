package alibaba

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog"
	"strings"
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
	auth   *ClientAuth
	region string
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
			route.ProviderId, err = r.providerIDFromInstanceId(resp.RouteEntrys.RouteEntry[0].NextHops.NextHop[0].NextHopId)
			if err != nil {
				return nil, err
			}
		}
		return route, nil
	}
	return nil, nil
}

func (r *VPCProvider) providerIDFromInstanceId(instanceID string) (pvid string, err error) {
	if r.region == "" {
		r.region, err = r.auth.Meta.Region()
		if err != nil {
			return "", fmt.Errorf("error get region id for route entry: %v", err)
		}
	}
	return providerIDFromInstance(r.region, instanceID), nil
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
		Name:            fmt.Sprintf("%s-%s", provideID, destinationCIDR),
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

func (r *VPCProvider) ListRoute(ctx context.Context, table string) (routes []*model.Route, err error) {
	err = r.listRouteBatch(table, "", &routes)
	if err != nil {
		return nil,
			fmt.Errorf("table %s get route entries error ,err %s", table, err.Error())
	}
	return routes, nil
}

func (r *VPCProvider) listRouteBatch(table, nextToken string, routes *[]*model.Route) error {
	routeEntryListRequest := vpc.CreateDescribeRouteEntryListRequest()
	routeEntryListRequest.NextHopType = model.RouteNextHopTypeInstance
	routeEntryListRequest.RouteEntryType = model.RouteEntryTypeCustom
	routeEntryListRequest.RouteTableId = table
	routeEntryListRequest.NextToken = nextToken
	routeEntryListRequest.MaxResult = requests.NewInteger(model.RouteMaxQueryRouteEntry)
	routeEntryListResponse, err := r.auth.VPC.DescribeRouteEntryList(routeEntryListRequest)
	if err != nil {
		return fmt.Errorf("describe route entry list error, err %v", err)
	}
	routeEntries := routeEntryListResponse.RouteEntrys.RouteEntry
	if len(routeEntries) <= 0 {
		klog.Warningf("alicloud: table [%s] has 0 route entry.", table)
	}
	for _, e := range routeEntries {

		//skip none custom route
		if e.Type != model.RouteEntryTypeCustom ||
			// ECMP is not supported yet, skip next hop not equals 1
			len(e.NextHops.NextHop) != 1 ||
			// skip none Instance route
			strings.ToLower(e.NextHops.NextHop[0].NextHopType) != "instance" ||
			// skip DNAT route
			e.DestinationCidrBlock == "0.0.0.0/0" {
			continue
		}
		pvid, err := r.providerIDFromInstanceId(e.NextHops.NextHop[0].NextHopId)
		if err != nil {
			return err
		}
		route := &model.Route{
			Name:            fmt.Sprintf("%s-%s", pvid, e.DestinationCidrBlock),
			DestinationCIDR: e.DestinationCidrBlock,
			ProviderId:      pvid,
		}
		*routes = append(*routes, route)
	}
	if routeEntryListResponse.NextToken != "" {
		return r.listRouteBatch(table, routeEntryListResponse.NextToken, routes)
	}
	return nil
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
		req.PageSize = requests.NewInteger(next.PageSize)
		req.PageNumber = requests.NewInteger(next.PageNumber)
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
