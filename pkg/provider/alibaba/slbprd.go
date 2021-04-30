package alibaba

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
)

func NewLBProvider(
	auth *metadata.ClientAuth,
) *ProviderSLB {
	return &ProviderSLB{auth: auth}
}

var _ prvd.ILoadBalancer = &ProviderSLB{}

type ProviderSLB struct {
	auth *metadata.ClientAuth
}

func (p ProviderSLB) FindSLB(ctx context.Context, mdl *model.LoadBalancer) (bool, *model.LoadBalancer, error) {

	// 1. find slb by loadbalancer id
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] try to find slb by loadbalancer id %s", mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		req := slb.CreateDescribeLoadBalancersRequest()
		req.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
		resp, err := p.auth.SLB.DescribeLoadBalancers(req)
		if err != nil {
			return false, nil, err
		}
		if resp != nil && len(resp.LoadBalancers.LoadBalancer) > 0 {
			return getModelFromResponse(p, mdl.NamespacedName, resp)
		}
	}

	// 2. find slb by tag
	items, err := json.Marshal(mdl.LoadBalancerAttribute.Tags)
	if err != nil {
		return false, nil, err
	}
	klog.Infof("[%s] try to find slb by tag %s", mdl.NamespacedName, items)
	req := slb.CreateDescribeLoadBalancersRequest()
	req.Tags = string(items)
	resp, err := p.auth.SLB.DescribeLoadBalancers(req)
	if err != nil {
		return false, nil, err
	}
	if resp != nil && len(resp.LoadBalancers.LoadBalancer) > 0 {
		return getModelFromResponse(p, mdl.NamespacedName, resp)
	}

	// 3. find slb by name
	if mdl.LoadBalancerAttribute.LoadBalancerName != nil {
		klog.Infof("[%s] try to find slb by name %s", mdl.NamespacedName, *mdl.LoadBalancerAttribute.LoadBalancerName)
		req := slb.CreateDescribeLoadBalancersRequest()
		req.LoadBalancerName = *mdl.LoadBalancerAttribute.LoadBalancerName
		resp, err := p.auth.SLB.DescribeLoadBalancers(req)
		if err != nil {
			return false, nil, err
		}
		if resp != nil && len(resp.LoadBalancers.LoadBalancer) > 0 {
			return getModelFromResponse(p, mdl.NamespacedName, resp)
		}
	}

	return false, nil, nil
}

func getModelFromResponse(p ProviderSLB, name types.NamespacedName, response *slb.DescribeLoadBalancersResponse) (bool, *model.LoadBalancer, error) {
	if len(response.LoadBalancers.LoadBalancer) > 1 {
		klog.Warningf("find %d load balances by model, use the first one", len(response.LoadBalancers.LoadBalancer))
	}
	// Describe LoadBalancer Attribute
	req := slb.CreateDescribeLoadBalancerAttributeRequest()
	req.LoadBalancerId = response.LoadBalancers.LoadBalancer[0].LoadBalancerId
	resp, err := p.auth.SLB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return true, nil, err
	}
	if resp == nil {
		// find the slb, but describe error. return true
		return true, nil, fmt.Errorf("describe loadbalancer attribute error: slb [%s] is nil", req.LoadBalancerId)
	}
	klog.Infof("%s successfully find loadbalancer, %s", name, util.PrettyJson(resp))

	lb := model.LoadBalancer{NamespacedName: name}
	lb.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	lb.LoadBalancerAttribute.LoadBalancerName = &resp.LoadBalancerName
	lb.LoadBalancerAttribute.Address = resp.Address
	lb.LoadBalancerAttribute.AddressType = &resp.AddressType
	lb.LoadBalancerAttribute.NetworkType = &resp.NetworkType
	lb.LoadBalancerAttribute.VpcId = resp.VpcId
	lb.LoadBalancerAttribute.VSwitchId = &resp.VSwitchId
	lb.LoadBalancerAttribute.Bandwidth = &resp.Bandwidth
	lb.LoadBalancerAttribute.MasterZoneId = &resp.MasterZoneId
	lb.LoadBalancerAttribute.SlaveZoneId = &resp.SlaveZoneId
	lb.LoadBalancerAttribute.DeleteProtection = &resp.DeleteProtection
	lb.LoadBalancerAttribute.InternetChargeType = &resp.InternetChargeType
	lb.LoadBalancerAttribute.LoadBalancerSpec = &resp.LoadBalancerSpec
	lb.LoadBalancerAttribute.ModificationProtectionStatus = &resp.ModificationProtectionStatus
	lb.LoadBalancerAttribute.ResourceGroupId = &resp.ResourceGroupId
	return true, &lb, nil

}

func (p ProviderSLB) CreateSLB(ctx context.Context, mdl *model.LoadBalancer) (*model.LoadBalancer, error) {
	req := slb.CreateCreateLoadBalancerRequest()
	setCreateSLBReqFromModel(req, mdl)
	req.ClientToken = utils.GetUUID()
	resp, err := p.auth.SLB.CreateLoadBalancer(req)
	if err != nil {
		return nil, err
	}
	mdl.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	mdl.LoadBalancerAttribute.Address = resp.Address
	return mdl, nil
}

func setCreateSLBReqFromModel(request *slb.CreateLoadBalancerRequest, mdl *model.LoadBalancer) {
	if mdl.LoadBalancerAttribute.AddressType != nil {
		request.AddressType = *mdl.LoadBalancerAttribute.AddressType
	}
	if mdl.LoadBalancerAttribute.InternetChargeType != nil {
		request.InternetChargeType = *mdl.LoadBalancerAttribute.InternetChargeType
	}
	if mdl.LoadBalancerAttribute.Bandwidth != nil {
		request.Bandwidth = requests.Integer(*mdl.LoadBalancerAttribute.Bandwidth)
	}
	if mdl.LoadBalancerAttribute.LoadBalancerName != nil {
		request.LoadBalancerName = *mdl.LoadBalancerAttribute.LoadBalancerName
	}
	if mdl.LoadBalancerAttribute.VpcId != "" {
		request.VpcId = mdl.LoadBalancerAttribute.VpcId
	}
	if mdl.LoadBalancerAttribute.VSwitchId != nil {
		request.VSwitchId = *mdl.LoadBalancerAttribute.VSwitchId
	}
	if mdl.LoadBalancerAttribute.MasterZoneId != nil {
		request.MasterZoneId = *mdl.LoadBalancerAttribute.MasterZoneId
	}
	if mdl.LoadBalancerAttribute.SlaveZoneId != nil {
		request.SlaveZoneId = *mdl.LoadBalancerAttribute.SlaveZoneId
	}
	if mdl.LoadBalancerAttribute.LoadBalancerSpec != nil {
		request.LoadBalancerSpec = *mdl.LoadBalancerAttribute.LoadBalancerSpec
	}
	if mdl.LoadBalancerAttribute.ResourceGroupId != nil {
		request.ResourceGroupId = *mdl.LoadBalancerAttribute.ResourceGroupId
	}
	if mdl.LoadBalancerAttribute.AddressIPVersion != nil {
		request.AddressIPVersion = *mdl.LoadBalancerAttribute.AddressIPVersion
	}
	if mdl.LoadBalancerAttribute.DeleteProtection != nil {
		request.DeleteProtection = *mdl.LoadBalancerAttribute.DeleteProtection
	}
	if mdl.LoadBalancerAttribute.ModificationProtectionStatus != nil {
		request.ModificationProtectionStatus = *mdl.LoadBalancerAttribute.ModificationProtectionStatus
		request.ModificationProtectionReason = mdl.LoadBalancerAttribute.ModificationProtectionReason
	}
}

func (p ProviderSLB) DeleteSLB(ctx context.Context, mdl *model.LoadBalancer) error {
	request := slb.CreateDeleteLoadBalancerRequest()
	request.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
	_, err := p.auth.SLB.DeleteLoadBalancer(request)
	return err
}
