package model

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"strings"
)

type ListenerStatus string

const (
	Starting    = ListenerStatus("starting")
	Running     = ListenerStatus("running")
	Configuring = ListenerStatus("configuring")
	Stopping    = ListenerStatus("stopping")
	Stopped     = ListenerStatus("stopped")
)

type AddressType string

const (
	InternetAddressType = AddressType("internet")
	IntranetAddressType = AddressType("intranet")
)

type InternetChargeType string

const (
	PayByBandwidth = InternetChargeType("paybybandwidth")
	PayByTraffic   = InternetChargeType("paybytraffic")
)

type AddressIPVersionType string

const (
	IPv4 = AddressIPVersionType("ipv4")
	IPv6 = AddressIPVersionType("ipv6")
)

type LoadBalancerSpecType string

const (
	S1Small  = "slb.s1.small"
	S2Small  = "slb.s2.small"
	S2Medium = "slb.s2.medium"
	S3Small  = "slb.s3.small"
	S3Medium = "slb.s3.medium"
	S3Large  = "slb.s3.large"
)

type ModificationProtectionType string

const (
	NonProtection     = ModificationProtectionType("NonProtection")
	ConsoleProtection = ModificationProtectionType("ConsoleProtection")
)

type SchedulerType string

const (
	WRRScheduler = SchedulerType("wrr")
	WLCScheduler = SchedulerType("wlc")
)

type FlagType string

const (
	OnFlag  = FlagType("on")
	OffFlag = FlagType("off")
)

type StickySessionType string

const (
	InsertStickySessionType = StickySessionType("insert")
	ServerStickySessionType = StickySessionType("server")
)

type Protocol string

const (
	HTTP  = "http"
	HTTPS = "https"
	TCP   = "tcp"
	UDP   = "udp"
)

// LoadBalancer represents a AlibabaCloud LoadBalancer.
type LoadBalancer struct {
	NamespacedName        types.NamespacedName
	LoadBalancerAttribute LoadBalancerAttribute
	Listeners             []ListenerAttribute
	VServerGroups         []VServerGroup
}

type LoadBalancerAttribute struct {
	IsUserManaged bool

	// values can be modified by annotation
	LoadBalancerName             *string
	AddressType                  *string
	VSwitchId                    *string
	NetworkType                  *string
	Bandwidth                    *int
	InternetChargeType           *string
	DeleteProtection             *string
	ModificationProtectionStatus *string
	ResourceGroupId              *string
	LoadBalancerSpec             *string
	MasterZoneId                 *string
	SlaveZoneId                  *string
	AddressIPVersion             *string
	Tags                         []slb.Tag

	// values are immutable
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

	Description     string
	ListenerForward string
	ListenerPort    int
	Protocol        string
	Status          string

	// VServerGroup
	VGroupName string
	VGroupId   string

	// The following parameters can be changed by annotation
	// Use pointer types to distinguish between real default values and user settings as default values
	Bandwidth          *int
	Scheduler          *string
	PersistenceTimeout *int
	CertId             *string
	EnableHttp2        *string
	ForwardPort        *int

	Cookie            *string
	CookieTimeout     *int
	StickySession     *string
	StickySessionType *string

	// XForwardedFor
	XForwardedFor      *string
	XForwardedForProto *string

	// ACL
	AclId     *string
	AclType   *string
	AclStatus *string

	// health check
	HealthCheck               *string
	HealthCheckType           *string
	HealthCheckDomain         *string
	HealthCheckURI            *string
	HealthCheckConnectPort    *int
	HealthyThreshold          *int
	UnhealthyThreshold        *int
	HealthCheckConnectTimeout *int
	HealthCheckInterval       *int
	HealthCheckHttpCode       *string

	// connection drain
	ConnectionDrain        *string
	ConnectionDrainTimeout *int
}

type VServerGroup struct {
	NamedKey *VGroupNamedKey

	VGroupId   string
	VGroupName string
	Backends   []BackendAttribute
}

type BackendAttribute struct {
	IsUserManaged bool
	Description   string
	ServerId      string
	ServerIp      string
	Weight        int
	Port          int
	Type          string
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
