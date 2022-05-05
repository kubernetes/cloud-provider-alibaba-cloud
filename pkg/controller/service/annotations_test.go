package service

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestGet(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
	anno := NewAnnotationRequest(svc)
	svc.Annotations[Annotation(AddressType)] = "Intranet"
	assert.Equal(t, anno.Get(AddressType), "Intranet")

	svc.Annotations["service.beta.kubernetes.io/alicloud-loadbalancer-name"] = "slb-test"
	assert.Equal(t, anno.Get(LoadBalancerName), "slb-test")

	svc.Annotations[Annotation(OverrideListener)] = "false"
	svc.Annotations["service.beta.kubernetes.io/alicloud-force-override-listeners"] = "true"
	assert.Equal(t, anno.Get(OverrideListener), "false")
}

func TestGetLoadBalancerAdditionalTags(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
	}
	anno := NewAnnotationRequest(svc)
	svc.Annotations[Annotation(AdditionalTags)] = "Key1=Value1,Key2=Value2"
	tags := anno.GetLoadBalancerAdditionalTags()
	assert.Equal(t, len(tags), 2)
}

func TestIsForceOverride(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
	}
	anno := NewAnnotationRequest(svc)
	assert.Equal(t, anno.Get(OverrideListener), "")

	svc.Annotations[Annotation(OverrideListener)] = "true"
	assert.Equal(t, anno.Get(OverrideListener), "true")

}

func TestGetDefaultValue(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
	}
	anno := NewAnnotationRequest(svc)
	assert.Equal(t, anno.GetDefaultValue(AddressType), "internet")
	assert.Equal(t, anno.GetDefaultValue(Spec), "slb.s1.small")
	assert.Equal(t, anno.GetDefaultValue(IPVersion), "ipv4")
	assert.Equal(t, anno.GetDefaultValue(DeleteProtection), "on")
	assert.Equal(t, anno.GetDefaultValue(ModificationProtection), "ConsoleProtection")
}

func TestGetDefaultLoadBalancerName(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
			UID:         "34fecb0d-69ef-4f7f-961e-81b6999dc630",
		},
	}
	anno := NewAnnotationRequest(svc)
	assert.Equal(t, anno.GetDefaultLoadBalancerName(), "a34fecb0d69ef4f7f961e81b6999dc63")
}
