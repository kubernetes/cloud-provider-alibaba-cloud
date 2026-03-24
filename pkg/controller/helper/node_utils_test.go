package helper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPatchM(t *testing.T) {
	t.Run("patch status", func(t *testing.T) {
		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*v1.Node)
			condition, ok := FindCondition(nins.Status.Conditions, v1.NodeNetworkUnavailable)
			condition.Type = v1.NodeNetworkUnavailable
			condition.Status = v1.ConditionFalse
			condition.Reason = "RouteCreated"
			condition.Message = "RouteController created a route"

			if !ok {
				nins.Status.Conditions = append(nins.Status.Conditions, *condition)
			}
			return nins, nil
		}

		mclient := getFakeKubeClient()
		node1 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"app": "nginx"},
			},
		}

		err := PatchM(mclient, node1, diff, PatchStatus)
		if err != nil {
			t.Error(err)
		}
		node1Patched := &v1.Node{}
		_ = mclient.Get(context.TODO(), client.ObjectKey{
			Name: node1.Name,
		}, node1Patched)

		for _, con := range node1Patched.Status.Conditions {
			if con.Type == v1.NodeNetworkUnavailable {
				assert.Equal(t, v1.ConditionFalse, con.Status)
				return
			}
		}
		t.Error("Fail to patch node network status")
	})

	t.Run("patch all", func(t *testing.T) {
		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*v1.Node)

			// Modify spec part
			nins.Spec.Unschedulable = true

			// Modify status part
			nins.Status.Conditions[0].Status = v1.ConditionFalse
			nins.Status.Conditions[0].Reason = "TestPatchAll"

			return nins, nil
		}

		mclient := getFakeKubeClient()
		node1 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"app": "nginx"},
			},
			Spec: v1.NodeSpec{
				Unschedulable: false,
			},
		}

		err := PatchM(mclient, node1, diff, PatchAll)
		assert.NoError(t, err)

		// Verify updated node
		node1Patched := &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{
			Name: node1.Name,
		}, node1Patched)
		assert.NoError(t, err)

		// Verify spec is updated
		assert.True(t, node1Patched.Spec.Unschedulable)

		// Verify status is updated
		assert.Equal(t, v1.ConditionFalse, node1Patched.Status.Conditions[0].Status)
		assert.Equal(t, "TestPatchAll", node1Patched.Status.Conditions[0].Reason)
	})

	t.Run("patch spec only", func(t *testing.T) {
		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*v1.Node)
			nins.Spec.Unschedulable = true
			return nins, nil
		}

		mclient := getFakeKubeClient()
		node1 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"app": "nginx"},
			},
			Spec: v1.NodeSpec{
				Unschedulable: false,
			},
		}

		err := PatchM(mclient, node1, diff, PatchSpec)
		assert.NoError(t, err)

		node1Patched := &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{
			Name: node1.Name,
		}, node1Patched)
		assert.NoError(t, err)
		assert.True(t, node1Patched.Spec.Unschedulable)
	})

	t.Run("empty patch - no changes", func(t *testing.T) {
		diff := func(copy runtime.Object) (client.Object, error) {
			// Return the same object without changes
			return copy.(*v1.Node), nil
		}

		mclient := getFakeKubeClient()
		node1 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"app": "nginx"},
			},
		}

		err := PatchM(mclient, node1, diff, PatchAll)
		assert.NoError(t, err)
	})

	t.Run("getter returns error", func(t *testing.T) {
		diff := func(copy runtime.Object) (client.Object, error) {
			return nil, fmt.Errorf("getter error")
		}

		mclient := getFakeKubeClient()
		node1 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-1",
				Labels: map[string]string{"app": "nginx"},
			},
		}

		err := PatchM(mclient, node1, diff, PatchAll)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get object diff patch")
	})

	t.Run("node not found", func(t *testing.T) {
		diff := func(copy runtime.Object) (client.Object, error) {
			nins := copy.(*v1.Node)
			nins.Spec.Unschedulable = true
			return nins, nil
		}

		mclient := getFakeKubeClient()
		node1 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "non-existent-node",
				Labels: map[string]string{"app": "nginx"},
			},
		}

		err := PatchM(mclient, node1, diff, PatchAll)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get origin object")
	})
}

func TestNodeFromProviderID(t *testing.T) {
	cases := []struct {
		name         string
		providerID   string
		expectError  bool
		expectRegion string
		expectID     string
	}{
		{
			name:         "normal provider id",
			providerID:   "cn-hangzhou.i-123456",
			expectRegion: "cn-hangzhou",
			expectID:     "i-123456",
		},
		{
			name:        "malformed provider id",
			providerID:  "cn-hangzhou123456",
			expectError: true,
		},
		{
			name:         "provider with alicloud://",
			providerID:   "alicloud://cn-hangzhou.i-123456",
			expectRegion: "cn-hangzhou",
			expectID:     "i-123456",
		},
		{
			name:        "malformed provider with alicloud://",
			providerID:  "alicloud://cn-hangzhou123456",
			expectError: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			region, id, err := NodeFromProviderID(tt.providerID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRegion, region)
				assert.Equal(t, tt.expectID, id)
			}
		})
	}
}

func TestIsExcludedNode(t *testing.T) {
	cases := []struct {
		name           string
		node           *v1.Node
		expectExcluded bool
	}{
		{
			name:           "normal node",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}},
			expectExcluded: false,
		},
		{
			name:           "nil node",
			node:           nil,
			expectExcluded: false,
		},
		{
			name:           "node with exclude node label",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{LabelNodeExcludeNode: "true"}}},
			expectExcluded: true,
		},
		{
			name:           "node with deprecated exclude node label",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{LabelNodeExcludeNodeDeprecated: "true"}}},
			expectExcluded: true,
		},
		{
			name:           "hybrid node",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}, Spec: v1.NodeSpec{ProviderID: "ack-hybrid://cn-hangzhou/test"}},
			expectExcluded: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectExcluded, IsExcludedNode(tt.node))
		})
	}
}

func TestIsNodeExcludeFromLoadBalancer(t *testing.T) {
	cases := []struct {
		name           string
		node           *v1.Node
		expectExcluded bool
	}{
		{
			name:           "normal node",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}},
			expectExcluded: false,
		},
		{
			name:           "node with exclude balancer label",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{LabelNodeExcludeBalancer: "true"}}},
			expectExcluded: true,
		},
		{
			name:           "node with deprecated exclude balancer label",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{LabelNodeExcludeBalancerDeprecated: "true"}}},
			expectExcluded: true,
		},
		{
			name:           "node with excluded label",
			node:           &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{LabelNodeExcludeNode: "true"}}},
			expectExcluded: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectExcluded, IsNodeExcludeFromLoadBalancer(tt.node))
		})
	}
}

func TestPatchNodeStatus(t *testing.T) {
	t.Run("update node status normally", func(t *testing.T) {
		mclient := getFakeKubeClient()
		testNode := &v1.Node{}
		err := mclient.Get(context.TODO(), client.ObjectKey{Name: "node-1"}, testNode)
		assert.NoError(t, err)

		getter := func(node *v1.Node) (*v1.Node, error) {
			node.Status.Conditions[0].Status = v1.ConditionFalse
			node.Status.Conditions[0].Reason = "TestReason"
			return node, nil
		}

		err = PatchNodeStatus(mclient, testNode, getter)
		assert.NoError(t, err)

		updatedNode := &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{Name: testNode.Name}, updatedNode)
		assert.NoError(t, err)
		assert.Equal(t, v1.ConditionFalse, updatedNode.Status.Conditions[0].Status)
		assert.Equal(t, "TestReason", updatedNode.Status.Conditions[0].Reason)
	})

	t.Run("update node addresses", func(t *testing.T) {
		mclient := getFakeKubeClient()
		testNode := &v1.Node{}
		err := mclient.Get(context.TODO(), client.ObjectKey{Name: "node-1"}, testNode)
		assert.NoError(t, err)

		getter := func(node *v1.Node) (*v1.Node, error) {
			node.Status.Addresses = append(node.Status.Addresses, v1.NodeAddress{
				Type:    v1.NodeInternalIP,
				Address: "10.0.0.1",
			})
			return node, nil
		}

		err = PatchNodeStatus(mclient, testNode, getter)
		assert.NoError(t, err)

		updatedNode := &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{Name: testNode.Name}, updatedNode)
		assert.NoError(t, err)
		assert.Len(t, updatedNode.Status.Addresses, 1)
		assert.Equal(t, v1.NodeInternalIP, updatedNode.Status.Addresses[0].Type)
		assert.Equal(t, "10.0.0.1", updatedNode.Status.Addresses[0].Address)
	})

	t.Run("add ipv6 address to node", func(t *testing.T) {
		mclient := getFakeKubeClient()
		testNode := &v1.Node{}
		err := mclient.Get(context.TODO(), client.ObjectKey{Name: "node-1"}, testNode)
		assert.NoError(t, err)

		getter := func(node *v1.Node) (*v1.Node, error) {
			node.Status.Addresses = []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: "10.0.0.1",
				},
			}
			return node, nil
		}

		err = PatchNodeStatus(mclient, testNode, getter)
		assert.NoError(t, err)

		updatedNode := &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{Name: testNode.Name}, updatedNode)
		assert.NoError(t, err)
		assert.Len(t, updatedNode.Status.Addresses, 1)
		assert.Equal(t, v1.NodeInternalIP, updatedNode.Status.Addresses[0].Type)
		assert.Equal(t, "10.0.0.1", updatedNode.Status.Addresses[0].Address)

		getter = func(node *v1.Node) (*v1.Node, error) {
			node.Status.Addresses = []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: "10.0.0.1",
				},
				{
					Type:    v1.NodeInternalIP,
					Address: "2001:db8::1",
				},
			}
			return node, nil
		}

		err = PatchNodeStatus(mclient, testNode, getter)
		assert.NoError(t, err)

		updatedNode = &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{Name: testNode.Name}, updatedNode)
		assert.Len(t, updatedNode.Status.Addresses, 2)
		assert.Equal(t, v1.NodeInternalIP, updatedNode.Status.Addresses[1].Type)
		assert.Equal(t, "2001:db8::1", updatedNode.Status.Addresses[1].Address)
	})

	t.Run("getter returns error", func(t *testing.T) {
		mclient := getFakeKubeClient()
		testNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
			},
		}

		getter := func(node *v1.Node) (*v1.Node, error) {
			return nil, fmt.Errorf("getter error")
		}

		err := PatchNodeStatus(mclient, testNode, getter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get object diff patch")
	})

	t.Run("node not found", func(t *testing.T) {
		mclient := getFakeKubeClient()
		testNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "non-existent-node",
			},
		}

		getter := func(node *v1.Node) (*v1.Node, error) {
			node.Status.Conditions[0].Status = v1.ConditionFalse
			return node, nil
		}

		err := PatchNodeStatus(mclient, testNode, getter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get origin object")
	})

	t.Run("no changes to node status", func(t *testing.T) {
		mclient := getFakeKubeClient()
		testNode := &v1.Node{}
		err := mclient.Get(context.TODO(), client.ObjectKey{Name: "node-1"}, testNode)
		assert.NoError(t, err)

		originalStatus := testNode.Status.DeepCopy()
		getter := func(node *v1.Node) (*v1.Node, error) {
			// Return node without changes
			return node, nil
		}

		err = PatchNodeStatus(mclient, testNode, getter)
		assert.NoError(t, err)

		// Verify status hasn't changed
		updatedNode := &v1.Node{}
		err = mclient.Get(context.TODO(), client.ObjectKey{Name: testNode.Name}, updatedNode)
		assert.NoError(t, err)
		assert.Equal(t, originalStatus.Conditions, updatedNode.Status.Conditions)
	})
}

func TestFixupPatchForNodeStatusAddresses(t *testing.T) {
	t.Run("normal patch fixup", func(t *testing.T) {
		patchBytes := []byte(`{"status": {"conditions": [{"type": "Ready", "status": "True"}]}}`)
		addresses := []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
			{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
		}

		result, err := fixupPatchForNodeStatusAddresses(patchBytes, addresses)
		assert.NoError(t, err)
		assert.Contains(t, string(result), `"addresses"`)
		assert.Contains(t, string(result), `"192.168.1.1"`)
		assert.Contains(t, string(result), `"1.2.3.4"`)
		assert.Contains(t, string(result), `"$patch":"replace"`)
	})

	t.Run("empty patch", func(t *testing.T) {
		patchBytes := []byte(`{}`)
		addresses := []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
		}

		result, err := fixupPatchForNodeStatusAddresses(patchBytes, addresses)
		assert.NoError(t, err)
		assert.Contains(t, string(result), `"addresses"`)
		assert.Contains(t, string(result), `"192.168.1.1"`)
		assert.Contains(t, string(result), `"$patch":"replace"`)
	})

	t.Run("invalid json", func(t *testing.T) {
		patchBytes := []byte(`{invalid json}`)
		addresses := []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
		}

		_, err := fixupPatchForNodeStatusAddresses(patchBytes, addresses)
		assert.Error(t, err)
	})
}

func getFakeKubeClient() client.Client {
	// Node
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-1",
					Labels: map[string]string{"app": "nginx"},
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
							Reason: "No Route Created",
							Status: v1.ConditionTrue,
						},
					},
				},
			},
		},
	}

	objs := []runtime.Object{nodeList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}

func TestFindCondition(t *testing.T) {
	conditions := []v1.NodeCondition{
		{
			Type:   v1.NodeReady,
			Status: v1.ConditionTrue,
		},
		{
			Type:   v1.NodeMemoryPressure,
			Status: v1.ConditionFalse,
		},
	}

	t.Run("find existing condition", func(t *testing.T) {
		condition, found := FindCondition(conditions, v1.NodeReady)
		assert.True(t, found)
		assert.Equal(t, v1.NodeReady, condition.Type)
		assert.Equal(t, v1.ConditionTrue, condition.Status)
	})

	t.Run("find non-existing condition", func(t *testing.T) {
		condition, found := FindCondition(conditions, v1.NodeDiskPressure)
		assert.False(t, found)
		assert.Equal(t, v1.NodeCondition{}, *condition)
	})

	t.Run("empty conditions", func(t *testing.T) {
		condition, found := FindCondition([]v1.NodeCondition{}, v1.NodeReady)
		assert.False(t, found)
		assert.Equal(t, v1.NodeCondition{}, *condition)
	})

	t.Run("multiple same type returns latest LastHeartbeatTime", func(t *testing.T) {
		oldTime := metav1.NewTime(time.Now().Add(-time.Hour))
		newTime := metav1.NewTime(time.Now())
		conds := []v1.NodeCondition{
			{Type: v1.NodeReady, Status: v1.ConditionTrue, LastHeartbeatTime: oldTime},
			{Type: v1.NodeReady, Status: v1.ConditionFalse, LastHeartbeatTime: newTime},
		}
		condition, found := FindCondition(conds, v1.NodeReady)
		assert.True(t, found)
		assert.True(t, condition.LastHeartbeatTime.Equal(&newTime))
		assert.Equal(t, v1.ConditionFalse, condition.Status)
	})
}

func TestGetNodeCondition(t *testing.T) {
	node := &v1.Node{
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
				{
					Type:   v1.NodeMemoryPressure,
					Status: v1.ConditionFalse,
				},
			},
		},
	}

	t.Run("find existing condition", func(t *testing.T) {
		condition := GetNodeCondition(node, v1.NodeReady)
		assert.NotNil(t, condition)
		assert.Equal(t, v1.NodeReady, condition.Type)
		assert.Equal(t, v1.ConditionTrue, condition.Status)
	})

	t.Run("find non-existing condition", func(t *testing.T) {
		condition := GetNodeCondition(node, v1.NodeDiskPressure)
		assert.Nil(t, condition)
	})

	t.Run("nil conditions slice", func(t *testing.T) {
		nodeNoConds := &v1.Node{Status: v1.NodeStatus{Conditions: nil}}
		condition := GetNodeCondition(nodeNoConds, v1.NodeReady)
		assert.Nil(t, condition)
	})

	t.Run("empty node status", func(t *testing.T) {
		emptyNode := &v1.Node{}
		condition := GetNodeCondition(emptyNode, v1.NodeReady)
		assert.Nil(t, condition)
	})
}

func TestIsMasterNode(t *testing.T) {
	cases := []struct {
		name     string
		node     *v1.Node
		isMaster bool
	}{
		{
			name:     "master node",
			node:     &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{LabelNodeRoleMaster: "true"}}},
			isMaster: true,
		},
		{
			name:     "worker node",
			node:     &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"node-role.kubernetes.io/worker": "true"}}},
			isMaster: false,
		},
		{
			name:     "node without labels",
			node:     &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}},
			isMaster: false,
		},
		{
			name:     "nil node labels",
			node:     &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: nil}},
			isMaster: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isMaster, IsMasterNode(tt.node))
		})
	}
}

func TestGetNodeInternalIP(t *testing.T) {
	t.Run("node with internal IP", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{Type: v1.NodeInternalIP, Address: "192.168.1.100"},
					{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
				},
			},
		}

		ip, err := GetNodeInternalIP(node)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.100", ip)
	})

	t.Run("node without addresses", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
			Status:     v1.NodeStatus{Addresses: []v1.NodeAddress{}},
		}

		_, err := GetNodeInternalIP(node)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "do not contains addresses")
	})

	t.Run("node without internal IP", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
				},
			},
		}

		_, err := GetNodeInternalIP(node)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not find InternalIP")
	})

	t.Run("node with IPv6 only", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
			Status: v1.NodeStatus{
				Addresses: []v1.NodeAddress{
					{Type: v1.NodeInternalIP, Address: "2001:db8::1"},
				},
			},
		}

		_, err := GetNodeInternalIP(node)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not find InternalIP")
	})
}

func TestNodeInfo(t *testing.T) {
	t.Run("normal node", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "test"},
				},
			},
			Status: v1.NodeStatus{
				Images: []v1.ContainerImage{
					{Names: []string{"image1"}},
				},
			},
		}

		info := NodeInfo(node)
		assert.NotEmpty(t, info)
		assert.Contains(t, info, "test-node")
		assert.NotContains(t, info, "ManagedFields")
		assert.NotContains(t, info, "Images")
	})

	t.Run("nil node", func(t *testing.T) {
		info := NodeInfo(nil)
		assert.Empty(t, info)
	})
}

func TestFindNodeByNodeName(t *testing.T) {
	nodes := []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "node-3"},
		},
	}

	t.Run("find existing node", func(t *testing.T) {
		node := FindNodeByNodeName(nodes, "node-2")
		assert.NotNil(t, node)
		assert.Equal(t, "node-2", node.Name)
	})

	t.Run("find first node", func(t *testing.T) {
		node := FindNodeByNodeName(nodes, "node-1")
		assert.NotNil(t, node)
		assert.Equal(t, "node-1", node.Name)
	})

	t.Run("find last node", func(t *testing.T) {
		node := FindNodeByNodeName(nodes, "node-3")
		assert.NotNil(t, node)
		assert.Equal(t, "node-3", node.Name)
	})

	t.Run("node not found", func(t *testing.T) {
		node := FindNodeByNodeName(nodes, "node-4")
		assert.Nil(t, node)
	})

	t.Run("empty node list", func(t *testing.T) {
		node := FindNodeByNodeName([]v1.Node{}, "node-1")
		assert.Nil(t, node)
	})

	t.Run("find in single node list", func(t *testing.T) {
		singleNode := []v1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "only-node"}},
		}
		node := FindNodeByNodeName(singleNode, "only-node")
		assert.NotNil(t, node)
		assert.Equal(t, "only-node", node.Name)
	})
}
