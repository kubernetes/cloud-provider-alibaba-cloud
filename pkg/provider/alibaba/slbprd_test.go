package alibaba

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"testing"
)

func TestNewLBProvider(t *testing.T) {
	client, err := slb.NewClientWithAccessKey(
		"cn-hangzhou",
		"key",
		"secret")
	if err != nil {
		t.Fatalf(err.Error())

	}

	request := slb.CreateCreateLoadBalancerRequest()
	request.LoadBalancerSpec = "slb.s1.small"
	request.AddressType = "Internet"
	request.LoadBalancerName = "test"

	resp, err := client.CreateLoadBalancer(request)
	request.ClientToken = utils.GetUUID()
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("success(%d)! loadBalancerId = %s\n", resp.GetHttpStatus(), resp.LoadBalancerId)
}
