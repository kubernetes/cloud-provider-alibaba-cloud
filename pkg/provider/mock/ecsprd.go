package mock

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	log "github.com/sirupsen/logrus"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"time"
)

type Dummy struct{}

var _ provider.Instance = &Dummy{}

func (d *Dummy) DetailECS(ctx *node.NodeContext) (*provider.DetailECS, error) { return nil, nil }

func (d *Dummy) RestartECS(ctx *node.NodeContext) error { return nil }

func (d *Dummy) DestroyECS(ctx *node.NodeContext) error { return nil }

func (d *Dummy) RunCommand(ctx *node.NodeContext, cmd string) (*ecs.Invocation, error) {
	return &ecs.Invocation{}, nil
}

func (d *Dummy) ReplaceSystemDisk(ctx *node.NodeContext) error {
	log.Infof("replace system disk done")
	return nil
}

/*
	Provider needs permission
	"alibaba:DeleteInstance", "alibaba:RunCommand",
*/
func NewEcsProvider(
	auth *metadata.ClientAuth,
) *EcsProvider {
	return &EcsProvider{auth: auth}
}

var _ provider.Instance = &EcsProvider{}

type EcsProvider struct {
	auth *metadata.ClientAuth
}

func (e *EcsProvider) DetailECS(ctx *node.NodeContext) (*provider.DetailECS, error) {
	return nil, nil
}

func (e *EcsProvider) ReplaceSystemDisk(ctx *node.NodeContext) error {
	return nil
}

func (e *EcsProvider) DestroyECS(ctx *node.NodeContext) error {
	return nil
}
func (e *EcsProvider) RestartECS(ctx *node.NodeContext) error { return nil }
func (e *EcsProvider) RunCommand(
	ctx *node.NodeContext, command string,
) (*ecs.Invocation, error) {

	return nil, nil
}

const (
	// 240s
	StopECSTimeout  = 240
	StartECSTimeout = 300

	// Default timeout value for WaitForInstance method
	InstanceDefaultTimeout = 120
	DefaultWaitForInterval = 5
)

func Wait4Instance(
	client *ecs.Client,
	id, status string,
	timeout int,
) error {
	if timeout <= 0 {
		timeout = InstanceDefaultTimeout
	}
	req := ecs.CreateDescribeInstanceAttributeRequest()
	req.InstanceId = id
	for {
		instance, err := client.DescribeInstanceAttribute(req)
		if err != nil {
			return err
		}
		if instance.Status == status {
			//TODO
			//Sleep one more time for timing issues
			time.Sleep(DefaultWaitForInterval * time.Second)
			break
		}
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("timeout waiting %s %s", id, status)
		}
		time.Sleep(DefaultWaitForInterval * time.Second)

	}
	return nil
}
