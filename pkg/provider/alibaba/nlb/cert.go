package nlb

import (
	"context"

	nlb "github.com/alibabacloud-go/nlb-20220430/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
)

func (p *NLBProvider) ListNLBListenerCertificates(ctx context.Context, listenerId string) ([]nlbmodel.ListenerCertificate, error) {
	var ret []nlbmodel.ListenerCertificate
	nextToken := ""

	for {
		req := &nlb.ListListenerCertificatesRequest{}
		req.ListenerId = tea.String(listenerId)
		req.MaxResults = tea.Int32(50)
		req.NextToken = tea.String(nextToken)

		resp, err := retryOnThrottlingT("ListListenerCertificates",
			p.auth.NLB.ListListenerCertificates, req)

		if err != nil {
			return nil, util.SDKError("ListListenerCertificates", err)
		}

		for _, cert := range resp.Body.Certificates {
			ret = append(ret, nlbmodel.ListenerCertificate{
				Id:        tea.StringValue(cert.CertificateId),
				IsDefault: tea.BoolValue(cert.IsDefault),
				Status:    tea.StringValue(cert.Status),
				Type:      tea.StringValue(cert.CertificateType),
			})
		}

		klog.V(5).Infof("RequestId: %s, API: %s, NextToken: %s", tea.StringValue(resp.Body.RequestId), "ListListenerCertificates", nextToken)

		nextToken = tea.StringValue(resp.Body.NextToken)
		if nextToken == "" {
			break
		}
	}

	return ret, nil
}

func (p *NLBProvider) AssociateAdditionalCertificatesWithNLBListener(ctx context.Context, listenerId string, certIds []string) error {
	req := &nlb.AssociateAdditionalCertificatesWithListenerRequest{}
	req.ListenerId = tea.String(listenerId)
	req.AdditionalCertificateIds = tea.StringSlice(certIds)

	resp, err := retryOnThrottlingT("AssociateAdditionalCertificatesWithListener",
		p.auth.NLB.AssociateAdditionalCertificatesWithListener, req)
	if err != nil {
		return util.SDKError("AssociateAdditionalCertificatesWithListener", err)
	}

	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "AssociateAdditionalCertificatesWithListener")
	jobId := tea.StringValue(resp.Body.JobId)
	if jobId != "" {
		return p.WaitJobFinish("AssociateAdditionalCertificatesWithListener", jobId)
	}

	return nil
}

func (p *NLBProvider) DisassociateAdditionalCertificatesWithNLBListener(ctx context.Context, listenerId string, certIds []string) error {
	req := &nlb.DisassociateAdditionalCertificatesWithListenerRequest{}
	req.ListenerId = tea.String(listenerId)
	req.AdditionalCertificateIds = tea.StringSlice(certIds)

	resp, err := retryOnThrottlingT("DisassociateAdditionalCertificatesWithListener",
		p.auth.NLB.DisassociateAdditionalCertificatesWithListener, req)
	if err != nil {
		return util.SDKError("DisassociateAdditionalCertificatesWithListener", err)
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DisassociateAdditionalCertificatesWithListener")

	jobId := tea.StringValue(resp.Body.JobId)
	if jobId != "" {
		return p.WaitJobFinish("DisassociateAdditionalCertificatesWithListener", jobId)
	}
	return nil
}
