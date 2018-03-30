package alicloud

import (
	"fmt"
	"testing"
	"strings"
	//"time"
	"errors"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/patrickmn/go-cache"
)

func NewMockRouteMgr(client RouteSDK) (*ClientMgr, error) {
	mgr := &ClientMgr{
		routes: &RoutesClient{
			client:	client,
			routers: cache.New(defaultCacheExpiration, defaultCacheExpiration),
			vpcs:    cache.New(defaultCacheExpiration, defaultCacheExpiration),
		},
	}
	return mgr, nil
}

func TestListRoutes(t *testing.T) {
	vpcid 		:= "vpc-2zeaybwqmvn6qgabfd3pe"
	vrouterid 	:= "vrt-2zegcm0ty46mq243fmxoj"
	vswitchid 	:= "vsw-2zeclpmxy66zzxj4cg4ls"
	routetableid 	:= "vtb-2zedne8cr43rp5oqsr9xg"

	entries := []ecs.RouteEntrySetType{
		{
			RouteTableId: routetableid,
			DestinationCidrBlock: "172.16.3.0/24",
			Type: ecs.RouteTableCustom,
			NextHopType: "Instance",
			InstanceId: "i-2zee0h6bdcgrocv2n9jb",
			Status: ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId: routetableid,
			DestinationCidrBlock: "172.16.2.0/24",
			Type: ecs.RouteTableCustom,
			NextHopType: "Instance",
			InstanceId: "i-2zecarjjmtkx3oru4233",
			Status: ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId: routetableid,
			DestinationCidrBlock: "172.16.0.0/24",
			Type: ecs.RouteTableCustom,
			NextHopType: "Instance",
			InstanceId: "i-2ze7q4vl8cosjsd56j0h",
			Status: ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId: routetableid,
			DestinationCidrBlock: "0.0.0.0/0",
			Type: ecs.RouteTableCustom,
			NextHopType: "NatGateway",
			InstanceId: "ngw-2zetlvdtq0zt9ubez3zz3",
			Status: ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId: routetableid,
			DestinationCidrBlock: "192.168.0.0/16",
			Type: ecs.RouteTableSystem,
			NextHopType: "local",
			Status: ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId: routetableid,
			DestinationCidrBlock: "100.64.0.0/10",
			Type: ecs.RouteTableSystem,
			NextHopType: "service",
			Status: ecs.RouteEntryStatusAvailable,
		},
	}
	cmgr, err := NewMockRouteMgr(&mockRouteSDK{
		describeVpcs: func(args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error){
			if args.RegionId != common.Beijing{
				return nil,nil, errors.New("pls specify the right region")
			}
			vpcs = []ecs.VpcSetType{
				{
					VpcId: 	vpcid,
					RegionId: common.Beijing,
					VSwitchIds:struct {
						VSwitchId []string
					}{
						VSwitchId: []string{vswitchid},
					},
					CidrBlock: "192.168.0.0/16",
					VRouterId: vrouterid,
				},
			}
			return vpcs, nil, nil
		},
		describeVRouters: func(args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error){
			if args.RegionId != common.Beijing{
				return nil,nil, errors.New("pls specify the right region")
			}
			vrouters = []ecs.VRouterSetType{
				{
					VRouterId: vrouterid,
					RegionId: common.Beijing,
					VpcId: vpcid,
					RouteTableIds: struct {
						RouteTableId []string
					}{
						RouteTableId: []string{routetableid},
					},
				},
			}
			return vrouters,nil,nil
		},
		describeRouteTables: func(args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error){
			if args.RouteTableId != routetableid{
				return nil,nil, errors.New("pls specify the right route table id")
			}
			routeTables = []ecs.RouteTableSetType {
				{
					VRouterId: vrouterid,
					RouteTableId: routetableid,
					RouteEntrys: struct {
						RouteEntry []ecs.RouteEntrySetType
					}{
						RouteEntry: entries,
					},
					RouteTableType: "System",
				},
			}
			return routeTables,nil, nil
		},
	})
	if err != nil {
		t.Fatal("failed to create client manager")
	}

	route, err := cmgr.Routes().ListRoutes(common.Beijing, []string{""})
	if err != nil {
		t.Fatal("failed to list routes")
	}
	for _, r := range route {
		found := false
		t.Log(PrettyJson(r))
		for _, entry := range entries {
			if entry.DestinationCidrBlock == r.DestinationCIDR &&
			string(r.TargetNode) == strings.Join([]string{"cn-beijing",entry.InstanceId}, ".") {
				if entry.Type != "Custom" {
					t.Fatal("should not return none cutomized routes")
				}
				found = true
			}
		}
		if !found {
			t.Fatal("route was not matched.")
		}
	}

}

type mockRouteSDK struct {
	describeVpcs func(args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error)
	describeVRouters func(args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error)
	describeRouteTables func(args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error)
	deleteRouteEntry func(args *ecs.DeleteRouteEntryArgs) error
	createRouteEntry func(args *ecs.CreateRouteEntryArgs) error
	waitForAllRouteEntriesAvailable func(vrouterId string, routeTableId string, timeout int) error
}

func (m *mockRouteSDK) DescribeVpcs(args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error){
	if m.describeVpcs != nil {
		return m.describeVpcs(args)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *mockRouteSDK) DescribeVRouters(args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error){
	if m.describeVRouters != nil {
		return m.describeVRouters(args)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *mockRouteSDK) DescribeRouteTables(args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error){
	if m.describeRouteTables != nil {
		return m.describeRouteTables(args)
	}
	return nil, nil, errors.New("not implemented")
}
func (m *mockRouteSDK) DeleteRouteEntry(args *ecs.DeleteRouteEntryArgs) error{
	if m.deleteRouteEntry != nil {
		return m.deleteRouteEntry(args)
	}
	return errors.New("not implemented")
}
func (m *mockRouteSDK) CreateRouteEntry(args *ecs.CreateRouteEntryArgs) error{
	if m.createRouteEntry != nil {
		return m.createRouteEntry(args)
	}
	return errors.New("not implemented")
}
func (m *mockRouteSDK) WaitForAllRouteEntriesAvailable(vrouterId string, routeTableId string, timeout int) error{
	if m.waitForAllRouteEntriesAvailable != nil {
		return m.waitForAllRouteEntriesAvailable(vrouterId, routeTableId, timeout)
	}
	return errors.New("not implemented")
}


func testCamel(t *testing.T, original, expected string) {
	converted := replaceCamel(original)
	if converted != expected {
		t.Errorf("failed to replace camel from %s to %s: %s", original, expected, converted)
	}
}

func TestSep(t *testing.T) {

	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-ProtocolPort", ServiceAnnotationLoadBalancerProtocolPort)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-AddressType", ServiceAnnotationLoadBalancerAddressType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-SLBNetworkType", ServiceAnnotationLoadBalancerSLBNetworkType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-ChargeType", ServiceAnnotationLoadBalancerChargeType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-Region", ServiceAnnotationLoadBalancerRegion)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-Bandwidth", ServiceAnnotationLoadBalancerBandwidth)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-CertID", ServiceAnnotationLoadBalancerCertID)

	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckFlag", ServiceAnnotationLoadBalancerHealthCheckFlag)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckType", ServiceAnnotationLoadBalancerHealthCheckType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckURI", ServiceAnnotationLoadBalancerHealthCheckURI)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckConnectPort", ServiceAnnotationLoadBalancerHealthCheckConnectPort)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthyThreshold", ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-UnhealthyThreshold", ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckInterval", ServiceAnnotationLoadBalancerHealthCheckInterval)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckConnectTimeout", ServiceAnnotationLoadBalancerHealthCheckConnectTimeout)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckTimeout", ServiceAnnotationLoadBalancerHealthCheckTimeout)

	testCamel(t, ServiceAnnotationLoadBalancerProtocolPort, ServiceAnnotationLoadBalancerProtocolPort)
	testCamel(t, ServiceAnnotationLoadBalancerAddressType, ServiceAnnotationLoadBalancerAddressType)
	testCamel(t, ServiceAnnotationLoadBalancerSLBNetworkType, ServiceAnnotationLoadBalancerSLBNetworkType)
	testCamel(t, ServiceAnnotationLoadBalancerChargeType, ServiceAnnotationLoadBalancerChargeType)
	testCamel(t, ServiceAnnotationLoadBalancerRegion, ServiceAnnotationLoadBalancerRegion)
	testCamel(t, ServiceAnnotationLoadBalancerBandwidth, ServiceAnnotationLoadBalancerBandwidth)
	testCamel(t, ServiceAnnotationLoadBalancerCertID, ServiceAnnotationLoadBalancerCertID)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckFlag, ServiceAnnotationLoadBalancerHealthCheckFlag)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckType, ServiceAnnotationLoadBalancerHealthCheckType)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckURI, ServiceAnnotationLoadBalancerHealthCheckURI)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckConnectPort, ServiceAnnotationLoadBalancerHealthCheckConnectPort)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold, ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold, ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckInterval, ServiceAnnotationLoadBalancerHealthCheckInterval)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckConnectTimeout, ServiceAnnotationLoadBalancerHealthCheckConnectTimeout)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckTimeout, ServiceAnnotationLoadBalancerHealthCheckTimeout)

	// Ignore the unsupported annotation
	testCamel(t, "alicloud-loadbalancer-HealthCheckTimeout", "alicloud-loadbalancer-HealthCheckTimeout")
}

func TestRealClient(t *testing.T) {
	realRouteClient(keyid,keysecret)
}

func realRouteClient(keyid,keysec string) {
	if keyid == "" || keysec == "" {
		return
	}
	cs := ecs.NewClient(keyid, keysec)

	vpc,_,_ := cs.DescribeRouteTables(&ecs.DescribeRouteTablesArgs{
		RouteTableId:   "vtb-2zedne8cr43rp5oqsr9xg",
		VRouterId: 	"vrt-2zegcm0ty46mq243fmxoj",
	})

	fmt.Printf(PrettyJson(vpc))
}