package service

import (
	alicloud "k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"testing"
)

func TestImplemntion(t *testing.T) {
	var cloud cloudprovider.Interface
	cloud = &alicloud.Cloud{}
	lb, implemented := cloud.LoadBalancer()
	if !implemented {
		t.Fatalf("alibaba cloud does not implement LoadBalancer")
	}
	_, ok := lb.(EnsureENI)
	if !ok {
		t.Fatalf("alibaba cloud does not implement EnsureENI")
	}
}
