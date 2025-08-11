package nlb

import (
	"context"
	"fmt"
	nlb "github.com/alibabacloud-go/nlb-20220430/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"strconv"
)

func (p *NLBProvider) ListNLBListeners(ctx context.Context, lbId string) ([]*nlbmodel.ListenerAttribute, error) {
	var respListeners []*nlb.ListListenersResponseBodyListeners
	nextToken := ""
	for {
		req := &nlb.ListListenersRequest{}
		req.LoadBalancerIds = []*string{tea.String(lbId)}
		req.MaxResults = tea.Int32(100)
		req.NextToken = tea.String(nextToken)

		var resp *nlb.ListListenersResponse
		err := retryOnThrottling("ListListeners", func() error {
			var err error
			resp, err = p.auth.NLB.ListListeners(req)
			return err
		})
		if err != nil {
			return nil, util.SDKError("ListListeners", err)
		}
		if resp == nil || resp.Body == nil {
			return nil, fmt.Errorf("OpenAPI ListNLBListeners resp is nil")
		}
		klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "ListNLBListeners")

		respListeners = append(respListeners, resp.Body.Listeners...)

		nextToken = tea.StringValue(resp.Body.NextToken)
		if nextToken == "" {
			break
		}
	}

	var listeners []*nlbmodel.ListenerAttribute
	for _, lis := range respListeners {
		n := &nlbmodel.ListenerAttribute{
			ListenerId:          tea.StringValue(lis.ListenerId),
			ListenerDescription: tea.StringValue(lis.ListenerDescription),
			ListenerProtocol:    tea.StringValue(lis.ListenerProtocol),
			ListenerPort:        tea.Int32Value(lis.ListenerPort),
			ServerGroupId:       tea.StringValue(lis.ServerGroupId),
			ListenerStatus:      nlbmodel.ListenerStatus(tea.StringValue(lis.ListenerStatus)),
		}
		if lis.IdleTimeout != nil {
			n.IdleTimeout = tea.Int32Value(lis.IdleTimeout)
		}
		if lis.SecurityPolicyId != nil {
			n.SecurityPolicyId = tea.StringValue(lis.SecurityPolicyId)
		}
		for _, c := range lis.CertificateIds {
			if c != nil {
				n.CertificateIds = append(n.CertificateIds, tea.StringValue(c))
			}
		}
		for _, c := range lis.CaCertificateIds {
			if c != nil {
				n.CaCertificateIds = append(n.CaCertificateIds, tea.StringValue(c))
			}
		}
		n.CaEnabled = lis.CaEnabled
		n.Cps = lis.Cps
		n.ProxyProtocolEnabled = lis.ProxyProtocolEnabled
		if lis.ProxyProtocolV2Config != nil {
			n.ProxyProtocolV2Config = nlbmodel.ProxyProtocolV2Config{
				PrivateLinkEpIdEnabled:  lis.ProxyProtocolV2Config.Ppv2PrivateLinkEpIdEnabled,
				PrivateLinkEpsIdEnabled: lis.ProxyProtocolV2Config.Ppv2PrivateLinkEpsIdEnabled,
				VpcIdEnabled:            lis.ProxyProtocolV2Config.Ppv2VpcIdEnabled,
			}
		}
		n.AlpnEnabled = lis.AlpnEnabled
		n.AlpnPolicy = tea.StringValue(lis.AlpnPolicy)

		nameKey, err := nlbmodel.LoadNLBListenerNamedKey(n.ListenerDescription)
		if err != nil {
			n.IsUserManaged = true
			klog.Warningf("listener description [%s], not expected format. skip user managed port",
				tea.StringValue(lis.ListenerDescription))
		}
		n.NamedKey = nameKey

		if lis.StartPort != nil {
			s, err := strconv.Atoi(tea.StringValue(lis.StartPort))
			if err != nil {
				return nil, fmt.Errorf("error converting listener start port to int: %w", err)
			}
			n.StartPort = int32(s)
		}

		if lis.EndPort != nil {
			e, err := strconv.Atoi(tea.StringValue(lis.EndPort))
			if err != nil {
				return nil, fmt.Errorf("error converting listener end port to int: %w", err)
			}
			n.EndPort = int32(e)
		}

		listeners = append(listeners, n)
	}
	return listeners, nil
}

func (p *NLBProvider) CreateNLBListener(ctx context.Context, lbId string, lis *nlbmodel.ListenerAttribute) error {
	_, err := p.CreateNLBListenerAsync(ctx, lbId, lis)
	return err
}

func (p *NLBProvider) UpdateNLBListener(ctx context.Context, lis *nlbmodel.ListenerAttribute) error {
	_, err := p.UpdateNLBListenerAsync(ctx, lis)
	return err
}

func (p *NLBProvider) DeleteNLBListener(ctx context.Context, listenerId string) error {
	jobId, err := p.DeleteNLBListenerAsync(ctx, listenerId)
	if err != nil {
		return err
	}
	return p.waitJobFinish("DeleteListener", jobId)
}

func (p *NLBProvider) StartNLBListener(ctx context.Context, listenerId string) error {
	req := &nlb.StartListenerRequest{}
	req.ListenerId = tea.String(listenerId)

	var resp *nlb.StartListenerResponse
	err := retryOnThrottling("StartListener", func() error {
		var err error
		resp, err = p.auth.NLB.StartListener(req)
		return err
	})
	if err != nil {
		return util.SDKError("StartListener", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI StartListener resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "StartListener")
	return nil
}

func (p *NLBProvider) CreateNLBListenerAsync(ctx context.Context, lbId string, lis *nlbmodel.ListenerAttribute) (string, error) {
	req := &nlb.CreateListenerRequest{}
	req.LoadBalancerId = tea.String(lbId)
	req.ListenerProtocol = tea.String(lis.ListenerProtocol)
	req.ListenerPort = tea.Int32(lis.ListenerPort)
	req.ListenerDescription = tea.String(lis.ListenerDescription)
	req.ServerGroupId = tea.String(lis.ServerGroupId)
	req.Cps = lis.Cps
	req.ProxyProtocolEnabled = lis.ProxyProtocolEnabled
	if lis.StartPort != 0 {
		req.StartPort = tea.Int32(lis.StartPort)
	}
	if lis.EndPort != 0 {
		req.EndPort = tea.Int32(lis.EndPort)
	}
	if lis.ProxyProtocolV2Config.PrivateLinkEpIdEnabled != nil ||
		lis.ProxyProtocolV2Config.PrivateLinkEpsIdEnabled != nil ||
		lis.ProxyProtocolV2Config.VpcIdEnabled != nil {
		req.ProxyProtocolV2Config = &nlb.CreateListenerRequestProxyProtocolV2Config{
			Ppv2PrivateLinkEpIdEnabled:  lis.ProxyProtocolV2Config.PrivateLinkEpIdEnabled,
			Ppv2PrivateLinkEpsIdEnabled: lis.ProxyProtocolV2Config.PrivateLinkEpsIdEnabled,
			Ppv2VpcIdEnabled:            lis.ProxyProtocolV2Config.VpcIdEnabled,
		}
	}

	req.AlpnEnabled = lis.AlpnEnabled
	if tea.BoolValue(lis.AlpnEnabled) {
		req.AlpnPolicy = tea.String(lis.AlpnPolicy)
	}

	if lis.IdleTimeout != 0 {
		req.IdleTimeout = tea.Int32(lis.IdleTimeout)
	}
	if lis.SecurityPolicyId != "" {
		req.SecurityPolicyId = tea.String(lis.SecurityPolicyId)
	}
	for _, cert := range lis.CertificateIds {
		req.CertificateIds = append(req.CertificateIds, tea.String(cert))
	}
	for _, cert := range lis.CaCertificateIds {
		req.CaCertificateIds = append(req.CaCertificateIds, tea.String(cert))
	}
	req.CaEnabled = lis.CaEnabled

	var resp *nlb.CreateListenerResponse
	err := retryOnThrottling("CreateListener", func() error {
		var err error
		resp, err = p.auth.NLB.CreateListener(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("CreateListener", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI CreateListener resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "CreateListener")
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) UpdateNLBListenerAsync(ctx context.Context, lis *nlbmodel.ListenerAttribute) (string, error) {
	req := &nlb.UpdateListenerAttributeRequest{}
	req.ListenerId = tea.String(lis.ListenerId)
	req.ListenerDescription = tea.String(lis.ListenerDescription)
	req.ServerGroupId = tea.String(lis.ServerGroupId)
	req.Cps = lis.Cps
	req.ProxyProtocolEnabled = lis.ProxyProtocolEnabled
	if lis.ProxyProtocolV2Config.PrivateLinkEpIdEnabled != nil ||
		lis.ProxyProtocolV2Config.PrivateLinkEpsIdEnabled != nil ||
		lis.ProxyProtocolV2Config.VpcIdEnabled != nil {
		req.ProxyProtocolV2Config = &nlb.UpdateListenerAttributeRequestProxyProtocolV2Config{
			Ppv2PrivateLinkEpIdEnabled:  lis.ProxyProtocolV2Config.PrivateLinkEpIdEnabled,
			Ppv2PrivateLinkEpsIdEnabled: lis.ProxyProtocolV2Config.PrivateLinkEpsIdEnabled,
			Ppv2VpcIdEnabled:            lis.ProxyProtocolV2Config.VpcIdEnabled,
		}
	}
	req.AlpnEnabled = lis.AlpnEnabled
	if tea.BoolValue(lis.AlpnEnabled) {
		req.AlpnPolicy = tea.String(lis.AlpnPolicy)
	}
	if lis.IdleTimeout != 0 {
		req.IdleTimeout = tea.Int32(lis.IdleTimeout)
	}
	if lis.SecurityPolicyId != "" {
		req.SecurityPolicyId = tea.String(lis.SecurityPolicyId)
	}
	for _, cert := range lis.CertificateIds {
		req.CertificateIds = append(req.CertificateIds, tea.String(cert))
	}
	for _, cert := range lis.CaCertificateIds {
		req.CaCertificateIds = append(req.CaCertificateIds, tea.String(cert))
	}
	req.CaEnabled = lis.CaEnabled

	var resp *nlb.UpdateListenerAttributeResponse
	err := retryOnThrottling("UpdateListenerAttribute", func() error {
		var err error
		resp, err = p.auth.NLB.UpdateListenerAttribute(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("UpdateListenerAttribute", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI UpdateListenerAttribute resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateListenerAttribute")
	return tea.StringValue(resp.Body.JobId), nil
}

func (p *NLBProvider) DeleteNLBListenerAsync(ctx context.Context, listenerId string) (string, error) {
	req := &nlb.DeleteListenerRequest{}
	req.ListenerId = tea.String(listenerId)

	var resp *nlb.DeleteListenerResponse
	err := retryOnThrottling("DeleteListener", func() error {
		var err error
		resp, err = p.auth.NLB.DeleteListener(req)
		return err
	})
	if err != nil {
		return "", util.SDKError("DeleteNLBListener", err)
	}
	if resp == nil || resp.Body == nil {
		return "", fmt.Errorf("OpenAPI DeleteNLBListener resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DeleteNLBListener")
	return tea.StringValue(resp.Body.JobId), nil
}
