package k8s

import (
	"fmt"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"

	"k8s.io/klog/v2"

	"golang.org/x/time/rate"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	keyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

// Queue manages a time work queue through an independent worker that invokes the
// given sync function for every work item inserted.
// The queue uses an internal timestamp that allows the removal of certain elements
// which timestamp is older than the last successful get operation.
type Queue struct {
	// queue is the work queue the worker polls
	queue workqueue.RateLimitingInterface
	// sync is called for each item in the queue
	sync func(interface{}) error
	// workerDone is closed when the worker exits
	workerDone chan bool
	// fn makes a key for an API object
	fn func(obj interface{}) (interface{}, error)
	// lastSync is the Unix epoch time of the last execution of 'sync'
	lastSync int64
}

// Element represents one item of the queue
type Element struct {
	Key         interface{}
	Event       helper.Event
	IsSkippable bool
}

// String returns the general purpose string representation
func (n Element) String() string {
	namespaceAndName, _ := cache.MetaNamespaceKeyFunc(n.Key)
	return namespaceAndName
}

// Run starts processing elements in the queue
func (t *Queue) Run(maxConcurrentReconciles int, period time.Duration, stopCh <-chan struct{}) {
	for i := 0; i < maxConcurrentReconciles; i++ {
		go wait.Until(t.worker, period, stopCh)
	}

}

// EnqueueTask enqueues ns/name of the given api object in the task queue.
func (t *Queue) EnqueueTask(obj interface{}) {
	t.enqueue(obj, false)
}

// EnqueueSkippableTask enqueues ns/name of the given api object in
// the task queue that can be skipped
func (t *Queue) EnqueueSkippableTask(obj interface{}) {
	t.enqueue(obj, true)
}

// enqueue enqueues ns/name of the given api object in the task queue.
func (t *Queue) enqueue(obj interface{}, skippable bool) {
	if t.IsShuttingDown() {
		klog.Errorf("queue has been shutdown, failed to enqueue. key: %+v", obj)
		return
	}

	klog.V(3).Infof("queuing, item: %+v", obj)
	e := obj.(helper.Event)
	key, err := t.fn(e.Obj)
	if err != nil {
		klog.Errorf("creating object key, item: %+v, error: %s", obj, err.Error())
		return
	}
	t.queue.Add(Element{
		Key:   key,
		Event: obj.(helper.Event),
	})
}

func (t *Queue) defaultKeyFunc(obj interface{}) (interface{}, error) {
	key, err := keyFunc(obj)
	if err != nil {
		return "", fmt.Errorf("could not get key for object %+v: %v", obj, err)
	}

	return key, nil
}

// worker processes work in the queue through sync.
func (t *Queue) worker() {
	for {
		key, quit := t.queue.Get()
		if quit {
			if !isClosed(t.workerDone) {
				close(t.workerDone)
			}
			return
		}
		ts := time.Now().UnixNano()

		item := key.(Element)
		klog.V(3).Infof("syncing: key: %s", item.Key)
		if err := t.sync(key); err != nil {
			klog.Errorf("requeuing: key: %s, error: %s", item.Key, err.Error())
			t.queue.AddRateLimited(Element{
				Key:   item.Key,
				Event: item.Event,
			})
		} else {
			t.queue.Forget(key)
			t.lastSync = ts
		}

		t.queue.Done(key)
	}
}

func isClosed(ch <-chan bool) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

// Shutdown shuts down the work queue and waits for the worker to ACK
func (t *Queue) Shutdown() {
	t.queue.ShutDown()
	<-t.workerDone
}

// IsShuttingDown returns if the method Shutdown was invoked
func (t *Queue) IsShuttingDown() bool {
	return t.queue.ShuttingDown()
}

// NewTaskQueue creates a new task queue with the given sync function.
// The sync function is called for every element inserted into the queue.
func NewTaskQueue(syncFn func(interface{}) error) *Queue {
	return NewCustomTaskQueue(syncFn, nil)
}

// NewCustomTaskQueue ...
func NewCustomTaskQueue(syncFn func(interface{}) error, fn func(interface{}) (interface{}, error)) *Queue {
	q := &Queue{
		queue: workqueue.NewRateLimitingQueue(workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(500*time.Millisecond, 1000*time.Second),
			// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		)),
		sync:       syncFn,
		workerDone: make(chan bool),
		fn:         fn,
	}

	if fn == nil {
		q.fn = q.defaultKeyFunc
	}

	return q
}

// GetDummyObject returns a valid object that can be used in the Queue
func GetDummyObject(name string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name: name,
	}
}
func GetServiceDummyObject(name, namespace string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
}
