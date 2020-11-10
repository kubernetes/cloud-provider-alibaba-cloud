package service

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"testing"
)

func TestGetServiceHash(t *testing.T) {
	ServiceAnnotationLoadBalancerAddressType := "service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type"
	serviceA := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-service",
			Namespace: "default",
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAddressType: "intranet",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector: map[string]string{
				"run": "nginx",
			},
		},
	}
	hashA, err := utils.GetServiceHash(serviceA)
	if err != nil {
		t.Logf("get service hash error")
		t.Fail()
	}

	// change svc annotation
	serviceB := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-service",
			Namespace: "default",
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAddressType: "internet",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector: map[string]string{
				"run": "nginx",
			},
		},
	}
	hashB, err := utils.GetServiceHash(serviceB)
	if err != nil {
		t.Logf("get service hash error")
		t.Fail()
	}
	if hashA == hashB {
		t.Logf("svc annotation changed, but hash svc is equal")
		t.Fail()
	}

	// change svc spec
	serviceC := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-service",
			Namespace: "default",
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAddressType: "intranet",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector: map[string]string{
				"run": "nginx",
			},
		},
	}
	hashC, err := utils.GetServiceHash(serviceC)
	if err != nil {
		t.Logf("get service hash error")
		t.Fail()
	}
	if hashA == hashC {
		t.Logf("svc annotation changed, but hash svc is equal")
		t.Fail()
	}

	// add empty attr
	serviceD := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-service",
			Namespace: "default",
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAddressType: "intranet",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector: map[string]string{
				"run": "nginx",
			},
			ExternalIPs: []string{},
		},
	}
	hashD, err := utils.GetServiceHash(serviceD)
	if err != nil {
		t.Logf("get service hash error")
		t.Fail()
	}
	if hashA != hashD {
		t.Logf("svc add empty attr, but hash changed, from %s -> %s", hashA, hashD)
		t.Fail()
	}

	// same svc
	serviceE := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "basic-service",
			Namespace: "default",
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerAddressType: "intranet",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
			Selector: map[string]string{
				"run": "nginx",
			},
		},
	}
	hashE, err := utils.GetServiceHash(serviceE)
	if err != nil {
		t.Logf("get service hash error")
		t.Fail()
	}
	if hashA != hashE {
		t.Logf("svc is same, but hash changed, from %s -> %s", hashA, hashE)
		t.Fail()
	}
}
