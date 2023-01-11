package vmock

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"

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

func (c MockCAS) DescribeSSLCertificatePublicKeyDetail(ctx context.Context, certId string) (*model.CertificateInfo, error) {
	return nil, nil
}

func (c MockCAS) DescribeSSLCertificateList(ctx context.Context) ([]model.CertificateInfo, error) {
	return nil, nil
}
