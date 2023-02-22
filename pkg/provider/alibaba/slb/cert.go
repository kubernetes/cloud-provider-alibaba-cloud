package slb

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
)

func (p SLBProvider) DescribeServerCertificateById(ctx context.Context, serverCertificateId string) (*model.CertAttribute, error) {
	req := slb.CreateDescribeServerCertificatesRequest()
	req.ServerCertificateId = serverCertificateId

	resp, err := p.auth.SLB.DescribeServerCertificates(req)
	if err != nil {
		return nil, util.SDKError("DescribeServerCertificates", err)
	}

	if len(resp.ServerCertificates.ServerCertificate) == 0 {
		return nil, nil
	}

	if len(resp.ServerCertificates.ServerCertificate) > 1 {
		return nil, fmt.Errorf("find more than 1 server cert by id %s", serverCertificateId)
	}

	return &model.CertAttribute{
		CreateTimeStamp:     resp.ServerCertificates.ServerCertificate[0].CreateTimeStamp,
		ExpireTimeStamp:     resp.ServerCertificates.ServerCertificate[0].ExpireTimeStamp,
		ServerCertificateId: resp.ServerCertificates.ServerCertificate[0].ServerCertificateId,
		CommonName:          resp.ServerCertificates.ServerCertificate[0].CertificateId,
	}, nil

}

// DescribeServerCertificates used for e2etest
func (p SLBProvider) DescribeServerCertificates(ctx context.Context) ([]string, error) {
	req := slb.CreateDescribeServerCertificatesRequest()
	resp, err := p.auth.SLB.DescribeServerCertificates(req)
	if err != nil {
		return nil, err
	}
	var certs []string
	for _, cert := range resp.ServerCertificates.ServerCertificate {
		certs = append(certs, cert.ServerCertificateId)
	}
	return certs, nil
}

// DescribeCACertificates used for e2etest
func (p SLBProvider) DescribeCACertificates(ctx context.Context) ([]string, error) {
	req := slb.CreateDescribeCACertificatesRequest()
	resp, err := p.auth.SLB.DescribeCACertificates(req)
	if err != nil {
		return nil, err
	}
	var certIds []string
	for _, cert := range resp.CACertificates.CACertificate {
		certIds = append(certIds, cert.CACertificateId)
	}
	return certIds, nil
}
