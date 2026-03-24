package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
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

func TestHas(t *testing.T) {
	svc := getDefaultService()
	anno := NewAnnotationRequest(svc)

	assert.False(t, anno.Has(AddressType))

	svc.Annotations[Annotation(AddressType)] = "intranet"
	assert.True(t, anno.Has(AddressType))

	svc = getDefaultService()
	anno = NewAnnotationRequest(svc)
	svc.Annotations["service.beta.kubernetes.io/alicloud-loadbalancer-address-type"] = "internet"
	assert.True(t, anno.Has(AddressType))

	svc = getDefaultService()
	anno = NewAnnotationRequest(svc)
	svc.Annotations[Annotation(AddressType)] = "Intranet"
	svc.Annotations["service.beta.kubernetes.io/alicloud-loadbalancer-address-type"] = "internet"
	assert.True(t, anno.Has(AddressType))
}

func TestGetDefaultTags(t *testing.T) {
	svc := getDefaultService()
	svc.UID = "5e4dbfc9-c2ae-4642-b033-5607860aef6a"
	anno := NewAnnotationRequest(svc)
	tags := anno.GetDefaultTags()

	assert.Equal(t, 2, len(tags))

	assert.Equal(t, "kubernetes.do.not.delete", tags[0].Key)
	assert.Equal(t, "a5e4dbfc9c2ae4642b0335607860aef6", tags[0].Value)

	assert.Equal(t, "ack.aliyun.com", tags[1].Key)
}

func TestGetLoadBalancerAdditionalTagsEdgeCases(t *testing.T) {
	svc := getDefaultService()
	anno := NewAnnotationRequest(svc)

	svc.Annotations[Annotation(AdditionalTags)] = ""
	tags := anno.GetLoadBalancerAdditionalTags()
	assert.Equal(t, 0, len(tags))

	svc.Annotations[Annotation(AdditionalTags)] = "Key1=Value1,Key2=,Key3"
	tags = anno.GetLoadBalancerAdditionalTags()
	assert.Equal(t, 3, len(tags))

	var key1Tag, key2Tag, key3Tag *tag.Tag
	for _, tag := range tags {
		if tag.Key == "Key1" {
			key1Tag = &tag
		} else if tag.Key == "Key2" {
			key2Tag = &tag
		} else if tag.Key == "Key3" {
			key3Tag = &tag
		}
	}

	assert.NotNil(t, key1Tag)
	assert.Equal(t, "Value1", key1Tag.Value)

	assert.NotNil(t, key2Tag)
	assert.Equal(t, "", key2Tag.Value)

	assert.NotNil(t, key3Tag)
	assert.Equal(t, "", key3Tag.Value)

	svc.Annotations[Annotation(AdditionalTags)] = " Key1 = Value1 , Key2 =Value2,Key3= Value3 "
	tags = anno.GetLoadBalancerAdditionalTags()
	assert.Equal(t, 3, len(tags))
}

func TestGetWithNilServiceOrAnnotations(t *testing.T) {
	anno := &AnnotationRequest{Service: nil}
	assert.Equal(t, "", anno.Get(AddressType))

	svc := getDefaultService()
	svc.Annotations = nil
	anno = NewAnnotationRequest(svc)
	assert.Equal(t, "", anno.Get(AddressType))
}

func TestHasWithNilServiceOrAnnotations(t *testing.T) {
	anno := &AnnotationRequest{Service: nil}
	assert.False(t, anno.Has(AddressType))

	svc := getDefaultService()
	svc.Annotations = nil
	anno = NewAnnotationRequest(svc)
	assert.False(t, anno.Has(AddressType))
}

func TestGetDefaultValueForLoadBalancerName(t *testing.T) {
	svc := getDefaultService()
	svc.UID = "5e4dbfc9-c2ae-4642-b033-5607860aef6a"
	anno := NewAnnotationRequest(svc)

	name := anno.GetDefaultValue(LoadBalancerName)
	expectedName := "a5e4dbfc9c2ae4642b0335607860aef6"
	assert.Equal(t, expectedName, name)
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

func TestAnnotationRequest_IsForceOverride(t *testing.T) {
	tests := []struct {
		name        string
		svc         *v1.Service
		expected    bool
		description string
	}{
		{
			name: "annotation value is true",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "true",
					},
				},
			},
			expected:    true,
			description: "should return true when annotation value is 'true'",
		},
		{
			name: "annotation value is false",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "false",
					},
				},
			},
			expected:    false,
			description: "should return false when annotation value is 'false'",
		},
		{
			name: "annotation value is empty string",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "",
					},
				},
			},
			expected:    false,
			description: "should return false when annotation value is empty",
		},
		{
			name: "annotation does not exist",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"some-other-annotation": "value",
					},
				},
			},
			expected:    false,
			description: "should return false when annotation does not exist",
		},
		{
			name: "annotations map is nil",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			expected:    false,
			description: "should return false when annotations map is nil",
		},
		{
			name:        "service is nil",
			svc:         nil,
			expected:    false,
			description: "should return false when service is nil",
		},
		{
			name: "legacy annotation prefix with true value",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners": "true",
					},
				},
			},
			expected:    true,
			description: "should return true when legacy annotation is 'true'",
		},
		{
			name: "annotation value is 'True' (uppercase)",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "True",
					},
				},
			},
			expected:    false,
			description: "should return false when annotation value is 'True' (not lowercase 'true')",
		},
		{
			name: "annotation value is '1'",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "1",
					},
				},
			},
			expected:    false,
			description: "should return false when annotation value is '1' instead of 'true'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewAnnotationRequest(tt.svc)
			result := req.IsForceOverride()
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}
