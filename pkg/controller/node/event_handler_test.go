package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnqueueRequestForNodeEvent_Update(t *testing.T) {
	handler := NewEnqueueRequestForNodeEvent()
	assert.NotNil(t, handler)

	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]())
	defer queue.ShutDown()

	oldNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}

	newNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}

	updateEvent := event.TypedUpdateEvent[*v1.Node]{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	}

	// Call Update - should not add anything to queue (empty implementation)
	handler.Update(context.Background(), updateEvent, queue)

	// Verify queue is empty since Update does nothing
	assert.Equal(t, 0, queue.Len(), "Update should not add items to queue")
}

func TestEnqueueRequestForNodeEvent_Delete(t *testing.T) {
	handler := NewEnqueueRequestForNodeEvent()
	assert.NotNil(t, handler)

	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]())
	defer queue.ShutDown()

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}

	deleteEvent := event.TypedDeleteEvent[*v1.Node]{
		Object: node,
	}

	// Call Delete - should not add anything to queue (empty implementation)
	handler.Delete(context.Background(), deleteEvent, queue)

	// Verify queue is empty since Delete does nothing
	assert.Equal(t, 0, queue.Len(), "Delete should not add items to queue")
}

func TestEnqueueRequestForNodeEvent_Generic(t *testing.T) {
	handler := NewEnqueueRequestForNodeEvent()
	assert.NotNil(t, handler)

	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]())
	defer queue.ShutDown()

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}

	genericEvent := event.TypedGenericEvent[*v1.Node]{
		Object: node,
	}

	// Call Generic - should not add anything to queue (empty implementation)
	handler.Generic(context.Background(), genericEvent, queue)

	// Verify queue is empty since Generic does nothing
	assert.Equal(t, 0, queue.Len(), "Generic should not add items to queue")
}

func TestEnqueueRequestForNodeEvent_Create(t *testing.T) {
	tests := []struct {
		name          string
		node          *v1.Node
		expectedQueue int
		description   string
	}{
		{
			name: "node with cloud taint should be enqueued",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-1",
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    "node.cloudprovider.kubernetes.io/uninitialized",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expectedQueue: 1,
			description:   "should enqueue node with cloud taint",
		},
		{
			name: "node without cloud taint should not be enqueued",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-2",
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{
							Key:    "some-other-taint",
							Value:  "true",
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expectedQueue: 0,
			description:   "should not enqueue node without cloud taint",
		},
		{
			name: "node with no taints should not be enqueued",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-3",
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{},
				},
			},
			expectedQueue: 0,
			description:   "should not enqueue node with no taints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEnqueueRequestForNodeEvent()
			queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]())
			defer queue.ShutDown()

			createEvent := event.TypedCreateEvent[*v1.Node]{
				Object: tt.node,
			}

			handler.Create(context.Background(), createEvent, queue)
			assert.Equal(t, tt.expectedQueue, queue.Len(), tt.description)
		})
	}
}
