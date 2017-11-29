package alicloud

import (
	"testing"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"time"
)

func TestRoute(t *testing.T) {
	rsdk,err := NewSDKClientRoutes("","")
	if err !=nil {
		t.Fail()
	}

	route,err := rsdk.ListRoutes(common.Hangzhou, []string{"",""})
	if err != nil {
		t.Fail()
	}
	for _,r := range route{
		t.Log(fmt.Sprintf("%+v\n",r))
	}
}


func TestRouteExpire(t *testing.T) {
	rsdk,err := NewSDKClientRoutes("","")
	if err !=nil {
		t.Fail()
	}

	for i:=0;i<20;i++{
		_,err := rsdk.ListRoutes(common.Hangzhou, []string{""})
		if err != nil {
			t.Fail()
		}
		time.Sleep(time.Duration(3*time.Second))
		fmt.Printf("Log 1 second: %d",i)
	}

}



func TestSep(t *testing.T) {
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerProtocolPort))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerAddressType))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerSLBNetworkType))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerChargeType))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerRegion))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerBandwidth))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerCertID))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckFlag))

	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckType))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckURI))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckConnectPort))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckInterval))
	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckConnectTimeout))

	fmt.Println(replaceCamel(ServiceAnnotationLoadBalancerHealthCheckTimeout))
}


