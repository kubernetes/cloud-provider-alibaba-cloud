package slb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/alibabacloud-go/tea/tea"
	"k8s.io/klog/v2"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
)

func (p SLBProvider) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	req := slb.CreateDescribeLoadBalancerListenersRequest()
	req.LoadBalancerId = &[]string{lbId}
	req.MaxResults = requests.NewInteger(50)

	var respListeners []slb.ListenerInDescribeLoadBalancerListeners
	for {
		resp, err := p.auth.SLB.DescribeLoadBalancerListeners(req)
		if err != nil {
			return nil, util.SDKError("DescribeLoadBalancerListeners", err)
		}
		klog.V(5).Infof("RequestId: %s, API: %s, lbId: %s", resp.RequestId, "DescribeLoadBalancerListeners", lbId)
		respListeners = append(respListeners, resp.Listeners...)

		if resp.NextToken == "" {
			break
		}
		req.NextToken = resp.NextToken
	}

	var listeners []model.ListenerAttribute
	for _, lis := range respListeners {
		n := model.ListenerAttribute{
			Description:  lis.Description,
			ListenerPort: lis.ListenerPort,
			Protocol:     lis.ListenerProtocol,
			Status:       model.ListenerStatus(lis.Status),
			Bandwidth:    lis.Bandwidth,
			Scheduler:    lis.Scheduler,
			VGroupId:     lis.VServerGroupId,
			AclId:        lis.AclId,
			AclStatus:    model.FlagType(lis.AclStatus),
			AclType:      lis.AclType,
		}
		switch n.Protocol {
		case model.TCP:
			loadTCPListener(lis.TCPListenerConfig, &n)
		case model.UDP:
			loadUDPListener(lis.UDPListenerConfig, &n)
		case model.HTTP:
			loadHTTPListener(lis.HTTPListenerConfig, &n)
		case model.HTTPS:
			loadHTTPSListener(lis.HTTPSListenerConfig, &n)
		default:
			return listeners, fmt.Errorf("not support protocol %s", n.Protocol)
		}

		namedKey, err := model.LoadListenerNamedKey(lis.Description)
		if err != nil {
			n.IsUserManaged = true
			klog.Warningf("listener description [%s], not expected format. skip user managed port", lis.Description)
		}
		n.NamedKey = namedKey

		listeners = append(listeners, n)
	}
	return listeners, nil
}

func (p SLBProvider) StartLoadBalancerListener(ctx context.Context, lbId string, port int, proto string) error {
	req := slb.CreateStartLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	req.ListenerProtocol = proto
	_, err := p.auth.SLB.StartLoadBalancerListener(req)
	return util.SDKError("StartLoadBalancerListener", err)
}

func (p SLBProvider) StopLoadBalancerListener(ctx context.Context, lbId string, port int, proto string) error {
	req := slb.CreateStopLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	req.ListenerProtocol = proto
	_, err := p.auth.SLB.StopLoadBalancerListener(req)
	return util.SDKError("StopLoadBalancerListener", err)
}

func (p SLBProvider) DeleteLoadBalancerListener(ctx context.Context, lbId string, port int, proto string) error {
	req := slb.CreateDeleteLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	req.ListenerProtocol = proto

	_, err := p.auth.SLB.DeleteLoadBalancerListener(req)
	return util.SDKError("DeleteLoadBalancerListener", err)

}

func (p SLBProvider) CreateLoadBalancerTCPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerTCPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	setTCPListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerTCPListener(req)
	return util.SDKError("CreateLoadBalancerTCPListener", err)
}

func (p SLBProvider) SetLoadBalancerTCPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerTCPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	setTCPListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerTCPListenerAttribute(req)
	return util.SDKError("SetLoadBalancerTCPListenerAttribute", err)
}

func (p SLBProvider) CreateLoadBalancerUDPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerUDPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	setUDPListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerUDPListener(req)
	return util.SDKError("CreateLoadBalancerUDPListener", err)
}

func (p SLBProvider) SetLoadBalancerUDPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerUDPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	setUDPListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerUDPListenerAttribute(req)
	return util.SDKError("SetLoadBalancerUDPListenerAttribute", err)
}

func (p SLBProvider) CreateLoadBalancerHTTPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	setHTTPListenerValue(req, &listener)
	// set params only for CreateLoadBalancerHTTPListenerRequest
	if listener.ListenerForward != "" {
		req.ListenerForward = string(listener.ListenerForward)
	}
	if listener.ForwardPort != 0 {
		req.ForwardPort = requests.NewInteger(listener.ForwardPort)
	}
	_, err := p.auth.SLB.CreateLoadBalancerHTTPListener(req)
	return util.SDKError("CreateLoadBalancerHTTPListener", err)
}

func (p SLBProvider) SetLoadBalancerHTTPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	setHTTPListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerHTTPListenerAttribute(req)
	return util.SDKError("SetLoadBalancerHTTPListenerAttribute", err)
}

func (p SLBProvider) CreateLoadBalancerHTTPSListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPSListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	setHTTPSListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerHTTPSListener(req)
	return util.SDKError("CreateLoadBalancerHTTPSListener", err)
}

func (p SLBProvider) SetLoadBalancerHTTPSListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPSListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	setHTTPSListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerHTTPSListenerAttribute(req)
	return util.SDKError("SetLoadBalancerHTTPSListenerAttribute", err)
}

func setGenericListenerValue(req interface{}, listener *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	listenerPort := v.FieldByName("ListenerPort")
	listenerPort.SetString(strconv.Itoa(listener.ListenerPort))

	vGroupId := v.FieldByName("VServerGroupId")
	vGroupId.SetString(listener.VGroupId)

	description := v.FieldByName("Description")
	description.SetString(listener.Description)

	if listener.AclId != "" {
		aclId := v.FieldByName("AclId")
		aclId.SetString(listener.AclId)
	}
	if listener.AclType != "" {
		aclType := v.FieldByName("AclType")
		aclType.SetString(listener.AclType)
	}
	if listener.AclStatus != "" {
		aclStatus := v.FieldByName("AclStatus")
		aclStatus.SetString(string(listener.AclStatus))
	}

	if listener.Bandwidth != 0 {
		bandwidth := v.FieldByName("Bandwidth")
		bandwidth.SetString(strconv.Itoa(listener.Bandwidth))
	}

	if listener.Scheduler != "" {
		scheduler := v.FieldByName("Scheduler")
		scheduler.SetString(listener.Scheduler)
	}

	if listener.HealthCheckConnectPort != 0 {
		connectPort := v.FieldByName("HealthCheckConnectPort")
		connectPort.SetString(strconv.Itoa(listener.HealthCheckConnectPort))
	}

	if listener.HealthCheckInterval != 0 {
		healthCheckInterval := v.FieldByName("HealthCheckInterval")
		healthCheckInterval.SetString(strconv.Itoa(listener.HealthCheckInterval))
	}

	if listener.HealthyThreshold != 0 {
		healthyThreshold := v.FieldByName("HealthyThreshold")
		healthyThreshold.SetString(strconv.Itoa(listener.HealthyThreshold))
	}

	if listener.UnhealthyThreshold != 0 {
		unHealthyThreshold := v.FieldByName("UnhealthyThreshold")
		unHealthyThreshold.SetString(strconv.Itoa(listener.UnhealthyThreshold))
	}
}

func setTCPListenerValue(req interface{}, listener *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	if listener.PersistenceTimeout != nil {
		persistenceTimeout := v.FieldByName("PersistenceTimeout")
		persistenceTimeout.SetString(strconv.Itoa(*listener.PersistenceTimeout))
	}
	if listener.EstablishedTimeout != 0 {
		establishedTimeout := v.FieldByName("EstablishedTimeout")
		establishedTimeout.SetString(strconv.Itoa(listener.EstablishedTimeout))
	}
	if listener.HealthCheckConnectTimeout != 0 {
		healthCheckConnectTimeout := v.FieldByName("HealthCheckConnectTimeout")
		healthCheckConnectTimeout.SetString(strconv.Itoa(listener.HealthCheckConnectTimeout))
	}
	if listener.HealthCheckHttpCode != "" {
		healthCheckHttpCode := v.FieldByName("HealthCheckHttpCode")
		healthCheckHttpCode.SetString(listener.HealthCheckHttpCode)
	}
	if listener.HealthCheckURI != "" {
		healthCheckURI := v.FieldByName("HealthCheckURI")
		healthCheckURI.SetString(listener.HealthCheckURI)
	}
	if listener.HealthCheckType != "" {
		healthCheckType := v.FieldByName("HealthCheckType")
		healthCheckType.SetString(listener.HealthCheckType)
	}
	if listener.HealthCheckDomain != "" {
		healthCheckDomain := v.FieldByName("HealthCheckDomain")
		healthCheckDomain.SetString(listener.HealthCheckDomain)
	}
	if listener.HealthCheckSwitch != "" {
		healthCheckSwitch := v.FieldByName("HealthCheckSwitch")
		healthCheckSwitch.SetString(string(listener.HealthCheckSwitch))
	}

	if listener.ConnectionDrain != "" {
		connectionDrain := v.FieldByName("ConnectionDrain")
		connectionDrain.SetString(string(listener.ConnectionDrain))
	}
	if listener.ConnectionDrainTimeout != 0 {
		connectionDrainTimeout := v.FieldByName("ConnectionDrainTimeout")
		connectionDrainTimeout.SetString(strconv.Itoa(listener.ConnectionDrainTimeout))
	}
	if listener.EnableProxyProtocolV2 != nil {
		enabled := v.FieldByName("ProxyProtocolV2Enabled")
		enabled.SetString(string(requests.NewBoolean(*listener.EnableProxyProtocolV2)))
	}
}

func setUDPListenerValue(req interface{}, listener *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	if listener.HealthCheckConnectTimeout != 0 {
		healthCheckConnectTimeout := v.FieldByName("HealthCheckConnectTimeout")
		healthCheckConnectTimeout.SetString(strconv.Itoa(listener.HealthCheckConnectTimeout))
	}
	if listener.ConnectionDrain != "" {
		connectionDrain := v.FieldByName("ConnectionDrain")
		connectionDrain.SetString(string(listener.ConnectionDrain))
	}
	if listener.ConnectionDrainTimeout != 0 {
		connectionDrainTimeout := v.FieldByName("ConnectionDrainTimeout")
		connectionDrainTimeout.SetString(strconv.Itoa(listener.ConnectionDrainTimeout))
	}
	if listener.HealthCheckSwitch != "" {
		healthCheckSwitch := v.FieldByName("HealthCheckSwitch")
		healthCheckSwitch.SetString(string(listener.HealthCheckSwitch))
	}
	if listener.EnableProxyProtocolV2 != nil {
		enabled := v.FieldByName("ProxyProtocolV2Enabled")
		enabled.SetString(string(requests.NewBoolean(*listener.EnableProxyProtocolV2)))
	}

}

func setHTTPListenerValue(req interface{}, listener *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	if listener.HealthCheck != "" {
		healthCheck := v.FieldByName("HealthCheck")
		healthCheck.SetString(string(listener.HealthCheck))
	}
	if listener.HealthCheckHttpCode != "" {
		healthCheckHttpCode := v.FieldByName("HealthCheckHttpCode")
		healthCheckHttpCode.SetString(listener.HealthCheckHttpCode)
	}
	if listener.HealthCheckURI != "" {
		healthCheckURI := v.FieldByName("HealthCheckURI")
		healthCheckURI.SetString(listener.HealthCheckURI)
	}
	if listener.HealthCheckDomain != "" {
		healthCheckDomain := v.FieldByName("HealthCheckDomain")
		healthCheckDomain.SetString(listener.HealthCheckDomain)
	}
	if listener.HealthCheckTimeout != 0 {
		healthCheckTimeout := v.FieldByName("HealthCheckTimeout")
		healthCheckTimeout.SetString(strconv.Itoa(listener.HealthCheckTimeout))
	}
	if listener.HealthCheckMethod != "" {
		healthCheckMethod := v.FieldByName("HealthCheckMethod")
		healthCheckMethod.SetString(listener.HealthCheckMethod)
	}
	if listener.Cookie != "" {
		cookie := v.FieldByName("Cookie")
		cookie.SetString(listener.Cookie)
	}
	if listener.CookieTimeout != 0 {
		cookieTimeout := v.FieldByName("CookieTimeout")
		cookieTimeout.SetString(strconv.Itoa(listener.CookieTimeout))
	}
	if listener.StickySession != "" {
		stickySession := v.FieldByName("StickySession")
		stickySession.SetString(string(listener.StickySession))
	}
	if listener.StickySessionType != "" {
		stickySessionType := v.FieldByName("StickySessionType")
		stickySessionType.SetString(listener.StickySessionType)
	}
	if listener.XForwardedForProto != "" {
		xForwardedForProto := v.FieldByName("XForwardedForProto")
		xForwardedForProto.SetString(string(listener.XForwardedForProto))
	}
	if listener.XForwardedForSLBPort != "" {
		xForwardedForSLBPORT := v.FieldByName("XForwardedForSLBPORT")
		xForwardedForSLBPORT.SetString(string(listener.XForwardedForSLBPort))
	}
	if listener.XForwardedForClientSrcPort != "" {
		xForwardedForClientSrcPort := v.FieldByName("XForwardedForClientSrcPort")
		xForwardedForClientSrcPort.SetString(string(listener.XForwardedForClientSrcPort))
	}
	if listener.IdleTimeout != 0 {
		idleTimeout := v.FieldByName("IdleTimeout")
		idleTimeout.SetString(strconv.Itoa(listener.IdleTimeout))
	}
	if listener.RequestTimeout != 0 {
		requestTimeout := v.FieldByName("RequestTimeout")
		requestTimeout.SetString(strconv.Itoa(listener.RequestTimeout))
	}
}

func setHTTPSListenerValue(req interface{}, listener *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	if listener.HealthCheck != "" {
		healthCheck := v.FieldByName("HealthCheck")
		healthCheck.SetString(string(listener.HealthCheck))
	}
	if listener.HealthCheckHttpCode != "" {
		healthCheckHttpCode := v.FieldByName("HealthCheckHttpCode")
		healthCheckHttpCode.SetString(listener.HealthCheckHttpCode)
	}
	if listener.HealthCheckURI != "" {
		healthCheckURI := v.FieldByName("HealthCheckURI")
		healthCheckURI.SetString(listener.HealthCheckURI)
	}
	if listener.HealthCheckDomain != "" {
		healthCheckDomain := v.FieldByName("HealthCheckDomain")
		healthCheckDomain.SetString(listener.HealthCheckDomain)
	}
	if listener.HealthCheckTimeout != 0 {
		healthCheckTimeout := v.FieldByName("HealthCheckTimeout")
		healthCheckTimeout.SetString(strconv.Itoa(listener.HealthCheckTimeout))
	}
	if listener.HealthCheckMethod != "" {
		healthCheckMethod := v.FieldByName("HealthCheckMethod")
		healthCheckMethod.SetString(listener.HealthCheckMethod)
	}
	if listener.Cookie != "" {
		cookie := v.FieldByName("Cookie")
		cookie.SetString(listener.Cookie)
	}
	if listener.CookieTimeout != 0 {
		cookieTimeout := v.FieldByName("CookieTimeout")
		cookieTimeout.SetString(strconv.Itoa(listener.CookieTimeout))
	}
	if listener.StickySession != "" {
		stickySession := v.FieldByName("StickySession")
		stickySession.SetString(string(listener.StickySession))
	}
	if listener.StickySessionType != "" {
		stickySessionType := v.FieldByName("StickySessionType")
		stickySessionType.SetString(listener.StickySessionType)
	}
	if listener.XForwardedForProto != "" {
		xForwardedForProto := v.FieldByName("XForwardedForProto")
		xForwardedForProto.SetString(string(listener.XForwardedForProto))
	}
	if listener.XForwardedForSLBPort != "" {
		xForwardedForSLBPORT := v.FieldByName("XForwardedForSLBPORT")
		xForwardedForSLBPORT.SetString(string(listener.XForwardedForSLBPort))
	}
	if listener.XForwardedForClientSrcPort != "" {
		xForwardedForClientSrcPort := v.FieldByName("XForwardedForClientSrcPort")
		xForwardedForClientSrcPort.SetString(string(listener.XForwardedForClientSrcPort))
	}

	certId := v.FieldByName("ServerCertificateId")
	certId.SetString(listener.CertId)

	if listener.IdleTimeout != 0 {
		idleTimeout := v.FieldByName("IdleTimeout")
		idleTimeout.SetString(strconv.Itoa(listener.IdleTimeout))
	}
	if listener.EnableHttp2 != "" {
		enabled := v.FieldByName("EnableHttp2")
		enabled.SetString(string(listener.EnableHttp2))
	}
	if listener.RequestTimeout != 0 {
		requestTimeout := v.FieldByName("RequestTimeout")
		requestTimeout.SetString(strconv.Itoa(listener.RequestTimeout))
	}
	if listener.TLSCipherPolicy != "" {
		tlsPolicy := v.FieldByName("TLSCipherPolicy")
		tlsPolicy.SetString(listener.TLSCipherPolicy)
	}
}

func loadTCPListener(config slb.TCPListenerConfig, listener *model.ListenerAttribute) {
	persistenceTimeout := config.PersistenceTimeout
	listener.PersistenceTimeout = &persistenceTimeout
	listener.EstablishedTimeout = config.EstablishedTimeout
	listener.ConnectionDrain = model.FlagType(config.ConnectionDrain)
	listener.ConnectionDrainTimeout = config.ConnectionDrainTimeout
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckConnectTimeout = config.HealthCheckConnectTimeout
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckInterval = config.HealthCheckInterval
	listener.HealthCheckHttpCode = config.HealthCheckHttpCode
	listener.HealthCheckDomain = config.HealthCheckDomain
	listener.HealthCheckURI = config.HealthCheckURI
	listener.HealthCheckType = config.HealthCheckType
	listener.HealthCheckSwitch = model.FlagType(config.HealthCheckSwitch)
	if config.ProxyProtocolV2Enabled != "" {
		listener.EnableProxyProtocolV2 = tea.Bool(config.ProxyProtocolV2Enabled == "true")
	}
}

func loadUDPListener(config slb.UDPListenerConfig, listener *model.ListenerAttribute) {
	listener.ConnectionDrain = model.FlagType(config.ConnectionDrain)
	listener.ConnectionDrainTimeout = config.ConnectionDrainTimeout
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckConnectTimeout = config.HealthCheckConnectTimeout
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckInterval = config.HealthCheckInterval
	listener.HealthCheckSwitch = model.FlagType(config.HealthCheckSwitch)
	if config.ProxyProtocolV2Enabled != "" {
		listener.EnableProxyProtocolV2 = tea.Bool(config.ProxyProtocolV2Enabled == "true")
	}
}

func loadHTTPListener(config slb.HTTPListenerConfig, listener *model.ListenerAttribute) {
	listener.StickySession = model.FlagType(config.StickySession)
	listener.StickySessionType = config.StickySessionType
	listener.CookieTimeout = config.CookieTimeout
	listener.Cookie = config.Cookie
	listener.RequestTimeout = config.RequestTimeout
	listener.XForwardedForProto = model.FlagType(config.XForwardedForProto)
	listener.XForwardedForSLBPort = model.FlagType(config.XForwardedForSLBPORT)
	listener.XForwardedForClientSrcPort = model.FlagType(config.XForwardedForClientSrcPort)
	listener.IdleTimeout = config.IdleTimeout
	listener.HealthCheck = model.FlagType(config.HealthCheck)
	listener.HealthCheckDomain = config.HealthCheckDomain
	listener.HealthCheckURI = config.HealthCheckURI
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckTimeout = config.HealthCheckTimeout
	listener.HealthCheckInterval = config.HealthCheckInterval
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckHttpCode = config.HealthCheckHttpCode
	listener.HealthCheckMethod = config.HealthCheckMethod
	listener.ListenerForward = model.FlagType(config.ListenerForward)
	listener.ForwardPort = config.ForwardPort
}

func loadHTTPSListener(config slb.HTTPSListenerConfig, listener *model.ListenerAttribute) {
	listener.StickySession = model.FlagType(config.StickySession)
	listener.StickySessionType = config.StickySessionType
	listener.CookieTimeout = config.CookieTimeout
	listener.Cookie = config.Cookie
	listener.RequestTimeout = config.RequestTimeout
	listener.XForwardedForProto = model.FlagType(config.XForwardedForProto)
	listener.XForwardedForSLBPort = model.FlagType(config.XForwardedForSLBPORT)
	listener.XForwardedForClientSrcPort = model.FlagType(config.XForwardedForClientSrcPort)
	listener.IdleTimeout = config.IdleTimeout
	listener.HealthCheck = model.FlagType(config.HealthCheck)
	listener.HealthCheckDomain = config.HealthCheckDomain
	listener.HealthCheckURI = config.HealthCheckURI
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckTimeout = config.HealthCheckTimeout
	listener.HealthCheckInterval = config.HealthCheckInterval
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckHttpCode = config.HealthCheckHttpCode
	listener.HealthCheckMethod = config.HealthCheckMethod
	listener.CertId = config.ServerCertificateId
	listener.EnableHttp2 = model.FlagType(config.EnableHttp2)
	listener.TLSCipherPolicy = config.TLSCipherPolicy
}
