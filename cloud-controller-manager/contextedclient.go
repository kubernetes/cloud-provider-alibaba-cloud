package alicloud

import (
	"context"
	"github.com/denverdino/aliyungo/cen"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/pvtz"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/client-go/kubernetes"
)

type BaseClient struct {
	core kubernetes.Interface
}

func (b *BaseClient) SetCoreClient(core kubernetes.Interface) { b.core = core }

func NewContextedClientSLB(key, secret, region string) *ContextedClientSLB {
	return &ContextedClientSLB{
		BaseClient: BaseClient{},
		slb:        slb.NewSLBClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region)),
	}
}

type ContextedClientSLB struct {
	BaseClient
	// base slb client
	slb *slb.Client
}

func (c *ContextedClientSLB) DescribeLoadBalancers(
	ctx context.Context,
	args *slb.DescribeLoadBalancersArgs,
) (loadBalancers []slb.LoadBalancerType, err error) {
	return c.slb.DescribeLoadBalancers(args)
}

func (c *ContextedClientSLB) DescribeLoadBalancerAttribute(
	ctx context.Context,
	loadBalancerId string,
) (loadBalancer *slb.LoadBalancerType, err error) {
	return c.slb.DescribeLoadBalancerAttribute(loadBalancerId)
}

func (c *ContextedClientSLB) DescribeLoadBalancerHTTPSListenerAttribute(
	ctx context.Context,
	loadBalancerId string,
	port int,
) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error) {
	return c.slb.DescribeLoadBalancerHTTPSListenerAttribute(loadBalancerId, port)
}

func (c *ContextedClientSLB) DescribeLoadBalancerTCPListenerAttribute(
	ctx context.Context,
	loadBalancerId string,
	port int,
) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {
	return c.slb.DescribeLoadBalancerTCPListenerAttribute(loadBalancerId, port)
}

func (c *ContextedClientSLB) DescribeLoadBalancerUDPListenerAttribute(
	ctx context.Context,
	loadBalancerId string,
	port int,
) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error) {
	return c.slb.DescribeLoadBalancerUDPListenerAttribute(loadBalancerId, port)
}

func (c *ContextedClientSLB) DescribeLoadBalancerHTTPListenerAttribute(
	ctx context.Context,
	loadBalancerId string,
	port int,
) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error) {
	return c.slb.DescribeLoadBalancerHTTPListenerAttribute(loadBalancerId, port)
}

func (c *ContextedClientSLB) DescribeTags(ctx context.Context, args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error) {
	return c.slb.DescribeTags(args)
}

func (c *ContextedClientSLB) DescribeVServerGroups(
	ctx context.Context,
	args *slb.DescribeVServerGroupsArgs,
) (response *slb.DescribeVServerGroupsResponse, err error) {
	return c.slb.DescribeVServerGroups(args)
}

func (c *ContextedClientSLB) DescribeVServerGroupAttribute(
	ctx context.Context,
	args *slb.DescribeVServerGroupAttributeArgs,
) (response *slb.DescribeVServerGroupAttributeResponse, err error) {
	return c.slb.DescribeVServerGroupAttribute(args)
}

func (c *ContextedClientSLB) CreateLoadBalancer(
	ctx context.Context,
	args *slb.CreateLoadBalancerArgs,
) (response *slb.CreateLoadBalancerResponse, err error) {
	return c.slb.CreateLoadBalancer(args)
}

func (c *ContextedClientSLB) SetLoadBalancerModificationProtection(
	ctx context.Context,
	args *slb.SetLoadBalancerModificationProtectionArgs,
) (err error) {
	return c.slb.SetLoadBalancerModificationProtection(args)
}

func (c *ContextedClientSLB) DeleteLoadBalancer(ctx context.Context, loadBalancerId string) (err error) {

	return c.slb.DeleteLoadBalancer(loadBalancerId)
}

func (c *ContextedClientSLB) SetLoadBalancerDeleteProtection(ctx context.Context, args *slb.SetLoadBalancerDeleteProtectionArgs) (err error) {
	return c.slb.SetLoadBalancerDeleteProtection(args)
}

func (c *ContextedClientSLB) SetLoadBalancerName(ctx context.Context, loadBalancerId string, loadBalancerName string) (err error) {

	return c.slb.SetLoadBalancerName(loadBalancerId, loadBalancerName)
}

func (c *ContextedClientSLB) ModifyLoadBalancerInstanceSpec(
	ctx context.Context,
	args *slb.ModifyLoadBalancerInstanceSpecArgs,
) (err error) {

	return c.slb.ModifyLoadBalancerInstanceSpec(args)
}
func (c *ContextedClientSLB) ModifyLoadBalancerInternetSpec(
	ctx context.Context,
	args *slb.ModifyLoadBalancerInternetSpecArgs,
) (err error) {

	return c.slb.ModifyLoadBalancerInternetSpec(args)
}

func (c *ContextedClientSLB) RemoveBackendServers(
	ctx context.Context,
	loadBalancerId string,
	backendServers []slb.BackendServerType,
) (result []slb.BackendServerType, err error) {

	return c.slb.RemoveBackendServers(loadBalancerId, backendServers)
}

func (c *ContextedClientSLB) AddBackendServers(ctx context.Context, loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error) {

	return c.slb.AddBackendServers(loadBalancerId, backendServers)
}

func (c *ContextedClientSLB) StopLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) (err error) {
	return c.slb.StopLoadBalancerListener(loadBalancerId, port)
}

func (c *ContextedClientSLB) StartLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) (err error) {
	return c.slb.StartLoadBalancerListener(loadBalancerId, port)
}

func (c *ContextedClientSLB) CreateLoadBalancerTCPListener(
	ctx context.Context,
	args *slb.CreateLoadBalancerTCPListenerArgs,
) (err error) {
	return c.slb.CreateLoadBalancerTCPListener(args)
}

func (c *ContextedClientSLB) CreateLoadBalancerUDPListener(
	ctx context.Context,
	args *slb.CreateLoadBalancerUDPListenerArgs,
) (err error) {
	return c.slb.CreateLoadBalancerUDPListener(args)
}

func (c *ContextedClientSLB) DeleteLoadBalancerListener(ctx context.Context, loadBalancerId string, port int) (err error) {

	return c.slb.DeleteLoadBalancerListener(loadBalancerId, port)
}

func (c *ContextedClientSLB) CreateLoadBalancerHTTPSListener(
	ctx context.Context,
	args *slb.CreateLoadBalancerHTTPSListenerArgs,
) (err error) {
	return c.slb.CreateLoadBalancerHTTPSListener(args)
}

func (c *ContextedClientSLB) CreateLoadBalancerHTTPListener(
	ctx context.Context,
	args *slb.CreateLoadBalancerHTTPListenerArgs,
) (err error) {
	return c.slb.CreateLoadBalancerHTTPListener(args)
}

func (c *ContextedClientSLB) SetLoadBalancerHTTPListenerAttribute(
	ctx context.Context,
	args *slb.SetLoadBalancerHTTPListenerAttributeArgs,
) (err error) {
	return c.slb.SetLoadBalancerHTTPListenerAttribute(args)
}

func (c *ContextedClientSLB) SetLoadBalancerHTTPSListenerAttribute(
	ctx context.Context,
	args *slb.SetLoadBalancerHTTPSListenerAttributeArgs,
) (err error) {
	return c.slb.SetLoadBalancerHTTPSListenerAttribute(args)
}

func (c *ContextedClientSLB) SetLoadBalancerTCPListenerAttribute(
	ctx context.Context,
	args *slb.SetLoadBalancerTCPListenerAttributeArgs,
) (err error) {
	return c.slb.SetLoadBalancerTCPListenerAttribute(args)
}

func (c *ContextedClientSLB) SetLoadBalancerUDPListenerAttribute(
	ctx context.Context,
	args *slb.SetLoadBalancerUDPListenerAttributeArgs,
) (err error) {
	return c.slb.SetLoadBalancerUDPListenerAttribute(args)
}

func (c *ContextedClientSLB) RemoveTags(ctx context.Context, args *slb.RemoveTagsArgs) error {
	return c.slb.RemoveTags(args)
}

func (c *ContextedClientSLB) AddTags(ctx context.Context, args *slb.AddTagsArgs) error {
	return c.slb.AddTags(args)
}

func (c *ContextedClientSLB) CreateVServerGroup(
	ctx context.Context,
	args *slb.CreateVServerGroupArgs,
) (response *slb.CreateVServerGroupResponse, err error) {
	return c.slb.CreateVServerGroup(args)
}

func (c *ContextedClientSLB) DeleteVServerGroup(
	ctx context.Context,
	args *slb.DeleteVServerGroupArgs,
) (response *slb.DeleteVServerGroupResponse, err error) {
	return c.slb.DeleteVServerGroup(args)
}

func (c *ContextedClientSLB) SetVServerGroupAttribute(
	ctx context.Context,
	args *slb.SetVServerGroupAttributeArgs,
) (response *slb.SetVServerGroupAttributeResponse, err error) {
	return c.slb.SetVServerGroupAttribute(args)
}

func (c *ContextedClientSLB) ModifyVServerGroupBackendServers(
	ctx context.Context,
	args *slb.ModifyVServerGroupBackendServersArgs,
) (response *slb.ModifyVServerGroupBackendServersResponse, err error) {
	return c.slb.ModifyVServerGroupBackendServers(args)
}

func (c *ContextedClientSLB) AddVServerGroupBackendServers(
	ctx context.Context,
	args *slb.AddVServerGroupBackendServersArgs,
) (response *slb.AddVServerGroupBackendServersResponse, err error) {
	return c.slb.AddVServerGroupBackendServers(args)
}

func (c *ContextedClientSLB) RemoveVServerGroupBackendServers(
	ctx context.Context,
	args *slb.RemoveVServerGroupBackendServersArgs,
) (response *slb.RemoveVServerGroupBackendServersResponse, err error) {
	return c.slb.RemoveVServerGroupBackendServers(args)
}

// =====================================================================================================================

func NewContextedClientINS(key, secret, region string) *ContextedClientINS {
	return &ContextedClientINS{
		BaseClient: BaseClient{},
		ecs:        ecs.NewECSClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region)),
	}
}

type ContextedClientINS struct {
	BaseClient
	// base ecs client
	ecs *ecs.Client
}

func (c *ContextedClientINS) AddTags(ctx context.Context, args *ecs.AddTagsArgs) error {
	return c.ecs.AddTags(args)
}
func (c *ContextedClientINS) DescribeInstances(
	ctx context.Context,
	args *ecs.DescribeInstancesArgs,
) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error) {
	return c.ecs.DescribeInstances(args)
}
func (c *ContextedClientINS) DescribeNetworkInterfaces(
	ctx context.Context,
	args *ecs.DescribeNetworkInterfacesArgs,
) (resp *ecs.DescribeNetworkInterfacesResponse, err error) {
	return c.ecs.DescribeNetworkInterfaces(args)
}

func (c *ContextedClientINS) DescribeEipAddresses(
	ctx context.Context,
	args *ecs.DescribeEipAddressesArgs,
) (eipAddresses []ecs.EipAddressSetType, pagination *common.PaginationResult, err error) {
	return c.ecs.DescribeEipAddresses(args)
}

// =====================================================================================================================
func NewContextedClientPVTZ(key, secret, region string) *ContextedClientPVTZ {
	return &ContextedClientPVTZ{
		BaseClient: BaseClient{},
		// TODO: change to regional client
		pvtz: pvtz.NewPVTZClientWithSecurityToken(key, secret, "", common.Region(region)),
	}
}

type ContextedClientPVTZ struct {
	BaseClient
	// base pvtz client
	pvtz *pvtz.Client
}

func (c *ContextedClientPVTZ) DescribeZones(ctx context.Context, args *pvtz.DescribeZonesArgs) (zones []pvtz.ZoneType, err error) {
	return c.pvtz.DescribeZones(args)
}

func (c *ContextedClientPVTZ) CheckZoneName(ctx context.Context, args *pvtz.CheckZoneNameArgs) (bool, error) {
	return c.pvtz.CheckZoneName(args)
}

func (c *ContextedClientPVTZ) DescribeZoneInfo(ctx context.Context, args *pvtz.DescribeZoneInfoArgs) (response *pvtz.DescribeZoneInfoResponse, err error) {
	return c.pvtz.DescribeZoneInfo(args)
}

func (c *ContextedClientPVTZ) DescribeRegions(ctx context.Context) (regions []pvtz.RegionType, err error) {
	return c.pvtz.DescribeRegions()
}

func (c *ContextedClientPVTZ) DescribeZoneRecords(ctx context.Context, args *pvtz.DescribeZoneRecordsArgs) (records []pvtz.ZoneRecordType, err error) {
	return c.pvtz.DescribeZoneRecords(args)
}

func (c *ContextedClientPVTZ) DescribeZoneRecordsByRR(ctx context.Context, zoneId string, rr string) (records []pvtz.ZoneRecordType, err error) {
	return c.pvtz.DescribeZoneRecordsByRR(zoneId, rr)
}

func (c *ContextedClientPVTZ) AddZone(ctx context.Context, args *pvtz.AddZoneArgs) (response *pvtz.AddZoneResponse, err error) {
	return c.pvtz.AddZone(args)
}
func (c *ContextedClientPVTZ) DeleteZone(ctx context.Context, args *pvtz.DeleteZoneArgs) (err error) {
	return c.pvtz.DeleteZone(args)
}
func (c *ContextedClientPVTZ) UpdateZoneRemark(ctx context.Context, args *pvtz.UpdateZoneRemarkArgs) error {
	return c.pvtz.UpdateZoneRemark(args)
}
func (c *ContextedClientPVTZ) BindZoneVpc(ctx context.Context, args *pvtz.BindZoneVpcArgs) (err error) {
	return c.pvtz.BindZoneVpc(args)
}
func (c *ContextedClientPVTZ) DeleteZoneRecordsByRR(ctx context.Context, zoneId string, rr string) error {
	return c.pvtz.DeleteZoneRecordsByRR(zoneId, rr)
}
func (c *ContextedClientPVTZ) AddZoneRecord(ctx context.Context, args *pvtz.AddZoneRecordArgs) (response *pvtz.AddZoneRecordResponse, err error) {
	return c.pvtz.AddZoneRecord(args)
}
func (c *ContextedClientPVTZ) UpdateZoneRecord(ctx context.Context, args *pvtz.UpdateZoneRecordArgs) (err error) {
	return c.pvtz.UpdateZoneRecord(args)
}
func (c *ContextedClientPVTZ) DeleteZoneRecord(ctx context.Context, args *pvtz.DeleteZoneRecordArgs) (err error) {
	return c.pvtz.DeleteZoneRecord(args)
}
func (c *ContextedClientPVTZ) SetZoneRecordStatus(ctx context.Context, args *pvtz.SetZoneRecordStatusArgs) (err error) {
	return c.pvtz.SetZoneRecordStatus(args)
}

// =====================================================================================================================

func NewContextedClientRoute(key, secret, region, vpcid string) *ContextedClientRoute {

	return &ContextedClientRoute{
		BaseClient: BaseClient{},
		ecs:        ecs.NewVPCClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region)),
		cen:        cen.NewCENClientWithSecurityToken4RegionalDomain(key, secret, "", common.Region(region)),
		vpcid:      vpcid,
	}
}

type ContextedClientRoute struct {
	BaseClient
	// base slb client
	ecs *ecs.Client
	// some api use alibaba-cloud-sdk-go
	cen   *cen.Client
	vpcid string
}

func (c *ContextedClientRoute) DescribeVpcs(ctx context.Context, args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error) {
	return c.ecs.DescribeVpcs(args)
}

func (c *ContextedClientRoute) DescribeVRouters(ctx context.Context, args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error) {
	return c.ecs.DescribeVRouters(args)
}

func (c *ContextedClientRoute) DescribeRouteTables(ctx context.Context, args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error) {
	return c.ecs.DescribeRouteTables(args)
}

func (c *ContextedClientRoute) DescribeRouteEntryList(ctx context.Context, args *ecs.DescribeRouteEntryListArgs) (response *ecs.DescribeRouteEntryListResponse, err error) {
	return c.ecs.DescribeRouteEntryList(args)
}

func (c *ContextedClientRoute) DeleteRouteEntry(ctx context.Context, args *ecs.DeleteRouteEntryArgs) error {
	return c.ecs.DeleteRouteEntry(args)
}

func (c *ContextedClientRoute) CreateRouteEntry(ctx context.Context, args *ecs.CreateRouteEntryArgs) error {
	return c.ecs.CreateRouteEntry(args)
}

func (c *ContextedClientRoute) PublishRouteEntry(ctx context.Context, args *cen.PublishRouteEntriesArgs) error {
	return c.cen.PublishRouteEntries(args)
}

func (c *ContextedClientRoute) WaitForAllRouteEntriesAvailable(ctx context.Context, vrouterId string, routeTableId string, timeout int) error {
	return c.ecs.WaitForAllRouteEntriesAvailable(vrouterId, routeTableId, timeout)
}
