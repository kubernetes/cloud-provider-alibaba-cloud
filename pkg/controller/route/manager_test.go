package route

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	globalCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"net"
	"testing"
)

func TestGetRouteTables(t *testing.T) {
	meta := vmock.NewMockMetaData("vpc-xxx")
	noRouteTableVPC := vmock.MockCloud{MockVPC: vmock.NewMockVPC(nil), IMetaData: meta}
	singleRouteTableVPC := vmock.MockCloud{MockVPC: vmock.NewMockVPC(nil), IMetaData: meta}
	multiRouteTableVPC := vmock.MockCloud{MockVPC: vmock.NewMockVPC(nil), IMetaData: meta}

	globalCtx.CloudCFG.Global.RouteTableIDS = "table-xxx,table-yyy"
	tables, err := getRouteTables(context.Background(), noRouteTableVPC)
	assert.Equal(t, 2, len(tables), "assert route tables from cloud config")
	assert.Equal(t, nil, err, "assert route tables from cloud config")

	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), noRouteTableVPC)
	assert.Equal(t, 0, len(tables), "assert route tables from no table vpc")
	assert.Equal(t, false, err == nil, "assert route tables from no table vpc")

	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), singleRouteTableVPC)
	assert.Equal(t, 1, len(tables), "assert route tables from no single vpc")
	assert.Equal(t, nil, err, "assert route tables from single table vpc")

	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), multiRouteTableVPC)
	assert.Equal(t, 0, len(tables), "assert route tables from multi table vpc")
	assert.Equal(t, false, err == nil, "assert route tables from multi table vpc")
}

func TestContainsRoute(t *testing.T) {
	testcases := []struct {
		outside      *net.IPNet
		route        string
		contains     bool
		realContains bool
		err          bool
	}{
		{outside: &net.IPNet{
			IP:   net.ParseIP("192.168.0.0"),
			Mask: net.CIDRMask(16, 32),
		}, route: "192.168.1.0/24", contains: true, realContains: true, err: false},
		{outside: &net.IPNet{
			IP:   net.ParseIP("172.16.0.0"),
			Mask: net.CIDRMask(12, 32),
		}, route: "192.168.1.0/24", contains: false, realContains: false, err: false},
		{outside: &net.IPNet{
			IP:   net.ParseIP("172.16.0.0"),
			Mask: net.CIDRMask(12, 32),
		}, route: "172.16.0.0/12", contains: true, realContains: false, err: false},
		{outside: nil, route: "192.168.1.0/24", contains: true, realContains: true, err: false},
		{outside: &net.IPNet{
			IP:   net.ParseIP("172.16.0.0"),
			Mask: net.CIDRMask(12, 32),
		}, route: "192.168.1/24", contains: false, realContains: false, err: true},
		{outside: &net.IPNet{
			IP:   net.ParseIP("fe80::0"),
			Mask: net.CIDRMask(16, 128),
		}, route: "fe80:af01::1/32", contains: true, realContains: true, err: false},
		{outside: &net.IPNet{
			IP:   net.ParseIP("fe80::0"),
			Mask: net.CIDRMask(16, 128),
		}, route: "fe80::0/16", contains: true, realContains: true, err: false},
	}
	for _, testcase := range testcases {
		contains, realContains, err := containsRoute(testcase.outside, testcase.route)
		assert.Equal(t, testcase.contains, contains, fmt.Sprintf("contains: outside: %v, route: %v", testcase.outside, testcase.route))
		assert.Equal(t, testcase.realContains, realContains, fmt.Sprintf("realContains: outside: %v, route: %v", testcase.outside, testcase.route))
		assert.Equal(t, testcase.err, err != nil, fmt.Sprintf("error: outside: %v, route: %v", testcase.outside, testcase.route))
	}
}

func TestFindRouteCached(t *testing.T) {
	cachedRoute := []*model.Route{
		{DestinationCIDR: "192.168.0.0/24", ProviderId: "a-123"},
		{DestinationCIDR: "192.168.1.0/24", ProviderId: "a-234"},
		{DestinationCIDR: "192.168.2.0/24", ProviderId: "a-345"},
	}
	testcases := []struct {
		routes []*model.Route
		pvid   string
		cidr   string
		found  string
		err    bool
	}{
		{
			routes: cachedRoute,
			pvid:   "a-123",
			cidr:   "192.168.0.0/24",
			found:  "a-123",
			err:    false,
		},
		{
			routes: cachedRoute,
			pvid:   "a-123",
			found:  "a-123",
			err:    false,
		},
		{
			routes: cachedRoute,
			cidr:   "192.168.0.0/24",
			found:  "a-123",
			err:    false,
		},
		{
			routes: cachedRoute,
			pvid:   "a-1234",
			cidr:   "192.168.0.0/24",
			found:  "",
			err:    false,
		},
	}

	for _, testcase := range testcases {
		route, err := findRoute(context.Background(), "", testcase.pvid, testcase.cidr, testcase.routes, nil)
		if testcase.found != "" {
			assert.Equal(t, testcase.found, route.ProviderId)
		} else {
			assert.Empty(t, route)
		}
		assert.Equal(t, testcase.err, err != nil)
	}
}
