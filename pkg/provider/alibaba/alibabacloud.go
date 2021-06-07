package alibaba

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/ecs"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/pvtz"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"

	log "github.com/sirupsen/logrus"
)

func NewAlibabaCloud() prvd.Provider {
	auth, err := base.NewClientAuth()
	if err != nil {
		log.Warnf("initialize alibaba cloud client auth: %s", err.Error())
	}
	if auth == nil {
		panic("auth should not be nil")
	}
	err = auth.Start(base.RefreshToken)
	if err != nil {
		log.Warnf("refresh token: %s", err.Error())
	}
	return &AlibabaCloud{
		IMetaData:    auth.Meta,
		EcsProvider:  ecs.NewEcsProvider(auth),
		ProviderSLB:  slb.NewLBProvider(auth),
		PVTZProvider: pvtz.NewPVTZProvider(auth),
		VPCProvider:  vpc.NewVPCProvider(auth),
	}
}

var _ prvd.Provider = AlibabaCloud{}

type AlibabaCloud struct {
	*ecs.EcsProvider
	*pvtz.PVTZProvider
	*vpc.VPCProvider
	*slb.ProviderSLB
	prvd.IMetaData
}
