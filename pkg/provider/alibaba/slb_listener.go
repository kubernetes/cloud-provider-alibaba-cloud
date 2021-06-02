package alibaba

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/klog"
	"reflect"
	"strconv"
)

func (p ProviderSLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	req := slb.CreateDescribeLoadBalancerListenersRequest()
	req.LoadBalancerId = &[]string{lbId}
	req.MaxResults = requests.NewInteger(50)

	var respListeners []slb.ListenerInDescribeLoadBalancerListeners
	for {
		resp, err := p.auth.SLB.DescribeLoadBalancerListeners(req)
		if err != nil {
			return nil, err
		}
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
			klog.Warningf("listener description [%s], not expected format. skip user managed port", lis.Description, err.Error())
			continue
		}
		n.NamedKey = namedKey

		listeners = append(listeners, n)
	}
	return listeners, nil
}

func (p ProviderSLB) StartLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	req := slb.CreateStartLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	_, err := p.auth.SLB.StartLoadBalancerListener(req)
	return err
}

func (p ProviderSLB) StopLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	req := slb.CreateStopLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	_, err := p.auth.SLB.StopLoadBalancerListener(req)
	return err
}

func (p ProviderSLB) DeleteLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	req := slb.CreateDeleteLoadBalancerListenerRequest()
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)

	_, err := p.auth.SLB.DeleteLoadBalancerListener(req)
	return err

}

func (p ProviderSLB) CreateLoadBalancerTCPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerTCPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	// set params only for tcp
	setTCPListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerTCPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerTCPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerTCPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	// set params only for tcp
	setTCPListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerTCPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerUDPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerUDPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	// set params only for udp
	setUDPListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerUDPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerUDPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerUDPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	// set params only for udp
	setUDPListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerUDPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerHTTPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	// set params only for http
	setHTTPListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerHTTPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerHTTPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	// set params only for http
	setHTTPListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerHTTPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerHTTPSListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPSListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	// set params only for https
	setHTTPSListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerHTTPSListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerHTTPSListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPSListenerAttributeRequest()
	req.LoadBalancerId = lbId
	req.VServerGroup = string(model.OnFlag)
	setGenericListenerValue(req, &listener)
	// set params only for https
	setHTTPSListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerHTTPSListenerAttribute(req)
	return err
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
	if listener.HealthCheckInterval != 0 {
		healthCheckInterval := v.FieldByName("HealthCheckInterval")
		healthCheckInterval.SetString(strconv.Itoa(listener.HealthCheckInterval))
	}

	if listener.ConnectionDrain != "" {
		connectionDrain := v.FieldByName("ConnectionDrain")
		connectionDrain.SetString(string(listener.ConnectionDrain))
	}
	if listener.ConnectionDrainTimeout != 0 {
		connectionDrainTimeout := v.FieldByName("ConnectionDrainTimeout")
		connectionDrainTimeout.SetString(strconv.Itoa(listener.ConnectionDrainTimeout))
	}
}

func setUDPListenerValue(req interface{}, listener *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()

	if listener.HealthCheckConnectTimeout != 0 {
		healthCheckConnectTimeout := v.FieldByName("HealthCheckConnectTimeout")
		healthCheckConnectTimeout.SetString(strconv.Itoa(listener.HealthCheckConnectTimeout))
	}
	if listener.HealthCheckInterval != 0 {
		healthCheckInterval := v.FieldByName("HealthCheckInterval")
		healthCheckInterval.SetString(strconv.Itoa(listener.HealthCheckInterval))
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
	if listener.HealthCheckInterval != 0 {
		healthCheckInterval := v.FieldByName("HealthCheckInterval")
		healthCheckInterval.SetString(strconv.Itoa(listener.HealthCheckInterval))
	}
	if listener.HealthCheckDomain != "" {
		healthCheckDomain := v.FieldByName("HealthCheckDomain")
		healthCheckDomain.SetString(listener.HealthCheckDomain)
	}
	if listener.HealthCheckTimeout != 0 {
		healthCheckTimeout := v.FieldByName("HealthCheckTimeout")
		healthCheckTimeout.SetString(strconv.Itoa(listener.HealthCheckTimeout))
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
	if listener.ListenerForward != "" {
		listenerForward := v.FieldByName("ListenerForward")
		listenerForward.SetString(string(listener.ListenerForward))
	}
	if listener.ForwardPort != 0 {
		forwardPort := v.FieldByName("ForwardPort")
		forwardPort.SetString(strconv.Itoa(listener.ForwardPort))
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
	if listener.HealthCheckInterval != 0 {
		healthCheckInterval := v.FieldByName("HealthCheckInterval")
		healthCheckInterval.SetString(strconv.Itoa(listener.HealthCheckInterval))
	}
	if listener.HealthCheckDomain != "" {
		healthCheckDomain := v.FieldByName("HealthCheckDomain")
		healthCheckDomain.SetString(listener.HealthCheckDomain)
	}
	if listener.HealthCheckTimeout != 0 {
		healthCheckTimeout := v.FieldByName("HealthCheckTimeout")
		healthCheckTimeout.SetString(strconv.Itoa(listener.HealthCheckTimeout))
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
	if listener.CertId != "" {
		certId := v.FieldByName("ServerCertificateId")
		certId.SetString(listener.CertId)
	}
}

func loadTCPListener(config slb.TCPListenerConfig, listener *model.ListenerAttribute) {
	persistenceTimeout := config.PersistenceTimeout
	listener.PersistenceTimeout = &persistenceTimeout
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
}

func loadUDPListener(config slb.UDPListenerConfig, listener *model.ListenerAttribute) {
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckConnectTimeout = config.HealthCheckConnectTimeout
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckInterval = config.HealthCheckInterval
}

func loadHTTPListener(config slb.HTTPListenerConfig, listener *model.ListenerAttribute) {
	listener.StickySession = model.FlagType(config.StickySession)
	listener.StickySessionType = config.StickySessionType
	listener.CookieTimeout = config.CookieTimeout
	listener.Cookie = config.Cookie
	listener.HealthCheck = model.FlagType(config.HealthCheck)
	listener.HealthCheckDomain = config.HealthCheckDomain
	listener.HealthCheckURI = config.HealthCheckURI
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckTimeout = config.HealthCheckTimeout
	listener.HealthCheckInterval = config.HealthCheckInterval
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckHttpCode = config.HealthCheckHttpCode
	listener.ListenerForward = model.FlagType(config.ListenerForward)
	listener.ForwardPort = config.ForwardPort
}

func loadHTTPSListener(config slb.HTTPSListenerConfig, listener *model.ListenerAttribute) {
	listener.StickySession = model.FlagType(config.StickySession)
	listener.StickySessionType = config.StickySessionType
	listener.CookieTimeout = config.CookieTimeout
	listener.Cookie = config.Cookie
	listener.HealthCheck = model.FlagType(config.HealthCheck)
	listener.HealthCheckDomain = config.HealthCheckDomain
	listener.HealthCheckURI = config.HealthCheckURI
	listener.HealthyThreshold = config.HealthyThreshold
	listener.UnhealthyThreshold = config.UnhealthyThreshold
	listener.HealthCheckTimeout = config.HealthCheckTimeout
	listener.HealthCheckInterval = config.HealthCheckInterval
	listener.HealthCheckConnectPort = config.HealthCheckConnectPort
	listener.HealthCheckHttpCode = config.HealthCheckHttpCode
	listener.CertId = config.ServerCertificateId
}
