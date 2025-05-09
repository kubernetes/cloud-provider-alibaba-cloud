package route

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type predicateForNodeEvent struct {
	predicate.Funcs
}

func (sp *predicateForNodeEvent) Create(e event.CreateEvent) bool {
	node, ok := e.Object.(*v1.Node)
	if ok && node.Spec.PodCIDR == "" {
		klog.V(5).Infof("node podCIDR is empty, ignore create event")
		return false
	}

	return true
}

func (sp *predicateForNodeEvent) Update(e event.UpdateEvent) bool {
	oldNode, ok1 := e.ObjectOld.(*v1.Node)
	newNode, ok2 := e.ObjectNew.(*v1.Node)
	if ok1 && ok2 {
		if oldNode.UID != newNode.UID {
			klog.Infof("node changed: %s UIDChanged: %v - %v", oldNode.Name, oldNode.UID, newNode.UID)
			return true
		}
		if oldNode.Spec.PodCIDR != newNode.Spec.PodCIDR {
			klog.Infof("node changed: %s Pod CIDR Changed: %v - %v", oldNode.Name, oldNode.Spec.PodCIDR, newNode.Spec.PodCIDR)
			return true
		}
		if !reflect.DeepEqual(oldNode.Spec.PodCIDRs, oldNode.Spec.PodCIDRs) {
			klog.Infof("node changed: %s Pod CIDRs Changed: %v - %v", oldNode.Name, oldNode.Spec.PodCIDRs, newNode.Spec.PodCIDRs)
			return true
		}
		return false
	}
	return true
}
