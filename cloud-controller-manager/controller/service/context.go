package service

import (
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
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
		glog.Errorf("not type of v1.svc: [%s]", reflect.TypeOf(v))
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
		record.Eventf(
			newm,
			v1.EventTypeNormal,
			"AnnotationChanged",
			"with count %v -> %v",
			len(old.Annotations),
			len(newm.Annotations),
		)
		return true
	}
	if old.UID != newm.UID {
		record.Eventf(
			newm,
			v1.EventTypeNormal,
			"UIDChanged",
			"%v -> %v",
			old.UID, newm.UID,
		)
		return true
	}
	if !reflect.DeepEqual(old.Spec, newm.Spec) {
		record.Eventf(
			newm,
			v1.EventTypeNormal,
			"ServiceSpecChanged",
			"%v -> %v",
			old.Spec,
			newm.Spec,
		)
		return true
	}

	return false
}

func NeedLoadBalancer(service *v1.Service) bool {
	return service.Spec.Type == v1.ServiceTypeLoadBalancer
}

func NodeSpecChanged(a, b *v1.Node) bool {
	return NodeLabelsChanged(a.Labels, b.Labels) ||
		a.Spec.Unschedulable != b.Spec.Unschedulable ||
		NodeConditionChanged(a.Status.Conditions, b.Status.Conditions)
}

func NodeConditionChanged(a, b []v1.NodeCondition) bool {
	if len(a) != len(b) {
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
			return true
		}
	}
	return false
}

func NodeLabelsChanged(a, b map[string]string) bool {
	if len(a) != len(b) {
		return true
	}
	for k, v := range a {
		if b[k] != v {
			return true
		}
	}
	// no need for reverse compare
	return false
}
