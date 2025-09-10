package node

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider/api"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNodeList(t *testing.T) {
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "normal-node",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "exclude-label",
					Labels: map[string]string{
						helper.LabelNodeExcludeNode: "true",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "legacy-exclude-label",
					Labels: map[string]string{
						helper.LabelNodeExcludeNodeDeprecated: "true",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "empty-provider-id",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "hybrid-node",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "ack-hybrid://123456",
				},
			},
		},
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(nodeList).Build()

	nodes, err := NodeList(client, false)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(nodes.Items))
	assert.Equal(t, "normal-node", nodes.Items[0].Name)
}

func TestGetNodeType(t *testing.T) {
	cases := []struct {
		name string
		node *v1.Node
		want NodeType
		ok   bool
	}{
		{
			name: "normal ecs node",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "normal-node",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			want: NodeTypeECS,
			ok:   true,
		},
		{
			name: "lingjun node",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node",
					Labels: map[string]string{
						LabelLingJunWorker: "true",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-123456",
				},
			},
			want: NodeTypeLingJun,
			ok:   true,
		},
		{
			name: "unknown node",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unknown-node",
				},
				Spec: v1.NodeSpec{
					ProviderID: "unknown-provider-id",
				},
			},
			ok: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, ok := getNodeType(c.node)
			assert.Equal(t, c.ok, ok)
			if c.ok {
				assert.Equal(t, c.want, actual)
			}
		})
	}
}

func TestSetNetworkUnavailable(t *testing.T) {
	t.Run("node without NodeNetworkUnavailable condition", func(t *testing.T) {
		// Test case 1: Node without NodeNetworkUnavailable condition
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{},
			},
		}

		setNetworkUnavailable(node)

		condition := getNodeCondition(node, v1.NodeNetworkUnavailable)
		assert.NotNil(t, condition)
		assert.Equal(t, v1.ConditionTrue, condition.Status)
		assert.Equal(t, "NoRouteCreated", condition.Reason)
	})

	t.Run("node with NodeNetworkUnavailable already set to False", func(t *testing.T) {
		// Test case 2: Node with NodeNetworkUnavailable already set to False
		node2 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-2",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeNetworkUnavailable,
						Status: v1.ConditionFalse,
					},
				},
			},
		}

		setNetworkUnavailable(node2)

		condition2 := getNodeCondition(node2, v1.NodeNetworkUnavailable)
		assert.NotNil(t, condition2)
		assert.Equal(t, v1.ConditionFalse, condition2.Status)
	})
}

func TestRemoveCloudTaints(t *testing.T) {
	// Test case 1: Node with cloud taint
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: v1.NodeSpec{
			Taints: []v1.Taint{
				{
					Key:    "node.cloudprovider.kubernetes.io/uninitialized",
					Value:  "true",
					Effect: v1.TaintEffectNoSchedule,
				},
				{
					Key:    "other-taint",
					Value:  "value",
					Effect: v1.TaintEffectNoExecute,
				},
			},
		},
	}

	removeCloudTaints(node)

	assert.Equal(t, 1, len(node.Spec.Taints))
	assert.Equal(t, "other-taint", node.Spec.Taints[0].Key)

	// Test case 2: Node without cloud taint
	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-2",
		},
		Spec: v1.NodeSpec{
			Taints: []v1.Taint{
				{
					Key:    "other-taint",
					Value:  "value",
					Effect: v1.TaintEffectNoExecute,
				},
			},
		},
	}

	removeCloudTaints(node2)

	assert.Equal(t, 1, len(node2.Spec.Taints))
	assert.Equal(t, "other-taint", node2.Spec.Taints[0].Key)
}

func TestFindCloudTaint(t *testing.T) {
	tests := []struct {
		name     string
		taints   []v1.Taint
		expected *v1.Taint
	}{
		{
			name: "has cloud taint",
			taints: []v1.Taint{
				{Key: v1.TaintNodeNetworkUnavailable, Value: "test", Effect: v1.TaintEffectNoSchedule},
				{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
			},
			expected: &v1.Taint{Key: api.TaintExternalCloudProvider, Value: "true", Effect: v1.TaintEffectNoSchedule},
		},
		{
			name: "no cloud taint",
			taints: []v1.Taint{
				{Key: v1.TaintNodeNetworkUnavailable, Value: "test", Effect: v1.TaintEffectNoSchedule},
			},
			expected: nil,
		},
		{
			name:     "empty taints",
			taints:   []v1.Taint{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCloudTaint(tt.taints)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Key, result.Key)
				assert.Equal(t, tt.expected.Value, result.Value)
				assert.Equal(t, tt.expected.Effect, result.Effect)
			}
		})
	}
}

func TestExcludeTaintFromList(t *testing.T) {
	taints := []v1.Taint{
		{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: v1.TaintEffectNoSchedule,
		},
		{
			Key:    "other-taint",
			Value:  "value",
			Effect: v1.TaintEffectNoExecute,
		},
	}

	toExclude := v1.Taint{
		Key:    "node.cloudprovider.kubernetes.io/uninitialized",
		Value:  "true",
		Effect: v1.TaintEffectNoSchedule,
	}

	result := excludeTaintFromList(taints, toExclude)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "other-taint", result[0].Key)
}

func TestFindCloudECSById(t *testing.T) {
	mockECS := &vmock.MockECS{}
	nodeWithProviderID := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-1",
		},
		Spec: v1.NodeSpec{
			ProviderID: "cn-hangzhou.i-123456",
		},
	}

	ins, err := findCloudECSById(mockECS, nodeWithProviderID)
	assert.NoError(t, err)
	assert.NotNil(t, ins)
	assert.Equal(t, "i-123456", ins.InstanceID)
	assert.Equal(t, vmock.InstanceType, ins.InstanceType)
	assert.Equal(t, vmock.ZoneID, ins.Zone)
	assert.Equal(t, vmock.RegionID, ins.Region)
}

func TestFindCloudECSById_ListInstancesError(t *testing.T) {
	mockECS := &vmock.MockECS{}
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
		Spec:       v1.NodeSpec{ProviderID: "cn-hangzhou.ecs-list-instances-error"},
	}
	ins, err := findCloudECSById(mockECS, node)
	assert.Error(t, err)
	assert.Nil(t, ins)
	assert.Contains(t, err.Error(), "cloud instance api fail")
}

func TestFindCloudECSById_NotFound(t *testing.T) {
	mockECS := &vmock.MockECS{}
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
		Spec:       v1.NodeSpec{ProviderID: "cn-hangzhou.not-exists"},
	}
	ins, err := findCloudECSById(mockECS, node)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, ins)
}

func TestFindCloudECSByIp(t *testing.T) {
	mockECS := &vmock.MockECS{}
	t.Run("node with single ip", func(t *testing.T) {
		nodeWithSingleIP := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-1",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    v1.NodeInternalIP,
						Address: "192.168.1.100",
					},
				},
			},
		}

		ins, err := findCloudECSByIp(mockECS, nodeWithSingleIP)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, ins)
	})

	t.Run("node without ip", func(t *testing.T) {
		nodeWithoutIP := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-2",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{},
			},
		}

		ins, err := findCloudECSByIp(mockECS, nodeWithoutIP)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, ins)
	})

	t.Run("node with multiple ips", func(t *testing.T) {
		nodeWithMultipleIPs := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-3",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    v1.NodeInternalIP,
						Address: "192.168.1.100",
					},
					{
						Type:    v1.NodeInternalIP,
						Address: "192.168.1.101",
					},
				},
			},
		}

		ins, err := findCloudECSByIp(mockECS, nodeWithMultipleIPs)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, ins)
	})
}

func TestFindCloudECS(t *testing.T) {
	mockECS := &vmock.MockECS{}
	t.Run("node with provider id", func(t *testing.T) {
		nodeWithProviderID := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-1",
			},
			Spec: v1.NodeSpec{
				ProviderID: "cn-hangzhou.i-123456",
			},
		}

		ins, err := findCloudECS(mockECS, nodeWithProviderID)
		assert.NoError(t, err)
		assert.NotNil(t, ins)
	})

	t.Run("node without provider id", func(t *testing.T) {
		nodeWithoutProviderID := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-2",
			},
			Spec: v1.NodeSpec{
				ProviderID: "",
			},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{
						Type:    v1.NodeInternalIP,
						Address: "192.168.1.100",
					},
				},
			},
		}

		ins, err := findCloudECS(mockECS, nodeWithoutProviderID)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, ins)
	})

}

func TestBatchOperate(t *testing.T) {
	t.Run("normal case with less than MAX_BATCH_NUM nodes", func(t *testing.T) {
		nodes := make([]v1.Node, 50)
		for i := 0; i < 50; i++ {
			nodes[i] = v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-" + string(rune(i)),
					Labels: map[string]string{},
				},
			}
		}

		count := 0
		batchFunc := func(batch []v1.Node, last bool) error {
			count += len(batch)
			return nil
		}

		err := batchOperate(nodes, batchFunc)
		assert.NoError(t, err)
		assert.Equal(t, 50, count)
	})

	t.Run("batch case with more than MAX_BATCH_NUM nodes", func(t *testing.T) {
		nodes := make([]v1.Node, 150)
		for i := 0; i < 150; i++ {
			nodes[i] = v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-" + string(rune(i)),
					Labels: map[string]string{},
				},
			}
		}

		count := 0
		batchSizes := []int{}
		batchFunc := func(batch []v1.Node, last bool) error {
			count += len(batch)
			batchSizes = append(batchSizes, len(batch))
			return nil
		}

		err := batchOperate(nodes, batchFunc)
		assert.NoError(t, err)
		assert.Equal(t, 150, count)
		// Should have two batches: 100 and 50
		assert.Equal(t, []int{100, 50}, batchSizes)
	})

	t.Run("error case", func(t *testing.T) {
		nodes := make([]v1.Node, 150)
		for i := 0; i < 150; i++ {
			nodes[i] = v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-" + string(rune(i)),
					Labels: map[string]string{},
				},
			}
		}

		callCount := 0
		batchFunc := func(batch []v1.Node, last bool) error {
			callCount++
			// Return error on second call to simulate failure in processing second batch
			if callCount == 2 {
				return errors.New("batch process failed")
			}
			return nil
		}

		err := batchOperate(nodes, batchFunc)
		assert.Error(t, err)
		assert.Equal(t, "batch process failed", err.Error())
	})

	t.Run("empty nodes", func(t *testing.T) {
		nodes := []v1.Node{}

		count := 0
		batchFunc := func(batch []v1.Node, last bool) error {
			count += len(batch)
			return nil
		}

		err := batchOperate(nodes, batchFunc)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestNodeConditionReady(t *testing.T) {
	t.Run("node with ready condition immediately", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-ready",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
						Reason: "KubeletReady",
					},
				},
			},
		}
		client := fake.NewClientBuilder().WithRuntimeObjects(node).Build()

		condition := nodeConditionReady(client, node)
		assert.NotNil(t, condition)
		assert.Equal(t, v1.NodeReady, condition.Type)
		assert.Equal(t, v1.ConditionTrue, condition.Status)
	})

	t.Run("node without ready condition initially but gets it after retry", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-retry",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{},
			},
		}
		// Create a node that will have ready condition after first Get
		nodeWithReady := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-retry",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
						Reason: "KubeletReady",
					},
				},
			},
		}
		client := fake.NewClientBuilder().WithRuntimeObjects(nodeWithReady).Build()

		condition := nodeConditionReady(client, node)
		// After retry, should find the condition
		assert.NotNil(t, condition)
		assert.Equal(t, v1.NodeReady, condition.Type)
	})

	t.Run("node deleted during retry", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-deleted",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{},
			},
		}
		// Client without the node (simulating deletion)
		client := fake.NewClientBuilder().Build()

		condition := nodeConditionReady(client, node)
		assert.Nil(t, condition)
	})

	t.Run("node never gets ready condition", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-never-ready",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeMemoryPressure,
						Status: v1.ConditionFalse,
					},
				},
			},
		}
		client := fake.NewClientBuilder().WithRuntimeObjects(node).Build()

		condition := nodeConditionReady(client, node)
		// After all retries, should return nil
		assert.Nil(t, condition)
	})

	t.Run("node with ready condition false", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-not-ready",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionFalse,
						Reason: "KubeletNotReady",
					},
				},
			},
		}
		client := fake.NewClientBuilder().WithRuntimeObjects(node).Build()

		condition := nodeConditionReady(client, node)
		assert.NotNil(t, condition)
		assert.Equal(t, v1.NodeReady, condition.Type)
		assert.Equal(t, v1.ConditionFalse, condition.Status)
	})
}

func getNodeCondition(node *v1.Node, conditionType v1.NodeConditionType) *v1.NodeCondition {
	for _, condition := range node.Status.Conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}
