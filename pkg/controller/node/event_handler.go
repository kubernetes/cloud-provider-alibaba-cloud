package node

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type enqueueRequestForNodeEvent struct{}

var _ handler.TypedEventHandler[*v1.Node, reconcile.Request] = (*enqueueRequestForNodeEvent)(nil)

// NewEnqueueRequestForNodeEvent, event handler for node event
func NewEnqueueRequestForNodeEvent() *enqueueRequestForNodeEvent {
	return &enqueueRequestForNodeEvent{}
}

func (h *enqueueRequestForNodeEvent) Create(_ context.Context, e event.TypedCreateEvent[*v1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	node := e.Object
	if needAdd(node) {
		queue.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: node.Name,
			},
		})
		util.NodeLog.Info("enqueue: node create", "node", util.Key(node), "queueLen", queue.Len())
	}
}

func (h *enqueueRequestForNodeEvent) Update(_ context.Context, e event.TypedUpdateEvent[*v1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {

}

func (h *enqueueRequestForNodeEvent) Delete(_ context.Context, e event.TypedDeleteEvent[*v1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {

}

func (h *enqueueRequestForNodeEvent) Generic(_ context.Context, e event.TypedGenericEvent[*v1.Node], queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {

}

func needAdd(node *v1.Node) bool {
	cloudTaint := findCloudTaint(node.Spec.Taints)
	return cloudTaint != nil
}
