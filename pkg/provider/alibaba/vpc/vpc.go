package vpc

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"net"
	"strings"
)

type AssociatedInstanceType string

const SlbInstance = AssociatedInstanceType("SlbInstance")

func NewVPCProvider(
	auth *base.ClientMgr,
) *VPCProvider {
	return &VPCProvider{auth: auth}
}

var _ prvd.IVPC = &VPCProvider{}

type VPCProvider struct {
	auth   *base.ClientMgr
	region string
}

func (r *VPCProvider) ListRouteTables(ctx context.Context, vpcID string) ([]string, error) {
	tableListRequest := vpc.CreateDescribeRouteTableListRequest()
	tableListRequest.VpcId = vpcID
	resp, err := r.auth.VPC.DescribeRouteTableList(tableListRequest)
	if err != nil {
		return nil, fmt.Errorf("error describe vpc: %v route tables, error: %v", vpcID, err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s, vpcId: %s", resp.RequestId, "DescribeRouteTableList", vpcID)
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
		_, instance, err := util.NodeFromProviderID(provID)
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
	klog.V(5).Infof("RequestId: %s, API: %s, providerId: %s",
		resp.RequestId, "DescribeRouteEntryList", provID)
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
	return util.ProviderIDFromInstance(r.region, instanceID), nil
}

func (r *VPCProvider) CreateRoute(
	ctx context.Context, table string, provideID string, destinationCIDR string,
) (*model.Route, error) {
	createRouteEntryRequest := vpc.CreateCreateRouteEntryRequest()
	createRouteEntryRequest.RouteTableId = table
	createRouteEntryRequest.DestinationCidrBlock = destinationCIDR
	createRouteEntryRequest.NextHopType = model.RouteNextHopTypeInstance
	_, instance, err := util.NodeFromProviderID(provideID)
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
	_, instance, err := util.NodeFromProviderID(provideID)
	if err != nil {
		return fmt.Errorf("invalid provide id: %v, err: %v", provideID, err)
	}
	deleteRouteEntryRequest.NextHopId = instance
	_, err = r.auth.VPC.DeleteRouteEntry(deleteRouteEntryRequest)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidRouteEntry.NotFound") {
			// route already removed
			return nil
		}
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
	klog.V(5).Infof("RequestId: %s, API: %s, tableId: %s",
		routeEntryListResponse.RequestId, "DescribeRouteEntryList", table)
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

func (r *VPCProvider) DescribeEipAddresses(ctx context.Context, instanceType string, instanceId string) (
	[]string, error,
) {
	req := vpc.CreateDescribeEipAddressesRequest()
	req.AssociatedInstanceType = instanceType
	req.AssociatedInstanceId = instanceId
	var ips []string
	next := &util.Pagination{
		PageNumber: 1,
		PageSize:   10,
	}

	for {
		req.PageSize = requests.NewInteger(next.PageSize)
		req.PageNumber = requests.NewInteger(next.PageNumber)
		resp, err := r.auth.VPC.DescribeEipAddresses(req)
		if err != nil {
			return nil, err
		}

		for _, eip := range resp.EipAddresses.EipAddress {
			ips = append(ips, eip.IpAddress)
		}

		pageResult := &util.PaginationResult{
			PageNumber: resp.PageNumber,
			PageSize:   resp.PageSize,
			TotalCount: resp.TotalCount,
		}
		next = pageResult.NextPage()
		if next == nil {
			break
		}
	}
	return ips, nil
}

func (r *VPCProvider) DescribeVpcCIDRBlock(ctx context.Context, vpcId string, ipVersion model.AddressIPVersionType) ([]*net.IPNet, error) {
	req := vpc.CreateDescribeVpcAttributeRequest()
	req.VpcId = vpcId
	resp, err := r.auth.VPC.DescribeVpcAttribute(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("DescribeVpcAttribute resp is nil")
	}

	var cidrs []*net.IPNet
	if ipVersion == model.IPv6 {
		var ipv6CidrBlocks []string
		ipv6CidrBlocks = append(ipv6CidrBlocks, resp.Ipv6CidrBlock)
		if len(resp.Ipv6CidrBlocks.Ipv6CidrBlock) > 0 {
			for _, ipv6 := range resp.Ipv6CidrBlocks.Ipv6CidrBlock {
				ipv6CidrBlocks = append(ipv6CidrBlocks, ipv6.Ipv6CidrBlock)
			}
		}

		for _, ipv6 := range ipv6CidrBlocks {
			_, ipv6CIDR, err := net.ParseCIDR(ipv6)
			if err != nil {
				klog.Warningf("can not parse ipv6 cidr %s, error: %s", ipv6, err.Error())
			} else {
				cidrs = append(cidrs, ipv6CIDR)
			}
		}
	} else {
		var ipv4CidrBlocks []string
		ipv4CidrBlocks = append(ipv4CidrBlocks, resp.CidrBlock)
		if len(resp.SecondaryCidrBlocks.SecondaryCidrBlock) > 0 {
			ipv4CidrBlocks = append(ipv4CidrBlocks, resp.SecondaryCidrBlocks.SecondaryCidrBlock...)
		}
		for _, ipv4 := range ipv4CidrBlocks {
			_, ipv4CIDR, err := net.ParseCIDR(ipv4)
			if err != nil {
				klog.Warningf("can not parse ipv4 cidr %s, error: %s", ipv4, err.Error())
			} else {
				cidrs = append(cidrs, ipv4CIDR)
			}
		}
	}

	return cidrs, err
}

// DescribeVSwitches used for e2etest
func (r *VPCProvider) DescribeVSwitches(ctx context.Context, vpcID string) ([]vpc.VSwitch, error) {
	req := vpc.CreateDescribeVSwitchesRequest()
	req.VpcId = vpcID
	var vSwitches []vpc.VSwitch
	next := &util.Pagination{
		PageNumber: 1,
		PageSize:   10,
	}
	for {
		req.PageSize = requests.NewInteger(next.PageSize)
		req.PageNumber = requests.NewInteger(next.PageNumber)
		resp, err := r.auth.VPC.DescribeVSwitches(req)
		if err != nil {
			return nil, err
		}
		vSwitches = append(vSwitches, resp.VSwitches.VSwitch...)
		pageResult := &util.PaginationResult{
			PageNumber: resp.PageNumber,
			PageSize:   resp.PageSize,
			TotalCount: resp.TotalCount,
		}
		next = pageResult.NextPage()
		if next == nil {
			break
		}
	}
	return vSwitches, nil
}

// CreateRouteTable used for e2etest
func (r *VPCProvider) CreateRouteTable(ctx context.Context, vpcId string, name string) (
	*vpc.CreateRouteTableResponse, error,
) {
	req := vpc.CreateCreateRouteTableRequest()
	req.VpcId = vpcId
	req.RouteTableName = name
	return r.auth.VPC.CreateRouteTable(req)
}

// CreateRouteTable used for e2etest
func (r *VPCProvider) DeleteRouteTable(ctx context.Context, routeTableId string) (
	*vpc.DeleteRouteTableResponse, error,
) {
	req := vpc.CreateDeleteRouteTableRequest()
	req.RouteTableId = routeTableId
	return r.auth.VPC.DeleteRouteTable(req)
}

// DescribeRouteTableList used for e2etest
func (r *VPCProvider) DescribeRouteTableList(ctx context.Context, vpcId string) ([]string, error) {
	req := vpc.CreateDescribeRouteTableListRequest()
	req.VpcId = vpcId
	resp, err := r.auth.VPC.DescribeRouteTableList(req)
	if err != nil {
		return nil, err
	}
	var rtIds []string
	for _, rt := range resp.RouterTableList.RouterTableListType {
		rtIds = append(rtIds, rt.RouteTableId)
	}
	return rtIds, nil
}
