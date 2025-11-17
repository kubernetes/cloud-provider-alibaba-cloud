package nlb

import (
	"fmt"
	"testing"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	nlb "github.com/alibabacloud-go/nlb-20220430/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
)

func NewNLBClient() (*nlb.Client, error) {
	config := &openapi.Config{
		RegionId:        tea.String("cn-shanghai"),
		AccessKeyId:     tea.String(""),
		AccessKeySecret: tea.String(""),
	}
	if tea.StringValue(config.AccessKeyId) == "" || tea.StringValue(config.AccessKeySecret) == "" {
		return nil, fmt.Errorf("init nlb client error: ak not set")
	}

	client, err := nlb.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("init nlb client error: %s", err.Error())
	}
	return client, nil
}

func TestCreateLoadBalancer(t *testing.T) {

	client, err := NewNLBClient()
	if err != nil {
		t.Skip("fail to create slb client, skip")
	}

	request := &nlb.CreateLoadBalancerRequest{
		AddressType: tea.String("Internet"),
		VpcId:       tea.String("vpc-uf6ixxxx"),
		ZoneMappings: []*nlb.CreateLoadBalancerRequestZoneMappings{
			{
				VSwitchId: tea.String("vsw-uf6xxxx"),
				ZoneId:    tea.String("cn-shanghai-f"),
			},
			{
				VSwitchId: tea.String("vsw-uf6xxxx"),
				ZoneId:    tea.String("cn-shanghai-g"),
			},
		},
	}

	resp, err := client.CreateLoadBalancer(request)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("resp: %+v", resp)

}

func TestDeleteLoadBalancer(t *testing.T) {
	client, err := NewNLBClient()
	if err != nil {
		t.Skip("fail to create slb client, skip")
	}

	req := &nlb.DeleteLoadBalancerRequest{
		LoadBalancerId: tea.String("nlb-id"),
	}
	resp, err := client.DeleteLoadBalancer(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("resp: %+v", resp)
}

func TestListLoadBalancers(t *testing.T) {
	client, err := NewNLBClient()
	if err != nil {
		t.Skip("fail to create slb client, skip")
	}
	req := &nlb.ListLoadBalancersRequest{}
	req.Tag = append(req.Tag,
		&nlb.ListLoadBalancersRequestTag{
			Key:   tea.String("kubernetes.do.not.delete"),
			Value: tea.String("afb401xxxxxxxxxx"),
		},
	)
	resp, err := client.ListLoadBalancers(req)
	if err != nil {
		t.Error(err)
	}

	t.Logf("RequestId: %s, API: %s, %+v", *resp.Body.RequestId, "DescribeLoadBalancers", resp.Body.LoadBalancers)
}

func TestTags(t *testing.T) {
	client, err := NewNLBClient()
	if err != nil {
		t.Skip("fail to create slb client, skip")
	}
	req := &nlb.ListTagResourcesRequest{}
	req.ResourceType = tea.String("loadbalancer")
	req.ResourceId = []*string{tea.String("nlb-xxxx-id")}

	resp, err := client.ListTagResources(req)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", resp)

	tagReq := &nlb.TagResourcesRequest{}
	tagReq.ResourceType = tea.String("loadbalancer")
	tagReq.ResourceId = []*string{tea.String("nlb-xxx-id")}

	tagReq.Tag = []*nlb.TagResourcesRequestTag{
		{
			Key:   tea.String("kubernetes.do.not.delete"),
			Value: tea.String("a99d2f1xxxxxxx"),
		},
	}
	_, err = client.TagResources(tagReq)
	if err != nil {
		t.Error(err)
	}
}

func TestGetLoadBalancer(t *testing.T) {
	client, err := NewNLBClient()
	if err != nil {
		t.Skip("fail to create slb client, skip")
	}
	var (
		retErr error
		resp   *nlb.GetLoadBalancerAttributeResponse
	)

	lbId := "nlb-xxx"
	_ = wait.PollImmediate(20*time.Second, 1*time.Minute, func() (bool, error) {
		req := &nlb.GetLoadBalancerAttributeRequest{}
		req.LoadBalancerId = tea.String(lbId)

		resp, retErr = client.GetLoadBalancerAttribute(req)
		if retErr != nil {
			return false, util.SDKError("GetLoadBalancerAttribute", retErr)
		}

		if resp == nil || resp.Body == nil {
			retErr = fmt.Errorf("nlbId %s GetLoadBalancerAttribute response is nil, resp [%+v]",
				lbId, resp)
			return false, retErr
		}

		if tea.StringValue(resp.Body.LoadBalancerStatus) == string(Provisioning) {
			retErr = fmt.Errorf("nlb %s is in creating status", lbId)
			return false, nil
		}

		retErr = nil
		return true, nil
	})
}

func TestRetryOnThrottling(t *testing.T) {
	count := 0
	err := retryOnThrottling("TestThrottling", func() error {
		count++
		return nil
	})
	assert.Equal(t, 1, count)
	assert.Nil(t, err)

	count = 0
	tErr := &tea.SDKError{
		Code:    tea.String("Throttling.User"),
		Message: tea.String("Request was denied due to user flow control."),
	}
	err = retryOnThrottling("TestThrottling", func() error {
		count++
		return tErr
	})
	assert.NotEqual(t, 0, count)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, tErr)

	count = 0
	err = retryOnThrottling("TestThrottling", func() error {
		count++
		if count > 1 {
			return nil
		}
		return &tea.SDKError{
			Code:    tea.String("Throttling.User"),
			Message: tea.String("Request was denied due to user flow control."),
		}
	})
	assert.Equal(t, 2, count)
	assert.Nil(t, err)

	count = 0
	nErr := &tea.SDKError{
		Code:    tea.String("TestError"),
		Message: tea.String("test error message"),
	}
	err = retryOnThrottling("TestThrottling", func() error {
		count++
		return nErr
	})
	assert.Equal(t, 1, count)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, nErr)
}
