package ecs

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ess"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/oos"
	log "github.com/sirupsen/logrus"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

type Dummy struct{}

var _ provider.Interface = &Dummy{}

func (d *Dummy) DetailECS(ctx *context.NodeContext) (*provider.DetailECS, error) { return nil, nil }

func (d *Dummy) RestartECS(ctx *context.NodeContext) error { return nil }

func (d *Dummy) DestroyECS(ctx *context.NodeContext) error { return nil }

func (d *Dummy) RunCommand(ctx *context.NodeContext, cmd string) (*ecs.Invocation, error) {
	return &ecs.Invocation{}, nil
}

func (d *Dummy) ReplaceSystemDisk(ctx *context.NodeContext) error {
	log.Infof("replace system disk done")
	return nil
}

/*
	Provider needs permission
	"ecs:DeleteInstance", "ecs:RunCommand",
*/

func NewProvider(
	ecs *ecs.Client,
	oos *oos.Client,
	ess *ess.Client,
) *Executor {
	return NewExecutor(ecs, oos, ess)
}

func NewExecutor(
	ecs *ecs.Client,
	oos *oos.Client,
	ess *ess.Client,
) *Executor {
	return &Executor{ECS: ecs, OOS: oos, ESS: ess}
}

var _ provider.Interface = &Executor{}

type Executor struct {
	ECS *ecs.Client
	OOS *oos.Client
	ESS *ess.Client
}

func (e *Executor) DetailECS(ctx *context.NodeContext) (*provider.DetailECS, error) {
	return nil,nil
}

func (e *Executor) ReplaceSystemDisk(ctx *context.NodeContext) error {
	return nil
}

func (e *Executor) DestroyECS(ctx *context.NodeContext) error {
	return nil
}
func (e *Executor) RestartECS(ctx *context.NodeContext) error { return nil }
func (e *Executor) RunCommand(
	ctx *context.NodeContext, command string,
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
