package alb

import "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"

type ALBLoadBalancerSpec struct {
	AddressAllocatedMode         string                       `json:"AddressAllocatedMode" xml:"AddressAllocatedMode"`
	AddressType                  string                       `json:"AddressType" xml:"AddressType"`
	DNSName                      string                       `json:"DNSName" xml:"DNSName"`
	LoadBalancerEdition          string                       `json:"LoadBalancerEdition" xml:"LoadBalancerEdition"`
	LoadBalancerId               string                       `json:"LoadBalancerId" xml:"LoadBalancerId"`
	LoadBalancerName             string                       `json:"LoadBalancerName" xml:"LoadBalancerName"`
	LoadBalancerStatus           string                       `json:"LoadBalancerStatus" xml:"LoadBalancerStatus"`
	ResourceGroupId              string                       `json:"ResourceGroupId" xml:"ResourceGroupId"`
	VpcId                        string                       `json:"VpcId" xml:"VpcId"`
	ForceOverride                *bool                        `json:"ForceOverride" xml:"ForceOverride"`
	AccessLogConfig              AccessLogConfig              `json:"AccessLogConfig" xml:"AccessLogConfig"`
	DeletionProtectionConfig     DeletionProtectionConfig     `json:"DeletionProtectionConfig" xml:"DeletionProtectionConfig"`
	LoadBalancerBillingConfig    LoadBalancerBillingConfig    `json:"LoadBalancerBillingConfig" xml:"LoadBalancerBillingConfig"`
	ModificationProtectionConfig ModificationProtectionConfig `json:"ModificationProtectionConfig" xml:"ModificationProtectionConfig"`
	LoadBalancerOperationLocks   []LoadBalancerOperationLock  `json:"LoadBalancerOperationLocks" xml:"LoadBalancerOperationLocks"`
	Tags                         []ALBTag                     `json:"Tags" xml:"Tags"`
	ZoneMapping                  []ZoneMapping                `json:"ZoneMapping" xml:"ZoneMapping"`
}

type ALBListenerSpec struct {
	DefaultActions      []Action            `json:"DefaultActions" xml:"DefaultActions"`
	Certificates        []Certificate       `json:"Certificates" xml:"Certificates"`
	CaCertificates      []Certificate       `json:"CaCertificates" xml:"CaCertificates"`
	GzipEnabled         bool                `json:"GzipEnabled" xml:"GzipEnabled"`
	Http2Enabled        bool                `json:"Http2Enabled" xml:"Http2Enabled"`
	IdleTimeout         int                 `json:"IdleTimeout" xml:"IdleTimeout"`
	ListenerDescription string              `json:"ListenerDescription" xml:"ListenerDescription"`
	ListenerId          string              `json:"ListenerId" xml:"ListenerId"`
	ListenerPort        int                 `json:"ListenerPort" xml:"ListenerPort"`
	ListenerProtocol    string              `json:"ListenerProtocol" xml:"ListenerProtocol"`
	ListenerStatus      string              `json:"ListenerStatus" xml:"ListenerStatus"`
	RequestTimeout      int                 `json:"RequestTimeout" xml:"RequestTimeout"`
	SecurityPolicyId    string              `json:"SecurityPolicyId" xml:"SecurityPolicyId"`
	LogConfig           LogConfig           `json:"LogConfig" xml:"LogConfig"`
	QuicConfig          QuicConfig          `json:"QuicConfig" xml:"QuicConfig"`
	XForwardedForConfig XForwardedForConfig `json:"XForwardedForConfig" xml:"XForwardedForConfig"`
}

type ALBListenerRuleSpec struct {
	Priority       int         `json:"Priority" xml:"Priority"`
	RuleId         string      `json:"RuleId" xml:"RuleId"`
	RuleName       string      `json:"RuleName" xml:"RuleName"`
	RuleStatus     string      `json:"RuleStatus" xml:"RuleStatus"`
	RuleActions    []Action    `json:"RuleActions" xml:"RuleActions"`
	RuleConditions []Condition `json:"RuleConditions" xml:"RuleConditions"`
}

type ALBServerGroupSpec struct {
	Protocol            string              `json:"Protocol" xml:"Protocol"`
	ResourceGroupId     string              `json:"ResourceGroupId" xml:"ResourceGroupId"`
	Scheduler           string              `json:"Scheduler" xml:"Scheduler"`
	ServerGroupId       string              `json:"ServerGroupId" xml:"ServerGroupId"`
	ServerGroupName     string              `json:"ServerGroupName" xml:"ServerGroupName"`
	ServerGroupStatus   string              `json:"ServerGroupStatus" xml:"ServerGroupStatus"`
	ServerGroupType     string              `json:"ServerGroupType" xml:"ServerGroupType"`
	VpcId               string              `json:"VpcId" xml:"VpcId"`
	HealthCheckConfig   HealthCheckConfig   `json:"HealthCheckConfig" xml:"HealthCheckConfig"`
	StickySessionConfig StickySessionConfig `json:"StickySessionConfig" xml:"StickySessionConfig"`
	Tags                []ALBTag            `json:"Tags" xml:"Tags"`
}

type AccessLogConfig struct {
	LogStore   string `json:"LogStore" xml:"LogStore"`
	LogProject string `json:"LogProject" xml:"LogProject"`
}
type DeletionProtectionConfig struct {
	Enabled     bool   `json:"Enabled" xml:"Enabled"`
	EnabledTime string `json:"EnabledTime" xml:"EnabledTime"`
}
type LoadBalancerBillingConfig struct {
	InternetBandwidth  int    `json:"InternetBandwidth" xml:"InternetBandwidth"`
	InternetChargeType string `json:"InternetChargeType" xml:"InternetChargeType"`
	PayType            string `json:"PayType" xml:"PayType"`
}
type ModificationProtectionConfig struct {
	Reason string `json:"Reason" xml:"Reason"`
	Status string `json:"Status" xml:"Status"`
}
type LoadBalancerOperationLock struct {
	LockReason string `json:"LockReason" xml:"LockReason"`
	LockType   string `json:"LockType" xml:"LockType"`
}

type AccessLog struct {
	LogProject string `json:"project"`
	LogStore   string `json:"logstore"`
}

type ZoneMapping struct {
	VSwitchId             string                `json:"VSwitchId" xml:"VSwitchId"`
	ZoneId                string                `json:"ZoneId" xml:"ZoneId"`
	LoadBalancerAddresses []LoadBalancerAddress `json:"LoadBalancerAddresses" xml:"LoadBalancerAddresses"`
}
type LoadBalancerAddress struct {
	Address string `json:"Address" xml:"Address"`
}

type HealthCheckConfig struct {
	HealthCheckConnectPort         int      `json:"HealthCheckConnectPort" xml:"HealthCheckConnectPort"`
	HealthCheckEnabled             bool     `json:"HealthCheckEnabled" xml:"HealthCheckEnabled"`
	HealthCheckHost                string   `json:"HealthCheckHost" xml:"HealthCheckHost"`
	HealthCheckHttpVersion         string   `json:"HealthCheckHttpVersion" xml:"HealthCheckHttpVersion"`
	HealthCheckInterval            int      `json:"HealthCheckInterval" xml:"HealthCheckInterval"`
	HealthCheckMethod              string   `json:"HealthCheckMethod" xml:"HealthCheckMethod"`
	HealthCheckPath                string   `json:"HealthCheckPath" xml:"HealthCheckPath"`
	HealthCheckProtocol            string   `json:"HealthCheckProtocol" xml:"HealthCheckProtocol"`
	HealthCheckTimeout             int      `json:"HealthCheckTimeout" xml:"HealthCheckTimeout"`
	HealthyThreshold               int      `json:"HealthyThreshold" xml:"HealthyThreshold"`
	UnhealthyThreshold             int      `json:"UnhealthyThreshold" xml:"UnhealthyThreshold"`
	HealthCheckTcpFastCloseEnabled bool     `json:"HealthCheckTcpFastCloseEnabled" xml:"HealthCheckTcpFastCloseEnabled"`
	HealthCheckHttpCodes           []string `json:"HealthCheckHttpCodes" xml:"HealthCheckHttpCodes"`
	HealthCheckCodes               []string `json:"HealthCheckCodes" xml:"HealthCheckCodes"`
}
type StickySessionConfig struct {
	Cookie               string `json:"Cookie" xml:"Cookie"`
	CookieTimeout        int    `json:"CookieTimeout" xml:"CookieTimeout"`
	StickySessionEnabled bool   `json:"StickySessionEnabled" xml:"StickySessionEnabled"`
	StickySessionType    string `json:"StickySessionType" xml:"StickySessionType"`
}

type Certificate struct {
	IsDefault     bool   `json:"IsDefault" xml:"IsDefault"`
	CertificateId string `json:"CertificateId" xml:"CertificateId"`
	Status        string `json:"Status" xml:"Status"`
}
type LogConfig struct {
	AccessLogRecordCustomizedHeadersEnabled bool                   `json:"AccessLogRecordCustomizedHeadersEnabled" xml:"AccessLogRecordCustomizedHeadersEnabled"`
	AccessLogTracingConfig                  AccessLogTracingConfig `json:"AccessLogTracingConfig" xml:"AccessLogTracingConfig"`
}
type AccessLogTracingConfig struct {
	TracingSample  int    `json:"TracingSample" xml:"TracingSample"`
	TracingType    string `json:"TracingType" xml:"TracingType"`
	TracingEnabled bool   `json:"TracingEnabled" xml:"TracingEnabled"`
}
type QuicConfig struct {
	QuicUpgradeEnabled bool   `json:"QuicUpgradeEnabled" xml:"QuicUpgradeEnabled"`
	QuicListenerId     string `json:"QuicListenerId" xml:"QuicListenerId"`
}

type XForwardedForConfig struct {
	XForwardedForClientCertSubjectDNAlias      string `json:"XForwardedForClientCertSubjectDNAlias" xml:"XForwardedForClientCertSubjectDNAlias"`
	XForwardedForClientCertSubjectDNEnabled    bool   `json:"XForwardedForClientCertSubjectDNEnabled" xml:"XForwardedForClientCertSubjectDNEnabled"`
	XForwardedForProtoEnabled                  bool   `json:"XForwardedForProtoEnabled" xml:"XForwardedForProtoEnabled"`
	XForwardedForClientCertIssuerDNEnabled     bool   `json:"XForwardedForClientCertIssuerDNEnabled" xml:"XForwardedForClientCertIssuerDNEnabled"`
	XForwardedForSLBIdEnabled                  bool   `json:"XForwardedForSLBIdEnabled" xml:"XForwardedForSLBIdEnabled"`
	XForwardedForClientSrcPortEnabled          bool   `json:"XForwardedForClientSrcPortEnabled" xml:"XForwardedForClientSrcPortEnabled"`
	XForwardedForClientCertFingerprintEnabled  bool   `json:"XForwardedForClientCertFingerprintEnabled" xml:"XForwardedForClientCertFingerprintEnabled"`
	XForwardedForEnabled                       bool   `json:"XForwardedForEnabled" xml:"XForwardedForEnabled"`
	XForwardedForSLBPortEnabled                bool   `json:"XForwardedForSLBPortEnabled" xml:"XForwardedForSLBPortEnabled"`
	XForwardedForClientCertClientVerifyAlias   string `json:"XForwardedForClientCertClientVerifyAlias" xml:"XForwardedForClientCertClientVerifyAlias"`
	XForwardedForClientCertIssuerDNAlias       string `json:"XForwardedForClientCertIssuerDNAlias" xml:"XForwardedForClientCertIssuerDNAlias"`
	XForwardedForClientCertFingerprintAlias    string `json:"XForwardedForClientCertFingerprintAlias" xml:"XForwardedForClientCertFingerprintAlias"`
	XForwardedForClientCertClientVerifyEnabled bool   `json:"XForwardedForClientCertClientVerifyEnabled" xml:"XForwardedForClientCertClientVerifyEnabled"`
}

type Action struct {
	Order               int                  `json:"Order" xml:"Order"`
	Type                string               `json:"Type" xml:"Type"`
	ForwardConfig       *ForwardActionConfig `json:"forwardConfig,omitempty"`
	FixedResponseConfig *FixedResponseConfig `json:"FixedResponseConfig" xml:"FixedResponseConfig"`
	RedirectConfig      *RedirectConfig      `json:"RedirectConfig" xml:"RedirectConfig"`
	InsertHeaderConfig  *InsertHeaderConfig  `json:"InsertHeaderConfig" xml:"InsertHeaderConfig"`
	RemoveHeaderConfig  *RemoveHeaderConfig  `json:"RemoveHeaderConfig" xml:"RemoveHeaderConfig"`
	RewriteConfig       *RewriteConfig       `json:"RewriteConfig" xml:"RewriteConfig"`
	TrafficMirrorConfig *TrafficMirrorConfig `json:"TrafficMirrorConfig" xml:"TrafficMirrorConfig"`
	TrafficLimitConfig  *TrafficLimitConfig  `json:"TrafficLimitConfig" xml:"TrafficLimitConfig"`
}

type ServerGroupTuple struct {
	ServerGroupID core.StringToken `json:"serverGroupID"`

	ServiceName string `json:"serviceName"`

	ServicePort int `json:"servicePort"`

	Weight int `json:"weight,omitempty"`
}
type InsertHeaderConfig struct {
	CoverEnabled bool   `json:"CoverEnabled" xml:"CoverEnabled"`
	Key          string `json:"Key" xml:"Key"`
	Value        string `json:"Value" xml:"Value"`
	ValueType    string `json:"ValueType" xml:"ValueType"`
}
type RemoveHeaderConfig struct {
	Key string `json:"Key" xml:"Key"`
}
type RewriteConfig struct {
	Host  string `json:"Host" xml:"Host"`
	Path  string `json:"Path" xml:"Path"`
	Query string `json:"Query" xml:"Query"`
}
type RedirectConfig struct {
	Host     string `json:"Host" xml:"Host"`
	HttpCode string `json:"HttpCode" xml:"HttpCode"`
	Path     string `json:"Path" xml:"Path"`
	Port     string `json:"Port" xml:"Port"`
	Protocol string `json:"Protocol" xml:"Protocol"`
	Query    string `json:"Query" xml:"Query"`
}

type ForwardActionConfig struct {
	ServerGroupStickySession *ServerGroupStickySession `json:"ServerGroupStickySession" xml:"ServerGroupStickySession"`
	ServerGroups             []ServerGroupTuple        `json:"serverGroups"`
}
type ServerGroupStickySession struct {
	Enabled bool `json:"Enabled" xml:"Enabled"`
	Timeout int  `json:"Timeout" xml:"Timeout"`
}

type TrafficMirrorConfig struct {
	TargetType        string            `json:"TargetType" xml:"TargetType"`
	MirrorGroupConfig MirrorGroupConfig `json:"MirrorGroupConfig" xml:"MirrorGroupConfig"`
}
type MirrorGroupConfig struct {
	ServerGroupTuples []ServerGroupTuple `json:"ServerGroupTuples" xml:"ServerGroupTuples"`
}
type TrafficLimitConfig struct {
	QPS int `json:"QPS" xml:"QPS"`
}
type FixedResponseConfig struct {
	Content     string `json:"Content" xml:"Content"`
	ContentType string `json:"ContentType" xml:"ContentType"`
	HttpCode    string `json:"HttpCode" xml:"HttpCode"`
}

type Condition struct {
	Type                     string                   `json:"Type" xml:"Type"`
	CookieConfig             CookieConfig             `json:"CookieConfig" xml:"CookieConfig"`
	HeaderConfig             HeaderConfig             `json:"HeaderConfig" xml:"HeaderConfig"`
	HostConfig               HostConfig               `json:"HostConfig" xml:"HostConfig"`
	MethodConfig             MethodConfig             `json:"MethodConfig" xml:"MethodConfig"`
	PathConfig               PathConfig               `json:"PathConfig" xml:"PathConfig"`
	QueryStringConfig        QueryStringConfig        `json:"QueryStringConfig" xml:"QueryStringConfig"`
	SourceIpConfig           SourceIpConfig           `json:"SourceIpConfig" xml:"SourceIpConfig"`
	ResponseStatusCodeConfig ResponseStatusCodeConfig `json:"ResponseStatusCodeConfig" xml:"ResponseStatusCodeConfig"`
	ResponseHeaderConfig     ResponseHeaderConfig     `json:"ResponseHeaderConfig" xml:"ResponseHeaderConfig"`
}
type CookieConfig struct {
	Values []Value `json:"Values" xml:"Values"`
}
type HeaderConfig struct {
	Key    string   `json:"Key" xml:"Key"`
	Values []string `json:"Values" xml:"Values"`
}
type Value struct {
	Key   string `json:"Key" xml:"Key"`
	Value string `json:"Value" xml:"Value"`
}
type HostConfig struct {
	Values []string `json:"Values" xml:"Values"`
}
type MethodConfig struct {
	Values []string `json:"Values" xml:"Values"`
}
type PathConfig struct {
	Values []string `json:"Values" xml:"Values"`
}
type QueryStringConfig struct {
	Values []Value `json:"Values" xml:"Values"`
}
type SourceIpConfig struct {
	Values []string `json:"Values" xml:"Values"`
}
type ResponseStatusCodeConfig struct {
	Values []string `json:"Values" xml:"Values"`
}
type ResponseHeaderConfig struct {
	Key    string   `json:"Key" xml:"Key"`
	Values []string `json:"Values" xml:"Values"`
}
type ALBTag struct {
	Key   string `json:"Key" xml:"Key"`
	Value string `json:"Value" xml:"Value"`
}
