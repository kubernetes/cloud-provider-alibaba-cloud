package annotation

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestGet(t *testing.T) {
	svc := getDefaultService()
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
	svc := getDefaultService()
	anno := NewAnnotationRequest(svc)
	svc.Annotations[Annotation(AdditionalTags)] = "Key1=Value1,Key2=Value2"
	tags := anno.GetLoadBalancerAdditionalTags()
	assert.Equal(t, len(tags), 2)
}

func TestIsForceOverride(t *testing.T) {
	svc := getDefaultService()
	anno := NewAnnotationRequest(svc)
	assert.Equal(t, anno.Get(OverrideListener), "")

	svc.Annotations[Annotation(OverrideListener)] = "true"
	assert.Equal(t, anno.Get(OverrideListener), "true")

}

func TestGetDefaultValue(t *testing.T) {
	svc := getDefaultService()
	anno := NewAnnotationRequest(svc)
	assert.Equal(t, anno.GetDefaultValue(AddressType), "internet")
	assert.Equal(t, anno.GetDefaultValue(Spec), "slb.s1.small")
	assert.Equal(t, anno.GetDefaultValue(IPVersion), "ipv4")
	assert.Equal(t, anno.GetDefaultValue(DeleteProtection), "on")
	assert.Equal(t, anno.GetDefaultValue(ModificationProtection), "ConsoleProtection")
}

func TestGetDefaultLoadBalancerName(t *testing.T) {
	svc := getDefaultService()
	svc.UID = "5e4dbfc9-c2ae-4642-b033-5607860aef6a"
	anno := NewAnnotationRequest(svc)
	assert.Equal(t, anno.GetDefaultLoadBalancerName(), "a5e4dbfc9c2ae4642b0335607860aef6")
}

func getDefaultService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					NodePort:   80,
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
}
