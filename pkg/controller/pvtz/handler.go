package pvtz

import (
	"k8s.io/klog/v2"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewEventHandlerWithClient() *EventHandlerWithClient {
	h := &EventHandlerWithClient{}
	h.EventHandler = handler.EnqueueRequestsFromMapFunc(h.norm)
	return h
}

type EventHandlerWithClient struct {
	client client.Client
	handler.EventHandler
}

// norm changes endpoints/services update into services
func (e *EventHandlerWithClient) norm(o client.Object) []reconcile.Request {
	var request []reconcile.Request
	switch o := o.(type) {
	case *v1.Endpoints:
		request = append(request, e.normEndpoint(o)...)
	case *v1.Service:
		request = append(request, e.normService(o)...)
	case *v1.Pod:
		request = append(request, e.normPod(o)...)
	default:
		klog.Warningf("unknown object: %s, %v", reflect.TypeOf(o), o)
	}
	return request
}

func (e *EventHandlerWithClient) normEndpoint(o *v1.Endpoints) []reconcile.Request {
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (e *EventHandlerWithClient) normService(o *v1.Service) []reconcile.Request {
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (e *EventHandlerWithClient) normPod(o *v1.Pod) []reconcile.Request {
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (e *EventHandlerWithClient) InjectClient(c client.Client) error {
	e.client = c
	return nil
}
