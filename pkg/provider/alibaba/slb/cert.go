package slb

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
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

func (p SLBProvider) DescribeDomainExtensions(ctx context.Context, lbId string, port int) ([]model.DomainExtension, error) {
	req := slb.CreateDescribeDomainExtensionsRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	resp, err := p.auth.SLB.DescribeDomainExtensions(req)
	if err != nil {
		return nil, util.SDKError("DescribeDomainExtensions", err)
	}
	var ret []model.DomainExtension
	for _, r := range resp.DomainExtensions.DomainExtension {
		ret = append(ret, model.DomainExtension{
			DomainExtensionId:   r.DomainExtensionId,
			Domain:              r.Domain,
			ServerCertificateId: r.ServerCertificateId,
		})
	}
	klog.V(5).Infof("RequestId: %s, API: %s", resp.RequestId, "DescribeDomainExtensions")
	return ret, nil
}

func (p SLBProvider) CreateDomainExtension(ctx context.Context, lbId string, port int, domain string, certId string) error {
	req := slb.CreateCreateDomainExtensionRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	req.Domain = domain
	req.ServerCertificateId = certId
	resp, err := p.auth.SLB.CreateDomainExtension(req)
	if err != nil {
		return util.SDKError("CreateDomainExtension", err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s", resp.RequestId, "CreateDomainExtension")
	return nil
}

func (p SLBProvider) DeleteDomainExtension(ctx context.Context, id string) error {
	req := slb.CreateDeleteDomainExtensionRequest()
	req.DomainExtensionId = id
	resp, err := p.auth.SLB.DeleteDomainExtension(req)
	if err != nil {
		return util.SDKError("DeleteDomainExtension", err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s", resp.RequestId, "DeleteDomainExtension")
	return nil
}

func (p SLBProvider) SetDomainExtensionAttribute(ctx context.Context, id string, certId string) error {
	req := slb.CreateSetDomainExtensionAttributeRequest()
	req.DomainExtensionId = id
	req.ServerCertificateId = certId
	resp, err := p.auth.SLB.SetDomainExtensionAttribute(req)
	if err != nil {
		return util.SDKError("SetDomainExtensionAttribute", err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s", resp.RequestId, "SetDomainExtensionAttribute")
	return nil
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
