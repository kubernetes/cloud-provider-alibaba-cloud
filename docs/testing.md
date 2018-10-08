# Testing

### unit test

Alibaba Cloud Controller use mocked cloud SDK to implement code unit test.

```
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
			t.Logf("PPPP: %v\n", li)
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
```

Faked SDK made unit test easier.

Use ```make test``` to run unit test.

### Integration Test

Alibaba Cloud Controller Manager integration test is expected to follow the kubernetes testgrid rule addressed in https://github.com/kubernetes/community/pull/2224#issuecomment-395410751 for consistency.
