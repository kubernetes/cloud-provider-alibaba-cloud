package route

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type enqueueRequestForNodeEvent struct {
	rateLimiter workqueue.RateLimiter
}

var _ handler.EventHandler = (*enqueueRequestForNodeEvent)(nil)

func (h *enqueueRequestForNodeEvent) Create(_ context.Context, e event.CreateEvent, queue workqueue.RateLimitingInterface) {
	if n, ok := e.Object.(*corev1.Node); ok {
		queue.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: n.Name,
			},
		})
	}
}

func (h *enqueueRequestForNodeEvent) Update(_ context.Context, e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	if n, ok := e.ObjectNew.(*corev1.Node); ok {
		queue.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: n.Name,
			},
		})
	}
}

func (h *enqueueRequestForNodeEvent) Delete(_ context.Context, e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	if n, ok := e.Object.(*corev1.Node); ok {
		queue.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: n.Name,
			},
		})
	}
}

func (h *enqueueRequestForNodeEvent) Generic(_ context.Context, e event.GenericEvent, queue workqueue.RateLimitingInterface) {
	if n, ok := e.Object.(*corev1.Node); ok {
		r := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: n.Name,
			},
		}
		queue.AddAfter(r, h.rateLimiter.When(r))
		log.Info("enqueue: route requeue", "node", n.Name, "queueLen", queue.Len())
	}
}
