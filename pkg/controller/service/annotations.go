package service

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	v1 "k8s.io/api/core/v1"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/klog"
	"strings"
)

// prefix
const (
	// AnnotationLegacyPrefix legacy prefix of service annotation
	AnnotationLegacyPrefix = "service.beta.kubernetes.io/alicloud"
	// AnnotationPrefix prefix of service annotation
	AnnotationPrefix = "service.beta.kubernetes.io/alibaba-cloud"
)

const (
	TAGKEY = "kubernetes.do.not.delete"
	ACKKEY = "ack.aliyun.com"
)

// load balancer annotations
const (
	// AnnotationLoadBalancerPrefix loadbalancer prefix
	AnnotationLoadBalancerPrefix = "loadbalancer-"

	// Load Balancer Attribute
	// ServiceAnnotationLoadBalancerAddressType loadbalancer address type
	AddressType = AnnotationLoadBalancerPrefix + "address-type"
	// ServiceAnnotationLoadBalancerVswitch loadbalancer vswitch id
	VswitchId = AnnotationLoadBalancerPrefix + "vswitch-id"
	// ServiceAnnotationLoadBalancerSLBNetworkType loadbalancer network type
	SLBNetworkType = AnnotationLoadBalancerPrefix + "slb-network-type"
	// ServiceAnnotationLoadBalancerChargeType lb charge type
	ChargeType = AnnotationLoadBalancerPrefix + "charge-type"
	// ServiceAnnotationLoadBalancerId lb id
	LoadBalancerId = AnnotationLoadBalancerPrefix + "id"
	// ServiceAnnotationLoadBalancerOverrideListener force override listeners
	OverrideListener = AnnotationLoadBalancerPrefix + "force-override-listeners"
	//ServiceAnnotationLoadBalancerName slb name
	LoadBalancerName = AnnotationLoadBalancerPrefix + "name"
	// ServiceAnnotationLoadBalancerMasterZoneID master zone id
	MasterZoneID = AnnotationLoadBalancerPrefix + "master-zoneid"
	// ServiceAnnotationLoadBalancerSlaveZoneID slave zone id
	SlaveZoneID = AnnotationLoadBalancerPrefix + "slave-zoneid"
	// ServiceAnnotationLoadBalancerBandwidth bandwidth
	Bandwidth = AnnotationLoadBalancerPrefix + "bandwidth"
	// ServiceAnnotationLoadBalancerAdditionalTags For example: "Key1=Val1,Key2=Val2,KeyNoVal1=,KeyNoVal2",same with aws
	AdditionalTags = AnnotationLoadBalancerPrefix + "additional-resource-tags"
	// ServiceAnnotationLoadBalancerSpec slb spec
	Spec = AnnotationLoadBalancerPrefix + "spec"
	// ServiceAnnotationLoadBalancerScheduler slb scheduler
	Scheduler = AnnotationLoadBalancerPrefix + "scheduler"
	//ServiceAnnotationLoadBalancerIPVersion ip version
	IPVersion = AnnotationLoadBalancerPrefix + "ip-version"
	// ServiceAnnotationLoadBalancerResourceGroupId resource group id
	ResourceGroupId = AnnotationLoadBalancerPrefix + "resource-group-id"
	// ServiceAnnotationLoadBalancerDeleteProtection delete protection
	DeleteProtection = AnnotationLoadBalancerPrefix + "delete-protection"
	// ServiceAnnotationLoadBalancerModificationProtection modification type
	ModificationProtection = AnnotationLoadBalancerPrefix + "modification-protection"
	// ServiceAnnotationLoadBalancerBackendType external ip type
	ExternalIPType = AnnotationLoadBalancerPrefix + "external-ip-type"

	// Listener Attribute
	// ServiceAnnotationLoadBalancerAclStatus enable or disable acl on all listener
	AclStatus = AnnotationLoadBalancerPrefix + "acl-status"
	// ServiceAnnotationLoadBalancerAclID acl id
	AclID = AnnotationLoadBalancerPrefix + "acl-id"
	// ServiceAnnotationLoadBalancerAclType acl type, black or white
	AclType = AnnotationLoadBalancerPrefix + "acl-type"
	// ServiceAnnotationLoadBalancerProtocolPort protocol port
	ProtocolPort = AnnotationLoadBalancerPrefix + "protocol-port"
	// ServiceAnnotationLoadBalancerForwardPort loadbalancer forward port
	ForwardPort = AnnotationLoadBalancerPrefix + "forward-port"
	// ServiceAnnotationLoadBalancerCertID cert id
	CertID = AnnotationLoadBalancerPrefix + "cert-id"
	// ServiceAnnotationLoadBalancerHealthCheckFlag health check flag
	HealthCheckFlag = AnnotationLoadBalancerPrefix + "health-check-flag"
	// ServiceAnnotationLoadBalancerHealthCheckType health check type
	HealthCheckType = AnnotationLoadBalancerPrefix + "health-check-type"
	// ServiceAnnotationLoadBalancerHealthCheckURI health check uri
	HealthCheckURI = AnnotationLoadBalancerPrefix + "health-check-uri"
	// ServiceAnnotationLoadBalancerHealthCheckConnectPort health check connect port
	HealthCheckConnectPort = AnnotationLoadBalancerPrefix + "health-check-connect-port"
	// ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold health check healthy thresh hold
	HealthyThreshold = AnnotationLoadBalancerPrefix + "healthy-threshold"
	// ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold health check unhealthy thresh hold
	UnhealthyThreshold = AnnotationLoadBalancerPrefix + "unhealthy-threshold"
	// ServiceAnnotationLoadBalancerHealthCheckInterval health check interval
	HealthCheckInterval = AnnotationLoadBalancerPrefix + "health-check-interval"
	// ServiceAnnotationLoadBalancerHealthCheckConnectTimeout health check connect timeout
	HealthCheckConnectTimeout = AnnotationLoadBalancerPrefix + "health-check-connect-timeout"
	// ServiceAnnotationLoadBalancerHealthCheckTimeout health check timeout
	HealthCheckTimeout = AnnotationLoadBalancerPrefix + "health-check-timeout"
	// ServiceAnnotationLoadBalancerHealthCheckDomain health check domain
	HealthCheckDomain = AnnotationLoadBalancerPrefix + "health-check-domain"
	// ServiceAnnotationLoadBalancerHealthCheckHTTPCode health check http code
	HealthCheckHTTPCode = AnnotationLoadBalancerPrefix + "health-check-httpcode"
	// ServiceAnnotationLoadBalancerSessionStick sticky session
	SessionStick = AnnotationLoadBalancerPrefix + "sticky-session"
	// ServiceAnnotationLoadBalancerSessionStickType session sticky type
	SessionStickType = AnnotationLoadBalancerPrefix + "sticky-session-type"
	// ServiceAnnotationLoadBalancerCookieTimeout cookie timeout
	CookieTimeout = AnnotationLoadBalancerPrefix + "cookie-timeout"
	//ServiceAnnotationLoadBalancerCookie lb cookie
	Cookie = AnnotationLoadBalancerPrefix + "cookie"
	// ServiceAnnotationLoadBalancerPersistenceTimeout persistence timeout
	PersistenceTimeout = AnnotationLoadBalancerPrefix + "persistence-timeout"
	// ConnectionDrain connection drain
	ConnectionDrain = AnnotationLoadBalancerPrefix + "connection-drain"
	// ConnectionDrainTimeout connection drain timeout
	ConnectionDrainTimeout = AnnotationLoadBalancerPrefix + "connection-drain-timeout"

	// VServerBackend Attribute
	// ServiceAnnotationLoadBalancerBackendLabel backend labels
	BackendLabel = AnnotationLoadBalancerPrefix + "backend-label"
	// ServiceAnnotationLoadBalancerBackendType backend type
	BackendType = "service.beta.kubernetes.io/backend-type"

	// RemoveUnscheduled remove unscheduled node from backends
	RemoveUnscheduled = AnnotationLoadBalancerPrefix + "remove-unscheduled-backend"
)

const (
	// ServiceAnnotationPrivateZonePrefix private zone prefix
	ServiceAnnotationPrivateZonePrefix = "private-zone-"
	// ServiceAnnotationLoadBalancerPrivateZoneName private zone name
	PrivateZoneName = ServiceAnnotationPrivateZonePrefix + "name"

	// ServiceAnnotationLoadBalancerPrivateZoneId private zone id
	PrivateZoneId = ServiceAnnotationPrivateZonePrefix + "id"

	// ServiceAnnotationLoadBalancerPrivateZoneRecordName private zone record name
	PrivateZoneRecordName = ServiceAnnotationPrivateZonePrefix + "record-name"

	// ServiceAnnotationLoadBalancerPrivateZoneRecordTTL private zone record ttl
	PrivateZoneRecordTTL = ServiceAnnotationPrivateZonePrefix + "record-ttl"
)

var DefaultValue = map[string]string{
	composite(AnnotationPrefix, AddressType):            string(model.InternetAddressType),
	composite(AnnotationPrefix, Spec):                   model.S1Small,
	composite(AnnotationPrefix, IPVersion):              string(model.IPv4),
	composite(AnnotationPrefix, DeleteProtection):       string(model.OnFlag),
	composite(AnnotationPrefix, ModificationProtection): string(model.ConsoleProtection),
}

type AnnotationRequest struct{ svc *v1.Service }

func (n *AnnotationRequest) Get(k string) string {
	if n.svc == nil {
		klog.Infof("extract annotation %s from empty service", k)
		return ""
	}

	if n.svc.Annotations == nil {
		return ""
	}

	key := composite(AnnotationPrefix, k)
	v, ok := n.svc.Annotations[key]
	if ok {
		return v
	}

	lkey := composite(AnnotationLegacyPrefix, k)
	v, ok = n.svc.Annotations[lkey]
	if ok {
		return v
	}

	return ""
}

func (n *AnnotationRequest) GetRaw(k string) string {
	return n.svc.Annotations[k]
}

func (n *AnnotationRequest) GetDefaultValue(k string) string {
	return DefaultValue[composite(AnnotationPrefix, k)]
}

func (n *AnnotationRequest) GetDefaultTags() []slb.Tag {
	return []slb.Tag{
		{
			TagKey:   TAGKEY,
			TagValue: n.GetDefaultValue(LoadBalancerName),
		},
		{
			TagKey:   ACKKEY,
			TagValue: ctx2.CFG.Global.ClusterID,
		},
	}
}

func (n *AnnotationRequest) GetDefaultLoadBalancerName() string {
	//GCE requires that the name of a load balancer starts with a lower case letter.
	ret := "a" + string(n.svc.UID)
	ret = strings.Replace(ret, "-", "", -1)
	//AWS requires that the name of a load balancer is shorter than 32 bytes.
	if len(ret) > 32 {
		ret = ret[:32]
	}
	return ret
}

func composite(p, k string) string {
	return fmt.Sprintf("%s-%s", p, k)
}

// getLoadBalancerAdditionalTags converts the comma separated list of key-value
// pairs in the ServiceAnnotationLoadBalancerAdditionalTags annotation and returns
// it as a map.
func (n *AnnotationRequest) GetLoadBalancerAdditionalTags() []slb.Tag {
	additionalTags := make(map[string]string)
	additionalTagsList := n.Get(AdditionalTags)
	if additionalTagsList != "" {
		additionalTagsList = strings.TrimSpace(additionalTagsList)
		// Break up list of "Key1=Val,Key2=Val2"
		tagList := strings.Split(additionalTagsList, ",")

		// Break up "Key=Val"
		for _, tagSet := range tagList {
			tag := strings.Split(strings.TrimSpace(tagSet), "=")

			// Accept "Key=val" or "Key=" or just "Key"
			if len(tag) >= 2 && len(tag[0]) != 0 {
				// There is a key and a value, so save it
				additionalTags[tag[0]] = tag[1]
			} else if len(tag) == 1 && len(tag[0]) != 0 {
				// Just "Key"
				additionalTags[tag[0]] = ""
			}
		}
	}
	var tags []slb.Tag
	for k, v := range additionalTags {
		tags = append(tags, slb.Tag{
			TagValue: v,
			TagKey:   k,
		})
	}
	return tags
}
