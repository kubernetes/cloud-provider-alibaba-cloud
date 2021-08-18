package service

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
	"strings"
)

func NewMapHandler() *MapEnqueue {
	omap := &MapEnqueue{}
	omap.EventHandler = handler.EnqueueRequestsFromMapFunc(omap.FanIN)
	return omap
}

type MapEnqueue struct {
	client client.Client
	handler.EventHandler
}

// FanIN aggregate multiple resource notification into service update.
func (m *MapEnqueue) FanIN(o client.Object) []reconcile.Request {

	var request []reconcile.Request
	switch o.(type) {
	case *v1.Node:
		request = append(request, m.mapNode(o.(*v1.Node))...)
	case *v1.Endpoints:
		request = append(request, m.mapEndpoint(o.(*v1.Endpoints))...)
	case *v1.Service:
		request = append(request, m.mapService(o.(*v1.Service))...)
	default:
		util.ServiceLog.Info(fmt.Sprintf("warning: unknown object: %s, %v", reflect.TypeOf(o), o))
	}
	return request
}

func (m *MapEnqueue) mapNode(o *v1.Node) []reconcile.Request {
	var request []reconcile.Request
	// node change would cause all service object reconcile
	svcs := v1.ServiceList{}
	err := m.client.List(context.TODO(), &svcs)
	if err != nil {
		util.ServiceLog.Error(err, fmt.Sprintf("fail to list services for node"),
			"node", o.Name)
		return request
	}

	for _, v := range svcs.Items {
		if !needLoadBalancer(&v) {
			continue
		}
		if !isProcessNeeded(&v) {
			continue
		}
		util.ServiceLog.Info(fmt.Sprintf("node change: enqueue service %s", util.Key(&v)),
			"node", o.Name)
		req := reconcile.Request{
			NamespacedName: client.ObjectKey{Namespace: v.Namespace, Name: v.Name},
		}
		request = append(request, req)
	}
	return request
}

func (m *MapEnqueue) mapEndpoint(o *v1.Endpoints) []reconcile.Request {

	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (m *MapEnqueue) mapService(o *v1.Service) []reconcile.Request {
	if !isProcessNeeded(o) {
		util.ServiceLog.Info("ccm class not empty, skip process", "service", util.Key(o))
		return nil
	}
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (m *MapEnqueue) InjectClient(c client.Client) error {
	m.client = c
	return nil
}

// PredicateForServiceEvent, filter service event
func NewPredicateForServiceEvent(eventRecorder record.EventRecorder) *predicateForServiceEvent {
	return &predicateForServiceEvent{eventRecorder: eventRecorder}
}

type predicateForServiceEvent struct {
	eventRecorder record.EventRecorder
}

var _ predicate.Predicate = (*predicateForServiceEvent)(nil)

func (p *predicateForServiceEvent) Create(e event.CreateEvent) bool {
	svc, ok := e.Object.(*v1.Service)
	if ok && needAdd(svc) {
		util.ServiceLog.Info("controller: service create event", "service", util.Key(svc))
		return true
	}
	return false
}

func (p *predicateForServiceEvent) Update(e event.UpdateEvent) bool {
	oldSvc, ok1 := e.ObjectOld.(*v1.Service)
	newSvc, ok2 := e.ObjectNew.(*v1.Service)

	if ok1 && ok2 && needUpdate(oldSvc, newSvc, p.eventRecorder) {
		util.ServiceLog.Info("controller: service update event",
			"service", util.Key(oldSvc))
		return true
	}
	return false
}

func (p *predicateForServiceEvent) Delete(e event.DeleteEvent) bool {
	return false
}

func (p *predicateForServiceEvent) Generic(event.GenericEvent) bool {
	return false
}

func needUpdate(oldSvc, newSvc *v1.Service, recorder record.EventRecorder) bool {
	if !needLoadBalancer(oldSvc) && !needLoadBalancer(newSvc) {
		return false
	}

	if needLoadBalancer(oldSvc) != needLoadBalancer(newSvc) {
		util.ServiceLog.Info(fmt.Sprintf("TypeChanged %v - %v", oldSvc.Spec.Type, newSvc.Spec.Type),
			"service", util.Key(oldSvc))
		recorder.Eventf(
			newSvc,
			v1.EventTypeNormal,
			helper.TypeChanged,
			"%v - %v",
			oldSvc.Spec.Type,
			newSvc.Spec.Type,
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
		recorder.Eventf(
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
		recorder.Eventf(
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
		recorder.Eventf(
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
	if needLoadBalancer(newService) {
		return true
	}

	// was LoadBalancer
	_, ok := newService.Labels[LabelServiceHash]
	if ok {
		util.ServiceLog.Info("service has hash label, which may was LoadBalancer", "service", util.Key(newService))
		return true
	}
	return false
}

// PredicateForEndpointEvent, filter endpoint event
func NewPredicateForEndpointEvent(client client.Client) *predicateForEndpointEvent {
	return &predicateForEndpointEvent{client}
}

type predicateForEndpointEvent struct {
	client client.Client
}

var _ predicate.Predicate = (*predicateForEndpointEvent)(nil)

func (p *predicateForEndpointEvent) Create(e event.CreateEvent) bool {
	ep, ok := e.Object.(*v1.Endpoints)
	if ok && isEndpointProcessNeeded(ep, p.client) {
		util.ServiceLog.Info("controller: endpoint create event", "endpoint", util.Key(ep))
		return true
	}
	return false
}

func (p *predicateForEndpointEvent) Update(e event.UpdateEvent) bool {
	ep1, ok1 := e.ObjectOld.(*v1.Endpoints)
	ep2, ok2 := e.ObjectNew.(*v1.Endpoints)

	if ok1 && ok2 &&
		isEndpointProcessNeeded(ep1, p.client) &&
		!reflect.DeepEqual(ep1.Subsets, ep2.Subsets) {
		util.ServiceLog.Info("controller: endpoint update event", "endpoint", util.Key(ep1))
		util.ServiceLog.Info(fmt.Sprintf("endpoints before [%s], afeter [%s]", LogEndpoints(*ep1), LogEndpoints(*ep2)), "endpoint", util.Key(ep1))
		return true
	}
	return false
}

func (p *predicateForEndpointEvent) Delete(e event.DeleteEvent) bool {
	ep, ok := e.Object.(*v1.Endpoints)
	if ok && isEndpointProcessNeeded(ep, p.client) {
		util.ServiceLog.Info("controller: endpoint delete event", "endpoint", util.Key(ep))
		return true
	}
	return false
}

func (p *predicateForEndpointEvent) Generic(event.GenericEvent) bool {
	return false
}

func isProcessNeeded(svc *v1.Service) bool { return svc.Annotations[CCMClass] == "" }

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
		if apierrors.IsNotFound(err) {
			util.ServiceLog.Error(err, "fail to get service, skip reconcile", "service", util.Key(ep))
		}
		return false
	}

	if !isProcessNeeded(svc) {
		util.ServiceLog.Info("endpoint change: class not empty, skip reconcile",
			"endpoint", util.Key(ep))
		return false
	}

	if !needLoadBalancer(svc) {
		// we are safe here to skip process syncEndpoint
		util.ServiceLog.V(5).Info("endpoint change: loadBalancer is not needed, skip",
			"endpoint", util.Key(ep))
		return false
	}
	return true
}

// PredicateForNodeEvent, filter node event
func NewPredicateForNodeEvent(record record.EventRecorder) *predicateForNodeEvent {
	return &predicateForNodeEvent{eventRecorder: record}
}

type predicateForNodeEvent struct {
	eventRecorder record.EventRecorder
}

var _ predicate.Predicate = (*predicateForNodeEvent)(nil)

func (p *predicateForNodeEvent) Create(e event.CreateEvent) bool {
	node, ok := e.Object.(*v1.Node)
	if ok && !canNodeSkipEventHandler(node) {
		util.ServiceLog.Info("controller: node create event", "node", node.Name)
		return true
	}
	return false
}

func (p *predicateForNodeEvent) Update(e event.UpdateEvent) bool {
	oldNode, ok1 := e.ObjectOld.(*v1.Node)
	newNode, ok2 := e.ObjectNew.(*v1.Node)

	if ok1 && ok2 {
		if canNodeSkipEventHandler(oldNode) && canNodeSkipEventHandler(newNode) {
			return false
		}

		//if node label and schedulable condition changed, need to reconcile svc
		if nodeSpecChanged(oldNode, newNode) {
			util.ServiceLog.Info("controller: node update event", "node", oldNode.Name)
			return true
		}
	}
	return false
}

func (p *predicateForNodeEvent) Delete(e event.DeleteEvent) bool {
	node, ok := e.Object.(*v1.Node)
	if ok && !canNodeSkipEventHandler(node) {
		util.ServiceLog.Info("controller: node delete event", "node", node.Name)
		return true
	}
	return false
}

func (p *predicateForNodeEvent) Generic(event.GenericEvent) bool {
	return false
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
	if nodeConditionChanged(oldNode.Name, oldNode.Status.Conditions, newNode.Status.Conditions) {
		return true
	}
	return false
}

func nodeConditionChanged(name string, oldC, newC []v1.NodeCondition) bool {
	if len(oldC) != len(newC) {
		util.ServiceLog.Info(fmt.Sprintf("node changed:  condition length not equal, from=%v, to=%v", oldC, newC),
			"node", name)
		return true
	}

	sort.SliceStable(oldC, func(i, j int) bool {
		return strings.Compare(string(oldC[i].Type), string(oldC[j].Type)) <= 0
	})

	sort.SliceStable(newC, func(i, j int) bool {
		return strings.Compare(string(newC[i].Type), string(newC[j].Type)) <= 0
	})

	for i := range oldC {
		if oldC[i].Type != newC[i].Type ||
			oldC[i].Status != newC[i].Status {
			util.ServiceLog.Info(
				fmt.Sprintf("node changed: condition type(%s,%s) | status(%s,%s)",
					oldC[i].Type, newC[i].Type, oldC[i].Status, newC[i].Status),
				"node", name)
			return true
		}
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
