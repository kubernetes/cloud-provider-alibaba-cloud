package vmock

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	log "github.com/sirupsen/logrus"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba"
)

/*
	Provider needs permission
	"alibaba:DeleteInstance", "alibaba:RunCommand",
*/
func NewMockECS(
	auth *alibaba.ClientAuth,
) *MockECS {
	return &MockECS{auth: auth}
}

type MockECS struct {
	auth *alibaba.ClientAuth
}

var _ prvd.IInstance = &MockECS{}

func (d *MockECS) ListInstances(ctx *node.NodeContext, ids []string) (map[string]*prvd.NodeAttribute, error) {
	return nil, nil
}
func (d *MockECS) SetInstanceTags(ctx *node.NodeContext, id string, tags map[string]string) error {
	return nil
}

func (d *MockECS) DetailECS(ctx *node.NodeContext) (*prvd.DetailECS, error) { return nil, nil }

func (d *MockECS) RestartECS(ctx *node.NodeContext) error { return nil }

func (d *MockECS) DestroyECS(ctx *node.NodeContext) error { return nil }

func (d *MockECS) RunCommand(ctx *node.NodeContext, cmd string) (*ecs.Invocation, error) {
	return &ecs.Invocation{}, nil
}

func (d *MockECS) ReplaceSystemDisk(ctx *node.NodeContext) error {
	log.Infof("replace system disk done")
	return nil
}

func (d *MockECS) DescribeNetworkInterfaces(vpcId string, ips *[]string, nextToken string) (*ecs.DescribeNetworkInterfacesResponse, error) {
	return nil, nil
}
