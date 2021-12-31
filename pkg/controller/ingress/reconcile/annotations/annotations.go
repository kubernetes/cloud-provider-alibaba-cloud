package annotations

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// prefix
const (
	// DefaultAnnotationsPrefix defines the common prefix used in the nginx ingress controller
	DefaultAnnotationsPrefix = "alb.ingress.kubernetes.io"

	// AnnotationLegacyPrefix legacy prefix of service annotation
	AnnotationLegacyPrefix = "service.beta.kubernetes.io/alicloud"
	// AnnotationPrefix prefix of service annotation
	AnnotationPrefix = "service.beta.kubernetes.io/alibaba-cloud"
)

const (
	TAGKEY  = "kubernetes.do.not.delete"
	ACKKEY  = "ack.aliyun.com"
	PostPay = "PostPay"
)
const (
	INGRESS_ALB_ANNOTATIONS = "alb.ingress.kubernetes.io/lb"
	// config the rule conditions
	INGRESS_ALB_CONDITIONS_ANNOTATIONS = "alb.ingress.kubernetes.io/conditions.%s"
	// config the rule actions
	INGRESS_ALB_ACTIONS_ANNOTATIONS = "alb.ingress.kubernetes.io/actions.%s"
	// config the certificate id
	INGRESS_ALB_CERTIFICATE_ANNOTATIONS = "alb.ingress.kubernetes.io/certificate-id"
)

// load balancer annotations
const (
	// AnnotationLoadBalancerPrefix loadbalancer prefix
	AnnotationLoadBalancerPrefix = "alb.ingress.kubernetes.io/"

	// Load Balancer Attribute
	AddressType            = AnnotationLoadBalancerPrefix + "address-type"             // AddressType loadbalancer address type
	VswitchIds             = AnnotationLoadBalancerPrefix + "vswitch-ids"              // VswitchId loadbalancer vswitch id
	SLBNetworkType         = AnnotationLoadBalancerPrefix + "slb-network-type"         // SLBNetworkType loadbalancer network type
	ChargeType             = AnnotationLoadBalancerPrefix + "charge-type"              // ChargeType lb charge type
	LoadBalancerId         = AnnotationLoadBalancerPrefix + "id"                       // LoadBalancerId lb id
	OverrideListener       = AnnotationLoadBalancerPrefix + "force-override-listeners" // OverrideListener force override listeners
	LoadBalancerName       = AnnotationLoadBalancerPrefix + "name"                     // LoadBalancerName slb name
	MasterZoneID           = AnnotationLoadBalancerPrefix + "master-zoneid"            // MasterZoneID master zone id
	SlaveZoneID            = AnnotationLoadBalancerPrefix + "slave-zoneid"             // SlaveZoneID slave zone id
	Bandwidth              = AnnotationLoadBalancerPrefix + "bandwidth"                // Bandwidth bandwidth
	AdditionalTags         = AnnotationLoadBalancerPrefix + "additional-resource-tags" // AdditionalTags For example: "Key1=Val1,Key2=Val2,KeyNoVal1=,KeyNoVal2",same with aws
	Spec                   = AnnotationLoadBalancerPrefix + "spec"                     // Spec slb spec
	Scheduler              = AnnotationLoadBalancerPrefix + "scheduler"                // Scheduler slb scheduler
	IPVersion              = AnnotationLoadBalancerPrefix + "ip-version"               // IPVersion ip version
	ResourceGroupId        = AnnotationLoadBalancerPrefix + "resource-group-id"        // ResourceGroupId resource group id
	DeleteProtection       = AnnotationLoadBalancerPrefix + "delete-protection"        // DeleteProtection delete protection
	ModificationProtection = AnnotationLoadBalancerPrefix + "modification-protection"  // ModificationProtection modification type
	ExternalIPType         = AnnotationLoadBalancerPrefix + "external-ip-type"         // ExternalIPType external ip type
	AddressAllocatedMode   = AnnotationLoadBalancerPrefix + "address-allocated-mode"
	LoadBalancerEdition    = AnnotationLoadBalancerPrefix + "edition"
	HTTPS                  = AnnotationLoadBalancerPrefix + "backend-protocol"
	ListenPorts            = AnnotationLoadBalancerPrefix + "listen-ports"
	AccessLog              = AnnotationLoadBalancerPrefix + "access-log"
	// Listener Attribute
	AclStatus              = AnnotationLoadBalancerPrefix + "acl-status"                   // AclStatus enable or disable acl on all listener
	AclID                  = AnnotationLoadBalancerPrefix + "acl-id"                       // AclID acl id
	AclType                = AnnotationLoadBalancerPrefix + "acl-type"                     // AclType acl type, black or white
	ProtocolPort           = AnnotationLoadBalancerPrefix + "protocol-port"                // ProtocolPort protocol port
	ForwardPort            = AnnotationLoadBalancerPrefix + "forward-port"                 // ForwardPort loadbalancer forward port
	CertID                 = AnnotationLoadBalancerPrefix + "cert-id"                      // CertID cert id
	HealthCheckEnabled     = AnnotationLoadBalancerPrefix + "healthcheck-enabled"          // HealthCheckFlag health check flag
	HealthCheckPath        = AnnotationLoadBalancerPrefix + "healthcheck-path"             // HealthCheckType health check type
	HealthCheckPort        = AnnotationLoadBalancerPrefix + "healthcheck-port"             // HealthCheckURI health check uri
	HealthCheckProtocol    = AnnotationLoadBalancerPrefix + "healthcheck-protocol"         // HealthCheckConnectPort health check connect port
	HealthCheckMethod      = AnnotationLoadBalancerPrefix + "healthcheck-method"           // HealthyThreshold health check healthy thresh hold
	HealthCheckInterval    = AnnotationLoadBalancerPrefix + "healthcheck-interval-seconds" // HealthCheckInterval health check interval
	HealthCheckTimeout     = AnnotationLoadBalancerPrefix + "healthcheck-timeout-seconds"  // HealthCheckTimeout health check timeout
	HealthThreshold        = AnnotationLoadBalancerPrefix + "healthy-threshold-count"      // HealthCheckDomain health check domain
	UnHealthThreshold      = AnnotationLoadBalancerPrefix + "unhealthy-threshold-count"    // HealthCheckHTTPCode health check http code
	HealthCheckHTTPCode    = AnnotationLoadBalancerPrefix + "healthcheck-httpcode"
	SessionStick           = AnnotationLoadBalancerPrefix + "sticky-session"           // SessionStick sticky session
	SessionStickType       = AnnotationLoadBalancerPrefix + "sticky-session-type"      // SessionStickType session sticky type
	CookieTimeout          = AnnotationLoadBalancerPrefix + "cookie-timeout"           // CookieTimeout cookie timeout
	Cookie                 = AnnotationLoadBalancerPrefix + "cookie"                   // Cookie lb cookie
	PersistenceTimeout     = AnnotationLoadBalancerPrefix + "persistence-timeout"      // PersistenceTimeout persistence timeout
	ConnectionDrain        = AnnotationLoadBalancerPrefix + "connection-drain"         // ConnectionDrain connection drain
	ConnectionDrainTimeout = AnnotationLoadBalancerPrefix + "connection-drain-timeout" // ConnectionDrainTimeout connection drain timeout
	PortVGroup             = AnnotationLoadBalancerPrefix + "port-vgroup"              // VGroupIDs binding user managed vGroup ids to ports

	// VServerBackend Attribute
	BackendLabel      = AnnotationLoadBalancerPrefix + "backend-label"              // BackendLabel backend labels
	BackendType       = "service.beta.kubernetes.io/backend-type"                   // BackendType backend type
	RemoveUnscheduled = AnnotationLoadBalancerPrefix + "remove-unscheduled-backend" // RemoveUnscheduled remove unscheduled node from backends
	VGroupWeight      = AnnotationLoadBalancerPrefix + "vgroup-weight"              // VGroupWeight total weight of a vGroup
)

const (
	AnnotationNginxPrefix    = "nginx.ingress.kubernetes.io/"
	NginxCanary              = AnnotationNginxPrefix + "canary"
	NginxCanaryByHeader      = AnnotationNginxPrefix + "canary-by-header"
	NginxCanaryByHeaderValue = AnnotationNginxPrefix + "canary-by-header-value"
	NginxCanaryByCookie      = AnnotationNginxPrefix + "canary-by-cookie"
	NginxCanaryWeight        = AnnotationNginxPrefix + "canary-weight"
	NginxSslRedirect         = AnnotationNginxPrefix + "ssl-redirect"

	AnnotationAlbPrefix    = "alb.ingress.kubernetes.io/"
	AlbCanary              = AnnotationAlbPrefix + "canary"
	AlbCanaryByHeader      = AnnotationAlbPrefix + "canary-by-header"
	AlbCanaryByHeaderValue = AnnotationAlbPrefix + "canary-by-header-value"
	AlbCanaryByCookie      = AnnotationAlbPrefix + "canary-by-cookie"
	AlbCanaryWeight        = AnnotationAlbPrefix + "canary-weight"
	AlbSslRedirect         = AnnotationAlbPrefix + "ssl-redirect"
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

func NewAnnotationRequest(svc *corev1.Service) *AnnotationRequest {
	return &AnnotationRequest{svc}
}

func (n *AnnotationRequest) Get(k string) string {
	if n.Svc == nil {
		klog.Infof("extract annotation %s from empty service", k)
		return ""
	}

	if n.Svc.Annotations == nil {
		return ""
	}

	key := composite(AnnotationPrefix, k)
	v, ok := n.Svc.Annotations[key]
	if ok {
		return v
	}

	lkey := composite(AnnotationLegacyPrefix, k)
	v, ok = n.Svc.Annotations[lkey]
	if ok {
		return v
	}

	return ""
}

func (n *AnnotationRequest) GetRaw(k string) string {
	return n.Svc.Annotations[k]
}

func (n *AnnotationRequest) GetDefaultLoadBalancerName() string {
	//GCE requires that the name of a load balancer starts with a lower case letter.
	ret := "a" + string(n.Svc.UID)
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

type AnnotationRequest struct{ Svc *corev1.Service }
