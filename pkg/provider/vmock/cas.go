package vmock

import (
	"context"

	cassdk "github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockCAS(
	auth *base.ClientMgr,
) *MockCAS {
	return &MockCAS{auth: auth}
}

type MockCAS struct {
	auth *base.ClientMgr
}

func (c MockCAS) DescribeSSLCertificateList(ctx context.Context, request *cassdk.DescribeSSLCertificateListRequest) (*cassdk.DescribeSSLCertificateListResponse, error) {
	return nil, nil
}

func (c MockCAS) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, request *cassdk.DescribeSSLCertificatePublicKeyDetailRequest) (*cassdk.DescribeSSLCertificatePublicKeyDetailResponse, error) {
	return nil, nil
}
