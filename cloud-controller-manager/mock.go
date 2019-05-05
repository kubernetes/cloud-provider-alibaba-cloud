package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/apimachinery/pkg/types"
)

type InitialSet func()

func PreSetCloudData(sets ...InitialSet) {
	for _, initialSet := range sets {
		initialSet()
	}
}

var (
	INSTANCEID = "i-xlakjbidlslkcdxxxx"
)

var (
	LOADBALANCER_ID           = "lb-bp1ids9hmq5924m6uk5w1"
	LOADBALANCER_NAME         = "a2cb99d47cc8311e899db00163e12560"
	LOADBALANCER_ADDRESS      = "47.97.241.114"
	LOADBALANCER_NETWORK_TYPE = "classic"
	LOADBALANCER_SPEC         = slb.LoadBalancerSpecType(slb.S1Small)

	SERVICE_UID = types.UID("2cb99d47-cc83-11e8-99db-00163e125603")
)

var (
	VPCID          = "vpc-2zeaybwqmvn6qgabfd3pe"
	VROUTER_ID     = "vrt-2zegcm0ty46mq243fmxoj"
	ROUTE_TABLE_ID = "vtb-2zedne8cr43rp5oqsr9xg"
	REGION         = common.Hangzhou
	REGION_A       = "cn-hangzhou-a"
	VSWITCH_ID     = "vsw-2zeclpmxy66zzxj4cg4ls"
	ROUTE_ENTRIES  = []ecs.RouteEntrySetType{
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "172.16.3.0/24",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "Instance",
			InstanceId:           "i-2zee0h6bdcgrocv2n9jb",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "172.16.2.0/24",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "Instance",
			InstanceId:           "i-2zecarjjmtkx3oru4233",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "172.16.0.0/24",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "Instance",
			InstanceId:           "i-2ze7q4vl8cosjsd56j0h",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "0.0.0.0/0",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "NatGateway",
			InstanceId:           "ngw-2zetlvdtq0zt9ubez3zz3",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "192.168.0.0/16",
			Type:                 ecs.RouteTableSystem,
			NextHopType:          "local",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "100.64.0.0/10",
			Type:                 ecs.RouteTableSystem,
			NextHopType:          "service",
			Status:               ecs.RouteEntryStatusAvailable,
		},
	}
)
