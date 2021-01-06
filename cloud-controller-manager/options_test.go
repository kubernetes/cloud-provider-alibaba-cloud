package alicloud

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestExtractAnnotationRequest(t *testing.T) {
	annotations := make(map[string]string)
	annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec"] = "slb.s1.small"
	annotations["service.beta.kubernetes.io/alicloud-loadbalancer-spec"] = "slb.s2.small"
	annotations["service.beta.kubernetes.io/alicloud-loadbalancer-address-type"] = "intranet"
	annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type"] = "internet"

	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
		},
	}
	for i := 0; i < 10; i++ {
		def, _ := ExtractAnnotationRequest(&svc)
		if def.LoadBalancerSpec != "slb.s1.small" {
			t.Fatal("old annotation works")
		}
		if def.AddressType != "internet" {
			t.Fatal("old annotation works")
		}
	}

}
