package helper

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

type mockObject struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Name          string
	Namespace     string
}

func (m *mockObject) GetObjectKind() schema.ObjectKind {
	return &m.TypeMeta
}

func (m *mockObject) DeepCopyObject() runtime.Object {
	return &mockObject{
		TypeMeta:   m.TypeMeta,
		ObjectMeta: m.ObjectMeta,
		Name:       m.Name,
		Namespace:  m.Namespace,
	}
}

func TestNewTaskQueue(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewTaskQueue(syncFn)

	assert.NotNil(t, queue)
	assert.NotNil(t, queue.sync)
	assert.NotNil(t, queue.queue)
	assert.NotNil(t, queue.workerDone)
	assert.NotNil(t, queue.fn)
}

func TestNewCustomTaskQueue(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	customKeyFunc := func(obj interface{}) (interface{}, error) {
		return "custom-key", nil
	}

	queue := NewCustomTaskQueue(syncFn, customKeyFunc)

	assert.NotNil(t, queue)
	assert.NotNil(t, queue.fn)

	key, err := queue.fn("test-obj")
	assert.NoError(t, err)
	assert.Equal(t, "custom-key", key)
}

func TestNewCustomTaskQueueWithNilKeyFunc(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewCustomTaskQueue(syncFn, nil)

	assert.NotNil(t, queue)
	assert.NotNil(t, queue.fn)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	key, err := queue.fn(mockObj)
	assert.NoError(t, err)
	assert.Equal(t, "default/test", key)
}

func TestElementString(t *testing.T) {
	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	element := Element{
		Key:   mockObj,
		Event: event,
	}

	result := element.String()
	assert.Equal(t, "default/test", result)
}

func TestEnqueueTask(t *testing.T) {
	var processedItems []interface{}
	var mu sync.Mutex

	syncFn := func(obj interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		processedItems = append(processedItems, obj)
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	queue.EnqueueTask(event)

	err := wait.PollImmediate(time.Millisecond*10, time.Second, func() (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		return len(processedItems) > 0, nil
	})

	assert.NoError(t, err)

	mu.Lock()
	assert.Equal(t, 1, len(processedItems))
	mu.Unlock()
}

func TestEnqueueSkippableTask(t *testing.T) {
	var processedItems []interface{}
	var mu sync.Mutex

	syncFn := func(obj interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		processedItems = append(processedItems, obj)
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	queue.EnqueueSkippableTask(event)

	err := wait.PollImmediate(time.Millisecond*10, time.Second, func() (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		return len(processedItems) > 0, nil
	})

	assert.NoError(t, err)

	mu.Lock()
	assert.Equal(t, 1, len(processedItems))
	mu.Unlock()
}

func TestEnqueueTaskWhenShuttingDown(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})

	// Start worker first
	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	// Then shutdown
	close(stopCh)
	queue.Shutdown()

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	assert.NotPanics(t, func() {
		queue.EnqueueTask(event)
	})
}

func TestWorkerProcessing(t *testing.T) {
	var processedItems []interface{}
	var mu sync.Mutex

	syncFn := func(obj interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		processedItems = append(processedItems, obj)
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	for i := 0; i < 3; i++ {
		mockObj := &mockObject{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("test-%d", i),
				Namespace: "default",
			},
		}
		event := Event{
			Type: CreateEvent,
			Obj:  mockObj,
		}
		queue.EnqueueTask(event)
	}

	err := wait.PollImmediate(time.Millisecond*10, time.Second*2, func() (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		return len(processedItems) == 3, nil
	})

	assert.NoError(t, err)

	mu.Lock()
	assert.Equal(t, 3, len(processedItems))
	mu.Unlock()
}

func TestWorkerErrorHandling(t *testing.T) {
	var processedItems []interface{}
	var mu sync.Mutex
	processCount := 0

	syncFn := func(obj interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		processCount++
		processedItems = append(processedItems, obj)

		if processCount <= 2 {
			return errors.New("processing failed")
		}
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	queue.EnqueueTask(event)

	err := wait.PollImmediate(time.Millisecond*50, time.Second*5, func() (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		return processCount >= 3, nil
	})

	assert.NoError(t, err)

	mu.Lock()
	assert.GreaterOrEqual(t, processCount, 3)
	mu.Unlock()
}

func TestIsShuttingDown(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})

	assert.False(t, queue.IsShuttingDown())

	// Start worker first
	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	// Then shutdown
	close(stopCh)
	queue.Shutdown()

	assert.True(t, queue.IsShuttingDown())
}

func TestShutdown(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	queue.Shutdown()

	assert.True(t, queue.IsShuttingDown())
}

func TestDefaultKeyFunc(t *testing.T) {
	queue := &Queue{}

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	key, err := queue.defaultKeyFunc(mockObj)
	assert.NoError(t, err)
	assert.Equal(t, "default/test", key)

	invalidObj := "invalid-object"
	_, err = queue.defaultKeyFunc(invalidObj)
	assert.Error(t, err)
}

func TestEnqueueWithInvalidKey(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	failingKeyFunc := func(obj interface{}) (interface{}, error) {
		return nil, errors.New("key generation failed")
	}

	queue := NewCustomTaskQueue(syncFn, failingKeyFunc)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	assert.NotPanics(t, func() {
		queue.EnqueueTask(event)
	})

	time.Sleep(time.Millisecond * 100)
}

func TestLastSyncTimestamp(t *testing.T) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	initialLastSync := queue.lastSync

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	queue.EnqueueTask(event)
	queue.Shutdown()

	assert.Greater(t, queue.lastSync, initialLastSync)
}

func TestRateLimiting(t *testing.T) {
	var processedItems []interface{}
	var mu sync.Mutex

	syncFn := func(obj interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		processedItems = append(processedItems, obj)
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	queue.EnqueueTask(event)

	err := wait.PollImmediate(time.Millisecond*10, time.Second, func() (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		return len(processedItems) > 0, nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, queue.queue)
}

func BenchmarkEnqueueTask(b *testing.B) {
	syncFn := func(obj interface{}) error {
		return nil
	}

	queue := NewTaskQueue(syncFn)
	stopCh := make(chan struct{})

	// Start worker
	go queue.Run(1, time.Millisecond*10, stopCh)
	time.Sleep(time.Millisecond * 50)

	mockObj := &mockObject{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	event := Event{
		Type: CreateEvent,
		Obj:  mockObj,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.EnqueueTask(event)
	}
	b.StopTimer()

	// Cleanup
	close(stopCh)
	queue.Shutdown()
}
