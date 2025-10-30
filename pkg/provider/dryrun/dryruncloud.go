package dryrun

import (
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/cas"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/ecs"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/eflo"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/pvtz"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/sls"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/vpc"
	"k8s.io/klog/v2"
)

func NewDryRunCloud() prvd.Provider {
	auth, err := base.NewClientMgr()
	if err != nil {
		klog.Warningf("initialize alibaba cloud client auth: %s", err.Error())
	}
	if auth == nil {
		panic("auth should not be nil")
	}
	err = auth.Start(base.RefreshToken)
	if err != nil {
		klog.Warningf("refresh token: %s", err.Error())
	}

	cloud := &alibaba.AlibabaCloud{
		IMetaData:    auth.Meta,
		ECSProvider:  ecs.NewECSProvider(auth),
		SLBProvider:  slb.NewLBProvider(auth),
		PVTZProvider: pvtz.NewPVTZProvider(auth),
		VPCProvider:  vpc.NewVPCProvider(auth),
		ALBProvider:  alb.NewALBProvider(auth),
		SLSProvider:  sls.NewSLSProvider(auth),
		CASProvider:  cas.NewCASProvider(auth),
		NLBProvider:  nlb.NewNLBProvider(auth),
		EFLOProvider: eflo.NewEFLOProvider(auth),
	}

	return &DryRunCloud{
		IMetaData:  auth.Meta,
		DryRunECS:  NewDryRunECS(auth, cloud.ECSProvider),
		DryRunPVTZ: NewDryRunPVTZ(auth, cloud.PVTZProvider),
		DryRunVPC:  NewDryRunVPC(auth, cloud.VPCProvider),
		DryRunSLB:  NewDryRunSLB(auth, cloud.SLBProvider),
		DryRunALB:  NewDryRunALB(auth, cloud.ALBProvider),
		DryRunSLS:  NewDryRunSLS(auth, cloud.SLSProvider),
		DryRunCAS:  NewDryRunCAS(auth, cloud.CASProvider),
		DryRunNLB:  NewDryRunNLB(auth, cloud.NLBProvider),
		DryRunEFLO: NewDryRunEFLO(auth, cloud.EFLOProvider),
	}
}

var _ prvd.Provider = &DryRunCloud{}

type DryRunCloud struct {
	*DryRunECS
	*DryRunPVTZ
	*DryRunVPC
	*DryRunSLB
	*DryRunALB
	*DryRunSLS
	*DryRunCAS
	*DryRunNLB
	*DryRunEFLO
	prvd.IMetaData
}
