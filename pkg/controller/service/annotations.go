package service

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
)

// prefix
const (
	// AnnotationLegacyPrefix legacy prefix of service annotation
	AnnotationLegacyPrefix = "service.beta.kubernetes.io/alicloud"
	// AnnotationPrefix prefix of service annotation
	AnnotationPrefix = "service.beta.kubernetes.io/alibaba-cloud"
)

const (
	TAGKEY   = "kubernetes.do.not.delete"
	REUSEKEY = "kubernetes.reused.by.user"
)

// load balancer annotations
const (
	// AnnotationLoadBalancerPrefix loadbalancer prefix
	AnnotationLoadBalancerPrefix = "loadbalancer-"

	// Load Balancer Attribute
	AddressType            = AnnotationLoadBalancerPrefix + "address-type"             // AddressType loadbalancer address type
	VswitchId              = AnnotationLoadBalancerPrefix + "vswitch-id"               // VswitchId loadbalancer vswitch id
	SLBNetworkType         = AnnotationLoadBalancerPrefix + "slb-network-type"         // SLBNetworkType loadbalancer network type
	ChargeType             = AnnotationLoadBalancerPrefix + "charge-type"              // InternetChargeType lb internet charge type, paybytraffic or paybybandwidth
	LoadBalancerId         = AnnotationLoadBalancerPrefix + "id"                       // LoadBalancerId lb id
	OverrideListener       = AnnotationLoadBalancerPrefix + "force-override-listeners" // OverrideListener force override listeners
	LoadBalancerName       = AnnotationLoadBalancerPrefix + "name"                     // LoadBalancerName slb name
	MasterZoneID           = AnnotationLoadBalancerPrefix + "master-zoneid"            // MasterZoneID master zone id
	SlaveZoneID            = AnnotationLoadBalancerPrefix + "slave-zoneid"             // SlaveZoneID slave zone id
	Bandwidth              = AnnotationLoadBalancerPrefix + "bandwidth"                // Bandwidth bandwidth
	AdditionalTags         = AnnotationLoadBalancerPrefix + "additional-resource-tags" // AdditionalTags For example: "Key1=Val1,Key2=Val2,KeyNoVal1=,KeyNoVal2",same with aws
	Spec                   = AnnotationLoadBalancerPrefix + "spec"                     // Spec slb spec
	InstanceChargeType     = AnnotationLoadBalancerPrefix + "instance-charge-type"     // InstanceChargeType the charge type of lb instance
	Scheduler              = AnnotationLoadBalancerPrefix + "scheduler"                // Scheduler slb scheduler
	IPVersion              = AnnotationLoadBalancerPrefix + "ip-version"               // IPVersion ip version
	ResourceGroupId        = AnnotationLoadBalancerPrefix + "resource-group-id"        // ResourceGroupId resource group id
	DeleteProtection       = AnnotationLoadBalancerPrefix + "delete-protection"        // DeleteProtection delete protection
	ModificationProtection = AnnotationLoadBalancerPrefix + "modification-protection"  // ModificationProtection modification type
	ExternalIPType         = AnnotationLoadBalancerPrefix + "external-ip-type"         // ExternalIPType external ip type
	HostName               = AnnotationLoadBalancerPrefix + "hostname"                 // HostName hostname for service.status.ingress.hostname

	// Listener Attribute
	AclStatus                 = AnnotationLoadBalancerPrefix + "acl-status"                   // AclStatus enable or disable acl on all listener
	AclID                     = AnnotationLoadBalancerPrefix + "acl-id"                       // AclID acl id
	AclType                   = AnnotationLoadBalancerPrefix + "acl-type"                     // AclType acl type, black or white
	ProtocolPort              = AnnotationLoadBalancerPrefix + "protocol-port"                // ProtocolPort protocol port
	ForwardPort               = AnnotationLoadBalancerPrefix + "forward-port"                 // ForwardPort loadbalancer forward port
	CertID                    = AnnotationLoadBalancerPrefix + "cert-id"                      // CertID cert id
	EnableHttp2               = AnnotationLoadBalancerPrefix + "http2-enabled"                //EnableHttp2 enable http2 on https port
	HealthCheckFlag           = AnnotationLoadBalancerPrefix + "health-check-flag"            // HealthCheckFlag health check flag
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
	ConnectionDrain           = AnnotationLoadBalancerPrefix + "connection-drain"             // ConnectionDrain connection drain
	ConnectionDrainTimeout    = AnnotationLoadBalancerPrefix + "connection-drain-timeout"     // ConnectionDrainTimeout connection drain timeout
	VGroupPort                = AnnotationLoadBalancerPrefix + "vgroup-port"                  // VGroupIDs binding user managed vGroup ids to ports
	XForwardedForProto        = AnnotationLoadBalancerPrefix + "xforwardedfor-proto"          // XForwardedForProto whether to use the X-Forwarded-Proto header to retrieve the listener protocol
	IdleTimeout               = AnnotationLoadBalancerPrefix + "idle-timeout"                 // IdleTimeout idle timeout for L7
	RequestTimeout            = AnnotationLoadBalancerPrefix + "request-timeout"              // RequestTimeout request timeout for L7
	EstablishedTimeout        = AnnotationLoadBalancerPrefix + "established-timeout"          // EstablishedTimeout connection established time out for TCP

	// VServerBackend Attribute
	BackendLabel      = AnnotationLoadBalancerPrefix + "backend-label"              // BackendLabel backend labels
	BackendType       = "service.beta.kubernetes.io/backend-type"                   // BackendType backend type
	BackendIPVersion  = AnnotationLoadBalancerPrefix + "backend-ip-version"         // BackendIPVersion backend ip version
	RemoveUnscheduled = AnnotationLoadBalancerPrefix + "remove-unscheduled-backend" // RemoveUnscheduled remove unscheduled node from backends
	VGroupWeight      = AnnotationLoadBalancerPrefix + "weight"                     // Weight total weight of the load balancer
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

// TODO get all annotations value from Get()
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

func (n *AnnotationRequest) GetDefaultValue(k string) string {
	if k == LoadBalancerName {
		return n.GetDefaultLoadBalancerName()
	}
	return DefaultValue[composite(AnnotationPrefix, k)]
}

func (n *AnnotationRequest) GetDefaultTags() []model.Tag {
	return []model.Tag{
		{
			TagKey:   TAGKEY,
			TagValue: n.GetDefaultLoadBalancerName(),
		},
		{
			TagKey:   ctrlCfg.CloudCFG.Global.KubernetesClusterTag,
			TagValue: ctrlCfg.CloudCFG.Global.ClusterID,
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
func (n *AnnotationRequest) GetLoadBalancerAdditionalTags() []model.Tag {
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
	var tags []model.Tag
	for k, v := range additionalTags {
		tags = append(tags, model.Tag{
			TagValue: v,
			TagKey:   k,
		})
	}
	return tags
}

func (n *AnnotationRequest) isForceOverride() bool {
	return n.Get(OverrideListener) == "true"
}
