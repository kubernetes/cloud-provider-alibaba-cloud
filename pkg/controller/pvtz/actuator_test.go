package pvtz

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

const (
	IP1                  = "10.0.0.1"
	IP2                  = "10.0.0.2"
	Domain1              = "test.com"
	testServiceName      = "test-svc"
	testServiceNamespace = "default"
	testServiceRr        = "test-svc.default.svc"
)

var (
	testCommonObjectMeta = metav1.ObjectMeta{
		Name:      testServiceName,
		Namespace: testServiceNamespace,
	}
	testDesiredEndpoints = map[*corev1.Service]*prvd.PvtzEndpoint{
		// Multiple Ingress IP LoadBalancer
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: IP1}, {IP: IP2}},
				},
			},
		}: {
			Rr:     testServiceRr,
			Type:   prvd.RecordTypeA,
			Values: []prvd.PvtzValue{{Data: IP1}, {Data: IP2}},
		},
		// TODO Headless ClusterIP, vmock client
		// Normal ClusterIP
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:       corev1.ServiceTypeClusterIP,
				ClusterIP:  IP1,
				ClusterIPs: []string{IP1, IP2},
			},
		}: {
			Rr:     testServiceRr,
			Type:   prvd.RecordTypeA,
			Values: []prvd.PvtzValue{{Data: IP1}, {Data: IP2}},
		},
		// Normal NodePort
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:       corev1.ServiceTypeNodePort,
				ClusterIP:  IP1,
				ClusterIPs: []string{IP1, IP2},
			},
		}: {
			Rr:     testServiceRr,
			Type:   prvd.RecordTypeA,
			Values: []prvd.PvtzValue{{Data: IP1}, {Data: IP2}},
		},
		// IP ExternalName
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: IP1,
			},
		}: {
			Rr:     testServiceRr,
			Type:   prvd.RecordTypeA,
			Values: []prvd.PvtzValue{{Data: IP1}},
		},
		// Domain ExternalName
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: Domain1,
			},
		}: {
			Rr:     testServiceRr,
			Type:   prvd.RecordTypeCNAME,
			Values: []prvd.PvtzValue{{Data: Domain1}},
		},
	}
)

func TestDesiredEndpoints(t *testing.T) {
	a := NewActuator(nil, nil)
	for k, v := range testDesiredEndpoints {
		ret, err := a.desiredEndpoints(k)
		if err != nil {
			t.Errorf("getting desired endpoitns error %s", err)
		}
		t.Logf("result is %v, expected is %v", ret.ValueString(), v.ValueString())
		if !ret.ValueEqual(v) {
			t.Fail()
		}
	}
}
