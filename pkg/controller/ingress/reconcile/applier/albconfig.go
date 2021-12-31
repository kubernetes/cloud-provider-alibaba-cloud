package applier

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/tracking"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/store"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/backend"
	albmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type AlbConfigManagerApplier interface {
	Apply(ctx context.Context, stack core.Manager) error
}

var _ AlbConfigManagerApplier = &defaultAlbConfigManagerApplier{}

func NewAlbConfigManagerApplier(store store.Storer, kubeClient client.Client, provider prvd.Provider, tagPrefix string, logger logr.Logger) *defaultAlbConfigManagerApplier {
	trackingProvider := tracking.NewDefaultProvider(tagPrefix, provider.ClusterID())
	backendManager := backend.NewBackendManager(store, kubeClient, provider, logger)
	return &defaultAlbConfigManagerApplier{
		trackingProvider: trackingProvider,
		backendManager:   *backendManager,
		kubeClient:       kubeClient,
		albProvider:      provider,
		logger:           logger,
	}
}

type defaultAlbConfigManagerApplier struct {
	kubeClient client.Client

	trackingProvider tracking.TrackingProvider
	backendManager   backend.Manager
	albProvider      prvd.Provider

	logger logr.Logger
}

type ResourceApply interface {
	Apply(ctx context.Context) error
	PostApply(ctx context.Context) error
}

func (m *defaultAlbConfigManagerApplier) Apply(ctx context.Context, stack core.Manager) error {

	// Reuse LoadBalancer
	var resLBs []*albmodel.AlbLoadBalancer
	stack.ListResources(&resLBs)
	if len(resLBs) > 1 {
		return fmt.Errorf("invalid res loadBalancers, at most one loadBalancer for stack: %s", stack.StackID())
	} else if len(resLBs) == 1 {
		resLb := resLBs[0]
		var isReuseLb bool
		if len(resLb.Spec.LoadBalancerId) != 0 {
			isReuseLb = true
		}
		if isReuseLb {
			if resLb.Spec.ForceOverride != nil && !*resLb.Spec.ForceOverride {
				applier := NewAlbLoadBalancerApplier(m.albProvider, m.trackingProvider, stack, m.logger)
				return applier.Apply(ctx)
			}
		}
	}

	appliers := []ResourceApply{
		NewServerGroupApplier(m.kubeClient, m.backendManager, m.albProvider, m.trackingProvider, stack, m.logger),
		NewAlbLoadBalancerApplier(m.albProvider, m.trackingProvider, stack, m.logger),
		NewListenerApplier(m.albProvider, stack, m.logger),
		NewListenerRuleApplier(m.albProvider, stack, m.logger),
	}

	for _, applier := range appliers {
		if err := applier.Apply(ctx); err != nil {
			return err
		}
	}

	for i := len(appliers) - 1; i >= 0; i-- {
		if err := appliers[i].PostApply(ctx); err != nil {
			return err
		}
	}

	return nil
}
