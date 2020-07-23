package service

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type Context struct{ ctx sync.Map }

func (c *Context) Get(name string) *v1.Service {
	v, ok := c.ctx.Load(name)
	if !ok {
		return nil
	}
	val, ok := v.(*v1.Service)
	if !ok {
		klog.Errorf("not type of v1.svc: [%s]", reflect.TypeOf(v))
		return nil
	}
	return val
}

func (c *Context) Set(name string, val *v1.Service) { c.ctx.Store(name, val) }

func (c *Context) Range(f func(key string, value *v1.Service) bool) {
	c.ctx.Range(
		func(key, value interface{}) bool {
			return f(key.(string), value.(*v1.Service))
		},
	)
}
func (c *Context) Remove(name string) { c.ctx.Delete(name) }

// NeedUpdate compare old and new service for possible changes
func NeedUpdate(old, newm *v1.Service, record record.EventRecorder) bool {
	if !NeedLoadBalancer(old) &&
		!NeedLoadBalancer(newm) {
		// no loadbalancer is needed
		return false
	}
	if NeedLoadBalancer(old) != NeedLoadBalancer(newm) {
		record.Eventf(
			newm,
			v1.EventTypeNormal,
			"TypeChanged",
			"%v -> %v",
			old.Spec.Type,
			newm.Spec.Type,
		)
		return true
	}

	if !reflect.DeepEqual(old.Annotations, newm.Annotations) {
		klog.Infof("AnnotationChanged: %v -> %v", old.Annotations, newm.Annotations)
		record.Eventf(
			newm,
			v1.EventTypeNormal,
			"AnnotationChanged",
			"The service will be updated because the annotations has been changed.",
		)
		return true
	}
	if old.UID != newm.UID {
		klog.Infof("UIDChanged: %v -> %v", old.UID, newm.UID)
		return true
	}
	if !reflect.DeepEqual(old.Spec, newm.Spec) {
		klog.Infof("ServiceSpecChanged: %v -> %v", old.Spec, newm.Spec)
		record.Eventf(
			newm,
			v1.EventTypeNormal,
			"ServiceSpecChanged",
			"The service will be updated because the spec has been changed.",
		)
		return true
	}

	return false
}

func NeedLoadBalancer(service *v1.Service) bool {
	return service.Spec.Type == v1.ServiceTypeLoadBalancer
}

func NodeSpecChanged(a, b *v1.Node) bool {
	if NodeLabelsChanged(a.Labels, b.Labels) {
		// log node label details for debug convenience
		klog.Infof("node label changed: %s, from=%v, to=%v", a.Name, a.Labels, b.Labels)
		return true
	}
	if a.Spec.Unschedulable != b.Spec.Unschedulable {
		klog.Infof(
			"spec.Unscheduleable changed: %s, from=%t, to=%t",
			a.Name, a.Spec.Unschedulable, b.Spec.Unschedulable,
		)
		return true
	}
	if NodeConditionChanged(a.Name, a.Status.Conditions, b.Status.Conditions) {
		klog.Infof(
			"node condition changed: %s, from=%d, to=%d",
			a.Name, len(a.Status.Conditions), len(b.Status.Conditions),
		)
		return true
	}
	return false
}

func NodeConditionChanged(name string, a, b []v1.NodeCondition) bool {
	if len(a) != len(b) {
		klog.Infof("Node Change: node condition changed, before %v, after %v", a, b)
		return true
	}

	sort.SliceStable(a, func(i, j int) bool {
		return strings.Compare(string(a[i].Type), string(a[j].Type)) <= 0
	})

	sort.SliceStable(b, func(i, j int) bool {
		return strings.Compare(string(b[i].Type), string(b[j].Type)) <= 0
	})

	for i := range a {
		if a[i].Type != b[i].Type ||
			a[i].Status != b[i].Status {
			klog.Infof(
				"node condition changed: %s, type(%s,%s) | status(%s,%s)",
				name, a[i].Type, b[i].Type, a[i].Status, b[i].Status,
			)
			return true
		}
	}
	return false
}

func NodeLabelsChanged(a, b map[string]string) bool {
	if len(a) != len(b) {
		klog.Infof("Node Change: the number of labels on the node are changed, before %v, after %v", a, b)
		return true
	}
	for k, v := range a {
		if b[k] != v {
			klog.Infof("Node Change: the %v label on the node are changed, before %v, after %v", k, a[k], b[k])
			return true
		}
	}
	// no need for reverse compare
	return false
}
