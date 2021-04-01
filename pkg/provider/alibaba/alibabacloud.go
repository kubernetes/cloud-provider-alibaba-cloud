package alibaba

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
)

func NewAlibabaCloud() prvd.Provider {
	auth, err := metadata.NewClientAuth()
	if err != nil {
		log.Warnf("initialize alibaba cloud client auth: %s", err.Error())
	}
	if auth == nil {
		panic("auth should not be nil")
	}
	err = auth.Start(metadata.RefreshToken)
	if err != nil {
		log.Warnf("refresh token: %s", err.Error())
	}
	return &AlibabaCloud{
		Auth:          auth,
		EcsProvider:   NewEcsProvider(auth),
		ProviderSLB:   NewLBProvider(auth),
		PVTZProvider:  NewPVTZProvider(auth),
		RouteProvider: NewRouteProvider(auth),
	}
}

var _ prvd.Provider = AlibabaCloud{}

type AlibabaCloud struct {
	*EcsProvider
	*PVTZProvider
	*RouteProvider
	*ProviderSLB
	Auth *metadata.ClientAuth
}
