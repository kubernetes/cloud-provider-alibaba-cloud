package route

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	globalCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestGetRouteTables(t *testing.T) {
	// Setup test cases with different VPC configurations
	globalCtx.CloudCFG.Global.VpcID = "vpc-test-id"

	noRouteTableVPC := vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-no-route-table"),
	}
	singleRouteTableVPC := vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
	multiRouteTableVPC := vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-multi-route-table"),
	}

	// Test with route table IDs specified in config
	globalCtx.CloudCFG.Global.RouteTableIDS = "table-xxx,table-yyy"
	tables, err := getRouteTables(context.Background(), noRouteTableVPC)
	assert.Equal(t, 2, len(tables), "assert route tables from cloud config")
	assert.Equal(t, []string{"table-xxx", "table-yyy"}, tables, "assert route tables from cloud config")
	assert.NoError(t, err, "assert route tables from cloud config")

	// Test with no route tables in VPC and no config
	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), noRouteTableVPC)
	assert.Equal(t, 0, len(tables), "assert route tables from no table vpc")
	assert.Error(t, err, "assert route tables from no table vpc")
	assert.Equal(t, "no route tables found by vpc id[vpc-test-id]", err.Error(), "assert route tables from no table vpc")

	// Test with single route table in VPC
	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), singleRouteTableVPC)
	assert.Equal(t, 1, len(tables), "assert route tables from no single vpc")
	assert.Equal(t, []string{"route-table-1"}, tables, "assert route tables from single table vpc")
	assert.NoError(t, err, "assert route tables from single table vpc")

	// Test with multiple route tables in VPC
	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), multiRouteTableVPC)
	assert.Equal(t, 0, len(tables), "assert route tables from multi table vpc")
	assert.Error(t, err, "assert route tables from multi table vpc")
	assert.Equal(t, "multiple route tables found by vpc id[vpc-test-id], length(tables)=2", err.Error(), "assert route tables from multi table vpc")

	// Reset global config
	globalCtx.CloudCFG.Global.RouteTableIDS = ""
	globalCtx.CloudCFG.Global.VpcID = ""
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

func TestConflictWithNodes(t *testing.T) {
	// Create test node list
	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
				Spec: v1.NodeSpec{
					ProviderID: "i-123",
					PodCIDR:    "192.168.1.0/24",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node2",
				},
				Spec: v1.NodeSpec{
					ProviderID: "i-456",
					PodCIDR:    "192.168.2.0/24",
					PodCIDRs:   []string{"192.168.3.0/24"},
				},
			},
		},
	}

	testCases := []struct {
		name     string
		route    *model.Route
		expected bool
	}{
		{
			name: "Conflict route - Same CIDR but different ProviderID",
			route: &model.Route{
				DestinationCIDR: "192.168.1.0/24",
				ProviderId:      "i-789", // Same CIDR as node1 but different ProviderID
			},
			expected: true,
		},
		{
			name: "Conflict route - Conflict with PodCIDRs",
			route: &model.Route{
				DestinationCIDR: "192.168.3.0/24",
				ProviderId:      "i-789", // Same CIDR as node1 but different ProviderID
			},
			expected: true,
		},
		{
			name: "Non-conflict route - Both CIDR and ProviderID match",
			route: &model.Route{
				DestinationCIDR: "192.168.1.0/24",
				ProviderId:      "i-123", // Complete match with node1
			},
			expected: false,
		},
		{
			name: "Non-conflict route - CIDR does not match",
			route: &model.Route{
				DestinationCIDR: "10.0.0.0/24",
				ProviderId:      "i-789",
			},
			expected: false,
		},
		{
			name: "Conflict route - Subnet inclusion relationship",
			route: &model.Route{
				DestinationCIDR: "192.168.1.0/30", // Contains node1's CIDR
				ProviderId:      "i-789",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := conflictWithNodes(tc.route, nodes)
			assert.Equal(t, tc.expected, result)
		})
	}

	t.Run("node with invalid PodCIDR", func(t *testing.T) {
		invalidNodes := &v1.NodeList{
			Items: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "bad-node"},
					Spec: v1.NodeSpec{
						ProviderID: "i-999",
						PodCIDR:    "invalid-cidr",
					},
				},
			},
		}
		route := &model.Route{DestinationCIDR: "10.0.0.0/24", ProviderId: "i-999"}
		result := conflictWithNodes(route, invalidNodes)
		assert.False(t, result)
	})

	t.Run("node with no PodCIDR", func(t *testing.T) {
		noCIDRNodes := &v1.NodeList{
			Items: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "no-cidr-node"},
					Spec:       v1.NodeSpec{ProviderID: "i-888"},
				},
			},
		}
		route := &model.Route{DestinationCIDR: "10.0.0.0/24", ProviderId: "i-888"}
		result := conflictWithNodes(route, noCIDRNodes)
		assert.False(t, result)
	})

	t.Run("route with unparsable DestinationCIDR", func(t *testing.T) {
		nodes := &v1.NodeList{
			Items: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node1"},
					Spec:       v1.NodeSpec{ProviderID: "i-123", PodCIDR: "192.168.1.0/24"},
				},
			},
		}
		route := &model.Route{DestinationCIDR: "invalid-cidr", ProviderId: "i-123"}
		result := conflictWithNodes(route, nodes)
		assert.False(t, result)
	})
}

func TestCreateRouteForInstance(t *testing.T) {
	ctx := context.Background()
	mockVPC := vmock.NewMockVPC(nil)
	route, err := createRouteForInstance(ctx, "route-table-1", "i-123", "10.96.0.0/24", mockVPC)
	assert.NoError(t, err)
	assert.NotNil(t, route)
	assert.Equal(t, "10.96.0.0/24", route.DestinationCIDR)
	assert.Equal(t, "i-123", route.ProviderId)

	t.Run("duplicate CIDR FindRoute succeeds", func(t *testing.T) {
		route2, err2 := createRouteForInstance(ctx, "route-table-1", "i-duplicate", "10.96.0.0/24", mockVPC)
		assert.NoError(t, err2)
		assert.NotNil(t, route2)
		assert.Equal(t, "i-duplicate", route2.ProviderId)
	})

	t.Run("duplicate CIDR FindRoute returns nil", func(t *testing.T) {
		_, err := createRouteForInstance(ctx, "route-table-1", "i-dup-find-fail", "10.96.0.0/24", mockVPC)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "InvalidCIDRBlock.Duplicate")
	})
}

func TestDeleteRouteForInstance(t *testing.T) {
	ctx := context.Background()
	mockVPC := vmock.NewMockVPC(nil)
	err := deleteRouteForInstance(ctx, "route-table-1", "i-123", "10.96.0.0/24", mockVPC)
	assert.NoError(t, err)
}

func TestFindRoute(t *testing.T) {
	ctx := context.Background()
	mockVPC := vmock.NewMockVPC(nil)

	_, err := findRoute(ctx, "t", "", "", nil, mockVPC)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty query condition")

	cached := []*model.Route{
		{DestinationCIDR: "10.96.0.0/24", ProviderId: "i-1", Name: "r1"},
		{DestinationCIDR: "10.96.0.128/26", ProviderId: "i-2", Name: "r2"},
	}
	r, err := findRoute(ctx, "t", "i-1", "10.96.0.0/24", cached, mockVPC)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "i-1", r.ProviderId)

	r, err = findRoute(ctx, "t", "i-2", "", cached, mockVPC)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "i-2", r.ProviderId)

	r, err = findRoute(ctx, "t", "", "10.96.0.128/26", cached, mockVPC)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "10.96.0.128/26", r.DestinationCIDR)

	r, err = findRoute(ctx, "t", "i-99", "10.0.0.0/24", cached, mockVPC)
	assert.NoError(t, err)
	assert.Nil(t, r)

	r, err = findRoute(ctx, "t", "i-123", "192.168.1.0/24", nil, mockVPC)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestNeedSyncRoute(t *testing.T) {
	// Normal node
	normalNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "normal-node",
		},
		Spec: v1.NodeSpec{
			ProviderID: "i-123",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	// Node with exclusion label
	excludedNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "excluded-node",
			Labels: map[string]string{
				helper.LabelNodeExcludeNode: "true",
			},
		},
		Spec: v1.NodeSpec{
			ProviderID: "i-456",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	// Node with unknown status
	unknownNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unknown-node",
		},
		Spec: v1.NodeSpec{
			ProviderID: "i-789",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionUnknown,
				},
			},
		},
	}

	// Node being deleted
	deletingNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-node",
			DeletionTimestamp: &metav1.Time{},
		},
		Spec: v1.NodeSpec{
			ProviderID: "i-101",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	// Node without ProviderID
	noProviderNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "no-provider-node",
		},
		Spec: v1.NodeSpec{
			ProviderID: "",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	testCases := []struct {
		name     string
		node     *v1.Node
		expected bool
	}{
		{
			name:     "Normal node",
			node:     normalNode,
			expected: true,
		},
		{
			name:     "Node with exclusion label",
			node:     excludedNode,
			expected: false,
		},
		{
			name:     "Node with unknown status",
			node:     unknownNode,
			expected: false,
		},
		{
			name:     "Node being deleted",
			node:     deletingNode,
			expected: false,
		},
		{
			name:     "Node without ProviderID",
			node:     noProviderNode,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := needSyncRoute(tc.node)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFindRouteWithoutCache(t *testing.T) {
	// Create mock provider with predefined routes
	mockProvider := &vmock.MockVPC{}

	testCases := []struct {
		name        string
		table       string
		pvid        string
		cidr        string
		expectFound bool
		expectName  string
		expectPVID  string
		expectCIDR  string
		expectError bool
	}{
		{
			name:        "Find by ProviderID and CIDR",
			table:       "table-1",
			pvid:        "i-123",
			cidr:        "192.168.1.0/24",
			expectFound: true,
			expectName:  "route-1",
			expectPVID:  "i-123",
			expectCIDR:  "192.168.1.0/24",
			expectError: false,
		},
		{
			name:        "Find by ProviderID only",
			table:       "table-1",
			pvid:        "i-456",
			cidr:        "",
			expectFound: true,
			expectName:  "route-2",
			expectPVID:  "i-456",
			expectCIDR:  "192.168.2.0/24",
			expectError: false,
		},
		{
			name:        "Find by CIDR only",
			table:       "table-1",
			pvid:        "",
			cidr:        "192.168.1.0/24",
			expectFound: true,
			expectName:  "route-1",
			expectPVID:  "i-123",
			expectCIDR:  "192.168.1.0/24",
			expectError: false,
		},
		{
			name:        "No matching route found",
			table:       "table-1",
			pvid:        "i-999",
			cidr:        "10.0.0.0/24",
			expectFound: false,
			expectError: false,
		},
		{
			name:        "Empty query conditions",
			table:       "table-1",
			pvid:        "",
			cidr:        "",
			expectFound: false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			route, err := findRoute(context.Background(), tc.table, tc.pvid, tc.cidr, nil, mockProvider)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, route)
			} else {
				assert.NoError(t, err)
			}

			if tc.expectFound {
				assert.NotNil(t, route)
				if tc.expectName != "" {
					assert.Equal(t, tc.expectName, route.Name)
				}
				if tc.expectPVID != "" {
					assert.Equal(t, tc.expectPVID, route.ProviderId)
				}
				if tc.expectCIDR != "" {
					assert.Equal(t, tc.expectCIDR, route.DestinationCIDR)
				}
			} else {
				assert.Nil(t, route)
			}
		})
	}
}

func TestBatchDeleteRoutes(t *testing.T) {
	cloud := vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
	eventRecord := record.NewFakeRecorder(100)
	rateLimiter := workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 10*time.Second)
	nodeCache := cmap.New()
	requeueChan := make(chan event.GenericEvent, 10)

	recon := &ReconcileRoute{
		cloud:       cloud,
		record:      eventRecord,
		nodeCache:   nodeCache,
		rateLimiter: rateLimiter,
		requeueChan: requeueChan,
	}

	ctx := context.Background()
	reconcileID := "test-reconcile-id"
	table := "table-1"

	t.Run("empty routes list", func(t *testing.T) {
		err := recon.batchDeleteRoutes(ctx, reconcileID, table, []*model.Route{})
		assert.NoError(t, err)
	})

	t.Run("successful delete", func(t *testing.T) {
		nodeRef := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
			},
		}
		routes := []*model.Route{
			{
				Name:            "route-1",
				DestinationCIDR: "192.168.1.0/24",
				ProviderId:      "i-123",
				NodeReference:   nodeRef,
			},
		}
		nodeCache.Set("node-1", routes[0])
		err := recon.batchDeleteRoutes(ctx, reconcileID, table, routes)
		assert.NoError(t, err)
		assert.False(t, nodeCache.Has("node-1"))
	})

	t.Run("route not exist ignore", func(t *testing.T) {
		nodeRef := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-not-exist"}}
		routes := []*model.Route{
			{ProviderId: "i-delete-not-exist", NodeReference: nodeRef},
		}
		go func() { <-requeueChan }()
		err := recon.batchDeleteRoutes(ctx, reconcileID, table, routes)
		assert.NoError(t, err)
	})

	t.Run("delete failed requeue", func(t *testing.T) {
		nodeRef := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-fail"}}
		routes := []*model.Route{
			{ProviderId: "i-delete-fail", NodeReference: nodeRef},
		}
		go func() { <-requeueChan }()
		err := recon.batchDeleteRoutes(ctx, reconcileID, table, routes)
		assert.NoError(t, err)
	})
}
