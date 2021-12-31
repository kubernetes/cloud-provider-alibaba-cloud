package albconfigmanager

import (
	"context"
	"fmt"

	v1 "k8s.io/cloud-provider-alibaba-cloud/pkg/apis/alibabacloud/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

func (t *defaultModelBuildTask) buildListener(ctx context.Context, lbID core.StringToken, lsSpec *v1.ListenerSpec) (*alb.Listener, error) {
	lsSpecDst, err := t.buildListenerSpec(ctx, lbID, lsSpec)
	if err != nil {
		return nil, err
	}
	lsResID := fmt.Sprintf("%v", lsSpecDst.ListenerPort)
	ls := alb.NewListener(t.stack, lsResID, lsSpecDst)
	return ls, nil
}

const (
	ListenerDescriptionPrefix = "ls"
)

func (t *defaultModelBuildTask) buildListenerSpec(ctx context.Context, lbID core.StringToken, apiLs *v1.ListenerSpec) (alb.ListenerSpec, error) {
	defaultAction, err := t.buildLsDefaultAction(ctx, apiLs.Port.IntValue())
	if err != nil {
		return alb.ListenerSpec{}, err
	}
	modelLs := alb.ListenerSpec{
		LoadBalancerID: lbID,
	}
	modelLs.DefaultActions = []alb.Action{defaultAction}
	modelLs.ListenerPort = apiLs.Port.IntValue()
	modelLs.ListenerDescription = apiLs.Description
	modelLs.ListenerProtocol = apiLs.Protocol
	modelLs.IdleTimeout = apiLs.IdleTimeout
	modelLs.RequestTimeout = apiLs.RequestTimeout
	modelLs.SecurityPolicyId = apiLs.SecurityPolicyId
	modelLs.LogConfig = alb.LogConfig{
		AccessLogRecordCustomizedHeadersEnabled: apiLs.LogConfig.AccessLogRecordCustomizedHeadersEnabled,
		AccessLogTracingConfig: alb.AccessLogTracingConfig{
			TracingSample:  apiLs.LogConfig.AccessLogTracingConfig.TracingSample,
			TracingType:    apiLs.LogConfig.AccessLogTracingConfig.TracingType,
			TracingEnabled: apiLs.LogConfig.AccessLogTracingConfig.TracingEnabled,
		},
	}
	modelLs.QuicConfig = alb.QuicConfig{
		QuicUpgradeEnabled: apiLs.QuicConfig.QuicUpgradeEnabled,
		QuicListenerId:     apiLs.QuicConfig.QuicListenerId,
	}
	modelLs.XForwardedForConfig = alb.XForwardedForConfig{
		XForwardedForClientCertSubjectDNAlias:      apiLs.XForwardedForConfig.XForwardedForClientCertSubjectDNAlias,
		XForwardedForClientCertSubjectDNEnabled:    apiLs.XForwardedForConfig.XForwardedForClientCertSubjectDNEnabled,
		XForwardedForProtoEnabled:                  apiLs.XForwardedForConfig.XForwardedForProtoEnabled,
		XForwardedForClientCertIssuerDNEnabled:     apiLs.XForwardedForConfig.XForwardedForClientCertIssuerDNEnabled,
		XForwardedForSLBIdEnabled:                  apiLs.XForwardedForConfig.XForwardedForSLBIdEnabled,
		XForwardedForClientSrcPortEnabled:          apiLs.XForwardedForConfig.XForwardedForClientSrcPortEnabled,
		XForwardedForClientCertFingerprintEnabled:  apiLs.XForwardedForConfig.XForwardedForClientCertFingerprintEnabled,
		XForwardedForEnabled:                       apiLs.XForwardedForConfig.XForwardedForEnabled,
		XForwardedForSLBPortEnabled:                apiLs.XForwardedForConfig.XForwardedForSLBPortEnabled,
		XForwardedForClientCertClientVerifyAlias:   apiLs.XForwardedForConfig.XForwardedForClientCertClientVerifyAlias,
		XForwardedForClientCertIssuerDNAlias:       apiLs.XForwardedForConfig.XForwardedForClientCertIssuerDNAlias,
		XForwardedForClientCertFingerprintAlias:    apiLs.XForwardedForConfig.XForwardedForClientCertFingerprintAlias,
		XForwardedForClientCertClientVerifyEnabled: apiLs.XForwardedForConfig.XForwardedForClientCertClientVerifyEnabled,
	}

	if modelLs.ListenerPort == 0 {
		modelLs.ListenerPort = t.defaultListenerPort
	}
	if len(modelLs.ListenerProtocol) == 0 {
		modelLs.ListenerProtocol = t.defaultListenerProtocol
	}
	if modelLs.IdleTimeout == 0 {
		modelLs.IdleTimeout = t.defaultListenerIdleTimeout
	}
	if modelLs.RequestTimeout == 0 {
		modelLs.RequestTimeout = t.defaultListenerRequestTimeout
	}
	if len(modelLs.ListenerDescription) == 0 {
		modelLs.ListenerDescription = fmt.Sprintf("%v-%v", ListenerDescriptionPrefix, modelLs.ListenerPort)
	}
	if apiLs.GzipEnabled == nil {
		modelLs.GzipEnabled = t.defaultListenerGzipEnabled
	}

	if apiLs.Protocol == string(ProtocolHTTPS) {
		modelLs.Certificates = transCertificatesFromAPIToSDK(apiLs.Certificates)
		if len(apiLs.CaCertificates) != 0 {
			modelLs.CaCertificates = transCertificatesFromAPIToSDK(apiLs.CaCertificates)
		}
		if len(modelLs.SecurityPolicyId) == 0 {
			modelLs.SecurityPolicyId = t.defaultListenerSecurityPolicyId
		}
		if apiLs.Http2Enabled == nil {
			modelLs.Http2Enabled = t.defaultListenerHttp2Enabled
		}
	}

	return modelLs, nil
}

func transCertificatesFromAPIToSDK(apiCerts []v1.Certificate) []alb.Certificate {
	sdkCerts := make([]alb.Certificate, 0)

	for _, apiCert := range apiCerts {
		sdkCerts = append(sdkCerts, alb.Certificate{
			IsDefault:     apiCert.IsDefault,
			CertificateId: apiCert.CertificateId,
		})
	}

	return sdkCerts
}
