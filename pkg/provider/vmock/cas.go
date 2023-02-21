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

func (c MockCAS) DeleteSSLCertificate(ctx context.Context, certId string) error {
	return nil
}
func (c MockCAS) CreateSSLCertificateWithName(ctx context.Context, certName, certificate, privateKey string) (string, error) {
	return "", nil
}
func (c MockCAS) DescribeSSLCertificateList(ctx context.Context) ([]model.CertificateInfo, error) {
	return nil, nil
}
