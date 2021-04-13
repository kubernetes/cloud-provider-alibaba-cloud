package vmock

import "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"

func NewRouteProvider(
	auth *metadata.ClientAuth,
) *RouteProvider {
	return &RouteProvider{auth: auth}
}

type RouteProvider struct {
	auth *metadata.ClientAuth
}

func (*RouteProvider) CreateRoute() {
	panic("implement me")
}

func (*RouteProvider) DeleteRoute() {
	panic("implement me")
}

func (*RouteProvider) ListRoute() {
	panic("implement me")
}
