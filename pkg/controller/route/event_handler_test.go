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
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	
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
	queue := workqueue.NewRateLimitingQueue(rateLimiter)
	
	// Create an event
	createEvent := event.CreateEvent{
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
	request, ok := item.(reconcile.Request)
	assert.True(t, ok)
	assert.Equal(t, "test-node", request.Name)
	
	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_Update(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	
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
	queue := workqueue.NewRateLimitingQueue(rateLimiter)
	
	// Create an update event
	updateEvent := event.UpdateEvent{
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
	request, ok := item.(reconcile.Request)
	assert.True(t, ok)
	assert.Equal(t, "new-node", request.Name)
	
	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_Delete(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	
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
	queue := workqueue.NewRateLimitingQueue(rateLimiter)
	
	// Create a delete event
	deleteEvent := event.DeleteEvent{
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
	request, ok := item.(reconcile.Request)
	assert.True(t, ok)
	assert.Equal(t, "deleted-node", request.Name)
	
	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_Generic(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	
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
	queue := workqueue.NewRateLimitingQueue(rateLimiter)
	
	// Create a generic event
	genericEvent := event.GenericEvent{
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
	request, ok := item.(reconcile.Request)
	assert.True(t, ok)
	assert.Equal(t, "generic-node", request.Name)
	
	// Done with the item
	queue.Done(item)
}

func TestEnqueueRequestForNodeEvent_NonNodeObjects(t *testing.T) {
	// Create a rate limiter
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	
	// Create the event handler
	handler := &enqueueRequestForNodeEvent{
		rateLimiter: rateLimiter,
	}
	
	// Create a fake queue
	queue := workqueue.NewRateLimitingQueue(rateLimiter)
	
	// Test with non-Node object in Create event
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}
	
	createEvent := event.CreateEvent{
		Object: pod,
	}
	
	// Call the Create method with non-Node object
	handler.Create(context.TODO(), createEvent, queue)
	
	// Verify the queue is empty
	assert.Equal(t, 0, queue.Len())
	
	// Test with non-Node object in Update event
	oldPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-pod",
			Namespace: "default",
		},
	}
	
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-pod",
			Namespace: "default",
		},
	}
	
	updateEvent := event.UpdateEvent{
		ObjectOld: oldPod,
		ObjectNew: newPod,
	}
	
	// Call the Update method with non-Node object
	handler.Update(context.TODO(), updateEvent, queue)
	
	// Verify the queue is still empty
	assert.Equal(t, 0, queue.Len())
	
	// Test with non-Node object in Delete event
	deleteEvent := event.DeleteEvent{
		Object: pod,
	}
	
	// Call the Delete method with non-Node object
	handler.Delete(context.TODO(), deleteEvent, queue)
	
	// Verify the queue is still empty
	assert.Equal(t, 0, queue.Len())
	
	// Test with non-Node object in Generic event
	genericEvent := event.GenericEvent{
		Object: pod,
	}
	
	// Call the Generic method with non-Node object
	handler.Generic(context.TODO(), genericEvent, queue)
	
	// Need to wait a bit for the AddAfter to process
	time.Sleep(100 * time.Millisecond)
	
	// Verify the queue is still empty
	assert.Equal(t, 0, queue.Len())
}