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
	"github.com/denverdino/aliyungo/slb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestUpdateListenerPorts(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "With TCP Listener")

	f.SVC.Spec.Ports = []v1.ServicePort{
		{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
		{Port: 443, TargetPort: intstr.FromInt(6443), Protocol: v1.ProtocolTCP, NodePort: 31443},
	}

	f.RunDefault(t, "Add Listener 443")
}

func TestUpdateListenerBackendPorts(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: 31000},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "With TCP Listener")

	f.SVC.Spec.Ports = []v1.ServicePort{
		{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: 32000},
	}

	f.RunDefault(t, "Change NodePort from 31000 to 32000")
}

func TestStartLoadBalancerListener(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: 31000},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalaner
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "With TCP Listener")

	_, lb, _ := f.LoadBalancer().findLoadBalancer(f.SVC)
	slbClient := f.SLBSDK()
	res, err := slbClient.DescribeLoadBalancerTCPListenerAttribute(lb.LoadBalancerId, int(listenPort1))
	if err != nil {
		t.Fatalf("DescribeLoadBalancerTCPListenerAttribute error: %s", err.Error())
	}
	if res.Status != slb.Running {
		t.Fatalf("listener start error.")
	}
}

func TestStopLoadBalancerListener(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: 31000},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalaner
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
	)

	f.RunDefault(t, "With TCP Listener")

	slbClient := f.SLBSDK()
	err := slbClient.StopLoadBalancerListener(LOADBALANCER_ID, int(listenPort1))
	if err != nil {
		t.Fatalf("StopLoadBalancerListener error: %s", err.Error())
	}

	res, err := slbClient.DescribeLoadBalancerTCPListenerAttribute(LOADBALANCER_ID, int(listenPort1))
	if err != nil {
		t.Fatalf("DescribeLoadBalancerTCPListenerAttribute error: %s", err.Error())
	}

	if res.Status != slb.Stopped {
		t.Fatalf("listener stop error.")
	}
}
