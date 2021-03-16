package alibaba

import "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"

func NewLBProvider(
	auth *metadata.ClientAuth,
) *LBProvider {
	return &LBProvider{auth: auth}
}

type LBProvider struct {
	auth *metadata.ClientAuth
}

func (*LBProvider) FindLoadBalancer() {
	panic("implement me")
}

func (*LBProvider) ListLoadBalancer() {
	panic("implement me")
}

func (*LBProvider) CreateLoadBalancer() {
	panic("implement me")
}
