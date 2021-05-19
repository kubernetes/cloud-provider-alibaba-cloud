package vmock

import "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/auth"

func NewRouteProvider(
	auth *auth.ClientAuth,
) *RouteProvider {
	return &RouteProvider{auth: auth}
}

type RouteProvider struct {
	auth *auth.ClientAuth
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
