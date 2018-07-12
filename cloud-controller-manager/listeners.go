/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"strings"
)

type ListenerInterface interface {
	Add(port v1.ServicePort) error
	Remove(port v1.ServicePort) error
	Update(port v1.ServicePort) error
}

type BaseListener struct {
	doAdd    func(port v1.ServicePort) error
	doUpdate func(port v1.ServicePort) error
	doRemove func(port v1.ServicePort) error
	manager  *ListenerManager
}

//####################################################################################
//# @date: 2018-03-16
//# @name: CommonListener
//# @desc: LoadBalancer TCP listener implementation.
//####################################################################################
type CommonListener struct {
	BaseListener
}

func (com *CommonListener) Add(port v1.ServicePort) error {
	if com.doAdd != nil {
		return com.doAdd(port)
	}
	return errors.New("unimplemented Add()")
}

func (com *CommonListener) Update(port v1.ServicePort) error {
	if com.doAdd != nil {
		return com.doUpdate(port)
	}
	return errors.New("unimplemented Update()")
}

func (com *CommonListener) Remove(port v1.ServicePort) error {
	if com.doRemove != nil {
		return com.doRemove(port)
	}
	err := com.manager.client.StopLoadBalancerListener(com.manager.loadbalancer.LoadBalancerId, int(port.Port))
	if err != nil {
		return err
	}
	return com.manager.client.DeleteLoadBalancerListener(com.manager.loadbalancer.LoadBalancerId, int(port.Port))
}

func (f *ListenerManager) NewTCP() ListenerInterface {

	def, request := ExtractAnnotationRequest(f.service)
	doAdd := func(targetPort v1.ServicePort) error {
		return f.client.CreateLoadBalancerTCPListener(
			&slb.CreateLoadBalancerTCPListenerArgs{
				LoadBalancerId:    f.loadbalancer.LoadBalancerId,
				ListenerPort:      int(targetPort.Port),
				BackendServerPort: int(targetPort.NodePort),
				//Health Check
				Bandwidth:          def.Bandwidth,
				PersistenceTimeout: def.PersistenceTimeout,

				HealthCheckType:           def.HealthCheckType,
				HealthCheckURI:            def.HealthCheckURI,
				HealthCheckConnectPort:    def.HealthCheckConnectPort,
				HealthyThreshold:          def.HealthyThreshold,
				UnhealthyThreshold:        def.UnhealthyThreshold,
				HealthCheckConnectTimeout: def.HealthCheckConnectTimeout,
				HealthCheckInterval:       def.HealthCheckInterval,
				HealthCheck:               def.HealthCheck,
				HealthCheckDomain:         def.HealthCheckDomain,
				HealthCheckHttpCode:       def.HealthCheckHttpCode,
			},
		)
	}
	doUpdate := func(port v1.ServicePort) error {

		response, err := f.client.DescribeLoadBalancerTCPListenerAttribute(f.loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.SetLoadBalancerTCPListenerAttributeArgs{
			LoadBalancerId:    f.loadbalancer.LoadBalancerId,
			ListenerPort:      int(port.Port),
			BackendServerPort: int(port.NodePort),
			//Health Check
			Bandwidth:          response.Bandwidth,
			PersistenceTimeout: response.PersistenceTimeout,

			HealthCheckType:           response.HealthCheckType,
			HealthCheckURI:            response.HealthCheckURI,
			HealthCheckConnectPort:    response.HealthCheckConnectPort,
			HealthyThreshold:          response.HealthyThreshold,
			UnhealthyThreshold:        response.UnhealthyThreshold,
			HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
			HealthCheckInterval:       response.HealthCheckInterval,
			HealthCheck:               response.HealthCheck,
			HealthCheckHttpCode:       response.HealthCheckHttpCode,
			HealthCheckDomain:         response.HealthCheckDomain,
		}
		needUpdate := false
		if request.Bandwidth != 0 &&
			def.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = def.Bandwidth
			glog.V(2).Infof("TCP listener checker [bandwidth] changed, request=%d. response=%d", def.Bandwidth, response.Bandwidth)
		}

		// todo: perform healthcheck update.
		if def.HealthCheckType != response.HealthCheckType {
			needUpdate = true
			config.HealthCheckType = def.HealthCheckType
		}
		if request.HealthCheckURI != "" &&
			def.HealthCheckURI != response.HealthCheckURI {
			needUpdate = true
			config.HealthCheckURI = def.HealthCheckURI
		}
		if request.HealthCheckConnectPort != 0 &&
			def.HealthCheckConnectPort != response.HealthCheckConnectPort {
			needUpdate = true
			config.HealthCheckConnectPort = def.HealthCheckConnectPort
		}
		if request.HealthyThreshold != 0 &&
			def.HealthyThreshold != response.HealthyThreshold {
			needUpdate = true
			config.HealthyThreshold = def.HealthyThreshold
		}
		if request.UnhealthyThreshold != 0 &&
			def.UnhealthyThreshold != response.UnhealthyThreshold {
			needUpdate = true
			config.UnhealthyThreshold = def.UnhealthyThreshold
		}
		if request.HealthCheckConnectTimeout != 0 &&
			def.HealthCheckConnectTimeout != response.HealthCheckConnectTimeout {
			needUpdate = true
			config.HealthCheckConnectTimeout = def.HealthCheckConnectTimeout
		}
		if request.HealthCheckInterval != 0 &&
			def.HealthCheckInterval != response.HealthCheckInterval {
			needUpdate = true
			config.HealthCheckInterval = def.HealthCheckInterval
		}
		if def.PersistenceTimeout != response.PersistenceTimeout {
			needUpdate = true
			config.PersistenceTimeout = def.PersistenceTimeout
		}
		if request.HealthCheckHttpCode != "" &&
			def.HealthCheckHttpCode != response.HealthCheckHttpCode {
			needUpdate = true
			config.HealthCheckHttpCode = def.HealthCheckHttpCode
		}
		if request.HealthCheckDomain != "" &&
			def.HealthCheckDomain != response.HealthCheckDomain {
			needUpdate = true
			config.HealthCheckDomain = def.HealthCheckDomain
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("tcp listener [BackendServerPort] changed, request=%d. response=%d, recreate.", port.NodePort, response.BackendServerPort)
			err := f.client.DeleteLoadBalancerListener(f.loadbalancer.LoadBalancerId, int(port.Port))
			if err != nil {
				return err
			}
			return f.client.CreateLoadBalancerTCPListener((*slb.CreateLoadBalancerTCPListenerArgs)(config))
		}
		if !needUpdate {
			glog.V(2).Infof("alicloud: tcp listener did not change, skip [update], port=[%d], nodeport=[%d]\n", port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("TCP listener checker changed, request recreate [%s]\n", f.loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(def))
		glog.V(5).Infof(PrettyJson(response))
		return f.client.SetLoadBalancerTCPListenerAttribute(config)
	}
	return &CommonListener{
		BaseListener{
			manager:  f,
			doAdd:    doAdd,
			doUpdate: doUpdate,
		},
	}
}

//####################################################################################
//# @date: 2018-03-16
//# @name: ListenerUDP
//# @desc: LoadBalancer UDP listener implementation.
//####################################################################################

func (f *ListenerManager) NewUDP() ListenerInterface {
	def, request := ExtractAnnotationRequest(f.service)
	doUpdate := func(port v1.ServicePort) error {
		response, err := f.client.DescribeLoadBalancerUDPListenerAttribute(f.loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.SetLoadBalancerUDPListenerAttributeArgs{
			LoadBalancerId:    f.loadbalancer.LoadBalancerId,
			ListenerPort:      int(port.Port),
			BackendServerPort: int(port.NodePort),
			//Health Check
			Bandwidth:          response.Bandwidth,
			PersistenceTimeout: response.PersistenceTimeout,
			//HealthCheckType:           response.HealthCheckType,
			//HealthCheckURI:            response.HealthCheckURI,
			HealthCheckConnectPort:    response.HealthCheckConnectPort,
			HealthyThreshold:          response.HealthyThreshold,
			UnhealthyThreshold:        response.UnhealthyThreshold,
			HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
			HealthCheckInterval:       response.HealthCheckInterval,
			HealthCheck:               response.HealthCheck,
		}
		needUpdate := false
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("UDP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}

		// todo: perform healthcheck update.
		if request.HealthCheckConnectPort != 0 &&
			def.HealthCheckConnectPort != response.HealthCheckConnectPort {
			needUpdate = true
			config.HealthCheckConnectPort = def.HealthCheckConnectPort
		}
		if request.HealthyThreshold != 0 &&
			def.HealthyThreshold != response.HealthyThreshold {
			needUpdate = true
			config.HealthyThreshold = def.HealthyThreshold
		}
		if request.UnhealthyThreshold != 0 &&
			def.UnhealthyThreshold != response.UnhealthyThreshold {
			needUpdate = true
			config.UnhealthyThreshold = def.UnhealthyThreshold
		}
		if request.HealthCheckConnectTimeout != 0 &&
			def.HealthCheckConnectTimeout != response.HealthCheckConnectTimeout {
			needUpdate = true
			config.HealthCheckConnectTimeout = def.HealthCheckConnectTimeout
		}
		if request.HealthCheckInterval != 0 &&
			def.HealthCheckInterval != response.HealthCheckInterval {
			needUpdate = true
			config.HealthCheckInterval = def.HealthCheckInterval
		}
		if def.PersistenceTimeout != response.PersistenceTimeout {
			needUpdate = true
			config.PersistenceTimeout = def.PersistenceTimeout
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("UDP listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
			err := f.client.DeleteLoadBalancerListener(f.loadbalancer.LoadBalancerId, int(port.Port))
			if err != nil {
				return err
			}
			return f.client.CreateLoadBalancerUDPListener((*slb.CreateLoadBalancerUDPListenerArgs)(config))
		}

		if !needUpdate {
			glog.V(2).Infof("alicloud: udp listener did not change, skip [update], port=[%d], nodeport=[%d]\n", port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("UDP listener checker changed, request recreate [%s]\n", f.loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(request))
		glog.V(5).Infof(PrettyJson(response))
		return f.client.SetLoadBalancerUDPListenerAttribute(config)
	}
	doAdd := func(targetPort v1.ServicePort) error {
		return f.client.CreateLoadBalancerUDPListener(
			&slb.CreateLoadBalancerUDPListenerArgs{
				LoadBalancerId:    f.loadbalancer.LoadBalancerId,
				ListenerPort:      int(targetPort.Port),
				BackendServerPort: int(targetPort.NodePort),
				//Health Check
				Bandwidth:          def.Bandwidth,
				PersistenceTimeout: def.PersistenceTimeout,

				//HealthCheckType:           request.HealthCheckType,
				//HealthCheckURI:            request.HealthCheckURI,
				HealthCheckConnectPort:    def.HealthCheckConnectPort,
				HealthyThreshold:          def.HealthyThreshold,
				UnhealthyThreshold:        def.UnhealthyThreshold,
				HealthCheckConnectTimeout: def.HealthCheckConnectTimeout,
				HealthCheckInterval:       def.HealthCheckInterval,
				HealthCheck:               def.HealthCheck,
			},
		)
	}
	return &CommonListener{
		BaseListener{
			manager:  f,
			doUpdate: doUpdate,
			doAdd:    doAdd,
		},
	}
}

//####################################################################################
//# @date: 2018-03-16
//# @name: ListenerHTTP
//# @desc: LoadBalancer HTTP listener implementation.
//####################################################################################

func (f *ListenerManager) NewHTTP() ListenerInterface {
	def, request := ExtractAnnotationRequest(f.service)
	doUpdate := func(port v1.ServicePort) error {
		response, err := f.client.DescribeLoadBalancerHTTPListenerAttribute(f.loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.SetLoadBalancerHTTPListenerAttributeArgs{
			LoadBalancerId:    f.loadbalancer.LoadBalancerId,
			ListenerPort:      int(port.Port),
			BackendServerPort: int(port.NodePort),
			//Health Check
			Bandwidth:         response.Bandwidth,
			StickySession:     response.StickySession,
			StickySessionType: response.StickySessionType,
			CookieTimeout:     response.CookieTimeout,
			Cookie:            response.Cookie,

			HealthCheck:            response.HealthCheck,
			HealthCheckURI:         response.HealthCheckURI,
			HealthCheckConnectPort: response.HealthCheckConnectPort,
			HealthyThreshold:       response.HealthyThreshold,
			UnhealthyThreshold:     response.UnhealthyThreshold,
			HealthCheckTimeout:     response.HealthCheckTimeout,
			HealthCheckDomain:      response.HealthCheckDomain,
			HealthCheckHttpCode:    response.HealthCheckHttpCode,
			HealthCheckInterval:    response.HealthCheckInterval,
		}
		needUpdate := false
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("HTTP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}

		// todo: perform healthcheck update.
		if def.HealthCheck != response.HealthCheck {
			needUpdate = true
			config.HealthCheck = def.HealthCheck
		}
		if request.HealthCheckURI != "" &&
			def.HealthCheckURI != response.HealthCheckURI {
			needUpdate = true
			config.HealthCheckURI = def.HealthCheckURI
		}
		if request.HealthCheckConnectPort != 0 &&
			def.HealthCheckConnectPort != response.HealthCheckConnectPort {
			needUpdate = true
			config.HealthCheckConnectPort = def.HealthCheckConnectPort
		}
		if request.HealthyThreshold != 0 &&
			def.HealthyThreshold != response.HealthyThreshold {
			needUpdate = true
			config.HealthyThreshold = def.HealthyThreshold
		}
		if request.UnhealthyThreshold != 0 &&
			def.UnhealthyThreshold != response.UnhealthyThreshold {
			needUpdate = true
			config.UnhealthyThreshold = def.UnhealthyThreshold
		}
		if request.HealthCheckTimeout != 0 &&
			def.HealthCheckTimeout != response.HealthCheckTimeout {
			needUpdate = true
			config.HealthCheckTimeout = def.HealthCheckTimeout
		}
		if request.HealthCheckInterval != 0 &&
			def.HealthCheckInterval != response.HealthCheckInterval {
			needUpdate = true
			config.HealthCheckInterval = def.HealthCheckInterval
		}
		if string(request.StickySession) != "" &&
			def.StickySession != response.StickySession {
			needUpdate = true
			config.StickySession = def.StickySession
		}
		if string(request.StickySessionType) != "" &&
			def.StickySessionType != response.StickySessionType {
			needUpdate = true
			config.StickySessionType = def.StickySessionType
		}
		if request.Cookie != "" &&
			def.Cookie != response.Cookie {
			needUpdate = true
			config.Cookie = def.Cookie
		}
		if request.CookieTimeout != 0 &&
			def.CookieTimeout != response.CookieTimeout {
			needUpdate = true
			config.CookieTimeout = def.CookieTimeout
		}
		if request.HealthCheckHttpCode != "" &&
			def.HealthCheckHttpCode != response.HealthCheckHttpCode {
			needUpdate = true
			config.HealthCheckHttpCode = def.HealthCheckHttpCode
		}
		if request.HealthCheckDomain != "" &&
			def.HealthCheckDomain != response.HealthCheckDomain {
			needUpdate = true
			config.HealthCheckDomain = def.HealthCheckDomain
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("HTTP listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
			err := f.client.DeleteLoadBalancerListener(f.loadbalancer.LoadBalancerId, int(port.Port))
			if err != nil {
				return err
			}
			return f.client.CreateLoadBalancerHTTPListener((*slb.CreateLoadBalancerHTTPListenerArgs)(config))
		}

		if !needUpdate {
			glog.V(2).Infof("alicloud: http listener did not change, skip [update], port=[%d], nodeport=[%d]\n", port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("HTTP listener checker changed, request recreate [%s]\n", f.loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(request))
		glog.V(5).Infof(PrettyJson(response))
		return f.client.SetLoadBalancerHTTPListenerAttribute(config)
	}
	doAdd := func(targetPort v1.ServicePort) error {
		return f.client.CreateLoadBalancerHTTPListener(
			&slb.CreateLoadBalancerHTTPListenerArgs{
				LoadBalancerId:    f.loadbalancer.LoadBalancerId,
				ListenerPort:      int(targetPort.Port),
				BackendServerPort: int(targetPort.NodePort),
				//Health Check
				Bandwidth:         request.Bandwidth,
				StickySession:     def.StickySession,
				StickySessionType: def.StickySessionType,
				CookieTimeout:     def.CookieTimeout,
				Cookie:            def.Cookie,

				//HealthCheckType:           request.HealthCheckType,
				HealthCheckURI:         request.HealthCheckURI,
				HealthCheckConnectPort: request.HealthCheckConnectPort,
				HealthyThreshold:       request.HealthyThreshold,
				UnhealthyThreshold:     request.UnhealthyThreshold,
				//HealthCheckConnectTimeout: request.HealthCheckConnectTimeout,
				HealthCheckInterval: request.HealthCheckInterval,
				HealthCheckDomain:   def.HealthCheckDomain,
				HealthCheck:         def.HealthCheck,
				HealthCheckTimeout:  def.HealthCheckTimeout,
				HealthCheckHttpCode: def.HealthCheckHttpCode,
			},
		)
	}
	return &CommonListener{
		BaseListener{
			manager:  f,
			doUpdate: doUpdate,
			doAdd:    doAdd,
		},
	}
}

//####################################################################################
//# @date: 2018-03-16
//# @name: ListenerHTTPS
//# @desc: LoadBalancer HTTPS listener implementation.
//####################################################################################

func (f *ListenerManager) NewHTTPS() ListenerInterface {
	def, request := ExtractAnnotationRequest(f.service)
	doUpdate := func(port v1.ServicePort) error {
		response, err := f.client.DescribeLoadBalancerHTTPSListenerAttribute(f.loadbalancer.LoadBalancerId, int(port.Port))
		if err != nil {
			return err
		}
		config := &slb.SetLoadBalancerHTTPSListenerAttributeArgs{
			HTTPListenerType: slb.HTTPListenerType{
				LoadBalancerId:    f.loadbalancer.LoadBalancerId,
				ListenerPort:      response.ListenerPort,
				BackendServerPort: response.BackendServerPort,
				//Health Check
				HealthCheck:       response.HealthCheck,
				Bandwidth:         response.Bandwidth,
				StickySession:     response.StickySession,
				StickySessionType: response.StickySessionType,
				CookieTimeout:     response.CookieTimeout,
				Cookie:            response.Cookie,

				HealthCheckURI:         response.HealthCheckURI,
				HealthCheckConnectPort: response.HealthCheckConnectPort,
				HealthyThreshold:       response.HealthyThreshold,
				UnhealthyThreshold:     response.UnhealthyThreshold,
				HealthCheckTimeout:     response.HealthCheckTimeout,
				HealthCheckInterval:    response.HealthCheckInterval,
				HealthCheckHttpCode:    response.HealthCheckHttpCode,
				HealthCheckDomain:      response.HealthCheckDomain,
			},
			ServerCertificateId: response.ServerCertificateId,
		}

		needUpdate := false
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("HTTPS listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}

		// todo: perform healthcheck update.
		if def.HealthCheck != response.HealthCheck {
			needUpdate = true
			config.HealthCheck = def.HealthCheck
		}
		if request.HealthCheckURI != "" &&
			def.HealthCheckURI != response.HealthCheckURI {
			needUpdate = true
			config.HealthCheckURI = def.HealthCheckURI
		}
		if request.HealthCheckConnectPort != 0 &&
			def.HealthCheckConnectPort != response.HealthCheckConnectPort {
			needUpdate = true
			config.HealthCheckConnectPort = def.HealthCheckConnectPort
		}
		if request.HealthyThreshold != 0 &&
			def.HealthyThreshold != response.HealthyThreshold {
			needUpdate = true
			config.HealthyThreshold = def.HealthyThreshold
		}
		if request.UnhealthyThreshold != 0 &&
			def.UnhealthyThreshold != response.UnhealthyThreshold {
			needUpdate = true
			config.UnhealthyThreshold = def.UnhealthyThreshold
		}
		if request.HealthCheckTimeout != 0 &&
			def.HealthCheckTimeout != response.HealthCheckTimeout {
			needUpdate = true
			config.HealthCheckTimeout = def.HealthCheckTimeout
		}
		if request.HealthCheckInterval != 0 &&
			def.HealthCheckInterval != response.HealthCheckInterval {
			needUpdate = true
			config.HealthCheckInterval = def.HealthCheckInterval
		}

		if string(request.StickySession) != "" &&
			def.StickySession != response.StickySession {
			needUpdate = true
			config.StickySession = def.StickySession
		}
		if string(request.StickySessionType) != "" &&
			def.StickySessionType != response.StickySessionType {
			needUpdate = true
			config.StickySessionType = def.StickySessionType
		}
		if request.Cookie != "" &&
			def.Cookie != response.Cookie {
			needUpdate = true
			config.Cookie = def.Cookie
		}
		if request.CookieTimeout != 0 &&
			def.CookieTimeout != response.CookieTimeout {
			needUpdate = true
			config.CookieTimeout = def.CookieTimeout
		}
		if request.HealthCheckHttpCode != "" &&
			def.HealthCheckHttpCode != response.HealthCheckHttpCode {
			needUpdate = true
			config.HealthCheckHttpCode = def.HealthCheckHttpCode
		}
		if request.HealthCheckDomain != "" &&
			def.HealthCheckDomain != response.HealthCheckDomain {
			needUpdate = true
			config.HealthCheckDomain = def.HealthCheckDomain
		}
		// backend server port has changed.
		if int(port.NodePort) != response.BackendServerPort {
			needUpdate = true
			config.BackendServerPort = int(port.NodePort)
			glog.V(2).Infof("HTTPS listener checker [BackendServerPort] changed, request=%d. response=%d", port.NodePort, response.BackendServerPort)
			err := f.client.DeleteLoadBalancerListener(f.loadbalancer.LoadBalancerId, int(port.Port))
			if err != nil {
				return err
			}
			return f.client.CreateLoadBalancerHTTPSListener((*slb.CreateLoadBalancerHTTPSListenerArgs)(config))
		}

		if !needUpdate {
			glog.V(2).Infof("alicloud: https listener did not change, skip [update], port=[%d], nodeport=[%d]\n", port.Port, port.NodePort)
			// no recreate needed.  skip
			return nil
		}
		glog.V(2).Infof("HTTPS listener checker changed, request recreate [%s]\n", f.loadbalancer.LoadBalancerId)
		glog.V(5).Infof(PrettyJson(request))
		glog.V(5).Infof(PrettyJson(response))
		return f.client.SetLoadBalancerHTTPSListenerAttribute(config)
	}
	doAdd := func(targetPort v1.ServicePort) error {
		return f.client.CreateLoadBalancerHTTPSListener(
			&slb.CreateLoadBalancerHTTPSListenerArgs{
				HTTPListenerType: slb.HTTPListenerType{
					LoadBalancerId:    f.loadbalancer.LoadBalancerId,
					ListenerPort:      int(targetPort.Port),
					BackendServerPort: int(targetPort.NodePort),
					//Health Check
					HealthCheck:       def.HealthCheck,
					Bandwidth:         def.Bandwidth,
					StickySession:     def.StickySession,
					StickySessionType: def.StickySessionType,
					Cookie:            def.Cookie,
					CookieTimeout:     def.CookieTimeout,

					HealthCheckURI:         def.HealthCheckURI,
					HealthCheckConnectPort: def.HealthCheckConnectPort,
					HealthyThreshold:       def.HealthyThreshold,
					UnhealthyThreshold:     def.UnhealthyThreshold,
					HealthCheckTimeout:     def.HealthCheckTimeout,
					HealthCheckInterval:    def.HealthCheckInterval,
					HealthCheckDomain:      def.HealthCheckDomain,
					HealthCheckHttpCode:    def.HealthCheckHttpCode,
				},
				ServerCertificateId: request.CertID,
			},
		)
	}
	return &CommonListener{
		BaseListener{
			manager:  f,
			doUpdate: doUpdate,
			doAdd:    doAdd,
		},
	}
}

type ListenerManager struct {
	client       ClientSLBSDK
	service      *v1.Service
	loadbalancer *slb.LoadBalancerType
}

func NewListenerManager(client ClientSLBSDK,
	servcie *v1.Service, loadbalancer *slb.LoadBalancerType) *ListenerManager {
	return &ListenerManager{
		client:       client,
		service:      servcie,
		loadbalancer: loadbalancer,
	}
}

func (f *ListenerManager) Build(proto string) ListenerInterface {

	switch proto {
	case "tcp":
		return f.NewTCP()
	case "udp":
		return f.NewUDP()
	case "http":
		return f.NewHTTP()
	case "https":
		return f.NewHTTPS()
	default:
		glog.Warningf("alicloud: unknown protocol specified. fallback to tcp.")
		return f.NewTCP()
	}
}

func Protocol(annotation string, port v1.ServicePort) (string, error) {

	if annotation == "" {
		return strings.ToLower(string(port.Protocol)), nil
	}
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return "", errors.New(fmt.Sprintf("port and "+
				"protocol format must be like 'https:443' with colon separated. got=[%+v]", pp))
		}

		if pp[0] != "http" &&
			pp[0] != "tcp" &&
			pp[0] != "https" &&
			pp[0] != "udp" {
			return "", errors.New(fmt.Sprintf("port protocol"+
				" format must be either [http|https|tcp|udp], protocol not supported wit [%s]\n", pp[0]))
		}

		if pp[1] == fmt.Sprintf("%d", port.Port) {
			return pp[0], nil
		}
	}
	return strings.ToLower(string(port.Protocol)), nil
}

//ApplyUpdate try to update current loadbalancer`s listeners based on the given service
func (f *ListenerManager) Apply() error {
	if err := f.Deletions(); err != nil {
		return err
	}
	return f.Updations()
}

func (f *ListenerManager) Updations() error {
	anno := serviceAnnotation(f.service, ServiceAnnotationLoadBalancerProtocolPort)

	for _, spec := range f.service.Spec.Ports {
		found := false
		realProtocol, err := Protocol(anno, spec)
		if err != nil {
			return err
		}
		pprot := f.loadbalancer.ListenerPortsAndProtocol
		for _, back := range pprot.ListenerPortAndProtocol {
			if int(spec.Port) == back.ListenerPort &&
				strings.ToUpper(back.ListenerProtocol) == strings.ToUpper(realProtocol) {
				found = true
				break
			}
		}
		listener := f.Build(realProtocol)
		if !found {
			// Add listener
			glog.V(4).Infof("alicloud: attempt to [add] listener, port=[%d],"+
				" nodeport=[%d],[%s]\n", spec.Port, spec.NodePort, f.loadbalancer.LoadBalancerId)
			err := listener.Add(spec)
			if err != nil {
				return err
			}
		} else {
			glog.V(4).Infof("alicloud: attempt to [update] listener, port=[%d],"+
				" nodeport=[%d],[%s]\n", spec.Port, spec.NodePort, f.loadbalancer.LoadBalancerId)
			if err := listener.Update(spec); err != nil {
				return err
			}
		}
		// todo : here should retry
		if err := f.client.StartLoadBalancerListener(
			f.loadbalancer.LoadBalancerId, int(spec.Port)); err != nil {
			return err
		}
	}
	return nil
}
func (f *ListenerManager) Deletions() error {
	anno := serviceAnnotation(f.service, ServiceAnnotationLoadBalancerProtocolPort)
	pprot := f.loadbalancer.ListenerPortsAndProtocol
	for _, back := range pprot.ListenerPortAndProtocol {
		found := false
		for _, spec := range f.service.Spec.Ports {
			if int(spec.Port) == back.ListenerPort {
				realProtocol, err := Protocol(anno, spec)
				if err != nil {
					return err
				}
				if strings.ToUpper(back.ListenerProtocol) == strings.ToUpper(realProtocol) {
					found = true
					break
				}
			}
		}
		if !found {
			// Add listener
			glog.V(4).Infof("alicloud: attempt to [delete] "+
				"listener, port=[%d],[%s]\n", back.ListenerPort, f.loadbalancer.LoadBalancerId)

			if err := f.Build("tcp").Remove(
				v1.ServicePort{
					Port: int32(back.ListenerPort),
				}); err != nil {
				return err
			}
		}
	}
	return nil
}
