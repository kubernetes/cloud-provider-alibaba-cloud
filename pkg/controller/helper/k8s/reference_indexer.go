package k8s

import (
	"context"

	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ReferenceIndexer interface {
	BuildServiceRefIndexes(ctx context.Context, ing *networking.Ingress) []string
}

func NewDefaultReferenceIndexer() *defaultReferenceIndexer {
	return &defaultReferenceIndexer{}
}

var _ ReferenceIndexer = &defaultReferenceIndexer{}

type defaultReferenceIndexer struct {
}

func (i *defaultReferenceIndexer) BuildServiceRefIndexes(ctx context.Context, ing *networking.Ingress) []string {
	var backends []networking.IngressBackend
	if ing.Spec.DefaultBackend != nil {
		backends = append(backends, *ing.Spec.DefaultBackend)
	}
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			backends = append(backends, path.Backend)
		}
	}

	serviceNames := sets.NewString()
	for _, backend := range backends {
		serviceNames.Insert(backend.Service.Name)
	}
	return serviceNames.List()
}
