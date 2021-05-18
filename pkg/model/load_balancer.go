package model

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

type AddressIPVersionType string

const (
	IPv4 = AddressIPVersionType("ipv4")
	IPv6 = AddressIPVersionType("ipv6")
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
	DeleteProtection             FlagType
	ModificationProtectionStatus ModificationProtectionType
	ResourceGroupId              string
	LoadBalancerSpec             LoadBalancerSpecType
	MasterZoneId                 string
	SlaveZoneId                  string
	AddressIPVersion             AddressIPVersionType
	Tags                         []slb.Tag

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
	ForwardPort               int
	EnableHttp2               FlagType
	Cookie                    string
	CookieTimeout             int // values: 1~86400
	StickySession             FlagType
	StickySessionType         string
	XForwardedFor             FlagType
	XForwardedForProto        FlagType
	AclId                     string
	AclType                   string
	AclStatus                 FlagType
	HealthCheck               FlagType
	HealthCheckType           string
	HealthCheckDomain         string
	HealthCheckURI            string
	HealthCheckConnectPort    int
	HealthyThreshold          int // values: 2~10
	UnhealthyThreshold        int // values: 2~10
	HealthCheckTimeout        int // values: 1~300
	HealthCheckConnectTimeout int // values: 1~300
	HealthCheckInterval       int // values: 1-50
	HealthCheckHttpCode       string
	ConnectionDrain           FlagType
	ConnectionDrainTimeout    int // values: 10~900

	// The following parameters can be changed by annotation
	// parameters can be set to the default value.
	// Use the pointer type to distinguish. If the user does not set the param, the param is nil
	PersistenceTimeout *int
}

type VServerGroup struct {
	IsUserManaged bool
	NamedKey      *VGroupNamedKey
	ServicePort   v1.ServicePort

	VGroupId   string
	VGroupName string
	Backends   []BackendAttribute
}

type BackendAttribute struct {
	IsUserManaged bool
	NodeName      *string

	Description string
	ServerId    string
	ServerIp    string
	Weight      int
	Port        int
	Type        string
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
