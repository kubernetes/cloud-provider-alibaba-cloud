package backend

import (
	"context"
	"fmt"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/ingress/reconcile/store"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewBackendManager(store store.Storer, kubeClient client.Client, cloud prvd.Provider, logger logr.Logger) *Manager {
	return &Manager{
		store:            store,
		k8sClient:        kubeClient,
		EndpointResolver: NewDefaultEndpointResolver(store, kubeClient, cloud, logger),
	}
}

type Manager struct {
	store     store.Storer
	k8sClient client.Client
	EndpointResolver
}

func (mgr *Manager) BuildServicePortSDKBackends(ctx context.Context, svcKey types.NamespacedName, port intstr.IntOrString) ([]alb.BackendItem, bool, error) {
	svc, err := helper.GetService(mgr.k8sClient, svcKey)
	if err != nil {
		return nil, false, err
	}

	var (
		modelBackends                   []alb.BackendItem
		endpoints                       []NodePortEndpoint
		containsPotentialReadyEndpoints bool
	)

	policy, err := helper.GetServiceTrafficPolicy(svc)
	if err != nil {
		return nil, false, err
	}
	switch policy {
	case helper.ENITrafficPolicy:
		endpoints, containsPotentialReadyEndpoints, err = mgr.ResolveENIEndpoints(ctx, util.NamespacedName(svc), port)
		if err != nil {
			return modelBackends, containsPotentialReadyEndpoints, err
		}
	case helper.LocalTrafficPolicy:
		endpoints, containsPotentialReadyEndpoints, err = mgr.ResolveLocalEndpoints(ctx, util.NamespacedName(svc), port)
		if err != nil {
			return modelBackends, containsPotentialReadyEndpoints, err
		}
	case helper.ClusterTrafficPolicy:
		endpoints, containsPotentialReadyEndpoints, err = mgr.ResolveClusterEndpoints(ctx, util.NamespacedName(svc), port)
		if err != nil {
			return modelBackends, containsPotentialReadyEndpoints, err
		}
	default:
		return modelBackends, containsPotentialReadyEndpoints, fmt.Errorf("not supported traffic policy [%s]", policy)
	}

	for _, endpoint := range endpoints {
		modelBackends = append(modelBackends, alb.BackendItem(endpoint))
	}

	return modelBackends, containsPotentialReadyEndpoints, nil
}

func (mgr *Manager) BuildServicePortsToSDKBackends(ctx context.Context, svc *v1.Service) (map[int32][]alb.BackendItem, bool, error) {
	svcPort2Backends := make(map[int32][]alb.BackendItem)
	containsPotentialReadyEndpoints := false
	for _, port := range svc.Spec.Ports {
		backends, _containsPotentialReadyEndpoints, err := mgr.BuildServicePortSDKBackends(ctx, util.NamespacedName(svc), intstr.FromInt(int(port.Port)))
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, _containsPotentialReadyEndpoints, err
		}
		containsPotentialReadyEndpoints = _containsPotentialReadyEndpoints
		svcPort2Backends[port.Port] = backends
	}

	return svcPort2Backends, containsPotentialReadyEndpoints, nil
}
