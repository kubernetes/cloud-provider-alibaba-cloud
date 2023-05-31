package pvtz

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ServicePredicate struct {
	predicate.Funcs
}

func (sp *ServicePredicate) filterLeaseEvents(obj client.Object) bool {
	empty := sets.Empty{}
	avoid := sets.Set[string]{
		"kube-system/kube-scheduler":          empty,
		"kube-system/ccm":                     empty,
		"kube-system/kube-controller-manager": empty,
	}
	target := fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
	return !avoid.Has(target)
}

func (sp *ServicePredicate) Create(e event.CreateEvent) bool {
	return sp.filterLeaseEvents(e.Object)
}

func (sp *ServicePredicate) Update(e event.UpdateEvent) bool {
	return sp.filterLeaseEvents(e.ObjectOld)
}
