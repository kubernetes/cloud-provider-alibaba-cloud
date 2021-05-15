package slbprd

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"reflect"
	"strconv"
)

func (p ProviderSLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	req := slb.CreateDescribeLoadBalancerListenersRequest()
	req.LoadBalancerId = &[]string{lbId}
	resp, err := p.auth.SLB.DescribeLoadBalancerListeners(req)
	if err != nil {
		return nil, err
	}
	var listeners []model.ListenerAttribute
	for _, lis := range resp.Listeners {
		listeners = append(listeners, model.ListenerAttribute{
			Description:  lis.Description,
			ListenerPort: lis.ListenerPort,
			Protocol:     lis.ListenerProtocol,
			Bandwidth:    lis.Bandwidth,
			Scheduler:    lis.Scheduler,
			VGroupId:     lis.VServerGroupId,
		})
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
	_, err := p.auth.SLB.CreateLoadBalancerTCPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerTCPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerTCPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerTCPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerUDPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerUDPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerUDPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerUDPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerUDPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerUDPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerHTTPListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	_, err := p.auth.SLB.CreateLoadBalancerHTTPListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerHTTPListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	_, err := p.auth.SLB.SetLoadBalancerHTTPListenerAttribute(req)
	return err
}

func (p ProviderSLB) CreateLoadBalancerHTTPSListener(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateCreateLoadBalancerHTTPSListenerRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
	if listener.CertId != "" {
		req.ServerCertificateId = listener.CertId
	}
	_, err := p.auth.SLB.CreateLoadBalancerHTTPSListener(req)
	return err
}

func (p ProviderSLB) SetLoadBalancerHTTPSListenerAttribute(
	ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	req := slb.CreateSetLoadBalancerHTTPSListenerAttributeRequest()
	req.LoadBalancerId = lbId
	setGenericListenerValue(req, &listener)
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

	if listener.Bandwidth != 0 {
		bandwidth := v.FieldByName("Bandwidth")
		bandwidth.SetString(strconv.Itoa(listener.Bandwidth))
	}
	if listener.Scheduler != "" {
		scheduler := v.FieldByName("Scheduler")
		scheduler.SetString(listener.Scheduler)
	}
	if listener.HealthyThreshold != 0 {
		healthyThreshold := v.FieldByName("HealthyThreshold")
		healthyThreshold.SetString(strconv.Itoa(listener.HealthyThreshold))
	}
	if listener.UnhealthyThreshold != 0 {
		unhealthyThreshold := v.FieldByName("UnhealthyThreshold")
		unhealthyThreshold.SetString(strconv.Itoa(listener.UnhealthyThreshold))
	}
	if listener.HealthCheckConnectTimeout != 0 {
		connectTimeout := v.FieldByName("HealthCheckConnectTimeout")
		connectTimeout.SetString(strconv.Itoa(listener.HealthCheckConnectTimeout))
	}
	if listener.HealthCheckConnectPort != 0 {
		connectPort := v.FieldByName("HealthCheckConnectPort")
		connectPort.SetString(strconv.Itoa(listener.HealthCheckConnectPort))
	}
	if listener.HealthCheckInterval != 0 {
		interval := v.FieldByName("HealthCheckInterval")
		interval.SetString(strconv.Itoa(listener.HealthCheckInterval))
	}
	if listener.HealthCheckDomain != "" {
		domain := v.FieldByName("HealthCheckDomain")
		domain.SetString(listener.HealthCheckDomain)
	}
	if listener.HealthCheckURI != "" {
		uri := v.FieldByName("HealthCheckURI")
		uri.SetString(listener.HealthCheckURI)
	}
	if listener.HealthCheckHttpCode != "" {
		httpCode := v.FieldByName("HealthCheckHttpCode")
		httpCode.SetString(listener.HealthCheckHttpCode)
	}
	if listener.HealthCheckType != "" {
		healthCheckType := v.FieldByName("HealthCheckType")
		healthCheckType.SetString(listener.HealthCheckType)
	}
	if listener.AclType != "" {
		aclType := v.FieldByName("AclType")
		aclType.SetString(listener.AclType)
	}
	if listener.AclStatus != "" {
		aclStatus := v.FieldByName("AclStatus")
		aclStatus.SetString(string(listener.AclStatus))
	}
}
