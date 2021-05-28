package alibaba

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

func NewAlibabaCloud() prvd.Provider {
	auth, err := NewClientAuth()
	if err != nil {
		log.Warnf("initialize alibaba cloud client auth: %s", err.Error())
	}
	if auth == nil {
		panic("auth should not be nil")
	}
	err = auth.Start(RefreshToken)
	if err != nil {
		log.Warnf("refresh token: %s", err.Error())
	}
	return &AlibabaCloud{
		IMetaData:    auth.Meta,
		EcsProvider:  NewEcsProvider(auth),
		ProviderSLB:  NewLBProvider(auth),
		PVTZProvider: NewPVTZProvider(auth),
		VPCProvider:  NewVPCProvider(auth),
	}
}

var _ prvd.Provider = AlibabaCloud{}

type AlibabaCloud struct {
	*EcsProvider
	*PVTZProvider
	*VPCProvider
	*ProviderSLB
	prvd.IMetaData
}
