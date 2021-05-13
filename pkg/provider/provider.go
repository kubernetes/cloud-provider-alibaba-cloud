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
	// DescribeNetworkInterfaces query one or more elastic network interfaces (ENIs)
	DescribeNetworkInterfaces(vpcId string, ips *[]string) (*ecs.DescribeNetworkInterfacesResponse, error)
}

type Route interface {
	CreateRoute()
	DeleteRoute()
	ListRoute()
}

type ILoadBalancer interface {
	FindSLB(ctx context.Context, slb *model.LoadBalancer) (bool, error)
	CreateSLB(ctx context.Context, slb *model.LoadBalancer) error
	DescribeSLB(ctx context.Context, slb *model.LoadBalancer) error
	DeleteSLB(ctx context.Context, slb *model.LoadBalancer) error
	// Listener
	DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error)
	StartLoadBalancerListener(ctx context.Context, lbId string, port int) error
	StopLoadBalancerListener(ctx context.Context, lbId string, port int) error
	DeleteLoadBalancerListener(ctx context.Context, lbId string, port int) error
	CreateLoadBalancerTCPListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	SetLoadBalancerTCPListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	CreateLoadBalancerUDPListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	SetLoadBalancerUDPListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	CreateLoadBalancerHTTPListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	SetLoadBalancerHTTPListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	CreateLoadBalancerHTTPSListener(ctx context.Context, lbId string, port *model.ListenerAttribute) error
	SetLoadBalancerHTTPSListenerAttribute(ctx context.Context, lbId string, port *model.ListenerAttribute) error

	// VServerGroup
	CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error
	DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error)
	DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (*model.VServerGroup, error)

	//DeleteVServerGroup(ctx context.Context, vGroupId string) (err error)
	//SetVServerGroupAttribute(ctx context.Context, args *slb.SetVServerGroupAttributeArgs) (response *slb.SetVServerGroupAttributeResponse, err error)
	//ModifyVServerGroupBackendServers(ctx context.Context, args *slb.ModifyVServerGroupBackendServersArgs) (response *slb.ModifyVServerGroupBackendServersResponse, err error)
}
