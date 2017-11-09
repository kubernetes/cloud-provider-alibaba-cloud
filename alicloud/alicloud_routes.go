package alicloud

import (
	"k8s.io/kubernetes/pkg/cloudprovider"

	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"sync"
	"github.com/patrickmn/go-cache"
	"time"
)

type SDKClientRoutes struct {
	client  *ecs.Client
	lock    *sync.RWMutex
	routers *cache.Cache
	vpcs    *cache.Cache

}

var defaultCacheExpiration = time.Duration(10 * time.Minute)

func NewSDKClientRoutes(access_key_id string, access_key_secret string) (*SDKClientRoutes, error) {
	c := ecs.NewClient(access_key_id, access_key_secret)
	c.SetUserAgent(KUBERNETES_ALICLOUD_IDENTITY)

	return &SDKClientRoutes{
		client: c,
		routers: cache.New(defaultCacheExpiration, defaultCacheExpiration),
		vpcs:    cache.New(defaultCacheExpiration, defaultCacheExpiration),
	}, nil
}

func (r *SDKClientRoutes) findRouter(region common.Region, vpcid string) (*ecs.VRouterSetType,error) {
	// look up for vpc
	vpckey := fmt.Sprintf("%s.%s",region,vpcid)
	v,ok := r.vpcs.Get(vpckey);
	if !ok {
		vpc, _, err := r.client.DescribeVpcs(&ecs.DescribeVpcsArgs{
			RegionId: region,
			VpcId:    vpcid,
		})
		if err != nil {
			glog.V(2).Infof("Alicloud: Error ecs.DescribeVpcs(%s,%s), %s . \n", region, vpc, r.getErrorString(err))
			return nil, err
		}
		if len(vpc) <= 0 {
			return nil, errors.New(fmt.Sprintf("Can not find vpc Meta by instance[region=%s][vpcid=%s], error: len = 0", region, vpc))
		}
		r.vpcs.Set(vpckey, &vpc[0], defaultCacheExpiration)
		v = &vpc[0]
		glog.V(2).Infof("call api: DescribeVpcs, %+v\n",vpc[0])
	}
	vpc := v.(*ecs.VpcSetType)
	// lookup for VRouter
	routerkey := fmt.Sprintf("%s.%s",region,vpc.VRouterId)
	route, ok := r.routers.Get(routerkey)
	if !ok {
		vroute, _, err := r.client.DescribeVRouters(
			&ecs.DescribeVRoutersArgs{
				VRouterId: vpc.VRouterId,
				RegionId:  region,
			})
		if err != nil {
			glog.V(2).Infof("Alicloud: Error ecs.DescribeVRouters(%s,%s), %s .\n", vpc.VRouterId, region, r.getErrorString(err))
			return nil, err
		}
		if len(vroute) <= 0 {
			return nil, errors.New(fmt.Sprintf("Can not find VRouter Meta by instance[region=%s][vpcid=%s], error: len = 0", region, vpc))
		}
		r.routers.Set(routerkey, &vroute[0],defaultCacheExpiration)
		route = &vroute[0]

		glog.V(2).Infof("call api: DescribeVRouters, %+v\n",vroute[0])
	}
	rts := route.(*ecs.VRouterSetType)
	return rts,nil
}

func (r *SDKClientRoutes) findRouteTables(region common.Region, vpcid string) ([]ecs.RouteTableSetType, *common.PaginationResult, error) {
	route, err := r.findRouter(region,vpcid)
	if err != nil {
		return nil,nil,err
	}
	return r.client.DescribeRouteTables(
		&ecs.DescribeRouteTablesArgs{
			VRouterId:    route.VRouterId,
			RouteTableId: route.RouteTableIds.RouteTableId[0],
	})
}

// ListRoutes lists all managed routes that belong to the specified clusterName
func (r *SDKClientRoutes) ListRoutes(region common.Region, vpcs []string) ([]*cloudprovider.Route, error) {

	routes := []*cloudprovider.Route{}
	for _, vpc := range vpcs {
		tables, _ ,err := r.findRouteTables(region, vpc)
		if err != nil {
			glog.V(2).Infof("Alicloud: Error ecs.DescribeRouteTables() %s.\n", r.getErrorString(err))
			return nil, err
		}

		if len(tables) <= 0 {
			glog.V(2).Infof("Alicloud WARINING: SDKClientRoutes.SDKClientRoutes,vpc=%s\n", vpc)
			continue
		}

		for _, e := range tables[0].RouteEntrys.RouteEntry {
			//skip none custom route
			if e.Type != ecs.RouteTableCustom {
				continue
			}
			// skip none Instance route
			if strings.ToLower(e.NextHopType) != "instance" {
				continue
			}
			// skip DNAT route
			if e.DestinationCidrBlock == "0.0.0.0/0" {
				continue
			}
			route := &cloudprovider.Route{
				Name:            fmt.Sprintf("%s.%s", region, e.InstanceId),
				TargetNode:      types.NodeName(fmt.Sprintf("%s.%s", region, e.InstanceId)),
				DestinationCIDR: e.DestinationCidrBlock,
			}
			routes = append(routes, route)
		}
	}

	return routes, nil
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
func (r *SDKClientRoutes) CreateRoute(route *cloudprovider.Route, region common.Region, vpcid string) error {

	tables, _, err := r.findRouteTables(region, vpcid)
	if err != nil {
		glog.V(2).Infof("Alicloud: Error ecs.DescribeRouteTables() %s.\n", err.Error())
		return err
	}
	if len(tables) <= 0 {
		glog.V(2).Infof("Alicloud WARINING: SDKClientRoutes.SDKClientRoutes,vRouteID=%s,rTableID=%s\n", tables[0].VRouterId, tables[0].RouteTableId)
		return errors.New(fmt.Sprintf("SDKClientRoutes.CreateRoute(%s),Error: cant find route table [vRouteID=%s,rTableID=%s]", route.DestinationCIDR, tables[0].VRouterId, tables[0].RouteTableId))
	}
	if err := r.reCreateRoute(tables[0], &ecs.CreateRouteEntryArgs{
		DestinationCidrBlock: route.DestinationCIDR,
		NextHopType:          ecs.NextHopIntance,
		NextHopId:            string(route.TargetNode),
		ClientToken:          "",
		RouteTableId:         tables[0].RouteTableId,
	}); err != nil {
		return err
	}

	return r.client.WaitForAllRouteEntriesAvailable(tables[0].VRouterId, tables[0].RouteTableId, 60)
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (r *SDKClientRoutes) DeleteRoute(route *cloudprovider.Route, region common.Region, vpcid string) error {

	tabels, _, err := r.findRouteTables(region,vpcid)
	if err!= nil {
		return err
	}
	if err := r.client.DeleteRouteEntry(&ecs.DeleteRouteEntryArgs{
		RouteTableId:         tabels[0].RouteTableId,
		DestinationCidrBlock: route.DestinationCIDR,
		NextHopId:            string(route.TargetNode),
	}); err != nil {
		return err
	}
	return r.client.WaitForAllRouteEntriesAvailable(tabels[0].VRouterId, tabels[0].RouteTableId, 60)
}

func (r *SDKClientRoutes) reCreateRoute(table ecs.RouteTableSetType, route *ecs.CreateRouteEntryArgs) error {

	exist := false
	for _, e := range table.RouteEntrys.RouteEntry {
		if e.RouteTableId == route.RouteTableId &&
			e.Type == ecs.RouteTableCustom &&
			e.InstanceId == route.NextHopId {

			if e.DestinationCidrBlock == route.DestinationCidrBlock &&
				e.Status == ecs.RouteEntryStatusAvailable {
				exist = true
				glog.V(2).Infof("Keep target entry: rtableid=%s, CIDR=%s, NextHop=%s \n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
				continue
			}

			// 0.0.0.0/0 => ECS1 this kind of route is used for DNAT. so we keep it
			if e.DestinationCidrBlock == "0.0.0.0/0" {
				glog.V(2).Infof("Keep route entry: rtableid=%s, CIDR=%s, NextHop=%s For DNAT\n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
				continue
			}
			// Fix: here we delete all the route which targeted to us(instance) except the specified route.
			// That means only one CIDR was allowed to target to the instance. Think if We need to change this
			// to adapt to multi CIDR and deal with unavailable route entry.
			if err := r.client.DeleteRouteEntry(&ecs.DeleteRouteEntryArgs{
				RouteTableId:         route.RouteTableId,
				DestinationCidrBlock: e.DestinationCidrBlock,
				NextHopId:            route.NextHopId,
			}); err != nil {
				return err
			}

			glog.V(2).Infof("Remove old route entry: rtableid=%s, CIDR=%s, NextHop=%s \n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
			continue
		}

		glog.V(2).Infof("Keep route entry: rtableid=%s, CIDR=%s, NextHop=%s \n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
	}
	if !exist {
		return r.client.CreateRouteEntry(route)
	}
	return nil
}

func (r *SDKClientRoutes) getErrorString(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}
