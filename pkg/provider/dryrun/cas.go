package dryrun

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	casprvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/cas"
)

func NewDryRunCAS(
	auth *base.ClientMgr, cas *casprvd.CASProvider,
) *DryRunCAS {
	return &DryRunCAS{auth: auth, cas: cas}
}

var _ prvd.ICAS = &DryRunCAS{}

type DryRunCAS struct {
	auth *base.ClientMgr
	cas  *casprvd.CASProvider
}

func (c DryRunCAS) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, certId string) (*model.CertificateInfo, error) {
	return c.cas.DescribeSSLCertificatePublicKeyDetail(ctx, certId)
}

func (c DryRunCAS) DescribeSSLCertificateList(ctx context.Context) ([]model.CertificateInfo, error) {
	return c.cas.DescribeSSLCertificateList(ctx)
}
