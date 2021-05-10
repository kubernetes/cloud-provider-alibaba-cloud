package controller

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	//Adds = append(Adds, node.Add)
	//Adds = append(Adds, route.Add)
	Adds = append(Adds, service.Add)
//	Adds = append(Adds, ingress.Add)
//	Adds = append(Adds, pvtz.Add)
}

// Adds is a list of functions to add all Controllers to the Manager
var Adds []func(manager.Manager, *shared.SharedContext) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, ctx *shared.SharedContext) error {

	for _, f := range Adds {
		if err := f(m, ctx); err != nil {
			return err
		}
	}
	return nil
}
