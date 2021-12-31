package alb

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"

	"github.com/pkg/errors"
)

var _ core.Resource = &Listener{}

type Listener struct {
	core.ResourceMeta `json:"-"`

	Spec ListenerSpec `json:"spec"`

	Status *ListenerStatus `json:"status,omitempty"`
}

func NewListener(stack core.Manager, id string, spec ListenerSpec) *Listener {
	ls := &Listener{
		ResourceMeta: core.NewResourceMeta(stack, "ALIYUN::ALB::LISTENER", id),
		Spec:         spec,
		Status:       nil,
	}
	stack.AddResource(ls)
	ls.registerDependencies(stack)
	return ls
}

func (ls *Listener) SetStatus(status ListenerStatus) {
	ls.Status = &status
}

func (ls *Listener) ListenerID() core.StringToken {
	return core.NewResourceFieldStringToken(ls, "status/listenerID",
		func(ctx context.Context, res core.Resource, fieldPath string) (s string, err error) {
			ls := res.(*Listener)
			if ls.Status == nil {
				return "", errors.Errorf("Listener is not fulfilled yet: %v", ls.ID())
			}
			return ls.Status.ListenerID, nil
		},
	)
}

func (ls *Listener) registerDependencies(stack core.Manager) {
	for _, dep := range ls.Spec.LoadBalancerID.Dependencies() {
		stack.AddDependency(dep, ls)
	}
}

type ListenerSpec struct {
	LoadBalancerID core.StringToken `json:"loadBalancerID"`
	ALBListenerSpec
}
type ListenerStatus struct {
	ListenerID string `json:"listenerID"`
}
