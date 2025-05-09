package helper

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type APIObject interface {
	metav1.Object
	runtime.Object
}

type FinalizerManager interface {
	AddFinalizers(ctx context.Context, obj APIObject, finalizers ...string) error
	RemoveFinalizers(ctx context.Context, obj APIObject, finalizers ...string) error
}

type defaultFinalizerManager struct {
	kubeClient client.Client
}

var _ FinalizerManager = &defaultFinalizerManager{}

func NewDefaultFinalizerManager(kclient client.Client) *defaultFinalizerManager {
	return &defaultFinalizerManager{kubeClient: kclient}
}

func (m *defaultFinalizerManager) AddFinalizers(ctx context.Context, obj APIObject, finalizers ...string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := m.kubeClient.Get(ctx, util.NamespacedName(obj), obj); err != nil {
			return err
		}

		oldObj := obj.DeepCopyObject().(client.Object)
		needsUpdate := false
		for _, finalizer := range finalizers {
			if !HasFinalizer(obj, finalizer) {
				controllerutil.AddFinalizer(obj, finalizer)
				needsUpdate = true
			}
		}
		if !needsUpdate {
			return nil
		}
		return m.kubeClient.Patch(ctx, obj, client.MergeFromWithOptions(oldObj, client.MergeFromWithOptimisticLock{}))
	})
}

func (m *defaultFinalizerManager) RemoveFinalizers(ctx context.Context, obj APIObject, finalizers ...string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := m.kubeClient.Get(ctx, util.NamespacedName(obj), obj); err != nil {
			return err
		}

		oldObj := obj.DeepCopyObject().(client.Object)
		needsUpdate := false
		for _, finalizer := range finalizers {
			if HasFinalizer(obj, finalizer) {
				controllerutil.RemoveFinalizer(obj, finalizer)
				needsUpdate = true
			}
		}
		if !needsUpdate {
			return nil
		}
		return m.kubeClient.Patch(ctx, obj, client.MergeFromWithOptions(oldObj, client.MergeFromWithOptimisticLock{}))
	})
}

// HasFinalizer tests whether k8s object has specified finalizer
func HasFinalizer(obj metav1.Object, finalizer string) bool {
	f := obj.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}
