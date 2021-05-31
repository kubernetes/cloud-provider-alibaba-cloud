package route

import (
	"context"
	"fmt"
	globalCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog"
	"strings"
)

func createRouteForInstance(ctx context.Context, table, providerID, cidr string, providerIns prvd.IVPC) (*model.Route, error) {
	klog.Infof("create routes for node: %v -> %v", providerID, cidr)
	return providerIns.CreateRoute(ctx, table, providerID, cidr)
}

func deleteRouteForInstance(ctx context.Context, table, providerID, cidr string, providerIns prvd.IVPC) error {
	klog.Infof("delete route for node: %v", providerID)
	return providerIns.DeleteRoute(ctx, table, providerID, cidr)
}

func getRouteTables(ctx context.Context, providerIns prvd.Provider) ([]string, error) {
	vpcId, err := providerIns.VpcID()
	if err != nil {
		return nil, fmt.Errorf("get vpc id from metadata error: %s", err.Error())
	}
	if globalCtx.CFG.Global.RouteTableIDS != "" {
		return strings.Split(globalCtx.CFG.Global.RouteTableIDS, ","), nil
	}
	tables, err := providerIns.ListRouteTables(ctx, vpcId)
	if err != nil {
		return nil, fmt.Errorf("alicloud: "+
			"can not found routetable by id[%s], error: %v", globalCtx.CFG.Global.VpcID, err)
	}
	if len(tables) > 1 {
		return nil, fmt.Errorf("alicloud: "+
			"multiple vpc found by id[%s], length(vpcs)=%d", globalCtx.CFG.Global.VpcID, len(tables))
	}
	return tables, nil
}
