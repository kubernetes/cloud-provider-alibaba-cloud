package route

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnqueueRequestForNodeEvent_Create(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]()

	// Create the event handler
	handler := &enqueueRequestForNodeEvent{
		rateLimiter: rateLimiter,
	}

	// Create a test node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}

	// Create a fake queue
	queue := workqueue.NewTypedRateLimitingQueue(rateLimiter)

	// Create an event
	createEvent := event.TypedCreateEvent[*corev1.Node]{
		Object: node,
	}

	// Call the Create method
	handler.Create(context.TODO(), createEvent, queue)

	// Verify the queue has one item
	assert.Equal(t, 1, queue.Len())

	// Get the item from the queue
	item, shutdown := queue.Get()
	assert.False(t, shutdown)

	// Verify the item is a reconcile.Request with correct name
	assert.Equal(t, "test-node", item.Name)

	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_Update(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]()

	// Create the event handler
	handler := &enqueueRequestForNodeEvent{
		rateLimiter: rateLimiter,
	}

	// Create test nodes
	oldNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "old-node",
		},
	}

	newNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-node",
		},
	}

	// Create a fake queue
	queue := workqueue.NewTypedRateLimitingQueue(rateLimiter)

	// Create an update event
	updateEvent := event.TypedUpdateEvent[*corev1.Node]{
		ObjectOld: oldNode,
		ObjectNew: newNode,
	}

	// Call the Update method
	handler.Update(context.TODO(), updateEvent, queue)

	// Verify the queue has one item
	assert.Equal(t, 1, queue.Len())

	// Get the item from the queue
	item, shutdown := queue.Get()
	assert.False(t, shutdown)

	// Verify the item is a reconcile.Request with correct name (should be new node name)
	assert.Equal(t, "new-node", item.Name)

	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_Delete(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]()

	// Create the event handler
	handler := &enqueueRequestForNodeEvent{
		rateLimiter: rateLimiter,
	}

	// Create a test node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deleted-node",
		},
	}

	// Create a fake queue
	queue := workqueue.NewTypedRateLimitingQueue(rateLimiter)

	// Create a delete event
	deleteEvent := event.TypedDeleteEvent[*corev1.Node]{
		Object: node,
	}

	// Call the Delete method
	handler.Delete(context.TODO(), deleteEvent, queue)

	// Verify the queue has one item
	assert.Equal(t, 1, queue.Len())

	// Get the item from the queue
	item, shutdown := queue.Get()
	assert.False(t, shutdown)

	// Verify the item is a reconcile.Request with correct name
	assert.Equal(t, "deleted-node", item.Name)

	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_Generic(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]()

	// Create the event handler
	handler := &enqueueRequestForNodeEvent{
		rateLimiter: rateLimiter,
	}

	// Create a test node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "generic-node",
		},
	}

	// Create a fake queue
	queue := workqueue.NewTypedRateLimitingQueue(rateLimiter)

	// Create a generic event
	genericEvent := event.TypedGenericEvent[*corev1.Node]{
		Object: node,
	}

	// Call the Generic method
	handler.Generic(context.TODO(), genericEvent, queue)

	// Need to wait a bit for the AddAfter to process
	time.Sleep(100 * time.Millisecond)

	// Verify the queue has one item
	assert.Equal(t, 1, queue.Len())

	// Get the item from the queue
	item, shutdown := queue.Get()
	assert.False(t, shutdown)

	// Verify the item is a reconcile.Request with correct name
	assert.Equal(t, "generic-node", item.Name)

	// Done with the item
	queue.Done(item)
}
