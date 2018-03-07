package alicloud

import (
	"fmt"
	"testing"
	//"time"

	"github.com/denverdino/aliyungo/common"
)

func TestRoute(t *testing.T) {
	cmgr, err := NewClientMgr("", "")
	if err != nil {
		t.Log("failed to create client manager")
		t.Fail()
	}

	route, err := cmgr.routes.ListRoutes(common.Hangzhou, []string{"", ""})
	if err != nil {
		t.Log("failed to list routes")
		t.Fail()
	}
	for _, r := range route {
		t.Log(fmt.Sprintf("%+v\n", r))
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

	testCamel(t, ServiceAnnotationLoadBalancerProtocolPort, ServiceAnnotationLoadBalancerProtocolPort)
	testCamel(t, ServiceAnnotationLoadBalancerAddressType, ServiceAnnotationLoadBalancerAddressType)
	testCamel(t, ServiceAnnotationLoadBalancerSLBNetworkType, ServiceAnnotationLoadBalancerSLBNetworkType)
	testCamel(t, ServiceAnnotationLoadBalancerChargeType, ServiceAnnotationLoadBalancerChargeType)
	testCamel(t, ServiceAnnotationLoadBalancerRegion, ServiceAnnotationLoadBalancerRegion)
	testCamel(t, ServiceAnnotationLoadBalancerBandwidth, ServiceAnnotationLoadBalancerBandwidth)
	testCamel(t, ServiceAnnotationLoadBalancerCertID, ServiceAnnotationLoadBalancerCertID)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckFlag, ServiceAnnotationLoadBalancerHealthCheckFlag)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckType, ServiceAnnotationLoadBalancerHealthCheckType)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckURI, ServiceAnnotationLoadBalancerHealthCheckURI)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckConnectPort, ServiceAnnotationLoadBalancerHealthCheckConnectPort)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold, ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold, ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckInterval, ServiceAnnotationLoadBalancerHealthCheckInterval)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckConnectTimeout, ServiceAnnotationLoadBalancerHealthCheckConnectTimeout)
	testCamel(t, ServiceAnnotationLoadBalancerHealthCheckTimeout, ServiceAnnotationLoadBalancerHealthCheckTimeout)

	// Ignore the unsupported annotation
	testCamel(t, "alicloud-loadbalancer-HealthCheckTimeout", "alicloud-loadbalancer-HealthCheckTimeout")
}
