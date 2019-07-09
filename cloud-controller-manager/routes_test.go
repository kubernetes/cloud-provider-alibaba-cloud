/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/denverdino/aliyungo/ecs"
)

func NewMockRouteMgr(tables string) (*ClientMgr, error) {
	mgr := &ClientMgr{
		routes: &RoutesClient{
			// mockRouteSDK can be override by implement its method
			client: &mockRouteSDK{},
			region: string(REGION),
		},
	}
	mgr.routes.WithVPC(VPCID, "")
	return mgr, nil
}

func TestListRoutes(t *testing.T) {

	cmgr, err := NewMockRouteMgr("")
	if err != nil {
		t.Fatal("failed to create client manager")
	}

	// ==========================================================================
	// init route cache.
	PreSetCloudData(
		WithNewRouteStore(),
		WithVpcs(),
		WithRouteTableEntrySet(),
		WithVRouter(),
	)

	route, err := cmgr.Routes().ListRoutes(ROUTE_TABLE_ID)
	if err != nil {
		t.Fatal(fmt.Sprintf("failed to list routes, %v", err))
	}
	for _, r := range route {
		found := false
		t.Log(PrettyJson(r))
		for _, entry := range ROUTE_ENTRIES {
			if entry.DestinationCidrBlock == r.DestinationCIDR &&
				string(r.TargetNode) == strings.Join([]string{string(REGION), entry.InstanceId}, ".") {
				if entry.Type != "Custom" {
					t.Fatal("should not return none cutomized routes")
				}
				found = true
			}
		}
		if !found {
			t.Fatal("route was not matched.")
		}
	}
}

func testCamel(t *testing.T, original, expected string) {
	converted := replaceCamel(normalizePrefix(original))
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

func RealClient(t *testing.T) {
	realRouteClient(keyid, keysecret)
}

func realRouteClient(keyid, keysec string) {
	if keyid == "" || keysec == "" {
		return
	}
	cs := ecs.NewClient(keyid, keysec)

	vpc, _, _ := cs.DescribeRouteTables(&ecs.DescribeRouteTablesArgs{
		RouteTableId: "vtb-2zedne8cr43rp5oqsr9xg",
		VRouterId:    "vrt-2zegcm0ty46mq243fmxoj",
	})

	fmt.Printf(PrettyJson(vpc))
}
