package service

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/klog"
)

// prefix
const (
	// AnnotationLegacyPrefix legacy prefix of service annotation
	AnnotationLegacyPrefix = "service.beta.kubernetes.io/alicloud"
	// AnnotationPrefix prefix of service annotation
	AnnotationPrefix = "service.beta.kubernetes.io/alibaba-cloud"
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
	HealthCheckHealthyThreshold = AnnotationLoadBalancerPrefix + "healthy-threshold"
	// ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold health check unhealthy thresh hold
	HealthCheckUnhealthyThreshold = AnnotationLoadBalancerPrefix + "unhealthy-threshold"
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

	// VServerBackend Attribute
	// ServiceAnnotationLoadBalancerBackendLabel backend labels
	BackendLabel = AnnotationLoadBalancerPrefix + "backend-label"
	// ServiceAnnotationLoadBalancerBackendType backend type
	BackendType = "service.beta.kubernetes.io/backend-type"
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
	composite(AnnotationPrefix, ModificationProtection): string(model.OnFlag),
}

type AnnotationRequest struct{ svc *v1.Service }

func (n *AnnotationRequest) Get(k string) *string {
	if n.svc == nil {
		klog.Infof("extract annotation %s from empty service", k)
		return nil
	}

	if n.svc.Annotations == nil {
		return nil
	}

	key := composite(AnnotationPrefix, k)
	v, ok := n.svc.Annotations[key]
	if ok {
		return &v
	}

	lkey := composite(AnnotationLegacyPrefix, k)
	v, ok = n.svc.Annotations[lkey]
	if ok {
		return &v
	}

	return nil
}

func (n *AnnotationRequest) GetRaw(k string) string {
	return n.svc.Annotations[k]
}

func (n *AnnotationRequest) GetDefaultValue(k string) string {
	return DefaultValue[composite(AnnotationPrefix, k)]
}

func composite(p, k string) string {
	return fmt.Sprintf("%s-%s", p, k)
}
