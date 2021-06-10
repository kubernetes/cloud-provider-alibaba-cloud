package vmock

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

/*
	Provider needs permission
	"alibaba:DeleteInstance", "alibaba:RunCommand",
*/
func NewMockECS(
	auth *base.ClientAuth,
) *MockECS {
	return &MockECS{auth: auth}
}

type MockECS struct {
	auth *base.ClientAuth
}

var _ prvd.IInstance = &MockECS{}

func (d *MockECS) ListInstances(ctx *node.NodeContext, ids []string) (map[string]*prvd.NodeAttribute, error) {
	return nil, nil
}

func (d *MockECS) SetInstanceTags(ctx *node.NodeContext, id string, tags map[string]string) error {
	return nil
}

func (d *MockECS) DescribeNetworkInterfaces(vpcId string, ips *[]string, nextToken string) (*ecs.DescribeNetworkInterfacesResponse, error) {
	return nil, nil
}
