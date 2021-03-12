package controller

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/privatezone"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/route"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	Adds = append(Adds, node.Add)
	Adds = append(Adds, route.Add)
	Adds = append(Adds, privatezone.Add)
	Adds = append(Adds, service.Add)
}

// Adds is a list of functions to add all Controllers to the Manager
var Adds []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {

	for _, f := range Adds {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}
