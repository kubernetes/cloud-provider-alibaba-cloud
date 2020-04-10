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
	"k8s.io/cloud-provider"
	"k8s.io/klog"

	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

type vpc struct {
	vpcid     string
	vrouterid string
	tableids  []string
}

//RoutesClient wrap route sdk
type RoutesClient struct {
	region string
	vpc    vpc
	client RouteSDK
}

var index = 1

//RouteSDK define route sdk interface
type RouteSDK interface {
	DescribeVpcs(args *ecs.DescribeVpcsArgs) (vpcs []ecs.VpcSetType, pagination *common.PaginationResult, err error)
	DescribeVRouters(args *ecs.DescribeVRoutersArgs) (vrouters []ecs.VRouterSetType, pagination *common.PaginationResult, err error)
	DescribeRouteTables(args *ecs.DescribeRouteTablesArgs) (routeTables []ecs.RouteTableSetType, pagination *common.PaginationResult, err error)
	DeleteRouteEntry(args *ecs.DeleteRouteEntryArgs) error
	CreateRouteEntry(args *ecs.CreateRouteEntryArgs) error
	WaitForAllRouteEntriesAvailable(vrouterId string, routeTableId string, timeout int) error
	DescribeRouteEntryList(args *ecs.DescribeRouteEntryListArgs) (response *ecs.DescribeRouteEntryListResponse, err error)
}

//WithVPC set vpc id and and route table ids.
func (r *RoutesClient) WithVPC(vpcid string, tableids string) error {
	args := &ecs.DescribeVpcsArgs{
		VpcId:    vpcid,
		RegionId: common.Region(r.region),
	}
	vpcs, _, err := r.client.DescribeVpcs(args)
	if err != nil {
		return fmt.Errorf("withvpc error: %s", err)
	}
	if len(vpcs) != 1 {
		return fmt.Errorf("alicloud: "+
			"multiple vpc found by id[%s], length(vpcs)=%d", vpcid, len(vpcs))
	}
	r.vpc.vrouterid = vpcs[0].VRouterId
	r.vpc.vpcid = vpcid
	if tableids != "" {
		for _, s := range strings.Split(tableids, ",") {
			r.vpc.tableids = append(r.vpc.tableids, strings.TrimSpace(s))
		}
		klog.Infof("using user customized route table ids (%v)", r.vpc.tableids)
	}
	return nil
}

// ListRoutes lists all managed routes that belong to the specified clusterName
func (r *RoutesClient) ListRoutes(tableid string) (routes []*cloudprovider.Route, err error) {

	klog.Infof("ListRoutes: for route table %s", tableid)
	// route will be overwritten by getRouteEntryBatch
	err = r.getRouteEntryBatch(tableid, "", &routes)
	if err != nil {
		return []*cloudprovider.Route{},
			fmt.Errorf("table %s get route entries error ,err %s", tableid, err.Error())
	} else {
		return routes, nil
	}
}

func (r *RoutesClient) getRouteEntryBatch(tableid string, nextToken string, routes *[]*cloudprovider.Route) error {

	args := &ecs.DescribeRouteEntryListArgs{
		RegionId:       r.region,
		RouteTableId:   tableid,
		RouteEntryType: "Custom",
		NextToken:      nextToken,
	}
	response, err := r.client.DescribeRouteEntryList(args)
	if err != nil || response == nil {
		return fmt.Errorf("describe route entry list error, err %v", err)
	}

	routeEntries := response.RouteEntrys.RouteEntry
	if len(routeEntries) <= 0 {
		klog.Warningf("alicloud: table [%s] has 0 route entry.", tableid)
	}

	for _, e := range routeEntries {

		//skip none custom route
		if e.Type != string(ecs.RouteTableCustom) ||
			// ECMP is not supported yet, skip next hop not equals 1
			len(e.NextHops.NextHop) != 1 ||
			// skip none Instance route
			strings.ToLower(e.NextHops.NextHop[0].NextHopType) != "instance" ||
			// skip DNAT route
			e.DestinationCidrBlock == "0.0.0.0/0" {
			continue
		}

		route := &cloudprovider.Route{
			Name:            nodeid(r.region, e.NextHops.NextHop[0].NextHopId),
			DestinationCIDR: e.DestinationCidrBlock,
			TargetNode:      types.NodeName(nodeid(r.region, e.NextHops.NextHop[0].NextHopId)),
		}
		*routes = append(*routes, route)
	}
	// get next batch
	if response.NextToken != "" {
		return r.getRouteEntryBatch(tableid, response.NextToken, routes)
	}
	return nil
}

//RouteTables return all the tables in the vpc network.
func (r *RoutesClient) RouteTables() ([]string, error) {
	if len(r.vpc.tableids) != 0 {
		return r.vpc.tableids, nil
	}
	// describe vpc attribute to get route table ids.
	args := &ecs.DescribeVpcsArgs{
		VpcId:    r.vpc.vpcid,
		RegionId: common.Region(r.region),
	}
	vpcs, _, err := r.client.DescribeVpcs(args)
	if err != nil {
		return []string{}, err
	}
	if len(vpcs) != 1 {
		return []string{}, fmt.Errorf("alicloud: "+
			"multiple vpc found by id[%s], length(vpcs)=%d", r.vpc.vpcid, len(vpcs))
	}
	if len(vpcs[0].RouterTableIds.RouterTableIds) != 1 {
		return []string{}, fmt.Errorf("alicloud: multiple "+
			"route table or no route table found in vpc %s, [%s]", r.vpc.vpcid, vpcs[0].RouterTableIds.RouterTableIds)
	}
	return vpcs[0].RouterTableIds.RouterTableIds, nil
}

// CreateRoute creates the described managed route
// route.Name will be ignored, although the cloud-provider may use nameHint
// to create a more user-meaningful name.
func (r *RoutesClient) CreateRoute(tabid string, route *cloudprovider.Route, region common.Region, vpcid string) error {
	describeRouteEntryListArgs := &ecs.DescribeRouteEntryListArgs{
		RegionId:             r.region,
		RouteTableId:         tabid,
		RouteEntryType:       string(ecs.RouteTableCustom),
		DestinationCidrBlock: route.DestinationCIDR,
		NextHopId:            string(route.TargetNode),
	}
	response, err := r.client.DescribeRouteEntryList(describeRouteEntryListArgs)
	if err != nil || response == nil {
		return fmt.Errorf("describe table %s RouteEntry list error, %v", tabid, err)
	}

	if len(response.RouteEntrys.RouteEntry) > 0 {
		klog.Infof("CreateRoute: skip exist route, %s -> %s", route.DestinationCIDR, route.TargetNode)
		return nil
	}

	args := &ecs.CreateRouteEntryArgs{
		ClientToken:          "",
		RouteTableId:         tabid,
		DestinationCidrBlock: route.DestinationCIDR,
		NextHopType:          ecs.NextHopInstance,
		NextHopId:            string(route.TargetNode),
	}
	klog.Infof("CreateRoute:[%s] start to create route, %s -> %s", tabid, route.DestinationCIDR, route.TargetNode)
	return WaitCreate(r, tabid, args)
}

func isRouteExists(routes []*cloudprovider.Route, route *cloudprovider.Route) bool {
	for _, r := range routes {
		if r.DestinationCIDR == route.DestinationCIDR &&
			strings.Contains(string(r.TargetNode), string(route.TargetNode)) {
			klog.Infof("CreateRoute: skip exist route, %s -> %s", route.DestinationCIDR, route.TargetNode)
			return true
		}
	}
	return false
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (r *RoutesClient) DeleteRoute(tabid string, route *cloudprovider.Route, region common.Region) error {
	args := &ecs.DeleteRouteEntryArgs{
		RouteTableId:         tabid,
		DestinationCidrBlock: route.DestinationCIDR,
		NextHopId:            string(route.TargetNode),
	}
	return WaitDelete(r, tabid, args)
}

// WaitCreate create route and wait for route ready
func WaitCreate(rc *RoutesClient, tableid string, route *ecs.CreateRouteEntryArgs) error {
	err := rc.client.CreateRouteEntry(route)
	if err != nil {
		return fmt.Errorf("WaitCreate: ceate route for table %s error, %s", tableid, err.Error())
	}
	return WaitForRouteEntryAvailable(rc.client, rc.vpc.vrouterid, tableid)
}

// WaitDelete delete route and wait for route ready
func WaitDelete(rc *RoutesClient, tableid string, route *ecs.DeleteRouteEntryArgs) error {
	if err := rc.client.DeleteRouteEntry(route); err != nil {
		if strings.Contains(err.Error(), "InvalidRouteEntry.NotFound") {
			klog.Warningf("WaitDelete:[%s] route not found %s -> %s", tableid, route.DestinationCidrBlock, route.NextHopId)
			return nil
		}
		return fmt.Errorf("WaitDelete:[%s] delete route entry error: %s", tableid, err.Error())
	}
	return WaitForRouteEntryAvailable(rc.client, rc.vpc.vrouterid, tableid)
}

// Error implement error
func (r *RoutesClient) Error(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}

// WaitForRouteEntryAvailable wait for route entry available
func WaitForRouteEntryAvailable(client RouteSDK, routeid, tableid string) error {
	return client.WaitForAllRouteEntriesAvailable(routeid, tableid, 60)
}
