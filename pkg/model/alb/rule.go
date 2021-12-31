package alb

import (
	albsdk "github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"
)

var _ core.Resource = &ListenerRule{}

type ListenerRule struct {
	core.ResourceMeta `json:"-"`

	Spec ListenerRuleSpec `json:"spec"`

	Status *ListenerRuleStatus `json:"status,omitempty"`
}

func NewListenerRule(stack core.Manager, id string, spec ListenerRuleSpec) *ListenerRule {
	lr := &ListenerRule{
		ResourceMeta: core.NewResourceMeta(stack, "ALIYUN::ALB::RULE", id),
		Spec:         spec,
		Status:       nil,
	}
	stack.AddResource(lr)
	lr.registerDependencies(stack)
	return lr
}

func (lr *ListenerRule) SetStatus(status ListenerRuleStatus) {
	lr.Status = &status
}

func (lr *ListenerRule) registerDependencies(stack core.Manager) {
	for _, dep := range lr.Spec.ListenerID.Dependencies() {
		stack.AddDependency(dep, lr)
	}
}

type ListenerRuleStatus struct {
	RuleID string `json:"ruleID"`
}

type ListenerRuleSpec struct {
	ListenerID core.StringToken `json:"listenerID"`
	ALBListenerRuleSpec
}

type ResAndSDKListenerRulePair struct {
	ResLR *ListenerRule
	SdkLR *albsdk.Rule
}
