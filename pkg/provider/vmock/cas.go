package vmock

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	casprvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/cas"

	cassdk "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
)

func NewMockCAS(
	auth *base.ClientMgr,
) *MockCAS {
	return &MockCAS{auth: auth}
}

type MockCAS struct {
	auth *base.ClientMgr
	cas  *casprvd.CASProvider
}

func (c MockCAS) DescribeSSLCertificateList(ctx context.Context, request *cassdk.DescribeSSLCertificateListRequest) (*cassdk.DescribeSSLCertificateListResponse, error) {
	return nil, nil
}

func (c MockCAS) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, request *cassdk.DescribeSSLCertificatePublicKeyDetailRequest) (*cassdk.DescribeSSLCertificatePublicKeyDetailResponse, error) {
	return nil, nil
}
