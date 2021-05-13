package alibaba

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"reflect"
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

func TestProviderSLB_DescribeLoadBalancerListeners(t *testing.T) {
	client, err := slb.NewClientWithAccessKey("cn-hangzhou",
		"key", "secret")
	if err != nil {
		t.Fatalf(err.Error())
	}

	req := slb.CreateDescribeLoadBalancerTCPListenerAttributeRequest()
	req.LoadBalancerId = "lb-xxxx"
	req.ListenerPort = requests.NewInteger(80)
	resp, err := client.DescribeLoadBalancerTCPListenerAttribute(req)
	if err != nil {
		t.Fatalf("DescribeLoadBalancerTCPListenerAttribute error: %s", err.Error())
	}
	t.Logf("%v", resp)
}

func TestNewLBProvider(t *testing.T) {
	ln := "local"
	rn := "remote"
	local := &model.LoadBalancer{}
	local.LoadBalancerAttribute = model.LoadBalancerAttribute{LoadBalancerName: &ln}
	remote := &model.LoadBalancer{}
	remote.LoadBalancerAttribute = model.LoadBalancerAttribute{LoadBalancerName: &rn}

	changeName(local, remote)

	t.Logf("local name: %s,remote name: %s", *local.LoadBalancerAttribute.LoadBalancerName,
		*remote.LoadBalancerAttribute.LoadBalancerName)

}

func changeName(local *model.LoadBalancer, remote *model.LoadBalancer) {
	*local.LoadBalancerAttribute.LoadBalancerName = "local2"

	*remote.LoadBalancerAttribute.LoadBalancerName = "remote2"

}

type dog struct {
	LegCount int
}

func TestProviderSLB_CreateLoadBalancerHTTPListener(t *testing.T) {
	req := slb.CreateSetLoadBalancerTCPListenerAttributeRequest()
	port := &model.ListenerAttribute{}
	port.VGroupId = "12345"
	setV(req, port)
	t.Logf("vgroup id %s", req.VServerGroupId)

}

func setV(req interface{}, port *model.ListenerAttribute) {
	v := reflect.ValueOf(req).Elem()
	vgroup := v.FieldByName("VServerGroupId")
	vgroup.SetString(port.VGroupId)
}
