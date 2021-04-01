package service

import (
	"context"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util/hash"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
		log.Warnf("unknown object: %s, %v", reflect.TypeOf(o), o)
	}
	return dropLeaseEndpoint(request)
}

func (m *MapEnqueue) mapNode(o *v1.Node) []reconcile.Request {
	var request []reconcile.Request
	// 1. node label & condition filter.
	if !isHashChanged(o) {
		return []reconcile.Request{}
	}

	// node change would cause all service object reconcile
	svcs := v1.ServiceList{}
	err := m.client.List(context.TODO(), &svcs)
	if err != nil {
		log.Errorf("list service cache for node: %s, %s", o, err.Error())
		return request
	}

	for _, v := range svcs.Items {
		// 2. need to filter out service which is
		// not LoadBalancer type. And, service change
		// would handle LoadBalancer->NodePort change
		if v.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}
		req := reconcile.Request{
			NamespacedName: client.ObjectKey{Namespace: v.Namespace, Name: v.Name},
		}
		request = append(request, req)
	}
	return request
}

func (m *MapEnqueue) mapEndpoint(o *v1.Endpoints) []reconcile.Request {

	if !isHashChanged(o) {
		return []reconcile.Request{}
	}

	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (m *MapEnqueue) mapService(o *v1.Service) []reconcile.Request {
	if !isHashChanged(o) {
		return []reconcile.Request{}
	}
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}},
	}
}

func (m *MapEnqueue) InjectClient(c client.Client) error {
	m.client = c
	return nil
}

func isHashChanged(o interface{}) bool {
	var (
		op  []interface{}
		lbl map[string]string
	)
	switch o.(type) {
	case *v1.Service:
		n := o.(*v1.Service)
		lbl = n.Labels
		op = append(op, n.Spec, n.Annotations)
	case *v1.Node:
		n := o.(*v1.Node)
		lbl = n.Labels
		op = append(op, n.Status.Conditions, n.Labels, n.Spec.Unschedulable)
	case *v1.Endpoints:
		e := o.(*v1.Endpoints)
		lbl = e.Labels
		op = append(op, e.Subsets, e.Labels)
	}
	return !strings.EqualFold(hash.HashObject(op), getPreHash(lbl))
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

func getPreHash(o map[string]string) string {
	if o == nil {
		return ""
	}
	return o[hash.ReconcileHashLable]
}
