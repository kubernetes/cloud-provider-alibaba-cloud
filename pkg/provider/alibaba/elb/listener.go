package elb

import (
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/klog/v2"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ens"
)

func (e ELBProvider) FindEdgeLoadBalancerListener(ctx context.Context, lbId string, listeners *elbmodel.EdgeListeners) error {
	req := ens.CreateDescribeLoadBalancerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId

	resp, err := e.auth.ELB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerAttribute", err)
	}
	if resp == nil {
		klog.Errorf("RequestId: %s, loadbalancer Id %s DescribeLoadBalancerAttribute response is nil", resp.RequestId, lbId)
		return fmt.Errorf("find no loadbalancer by id %s", lbId)
	}
	for _, v := range resp.ListenerPortsAndProtocols {
		listeners.BackListener = append(listeners.BackListener, elbmodel.EdgeListenerAttribute{
			ListenerProtocol: v.ListenerProtocol,
			ListenerPort:     v.ListenerPort,
		})
	}
	return nil
}

func (e ELBProvider) DescribeEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateDescribeLoadBalancerTCPListenerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	resp, err := e.auth.ELB.DescribeLoadBalancerTCPListenerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerTCPListenerAttribute", err)
	}
	return loadTCPListenResponse(resp, listener)
}

func (e ELBProvider) DescribeEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateDescribeLoadBalancerUDPListenerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	resp, err := e.auth.ELB.DescribeLoadBalancerUDPListenerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerUDPListenerAttribute", err)
	}

	return LoadUDPListenResponse(resp, listener)
}

func loadTCPListenResponse(resp *ens.DescribeLoadBalancerTCPListenerAttributeResponse, listener *elbmodel.EdgeListenerAttribute) error {
	listener.ListenerProtocol = elbmodel.ProtocolTCP
	listener.ListenerPort = resp.ListenerPort
	listener.Description = resp.Description
	listener.Scheduler = resp.Scheduler
	listener.Status = resp.Status
	listener.PersistenceTimeout = resp.PersistenceTimeout
	listener.EstablishedTimeout = resp.EstablishedTimeout
	if resp.HealthCheck != "on" {
		klog.Warningf("listen port %s:%d close healthy check", listener.ListenerProtocol, listener.ListenerPort)
		return nil
	}
	listener.HealthCheckType = resp.HealthCheckType
	listener.HealthThreshold = resp.HealthyThreshold
	listener.UnhealthyThreshold = resp.UnhealthyThreshold
	listener.HealthCheckConnectTimeout = resp.HealthCheckConnectTimeout
	listener.HealthCheckInterval = resp.HealthCheckInterval
	listener.HealthCheckConnectPort = resp.HealthCheckConnectPort
	return nil
}

func LoadUDPListenResponse(resp *ens.DescribeLoadBalancerUDPListenerAttributeResponse, listener *elbmodel.EdgeListenerAttribute) error {
	listener.ListenerProtocol = elbmodel.ProtocolUDP
	listener.ListenerPort = resp.ListenerPort
	listener.Description = resp.Description
	listener.Scheduler = resp.Scheduler
	listener.Status = resp.Status
	if resp.HealthCheck != "on" {
		klog.Warningf("listen port %s:%d close healthy check", listener.ListenerProtocol, listener.ListenerPort)
		return nil
	}
	listener.HealthThreshold = resp.HealthyThreshold
	listener.UnhealthyThreshold = resp.UnhealthyThreshold
	listener.HealthCheckConnectTimeout = resp.HealthCheckConnectTimeout
	listener.HealthCheckInterval = resp.HealthCheckInterval
	listener.HealthCheckConnectPort = resp.HealthCheckConnectPort
	return nil
}

func (e ELBProvider) DescribeEdgeLoadBalancerHTTPListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateDescribeLoadBalancerHTTPListenerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	resp, err := e.auth.ELB.DescribeLoadBalancerHTTPListenerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerHTTPListenerAttribute", err)
	}

	listener.Description = resp.Description
	listener.ListenerPort = resp.ListenerPort
	listener.ListenerProtocol = elbmodel.ProtocolHTTP
	return nil
}

func (e ELBProvider) DescribeEdgeLoadBalancerHTTPSListener(ctx context.Context, lbId string, port int, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateDescribeLoadBalancerHTTPSListenerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	resp, err := e.auth.ELB.DescribeLoadBalancerHTTPSListenerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerHTTPSListenerAttribute", err)
	}
	listener.Description = resp.Description
	listener.ListenerPort = resp.ListenerPort
	listener.ListenerProtocol = elbmodel.ProtocolHTTPS
	return nil
}

func (e ELBProvider) StartEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	req := ens.CreateStartLoadBalancerListenerRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerProtocol = protocol
	req.ListenerPort = requests.NewInteger(port)
	_, err := e.auth.ELB.StartLoadBalancerListener(req)
	if err != nil {
		return util.SDKError("StartLoadBalancerListener", err)
	}
	return nil
}

func (e ELBProvider) StopEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	req := ens.CreateStopLoadBalancerListenerRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerProtocol = protocol
	req.ListenerPort = requests.NewInteger(port)
	_, err := e.auth.ELB.StopLoadBalancerListener(req)
	if err != nil {
		return util.SDKError("StopLoadBalancerListener", err)
	}
	return nil
}

func (e ELBProvider) CreateEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateCreateLoadBalancerTCPListenerRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.Description = listener.Description
	req.ListenerPort = requests.NewInteger(listener.ListenerPort)
	req.Scheduler = listener.Scheduler
	req.HealthyThreshold = requests.NewInteger(listener.HealthThreshold)
	req.UnhealthyThreshold = requests.NewInteger(listener.UnhealthyThreshold)
	req.HealthCheckConnectTimeout = requests.NewInteger(listener.HealthCheckConnectTimeout)
	req.HealthCheckInterval = requests.NewInteger(listener.HealthCheckInterval)
	req.HealthCheckType = listener.HealthCheckType
	req.PersistenceTimeout = requests.NewInteger(listener.PersistenceTimeout)
	req.EstablishedTimeout = requests.NewInteger(listener.EstablishedTimeout)
	if listener.HealthCheckConnectPort != 0 {
		req.HealthCheckConnectPort = requests.NewInteger(listener.HealthCheckConnectPort)
	}

	_, err := e.auth.ELB.CreateLoadBalancerTCPListener(req)
	if err != nil {
		return util.SDKError("CreateLoadBalancerTCPListener", err)
	}
	return nil
}

func (e ELBProvider) CreateEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateCreateLoadBalancerUDPListenerRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.Description = listener.Description
	req.ListenerPort = requests.NewInteger(listener.ListenerPort)
	req.Scheduler = listener.Scheduler
	req.HealthyThreshold = requests.NewInteger(listener.HealthThreshold)
	req.UnhealthyThreshold = requests.NewInteger(listener.UnhealthyThreshold)
	req.HealthCheckConnectTimeout = requests.NewInteger(listener.HealthCheckConnectTimeout)
	req.HealthCheckInterval = requests.NewInteger(listener.HealthCheckInterval)
	if listener.HealthCheckConnectPort != 0 {
		req.HealthCheckConnectPort = requests.NewInteger(listener.HealthCheckConnectPort)
	}
	_, err := e.auth.ELB.CreateLoadBalancerUDPListener(req)
	if err != nil {
		return util.SDKError("CreateLoadBalancerUDPListener", err)
	}
	return nil
}

func (e ELBProvider) ModifyEdgeLoadBalancerTCPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateSetLoadBalancerTCPListenerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(listener.ListenerPort)
	req.Scheduler = listener.Scheduler
	req.HealthyThreshold = requests.NewInteger(listener.HealthThreshold)
	req.UnhealthyThreshold = requests.NewInteger(listener.UnhealthyThreshold)
	req.HealthCheckConnectTimeout = requests.NewInteger(listener.HealthCheckConnectTimeout)
	req.HealthCheckInterval = requests.NewInteger(listener.HealthCheckInterval)
	req.HealthCheckType = listener.HealthCheckType
	if listener.HealthCheckConnectPort != 0 {
		req.HealthCheckConnectPort = requests.NewInteger(listener.HealthCheckConnectPort)
	}
	req.PersistenceTimeout = requests.NewInteger(listener.PersistenceTimeout)
	req.EstablishedTimeout = requests.NewInteger(listener.EstablishedTimeout)

	_, err := e.auth.ELB.SetLoadBalancerTCPListenerAttribute(req)
	if err != nil {
		return util.SDKError("SetLoadBalancerTCPListenerAttribute", err)
	}
	return nil
}

func (e ELBProvider) ModifyEdgeLoadBalancerUDPListener(ctx context.Context, lbId string, listener *elbmodel.EdgeListenerAttribute) error {
	req := ens.CreateSetLoadBalancerUDPListenerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(listener.ListenerPort)
	req.Scheduler = listener.Scheduler
	req.HealthyThreshold = requests.NewInteger(listener.HealthThreshold)
	req.UnhealthyThreshold = requests.NewInteger(listener.UnhealthyThreshold)
	req.HealthCheckConnectTimeout = requests.NewInteger(listener.HealthCheckConnectTimeout)
	req.HealthCheckInterval = requests.NewInteger(listener.HealthCheckInterval)
	if listener.HealthCheckConnectPort != 0 {
		req.HealthCheckConnectPort = requests.NewInteger(listener.HealthCheckConnectPort)
	}
	_, err := e.auth.ELB.SetLoadBalancerUDPListenerAttribute(req)
	if err != nil {
		return util.SDKError("SetLoadBalancerUDPListenerAttribute", err)
	}
	return nil
}

func (e ELBProvider) DeleteEdgeLoadBalancerListener(ctx context.Context, lbId string, port int, protocol string) error {
	req := ens.CreateDeleteLoadBalancerListenerRequest()
	req.Scheme = "http"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.ListenerPort = requests.NewInteger(port)
	req.ListenerProtocol = protocol
	_, err := e.auth.ELB.DeleteLoadBalancerListener(req)
	if err != nil {
		return util.SDKError("DeleteEdgeLoadBalancerListener", err)
	}
	return nil
}
