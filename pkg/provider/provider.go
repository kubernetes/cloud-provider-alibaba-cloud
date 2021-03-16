package provider

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
)

type DetailECS struct {
	// ImageID alibaba image id
	ImageID string
}

type Provider interface {
	Instance
	Route
	LoadBalancer
	PrivateZone
}

type Instance interface {
	DetailECS(ctx *node.NodeContext) (*DetailECS, error)
	// RestartECS
	// restart alibaba and wait for running
	RestartECS(ctx *node.NodeContext) error

	DestroyECS(ctx *node.NodeContext) error

	RunCommand(ctx *node.NodeContext, cmd string) (*ecs.Invocation, error)
	// ReplaceSystemDisk replace system disk and run user data
	ReplaceSystemDisk(ctx *node.NodeContext) error
}

type Route interface {
	CreateRoute()
	DeleteRoute()
	ListRoute()
}

type PrivateZone interface {
	CreatePVTZ()
}

type LoadBalancer interface {
	FindLoadBalancer()
	ListLoadBalancer()
	CreateLoadBalancer()
}
