package service

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	hashA := GetServiceHash(serviceA)

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
	hashB := GetServiceHash(serviceB)
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
	hashC := GetServiceHash(serviceC)
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
	hashD := GetServiceHash(serviceD)
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
	hashE := GetServiceHash(serviceE)
	if hashA != hashE {
		t.Logf("svc is same, but hash changed, from %s -> %s", hashA, hashE)
		t.Fail()
	}
}
