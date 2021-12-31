package alb

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type ServiceManager struct {
	ClusterID string

	Namespace string
	Name      string

	PortToServerGroup map[int32]*ServerGroupWithIngress

	TrafficPolicy                   string
	ContainsPotentialReadyEndpoints bool
}

type ServerGroupNamedKey struct {
	Prefix      string
	ClusterID   string
	Namespace   string
	IngressName string
	ServiceName string
	ServicePort int
}

type BackendItem struct {
	Pod         *v1.Pod
	Description string
	ServerId    string
	ServerIp    string
	Weight      int
	Port        int
	Type        string
}

type ServiceGroupWithNameKey struct {
	NamedKey *ServerGroupNamedKey
	Backends []BackendItem
}

type ServerGroupWithIngress struct {
	IngressNames []string
	Backends     []BackendItem
}

const (
	ECSBackendType = "ecs"
	ENIBackendType = "eni"
)

const (
	OnFlag  = FlagType("on")
	OffFlag = FlagType("off")
)

type FlagType string

func (n *ServerGroupNamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

const DefaultPrefix = "k8s"

func (n *ServerGroupNamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = DefaultPrefix
	}
	return fmt.Sprintf("%s_%d_%s_%s_%s_%.6s", n.Prefix, n.ServicePort, n.ServiceName, n.IngressName, n.Namespace, n.ClusterID)
}

type ServiceStackContext struct {
	ClusterID string

	ServiceNamespace string
	ServiceName      string

	Service *v1.Service

	ServicePortToIngressNames map[int32][]string

	IsServiceNotFound bool
}
