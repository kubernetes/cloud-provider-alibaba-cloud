package alicloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/denverdino/aliyungo/common"
)

func TestRoute(t *testing.T) {
	rsdk, err := NewSDKClientRoutes("", "")
	if err != nil {
		t.Fail()
	}

	route, err := rsdk.ListRoutes(common.Hangzhou, []string{"", ""})
	if err != nil {
		t.Fail()
	}
	for _, r := range route {
		t.Log(fmt.Sprintf("%+v\n", r))
	}
}

func TestRouteExpire(t *testing.T) {
	rsdk, err := NewSDKClientRoutes("", "")
	if err != nil {
		t.Fail()
	}

	for i := 0; i < 20; i++ {
		_, err := rsdk.ListRoutes(common.Hangzhou, []string{""})
		if err != nil {
			t.Fail()
		}
		time.Sleep(time.Duration(3 * time.Second))
		fmt.Printf("Log 1 second: %d", i)
	}

}

func testCamel(t *testing.T, original, expected string) {
	converted := replaceCamel(original)
	if converted != expected {
		t.Errorf("failed to replace camel from %s to %s: %s", original, expected, converted)
	}
}

func TestSep(t *testing.T) {

	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-ProtocolPort", ServiceAnnotationLoadBalancerProtocolPort)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-AddressType", ServiceAnnotationLoadBalancerAddressType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-SLBNetworkType", ServiceAnnotationLoadBalancerSLBNetworkType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-ChargeType", ServiceAnnotationLoadBalancerChargeType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-Region", ServiceAnnotationLoadBalancerRegion)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-Bandwidth", ServiceAnnotationLoadBalancerBandwidth)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-CertID", ServiceAnnotationLoadBalancerCertID)

	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckFlag", ServiceAnnotationLoadBalancerHealthCheckFlag)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckType", ServiceAnnotationLoadBalancerHealthCheckType)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckURI", ServiceAnnotationLoadBalancerHealthCheckURI)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckConnectPort", ServiceAnnotationLoadBalancerHealthCheckConnectPort)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthyThreshold", ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-UnhealthyThreshold", ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckInterval", ServiceAnnotationLoadBalancerHealthCheckInterval)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckConnectTimeout", ServiceAnnotationLoadBalancerHealthCheckConnectTimeout)
	testCamel(t, "service.beta.kubernetes.io/alicloud-loadbalancer-HealthCheckTimeout", ServiceAnnotationLoadBalancerHealthCheckTimeout)

	// Ignore the unsupported annotation
	testCamel(t, "alicloud-loadbalancer-HealthCheckTimeout", "alicloud-loadbalancer-HealthCheckTimeout")
}
