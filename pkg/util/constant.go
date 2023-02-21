package util

import (
	"math/rand"
	"time"
)

const (
	Action = "action"

	CreateALBListener                               = "CreateALBListener"
	CreateALBListenerAsynchronous                   = "CreateALBListenerAsynchronous"
	DeleteALBListener                               = "DeleteALBListener"
	UpdateALBListenerAttribute                      = "UpdateALBListenerAttribute"
	ListALBListeners                                = "ListALBListeners"
	GetALBListenerAttribute                         = "GetALBListenerAttribute"
	ListALBListenerCertificates                     = "ListALBListenerCertificates"
	AssociateALBAdditionalCertificatesWithListener  = "AssociateALBAdditionalCertificatesWithListener"
	DissociateALBAdditionalCertificatesFromListener = "DissociateALBAdditionalCertificatesFromListener"

	CreateALBLoadBalancer             = "CreateALBLoadBalancer"
	CreateALBLoadBalancerAsynchronous = "CreateALBLoadBalancerAsynchronous"
	DeleteALBLoadBalancer             = "DeleteALBLoadBalancer"
	ListALBLoadBalancers              = "ListALBLoadBalancers"
	GetALBLoadBalancerAttribute       = "GetALBLoadBalancerAttribute"
	UpdateALBLoadBalancerAttribute    = "UpdateALBLoadBalancerAttribute"
	UpdateALBLoadBalancerEdition      = "UpdateALBLoadBalancerEdition"
	EnableALBLoadBalancerAccessLog    = "EnableALBLoadBalancerAccessLog"
	DisableALBLoadBalancerAccessLog   = "DisableALBLoadBalancerAccessLog"
	EnableALBDeletionProtection       = "EnableALBDeletionProtection"
	DisableALBDeletionProtection      = "DisableALBDeletionProtection"

	TagALBResource             = "TagALBResource"
	UnTagALBResource           = "UnTagALBResource"
	AnalyzeProductLog          = "AnalyzeProductLog"
	OpenProductDataCollection  = "OpenProductDataCollection"
	CloseProductDataCollection = "CloseProductDataCollection"

	CreateALBRule           = "CreateALBRule"
	CreateALBRules          = "CreateALBRules"
	DeleteALBRule           = "DeleteALBRule"
	DeleteALBRules          = "DeleteALBRules"
	ListALBRules            = "ListALBRules"
	UpdateALBRuleAttribute  = "UpdateALBRuleAttribute"
	UpdateALBRulesAttribute = "UpdateALBRulesAttribute"

	CreateALBServerGroup                        = "CreateALBServerGroup"
	DeleteALBServerGroup                        = "DeleteALBServerGroup"
	UpdateALBServerGroupAttribute               = "UpdateALBServerGroupAttribute"
	ListALBServerGroups                         = "ListALBServerGroups"
	ListALBServerGroupServers                   = "ListALBServerGroupServers"
	AddALBServersToServerGroupAsynchronous      = "AddALBServersToServerGroupAsynchronous"
	AddALBServersToServerGroup                  = "AddALBServersToServerGroup"
	RemoveALBServersFromServerGroupAsynchronous = "RemoveALBServersFromServerGroupAsynchronous"
	RemoveALBServersFromServerGroup             = "RemoveALBServersFromServerGroup"
	ReplaceALBServersInServerGroupAsynchronous  = "ReplaceALBServersInServerGroupAsynchronous"
	ReplaceALBServersInServerGroup              = "ReplaceALBServersInServerGroup"

	ALBInnerServiceManagedControl = "InnerServiceManagedControl"

	CreateAcl                  = "CreateAcl"
	DeleteAcl                  = "DeleteAcl"
	ListAcl                    = "ListAcl"
	AddEntriesToAclALBAcl      = "AddEntriesToAclALBAcl"
	AssociateAclsWithListener  = "AssociateAclsWithListener"
	ListAclRelations           = "ListAclRelations"
	ListAclEntries             = "ListAclEntries"
	RemoveEntriesFromAcl       = "RemoveEntriesFromAcl"
	DissociateAclsFromListener = "DissociateAclsFromListener"
)
const (
	// IngressClass
	IngressClass   = "kubernetes.io/ingress.class"
	KnativeIngress = "knative.aliyun.com/ingress"

	// Ingress annotation suffixes
	IngressSuffixAlbConfigName  = "albconfig.name"
	IngressSuffixAlbConfigOrder = "alb.ingress.kubernetes.io/albconfig.order"
	IngressSuffixListenPorts    = "listen-ports"
)

const (
	MgrLogLevel = ApplierManagerLogLevel
)

const (
	BatchRegisterDeregisterServersMaxNum = 40
	BatchRegisterServersDefaultNum       = 30
	BatchDeregisterServersDefaultNum     = 30

	ServerStatusAdding      = "Adding"
	ServerStatusAvailable   = "Available"
	ServerStatusConfiguring = "Configuring"
	ServerStatusRemoving    = "Removing"

	LoadBalancerStatusActive       = "Active"
	LoadBalancerStatusInactive     = "Inactive"
	LoadBalancerStatusProvisioning = "Provisioning"
	LoadBalancerStatusConfiguring  = "Configuring"
	LoadBalancerStatusCreateFailed = "CreateFailed"

	ListenerStatusProvisioning = "Provisioning"
	ListenerStatusRunning      = "Running"
	ListenerStatusConfiguring  = "Configuring"
	ListenerStatusStopped      = "Stopped"

	AclStatusAvailable = "Available"
)

const (
	BatchCreateDeleteUpdateRulesMaxNum = 10
	BatchCreateRulesDefaultNum         = 9
	BatchDeleteRulesDefaultNum         = 9
	BatchUpdateRulesDefaultNum         = 9
)

const (
	CreateLoadBalancerWaitActiveMaxRetryTimes = 10
	CreateLoadBalancerWaitActiveRetryInterval = 2 * time.Second

	CreateListenerWaitRunningMaxRetryTimes = 15
	CreateListenerWaitRunningRetryInterval = 1 * time.Second
)

const (
	DefaultLogAcceptFormat = "xml"
	DefaultLogCloudProduct = "k8s-nginx-ingress"
	DefaultLogLang         = "cn"
	DefaultLogDomainSuffix = "-intranet.log.aliyuncs.com/open-api"

	MinLogProjectNameLen = 4
	MaxLogProjectNameLen = 63
	MinLogStoreNameLen   = 2
	MaxLogStoreNameLen   = 64
)

const (
	ServerGroupResourceType  = "ServerGroup"
	LoadBalancerResourceType = "LoadBalancer"
)

const (
	AclTypeBlack                  = "Black"
	AclTypeWhite                  = "White"
	BatchAddEntriesToAclMaxNum    = 20
	BatchRemoveEntriesToAclMaxNum = 20
)

const (
	ServerTypeEcs = "Ecs"
	ServerTypeEni = "Eni"
	ServerTypeEci = "Eci"
)

const (
	AddALBServersToServerGroupWaitAvailableMaxRetryTimes = 60
	AddALBServersToServerGroupWaitAvailableRetryInterval = time.Second
	RemoveALBServersFromServerGroupMaxRetryTimes         = 60
	RemoveALBServersFromServerGroupRetryInterval         = time.Second
	ReplaceALBServersInServerGroupMaxRetryTimes          = 60
	ReplaceALBServersInServerGroupRetryInterval          = time.Second
)

const IsWaitServersAsynchronousComplete = true

const (
	SynLogLevel = ApplierSynthesizerLogLevel
)

const (
	BatchReplaceServersMaxNum = 40
)

const (
	TrafficPolicyEni     = "eni"
	TrafficPolicyLocal   = "local"
	TrafficPolicyCluster = "cluster"
)

const (
	ServerGroupConcurrentNum    = 5
	ListenerConcurrentNum       = 5
	SSLCertificateConcurrentNum = 5
)

const (
	LabelNodeRoleMaster = "node-role.kubernetes.io/master"

	LabelNodeExcludeApplicationLoadBalancer = "alpha.service-controller.kubernetes.io/exclude-alb"

	LabelNodeTypeVK = "virtual-kubelet"

	DefaultServerWeight = 100
)
const ConcurrentMaxSleepMillisecondTime = 200

const IndexKeyServiceRefName = "spec.serviceRef.name"

const (
	ClusterTagKey          = "ack.aliyun.com"
	ClusterNameTagKey      = ClusterTagKey
	ServiceNamespaceTagKey = IngressTagKeyPrefix + "/service_ns"
	ServiceNameTagKey      = IngressTagKeyPrefix + "/service_name"
	ServicePortTagKey      = IngressTagKeyPrefix + "/service_port"
	IngressNameTagKey      = IngressTagKeyPrefix + "/ingress_name"

	AlbConfigTagKey     = "albconfig"
	AlbConfigFullTagKey = IngressTagKeyPrefix + "/" + AlbConfigTagKey
)

const (
	IngressClassALB = "alb"
)

const (
	IngressFinalizer = IngressTagKeyPrefix + "/resources"
)

const (
	IngressTagKeyPrefix = "ingress.k8s.alibaba"
)

const (
	DefaultListenerFlag       = "-listener-"
	ListenerDescriptionPrefix = "ingress-auto-listener"
)

const (
	RuleActionTypeFixedResponse string = "FixedResponse"
	RuleActionTypeRedirect      string = "Redirect"
	RuleActionTypeForward       string = "ForwardGroup"
	RuleActionTypeRewrite       string = "Rewrite"
	RuleActionTypeInsertHeader  string = "InsertHeader"
	RuleActionTypeRemoveHeader  string = "RemoveHeader"
	RuleActionTypeTrafficLimit  string = "TrafficLimit"
	RuleActionTypeTrafficMirror string = "TrafficMirror"
	RuleActionTypeCors          string = "Cors"
)

const (
	DefaultCorsAllowOrigin      string = "*"
	DefaultCorsAllowMethods     string = "GET, PUT, POST, DELETE, PATCH, OPTIONS"
	DefaultCorsAllowHeaders     string = "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization"
	DefaultCorsExposeHeaders    string = ""
	DefaultCorsAllowCredentials string = "on"
	DefaultCorsMaxAge           string = "172800"
)

const (
	RuleConditionFieldHost        string = "Host"
	RuleConditionFieldPath        string = "Path"
	RuleConditionFieldHeader      string = "Header"
	RuleConditionFieldQueryString string = "QueryString"
	RuleConditionFieldMethod      string = "Method"
	RuleConditionFieldCookie      string = "Cookie"
	RuleConditionFieldSourceIp    string = "SourceIp"

	// response rule condition
	RuleConditionResponseStatusCode string = "ResponseStatusCode"
	RuleConditionResponseHeader     string = "ResponseHeader"
)

const (
	RuleRequestDirection  string = "Request"
	RuleResponseDirection string = "Response"
)

const (
	DefaultServerGroupScheduler string = ServerGroupSchedulerWrr
	DefaultServerGroupProtocol  string = ServerGroupProtocolHTTP
	DefaultServerGroupType      string = "instance"

	DefaultServerGroupHealthCheckInterval            = 2                                   // 1~50
	DefaultServerGroupHealthyThreshold               = 3                                   // 2～10
	DefaultServerGroupHealthCheckHost                = "$SERVER_IP"                        // GET OR HEAD
	DefaultServerGroupHealthCheckPath                = "/"                                 // GET OR HEAD
	DefaultServerGroupHealthCheckHttpVersion         = ServerGroupHealthCheckHttpVersion11 // HTTP1.0 OR HTTP1.1
	DefaultServerGroupHealthCheckEnabled             = false
	DefaultServerGroupHealthCheckTimeout             = 5 // 1～300
	DefaultServerGroupHealthCheckTcpFastCloseEnabled = false
	DefaultServerGroupHealthCheckConnectPort         = 0                                  // 0~65535
	DefaultServerGroupHealthCheckHTTPCodes           = ServerGroupHealthCheckCodes2xx     // http_2xx、http_3xx、http_4xx OR http_5xx
	DefaultServerGroupHealthCheckCodes               = ServerGroupHealthCheckCodes2xx     // http_2xx、http_3xx、http_4xx OR http_5xx
	DefaultServerGroupHealthCheckMethod              = ServerGroupHealthCheckMethodHEAD   // GET OR HEAD
	DefaultServerGroupUnhealthyThreshold             = 3                                  // 2～10
	DefaultServerGroupHealthCheckProtocol            = ServerGroupHealthCheckProtocolHTTP // HTTP、HTTPS

	// Cookie timeout period. Unit: second
	// Value: 1~86400
	// Default value: 1000
	// Description: When StickySessionEnabled is true and StickySessionType is Insert, this parameter is mandatory.
	DefaultServerGroupStickySessionCookieTimeout = 1000
	// Whether to enable session retention, value: true or false (default value).
	DefaultServerGroupStickySessionEnabled = false
	//Cookie processing method. Value:
	//Insert (default value): Insert Cookie.
	//When the client visits for the first time, the load balancer will implant a cookie in the return request (that is, insert the SERVERID in the HTTP or HTTPS response message), the next time the client visits with this cookie, the load balance service will direct the request Forward to the back-end server previously recorded.
	//Server: Rewrite Cookie.
	//The load balancer finds that the user has customized the cookie and will rewrite the original cookie. The next time the client visits with a new cookie, the load balancer service will direct the request to the back-end server that was previously recorded.
	DefaultServerGroupStickySessionType = ServerGroupStickySessionTypeInsert

	DefaultLoadBalancerAddressType                        string = LoadBalancerAddressTypeInternet
	DefaultLoadBalancerAddressAllocatedMode               string = LoadBalancerAddressAllocatedModeDynamic
	DefaultLoadBalancerEdition                            string = LoadBalancerEditionBasic
	DefaultLoadBalancerBillingConfigPayType               string = LoadBalancerPayTypePostPay
	DefaultLoadBalancerDeletionProtectionConfigEnabled    bool   = true
	DefaultLoadBalancerModificationProtectionConfigStatus string = LoadBalancerModificationProtectionStatusConsoleProtection

	DefaultListenerProtocol         = ListenerProtocolHTTP
	DefaultListenerPort             = 80
	DefaultListenerIdleTimeout      = 15
	DefaultListenerRequestTimeout   = 60
	DefaultListenerGzipEnabled      = true
	DefaultListenerHttp2Enabled     = true
	DefaultListenerSecurityPolicyId = "tls_cipher_policy_1_0"

	ActionTrafficLimitQpsMax = 100000
	ActionTrafficLimitQpsMin = 1
)
const IsReuseLb string = "is_reuse_lb"
const (
	ServerGroupSchedulerWrr = "Wrr"
	ServerGroupSchedulerWlc = "Wlc"
	ServerGroupSchedulerSch = "Sch"

	ServerGroupProtocolHTTP  = "HTTP"
	ServerGroupProtocolHTTPS = "HTTPS"
	ServerGroupProtocolGRPC  = "GRPC"

	ServerGroupHealthCheckMethodGET     = "GET"
	ServerGroupHealthCheckMethodHEAD    = "HEAD"
	ServerGroupHealthCheckProtocolHTTP  = "HTTP"
	ServerGroupHealthCheckProtocolHTTPS = "HTTPS"
	ServerGroupHealthCheckProtocolTCP   = "TCP"
	ServerGroupHealthCheckCodes2xx      = "http_2xx"
	ServerGroupHealthCheckCodes3xx      = "http_3xx"
	ServerGroupHealthCheckCodes4xx      = "http_4xx"
	ServerGroupHealthCheckCodes5xx      = "http_5xx"
	ServerGroupHealthCheckHttpVersion10 = "HTTP1.0"
	ServerGroupHealthCheckHttpVersion11 = "HTTP1.1"

	ServerGroupStickySessionTypeInsert = "Insert"
	ServerGroupStickySessionTypeServer = "Server"
)

const (
	LoadBalancerEditionBasic    = "Basic"
	LoadBalancerEditionStandard = "Standard"
	LoadBalancerEditionWaf      = "StandardWithWaf"

	LoadBalancerAddressTypeInternet = "Internet"
	LoadBalancerAddressTypeIntranet = "Intranet"

	LoadBalancerPayTypePostPay = "PostPay"

	LoadBalancerAddressAllocatedModeFixed   = "Fixed"
	LoadBalancerAddressAllocatedModeDynamic = "Dynamic"

	LoadBalancerModificationProtectionStatusNonProtection     = "NonProtection"
	LoadBalancerModificationProtectionStatusConsoleProtection = "ConsoleProtection"
)

const (
	ListenerProtocolHTTP  = "HTTP"
	ListenerProtocolHTTPS = "HTTPS"
	ListenerProtocolQUIC  = "QUIC"
)

type ContextTraceID string

const (
	TraceID = ContextTraceID("traceID")
)

const (
	ApplierManagerLogLevel     = 0
	ApplierSynthesizerLogLevel = 0
)

var RandomSleepFunc = func(max int) {
	if max <= 0 {
		max = ConcurrentMaxSleepMillisecondTime
	}
	time.Sleep(time.Duration(rand.Intn(max)) * time.Millisecond)
}
