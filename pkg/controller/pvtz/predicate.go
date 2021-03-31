package pvtz

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type PredicateDNS struct {
	predicate.Funcs
}

func (pd *PredicateDNS) filterLeaseEvents(obj client.Object) bool {
	empty := sets.Empty{}
	avoid := sets.String{
		"kube-system/kube-scheduler":          empty,
		"kube-system/ccm":                     empty,
		"kube-system/kube-controller-manager": empty,
	}
	target := fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
	if avoid.Has(target) {
		return false
	}
	return true
}

func (pd *PredicateDNS) Create(e event.CreateEvent) bool {
	return pd.filterLeaseEvents(e.Object)
}

func (pd *PredicateDNS) Update(e event.UpdateEvent) bool {
	return pd.filterLeaseEvents(e.ObjectOld)
}
