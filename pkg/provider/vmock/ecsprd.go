package vmock

import (
	"context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

/*
	Provider needs permission
	"alibaba:DeleteInstance", "alibaba:RunCommand",
*/
func NewMockECS(
	auth *base.ClientMgr,
) *MockECS {
	return &MockECS{auth: auth}
}

type MockECS struct {
	auth *base.ClientMgr
}

var _ prvd.IInstance = &MockECS{}

func (d *MockECS) ListInstances(ctx context.Context, ids []string) (map[string]*prvd.NodeAttribute, error) {
	return nil, nil
}

func (d *MockECS) SetInstanceTags(ctx context.Context, id string, tags map[string]string) error {
	return nil
}

func (d *MockECS) DescribeNetworkInterfaces(vpcId string, ips []string) (map[string]string, error) {
	return nil, nil
}
