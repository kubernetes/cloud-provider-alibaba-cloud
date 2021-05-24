package vmock

import "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"

func NewRouteProvider(
	auth *alibaba.ClientAuth,
) *RouteProvider {
	return &RouteProvider{auth: auth}
}

type RouteProvider struct {
	auth *alibaba.ClientAuth
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
