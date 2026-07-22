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
	rateLimiter workqueue.TypedRateLimiter[reconcile.Request]
}

var _ handler.TypedEventHandler[*corev1.Node, reconcile.Request] = (*enqueueRequestForNodeEvent)(nil)

func (h *enqueueRequestForNodeEvent) Create(_ context.Context, e event.TypedCreateEvent[*corev1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	n := e.Object
	queue.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: n.Name,
		},
	})
}

func (h *enqueueRequestForNodeEvent) Update(_ context.Context, e event.TypedUpdateEvent[*corev1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	n := e.ObjectNew
	queue.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: n.Name,
		},
	})
}

func (h *enqueueRequestForNodeEvent) Delete(_ context.Context, e event.TypedDeleteEvent[*corev1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	n := e.Object
	queue.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: n.Name,
		},
	})
}

func (h *enqueueRequestForNodeEvent) Generic(_ context.Context, e event.TypedGenericEvent[*corev1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	n := e.Object
	r := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: n.Name,
		},
	}
	queue.AddAfter(r, h.rateLimiter.When(r))
	log.Info("enqueue: route requeue", "node", n.Name, "queueLen", queue.Len())
}
