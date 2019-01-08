/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"k8s.io/kubernetes/pkg/cloudprovider"

	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/golang/glog"
	"github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"sync"
	"time"
)

type RoutesClient struct {
	client  RouteSDK
	lock    *sync.RWMutex
	routers *cache.Cache
	vpcs    *cache.Cache
}

type RouteSDK interface {
	DescribeVpcs(args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error)
	DescribeVRouters(args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error)
	DescribeRouteTables(args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error)
	DeleteRouteEntry(args *ecs.DeleteRouteEntryArgs) error
	CreateRouteEntry(args *ecs.CreateRouteEntryArgs) error
	WaitForAllRouteEntriesAvailable(vrouterId string, routeTableId string, timeout int) error
}

var defaultCacheExpiration = time.Duration(10 * time.Minute)

func (r *RoutesClient) findRouter(region common.Region, vpcid string) (*ecs.VRouterSetType, error) {
	// look up for vpc
	vpckey := fmt.Sprintf("%s.%s", region, vpcid)
	v, ok := r.vpcs.Get(vpckey)
	if !ok {
		vpc, _, err := r.client.DescribeVpcs(&ecs.DescribeVpcsArgs{
			RegionId: region,
			VpcId:    vpcid,
		})
		if err != nil {
			glog.Errorf("alicloud: error from ecs.DescribeVpcs(%s,%s,%v). message=%s\n", vpcid, region, vpc, r.getErrorString(err))
			return nil, err
		}
		if len(vpc) <= 0 {
			return nil, fmt.Errorf("can not find vpc metadata by "+
				"instance. region=%s, vpcid=%s, error=[no vpc found by specified id (%v)]", region, vpcid, vpc)
		}
		r.vpcs.Set(vpckey, &vpc[0], defaultCacheExpiration)
		v = &vpc[0]
	}
	vpc := v.(*ecs.VpcSetType)
	// lookup for VRouter
	routerkey := fmt.Sprintf("%s.%s", region, vpc.VRouterId)
	route, ok := r.routers.Get(routerkey)
	if !ok {
		vroute, _, err := r.client.DescribeVRouters(
			&ecs.DescribeVRoutersArgs{
				VRouterId: vpc.VRouterId,
				RegionId:  region,
			})
		if err != nil {
			glog.Errorf("alicloud: error ecs.DescribeVRouters(%s,%s). "+
				"message=%s\n", vpc.VRouterId, region, r.getErrorString(err))
			return nil, err
		}
		if len(vroute) <= 0 {
			return nil, fmt.Errorf("can not find VRouter metadata "+
				"by instance. region=%s,vpcid=%v, error=[no VRouter found by specified id]", region, vpc)
		}
		r.routers.Set(routerkey, &vroute[0], defaultCacheExpiration)
		route = &vroute[0]
	}
	rts := route.(*ecs.VRouterSetType)
	return rts, nil
}

func (r *RoutesClient) findRouteTables(region common.Region, vpcid string) ([]ecs.RouteTableSetType, *common.PaginationResult, error) {

	route, err := r.findRouter(region, vpcid)
	if err != nil {
		return nil, nil, err
	}
	return r.client.DescribeRouteTables(
		&ecs.DescribeRouteTablesArgs{
			VRouterId:    route.VRouterId,
			RouteTableId: route.RouteTableIds.RouteTableId[0],
		})
}

// ListRoutes lists all managed routes that belong to the specified clusterName
func (r *RoutesClient) ListRoutes(region common.Region, vpcs []string) ([]*cloudprovider.Route, error) {

	routes := []*cloudprovider.Route{}
	for _, vpc := range vpcs {
		tables, _, err := r.findRouteTables(region, vpc)
		if err != nil {
			glog.Errorf("alicloud: error ecs.DescribeRouteTables() message=%s.\n", r.getErrorString(err))
			return nil, err
		}

		if len(tables) <= 0 {
			glog.Warningf("alicloud: error SDKClientRoutes.SDKClientRoutes. vpc=%s, message=[no route table found]\n", vpc)
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
func (r *RoutesClient) CreateRoute(route *cloudprovider.Route, region common.Region, vpcid string) error {

	tables, _, err := r.findRouteTables(region, vpcid)
	if err != nil {
		glog.Errorf("alicloud: error ecs.DescribeRouteTables(). message=%s.\n", err.Error())
		return err
	}
	if len(tables) <= 0 {
		glog.Errorf("alicloud: returned zero items. abort creating. \n")
		return fmt.Errorf("error: cant find route table for route. cidr=%s", route.DestinationCIDR)
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
func (r *RoutesClient) DeleteRoute(route *cloudprovider.Route, region common.Region, vpcid string) error {

	tabels, _, err := r.findRouteTables(region, vpcid)
	if err != nil {
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

func (r *RoutesClient) reCreateRoute(table ecs.RouteTableSetType, route *ecs.CreateRouteEntryArgs) error {

	exist := false
	for _, e := range table.RouteEntrys.RouteEntry {
		if e.RouteTableId == route.RouteTableId &&
			e.Type == ecs.RouteTableCustom &&
			e.InstanceId == route.NextHopId {

			if e.DestinationCidrBlock == route.DestinationCidrBlock &&
				e.Status == ecs.RouteEntryStatusAvailable {
				exist = true
				glog.V(2).Infof("keep target entry: rtableid=%s, "+
					"CIDR=%s, NextHop=%s \n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
				continue
			}

			// 0.0.0.0/0 => ECS1 this kind of route is used for DNAT. so we keep it
			if e.DestinationCidrBlock == "0.0.0.0/0" {
				glog.V(2).Infof("keep route entry: rtableid=%s, CIDR=%s, "+
					"NextHop=%s For DNAT\n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
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

			glog.V(2).Infof("remove old route entry: rtableid=%s, "+
				"CIDR=%s, NextHop=%s \n", e.RouteTableId, e.DestinationCidrBlock, e.InstanceId)
			continue
		}
	}
	if !exist {
		glog.V(2).Infof("create route entry: rtableid=%s, "+
			"CIDR=%s, NextHop=%s \n", route.RouteTableId, route.DestinationCidrBlock, route.NextHopId)
		return r.client.CreateRouteEntry(route)
	}
	glog.V(2).Infof("route keeped unchanged: rtableid=%s, "+
		"CIDR=%s, NextHop=%s \n", route.RouteTableId, route.DestinationCidrBlock, route.NextHopId)
	return nil
}

func (r *RoutesClient) getErrorString(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}
