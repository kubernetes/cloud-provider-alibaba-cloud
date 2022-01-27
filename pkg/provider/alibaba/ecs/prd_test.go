package ecs

import (
	"context"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func NewECSClient() (*ecs.Client, error) {
	var ak, sk, regionId string
	if ak == "" || sk == "" {
		return nil, fmt.Errorf("ak or sk is empty")
	}
	return ecs.NewClientWithAccessKey(regionId, ak, sk)
}

func TestEcsProvider_ListInstances(t *testing.T) {
	client, err := NewECSClient()
	if err != nil {
		t.Skip("fail to create ecs client, skip")
		return
	}
	ids := []string{
		"cn-hangzhou.i-xxxx",
	}

	ecsProvider := NewEcsProvider(&base.ClientMgr{
		Meta: nil,
		ECS:  client,
		VPC:  nil,
		SLB:  nil,
		PVTZ: nil,
	})

	cloudNodes, err := ecsProvider.ListInstances(context.TODO(), ids)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	for _, id := range ids {
		c, ok := cloudNodes[id]
		if !ok || c == nil {
			t.Errorf("cannot find %s ecs from cloud", id)
		}
		t.Logf("%s, instance address: %s", id, c.Addresses)
	}

	t.Logf("ListInstances test successfully")
}
