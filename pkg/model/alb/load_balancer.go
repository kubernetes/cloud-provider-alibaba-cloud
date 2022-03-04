package alb

import (
	"context"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"

	"github.com/pkg/errors"
)

var _ core.Resource = &AlbLoadBalancer{}

type AlbLoadBalancer struct {
	core.ResourceMeta `json:"-"`

	Spec ALBLoadBalancerSpec `json:"spec"`

	Status *LoadBalancerStatus `json:"status,omitempty"`
}

func NewAlbLoadBalancer(stack core.Manager, id string, spec ALBLoadBalancerSpec) *AlbLoadBalancer {
	lb := &AlbLoadBalancer{
		ResourceMeta: core.NewResourceMeta(stack, "ALIYUN::ALB::LOADBALANCER", id),
		Spec:         spec,
		Status:       nil,
	}
	_ = stack.AddResource(lb)
	return lb
}

func (lb *AlbLoadBalancer) SetStatus(status LoadBalancerStatus) {
	lb.Status = &status
}

func (lb *AlbLoadBalancer) LoadBalancerID() core.StringToken {
	return core.NewResourceFieldStringToken(lb, "status/loadBalancerID",
		func(ctx context.Context, res core.Resource, fieldPath string) (s string, err error) {
			lb := res.(*AlbLoadBalancer)
			if lb.Status == nil {
				return "", errors.Errorf("LoadBalancer is not fulfilled yet: %v", lb.ID())
			}
			return lb.Status.LoadBalancerID, nil
		},
	)
}

func (lb *AlbLoadBalancer) DNSName() core.StringToken {
	return core.NewResourceFieldStringToken(lb, "status/dnsName",
		func(ctx context.Context, res core.Resource, fieldPath string) (s string, err error) {
			lb := res.(*AlbLoadBalancer)
			if lb.Status == nil {
				return "", errors.Errorf("LoadBalancer is not fulfilled yet: %v", lb.ID())
			}
			return lb.Status.DNSName, nil
		},
	)
}

type LoadBalancerStatus struct {
	DNSName        string `json:"dnsName"`
	LoadBalancerID string `json:"loadBalancerID"`
}
