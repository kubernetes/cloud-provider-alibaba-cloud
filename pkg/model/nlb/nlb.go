package nlb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
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

type ModificationProtectionType string

const ConsoleProtection = ModificationProtectionType("ConsoleProtection")
const NonProtection = ModificationProtectionType("NonProtection")
const ModificationProtectionReason = "managed.by.ack"

// NetworkLoadBalancer represents a AlibabaCloud NetworkLoadBalancer.
type NetworkLoadBalancer struct {
	NamespacedName                  types.NamespacedName
	LoadBalancerAttribute           *LoadBalancerAttribute
	Listeners                       []*ListenerAttribute
	ServerGroups                    []*ServerGroup
	ContainsPotentialReadyEndpoints bool
}

func (l *NetworkLoadBalancer) GetLoadBalancerId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.LoadBalancerId
}

type LoadBalancerAttribute struct {
	IsUserManaged bool

	Name                         string
	AddressType                  string
	AddressIpVersion             string
	IPv6AddressType              string
	VpcId                        string
	ZoneMappings                 []ZoneMapping
	ResourceGroupId              string
	Tags                         []tag.Tag
	SecurityGroupIds             []string
	BandwidthPackageId           *string
	DeletionProtectionConfig     *DeletionProtectionConfig
	ModificationProtectionConfig *ModificationProtectionConfig
	PreserveOnDelete             bool

	// auto-generated parameters
	LoadBalancerId             string
	LoadBalancerStatus         string
	LoadBalancerBusinessStatus string
	DNSName                    string
}

type DeletionProtectionConfig struct {
	Enabled bool
	Reason  string
}

type ModificationProtectionConfig struct {
	Status ModificationProtectionType
	Reason string
}

type ProxyProtocolV2Config struct {
	PrivateLinkEpIdEnabled  *bool
	PrivateLinkEpsIdEnabled *bool
	VpcIdEnabled            *bool
}

type ListenerAttribute struct {
	IsUserManaged   bool
	NamedKey        *ListenerNamedKey
	ServerGroupName string
	ServicePort     *v1.ServicePort

	ListenerProtocol      string
	ListenerPort          int32
	ListenerDescription   string
	ServerGroupId         string
	LoadBalancerId        string
	IdleTimeout           int32 // 1-900
	SecurityPolicyId      string
	CertificateIds        []string // tcpssl
	CaCertificateIds      []string
	CaEnabled             *bool
	ProxyProtocolEnabled  *bool
	ProxyProtocolV2Config ProxyProtocolV2Config
	SecSensorEnabled      *bool
	AlpnEnabled           *bool
	AlpnPolicy            string
	StartPort             int32  //0-65535
	EndPort               int32  //0-65535
	Cps                   *int32 //0-1000000

	// auto-generated parameters
	ListenerId string
	ListenerStatus
}

func (a *ListenerAttribute) PortString() string {
	if a.ListenerPort != 0 {
		return fmt.Sprintf("%d", a.ListenerPort)
	}
	return fmt.Sprintf("%d-%d", a.StartPort, a.EndPort)
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
	InvalidServers          []ServerGroupServer
	InitialServers          []ServerGroupServer
	Tags                    []tag.Tag
	AnyPortEnabled          bool
	IgnoreWeightUpdate      bool

	// auto-generated parameters
	ServerGroupId string
}

func (s *ServerGroup) BackendInfo() string {
	if len(s.Servers) > 100 {
		s.Servers = s.Servers[:100]
	}
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

	TargetRef *v1.ObjectReference
	Invalid   bool
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
	Port      int32
	StartPort int32
	EndPort   int32
	Protocol  string
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
	if n.Port != 0 {
		return fmt.Sprintf("%s.%d.%s.%s.%s.%s", n.Prefix, n.Port, n.Protocol, n.ServiceName, n.Namespace, n.CID)
	} else {
		return fmt.Sprintf("%s.%d_%d.%s.%s.%s.%s", n.Prefix, n.StartPort, n.EndPort, n.Protocol, n.ServiceName, n.Namespace, n.CID)
	}
}

func LoadNLBListenerNamedKey(key string) (*ListenerNamedKey, error) {
	metas := strings.Split(key, ".")
	if len(metas) != 6 || metas[0] != model.DEFAULT_PREFIX {
		return nil, fmt.Errorf("ListenerName Format Error: k8s.${port}.${protocol}.${service}.${namespace}.${clusterid} format is expected. Got [%s]", key)
	}
	port, startPort, endPort, err := parseListenerPortKey(metas[1])
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
		Port:      port,
		StartPort: startPort,
		EndPort:   endPort,
		Protocol:  metas[2],
	}, nil
}

func parseListenerPortKey(p string) (int32, int32, int32, error) {
	port, err := strconv.Atoi(p)
	if err == nil {
		return int32(port), 0, 0, nil
	}
	ports := strings.Split(p, "_")
	if len(ports) != 2 {
		return 0, 0, 0, fmt.Errorf("parse listener port key failed")
	}
	startPort, err := strconv.Atoi(ports[0])
	if err != nil {
		return 0, 0, 0, err
	}
	endPort, err := strconv.Atoi(ports[1])
	if err != nil {
		return 0, 0, 0, err
	}
	return 0, int32(startPort), int32(endPort), nil
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
