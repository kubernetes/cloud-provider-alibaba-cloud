package model

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/apimachinery/pkg/types"
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

// LoadBalancer represents a AlibabaCloud LoadBalancer.
type LoadBalancer struct {
	NamespacedName        types.NamespacedName
	LoadBalancerAttribute LoadBalancerAttribute
	Listeners             []ListenerAttribute
	Backends              []BackendAttribute
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
	// values can be modified by annotation
	LoadBalancerId            string
	ListenerPort              int
	BackendServerPort         int
	Bandwidth                 int
	Scheduler                 string
	PersistenceTimeout        int
	HealthCheck               FlagType
	HealthCheckType           string
	HealthCheckDomain         string
	HealthCheckURI            string
	HealthCheckConnectPort    int
	HealthyThreshold          int
	UnhealthyThreshold        int
	HealthCheckConnectTimeout int
	HealthCheckInterval       int
	HealthCheckHttpCode       string

	// values are immutable
	VServerGroup   FlagType
	VServerGroupId string
	Description    string
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
