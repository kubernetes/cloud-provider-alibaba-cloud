package prvd

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/node"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
)

type DetailECS struct {
	// ImageID alibaba image id
	ImageID string
}

type Provider interface {
	Instance
	Route
	ILoadBalancer
	PrivateZone
}

// NodeAttribute node attribute from cloud instance
type NodeAttribute struct {
	InstanceID   string
	Addresses    []v1.NodeAddress
	InstanceType string
	Zone         string
	Region       string
}

type Instance interface {
	ListInstances(ctx *node.NodeContext, ids []string) (map[string]*NodeAttribute, error)
	SetInstanceTags(ctx *node.NodeContext, id string, tags map[string]string) error
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

type ILoadBalancer interface {
	FindSLB(ctx context.Context, id string) ([]model.LoadBalancer, error)
	ListSLB(ctx context.Context, slb model.LoadBalancer) ([]model.LoadBalancer, error)
	CreateSLB(ctx context.Context, slb *model.LoadBalancer) (*model.LoadBalancer, error)
	DeleteSLB(ctx context.Context, slb model.LoadBalancer) error
}
