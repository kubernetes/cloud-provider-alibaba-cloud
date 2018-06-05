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
	"fmt"
	"k8s.io/api/core/v1"
	"testing"

	"github.com/denverdino/aliyungo/slb"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func loadbalancerAttrib(loadbalancer *slb.LoadBalancerType) *slb.LoadBalancerType {

	lports := struct {
		ListenerPort []int
	}{
		ListenerPort: []int{442, 80},
	}
	pproto := struct {
		ListenerPortAndProtocol []slb.ListenerPortAndProtocolType
	}{
		ListenerPortAndProtocol: []slb.ListenerPortAndProtocolType{
			{
				ListenerPort:     442,
				ListenerProtocol: "tcp",
			},
			{
				ListenerPort:     80,
				ListenerProtocol: "tcp",
			},
		},
	}
	backend := struct {
		BackendServer []slb.BackendServerType
	}{
		BackendServer: []slb.BackendServerType{
			{
				ServerId: "i-bp152coo41mv2dqry64j",
				Weight:   100,
			},
			{
				ServerId: "i-bp152coo41mv2dqry64i",
				Weight:   100,
			},
		},
	}
	loadbalancer.ListenerPorts = lports
	loadbalancer.ListenerPortsAndProtocol = pproto
	loadbalancer.BackendServers = backend
	return loadbalancer
}

func TestUpdateListenerPorts(t *testing.T) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "service-test",
			UID:         "abcdefghigklmnopqrstu",
			Annotations: map[string]string{
			//ServiceAnnotationLoadBalancerId: LOADBALANCER_ID,
			},
		},
		Spec: v1.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []v1.ServicePort{
				{
					Name:       "tcp",
					Protocol:   protocolTcp,
					Port:       80,
					TargetPort: intstr.FromInt(80),
					NodePort:   30480,
				}, {
					Name:       "tcp",
					Protocol:   protocolTcp,
					Port:       443,
					TargetPort: intstr.FromInt(443),
					NodePort:   30443,
				},
			},
		},
	}

	base := newBaseLoadbalancer()
	detail := loadbalancerAttrib(&base[0])
	mgr, _ := NewMockClientMgr(&mockClientSLB{
		startLoadBalancerListener: func(loadBalancerId string, port int) (err error) {

			return nil
		},
		stopLoadBalancerListener: func(loadBalancerId string, port int) (err error) {
			return nil
		},
		createLoadBalancerTCPListener: func(args *slb.CreateLoadBalancerTCPListenerArgs) (err error) {
			li := slb.ListenerPortAndProtocolType{
				ListenerPort:     args.ListenerPort,
				ListenerProtocol: "tcp",
			}
			t.Log("PPPP: %v\n",li)
			detail.ListenerPorts.ListenerPort = append(detail.ListenerPorts.ListenerPort, args.ListenerPort)
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = append(detail.ListenerPortsAndProtocol.ListenerPortAndProtocol, li)
			return nil
		},
		deleteLoadBalancerListener: func(loadBalancerId string, port int) (err error) {
			response := []slb.ListenerPortAndProtocolType{}
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol
			for _, p := range ports {
				if p.ListenerPort == port {
					continue
				}
				response = append(response, p)
			}

			listports := []int{}
			lports := detail.ListenerPorts.ListenerPort
			for _, po := range lports {
				if po == port {
					continue
				}
				listports = append(listports, po)
			}
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = response
			detail.ListenerPorts.ListenerPort = listports
			return nil
		},
		describeLoadBalancerTCPListenerAttribute: func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol

			for _, p := range ports {
				if p.ListenerPort == port {
					return &slb.DescribeLoadBalancerTCPListenerAttributeResponse{
						DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{
							Status: slb.Running,
						},
						TCPListenerType: slb.TCPListenerType{
							LoadBalancerId:    loadBalancerId,
							ListenerPort:      port,
							BackendServerPort: 31789,
							Bandwidth:         50,
						},
					}, nil
				}
			}
			return nil, errors.New("not found")
		},
	})

	err := NewListenerManager(mgr.loadbalancer.c, service, detail).Apply()

	if err != nil {
		t.Fatal("listener update error! ")
	}
	t.Log(PrettyJson(service))
	t.Log(PrettyJson(detail))

	for _, sport := range service.Spec.Ports {
		found := false
		for _, port := range detail.ListenerPortsAndProtocol.ListenerPortAndProtocol {
			if int(sport.Port) == port.ListenerPort {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("1. listen port protocol not found [%d]\n", sport.Port))
		}
		found = false
		for _, port := range detail.ListenerPorts.ListenerPort {
			if int(sport.Port) == port {
				found = true
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("2. listen port not found [%d]\n", sport.Port))
		}
	}

	for _, sport := range detail.ListenerPortsAndProtocol.ListenerPortAndProtocol {
		found := false
		for _, port := range service.Spec.Ports {
			if int(port.Port) == sport.ListenerPort {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("3. port not found [%d]\n", sport.ListenerPort))
		}
	}

	for _, sport := range detail.ListenerPorts.ListenerPort {
		found := false
		for _, port := range service.Spec.Ports {
			if int(port.Port) == sport {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("4. port not found [%d]\n", sport))
		}
	}

}

func TestUpdateListenerBackendPorts(t *testing.T) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "service-test",
			UID:         "abcdefghigklmnopqrstu",
			Annotations: map[string]string{
			//ServiceAnnotationLoadBalancerId: LOADBALANCER_ID,
			},
		},
		Spec: v1.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []v1.ServicePort{
				{
					Name:       "tcp",
					Protocol:   protocolTcp,
					Port:       80,
					TargetPort: intstr.FromInt(80),
					NodePort:   30480,
				}, {
					Name:       "tcp",
					Protocol:   protocolTcp,
					Port:       442,
					TargetPort: intstr.FromInt(443),
					NodePort:   30443,
				},
			},
		},
	}

	base := newBaseLoadbalancer()
	detail := loadbalancerAttrib(&base[0])
	mgr, _ := NewMockClientMgr(&mockClientSLB{
		startLoadBalancerListener: func(loadBalancerId string, port int) (err error) {

			return nil
		},
		stopLoadBalancerListener: func(loadBalancerId string, port int) (err error) {
			return nil
		},
		createLoadBalancerTCPListener: func(args *slb.CreateLoadBalancerTCPListenerArgs) (err error) {
			li := slb.ListenerPortAndProtocolType{
				ListenerPort:     args.ListenerPort,
				ListenerProtocol: "tcp",
			}
			detail.ListenerPorts.ListenerPort = append(detail.ListenerPorts.ListenerPort, args.ListenerPort)
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = append(detail.ListenerPortsAndProtocol.ListenerPortAndProtocol, li)
			return nil
		},
		deleteLoadBalancerListener: func(loadBalancerId string, port int) (err error) {
			response := []slb.ListenerPortAndProtocolType{}
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol
			for _, p := range ports {
				if p.ListenerPort == port {
					continue
				}
				response = append(response, p)
			}

			listports := []int{}
			lports := detail.ListenerPorts.ListenerPort
			for _, po := range lports {
				if po == port {
					continue
				}
				listports = append(listports, po)
			}
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = response
			detail.ListenerPorts.ListenerPort = listports
			return nil
		},
		describeLoadBalancerTCPListenerAttribute: func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol

			for _, p := range ports {
				if p.ListenerPort == port {
					return &slb.DescribeLoadBalancerTCPListenerAttributeResponse{
						DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{
							Status: slb.Running,
						},
						TCPListenerType: slb.TCPListenerType{
							LoadBalancerId:    loadBalancerId,
							ListenerPort:      port,
							BackendServerPort: 31789,
							Bandwidth:         50,
						},
					}, nil
				}
			}
			return nil, errors.New("not found")
		},
	})

	err := NewListenerManager(mgr.loadbalancer.c, service, detail).Apply()

	if err != nil {
		t.Fatal("listener update error! ")
	}
	t.Log(PrettyJson(service))
	t.Log(PrettyJson(detail))

	for _, sport := range service.Spec.Ports {
		found := false
		for _, port := range detail.ListenerPortsAndProtocol.ListenerPortAndProtocol {
			if int(sport.Port) == port.ListenerPort {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("1. listen port protocol not found [%d]\n", sport.Port))
		}
		found = false
		for _, port := range detail.ListenerPorts.ListenerPort {
			if int(sport.Port) == port {
				found = true
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("2. listen port not found [%d]\n", sport.Port))
		}
	}

	for _, sport := range detail.ListenerPortsAndProtocol.ListenerPortAndProtocol {
		found := false
		for _, port := range service.Spec.Ports {
			if int(port.Port) == sport.ListenerPort {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("3. port not found [%d]\n", sport.ListenerPort))
		}
	}

	for _, sport := range detail.ListenerPorts.ListenerPort {
		found := false
		for _, port := range service.Spec.Ports {
			if int(port.Port) == sport {
				found = true
				break
			}
		}
		if !found {
			t.Fatal(fmt.Sprintf("4. port not found [%d]\n", sport))
		}
	}

}
