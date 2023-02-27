package nlb

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"strconv"
	"strings"
)

const (
	InternetAddressType = "Internet"
	IntranetAddressType = "Intranet"
)

func GetAddressType(addressType string) string {
	if strings.EqualFold(addressType, InternetAddressType) {
		return InternetAddressType
	}
	if strings.EqualFold(addressType, IntranetAddressType) {
		return IntranetAddressType
	}
	return addressType
}

const (
	IPv4      = "ipv4"
	DualStack = "DualStack"
)

func GetAddressIpVersion(addressIpVersion string) string {
	if strings.EqualFold(addressIpVersion, IPv4) {
		return IPv4
	}
	if strings.EqualFold(addressIpVersion, DualStack) {
		return DualStack
	}
	return addressIpVersion
}

const (
	TCP    = "TCP"
	UDP    = "UDP"
	TCPSSL = "TCPSSL"
)

func GetListenerProtocolType(protocol string) string {
	if strings.EqualFold(protocol, TCP) {
		return TCP
	}
	if strings.EqualFold(protocol, UDP) {
		return UDP
	}
	if strings.EqualFold(protocol, TCPSSL) {
		return TCPSSL
	}
	return protocol
}

type ServerGroupType string

const (
	InstanceServerGroupType = ServerGroupType("Instance")
	IpServerGroupType       = ServerGroupType("Ip")
)

type ServerType string

const (
	EcsServerType = ServerType("Ecs")
	EniServerType = ServerType("Eni")
	IpServerType  = ServerType("Ip")
)

type ListenerStatus string

const (
	StoppedListenerStatus = ListenerStatus("Stopped")
)

type TagResourceType string

const (
	LoadBalancerTagType = TagResourceType("loadbalancer")
	ServerGroupTagType  = TagResourceType("servergroup")
)

// NetworkLoadBalancer represents a AlibabaCloud NetworkLoadBalancer.
type NetworkLoadBalancer struct {
	NamespacedName        types.NamespacedName
	LoadBalancerAttribute *LoadBalancerAttribute
	Listeners             []*ListenerAttribute
	ServerGroups          []*ServerGroup
}

func (l *NetworkLoadBalancer) GetLoadBalancerId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.LoadBalancerId
}

type LoadBalancerAttribute struct {
	IsUserManaged bool

	Name             string
	AddressType      string
	AddressIpVersion string
	VpcId            string
	ZoneMappings     []ZoneMapping
	ResourceGroupId  string
	Tags             []tag.Tag
	SecurityGroupIds []string

	// auto-generated parameters
	LoadBalancerId             string
	LoadBalancerStatus         string
	LoadBalancerBusinessStatus string
	DNSName                    string
}

type ListenerAttribute struct {
	IsUserManaged   bool
	NamedKey        *ListenerNamedKey
	ServerGroupName string
	ServicePort     *v1.ServicePort

	ListenerProtocol     string
	ListenerPort         int32
	ListenerDescription  string
	ServerGroupId        string
	LoadBalancerId       string
	IdleTimeout          int32 // 1-900
	SecurityPolicyId     string
	CertificateIds       []string // tcpssl
	CaCertificateIds     []string
	CaEnabled            *bool
	ProxyProtocolEnabled *bool
	SecSensorEnabled     *bool
	AlpnEnabled          *bool
	AlpnPolicy           string
	StartPort            *int32 //0-65535
	EndPort              *int32 //0-65535
	Cps                  *int32 //0-1000000

	// auto-generated parameters
	ListenerId string
	ListenerStatus
}

type ServerGroup struct {
	IsUserManaged bool
	NamedKey      *SGNamedKey
	ServicePort   *v1.ServicePort
	Weight        *int

	VPCId                   string
	ServerGroupName         string
	ServerGroupType         ServerGroupType
	ResourceGroupId         string
	AddressIPVersion        string
	Protocol                string
	ConnectionDrainEnabled  *bool
	ConnectionDrainTimeout  int32 // 10-900
	Scheduler               string
	PreserveClientIpEnabled *bool
	HealthCheckConfig       *HealthCheckConfig
	Servers                 []ServerGroupServer
	Tags                    []tag.Tag

	// auto-generated parameters
	ServerGroupId string
}

func (s *ServerGroup) BackendInfo() string {
	backendJson, err := json.Marshal(s.Servers)
	if err != nil {
		return fmt.Sprintf("%v", s.Servers)
	}
	return string(backendJson)
}

type ServerGroupServer struct {
	IsUserManaged bool
	NodeName      *string

	ServerGroupId string
	Description   string
	ServerId      string
	ServerIp      string
	ServerType    ServerType
	Port          int32
	Weight        int32
	ZoneId        string
	Status        string
}

type ZoneMapping struct {
	VSwitchId    string
	ZoneId       string
	IPv4Addr     string
	AllocationId string
}

type HealthCheckConfig struct {
	HealthCheckEnabled        *bool
	HealthCheckType           string
	HealthCheckConnectPort    int32
	HealthyThreshold          int32
	UnhealthyThreshold        int32
	HealthCheckConnectTimeout int32
	HealthCheckInterval       int32
	HealthCheckDomain         string
	HealthCheckUrl            string
	HealthCheckHttpCode       []string
	HttpCheckMethod           string
}

type NamedKey struct {
	Prefix      string
	CID         string
	Namespace   string
	ServiceName string
}

func (n *NamedKey) IsManagedByService(svc *v1.Service, clusterId string) bool {
	return n != nil && n.ServiceName == svc.Name &&
		n.Namespace == svc.Namespace &&
		n.CID == clusterId
}

type ListenerNamedKey struct {
	NamedKey
	Port     int32
	Protocol string
}

func (n *ListenerNamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

func (n *ListenerNamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = model.DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s.%d.%s.%s.%s.%s", n.Prefix, n.Port, n.Protocol, n.ServiceName, n.Namespace, n.CID)
}

func LoadNLBListenerNamedKey(key string) (*ListenerNamedKey, error) {
	metas := strings.Split(key, ".")
	if len(metas) != 6 || metas[0] != model.DEFAULT_PREFIX {
		return nil, fmt.Errorf("ListenerName Format Error: k8s.${port}.${protocol}.${service}.${namespace}.${clusterid} format is expected. Got [%s]", key)
	}
	port, err := strconv.Atoi(metas[1])
	if err != nil {
		return nil, err
	}
	return &ListenerNamedKey{
		NamedKey: NamedKey{
			CID:         metas[5],
			Namespace:   metas[4],
			ServiceName: metas[3],
			Prefix:      model.DEFAULT_PREFIX,
		},
		Port:     int32(port),
		Protocol: metas[2],
	}, nil
}

type SGNamedKey struct {
	NamedKey
	Protocol string
	// SGGroupPort the port in the vgroup name
	SGGroupPort string
}

func (n *SGNamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

func (n *SGNamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = model.DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s.%s.%s.%s.%s.%s", n.Prefix, n.SGGroupPort, n.Protocol, n.ServiceName, n.Namespace, n.CID)
}

func LoadNLBSGNamedKey(key string) (*SGNamedKey, error) {
	metas := strings.Split(key, ".")
	if len(metas) != 6 || metas[0] != model.DEFAULT_PREFIX {
		return nil, fmt.Errorf("ServerGroupName Format Error: k8s.${port}.${protocol}.${service}.${namespace}.${clusterid} format is expected. Got [%s]", key)
	}

	return &SGNamedKey{
		NamedKey: NamedKey{
			CID:         metas[5],
			Namespace:   metas[4],
			ServiceName: metas[3],
			Prefix:      model.DEFAULT_PREFIX,
		},
		Protocol:    metas[2],
		SGGroupPort: metas[1],
	}, nil
}
