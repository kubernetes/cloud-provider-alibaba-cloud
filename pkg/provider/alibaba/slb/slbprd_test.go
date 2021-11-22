package slb

import (
	"encoding/json"
	"testing"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
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
	client, err := slb.NewClientWithAccessKey("cn-hangzhou", "key", "secert")
	if err != nil {
		t.Fatalf(err.Error())
	}

	lbIds := []string{
		"lb-xxx",
		"lb-xxxx",
	}
	for _, id := range lbIds {
		req := slb.CreateSetLoadBalancerDeleteProtectionRequest()
		req.LoadBalancerId = id
		req.DeleteProtection = "off"
		_, err := client.SetLoadBalancerDeleteProtection(req)
		if err != nil {
			t.Fatalf(err.Error())
		}

		request := slb.CreateDeleteLoadBalancerRequest()
		request.LoadBalancerId = id
		if _, err = client.DeleteLoadBalancer(request); err != nil {
			t.Fatalf(err.Error())
		}
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

func TestProviderSLB_DescribeVServerGroups(t *testing.T) {
	key := ""
	secert := ""
	client, err := slb.NewClientWithAccessKey("cn-hangzhou", key, secert)
	if err != nil {
		t.Fatalf(err.Error())
	}

	req := slb.CreateDescribeVServerGroupsRequest()
	req.LoadBalancerId = "lb-xxxx"
	resp, err := client.DescribeVServerGroups(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	var vgs []model.VServerGroup
	for _, v := range resp.VServerGroups.VServerGroup {

		req := slb.CreateDescribeVServerGroupAttributeRequest()
		req.VServerGroupId = v.VServerGroupId
		resp, err := client.DescribeVServerGroupAttribute(req)
		if err != nil {
			t.Fatalf(err.Error())
		}
		vg := setVServerGroupFromResponse(resp)

		namedKey, err := model.LoadVGroupNamedKey(vg.VGroupName)
		if err != nil {
			t.Fatalf(err.Error())
		}
		vg.NamedKey = namedKey
		vgs = append(vgs, vg)
	}
	jsonStr, _ := json.Marshal(vgs)
	t.Logf(string(jsonStr))
}

func TestSLBProvider_AddTags(t *testing.T) {
	key := ""
	secert := ""
	client, err := slb.NewClientWithAccessKey("cn-hangzhou", key, secert)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tags := []model.Tag{{TagKey: "testkey", TagValue: "testvalue"}}
	tagBytes, _ := json.Marshal(tags)
	req := slb.CreateAddTagsRequest()
	req.LoadBalancerId = "lb-xxxxx"
	req.Tags = string(tagBytes)
	_, err = client.AddTags(req)
	if err != nil {
		panic(err)
	}
}
