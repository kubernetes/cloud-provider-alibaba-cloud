package alibaba

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
)

func NewLBProvider(
	auth *metadata.ClientAuth,
) *ProviderSLB {
	return &ProviderSLB{auth: auth}
}

type ProviderSLB struct {
	auth *metadata.ClientAuth
}

func (ProviderSLB) FindSLB(ctx context.Context, id string) ([]model.LoadBalancer, error) {
	panic("implement me")
}

func (ProviderSLB) ListSLB(ctx context.Context, slb model.LoadBalancer) ([]model.LoadBalancer, error) {
	panic("implement me")
}

func (p ProviderSLB) CreateSLB(ctx context.Context, balancer *model.LoadBalancer) (*model.LoadBalancer, error) {
	request := slb.CreateCreateLoadBalancerRequest()
	request.AddressType = *balancer.LoadBalancerAttribute.AddressType
	request.LoadBalancerSpec = *balancer.LoadBalancerAttribute.LoadBalancerSpec
	resp, err := p.auth.SLB.CreateLoadBalancer(request)
	request.ClientToken = utils.GetUUID()
	if err != nil {
		return nil, err
	}
	balancer.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	return balancer, nil
}

func (ProviderSLB) DeleteSLB(ctx context.Context, slb model.LoadBalancer) error {
	panic("implement me")
}
