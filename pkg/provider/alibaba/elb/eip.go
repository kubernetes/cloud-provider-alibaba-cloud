package elb

import (
	"context"
	"fmt"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"

	sdkerror "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ens"
)

func (e ELBProvider) CreateEip(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateCreateEipInstanceRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.EnsRegionId = mdl.EipAttribute.EnsRegionId
	req.InstanceChargeType = mdl.EipAttribute.InstanceChargeType
	req.InternetChargeType = mdl.EipAttribute.InternetChargeType
	req.Name = mdl.EipAttribute.Name
	req.Bandwidth = requests.NewInteger(mdl.EipAttribute.Bandwidth)
	resp, err := e.auth.ELB.CreateEipInstance(req)
	if err != nil {
		return util.SDKError("CreateEipInstance", err)
	}
	mdl.EipAttribute.AllocationId = resp.AllocationId
	return nil
}

func (e ELBProvider) ReleaseEip(ctx context.Context, eipId string) error {
	req := ens.CreateReleaseInstanceRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.InstanceId = eipId
	_, err := e.auth.ELB.ReleaseInstance(req)
	if err != nil {
		if serverErr, ok := err.(*sdkerror.ServerError); ok && serverErr.ErrorCode() == "5100" {
			klog.Warningf("eip %s might already been deleted, error code %s", eipId, err)
			return nil
		}
		return util.SDKError("ReleaseInstance", err)
	}
	return nil
}

func (e ELBProvider) AssociateElbEipAddress(ctx context.Context, eipId, elbId string) error {
	req := ens.CreateAssociateEnsEipAddressRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.AllocationId = eipId
	req.InstanceId = elbId
	req.InstanceType = "SlbInstance"
	_, err := e.auth.ELB.AssociateEnsEipAddress(req)
	if err != nil {
		return util.SDKError("AssociateEnsEipAddress", err)
	}
	return nil
}

func (e ELBProvider) UnAssociateElbEipAddress(ctx context.Context, eipId string) error {
	req := ens.CreateUnAssociateEnsEipAddressRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.AllocationId = eipId
	_, err := e.auth.ELB.UnAssociateEnsEipAddress(req)
	if err != nil {
		return util.SDKError("UnAssociateEnsEipAddress", err)
	}
	return nil
}

func (e ELBProvider) ModifyEipAttribute(ctx context.Context, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateModifyEnsEipAddressAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.AllocationId = eipId
	req.Bandwidth = requests.NewInteger(mdl.EipAttribute.Bandwidth)
	req.Description = mdl.EipAttribute.Description
	_, err := e.auth.ELB.ModifyEnsEipAddressAttribute(req)
	if err != nil {
		return util.SDKError("ModifyEnsEipAddressAttribute", err)
	}
	return nil
}

func (e ELBProvider) DescribeEnsEipAddressesById(ctx context.Context, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateDescribeEnsEipAddressesRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.AllocationId = eipId
	resp, err := e.auth.ELB.DescribeEnsEipAddresses(req)
	if err != nil {
		return util.SDKError("DescribeEnsEipAddresses", err)
	}
	if resp == nil {
		klog.Errorf("RequestId: %s, eip address %s DescribeLoadBalancerAttribute response is nil", resp.RequestId, eipId)
		return fmt.Errorf("find no instance by eip allocation Id %s", eipId)
	}
	var found = false
	var instance ens.EipAddress
	for _, instance = range resp.EipAddresses.EipAddress {
		if instance.AllocationId == eipId {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("find no instance by eip allocation Id %s", eipId)
	}
	loadEipResponse(instance, mdl)
	return nil
}

func (e ELBProvider) DescribeEnsEipAddressesByName(ctx context.Context, eipName string, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateDescribeEnsEipAddressesRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	resp, err := e.auth.ELB.DescribeEnsEipAddresses(req)
	if err != nil {
		return util.SDKError("DescribeEnsEipAddresses", err)
	}
	if resp == nil {
		klog.Errorf("RequestId: %s, eip address %s DescribeLoadBalancerAttribute response is nil", resp.RequestId, eipName)
		return fmt.Errorf("find no instance by eip name %s", eipName)
	}
	var found = false
	var instance ens.EipAddress
	for _, instance = range resp.EipAddresses.EipAddress {
		if instance.Name == eipName {
			found = true
			break
		}
	}
	if found {
		loadEipResponse(instance, mdl)
		mdl.EipAttribute.IsUserManaged = false
		return nil
	}
	return nil
}

func (e ELBProvider) FindAssociatedInstance(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateDescribeEnsEipAddressesRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.AssociatedInstanceId = mdl.GetLoadBalancerId()
	req.AssociatedInstanceType = "SlbInstance"
	resp, err := e.auth.ELB.DescribeEnsEipAddresses(req)
	if err != nil {
		return util.SDKError("DescribeEnsEipAddresses", err)
	}
	if resp == nil {
		klog.Errorf("RequestId: %s, eip address %s DescribeLoadBalancerAttribute response is nil", resp.RequestId, mdl.GetEipName())
		return fmt.Errorf("find no instance by eip name %s", mdl.GetEipName())
	}
	if resp.TotalCount == 0 {
		return nil
	}
	if resp.TotalCount == 1 && resp.EipAddresses.EipAddress[0].InstanceId == mdl.GetLoadBalancerId() {
		if resp.EipAddresses.EipAddress[0].AllocationId == "" || resp.EipAddresses.EipAddress[0].Name == "" {
			return fmt.Errorf("lacks eip id and name")
		}
		mdl.LoadBalancerAttribute.AssociatedEipId = resp.EipAddresses.EipAddress[0].AllocationId
		mdl.LoadBalancerAttribute.AssociatedEipName = resp.EipAddresses.EipAddress[0].Name
		mdl.LoadBalancerAttribute.AssociatedEipAddress = resp.EipAddresses.EipAddress[0].IpAddress
		return nil
	}
	return nil
}

func loadEipResponse(eip ens.EipAddress, mdl *elbmodel.EdgeLoadBalancer) {
	mdl.EipAttribute.Name = eip.Name
	mdl.EipAttribute.AllocationId = eip.AllocationId
	mdl.EipAttribute.EnsRegionId = eip.EnsRegionId
	mdl.EipAttribute.IpAddress = eip.IpAddress
	mdl.EipAttribute.InternetChargeType = eip.InternetChargeType
	mdl.EipAttribute.InstanceChargeType = eip.ChargeType
	mdl.EipAttribute.Bandwidth = eip.Bandwidth
	mdl.EipAttribute.InstanceId = eip.InstanceId
	mdl.EipAttribute.InstanceType = eip.InstanceType
	mdl.EipAttribute.Status = eip.Status
	mdl.EipAttribute.Description = eip.Description

}
