package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Gateway is a collection of rules that allow inbound connections to reach the
// endpoints defined by a backend. An Gateway can be configured to give services
// externally-reachable urls, load balance traffic, terminate SSL, offer name
// based virtual hosting etc.
type Gateway struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec is the desired state of the Gateway.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec GatewaySpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status is the current state of the Gateway.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status IngressStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayList is a collection of Gateway.
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Gateway.
	Items []Gateway `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// GatewaySpec describes the Gateway the user wishes to exist.
type GatewaySpec struct {
	// IngressClassName is the name of the IngressClass cluster resource. The
	// associated IngressClass defines which controller will implement the
	// resource. This replaces the deprecated `kubernetes.io/ingress.class`
	// annotation. For backwards compatibility, when that annotation is set, it
	// must be given precedence over this field. The controller may emit a
	// warning if the field and annotation have different values.
	// Implementations of this API should ignore Ingresses without a class
	// specified. An IngressClass resource may be marked as default, which can
	// be used to set a default value for this field. For more information,
	// refer to the IngressClass documentation.
	// +optional
	IngressClassName *string `json:"ingressClassName,omitempty" protobuf:"bytes,4,opt,name=ingressClassName"`
}
