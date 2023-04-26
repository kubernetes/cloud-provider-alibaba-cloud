package elb

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

var DEFAULT_PREFIX = "k8s"

type NamedKey struct {
	Prefix      string
	CID         string
	Namespace   string
	ServiceName string
}

func (n *NamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%s/%s/%s", n.Prefix, n.ServiceName, n.Namespace, n.CID)
}

func (n *NamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

func LoadNamedKey(key string) (*NamedKey, error) {
	metas := strings.Split(key, "/")
	if len(metas) != 4 || metas[0] != DEFAULT_PREFIX {
		return nil, fmt.Errorf("NamedKey Format Error: k8s.${port}.${protocol}.${service}.${namespace}.${clusterid} format is expected. Got [%s]", key)
	}
	return &NamedKey{
		CID:         metas[3],
		Namespace:   metas[2],
		ServiceName: metas[1],
		Prefix:      metas[0],
	}, nil
}

// EdgeLoadBalancer represents a AlibabaCloud ENS LoadBalancer.
type EdgeLoadBalancer struct {
	NamespacedName        types.NamespacedName
	LoadBalancerAttribute EdgeLoadBalancerAttribute
	EipAttribute          EdgeEipAttribute
	ServerGroup           EdgeServerGroup
	Listeners             EdgeListeners
}

func (l *EdgeLoadBalancer) GetLoadBalancerId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.LoadBalancerId
}

func (l *EdgeLoadBalancer) GetLoadBalancerName() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.LoadBalancerName
}

func (l *EdgeLoadBalancer) GetNetworkId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.NetworkId
}

func (l *EdgeLoadBalancer) GetVSwitchId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.VSwitchId
}

func (l *EdgeLoadBalancer) GetAssociatedEipId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.AssociatedEipId
}

func (l *EdgeLoadBalancer) GetAssociatedEipAddress() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.AssociatedEipAddress
}

func (l *EdgeLoadBalancer) CanReUse() bool {
	if l == nil {
		return false
	}
	return l.LoadBalancerAttribute.IsReUsed
}

func (l *EdgeLoadBalancer) GetEipAddress() string {
	if l == nil {
		return ""
	}
	return l.EipAttribute.IpAddress
}
func (l *EdgeLoadBalancer) GetEipId() string {
	if l == nil {
		return ""
	}
	return l.EipAttribute.AllocationId
}
func (l *EdgeLoadBalancer) GetEipName() string {
	if l == nil {
		return ""
	}
	return l.EipAttribute.Name
}
