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
	IPv61                = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	IPv62                = "2001:0db8:85a3:::8a2e:0370:7334"
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
	testDesiredAandAAAAEndpoints = map[*corev1.Service][]*prvd.PvtzEndpoint{
		// Multiple Ingress IP LoadBalancer
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: IP1}, {IP: IP2}, {IP: IPv61}, {IP: IPv62}},
				},
			},
		}: {
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeA,
				Values: []prvd.PvtzValue{{Data: IP1}, {Data: IP2}},
			},
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeAAAA,
				Values: []prvd.PvtzValue{{Data: IPv61}, {Data: IPv62}},
			},
		},
		// TODO Headless ClusterIP, vmock client
		// Normal ClusterIP
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:       corev1.ServiceTypeClusterIP,
				ClusterIP:  IP1,
				ClusterIPs: []string{IP1, IPv61},
			},
		}: {
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeA,
				Values: []prvd.PvtzValue{{Data: IP1}},
			},
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeAAAA,
				Values: []prvd.PvtzValue{{Data: IPv61}},
			},
		},
		// Normal NodePort
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:       corev1.ServiceTypeNodePort,
				ClusterIP:  IP1,
				ClusterIPs: []string{IP1, IPv61},
			},
		}: {
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeA,
				Values: []prvd.PvtzValue{{Data: IP1}},
			},
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeAAAA,
				Values: []prvd.PvtzValue{{Data: IPv61}},
			},
		},
		// IPv4 ExternalName
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: IP1,
			},
		}: {
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeA,
				Values: []prvd.PvtzValue{{Data: IP1}},
			},
		},
		// IPv6 ExternalName
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: IPv61,
			},
		}: {
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeAAAA,
				Values: []prvd.PvtzValue{{Data: IPv61}},
			},
		},
	}

	testDesiredCNAMEEndpoints = map[*corev1.Service][]*prvd.PvtzEndpoint{
		// Domain ExternalName
		&corev1.Service{
			ObjectMeta: testCommonObjectMeta,
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: Domain1,
			},
		}: {
			{
				Rr:     testServiceRr,
				Type:   prvd.RecordTypeCNAME,
				Values: []prvd.PvtzValue{{Data: Domain1}},
			},
		},
	}
)

func TestDesiredAandAAAAEndpoints(t *testing.T) {
	a := NewActuator(nil, nil)
	testDesiredEndpoints(t, testDesiredAandAAAAEndpoints, a.desiredAandAAAA)
}

func TestDesiredCNAMEEndpoints(t *testing.T) {
	a := NewActuator(nil, nil)
	testDesiredEndpoints(t, testDesiredCNAMEEndpoints, a.desiredCNAME)
}

func testDesiredEndpoints(t *testing.T, cases map[*corev1.Service][]*prvd.PvtzEndpoint, desiredFunc func(*corev1.Service) ([]*prvd.PvtzEndpoint, error)) {
	i := 0
	for k, v := range cases {
		eps, err := desiredFunc(k)
		if err != nil {
			t.Errorf("testcase %d getting desired endpoitns error %s", i, err)
		}
		if len(eps) != len(v) {
			t.Logf("testcase %d len of result and expected result is not same, len(result) is %d, len(expected) is %d", i, len(eps), len(v))
			t.Fail()
		}
		for _, ep := range eps {
			found := false
			for _, expectEp := range v {
				if ep.ValueEqual(expectEp) {
					t.Logf("testcase %d found ep in expected, ep is %v, expected is %v", i, ep.ValueString(), expectEp.ValueString())
					found = true
					break
				}
			}
			if !found {
				t.Logf("testcase %d can not find ep in expected, ep is %v, test case is %v", i, ep.ValueString(), k)
				t.Fail()
			}
		}
		i++
	}
}
