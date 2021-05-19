package vmock

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/auth"
)

func NewMockCloud(auth *auth.ClientAuth) prvd.Provider {

	return &MockCloud{
		Auth:          auth,
		MockECS:       NewMockECS(auth),
		MockCLB:       NewMockCLB(auth),
		MockPVTZ:      NewPVTZProvider(auth),
		RouteProvider: NewRouteProvider(auth),
	}
}

var _ prvd.Provider = alibaba.AlibabaCloud{}

// MockCloud for unit test
type MockCloud struct {
	*MockECS
	*MockPVTZ
	*RouteProvider
	*MockCLB
	Auth *auth.ClientAuth
}
