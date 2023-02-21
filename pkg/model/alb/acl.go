package alb

import (
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

type Acl struct {
	core.ResourceMeta `json:"-"`

	Spec AclSpec `json:"spec"`

	Status *AclStatus `json:"status,omitempty"`
}

func NewAcl(stack core.Manager, id string, spec AclSpec) *Acl {
	acl := &Acl{
		ResourceMeta: core.NewResourceMeta(stack, "ALIYUN::ALB::ACL", id),
		Spec:         spec,
		Status:       nil,
	}
	stack.AddResource(acl)
	acl.registerDependencies(stack)
	return acl
}

func (acl *Acl) SetStatus(status AclStatus) {
	acl.Status = &status
}

func (acl *Acl) registerDependencies(stack core.Manager) {
	for _, dep := range acl.Spec.ListenerID.Dependencies() {
		stack.AddDependency(dep, acl)
	}
}

type AclStatus struct {
	AclID string `json:"aclID"`
}

type AclSpec struct {
	ListenerID core.StringToken `json:"listenerID"`
	AclId      string           `json:"AclId" xml:"AclId"`
	AclName    string           `json:"AclName" xml:"AclName"`
	AclType    string           `json:"AclType" xml:"AclType"`
	AclStatus  string           `json:"AclStatus" xml:"AclStatus"`
	AclEntries []AclEntry       `json:"AclEntries" xml:"AclEntries"`
}

type AclEntry struct {
	Entry  string `json:"Entry" xml:"Entry"`
	Status string `json:"Status" xml:"Status"`
}
type ResAndSDKAclPair struct {
	ResAcl *Acl
	SdkAcl *albsdk.Acl
}
