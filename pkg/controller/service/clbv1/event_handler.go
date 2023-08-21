package clbv1

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewEnqueueRequestForServiceEvent(eventRecorder record.EventRecorder) *enqueueRequestForServiceEvent {
	return &enqueueRequestForServiceEvent{eventRecorder: eventRecorder}
}

type enqueueRequestForServiceEvent struct {
	eventRecorder record.EventRecorder
}

var _ handler.EventHandler = (*enqueueRequestForServiceEvent)(nil)

func (h *enqueueRequestForServiceEvent) Create(e event.CreateEvent, queue workqueue.RateLimitingInterface) {
	svc, ok := e.Object.(*v1.Service)
	if ok && needAdd(svc) {
		util.ServiceLog.Info("controller: service create event", "service", util.Key(svc))
		h.enqueueManagedService(queue, svc)
	}
}

func (h *enqueueRequestForServiceEvent) Update(e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	oldSvc, ok1 := e.ObjectOld.(*v1.Service)
	newSvc, ok2 := e.ObjectNew.(*v1.Service)

	if ok1 && ok2 && needUpdate(oldSvc, newSvc, h.eventRecorder) {
		util.ServiceLog.Info("controller: service update event", "service", util.Key(oldSvc))
		h.enqueueManagedService(queue, newSvc)
	}
}

func (h *enqueueRequestForServiceEvent) Delete(e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	// Services have the finalizer. When a service is deleted, it will update the deletionTimestamp of the service.
	// Since a delete event has changed to an update event, it is safe to ignore it.
}

func (h *enqueueRequestForServiceEvent) Generic(e event.GenericEvent, queue workqueue.RateLimitingInterface) {
	// unknown type event, ignore
}

func (h *enqueueRequestForServiceEvent) enqueueManagedService(queue workqueue.RateLimitingInterface, service *v1.Service) {
	queue.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: service.Namespace,
			Name:      service.Name,
		},
	})
	util.ServiceLog.Info("enqueue", "service", util.Key(service), "queueLen", queue.Len())
}

func needUpdate(oldSvc, newSvc *v1.Service, recorder record.EventRecorder) bool {
	if !helper.NeedCLB(oldSvc) && !helper.NeedCLB(newSvc) {
		return false
	}

	if helper.NeedCLB(oldSvc) != helper.NeedCLB(newSvc) {
		if newSvc.Annotations[helper.LoadBalancerClass] != "" {
			util.ServiceLog.Info(fmt.Sprintf("service has LoadBalancerClass %s, skip update", newSvc.Annotations[helper.LoadBalancerClass]),
				"service", util.Key(oldSvc))
			return false
		}
		util.ServiceLog.Info(fmt.Sprintf("TypeChanged %v - %v", oldSvc.Spec.Type, newSvc.Spec.Type),
			"service", util.Key(oldSvc))
		recorder.Event(
			newSvc,
			v1.EventTypeNormal,
			helper.TypeChanged,
			fmt.Sprintf("type change %v - %v", oldSvc.Spec.Type, newSvc.Spec.Type),
		)
		return true
	}

	if oldSvc.UID != newSvc.UID {
		util.ServiceLog.Info(fmt.Sprintf("UIDChanged: %v - %v", oldSvc.UID, newSvc.UID),
			"service", util.Key(oldSvc))
		return true
	}

	if !reflect.DeepEqual(oldSvc.Annotations, newSvc.Annotations) {
		util.ServiceLog.Info(fmt.Sprintf("AnnotationChanged: %v - %v",
			oldSvc.Annotations, newSvc.Annotations),
			"service", util.Key(oldSvc))
		recorder.Event(
			newSvc,
			v1.EventTypeNormal,
			helper.AnnoChanged,
			"The service will be updated because the annotations has been changed.",
		)
		return true
	}

	if !reflect.DeepEqual(oldSvc.Spec, newSvc.Spec) {
		util.ServiceLog.Info(fmt.Sprintf("SpecChanged: %v - %v", oldSvc.Spec, newSvc.Spec),
			"service", util.Key(oldSvc))
		recorder.Event(
			newSvc,
			v1.EventTypeNormal,
			helper.SpecChanged,
			"The service will be updated because the spec has been changed.",
		)
		return true
	}

	if !reflect.DeepEqual(oldSvc.DeletionTimestamp.IsZero(), newSvc.DeletionTimestamp.IsZero()) {
		util.ServiceLog.Info(fmt.Sprintf("DeleteTimestampChanged: %v - %v",
			oldSvc.DeletionTimestamp.IsZero(), newSvc.DeletionTimestamp.IsZero()),
			"service", util.Key(oldSvc))
		recorder.Event(
			newSvc,
			v1.EventTypeNormal,
			helper.DeleteTimestampChanged,
			"The service will be updated because the delete timestamp has been changed.",
		)
		return true
	}

	return false
}

func needAdd(newService *v1.Service) bool {
	if helper.NeedCLB(newService) {
		return true
	}

	// was LoadBalancer
	if helper.HasFinalizer(newService, helper.ServiceFinalizer) {
		util.ServiceLog.Info("service has service finalizer, which may was LoadBalancer", "service", util.Key(newService))
		return true
	}
	return false
}

// NewEnqueueRequestForEndpointEvent, event handler for endpoint events
func NewEnqueueRequestForEndpointEvent(eventRecorder record.EventRecorder) *enqueueRequestForEndpointEvent {
	return &enqueueRequestForEndpointEvent{eventRecorder: eventRecorder}
}

type enqueueRequestForEndpointEvent struct {
	client        client.Client
	eventRecorder record.EventRecorder
}

func (h *enqueueRequestForEndpointEvent) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

var _ handler.EventHandler = (*enqueueRequestForEndpointEvent)(nil)

func (h *enqueueRequestForEndpointEvent) Create(e event.CreateEvent, queue workqueue.RateLimitingInterface) {
	ep, ok := e.Object.(*v1.Endpoints)
	if ok && isEndpointProcessNeeded(ep, h.client) {
		util.ServiceLog.Info("controller: endpoint create event", "endpoint", util.Key(ep))
		h.enqueueManagedEndpoint(queue, ep)
	}
}

func (h *enqueueRequestForEndpointEvent) Update(e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	ep1, ok1 := e.ObjectOld.(*v1.Endpoints)
	ep2, ok2 := e.ObjectNew.(*v1.Endpoints)

	if ok1 && ok2 && isEndpointProcessNeeded(ep1, h.client) &&
		!reflect.DeepEqual(ep1.Subsets, ep2.Subsets) {
		util.ServiceLog.Info("controller: endpoint update event", "endpoint", util.Key(ep1))
		util.ServiceLog.Info(fmt.Sprintf("endpoints before [%s], afeter [%s]",
			helper.LogEndpoints(ep1), helper.LogEndpoints(ep2)), "endpoint", util.Key(ep1))
		h.enqueueManagedEndpoint(queue, ep1)
	}
}

func (h *enqueueRequestForEndpointEvent) Delete(e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	ep, ok := e.Object.(*v1.Endpoints)
	if ok && isEndpointProcessNeeded(ep, h.client) {
		util.ServiceLog.Info("controller: endpoint delete event", "endpoint", util.Key(ep))
		h.enqueueManagedEndpoint(queue, ep)
	}
}

func (h *enqueueRequestForEndpointEvent) Generic(e event.GenericEvent, queue workqueue.RateLimitingInterface) {
	// unknown event, ignore
}

func (h *enqueueRequestForEndpointEvent) enqueueManagedEndpoint(queue workqueue.RateLimitingInterface, endpoint *v1.Endpoints) {
	queue.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: endpoint.Namespace,
			Name:      endpoint.Name,
		},
	})
	util.ServiceLog.Info("enqueue", "endpoint", util.Key(endpoint), "queueLen", queue.Len())
}

func isEndpointProcessNeeded(ep *v1.Endpoints, client client.Client) bool {
	if ep == nil {
		return false
	}

	if len(ep.Annotations) != 0 {
		// skip eps which are used for leader election
		if _, ok := ep.Annotations[resourcelock.LeaderElectionRecordAnnotationKey]; ok {
			return false
		}
	}

	svc := &v1.Service{}
	err := client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: ep.GetNamespace(),
			Name:      ep.GetName(),
		}, svc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			util.ServiceLog.Error(err, "fail to get service, skip reconcile endpoint", "service", util.Key(ep))
		}
		return false
	}

	if !helper.NeedCLB(svc) {
		// it is safe not to reconcile endpoints which belongs to the non-loadbalancer svc
		util.ServiceLog.V(5).Info("endpoint change: loadBalancer is not needed, skip",
			"endpoint", util.Key(ep))
		return false
	}
	return true
}

// NewEnqueueRequestForNodeEvent, event handler for node event
func NewEnqueueRequestForNodeEvent(record record.EventRecorder) *enqueueRequestForNodeEvent {
	return &enqueueRequestForNodeEvent{eventRecorder: record}
}

type enqueueRequestForNodeEvent struct {
	client        client.Client
	eventRecorder record.EventRecorder
}

var _ handler.EventHandler = (*enqueueRequestForNodeEvent)(nil)

func (h *enqueueRequestForNodeEvent) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

func (h *enqueueRequestForNodeEvent) Create(e event.CreateEvent, queue workqueue.RateLimitingInterface) {
	node, ok := e.Object.(*v1.Node)
	if ok && !canNodeSkipEventHandler(node) {
		util.ServiceLog.Info("controller: node create event", "node", node.Name)
		h.enqueueManagedNode(queue, node)
	}
}

func (h *enqueueRequestForNodeEvent) Update(e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	oldNode, ok1 := e.ObjectOld.(*v1.Node)
	newNode, ok2 := e.ObjectNew.(*v1.Node)

	if ok1 && ok2 {
		if canNodeSkipEventHandler(oldNode) && canNodeSkipEventHandler(newNode) {
			return
		}

		//if node label and schedulable condition changed, need to reconcile svc
		if nodeSpecChanged(oldNode, newNode) {
			util.ServiceLog.Info("controller: node update event", "node", oldNode.Name)
			h.enqueueManagedNode(queue, newNode)
		}
	}
}

func (h *enqueueRequestForNodeEvent) Delete(e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	node, ok := e.Object.(*v1.Node)
	if ok && !canNodeSkipEventHandler(node) {
		util.ServiceLog.Info("controller: node delete event", "node", node.Name)
		h.enqueueManagedNode(queue, node)
	}
}

func (h *enqueueRequestForNodeEvent) Generic(e event.GenericEvent, queue workqueue.RateLimitingInterface) {
	// unknown event, ignore
}

func (h *enqueueRequestForNodeEvent) enqueueManagedNode(queue workqueue.RateLimitingInterface, node *v1.Node) {

	// node change would cause all service object reconcile
	svcs := v1.ServiceList{}
	err := h.client.List(context.TODO(), &svcs)
	if err != nil {
		util.ServiceLog.Error(err, "fail to list services for node",
			"node", node.Name)
		return
	}

	for _, v := range svcs.Items {
		if !helper.NeedCLB(&v) {
			continue
		}
		queue.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: v.Namespace,
				Name:      v.Name,
			},
		})
		util.ServiceLog.Info(fmt.Sprintf("node change: enqueue service %s", util.Key(&v)),
			"node", node.Name, "queueLen", queue.Len())
	}
}

// NewEnqueueRequestForEndpointSliceEvent, event handler for endpointslice event
func NewEnqueueRequestForEndpointSliceEvent(record record.EventRecorder) *enqueueRequestForEndpointSliceEvent {
	return &enqueueRequestForEndpointSliceEvent{eventRecorder: record}
}

type enqueueRequestForEndpointSliceEvent struct {
	client        client.Client
	eventRecorder record.EventRecorder
}

var _ handler.EventHandler = (*enqueueRequestForEndpointSliceEvent)(nil)

func (h *enqueueRequestForEndpointSliceEvent) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

func (h *enqueueRequestForEndpointSliceEvent) Create(e event.CreateEvent, queue workqueue.RateLimitingInterface) {
	es, ok := e.Object.(*discovery.EndpointSlice)
	if ok && isEndpointSliceProcessNeeded(es, h.client) {
		util.ServiceLog.Info("controller: endpointslice create event", "endpointslice", util.Key(es))
		h.enqueueManagedEndpointSlice(queue, es)
	}
}

func (h *enqueueRequestForEndpointSliceEvent) Update(e event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	es1, ok1 := e.ObjectOld.(*discovery.EndpointSlice)
	es2, ok2 := e.ObjectNew.(*discovery.EndpointSlice)

	if ok1 && ok2 && isEndpointSliceProcessNeeded(es1, h.client) &&
		isEndpointSliceUpdateNeeded(es1, es2) {
		util.ServiceLog.Info("controller: endpointslice update event", "endpointslice", util.Key(es1))
		util.ServiceLog.Info(fmt.Sprintf("endpoints before [%s], afeter [%s]",
			helper.LogEndpointSlice(es1), helper.LogEndpointSlice(es2)), "endpointslice", util.Key(es1))
		h.enqueueManagedEndpointSlice(queue, es1)
	}
}

func (h *enqueueRequestForEndpointSliceEvent) Delete(e event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	es, ok := e.Object.(*discovery.EndpointSlice)
	if ok && isEndpointSliceProcessNeeded(es, h.client) {
		util.ServiceLog.Info("controller: endpointslice delete event", "endpointslice", util.Key(es))
		h.enqueueManagedEndpointSlice(queue, es)
	}
}

func (h *enqueueRequestForEndpointSliceEvent) Generic(e event.GenericEvent, queue workqueue.RateLimitingInterface) {
	// unknown event, ignore
}

func (h *enqueueRequestForEndpointSliceEvent) enqueueManagedEndpointSlice(queue workqueue.RateLimitingInterface, endpointSlice *discovery.EndpointSlice) {
	serviceName, ok := endpointSlice.Labels[discovery.LabelServiceName]
	if !ok {
		return
	}

	queue.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: endpointSlice.Namespace,
			Name:      serviceName,
		},
	})

	util.ServiceLog.Info("enqueue", "endpointslice", util.Key(endpointSlice), "queueLen", queue.Len())
}

func isEndpointSliceProcessNeeded(es *discovery.EndpointSlice, client client.Client) bool {
	if es == nil {
		return false
	}

	serviceName, ok := es.Labels[discovery.LabelServiceName]
	if !ok {
		return false
	}

	svc := &v1.Service{}
	err := client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: es.Namespace,
			Name:      serviceName,
		}, svc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			util.ServiceLog.Error(err, "fail to get service, skip reconcile endpointslice",
				"endpointslice", util.Key(es), "service", serviceName)
		}
		return false
	}

	if !helper.NeedCLB(svc) {
		// it is safe not to reconcile endpointslice which belongs to the non-loadbalancer svc
		util.ServiceLog.V(5).Info("endpointslice change: loadBalancer is not needed, skip",
			"endpointslice", util.Key(es))
		return false
	}
	return true
}

func isEndpointSliceUpdateNeeded(old, new *discovery.EndpointSlice) bool {
	return !reflect.DeepEqual(old.Endpoints, new.Endpoints) || !reflect.DeepEqual(old.Ports, new.Ports)
}

func nodeSpecChanged(oldNode, newNode *v1.Node) bool {
	if nodeLabelsChanged(oldNode.Name, oldNode.Labels, newNode.Labels) {
		return true
	}
	if oldNode.Spec.Unschedulable != newNode.Spec.Unschedulable {
		util.ServiceLog.Info(fmt.Sprintf(
			"node changed: %s, spec from=%t, to=%t",
			oldNode.Name, oldNode.Spec.Unschedulable, newNode.Spec.Unschedulable),
			"node", oldNode.Name)
		return true
	}

	if nodeReadyChanged(oldNode, newNode) {
		return true
	}

	return false
}

func nodeReadyChanged(oldNode, newNode *v1.Node) bool {
	oldNodeReadyCondition := v1.ConditionFalse
	newNodeReadyCondition := v1.ConditionFalse
	if oldNode != nil {
		if readyCond := helper.GetNodeCondition(oldNode, v1.NodeReady); readyCond != nil {
			oldNodeReadyCondition = readyCond.Status
		}
	}
	if newNode != nil {
		if readyCond := helper.GetNodeCondition(newNode, v1.NodeReady); readyCond != nil {
			newNodeReadyCondition = readyCond.Status
		}
	}
	if oldNodeReadyCondition != newNodeReadyCondition {
		util.ServiceLog.Info(fmt.Sprintf(
			"node changed: %s, ready condition from=%s, to=%s. old condition [%v], new condition[%v]",
			oldNode.Name, oldNodeReadyCondition, newNodeReadyCondition, oldNode.Status.Conditions, newNode.Status.Conditions),
			"node", oldNode.Name)
		return true
	}

	return false
}

func nodeLabelsChanged(nodeName string, oldL, newL map[string]string) bool {
	if len(oldL) != len(newL) {
		util.ServiceLog.Info(fmt.Sprintf("node changed: label size not equal, from=%v, to=%v", oldL, newL),
			"node", nodeName)
		return true
	}
	for k, v := range oldL {
		if newL[k] != v {
			util.ServiceLog.Info(fmt.Sprintf("node changed: label key=%s, value from=%v, to=%v",
				k, oldL[k], newL[k]),
				"node", nodeName)
			return true
		}
	}
	// no need for reverse compare
	return false
}

// only for node event
func canNodeSkipEventHandler(node *v1.Node) bool {
	if node == nil || node.Labels == nil {
		return false
	}

	if helper.HasExcludeLabel(node) {
		klog.V(5).Infof("node %s has exclude label, skip", node.Name)
		return true
	}
	if helper.IsMasterNode(node) {
		klog.V(5).Infof("node %s is master node, skip", node.Name)
		return true
	}
	return false
}
