package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func init() {
	SchemeBuilder.Register(&AlbConfig{}, &AlbConfigList{})
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlbConfig is a collection of rules that allow inbound connections to reach the
// endpoints defined by a backend. An AlbConfig can be configured to give services
// externally-reachable urls, load balance traffic, terminate SSL, offer name
// based virtual hosting etc.
type AlbConfig struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec is the desired state of the Gateway.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec AlbConfigSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status is the current state of the Gateway.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status IngressStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// LoadBalancerStatus represents the status of a load-balancer.
type LoadBalancerStatus struct {
	DNSName   string           `json:"dnsname,omitempty" protobuf:"bytes,1,opt,name=dnsname"`
	Id        string           `json:"id,omitempty" protobuf:"bytes,2,opt,name=id"`
	Listeners []ListenerStatus `json:"listeners,omitempty" protobuf:"bytes,3,opt,name=listeners"`
}

type ListenerStatus struct {
	PortAndProtocol string               `json:"portAndProtocol,omitempty" protobuf:"bytes,1,opt,name=portAndProtocol"`
	Certificates    []AppliedCertificate `json:"certificates,omitempty" protobuf:"bytes,2,opt,name=certificates"`
}

type AppliedCertificate struct {
	CertificateId string `json:"certificateId,omitempty" protobuf:"bytes,1,opt,name=certificateId"`
	IsDefault     bool   `json:"isDefault,omitempty" protobuf:"bytes,2,opt,name=isDefault"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlbConfigList is a collection of AlbConfig.
type AlbConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Gateway.
	Items []AlbConfig `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// AlbConfigSpec describes the AlbConfig the user wishes to exist.
type AlbConfigSpec struct {
	LoadBalancer *LoadBalancerSpec `json:"config" protobuf:"bytes,1,rep,name=config"`
	Listeners    []*ListenerSpec   `json:"listeners" protobuf:"bytes,2,rep,name=listeners"`
}

// IngressStatus describe the current state of the AckIngress.
type IngressStatus struct {
	// LoadBalancer contains the current status of the load-balancer.
	// +optional
	LoadBalancer LoadBalancerStatus `json:"loadBalancer,omitempty" protobuf:"bytes,1,opt,name=loadBalancer"`
}

// LoadBalancer is a nested struct in alb response
type LoadBalancerSpec struct {
	Id                           string                       `json:"id" protobuf:"bytes,1,opt,name=id"`
	Name                         string                       `json:"name" protobuf:"bytes,2,opt,name=name"`
	AddressAllocatedMode         string                       `json:"addressAllocatedMode" protobuf:"bytes,3,opt,name=addressAllocatedMode"`
	AddressType                  string                       `json:"addressType" protobuf:"bytes,4,opt,name=addressType"`
	ResourceGroupId              string                       `json:"resourceGroupId" protobuf:"bytes,5,opt,name=resourceGroupId"`
	Edition                      string                       `json:"edition" protobuf:"bytes,6,opt,name=edition"`
	ZoneMappings                 []ZoneMapping                `json:"zoneMappings" protobuf:"bytes,7,rep,name=zoneMappings"`
	AccessLogConfig              AccessLogConfig              `json:"accessLogConfig" protobuf:"bytes,8,opt,name=accessLogConfig"`
	DeletionProtectionEnabled    *bool                        `json:"deletionProtectionEnabled" protobuf:"bytes,9,opt,name=deletionProtectionEnabled"`
	BillingConfig                BillingConfig                `json:"billingConfig" protobuf:"bytes,10,opt,name=billingConfig"`
	ForceOverride                *bool                        `json:"forceOverride" protobuf:"bytes,11,opt,name=forceOverride"`
	ModificationProtectionConfig ModificationProtectionConfig `json:"modificationProtectionConfig" protobuf:"bytes,12,opt,name=modificationProtectionConfig"`
	Tags                         []Tag                        `json:"tags" protobuf:"bytes,13,opt,name=tags"`
	ListenerForceOverride        *bool                        `json:"listenerForceOverride" protobuf:"bytes,14,opt,name=listenerForceOverride"`
}

type Tag struct {
	Key   string `json:"key" protobuf:"bytes,1,opt,name=key"`
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

// ZoneMapping is a nested struct in alb response
type ZoneMapping struct {
	VSwitchId string `json:"vSwitchId" protobuf:"bytes,1,opt,name=vSwitchId"`
	ZoneId    string `json:"zoneId" protobuf:"bytes,2,opt,name=zoneId"`
}

type AccessLogConfig struct {
	LogStore   string `json:"logStore" protobuf:"bytes,1,opt,name=logStore"`
	LogProject string `json:"logProject" protobuf:"bytes,2,opt,name=logProject"`
}
type DeletionProtectionConfig struct {
	Enabled     bool   `json:"enabled" protobuf:"bytes,1,opt,name=enabled"`
	EnabledTime string `json:"enabledTime" protobuf:"bytes,2,opt,name=enabledTime"`
}
type BillingConfig struct {
	InternetBandwidth  int    `json:"internetBandwidth" protobuf:"bytes,1,opt,name=internetBandwidth"`
	InternetChargeType string `json:"internetChargeType" protobuf:"bytes,2,opt,name=internetChargeType"`
	PayType            string `json:"payType" protobuf:"bytes,3,opt,name=payType"`
}
type ModificationProtectionConfig struct {
	Reason string `json:"reason" protobuf:"bytes,1,opt,name=reason"`
	Status string `json:"status" protobuf:"bytes,2,opt,name=status"`
}

type Certificate struct {
	IsDefault     bool   `json:"IsDefault" protobuf:"bytes,1,opt,name=IsDefault"`
	CertificateId string `json:"CertificateId" protobuf:"bytes,2,opt,name=CertificateId"`
}

type QuicConfig struct {
	QuicUpgradeEnabled bool   `json:"quicUpgradeEnabled" protobuf:"bytes,1,opt,name=quicUpgradeEnabled"`
	QuicListenerId     string `json:"quicListenerId" protobuf:"bytes,2,opt,name=quicListenerId"`
}

type XForwardedForConfig struct {
	XForwardedForClientCertSubjectDNAlias      string `json:"XForwardedForClientCertSubjectDNAlias" protobuf:"bytes,1,opt,name=XForwardedForClientCertSubjectDNAlias"`
	XForwardedForClientCertSubjectDNEnabled    bool   `json:"XForwardedForClientCertSubjectDNEnabled" protobuf:"bytes,2,opt,name=XForwardedForClientCertSubjectDNEnabled"`
	XForwardedForProtoEnabled                  bool   `json:"XForwardedForProtoEnabled" protobuf:"bytes,3,opt,name=XForwardedForProtoEnabled"`
	XForwardedForClientCertIssuerDNEnabled     bool   `json:"XForwardedForClientCertIssuerDNEnabled" protobuf:"bytes,4,opt,name=XForwardedForClientCertIssuerDNEnabled"`
	XForwardedForSLBIdEnabled                  bool   `json:"XForwardedForSLBIdEnabled" protobuf:"bytes,5,opt,name=XForwardedForSLBIdEnabled"`
	XForwardedForClientSrcPortEnabled          bool   `json:"XForwardedForClientSrcPortEnabled" protobuf:"bytes,6,opt,name=XForwardedForClientSrcPortEnabled"`
	XForwardedForClientCertFingerprintEnabled  bool   `json:"XForwardedForClientCertFingerprintEnabled" protobuf:"bytes,7,opt,name=XForwardedForClientCertFingerprintEnabled"`
	XForwardedForEnabled                       bool   `json:"XForwardedForEnabled" protobuf:"bytes,8,opt,name=XForwardedForEnabled"`
	XForwardedForSLBPortEnabled                bool   `json:"XForwardedForSLBPortEnabled" protobuf:"bytes,9,opt,name=XForwardedForSLBPortEnabled"`
	XForwardedForClientCertClientVerifyAlias   string `json:"XForwardedForClientCertClientVerifyAlias" protobuf:"bytes,10,opt,name=XForwardedForClientCertClientVerifyAlias"`
	XForwardedForClientCertIssuerDNAlias       string `json:"XForwardedForClientCertIssuerDNAlias" protobuf:"bytes,11,opt,name=XForwardedForClientCertIssuerDNAlias"`
	XForwardedForClientCertFingerprintAlias    string `json:"XForwardedForClientCertFingerprintAlias" protobuf:"bytes,12,opt,name=XForwardedForClientCertFingerprintAlias"`
	XForwardedForClientCertClientVerifyEnabled bool   `json:"XForwardedForClientCertClientVerifyEnabled" protobuf:"bytes,13,opt,name=XForwardedForClientCertClientVerifyEnabled"`
}

type AccessLogTracingConfig struct {
	TracingSample  int    `json:"tracingSample" protobuf:"bytes,1,opt,name=tracingSample"`
	TracingType    string `json:"tracingType" protobuf:"bytes,2,opt,name=tracingType"`
	TracingEnabled bool   `json:"tracingEnabled" protobuf:"bytes,3,opt,name=tracingEnabled"`
}

type LogConfig struct {
	AccessLogRecordCustomizedHeadersEnabled bool                   `json:"accessLogRecordCustomizedHeadersEnabled" protobuf:"bytes,1,opt,name=accessLogRecordCustomizedHeadersEnabled"`
	AccessLogTracingConfig                  AccessLogTracingConfig `json:"accessLogTracingConfig" protobuf:"bytes,2,opt,name=accessLogTracingConfig"`
}

type ListenerSpec struct {
	GzipEnabled         *bool               `json:"gzipEnabled" protobuf:"bytes,1,opt,name=gzipEnabled"`
	QuicConfig          QuicConfig          `json:"quicConfig" protobuf:"bytes,2,opt,name=quicConfig"`
	Http2Enabled        *bool               `json:"http2Enabled" protobuf:"bytes,3,opt,name=http2Enabled"`
	DefaultActions      []Action            `json:"defaultActions" protobuf:"bytes,4,opt,name=defaultActions"`
	Port                intstr.IntOrString  `json:"port" protobuf:"bytes,5,opt,name=port"`
	CaCertificates      []Certificate       `json:"caCertificates" protobuf:"bytes,6,opt,name=caCertificates"`
	XForwardedForConfig XForwardedForConfig `json:"xForwardedForConfig" protobuf:"bytes,7,opt,name=xForwardedForConfig"`
	Protocol            string              `json:"protocol" protobuf:"bytes,8,opt,name=protocol"`
	SecurityPolicyId    string              `json:"securityPolicyId" protobuf:"bytes,9,opt,name=securityPolicyId"`
	IdleTimeout         int                 `json:"idleTimeout" protobuf:"bytes,10,opt,name=idleTimeout"`
	LoadBalancerId      string              `json:"loadBalancerId" protobuf:"bytes,11,opt,name=loadBalancerId"`
	Certificates        []Certificate       `json:"certificates" protobuf:"bytes,12,opt,name=certificates"`
	Description         string              `json:"description" protobuf:"bytes,13,opt,name=description"`
	CaEnabled           bool                `json:"caEnabled" protobuf:"bytes,14,opt,name=caEnabled"`
	LogConfig           LogConfig           `json:"logConfig" protobuf:"bytes,15,opt,name=logConfig"`
	RequestTimeout      int                 `json:"requestTimeout" protobuf:"bytes,16,opt,name=requestTimeout"`
	AclConfig           AclConfig           `json:"aclConfig" protobuf:"bytes,17,opt,name=aclConfig"`
}
type Action struct {
	Type string `json:"actionType" protobuf:"bytes,1,opt,name=actionType"`

	FixedResponseConfig *FixedResponseActionConfig `json:"fixedResponseConfig,omitempty" protobuf:"bytes,2,opt,name=fixedResponseConfig"`

	RedirectConfig *RedirectActionConfig `json:"redirectConfig,omitempty" protobuf:"bytes,3,opt,name=redirectConfig"`

	ForwardConfig *ForwardActionConfig `json:"forwardConfig,omitempty" protobuf:"bytes,4,opt,name=forwardConfig"`
}

type TargetGroupTuple struct {
	TargetGroupARN string `json:"targetGroupARN" protobuf:"bytes,1,opt,name=targetGroupARN"`

	ServiceName string `json:"serviceName" protobuf:"bytes,2,opt,name=serviceName"`

	ServicePort intstr.IntOrString `json:"servicePort" protobuf:"bytes,3,opt,name=servicePort"`

	Weight int `json:"weight,omitempty" protobuf:"bytes,4,opt,name=weight"`
}

type ForwardActionConfig struct {
	TargetGroups []TargetGroupTuple `json:"targetGroups" protobuf:"bytes,1,opt,name=targetGroups"`
}
type FixedResponseActionConfig struct {
	ContentType string `json:"contentType,omitempty" protobuf:"bytes,1,opt,name=contentType"`

	MessageBody string `json:"messageBody,omitempty" protobuf:"bytes,2,opt,name=messageBody"`

	StatusCode string `json:"statusCode" protobuf:"bytes,3,opt,name=statusCode"`
}

type RedirectActionConfig struct {
	Host string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`

	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	Port string `json:"port,omitempty" protobuf:"bytes,3,opt,name=port"`

	Protocol string `json:"protocol,omitempty" protobuf:"bytes,4,opt,name=protocol"`

	Query string `json:"query,omitempty" protobuf:"bytes,5,opt,name=query"`

	StatusCode string `json:"statusCode" protobuf:"bytes,6,opt,name=statusCode"`
}

type AclConfig struct {
	AclName    string   `json:"aclName" protobuf:"bytes,1,opt,name=aclName"`
	AclType    string   `json:"aclType" protobuf:"bytes,2,opt,name=aclType"`
	AclEntries []string `json:"aclEntries" protobuf:"bytes,3,opt,name=aclEntries"`
}
