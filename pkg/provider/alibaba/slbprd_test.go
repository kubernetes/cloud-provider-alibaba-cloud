package alibaba

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/klog"
	"testing"
)

func TestNewLBProvider_CreateSLB(t *testing.T) {
	client, err := slb.NewClientWithAccessKey("cn-hangzhou", "key", "secret")
	if err != nil {
		t.Fatalf(err.Error())
	}

	request := slb.CreateCreateLoadBalancerRequest()
	request.LoadBalancerSpec = "slb.s1.small"
	request.AddressType = "Internet"
	request.LoadBalancerName = "test"
	request.ClientToken = utils.GetUUID()

	resp, err := client.CreateLoadBalancer(request)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("success(%d)! loadBalancerId = %s\n", resp.GetHttpStatus(), resp.LoadBalancerId)
}

func TestProviderSLB_DeleteSLB(t *testing.T) {
	client, err := slb.NewClientWithAccessKey("cn-hangzhou", "key", "secret")
	if err != nil {
		t.Fatalf(err.Error())

	}

	request := slb.CreateDeleteLoadBalancerRequest()
	request.LoadBalancerId = "lb-xxx"
	if _, err = client.DeleteLoadBalancer(request); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewLBProvider(t *testing.T) {
	var p *string
	p = nil
	klog.Info(*p)
}
