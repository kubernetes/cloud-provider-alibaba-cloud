package alibaba

func NewRouteProvider(
	auth *ClientAuth,
) *RouteProvider {
	return &RouteProvider{auth: auth}
}

type RouteProvider struct {
	auth *ClientAuth
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
