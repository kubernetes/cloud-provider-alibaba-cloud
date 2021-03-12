package provider

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/context"
)

type DetailECS struct {
	// ImageID ecs image id
	ImageID string
}

type Interface interface {
	DetailECS(ctx *context.NodeContext) (*DetailECS, error)
	// RestartECS
	// restart ecs and wait for running
	RestartECS(ctx *context.NodeContext) error

	DestroyECS(ctx *context.NodeContext) error

	RunCommand(ctx *context.NodeContext, cmd string) (*ecs.Invocation, error)
	// ReplaceSystemDisk replace system disk and run user data
	ReplaceSystemDisk(ctx *context.NodeContext) error
}
