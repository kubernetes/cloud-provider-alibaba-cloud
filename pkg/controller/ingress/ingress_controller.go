package ingress

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/shared"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Add(mgr manager.Manager, ctx *shared.SharedContext) error {
	albconfigReconciler, err := NewAlbConfigReconciler(mgr, ctx)
	if err != nil {
		return err
	}
	return albconfigReconciler.SetupWithManager(context.Background(), mgr)
}
