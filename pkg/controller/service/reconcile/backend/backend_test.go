package backend

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestBatch(t *testing.T) {
	sum := 0
	addFunc := func(a []int) error {
		for _, num := range a {
			sum += num
		}
		return nil
	}
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if err := Batch(nums, 3, addFunc); err != nil {
		t.Fatalf("Batch error: %s", err.Error())
	}
	assert.Equal(t, sum, 55)
}

func TestSetAddressIpVersion(t *testing.T) {
	tests := []struct {
		name                string
		backendIPVersion    string
		ipVersion           string
		isNLB               bool
		enableDualStack     bool
		enableEndpointSlice bool
		expected            model.AddressIPVersionType
	}{
		{
			name:                "default IPv4 when no backend-ip-version annotation",
			backendIPVersion:    "",
			ipVersion:           "",
			isNLB:               false,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
		{
			name:                "IPv4 when backend-ip-version is ipv4",
			backendIPVersion:    "ipv4",
			ipVersion:           "",
			isNLB:               false,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
		{
			name:                "IPv4 when IPv6DualStack feature gate disabled",
			backendIPVersion:    "ipv6",
			ipVersion:           "ipv6",
			isNLB:               false,
			enableDualStack:     false,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
		{
			name:                "IPv4 when EndpointSlice feature gate disabled",
			backendIPVersion:    "ipv6",
			ipVersion:           "ipv6",
			isNLB:               false,
			enableDualStack:     true,
			enableEndpointSlice: false,
			expected:            model.IPv4,
		},
		{
			name:                "IPv4 when ip-version annotation mismatch for CLB",
			backendIPVersion:    "ipv6",
			ipVersion:           "ipv4",
			isNLB:               false,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
		{
			name:                "IPv6 for CLB when all conditions met",
			backendIPVersion:    "ipv6",
			ipVersion:           "ipv6",
			isNLB:               false,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv6,
		},
		{
			name:                "IPv4 for NLB when ip-version is ipv6 and backend-ip-version is ipv6 (NLB only supports DualStack)",
			backendIPVersion:    "ipv6",
			ipVersion:           "ipv6",
			isNLB:               true,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
		{
			name:                "DualStack for NLB when all conditions met",
			backendIPVersion:    "dualstack",
			ipVersion:           "dualstack",
			isNLB:               true,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.DualStack,
		},
		{
			name:                "IPv4 for NLB when ip-version is ipv6 but backend-ip-version is dualstack",
			backendIPVersion:    "dualstack",
			ipVersion:           "ipv6",
			isNLB:               true,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
		{
			name:                "IPv4 for CLB when backend-ip-version is dualstack (CLB does not support dualstack)",
			backendIPVersion:    "dualstack",
			ipVersion:           "dualstack",
			isNLB:               false,
			enableDualStack:     true,
			enableEndpointSlice: true,
			expected:            model.IPv4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultMutableFeatureGate, ctrlCfg.IPv6DualStack, tt.enableDualStack)()
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultMutableFeatureGate, ctrlCfg.EndpointSlice, tt.enableEndpointSlice)()

			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
			}

			if tt.isNLB {
				svc.Spec.LoadBalancerClass = stringPtr(helper.NLBClass)
			}

			if tt.backendIPVersion != "" {
				svc.Annotations[annotation.Annotation(annotation.BackendIPVersion)] = tt.backendIPVersion
			}

			if tt.ipVersion != "" {
				svc.Annotations[annotation.Annotation(annotation.IPVersion)] = tt.ipVersion
			}

			reqCtx := &svcCtx.RequestContext{
				Service: svc,
				Anno:    annotation.NewAnnotationRequest(svc),
				Log:     logr.Discard(),
			}

			e := &EndpointWithENI{}
			e.setAddressIpVersion(reqCtx)

			assert.Equal(t, tt.expected, e.AddressIPVersion)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
