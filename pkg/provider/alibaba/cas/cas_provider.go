package cas

import (
	"context"

	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"

	cassdk "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
)

func NewCASProvider(
	auth *base.ClientMgr,
) *CASProvider {
	return &CASProvider{auth: auth}
}

var _ prvd.ICAS = &CASProvider{}

type CASProvider struct {
	auth *base.ClientMgr
}

func (c CASProvider) DescribeSSLCertificateList(ctx context.Context, request *cassdk.DescribeSSLCertificateListRequest) (*cassdk.DescribeSSLCertificateListResponse, error) {
	return c.auth.CAS.DescribeSSLCertificateList(request)
}

func (c CASProvider) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, request *cassdk.DescribeSSLCertificatePublicKeyDetailRequest) (*cassdk.DescribeSSLCertificatePublicKeyDetailResponse, error) {
	return c.auth.CAS.DescribeSSLCertificatePublicKeyDetail(request)
}
