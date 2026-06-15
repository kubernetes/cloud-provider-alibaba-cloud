package route

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type predicateForNodeEvent struct {
	predicate.TypedFuncs[*v1.Node]
}

func (sp *predicateForNodeEvent) Create(e event.TypedCreateEvent[*v1.Node]) bool {
	node := e.Object
	if node.Spec.PodCIDR == "" {
		klog.V(5).Infof("node %s podCIDR is empty, ignore create event", node.Name)
		return false
	}

	return true
}

func (sp *predicateForNodeEvent) Update(e event.TypedUpdateEvent[*v1.Node]) bool {
	oldNode := e.ObjectOld
	newNode := e.ObjectNew
	if oldNode.UID != newNode.UID {
		klog.Infof("node changed: %s UIDChanged: %v - %v", oldNode.Name, oldNode.UID, newNode.UID)
		return true
	}
	if oldNode.Spec.PodCIDR != newNode.Spec.PodCIDR {
		klog.Infof("node changed: %s Pod CIDR Changed: %v - %v", oldNode.Name, oldNode.Spec.PodCIDR, newNode.Spec.PodCIDR)
		return true
	}
	if !reflect.DeepEqual(oldNode.Spec.PodCIDRs, newNode.Spec.PodCIDRs) {
		klog.Infof("node changed: %s Pod CIDRs Changed: %v - %v", oldNode.Name, oldNode.Spec.PodCIDRs, newNode.Spec.PodCIDRs)
		return true
	}
	return false
}
