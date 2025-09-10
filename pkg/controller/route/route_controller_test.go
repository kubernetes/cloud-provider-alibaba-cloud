package route

import (
	"context"
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	NodeName = "cn-hangzhou.192.0.168.68"
)

func TestUpdateNetworkingCondition(t *testing.T) {
	r := getReconcileRoute()
	nodes := &v1.NodeList{}
	err := r.client.List(context.TODO(), nodes)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("routeCreated true patch success", func(t *testing.T) {
		for i := range nodes.Items {
			node := &nodes.Items[i]
			err := r.updateNetworkingCondition(context.TODO(), node, true)
			if err != nil {
				t.Error(err)
			}
			updatedNode := &v1.Node{}
			err = r.client.Get(context.TODO(), util.NamespacedName(node), updatedNode)
			if err != nil {
				t.Error(err)
			}
			networkCondition, ok := helper.FindCondition(updatedNode.Status.Conditions, v1.NodeNetworkUnavailable)
			if !ok || networkCondition.Status != v1.ConditionFalse {
				t.Error("node condition update failed")
			}
		}
	})

	t.Run("routeCreated true already ConditionFalse early return", func(t *testing.T) {
		var nodeWithFalse *v1.Node
		for i := range nodes.Items {
			if nodes.Items[i].Status.Conditions != nil {
				for _, c := range nodes.Items[i].Status.Conditions {
					if c.Type == v1.NodeNetworkUnavailable && c.Status == v1.ConditionFalse {
						nodeWithFalse = &nodes.Items[i]
						break
					}
				}
			}
			if nodeWithFalse != nil {
				break
			}
		}
		if nodeWithFalse == nil {
			t.Skip("no node with NodeNetworkUnavailable=False in fixture")
		}
		err := r.updateNetworkingCondition(context.TODO(), nodeWithFalse, true)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("routeCreated false already ConditionTrue early return", func(t *testing.T) {
		var nodeWithTrue *v1.Node
		for i := range nodes.Items {
			if nodes.Items[i].Name == NodeName {
				nodeWithTrue = &nodes.Items[i]
				break
			}
		}
		if nodeWithTrue == nil {
			t.Fatal("fixture node not found")
		}
		err := r.updateNetworkingCondition(context.TODO(), nodeWithTrue, false)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("routeCreated false patch to True", func(t *testing.T) {
		var nodeWithFalse *v1.Node
		for i := range nodes.Items {
			for _, c := range nodes.Items[i].Status.Conditions {
				if c.Type == v1.NodeNetworkUnavailable && c.Status == v1.ConditionFalse {
					nodeWithFalse = &nodes.Items[i]
					break
				}
			}
			if nodeWithFalse != nil {
				break
			}
		}
		if nodeWithFalse == nil {
			t.Skip("no node with NodeNetworkUnavailable=False")
		}
		err := r.updateNetworkingCondition(context.TODO(), nodeWithFalse, false)
		if err != nil {
			t.Error(err)
		}
		updatedNode := &v1.Node{}
		err = r.client.Get(context.TODO(), util.NamespacedName(nodeWithFalse), updatedNode)
		if err != nil {
			t.Error(err)
		}
		networkCondition, ok := helper.FindCondition(updatedNode.Status.Conditions, v1.NodeNetworkUnavailable)
		if !ok || networkCondition.Status != v1.ConditionTrue {
			t.Error("expected ConditionTrue after routeCreated false")
		}
	})
}

func TestSyncTableRoutes(t *testing.T) {
	oldCIDR := ctrlCfg.ControllerCFG.ClusterCIDR
	defer func() { ctrlCfg.ControllerCFG.ClusterCIDR = oldCIDR }()
	ctrlCfg.ControllerCFG.ClusterCIDR = "10.96.0.0/16"

	r := getReconcileRoute()
	nodes := &v1.NodeList{}
	err := r.client.List(context.TODO(), nodes)
	if err != nil {
		t.Fatal(err)
	}
	err = r.syncTableRoutes(context.Background(), "route-table-1", nodes)
	if err != nil {
		t.Error(err)
	}

	t.Run("ListRoute error", func(t *testing.T) {
		err := r.syncTableRoutes(context.Background(), "route-table-list-err", nodes)
		if err == nil {
			t.Error("expected error from ListRoute")
		}
	})

	t.Run("invalid ClusterCIDR parse error", func(t *testing.T) {
		ctrlCfg.ControllerCFG.ClusterCIDR = "invalid"
		defer func() { ctrlCfg.ControllerCFG.ClusterCIDR = "10.96.0.0/16" }()
		err := r.syncTableRoutes(context.Background(), "route-table-1", nodes)
		if err == nil {
			t.Error("expected parse error for invalid ClusterCIDR")
		}
	})

	t.Run("node with needSyncRoute false skipped", func(t *testing.T) {
		excludedNode := v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "excluded-node",
				Labels: map[string]string{
					helper.LabelNodeExcludeNode: "true",
				},
			},
			Spec: v1.NodeSpec{ProviderID: "cn-hangzhou.i-123", PodCIDR: "10.96.0.0/24"},
		}
		customNodes := &v1.NodeList{Items: append(nodes.Items, excludedNode)}
		err := r.syncTableRoutes(context.Background(), "route-table-1", customNodes)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("node with empty ProviderID skipped", func(t *testing.T) {
		noProviderNode := v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "no-provider-node"},
			Spec:       v1.NodeSpec{ProviderID: "", PodCIDR: "10.96.0.0/24"},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}},
			},
		}
		customNodes := &v1.NodeList{Items: append(nodes.Items, noProviderNode)}
		err := r.syncTableRoutes(context.Background(), "route-table-1", customNodes)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("empty ClusterCIDR", func(t *testing.T) {
		oldCIDR := ctrlCfg.ControllerCFG.ClusterCIDR
		ctrlCfg.ControllerCFG.ClusterCIDR = ""
		defer func() { ctrlCfg.ControllerCFG.ClusterCIDR = oldCIDR }()
		err := r.syncTableRoutes(context.Background(), "route-table-1", nodes)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("route with invalid DestinationCIDR", func(t *testing.T) {
		err := r.syncTableRoutes(context.Background(), "route-table-invalid-cidr", nodes)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestAddRouteForNode(t *testing.T) {
	r := getReconcileRoute()
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
		Spec:       v1.NodeSpec{ProviderID: "cn-hangzhou.ecs-id", PodCIDR: "10.96.0.64/26"},
	}
	err := r.addRouteForNode(context.Background(), "route-table-1", "10.96.0.64/26", "cn-hangzhou.ecs-id", node, nil)
	if err != nil {
		t.Error(err)
	}

	err = r.addRouteForNode(context.Background(), "route-table-1", "", "", node, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestReconcileRoute_Reconcile(t *testing.T) {
	t.Run("configRoutes disabled", func(t *testing.T) {
		r := getReconcileRoute()
		r.configRoutes = false

		request := reconcile.Request{
			NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: NodeName}}),
		}

		result, err := r.Reconcile(context.TODO(), request)
		if err != nil {
			t.Errorf("Reconcile failed: %v", err)
		}

		if result != (reconcile.Result{}) {
			t.Errorf("Expected empty result, got: %v", result)
		}

		select {
		case <-r.requestChan:
			t.Error("Expected no message in requestChan when configRoutes is false")
		default:
			// Expected behavior - channel should be empty
		}
	})

	t.Run("configRoutes enabled", func(t *testing.T) {
		r := getReconcileRoute()
		r.configRoutes = true

		request := reconcile.Request{
			NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: NodeName}}),
		}

		result, err := r.Reconcile(context.TODO(), request)
		if err != nil {
			t.Errorf("Reconcile failed: %v", err)
		}

		if result != (reconcile.Result{}) {
			t.Errorf("Expected empty result, got: %v", result)
		}

		select {
		case receivedRequest := <-r.requestChan:
			if receivedRequest != request {
				t.Errorf("Expected request %v, got: %v", request, receivedRequest)
			}
		case <-time.After(1 * time.Second):
			t.Error("Expected message in requestChan but timed out")
		}
	})
}

func TestBatchAddRoutes(t *testing.T) {
	r := getReconcileRoute()
	ctx := context.Background()
	nodes := &v1.NodeList{}
	if err := r.client.List(ctx, nodes); err != nil {
		t.Fatal(err)
	}
	nodeByName := make(map[string]*v1.Node)
	for i := range nodes.Items {
		nodeByName[nodes.Items[i].Name] = &nodes.Items[i]
	}
	makeRoute := func(name, cidr, providerID string, node *v1.Node) *model.Route {
		return &model.Route{
			Name:            name,
			DestinationCIDR: cidr,
			ProviderId:      providerID,
			NodeReference:   node,
		}
	}

	t.Run("empty routes", func(t *testing.T) {
		err := r.batchAddRoutes(ctx, "rid", "route-table-1", nil)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("duplicate and status_error treated as success", func(t *testing.T) {
		dupNode := nodeByName["dup-cidr"]
		statusNode := nodeByName["status-err"]
		if dupNode == nil || statusNode == nil {
			t.Skip("fixture nodes not found")
		}
		routes := []*model.Route{
			makeRoute("r1", "10.96.0.192/26", "i-dup", dupNode),
			makeRoute("r2", "10.96.0.0/28", "i-se", statusNode),
		}
		err := r.batchAddRoutes(ctx, "rid", "route-table-1", routes)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create failed requeue", func(t *testing.T) {
		cfNode := nodeByName["create-fail"]
		if cfNode == nil {
			t.Skip("create-fail node not found")
		}
		routes := []*model.Route{
			makeRoute("r-cf", "10.96.0.16/28", "i-cf", cfNode),
		}
		err := r.batchAddRoutes(ctx, "rid", "route-table-1", routes)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestReconcileRoute_BatchSyncCloudRoutes(t *testing.T) {
	t.Run("no nodes to process", func(t *testing.T) {
		r := getReconcileRoute()

		// Test empty request list
		err := r.batchSyncCloudRoutes(context.Background(), "test-reconcile-id", []reconcile.Request{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("node not found", func(t *testing.T) {
		r := getReconcileRoute()
		requests := []reconcile.Request{
			{
				NamespacedName: util.NamespacedName(&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "non-existent-node"},
				}),
			},
		}

		err := r.batchSyncCloudRoutes(context.Background(), "test-reconcile-id", requests)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("node not found with cache entry", func(t *testing.T) {
		r := getReconcileRoute()
		r.nodeCache.Set("deleted-node", &model.Route{
			Name:            "r-deleted",
			DestinationCIDR: "10.96.0.0/24",
			ProviderId:      "i-deleted",
			NodeReference:   &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "deleted-node"}},
		})
		requests := []reconcile.Request{
			{NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "deleted-node"}})},
		}
		err := r.batchSyncCloudRoutes(context.Background(), "rid", requests)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("node with no provider id", func(t *testing.T) {
		r := getReconcileRoute()
		requests := []reconcile.Request{
			{
				NamespacedName: util.NamespacedName(&v1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: "cn-hangzhou.192.0.168.69"},
				}),
			},
		}
		node := &v1.Node{}
		err := r.client.Get(context.Background(), util.NamespacedName(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "cn-hangzhou.192.0.168.69"},
		}), node)
		if err != nil {
			t.Fatalf("Failed to get node: %v", err)
		}

		node.Spec.ProviderID = ""
		err = r.client.Update(context.Background(), node)
		if err != nil {
			t.Fatalf("Failed to update node: %v", err)
		}

		err = r.batchSyncCloudRoutes(context.Background(), "test-reconcile-id", requests)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("invalid providerID requeue", func(t *testing.T) {
		r := getReconcileRoute()
		requests := []reconcile.Request{
			{NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "invalid-provider"}})},
		}
		err := r.batchSyncCloudRoutes(context.Background(), "rid", requests)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("bad PodCIDR update condition and requeue", func(t *testing.T) {
		r := getReconcileRoute()
		requests := []reconcile.Request{
			{NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "bad-podcidr"}})},
		}
		err := r.batchSyncCloudRoutes(context.Background(), "rid", requests)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("getRouteTables error", func(t *testing.T) {
		cloudNoTable := vmock.MockCloud{
			MockVPC:   vmock.NewMockVPC(nil),
			IMetaData: vmock.NewMockMetaData("vpc-no-route-table"),
		}
		oldRouteTableIDS := ctrlCfg.CloudCFG.Global.RouteTableIDS
		ctrlCfg.CloudCFG.Global.RouteTableIDS = ""
		defer func() { ctrlCfg.CloudCFG.Global.RouteTableIDS = oldRouteTableIDS }()
		r := getReconcileRouteWithCloud(cloudNoTable)
		requests := []reconcile.Request{
			{NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: NodeName}})},
		}
		err := r.batchSyncCloudRoutes(context.Background(), "rid", requests)
		if err == nil {
			t.Error("expected getRouteTables error")
		}
	})
}

func TestReconcileRoute_ReconcileForCluster(t *testing.T) {
	t.Run("reconcileForCluster success", func(t *testing.T) {
		r := getReconcileRoute()
		r.reconcileForCluster()
	})

	t.Run("reconcileForCluster getRouteTables error", func(t *testing.T) {
		cloudNoTable := vmock.MockCloud{
			MockVPC:   vmock.NewMockVPC(nil),
			IMetaData: vmock.NewMockMetaData("vpc-no-route-table"),
		}
		oldRouteTableIDS := ctrlCfg.CloudCFG.Global.RouteTableIDS
		ctrlCfg.CloudCFG.Global.RouteTableIDS = ""
		defer func() { ctrlCfg.CloudCFG.Global.RouteTableIDS = oldRouteTableIDS }()
		r := getReconcileRouteWithCloud(cloudNoTable)
		r.reconcileForCluster()
	})

	t.Run("reconcileForCluster syncTableRoutes error", func(t *testing.T) {
		oldRouteTableIDS := ctrlCfg.CloudCFG.Global.RouteTableIDS
		ctrlCfg.CloudCFG.Global.RouteTableIDS = "route-table-list-err"
		defer func() { ctrlCfg.CloudCFG.Global.RouteTableIDS = oldRouteTableIDS }()
		r := getReconcileRoute()
		r.reconcileForCluster()
	})
}

func TestReconcileRoute_BatchWorker(t *testing.T) {
	ctrlCfg.ControllerCFG.RouteReconcileBatchSize = 10

	t.Run("batchWorker context cancel", func(t *testing.T) {
		r := getReconcileRoute()

		// Create a cancellable context
		ctx, cancel := context.WithCancel(context.Background())

		// Start batchWorker
		go r.batchWorker(ctx, 0)

		// Cancel context immediately
		cancel()

		// Wait a short time to ensure goroutine has time to process cancel signal
		time.Sleep(100 * time.Millisecond)

		// Test passes if no panic occurs
	})

	t.Run("batchWorker process requests", func(t *testing.T) {
		r := getReconcileRoute()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go r.batchWorker(ctx, 1)

		request := reconcile.Request{
			NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: NodeName}}),
		}
		r.requestChan <- request

		time.Sleep(2 * time.Second)
	})

	t.Run("batchWorker handle multiple requests", func(t *testing.T) {
		r := getReconcileRoute()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go r.batchWorker(ctx, 2)
		request1 := reconcile.Request{
			NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: NodeName}}),
		}
		request2 := reconcile.Request{
			NamespacedName: util.NamespacedName(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "cn-hangzhou.192.0.168.69"}}),
		}

		r.requestChan <- request1
		r.requestChan <- request2
		r.requestChan <- request1 // Duplicate request

		time.Sleep(2 * time.Second)
	})
}

func TestRequeueNode(t *testing.T) {
	t.Run("requeue sent", func(t *testing.T) {
		requeueCh := make(chan event.GenericEvent, 2)
		r := &ReconcileRoute{
			cloud:       getMockCloudProvider(),
			client:      getFakeKubeClient(),
			record:      record.NewFakeRecorder(10),
			nodeCache:   cmap.New(),
			requeueChan: requeueCh,
			requestChan: make(chan reconcile.Request, 10),
		}
		n := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		r.requeueNode(n)
		ev := <-requeueCh
		if ev.Object.GetName() != "test-node" {
			t.Errorf("unexpected node name: %s", ev.Object.GetName())
		}
	})

	t.Run("requeue channel full", func(t *testing.T) {
		requeueCh := make(chan event.GenericEvent, 1)
		r := &ReconcileRoute{
			cloud:        getMockCloudProvider(),
			client:       getFakeKubeClient(),
			record:       record.NewFakeRecorder(10),
			nodeCache:    cmap.New(),
			requeueChan:  requeueCh,
			requestChan:  make(chan reconcile.Request, 10),
		}
		requeueCh <- event.GenericEvent{Object: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "blocking"}}}
		n := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "dropped"}}
		r.requeueNode(n)
		ev := <-requeueCh
		if ev.Object.GetName() != "blocking" {
			t.Errorf("expected blocking, got %s", ev.Object.GetName())
		}
	})
}

func getReconcileRoute() *ReconcileRoute {
	eventRecord := record.NewFakeRecorder(100)
	requeueCh := make(chan event.GenericEvent, 10)
	recon := &ReconcileRoute{
		cloud:        getMockCloudProvider(),
		client:       getFakeKubeClient(),
		record:       eventRecord,
		nodeCache:    cmap.New(),
		configRoutes: true,
		requeueChan:  requeueCh,
		rateLimiter: workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		),
		requestChan: make(chan reconcile.Request, 10),
	}
	return recon
}

func getMockCloudProvider() prvd.Provider {
	return vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
}

func getReconcileRouteWithCloud(cloud prvd.Provider) *ReconcileRoute {
	eventRecord := record.NewFakeRecorder(100)
	requeueCh := make(chan event.GenericEvent, 10)
	return &ReconcileRoute{
		cloud:        cloud,
		client:       getFakeKubeClient(),
		record:       eventRecord,
		nodeCache:    cmap.New(),
		configRoutes: true,
		requeueChan:  requeueCh,
		rateLimiter: workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		),
		requestChan: make(chan reconcile.Request, 10),
	}
}

func getFakeKubeClient() client.Client {
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: NodeName,
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.64/26",
					ProviderID: "cn-hangzhou.ecs-id",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
						},
						{
							Type:   v1.NodeNetworkUnavailable,
							Reason: string(v1.NodeNetworkUnavailable),
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cn-hangzhou.192.0.168.69",
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.128/26",
					ProviderID: "alicloud://cn-hangzhou.ecs-id",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
						},
						{
							Type:   v1.NodeNetworkUnavailable,
							Reason: string(v1.NodeNetworkUnavailable),
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "dup-cidr"},
				Spec:       v1.NodeSpec{ProviderID: "alicloud://i-dup", PodCIDR: "10.96.0.192/26"},
				Status:     v1.NodeStatus{Conditions: []v1.NodeCondition{}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "status-err"},
				Spec:       v1.NodeSpec{ProviderID: "alicloud://i-se", PodCIDR: "10.96.0.0/28"},
				Status:     v1.NodeStatus{Conditions: []v1.NodeCondition{}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "create-fail"},
				Spec:       v1.NodeSpec{ProviderID: "alicloud://i-cf", PodCIDR: "10.96.0.16/28"},
				Status:     v1.NodeStatus{Conditions: []v1.NodeCondition{}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "invalid-provider"},
				Spec:       v1.NodeSpec{ProviderID: "x", PodCIDR: "10.96.0.32/28"},
				Status:     v1.NodeStatus{Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}}},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "bad-podcidr"},
				Spec:       v1.NodeSpec{ProviderID: "alicloud://cn-hangzhou.i-bad", PodCIDR: ""},
				Status:     v1.NodeStatus{Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}}},
			},
		},
	}

	objs := []runtime.Object{nodeList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}
