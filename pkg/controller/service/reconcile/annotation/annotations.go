package annotation

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
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
	AddressType      = AnnotationLoadBalancerPrefix + "address-type"             // AddressType loadbalancer address type
	LoadBalancerId   = AnnotationLoadBalancerPrefix + "id"                       // LoadBalancerId lb id
	LoadBalancerName = AnnotationLoadBalancerPrefix + "name"                     // LoadBalancerName slb name
	ResourceGroupId  = AnnotationLoadBalancerPrefix + "resource-group-id"        // ResourceGroupId resource group id
	AdditionalTags   = AnnotationLoadBalancerPrefix + "additional-resource-tags" // AdditionalTags For example: "Key1=Val1,Key2=Val2,KeyNoVal1=,KeyNoVal2",same with aws

	CertID          = AnnotationLoadBalancerPrefix + "cert-id"           // CertID cert id
	ProtocolPort    = AnnotationLoadBalancerPrefix + "protocol-port"     // ProtocolPort protocol port
	IdleTimeout     = AnnotationLoadBalancerPrefix + "idle-timeout"      // IdleTimeout idle timeout for L7
	TLSCipherPolicy = AnnotationLoadBalancerPrefix + "tls-cipher-policy" //TLSCipherPolicy TLS security policy for https

	Scheduler              = AnnotationLoadBalancerPrefix + "scheduler"                // Scheduler slb scheduler
	ConnectionDrain        = AnnotationLoadBalancerPrefix + "connection-drain"         // ConnectionDrain connection drain
	ConnectionDrainTimeout = AnnotationLoadBalancerPrefix + "connection-drain-timeout" // ConnectionDrainTimeout connection drain timeout

)

// classic load balancer

const (
	VswitchId              = AnnotationLoadBalancerPrefix + "vswitch-id"               // VswitchId loadbalancer vswitch id
	SLBNetworkType         = AnnotationLoadBalancerPrefix + "slb-network-type"         // SLBNetworkType loadbalancer network type
	ChargeType             = AnnotationLoadBalancerPrefix + "charge-type"              // InternetChargeType lb internet charge type, paybytraffic or paybybandwidth
	OverrideListener       = AnnotationLoadBalancerPrefix + "force-override-listeners" // OverrideListener force override listeners
	MasterZoneID           = AnnotationLoadBalancerPrefix + "master-zoneid"            // MasterZoneID master zone id
	SlaveZoneID            = AnnotationLoadBalancerPrefix + "slave-zoneid"             // SlaveZoneID slave zone id
	Bandwidth              = AnnotationLoadBalancerPrefix + "bandwidth"                // Bandwidth bandwidth
	Spec                   = AnnotationLoadBalancerPrefix + "spec"                     // Spec slb spec
	InstanceChargeType     = AnnotationLoadBalancerPrefix + "instance-charge-type"     // InstanceChargeType the charge type of lb instance
	IPVersion              = AnnotationLoadBalancerPrefix + "ip-version"               // IPVersion ip version
	DeleteProtection       = AnnotationLoadBalancerPrefix + "delete-protection"        // DeleteProtection delete protection
	ModificationProtection = AnnotationLoadBalancerPrefix + "modification-protection"  // ModificationProtection modification type
	ExternalIPType         = AnnotationLoadBalancerPrefix + "external-ip-type"         // ExternalIPType external ip type
	HostName               = AnnotationLoadBalancerPrefix + "hostname"                 // HostName hostname for service.status.ingress.hostname

	// Listener Attribute
	AclStatus                 = AnnotationLoadBalancerPrefix + "acl-status"                   // AclStatus enable or disable acl on all listener
	AclID                     = AnnotationLoadBalancerPrefix + "acl-id"                       // AclID acl id
	AclType                   = AnnotationLoadBalancerPrefix + "acl-type"                     // AclType acl type, black or white
	ForwardPort               = AnnotationLoadBalancerPrefix + "forward-port"                 // ForwardPort loadbalancer forward port
	EnableHttp2               = AnnotationLoadBalancerPrefix + "http2-enabled"                // EnableHttp2 enable http2 on https port
	HealthCheckSwitch         = AnnotationLoadBalancerPrefix + "health-check-switch"          // HealthCheckSwitch health check switch flag, only for tcp & udp
	HealthCheckFlag           = AnnotationLoadBalancerPrefix + "health-check-flag"            // HealthCheckFlag health check flag, only for http & https
	HealthCheckType           = AnnotationLoadBalancerPrefix + "health-check-type"            // HealthCheckType health check type
	HealthCheckURI            = AnnotationLoadBalancerPrefix + "health-check-uri"             // HealthCheckURI health check uri
	HealthCheckConnectPort    = AnnotationLoadBalancerPrefix + "health-check-connect-port"    // HealthCheckConnectPort health check connect port
	HealthyThreshold          = AnnotationLoadBalancerPrefix + "healthy-threshold"            // HealthyThreshold health check healthy thresh hold
	UnhealthyThreshold        = AnnotationLoadBalancerPrefix + "unhealthy-threshold"          // UnhealthyThreshold health check unhealthy thresh hold
	HealthCheckInterval       = AnnotationLoadBalancerPrefix + "health-check-interval"        // HealthCheckInterval health check interval
	HealthCheckConnectTimeout = AnnotationLoadBalancerPrefix + "health-check-connect-timeout" // HealthCheckConnectTimeout health check connect timeout
	HealthCheckTimeout        = AnnotationLoadBalancerPrefix + "health-check-timeout"         // HealthCheckTimeout health check timeout
	HealthCheckDomain         = AnnotationLoadBalancerPrefix + "health-check-domain"          // HealthCheckDomain health check domain
	HealthCheckHTTPCode       = AnnotationLoadBalancerPrefix + "health-check-httpcode"        // HealthCheckHTTPCode health check http code
	HealthCheckMethod         = AnnotationLoadBalancerPrefix + "health-check-method"          // HealthCheckMethod health check method for L7
	SessionStick              = AnnotationLoadBalancerPrefix + "sticky-session"               // SessionStick sticky session
	SessionStickType          = AnnotationLoadBalancerPrefix + "sticky-session-type"          // SessionStickType session sticky type
	CookieTimeout             = AnnotationLoadBalancerPrefix + "cookie-timeout"               // CookieTimeout cookie timeout
	Cookie                    = AnnotationLoadBalancerPrefix + "cookie"                       // Cookie lb cookie
	PersistenceTimeout        = AnnotationLoadBalancerPrefix + "persistence-timeout"          // PersistenceTimeout persistence timeout
	VGroupPort                = AnnotationLoadBalancerPrefix + "vgroup-port"                  // VGroupIDs binding user managed vGroup ids to ports
	XForwardedForProto        = AnnotationLoadBalancerPrefix + "xforwardedfor-proto"          // XForwardedForProto whether to use the X-Forwarded-Proto header to retrieve the listener protocol
	RequestTimeout            = AnnotationLoadBalancerPrefix + "request-timeout"              // RequestTimeout request timeout for L7
	EstablishedTimeout        = AnnotationLoadBalancerPrefix + "established-timeout"          // EstablishedTimeout connection established time out for TCP

	// VServerBackend Attribute
	BackendLabel      = AnnotationLoadBalancerPrefix + "backend-label"              // BackendLabel backend labels
	BackendType       = helper.BackendType                                          // BackendType backend type
	BackendIPVersion  = AnnotationLoadBalancerPrefix + "backend-ip-version"         // BackendIPVersion backend ip version
	RemoveUnscheduled = AnnotationLoadBalancerPrefix + "remove-unscheduled-backend" // RemoveUnscheduled remove unscheduled node from backends
	VGroupWeight      = AnnotationLoadBalancerPrefix + "weight"                     // Weight total weight of the load balancer
)

// network load balancer
const (
	ZoneMaps         = AnnotationLoadBalancerPrefix + "zone-maps" // ZoneMaps zone maps
	SecurityGroupIds = AnnotationLoadBalancerPrefix + "security-group-ids"

	ProxyProtocol = AnnotationLoadBalancerPrefix + "proxy-protocol"
	CaCertID      = AnnotationLoadBalancerPrefix + "cacert-id" // CertID cert id
	CaCert        = AnnotationLoadBalancerPrefix + "cacert"    // CaCert enable ca
	Cps           = AnnotationLoadBalancerPrefix + "cps"

	PreserveClientIp = AnnotationLoadBalancerPrefix + "preserve-client-ip"
)

var DefaultValue = map[string]string{
	composite(AnnotationPrefix, AddressType):            string(model.InternetAddressType),
	composite(AnnotationPrefix, Spec):                   model.S1Small,
	composite(AnnotationPrefix, IPVersion):              string(model.IPv4),
	composite(AnnotationPrefix, DeleteProtection):       string(model.OnFlag),
	composite(AnnotationPrefix, ModificationProtection): string(model.ConsoleProtection),
}

type AnnotationRequest struct{ Service *v1.Service }

func NewAnnotationRequest(svc *v1.Service) *AnnotationRequest {
	return &AnnotationRequest{svc}
}

func (n *AnnotationRequest) Get(k string) string {
	if n.Service == nil {
		return ""
	}

	if n.Service.Annotations == nil {
		return ""
	}

	key := composite(AnnotationPrefix, k)
	v, ok := n.Service.Annotations[key]
	if ok {
		return v
	}

	lkey := composite(AnnotationLegacyPrefix, k)
	v, ok = n.Service.Annotations[lkey]
	if ok {
		return v
	}

	return ""
}

func (n *AnnotationRequest) Has(k string) bool {
	if n.Service == nil {
		return false
	}

	if n.Service.Annotations == nil {
		return false
	}

	key := composite(AnnotationPrefix, k)
	if _, ok := n.Service.Annotations[key]; ok {
		return true
	}

	key = composite(AnnotationLegacyPrefix, k)
	if _, ok := n.Service.Annotations[key]; ok {
		return true
	}

	return false
}

func (n *AnnotationRequest) GetDefaultValue(k string) string {
	if k == LoadBalancerName {
		return n.GetDefaultLoadBalancerName()
	}
	return DefaultValue[composite(AnnotationPrefix, k)]
}

func (n *AnnotationRequest) GetDefaultTags() []tag.Tag {
	return []tag.Tag{
		{
			Key:   helper.TAGKEY,
			Value: n.GetDefaultLoadBalancerName(),
		},
		{
			Key:   util.ClusterTagKey,
			Value: base.CLUSTER_ID,
		},
	}
}

func (n *AnnotationRequest) GetDefaultLoadBalancerName() string {
	//GCE requires that the name of a load balancer starts with a lower case letter.
	ret := "a" + string(n.Service.UID)
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

func Annotation(k string) string {
	return composite(AnnotationPrefix, k)
}

// getLoadBalancerAdditionalTags converts the comma separated list of key-value
// pairs in the ServiceAnnotationLoadBalancerAdditionalTags annotation and returns
// it as a map.
func (n *AnnotationRequest) GetLoadBalancerAdditionalTags() []tag.Tag {
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
	var tags []tag.Tag
	for k, v := range additionalTags {
		tags = append(tags, tag.Tag{
			Key:   k,
			Value: v,
		})
	}
	return tags
}

func (n *AnnotationRequest) IsForceOverride() bool {
	return n.Get(OverrideListener) == "true"
}
