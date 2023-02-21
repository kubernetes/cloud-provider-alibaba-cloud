package model

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"strconv"
	"strings"
)

type ListenerStatus string

const Stopped = ListenerStatus("stopped")

type AddressType string

const (
	InternetAddressType = AddressType("internet")
	IntranetAddressType = AddressType("intranet")
)

type InternetChargeType string

const PayByBandwidth = InternetChargeType("paybybandwidth")

// InstanceChargeType slb instance charge type
type InstanceChargeType string

func (t InstanceChargeType) IsPayBySpec() bool {
	return t == "" || strings.ToLower(string(t)) == "paybyspec"
}

func (t InstanceChargeType) IsPayByCLCU() bool {
	return strings.ToLower(string(t)) == "paybyclcu"
}

type AddressIPVersionType string

const (
	IPv4      = AddressIPVersionType("ipv4")
	IPv6      = AddressIPVersionType("ipv6")
	DualStack = AddressIPVersionType("dualstack")
)

type LoadBalancerSpecType string

const S1Small = "slb.s1.small"

type ModificationProtectionType string

const ConsoleProtection = ModificationProtectionType("ConsoleProtection")

type FlagType string

const (
	OnFlag  = FlagType("on")
	OffFlag = FlagType("off")
)

const (
	HTTP  = "http"
	HTTPS = "https"
	TCP   = "tcp"
	UDP   = "udp"
)

const (
	ECSBackendType = "ecs"
	ENIBackendType = "eni"
)

const ModificationProtectionReason = "managed.by.ack"

// LoadBalancer represents a AlibabaCloud LoadBalancer.
type LoadBalancer struct {
	NamespacedName        types.NamespacedName
	LoadBalancerAttribute LoadBalancerAttribute
	Listeners             []ListenerAttribute
	VServerGroups         []VServerGroup
}

func (l *LoadBalancer) GetLoadBalancerId() string {
	if l == nil {
		return ""
	}
	return l.LoadBalancerAttribute.LoadBalancerId
}

type LoadBalancerAttribute struct {
	IsUserManaged bool

	// parameters can be modified by annotation
	// values of these parameters can not be set to the default value, no need to use ptr type
	LoadBalancerName             string
	AddressType                  AddressType
	VSwitchId                    string
	NetworkType                  string
	Bandwidth                    int
	InternetChargeType           InternetChargeType
	InstanceChargeType           InstanceChargeType
	DeleteProtection             FlagType
	ModificationProtectionStatus ModificationProtectionType
	ResourceGroupId              string
	LoadBalancerSpec             LoadBalancerSpecType
	MasterZoneId                 string
	SlaveZoneId                  string
	AddressIPVersion             AddressIPVersionType
	Tags                         []tag.Tag

	// parameters are immutable
	RegionId                     string
	LoadBalancerId               string
	LoadBalancerStatus           string
	Address                      string
	VpcId                        string
	CreateTime                   string
	ModificationProtectionReason string
}

type ListenerAttribute struct {
	IsUserManaged bool
	NamedKey      *ListenerNamedKey

	// parameters are immutable
	ListenerPort    int
	Description     string
	Status          ListenerStatus
	ListenerForward FlagType

	VGroupName string
	VGroupId   string

	// parameters can be modified by annotation
	// values of these parameters can not be set to the default value, no need to use ptr type
	Protocol                  string
	Bandwidth                 int // values: -1 or 1~5120
	Scheduler                 string
	CertId                    string
	TLSCipherPolicy           string
	ForwardPort               int
	EnableHttp2               FlagType
	EnableProxyProtocolV2     *bool
	StickySession             FlagType
	StickySessionType         string
	Cookie                    string
	CookieTimeout             int // values: 1~86400
	XForwardedFor             FlagType
	XForwardedForProto        FlagType
	AclId                     string
	AclType                   string
	AclStatus                 FlagType
	ConnectionDrain           FlagType
	ConnectionDrainTimeout    int // values: 10~900
	IdleTimeout               int // values: 1~60
	RequestTimeout            int // values: 1~180, http & https
	EstablishedTimeout        int // values: 10~900, tcp
	HealthCheckConnectPort    int
	HealthCheckInterval       int      // values: 1~50
	HealthyThreshold          int      // values: 2~10
	UnhealthyThreshold        int      // values: 2~10
	HealthCheckType           string   // tcp
	HealthCheckConnectTimeout int      // tcp & udp values: 1~300
	HealthCheckTimeout        int      // http & https values: 1~300
	HealthCheck               FlagType // http & https
	HealthCheckDomain         string   // tcp & http & https
	HealthCheckURI            string   // tcp & http & https
	HealthCheckHttpCode       string   // tcp & http & https
	HealthCheckMethod         string   // http & https

	// The following parameters can be set to the default value.
	// Use the pointer type to distinguish. If the user does not set the param, the param is nil
	PersistenceTimeout *int
}

type VServerGroup struct {
	IsUserManaged bool
	NamedKey      *VGroupNamedKey
	ServicePort   v1.ServicePort

	VGroupId     string
	VGroupName   string
	VGroupWeight *int
	Backends     []BackendAttribute
}

func (v *VServerGroup) BackendInfo() string {
	backendJson, err := json.Marshal(v.Backends)
	if err != nil {
		return fmt.Sprintf("%v", v.Backends)
	}
	return string(backendJson)
}

type BackendAttribute struct {
	IsUserManaged bool
	NodeName      *string

	Description string `json:"description"`
	ServerId    string `json:"serverId"`
	ServerIp    string `json:"serverIp"`
	Weight      int    `json:"weight"`
	Port        int    `json:"port"`
	Type        string `json:"type"`
}

// DEFAULT_PREFIX default prefix for listener
var DEFAULT_PREFIX = "k8s"

// NamedKey identify listeners on grouped attributes
type ListenerNamedKey struct {
	Prefix      string
	CID         string
	Namespace   string
	ServiceName string
	Port        int32
}

func (n *ListenerNamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

func (n *ListenerNamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%d/%s/%s/%s", n.Prefix, n.Port, n.ServiceName, n.Namespace, n.CID)
}

func LoadListenerNamedKey(key string) (*ListenerNamedKey, error) {
	metas := strings.Split(key, "/")
	if len(metas) != 5 || metas[0] != DEFAULT_PREFIX {
		return nil, formatError{key: key}
	}
	port, err := strconv.Atoi(metas[1])
	if err != nil {
		return nil, err
	}
	return &ListenerNamedKey{
		CID:         metas[4],
		Namespace:   metas[3],
		ServiceName: metas[2],
		Port:        int32(port),
		Prefix:      DEFAULT_PREFIX}, nil
}

var FORMAT_ERROR = "ListenerName Format Error: k8s/${port}/${service}/${namespace}/${clusterid} format is expected"

type formatError struct{ key string }

func (f formatError) Error() string { return fmt.Sprintf("%s. Got [%s]", FORMAT_ERROR, f.key) }

// NamedKey identify listeners on grouped attributes
type VGroupNamedKey struct {
	Prefix      string
	CID         string
	Namespace   string
	ServiceName string
	// VGroupPort the port in the vgroup name
	VGroupPort string
}

func (n *VGroupNamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

func (n *VGroupNamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s", n.Prefix, n.VGroupPort, n.ServiceName, n.Namespace, n.CID)
}

func LoadVGroupNamedKey(key string) (*VGroupNamedKey, error) {
	metas := strings.Split(key, "/")
	if len(metas) != 5 || metas[0] != DEFAULT_PREFIX {
		return nil, formatError{key: key}
	}

	return &VGroupNamedKey{
		CID:         metas[4],
		Namespace:   metas[3],
		ServiceName: metas[2],
		VGroupPort:  metas[1],
		Prefix:      DEFAULT_PREFIX}, nil
}
