package dryrun

import (
	"context"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	casprvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/cas"

	cassdk "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
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

func (c DryRunCAS) DescribeSSLCertificateList(ctx context.Context, request *cassdk.DescribeSSLCertificateListRequest) (*cassdk.DescribeSSLCertificateListResponse, error) {
	return c.auth.CAS.DescribeSSLCertificateList(request)
}

func (c DryRunCAS) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, request *cassdk.DescribeSSLCertificatePublicKeyDetailRequest) (*cassdk.DescribeSSLCertificatePublicKeyDetailResponse, error) {
	return c.auth.CAS.DescribeSSLCertificatePublicKeyDetail(request)
}
