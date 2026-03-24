package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	NodeName = fmt.Sprintf("cn-hangzhou.%s", vmock.InstanceIP)
)

func TestSyncLingJunNodes(t *testing.T) {
	cases := []struct {
		node                v1.Node
		delete              bool
		removeUninitialized bool
	}{
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node-ready",
					Labels: map[string]string{
						LabelLingJunWorker:      "true",
						LabelLingJunNodeGroupID: "node-group-id",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-test-1",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			delete: false,
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node-uninitialized",
					Labels: map[string]string{
						LabelLingJunWorker:      "true",
						LabelLingJunNodeGroupID: "node-group-id",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-test-2",
					Taints: []v1.Taint{
						{
							Key:   api.TaintExternalCloudProvider,
							Value: "true",
						},
					},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			delete:              false,
			removeUninitialized: true,
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node-unknown",
					Labels: map[string]string{
						LabelLingJunWorker:      "true",
						LabelLingJunNodeGroupID: "node-group-id",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-test-3",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			delete: false,
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node-unknown-notfound",
					Labels: map[string]string{
						LabelLingJunWorker:      "true",
						LabelLingJunNodeGroupID: "node-group-id",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-test-notfound",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			delete: true,
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node-unknown-notfound-tianwen",
					Labels: map[string]string{
						LabelLingJunWorker:             "true",
						LabelLingJunNodeGroupID:        "node-group-id-2",
						LabelLingJunTianwenEnvironment: "tianwen-123",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-test-notfound",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			delete: false,
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node-unknown-different-nodegroup",
					Labels: map[string]string{
						LabelLingJunWorker:      "true",
						LabelLingJunNodeGroupID: "node-group-id-2",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-test",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionUnknown,
						},
					},
				},
			},
			delete: false,
		},
	}

	client := fake.NewClientBuilder().Build()
	for _, c := range cases {
		err := client.Create(context.Background(), &c.node)
		assert.NoError(t, err)
	}

	list, err := NodeList(client, false)
	assert.NoError(t, err)

	recon := getReconcileNode()
	recon.client = client
	err = recon.syncLingJunNodes(list.Items, false)
	assert.NoError(t, err)

	// wait for nodes to be deleted
	time.Sleep(1 * time.Second)

	list, err = NodeList(client, false)
	assert.NoError(t, err)

	for _, c := range cases {
		t.Run(c.node.Name, func(t *testing.T) {
			var node *v1.Node
			for _, n := range list.Items {
				if c.node.Name == n.Name {
					node = &n
					break
				}
			}
			if c.delete {
				assert.Nil(t, node)
			} else {
				assert.NotNil(t, node)
				if node != nil {
					if c.removeUninitialized {
						assert.Nil(t, findCloudTaint(node.Spec.Taints))
					}
				}
			}
		})
	}
}

func getReconcileNode() *ReconcileNode {
	eventRecord := record.NewFakeRecorder(100)
	recon := &ReconcileNode{
		cloud:       getMockCloudProvider(),
		client:      getFakeKubeClient(),
		record:      eventRecord,
		requestChan: make(chan *v1.Node, 10),
	}

	return recon
}

func TestReconcileNode(t *testing.T) {
	t.Run("node with cloud taint", func(t *testing.T) {
		recon := getReconcileNode()

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{Name: NodeName},
		}

		result, err := recon.Reconcile(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, reconcile.Result{}, result)

		select {
		case n := <-recon.requestChan:
			assert.Equal(t, NodeName, n.Name)
		default:
			t.Error("Expected node to be added to request channel")
		}
	})

	t.Run("node without cloud taint", func(t *testing.T) {
		recon := getReconcileNode()
		node := &v1.Node{}
		err := recon.client.Get(context.TODO(), types.NamespacedName{Name: NodeName}, node)
		assert.NoError(t, err)

		node.Spec.Taints = []v1.Taint{}
		err = recon.client.Update(context.TODO(), node)
		assert.NoError(t, err)

		req := reconcile.Request{
			NamespacedName: util.NamespacedName(node),
		}

		// Test reconcile without cloud taint
		result, err := recon.Reconcile(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, reconcile.Result{}, result)

		// Verify nothing was added to request channel
		select {
		case n := <-recon.requestChan:
			t.Errorf("Expected no node to be added to request channel, but got %s", n.Name)
		default:
			// This is expected
		}
	})

	t.Run("nonexistent node", func(t *testing.T) {
		recon := getReconcileNode()

		req := reconcile.Request{
			NamespacedName: client.ObjectKey{
				Name: "non-existent-node",
			},
		}

		result, err := recon.Reconcile(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, reconcile.Result{}, result)

		select {
		case n := <-recon.requestChan:
			t.Errorf("Expected no node to be added to request channel, but got %s", n.Name)
		default:
			// This is expected
		}
	})
}

func TestBatchWorker(t *testing.T) {
	oldBatchSize := ctrlCfg.ControllerCFG.NodeReconcileBatchSize
	oldAggregation := ctrlCfg.ControllerCFG.NodeEventAggregationWaitSeconds
	defer func() {
		time.Sleep(500 * time.Millisecond)
		ctrlCfg.ControllerCFG.NodeReconcileBatchSize = oldBatchSize
		ctrlCfg.ControllerCFG.NodeEventAggregationWaitSeconds = oldAggregation
	}()
	ctrlCfg.ControllerCFG.NodeReconcileBatchSize = 10
	ctrlCfg.ControllerCFG.NodeEventAggregationWaitSeconds = 0

	t.Run("batch worker process nodes", func(t *testing.T) {
		eventRecord := record.NewFakeRecorder(100)
		recon := &ReconcileNode{
			cloud:       getMockCloudProvider(),
			client:      getFakeKubeClient(),
			record:      eventRecord,
			requestChan: make(chan *v1.Node, 10),
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go recon.batchWorker(ctx, 0)

		nodeList := &v1.NodeList{}
		err := recon.client.List(context.TODO(), nodeList)
		assert.NoError(t, err)
		assert.NotEmpty(t, nodeList.Items)

		nodeCopy := nodeList.Items[0].DeepCopy()
		recon.requestChan <- nodeCopy

		time.Sleep(2 * time.Second)

		updatedNode := &v1.Node{}
		err = recon.client.Get(context.TODO(), types.NamespacedName{Name: nodeCopy.Name}, updatedNode)
		assert.NoError(t, err)

		cloudTaint := findCloudTaint(updatedNode.Spec.Taints)
		assert.Nil(t, cloudTaint)
	})

	t.Run("batch worker handle multiple nodes", func(t *testing.T) {
		eventRecord := record.NewFakeRecorder(100)
		recon := &ReconcileNode{
			cloud:       getMockCloudProvider(),
			client:      getFakeKubeClient(),
			record:      eventRecord,
			requestChan: make(chan *v1.Node, 10),
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go recon.batchWorker(ctx, 1)

		nodeList := &v1.NodeList{}
		err := recon.client.List(context.TODO(), nodeList)
		assert.NoError(t, err)
		assert.NotEmpty(t, nodeList.Items)

		n1, n2 := nodeList.Items[0].DeepCopy(), nodeList.Items[0].DeepCopy()
		recon.requestChan <- n1
		recon.requestChan <- n2

		time.Sleep(2 * time.Second)
	})

	t.Run("batch worker context cancellation", func(t *testing.T) {
		eventRecord := record.NewFakeRecorder(100)
		recon := &ReconcileNode{
			cloud:       getMockCloudProvider(),
			client:      getFakeKubeClient(),
			record:      eventRecord,
			requestChan: make(chan *v1.Node, 10),
		}

		ctx, cancel := context.WithCancel(context.Background())
		go recon.batchWorker(ctx, 2)
		cancel()
		time.Sleep(1 * time.Second)
	})

	t.Run("batch worker syncNode error", func(t *testing.T) {
		eventRecord := record.NewFakeRecorder(100)
		recon := &ReconcileNode{
			cloud:       getMockCloudProvider(),
			client:      getFakeKubeClient(),
			record:      eventRecord,
			requestChan: make(chan *v1.Node, 10),
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go recon.batchWorker(ctx, 3)
		nodeWithListError := v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-list-instances-error"},
			Spec: v1.NodeSpec{
				ProviderID: "cn-hangzhou.ecs-list-instances-error",
				Taints: []v1.Taint{
					{Key: api.TaintExternalCloudProvider, Value: "true"},
				},
			},
		}
		nodeCopy := nodeWithListError.DeepCopy()
		recon.requestChan <- nodeCopy
		time.Sleep(2 * time.Second)
	})

	t.Run("batch worker duplicated request", func(t *testing.T) {
		eventRecord := record.NewFakeRecorder(100)
		recon := &ReconcileNode{
			cloud:       getMockCloudProvider(),
			client:      getFakeKubeClient(),
			record:      eventRecord,
			requestChan: make(chan *v1.Node, 10),
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go recon.batchWorker(ctx, 4)
		nodeList := &v1.NodeList{}
		err := recon.client.List(context.TODO(), nodeList)
		assert.NoError(t, err)
		assert.NotEmpty(t, nodeList.Items)
		base := nodeList.Items[0]
		n1, n2, n3 := base.DeepCopy(), base.DeepCopy(), base.DeepCopy()
		recon.requestChan <- n1
		recon.requestChan <- n2
		recon.requestChan <- n3
		time.Sleep(2 * time.Second)
	})

	time.Sleep(2 * time.Second)
}

func TestSyncNode(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		recon := getReconcileNode()
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeName,
			},
		}
		err := getFakeKubeClient().Get(context.TODO(), util.NamespacedName(node), node)
		if err != nil {
			t.Error(err)
		}

		if err = recon.syncNode([]v1.Node{*node}, true); err != nil {
			t.Error(err)
		}

		updatedNode := &v1.Node{}
		if err := recon.client.Get(context.TODO(), util.NamespacedName(node), updatedNode); err != nil {
			t.Error(err)
		}

		if len(updatedNode.Spec.Taints) != 0 {
			t.Errorf("remove taint error, taints: %+v", updatedNode.Spec.Taints)
		}
		if instanceType, ok := updatedNode.Labels[v1.LabelInstanceType]; !ok || instanceType != vmock.InstanceType {
			t.Errorf("node label LabelInstanceType not equal, expect %s, got %s", vmock.InstanceType, instanceType)
		}
		if zone, ok := updatedNode.Labels[v1.LabelTopologyZone]; !ok || zone != vmock.ZoneID {
			t.Errorf("node label LabelTopologyZone not equal, expect %s, got %s", vmock.ZoneID, zone)
		}
		if region, ok := updatedNode.Labels[v1.LabelTopologyRegion]; !ok || region != vmock.RegionID {
			t.Errorf("node label LabelTopologyRegion not equal, expect %s, got %s", vmock.RegionID, region)
		}
		if nodePoolID, ok := updatedNode.Labels[LabelNodePoolID]; !ok || nodePoolID != vmock.NodePoolID {
			t.Errorf("node label LabelNodePoolID not equal, expect %s, got %s", vmock.NodePoolID, nodePoolID)
		}
		if instanceChargeType, ok := updatedNode.Labels[LabelInstanceChargeType]; !ok || instanceChargeType != vmock.InstanceChargeType {
			t.Errorf("node label LabelInstanceChargeType not equal, expect %s, got %s", vmock.InstanceChargeType, instanceChargeType)
		}
		if spotStrategy, ok := updatedNode.Labels[LabelSpotStrategy]; !ok || spotStrategy != vmock.SpotStrategy {
			t.Errorf("node label LabelSpotStrategy not equal, expect %s, got %s", vmock.SpotStrategy, spotStrategy)
		}
		if len(updatedNode.Status.Addresses) == 0 {
			t.Error("node address is empty")
		}
		for _, addr := range updatedNode.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				if addr.Address != vmock.InstanceIP {
					t.Errorf("node internal ip address not equal, expect %s, got %s", vmock.InstanceIP, addr.Address)
				}
			}
		}
	})

	t.Run("unrecognized node type skipped", func(t *testing.T) {
		recon := getReconcileNode()
		unknownNode := v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "unknown-node"},
			Spec:       v1.NodeSpec{ProviderID: ""},
		}
		err := recon.syncNode([]v1.Node{unknownNode}, true)
		assert.NoError(t, err)
	})

	t.Run("node without cloud taint", func(t *testing.T) {
		recon := getReconcileNode()
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-without-cloud-taint",
			},
			Spec: v1.NodeSpec{
				PodCIDR:    "10.96.0.64/26",
				ProviderID: "cn-hangzhou.ecs-id-1",
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

		err := recon.client.Create(context.TODO(), node)
		assert.NoError(t, err)

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err)

		n := &v1.Node{}
		err = recon.client.Get(context.TODO(), util.NamespacedName(node), n)
		assert.NoError(t, err)
	})

	t.Run("not exists on cloud with status unknown", func(t *testing.T) {
		recon := getReconcileNode()
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-not-exists",
			},
			Spec: v1.NodeSpec{
				PodCIDR:    "10.96.0.64/26",
				ProviderID: "cn-hangzhou.not-exists-ecs-id",
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

		err := recon.client.Create(context.TODO(), node)
		assert.NoError(t, err)

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err)

		// wait for node to be deleted
		time.Sleep(100 * time.Millisecond)

		n := &v1.Node{}
		err = recon.client.Get(context.TODO(), util.NamespacedName(node), n)
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
	})

	t.Run("not exists on cloud with status ready", func(t *testing.T) {
		recon := getReconcileNode()
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-not-exists-1",
			},
			Spec: v1.NodeSpec{
				PodCIDR:    "10.96.0.64/26",
				ProviderID: "cn-hangzhou.not-exists-ecs-id-1",
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

		err := recon.client.Create(context.TODO(), node)
		assert.NoError(t, err)

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err)

		n := &v1.Node{}
		err = recon.client.Get(context.TODO(), util.NamespacedName(node), n)
		assert.NoError(t, err)
	})

	t.Run("node with source dest check disabled", func(t *testing.T) {
		// Save original config value
		originalValue := ctrlCfg.ControllerCFG.SkipDisableSourceDestCheck
		ctrlCfg.ControllerCFG.SkipDisableSourceDestCheck = true
		defer func() {
			ctrlCfg.ControllerCFG.SkipDisableSourceDestCheck = originalValue
		}()

		recon := getReconcileNode()
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: NodeName,
			},
		}
		err := getFakeKubeClient().Get(context.TODO(), util.NamespacedName(node), node)
		assert.NoError(t, err)

		// Ensure node has cloud taint to trigger source dest check logic
		node.Spec.Taints = append(node.Spec.Taints, v1.Taint{
			Key:   api.TaintExternalCloudProvider,
			Value: "true",
		})

		err = recon.client.Update(context.TODO(), node)
		assert.NoError(t, err)

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err)

		// Verify node was processed correctly
		updatedNode := &v1.Node{}
		err = recon.client.Get(context.TODO(), util.NamespacedName(node), updatedNode)
		assert.NoError(t, err)
	})

	t.Run("multiple nodes processing", func(t *testing.T) {
		recon := getReconcileNode()

		// Get existing node
		existingNode := &v1.Node{}
		err := getFakeKubeClient().Get(context.TODO(), types.NamespacedName{Name: NodeName}, existingNode)
		assert.NoError(t, err)

		// Create additional node
		additionalNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "additional-node",
			},
			Spec: v1.NodeSpec{
				ProviderID: "cn-hangzhou.ecs-additional",
				Taints: []v1.Taint{
					{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
				},
			},
		}
		err = recon.client.Create(context.TODO(), additionalNode)
		assert.NoError(t, err)

		nodes := []v1.Node{*existingNode, *additionalNode}
		err = recon.syncNode(nodes, true)
		assert.NoError(t, err)

		// Verify both nodes were processed
		for _, node := range nodes {
			updatedNode := &v1.Node{}
			err = recon.client.Get(context.TODO(), util.NamespacedName(&node), updatedNode)
			assert.NoError(t, err)

			// Cloud taint should be removed
			cloudTaint := findCloudTaint(updatedNode.Spec.Taints)
			assert.Nil(t, cloudTaint)
		}
	})

	t.Run("node with user specified ip", func(t *testing.T) {
		recon := getReconcileNode()

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-with-user-ip",
				Annotations: map[string]string{
					"alpha.kubernetes.io/provided-node-ip": vmock.InstanceIP,
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "cn-hangzhou.ecs-user-ip",
				Taints: []v1.Taint{
					{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
				},
			},
		}
		err := recon.client.Create(context.TODO(), node)
		assert.NoError(t, err)

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err)

		updatedNode := &v1.Node{}
		err = recon.client.Get(context.TODO(), util.NamespacedName(node), updatedNode)
		assert.NoError(t, err)

		// Should have only the user-specified IP address
		assert.Len(t, updatedNode.Status.Addresses, 1)
		assert.Equal(t, v1.NodeInternalIP, updatedNode.Status.Addresses[0].Type)
		assert.Equal(t, vmock.InstanceIP, updatedNode.Status.Addresses[0].Address)
	})

	t.Run("node with invalid user specified ip", func(t *testing.T) {
		recon := getReconcileNode()

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-with-invalid-user-ip",
				Annotations: map[string]string{
					"alpha.kubernetes.io/provided-node-ip": "1.2.3.4", // IP that doesn't match cloud provider
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "cn-hangzhou.ecs-invalid-user-ip",
				Taints: []v1.Taint{
					{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
				},
			},
		}
		err := recon.client.Create(context.TODO(), node)
		assert.NoError(t, err)

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err)

		updatedNode := &v1.Node{}
		err = recon.client.Get(context.TODO(), util.NamespacedName(node), updatedNode)
		assert.NoError(t, err)
	})

	t.Run("patch node status error", func(t *testing.T) {
		recon := getReconcileNode()

		// Create a node
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "patch-error-node",
			},
			Spec: v1.NodeSpec{
				ProviderID: "cn-hangzhou.ecs-patch-error",
				Taints: []v1.Taint{
					{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
				},
			},
		}
		err := recon.client.Create(context.TODO(), node)
		assert.NoError(t, err)

		// Use a client that will fail on patch operations
		mockClient := &MockClientWithPatchError{
			Client: recon.client,
		}
		recon.client = mockClient

		err = recon.syncNode([]v1.Node{*node}, true)
		assert.NoError(t, err) // syncNode doesn't return error even if patch fails

		// But the node should still exist
		updatedNode := &v1.Node{}
		err = mockClient.Get(context.TODO(), util.NamespacedName(node), updatedNode)
		assert.NoError(t, err)
	})
}

func TestDisableNetworkInterfaceSourceDestCheck(t *testing.T) {
	t.Run("with API error", func(t *testing.T) {
		enis := []eniInfo{
			{
				ENI: "eni-error",
				NodeRef: &v1.ObjectReference{
					Kind: "Node",
					Name: "test",
				},
			},
		}

		recon := getReconcileNode()
		failed, err := recon.disableNetworkInterfaceSourceDestCheck(enis)
		assert.NoError(t, err)
		assert.Len(t, failed, 1)
	})
}

// MockClientWithPatchError implements client.Client but returns error for Patch operations
type MockClientWithPatchError struct {
	client.Client
}

func (m *MockClientWithPatchError) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	// Return error for patch operations
	return fmt.Errorf("simulated patch error")
}

func (m *MockClientWithPatchError) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return m.Client.Get(ctx, key, obj, opts...)
}

func TestIsProvidedAddrExist(t *testing.T) {
	nodeWithProvidedAddr := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"alpha.kubernetes.io/provided-node-ip": "192.168.1.100",
			},
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
				{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
			},
		},
	}

	nodeWithoutProvidedAddr := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
			},
		},
	}

	addressesWithMatch := []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
		{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
	}

	addressesWithoutMatch := []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: "192.168.1.101"},
		{Type: v1.NodeExternalIP, Address: "1.2.3.5"},
	}

	tests := []struct {
		name         string
		node         *v1.Node
		addresses    []v1.NodeAddress
		expectedOK   bool
		expectedAddr *v1.NodeAddress
	}{
		{
			name:         "provided addr exists and matches",
			node:         nodeWithProvidedAddr,
			addresses:    addressesWithMatch,
			expectedOK:   true,
			expectedAddr: &v1.NodeAddress{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
		},
		{
			name:         "provided addr exists but no match",
			node:         nodeWithProvidedAddr,
			addresses:    addressesWithoutMatch,
			expectedOK:   true,
			expectedAddr: nil,
		},
		{
			name:         "no provided addr",
			node:         nodeWithoutProvidedAddr,
			addresses:    addressesWithMatch,
			expectedOK:   false,
			expectedAddr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, ok := isProvidedAddrExist(tt.node, tt.addresses)
			assert.Equal(t, tt.expectedOK, ok)
			if tt.expectedAddr == nil {
				assert.Nil(t, addr)
			} else {
				assert.NotNil(t, addr)
				assert.Equal(t, tt.expectedAddr.Type, addr.Type)
				assert.Equal(t, tt.expectedAddr.Address, addr.Address)
			}
		})
	}
}

func TestSetHostnameAddress(t *testing.T) {
	nodeWithoutHostname := &v1.Node{
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
			},
		},
	}

	nodeWithHostname := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"alpha.kubernetes.io/provided-node-ip": "192.168.1.100",
			},
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
				{Type: v1.NodeHostName, Address: "test-hostname"},
			},
		},
	}

	addressWithHostname := []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
		{Type: v1.NodeHostName, Address: "cloud-hostname"},
	}

	addressWithoutHostname := []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
		{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
	}

	tests := []struct {
		name             string
		node             *v1.Node
		addresses        []v1.NodeAddress
		expectedCount    int
		expectedHostname string
		hasHostname      bool
	}{
		{
			name:             "add hostname when not present",
			node:             nodeWithHostname,
			addresses:        addressWithoutHostname,
			expectedCount:    3,
			expectedHostname: "test-hostname",
			hasHostname:      true,
		},
		{
			name:             "keep existing hostname",
			node:             nodeWithHostname,
			addresses:        addressWithHostname,
			expectedCount:    2,
			expectedHostname: "cloud-hostname",
			hasHostname:      true,
		},
		{
			name:          "no hostname",
			node:          nodeWithoutHostname,
			addresses:     addressWithoutHostname,
			expectedCount: 2,
			hasHostname:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setHostnameAddress(tt.node, tt.addresses)
			assert.Equal(t, tt.expectedCount, len(result))

			hasHostname := false
			for _, addr := range result {
				if addr.Type == v1.NodeHostName {
					hasHostname = true
					if tt.expectedHostname != "" {
						assert.Equal(t, tt.expectedHostname, addr.Address)
					}
					break
				}
			}
			assert.Equal(t, tt.hasHostname, hasHostname)
		})
	}
}

func TestSetFields(t *testing.T) {
	cloudNode := &prvd.NodeAttribute{
		InstanceID:   "i-test",
		InstanceType: "ecs.c1m1.large",
		Addresses: []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
		},
		Zone:   "cn-hangzhou-a",
		Region: "cn-hangzhou",
	}

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: v1.NodeSpec{
			Taints: []v1.Taint{
				{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
			},
		},
	}

	// Test with removeCloudTaint=true
	setFields(node, cloudNode, true, true)

	// Check labels
	assert.Equal(t, "ecs.c1m1.large", node.Labels[v1.LabelInstanceTypeStable])
	assert.Equal(t, "cn-hangzhou-a", node.Labels[v1.LabelTopologyZone])
	assert.Equal(t, "cn-hangzhou", node.Labels[v1.LabelTopologyRegion])

	// Check that cloud taint was removed
	hasCloudTaint := false
	for _, taint := range node.Spec.Taints {
		if taint.Key == api.TaintExternalCloudProvider {
			hasCloudTaint = true
			break
		}
	}
	assert.False(t, hasCloudTaint)

	// Reset node for second test
	node.Spec.Taints = []v1.Taint{
		{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
	}

	// Test with removeCloudTaint=false
	setFields(node, cloudNode, true, false)

	// Check that cloud taint was not removed
	hasCloudTaint = false
	for _, taint := range node.Spec.Taints {
		if taint.Key == api.TaintExternalCloudProvider {
			hasCloudTaint = true
			break
		}
	}
	assert.True(t, hasCloudTaint)
}

func TestAdd(t *testing.T) {
	mgr, err := manager.New(&rest.Config{}, manager.Options{})
	assert.NoError(t, err)

	assert.NotPanics(t, func() {
		err := Add(mgr, &shared.SharedContext{})
		assert.NoError(t, err)
	})
}

func TestPeriodicalSync(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		recon := getReconcileNode()
		recon.statusFrequency = 100 * time.Millisecond
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		recon.PeriodicalSync(ctx)
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("NodeList error", func(t *testing.T) {
		recon := getReconcileNode()
		recon.statusFrequency = 50 * time.Millisecond
		recon.client = &MockClientWithListError{
			Client: recon.client,
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		recon.PeriodicalSync(ctx)
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("context cancellation", func(t *testing.T) {
		recon := getReconcileNode()
		recon.statusFrequency = 50 * time.Millisecond
		ctx, cancel := context.WithCancel(context.Background())

		recon.PeriodicalSync(ctx)
		time.Sleep(50 * time.Millisecond)
		cancel()
		time.Sleep(100 * time.Millisecond)
	})
}

type MockClientWithListError struct {
	client.Client
}

func (m *MockClientWithListError) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return fmt.Errorf("simulated list error")
}

func getMockCloudProvider() prvd.Provider {
	return vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
}

func getFakeKubeClient() client.Client {
	// Node
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: NodeName,
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.64/26",
					ProviderID: "cn-hangzhou.ecs-id",
					Taints: []v1.Taint{
						{
							Key:   api.TaintExternalCloudProvider,
							Value: "true",
						},
					},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
						},
						{
							Reason: string(v1.NodeNetworkUnavailable),
							Status: v1.ConditionFalse,
						},
					},
				},
			},
		},
	}

	objs := []runtime.Object{nodeList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}
