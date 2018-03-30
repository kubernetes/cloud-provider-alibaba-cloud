package alicloud

import (
	"errors"
	"strings"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

type Listener interface {
	AttemptAdd(client ClientSLBSDK, targetPort v1.ServicePort, transformProtocol string) error
	AttemptRemove(client ClientSLBSDK, targetPort int, transformProtocol string) error
	AttemptUpdate(client ClientSLBSDK, targetPort v1.ServicePort, transformProtocol string) error
}

type BaseListener struct {
	doAdd        func(client ClientSLBSDK, port v1.ServicePort) error
	doUpdate     func(client ClientSLBSDK, port v1.ServicePort) error
	loadbalancer *slb.LoadBalancerType
	service      *v1.Service
}

func (l *BaseListener) AttemptRemove(client ClientSLBSDK, targetPort int, backProtocol string) error {
	found := false
	protocol := serviceAnnotation(l.service, ServiceAnnotationLoadBalancerProtocolPort)
	for _, port := range l.service.Spec.Ports {
		proto, err := getProtocol(protocol, port)
		if err != nil {
			return err
		}
		if port.Port == int32(targetPort) &&
			strings.ToUpper(proto) == strings.ToUpper(backProtocol){
			found = true
			break
		}
	}
	if !found {
		glog.V(4).Infof("alicloud: attempt to [remove] listener, [%d]\n", targetPort)
		err := client.StopLoadBalancerListener(l.loadbalancer.LoadBalancerId, targetPort)
		if err != nil {
			return err
		}
		return client.DeleteLoadBalancerListener(l.loadbalancer.LoadBalancerId, targetPort)
	}
	return nil
}
func (l *BaseListener) AttemptAdd(client ClientSLBSDK, targetPort v1.ServicePort, transformProtocol string) error {
	found := false
	ports := l.loadbalancer.ListenerPortsAndProtocol
	for _, port := range ports.ListenerPortAndProtocol {
		if port.ListenerPort == int(targetPort.Port) &&
			strings.ToUpper(port.ListenerProtocol) == strings.ToUpper(transformProtocol) {
			found = true
			break
		}
	}
	if !found {
		glog.V(4).Infof("alicloud: attempt to [add] listener, port=[%d], nodeport=[%d]\n", targetPort.Port,targetPort.NodePort)
		if l.doAdd == nil {
			return errors.New("AttemptAdd needs doAdd implementation.\n")
		}
		err := l.doAdd(client, targetPort)
		if err != nil {
			return err
		}
		// todo : here should retry
		return client.StartLoadBalancerListener(l.loadbalancer.LoadBalancerId, int(targetPort.Port))
	}
	return nil
}
func (l *BaseListener) AttemptUpdate(client ClientSLBSDK, targetPort v1.ServicePort, transformProtocol string) error {
	found := false
	backports := l.loadbalancer.ListenerPortsAndProtocol
	for _, backport := range backports.ListenerPortAndProtocol {
		if backport.ListenerPort == int(targetPort.Port)&&
			strings.ToUpper(backport.ListenerProtocol) == strings.ToUpper(transformProtocol) {
			found = true
			break
		}
	}
	if !found {
		// no update needed, just return
		return nil
	}
	if l.doUpdate == nil {
		return errors.New("AttemptUpdate need doUpdate implementation.\n")
	}
	glog.V(4).Infof("alicloud: attempt to [update] listener, port=[%d], nodeport=[%d]\n", targetPort.Port, targetPort.NodePort)
	if err := l.doUpdate(client, targetPort); err != nil {
		return err
	}
	return client.StartLoadBalancerListener(l.loadbalancer.LoadBalancerId, int(targetPort.Port))
}

//####################################################################################
//# @date: 2018-03-16
//# @name: CommonListener
//# @desc: LoadBalancer TCP listener implementation.
//####################################################################################
type CommonListener struct {
	BaseListener
}

func NewTCP(service *v1.Service, loadbalancer *slb.LoadBalancerType) Listener {

	def, request := ExtractAnnotationRequest(service)
	doAdd := func(client ClientSLBSDK, targetPort v1.ServicePort) error {
		return client.CreateLoadBalancerTCPListener(
			&slb.CreateLoadBalancerTCPListenerArgs{
				LoadBalancerId:    loadbalancer.LoadBalancerId,
				ListenerPort:      int(targetPort.Port),
				BackendServerPort: int(targetPort.NodePort),
				//Health Check
				Bandwidth: def.Bandwidth,

				HealthCheckType:           def.HealthCheckType,
				HealthCheckURI:            def.HealthCheckURI,
				HealthCheckConnectPort:    def.HealthCheckConnectPort,
				HealthyThreshold:          def.HealthyThreshold,
				UnhealthyThreshold:        def.UnhealthyThreshold,
				HealthCheckConnectTimeout: def.HealthCheckConnectTimeout,
				HealthCheckInterval:       def.HealthCheckInterval,
			},
		)
	}
	doUpdate := func(client ClientSLBSDK, port v1.ServicePort) error {

		response, err := client.DescribeLoadBalancerTCPListenerAttribute(loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.CreateLoadBalancerTCPListenerArgs{
			LoadBalancerId:    loadbalancer.LoadBalancerId,
			ListenerPort:      int(port.Port),
			BackendServerPort: int(port.NodePort),
			//Health Check
			Bandwidth: response.Bandwidth,

			HealthCheckType:           response.HealthCheckType,
			HealthCheckURI:            response.HealthCheckURI,
			HealthCheckConnectPort:    response.HealthCheckConnectPort,
			HealthyThreshold:          response.HealthyThreshold,
			UnhealthyThreshold:        response.UnhealthyThreshold,
			HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
			HealthCheckInterval:       response.HealthCheckInterval,
		}
		recreate := false
		if def.Bandwidth != response.Bandwidth {
			recreate = true
			config.Bandwidth = def.Bandwidth
			glog.V(2).Infof("TCP listener checker [bandwidth] changed, request=%d. response=%d", def.Bandwidth, response.Bandwidth)
		}

		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			recreate = true
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("TCP listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
		}
		// todo: perform healthcheck update.
		if def.HealthCheckType != response.HealthCheckType {
			recreate = true
			config.HealthCheckType = def.HealthCheckType
		}
		if request.HealthCheckURI != "" &&
			def.HealthCheckURI != response.HealthCheckURI {
			recreate = true
			config.HealthCheckURI = def.HealthCheckURI
		}
		if request.HealthCheckConnectPort != 0 &&
			def.HealthCheckConnectPort != response.HealthCheckConnectPort {
			recreate = true
			config.HealthCheckConnectPort = def.HealthCheckConnectPort
		}
		if request.HealthyThreshold != 0 &&
			def.HealthyThreshold != response.HealthyThreshold {
			recreate = true
			config.HealthyThreshold = def.HealthyThreshold
		}
		if request.UnhealthyThreshold != 0 &&
			def.UnhealthyThreshold != response.UnhealthyThreshold {
			recreate = true
			config.UnhealthyThreshold = def.UnhealthyThreshold
		}
		if request.HealthCheckConnectTimeout != 0 &&
			def.HealthCheckConnectTimeout != response.HealthCheckConnectTimeout {
			recreate = true
			config.HealthCheckConnectTimeout = def.HealthCheckConnectTimeout
		}
		if request.HealthCheckInterval != 0 &&
			def.HealthCheckInterval != response.HealthCheckInterval {
			recreate = true
			config.HealthCheckInterval = def.HealthCheckInterval
		}

		// because backendserverport update is not supported, so we use recreate instead.
		if !recreate {
			glog.V(2).Infof("alicloud: tcp listener did not change, skip [update], port=[%d], nodeport=[%d]\n",port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("TCP listener checker changed, request recreate [%s]\n",loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(def))
		glog.V(5).Infof(PrettyJson(response))
		if err := client.DeleteLoadBalancerListener(
			loadbalancer.LoadBalancerId, int(port.Port)); err != nil {

			return err
		}
		return client.CreateLoadBalancerTCPListener(config)
	}
	return &CommonListener{
		BaseListener{
			loadbalancer: loadbalancer,
			service:      service,
			doAdd:        doAdd,
			doUpdate:     doUpdate,
		},
	}
}

//####################################################################################
//# @date: 2018-03-16
//# @name: ListenerUDP
//# @desc: LoadBalancer UDP listener implementation.
//####################################################################################

func NewUDP(service *v1.Service, loadbalancer *slb.LoadBalancerType) Listener {
	doUpdate := func(client ClientSLBSDK, port v1.ServicePort) error {
		request,_ := ExtractAnnotationRequest(service)
		response, err := client.DescribeLoadBalancerUDPListenerAttribute(loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.CreateLoadBalancerUDPListenerArgs{
			LoadBalancerId:    loadbalancer.LoadBalancerId,
			ListenerPort:      int(port.Port),
			BackendServerPort: int(port.NodePort),
			//Health Check
			Bandwidth: response.Bandwidth,

			//HealthCheckType:           response.HealthCheckType,
			//HealthCheckURI:            response.HealthCheckURI,
			HealthCheckConnectPort:    response.HealthCheckConnectPort,
			HealthyThreshold:          response.HealthyThreshold,
			UnhealthyThreshold:        response.UnhealthyThreshold,
			HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
			HealthCheckInterval:       response.HealthCheckInterval,
		}
		recreate := false
		if request.Bandwidth != response.Bandwidth {
			recreate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("UDP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			recreate = true
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("UDP listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
		}

		// todo: perform healthcheck update.
		// because backendserverport update is not supported, so we use recreate instead.
		if !recreate {
			glog.V(2).Infof("alicloud: udp listener did not change, skip [update], port=[%d], nodeport=[%d]\n",port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("UDP listener checker changed, request recreate [%s]\n",loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(request))
		glog.V(5).Infof(PrettyJson(response))
		if err := client.DeleteLoadBalancerListener(
			loadbalancer.LoadBalancerId, int(port.Port)); err != nil {

			return err
		}
		return client.CreateLoadBalancerUDPListener(config)
	}
	doAdd := func(client ClientSLBSDK, targetPort v1.ServicePort) error {
		request,_ := ExtractAnnotationRequest(service)
		return client.CreateLoadBalancerUDPListener(
			&slb.CreateLoadBalancerUDPListenerArgs{
				LoadBalancerId:    loadbalancer.LoadBalancerId,
				ListenerPort:      int(targetPort.Port),
				BackendServerPort: int(targetPort.NodePort),
				//Health Check
				Bandwidth: request.Bandwidth,

				//HealthCheckType:           request.HealthCheckType,
				//HealthCheckURI:            request.HealthCheckURI,
				HealthCheckConnectPort:    request.HealthCheckConnectPort,
				HealthyThreshold:          request.HealthyThreshold,
				UnhealthyThreshold:        request.UnhealthyThreshold,
				HealthCheckConnectTimeout: request.HealthCheckConnectTimeout,
				HealthCheckInterval:       request.HealthCheckInterval,
			},
		)
	}
	return &CommonListener{
		BaseListener{
			loadbalancer: loadbalancer,
			service:      service,
			doUpdate:     doUpdate,
			doAdd:        doAdd,
		},
	}
}

//####################################################################################
//# @date: 2018-03-16
//# @name: ListenerHTTP
//# @desc: LoadBalancer HTTP listener implementation.
//####################################################################################

func NewHTTP(service *v1.Service, loadbalancer *slb.LoadBalancerType) Listener {
	doUpdate := func(client ClientSLBSDK, port v1.ServicePort) error {
		request,_ := ExtractAnnotationRequest(service)
		response, err := client.DescribeLoadBalancerHTTPListenerAttribute(loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.CreateLoadBalancerHTTPListenerArgs{
			LoadBalancerId:    loadbalancer.LoadBalancerId,
			ListenerPort:      int(port.Port),
			BackendServerPort: int(port.NodePort),
			//Health Check
			Bandwidth: response.Bandwidth,

			//HealthCheckType:           response.HealthCheckType,
			HealthCheckURI:         response.HealthCheckURI,
			HealthCheckConnectPort: response.HealthCheckConnectPort,
			HealthyThreshold:       response.HealthyThreshold,
			UnhealthyThreshold:     response.UnhealthyThreshold,
			//HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
			HealthCheckInterval: response.HealthCheckInterval,
		}
		recreate := false
		if request.Bandwidth != response.Bandwidth {
			recreate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("HTTP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			recreate = true
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("HTTP listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
		}

		// todo: perform healthcheck update.
		// because backendserverport update is not supported, so we use recreate instead.
		if !recreate {
			glog.V(2).Infof("alicloud: http listener did not change, skip [update], port=[%d], nodeport=[%d]\n",port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("HTTP listener checker changed, request recreate [%s]\n",loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(request))
		glog.V(5).Infof(PrettyJson(response))
		if err := client.DeleteLoadBalancerListener(
			loadbalancer.LoadBalancerId, int(port.Port)); err != nil {

			return err
		}
		return client.CreateLoadBalancerHTTPListener(config)
	}
	doAdd := func(client ClientSLBSDK, targetPort v1.ServicePort) error {
		request,_ := ExtractAnnotationRequest(service)
		return client.CreateLoadBalancerHTTPListener(
			&slb.CreateLoadBalancerHTTPListenerArgs{
				LoadBalancerId:    loadbalancer.LoadBalancerId,
				ListenerPort:      int(targetPort.Port),
				BackendServerPort: int(targetPort.NodePort),
				//Health Check
				Bandwidth: request.Bandwidth,

				//HealthCheckType:           request.HealthCheckType,
				HealthCheckURI:         request.HealthCheckURI,
				HealthCheckConnectPort: request.HealthCheckConnectPort,
				HealthyThreshold:       request.HealthyThreshold,
				UnhealthyThreshold:     request.UnhealthyThreshold,
				//HealthCheckConnectTimeout: request.HealthCheckConnectTimeout,
				HealthCheckInterval: request.HealthCheckInterval,
			},
		)
	}
	return &CommonListener{
		BaseListener{
			loadbalancer: loadbalancer,
			service:      service,
			doUpdate:     doUpdate,
			doAdd:        doAdd,
		},
	}
}

//####################################################################################
//# @date: 2018-03-16
//# @name: ListenerHTTPS
//# @desc: LoadBalancer HTTPS listener implementation.
//####################################################################################

func NewHTTPS(service *v1.Service, loadbalancer *slb.LoadBalancerType) Listener {
	doUpdate := func(client ClientSLBSDK, port v1.ServicePort) error {
		request,_ := ExtractAnnotationRequest(service)
		response, err := client.DescribeLoadBalancerHTTPSListenerAttribute(loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.CreateLoadBalancerHTTPSListenerArgs{
			HTTPListenerType: slb.HTTPListenerType{
				LoadBalancerId:    response.LoadBalancerId,
				ListenerPort:      response.ListenerPort,
				BackendServerPort: response.BackendServerPort,
				//Health Check
				HealthCheck:   response.HealthCheck,
				Bandwidth:     response.Bandwidth,
				StickySession: response.StickySession,

				HealthCheckURI:         response.HealthCheckURI,
				HealthCheckConnectPort: response.HealthCheckConnectPort,
				HealthyThreshold:       response.HealthyThreshold,
				UnhealthyThreshold:     response.UnhealthyThreshold,
				HealthCheckTimeout:     response.HealthCheckTimeout,
				HealthCheckInterval:    response.HealthCheckInterval,
			},
			ServerCertificateId: response.ServerCertificateId,
		}

		recreate := false
		if request.Bandwidth != response.Bandwidth {
			recreate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("HTTPS listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			recreate = true
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("HTTPS listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
		}

		// todo: perform healthcheck update.
		// because backendserverport update is not supported, so we use recreate instead.
		if !recreate {
			glog.V(2).Infof("alicloud: https listener did not change, skip [update], port=[%d], nodeport=[%d]\n",port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("HTTPS listener checker changed, request recreate [%s]\n",loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(request))
		glog.V(5).Infof(PrettyJson(response))
		if err := client.DeleteLoadBalancerListener(
			loadbalancer.LoadBalancerId, int(port.Port)); err != nil {

			return err
		}
		return client.CreateLoadBalancerHTTPSListener(config)
	}
	doAdd := func(client ClientSLBSDK, targetPort v1.ServicePort) error {
		request,_ := ExtractAnnotationRequest(service)
		return client.CreateLoadBalancerHTTPSListener(
			&slb.CreateLoadBalancerHTTPSListenerArgs{
				HTTPListenerType: slb.HTTPListenerType{
					LoadBalancerId:    loadbalancer.LoadBalancerId,
					ListenerPort:      int(targetPort.Port),
					BackendServerPort: int(targetPort.NodePort),
					//Health Check
					HealthCheck: request.HealthCheck,
					Bandwidth:   request.Bandwidth,
					StickySession: slb.OffFlag,

					HealthCheckURI:         request.HealthCheckURI,
					HealthCheckConnectPort: request.HealthCheckConnectPort,
					HealthyThreshold:       request.HealthyThreshold,
					UnhealthyThreshold:     request.UnhealthyThreshold,
					HealthCheckTimeout:     request.HealthCheckTimeout,
					HealthCheckInterval:    request.HealthCheckInterval,
				},
				ServerCertificateId: request.CertID,
			},
		)
	}
	return &CommonListener{
		BaseListener{
			loadbalancer: loadbalancer,
			service:      service,
			doUpdate:     doUpdate,
			doAdd:        doAdd,
		},
	}
}

type factory struct {
	client       ClientSLBSDK
	service      *v1.Service
	loadbalancer *slb.LoadBalancerType
}

func NewListenerManager(client ClientSLBSDK,
	servcie *v1.Service, loadbalancer *slb.LoadBalancerType) *factory {
	return &factory{
		client:       client,
		service:      servcie,
		loadbalancer: loadbalancer,
	}
}

func (f *factory) newListener(proto string) Listener {

	switch proto {
	case "tcp":
		return NewTCP(f.service, f.loadbalancer)
	case "udp":
		return NewUDP(f.service, f.loadbalancer)
	case "http":
		return NewHTTP(f.service, f.loadbalancer)
	case "https":
		return NewHTTPS(f.service, f.loadbalancer)
	default:
		glog.Warningf("alicloud: unknown protocol specified. fallback to tcp.")
		return NewTCP(f.service, f.loadbalancer)
	}
}

//ApplyUpdate try to update current loadbalancer`s listeners based on the given service
func (f *factory) ApplyUpdate() error {
	protocol := serviceAnnotation(f.service, ServiceAnnotationLoadBalancerProtocolPort)
	// remove listeners which was no longer exist.
	port := f.loadbalancer.ListenerPortsAndProtocol.ListenerPortAndProtocol
	glog.V(4).Infof("alicloud: loadbalancer backend listener and protocol. [%v]\n",port )
	for _, port := range port {
		listener := f.newListener("tcp")
		if err := listener.AttemptRemove(f.client, port.ListenerPort, port.ListenerProtocol); err != nil {
			return err
		}
	}
	// attempt to add and update new or existing listeners
	for _, port := range f.service.Spec.Ports {
		proto, err := getProtocol(protocol, port)
		if err != nil {
			return err
		}
		listener := f.newListener(proto)
		if err := listener.AttemptAdd(f.client, port,proto); err != nil {
			return err
		}
		if err := listener.AttemptUpdate(f.client, port,proto); err != nil {
			return err
		}
	}
	return nil
}
