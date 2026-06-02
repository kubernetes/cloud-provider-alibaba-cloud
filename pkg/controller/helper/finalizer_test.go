package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestHasFinalizer(t *testing.T) {
	t.Run("object has no finalizers", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{},
			},
		}
		assert.False(t, HasFinalizer(svc, "test-finalizer"))
	})

	t.Run("object has the specified finalizer", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{"test-finalizer", "another-finalizer"},
			},
		}
		assert.True(t, HasFinalizer(svc, "test-finalizer"))
	})

	t.Run("object has other finalizers but not the specified one", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{"another-finalizer", "yet-another-finalizer"},
			},
		}
		assert.False(t, HasFinalizer(svc, "test-finalizer"))
	})
}

func TestAddFinalizers(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	t.Run("add new finalizer to object without finalizers", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
		manager := NewDefaultFinalizerManager(c)

		err := manager.AddFinalizers(context.Background(), svc, "test-finalizer")
		assert.NoError(t, err)

		// Get the updated object
		updatedSvc := &v1.Service{}
		err = c.Get(context.Background(), client.ObjectKeyFromObject(svc), updatedSvc)
		assert.NoError(t, err)
		assert.True(t, HasFinalizer(updatedSvc, "test-finalizer"))
	})

	t.Run("add existing finalizer to object", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{"test-finalizer"},
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
		manager := NewDefaultFinalizerManager(c)

		err := manager.AddFinalizers(context.Background(), svc, "test-finalizer")
		assert.NoError(t, err)

		// Get the updated object
		updatedSvc := &v1.Service{}
		err = c.Get(context.Background(), client.ObjectKeyFromObject(svc), updatedSvc)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(updatedSvc.Finalizers))
		assert.True(t, HasFinalizer(updatedSvc, "test-finalizer"))
	})

	t.Run("add multiple finalizers", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
		manager := NewDefaultFinalizerManager(c)

		err := manager.AddFinalizers(context.Background(), svc, "finalizer-1", "finalizer-2")
		assert.NoError(t, err)

		// Get the updated object
		updatedSvc := &v1.Service{}
		err = c.Get(context.Background(), client.ObjectKeyFromObject(svc), updatedSvc)
		assert.NoError(t, err)
		assert.True(t, HasFinalizer(updatedSvc, "finalizer-1"))
		assert.True(t, HasFinalizer(updatedSvc, "finalizer-2"))
		assert.Equal(t, 2, len(updatedSvc.Finalizers))
	})
}

func TestRemoveFinalizers(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	t.Run("remove existing finalizer", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{"test-finalizer", "another-finalizer"},
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
		manager := NewDefaultFinalizerManager(c)

		err := manager.RemoveFinalizers(context.Background(), svc, "test-finalizer")
		assert.NoError(t, err)

		// Get the updated object
		updatedSvc := &v1.Service{}
		err = c.Get(context.Background(), client.ObjectKeyFromObject(svc), updatedSvc)
		assert.NoError(t, err)
		assert.False(t, HasFinalizer(updatedSvc, "test-finalizer"))
		assert.True(t, HasFinalizer(updatedSvc, "another-finalizer"))
		assert.Equal(t, 1, len(updatedSvc.Finalizers))
	})

	t.Run("remove non-existing finalizer", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{"another-finalizer"},
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
		manager := NewDefaultFinalizerManager(c)

		err := manager.RemoveFinalizers(context.Background(), svc, "test-finalizer")
		assert.NoError(t, err)

		// Get the updated object
		updatedSvc := &v1.Service{}
		err = c.Get(context.Background(), client.ObjectKeyFromObject(svc), updatedSvc)
		assert.NoError(t, err)
		assert.False(t, HasFinalizer(updatedSvc, "test-finalizer"))
		assert.True(t, HasFinalizer(updatedSvc, "another-finalizer"))
		assert.Equal(t, 1, len(updatedSvc.Finalizers))
	})

	t.Run("remove multiple finalizers", func(t *testing.T) {
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test",
				Namespace:  "default",
				Finalizers: []string{"finalizer-1", "finalizer-2", "finalizer-3"},
			},
		}

		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
		manager := NewDefaultFinalizerManager(c)

		err := manager.RemoveFinalizers(context.Background(), svc, "finalizer-1", "finalizer-2")
		assert.NoError(t, err)

		// Get the updated object
		updatedSvc := &v1.Service{}
		err = c.Get(context.Background(), client.ObjectKeyFromObject(svc), updatedSvc)
		assert.NoError(t, err)
		assert.False(t, HasFinalizer(updatedSvc, "finalizer-1"))
		assert.False(t, HasFinalizer(updatedSvc, "finalizer-2"))
		assert.True(t, HasFinalizer(updatedSvc, "finalizer-3"))
		assert.Equal(t, 1, len(updatedSvc.Finalizers))
	})
}
