package mock

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
)

func NewMockCloud(auth *metadata.ClientAuth) prvd.Provider {

	return &MockCloud{
		Auth:          auth,
		EcsProvider:   NewEcsProvider(auth),
		LBProvider:    NewLBProvider(auth),
		PVTZProvider:  NewPVTZProvider(auth),
		RouteProvider: NewRouteProvider(auth),
	}
}

var _ prvd.Provider = alibaba.AlibabaCloud{}

// MockCloud for unit test
type MockCloud struct {
	*EcsProvider
	*PVTZProvider
	*RouteProvider
	*LBProvider
	Auth *metadata.ClientAuth
}
