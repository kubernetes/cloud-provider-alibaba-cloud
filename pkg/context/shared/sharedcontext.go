package shared

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

func NewSharedContext(
	prvd prvd.Provider,
) *SharedContext {
	ctxs := SharedContext{}
	ctxs.SetKV(Provider, prvd)
	return &ctxs
}

const (
	Provider = "Provider"
)

type SharedContext struct{ base.Context }

// Provider
func (c *SharedContext) Provider() prvd.Provider {
	provider, ok := c.Value(Provider)
	if !ok {
		return nil
	}
	return provider.(prvd.Provider)
}
