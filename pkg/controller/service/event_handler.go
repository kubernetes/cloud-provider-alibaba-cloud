package service

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
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
		klog.Warningf("unknown object: %s, %v", reflect.TypeOf(o), o)
	}
	return dropLeaseEndpoint(request)
}

func (m *MapEnqueue) mapNode(o *v1.Node) []reconcile.Request {
	var request []reconcile.Request
	// node change would cause all service object reconcile
	svcs := v1.ServiceList{}
	err := m.client.List(context.TODO(), &svcs)
	if err != nil {
		klog.Errorf("list service cache for node: %s, %s", o, err.Error())
		return request
	}

	for _, v := range svcs.Items {
		if !needLoadBalancer(&v) {
			continue
		}
		if !isProcessNeeded(&v) {
			continue
		}
		klog.Infof("%s node change: enqueue service.", util.Key(&v))
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
		klog.Infof("%s ccm class not empty, skip process", util.Key(o))
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

func dropLeaseEndpoint(req []reconcile.Request) []reconcile.Request {
	e := sets.Empty{}
	avoid := sets.String{
		"kube-system/kube-scheduler":          e,
		"kube-system/ccm":                     e,
		"kube-system/kube-controller-manager": e,
	}
	var reqs []reconcile.Request
	for _, r := range req {
		if avoid.Has(r.String()) {
			continue
		}
		reqs = append(reqs, r)
	}
	return reqs
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
		klog.Infof("controller: %s service create event", util.Key(svc))
		return true
	}
	return false
}

func (p *predicateForServiceEvent) Update(e event.UpdateEvent) bool {
	oldSvc, ok1 := e.ObjectOld.(*v1.Service)
	newSvc, ok2 := e.ObjectNew.(*v1.Service)

	if ok1 && ok2 && needUpdate(oldSvc, newSvc, p.eventRecorder) {
		klog.Infof("controller: %s service update event", util.Key(oldSvc))
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
		recorder.Eventf(
			newSvc,
			v1.EventTypeNormal,
			helper.TypeChanged,
			"%v->%v",
			oldSvc.Spec.Type,
			newSvc.Spec.Type,
		)
		return true
	}

	if oldSvc.UID != newSvc.UID {
		klog.Info("%s UIDChanged: %v -> %v", util.Key(oldSvc), oldSvc.UID, newSvc.UID)
		return true
	}

	if !reflect.DeepEqual(oldSvc.Annotations, newSvc.Annotations) {
		klog.Infof("%s AnnotationChanged: %v -> %v", util.Key(oldSvc), oldSvc.Annotations, newSvc.Annotations)
		recorder.Eventf(
			newSvc,
			v1.EventTypeNormal,
			helper.AnnoChanged,
			"The service will be updated because the annotations has been changed.",
		)
		return true
	}

	if !reflect.DeepEqual(oldSvc.Spec, newSvc.Spec) {
		klog.Infof("%s SpecChanged: %v -> %v", util.Key(oldSvc), oldSvc.Spec, newSvc.Spec)
		recorder.Eventf(
			newSvc,
			v1.EventTypeNormal,
			helper.SpecChanged,
			"The service will be updated because the spec has been changed.",
		)
		return true
	}

	if !reflect.DeepEqual(oldSvc.DeletionTimestamp.IsZero(), newSvc.DeletionTimestamp.IsZero()) {
		klog.Infof("%s DeleteTimestampChanged: %v -> %v", util.Key(oldSvc),
			oldSvc.DeletionTimestamp.IsZero(), newSvc.DeletionTimestamp.IsZero())
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
		klog.Infof("%s has hash label, which may was LoadBalancer", util.Key(newService))
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
		klog.Infof("controller: %s endpoint create event", util.NamespacedName(ep).String())
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
		klog.Infof("controller: %s endpoint update event, before [%v], after [%v]",
			util.NamespacedName(ep1).String(), ep1.Subsets, ep2.Subsets)
		return true
	}
	return false
}

func (p *predicateForEndpointEvent) Delete(e event.DeleteEvent) bool {
	ep, ok := e.Object.(*v1.Endpoints)
	return ok && isEndpointProcessNeeded(ep, p.client)
}

func (p *predicateForEndpointEvent) Generic(event.GenericEvent) bool {
	return false
}

func isProcessNeeded(svc *v1.Service) bool { return svc.Annotations[CCMClass] == "" }

func isEndpointProcessNeeded(ep *v1.Endpoints, client client.Client) bool {
	if ep == nil {
		return false
	}

	svc := &v1.Service{}
	err := client.Get(context.TODO(),
		types.NamespacedName{
			Namespace: ep.GetNamespace(),
			Name:      ep.GetName(),
		}, svc)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			klog.Warningf("can not get service %s/%s, error: %s", ep.Namespace, ep.Name, err.Error())
		}
		return false
	}

	if !isProcessNeeded(svc) {
		klog.Infof("endpoint change: class not empty, skip process")
		return false
	}

	if !needLoadBalancer(svc) {
		// we are safe here to skip process syncEnpoint.
		klog.V(5).Infof("endpoint change: loadBalancer is not needed, skip")
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
	if ok && !isExcludeNode(node) {
		klog.Infof("controller: node %s create event", e.Object.GetName())
		return true
	}
	return false
}

func (p *predicateForNodeEvent) Update(e event.UpdateEvent) bool {
	oldNode, ok1 := e.ObjectOld.(*v1.Node)
	newNode, ok2 := e.ObjectNew.(*v1.Node)
	if ok1 && ok2 &&
		nodeSpecChanged(oldNode, newNode) {
		// label and schedulable changed .
		// status healthy should be considered
		klog.Infof("controller: node %s update event", oldNode.Namespace, oldNode.Name)
		return true
	}
	return false
}

func (p *predicateForNodeEvent) Delete(e event.DeleteEvent) bool {
	node, ok := e.Object.(*v1.Node)
	if ok && !isExcludeNode(node) {
		klog.Infof("controller: node %s delete event", e.Object.GetName())
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
		klog.Infof(
			"node spec changed: %s, from=%t, to=%t",
			oldNode.Name, oldNode.Spec.Unschedulable, newNode.Spec.Unschedulable,
		)
		return true
	}
	if nodeConditionChanged(oldNode.Name, oldNode.Status.Conditions, newNode.Status.Conditions) {
		return true
	}
	return false
}

func nodeConditionChanged(name string, oldC, newC []v1.NodeCondition) bool {
	if len(oldC) != len(newC) {
		klog.Infof("node condition changed: %s, length not equal, from=%v, to=%v", name, oldC, newC)
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
			klog.Infof(
				"node condition changed: %s, type(%s,%s) | status(%s,%s)",
				name, oldC[i].Type, newC[i].Type, oldC[i].Status, newC[i].Status,
			)
			return true
		}
	}
	return false
}

func nodeLabelsChanged(nodeName string, oldL, newL map[string]string) bool {
	if len(oldL) != len(newL) {
		klog.Infof("node label changed: %s, size not equal, from=%v, to=%v", nodeName, oldL, newL)
		return true
	}
	for k, v := range oldL {
		if newL[k] != v {
			klog.Infof("node label changed: %s, key=%s, value from=%v, to=%v", nodeName, k, oldL[k], newL[k])
			return true
		}
	}
	// no need for reverse compare
	return false
}
