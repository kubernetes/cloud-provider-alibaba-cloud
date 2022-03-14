package slb

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"reflect"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

func NewLBProvider(
	auth *base.ClientMgr,
) *SLBProvider {
	return &SLBProvider{auth: auth}
}

var _ prvd.ILoadBalancer = &SLBProvider{}

type SLBProvider struct {
	auth *base.ClientMgr
}

func (p SLBProvider) FindLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {

	// 1. find by loadbalancer id
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] try to find loadbalancer by id %s",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return p.DescribeLoadBalancer(ctx, mdl)
	}

	// 2. find by tags
	err := p.findLoadBalancerByTag(mdl)
	if err != nil {
		return err
	}

	// 3. find by loadbalancer name
	if mdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return p.findLoadBalancerByName(mdl)
	}

	return nil
}

func (p SLBProvider) findLoadBalancerByTag(mdl *model.LoadBalancer) error {
	items, err := json.Marshal(mdl.LoadBalancerAttribute.Tags)
	if err != nil {
		return fmt.Errorf("tags marshal error: %s", err.Error())
	}

	klog.Infof("[%s] try to find loadbalancer by tag %s", mdl.NamespacedName, string(items))
	req := slb.CreateDescribeLoadBalancersRequest()
	req.Tags = string(items)
	resp, err := p.auth.SLB.DescribeLoadBalancers(req)
	if err != nil {
		return fmt.Errorf("[%s] find loadbalancer by tag error: %s", mdl.NamespacedName, util.FormatErrorMessage(err))
	}
	klog.V(5).Infof("RequestId: %s, API: %s", resp.RequestId, "DescribeLoadBalancers")

	num := len(resp.LoadBalancers.LoadBalancer)
	if num == 0 {
		return nil
	}

	if num > 1 {
		var lbIds []string
		for _, lb := range resp.LoadBalancers.LoadBalancer {
			lbIds = append(lbIds, lb.LoadBalancerId)
		}
		return fmt.Errorf("[%s] find multiple loadbalances by tag, lbIds[%s]", mdl.NamespacedName,
			strings.Join(lbIds, ","))
	}

	loadResponse(resp.LoadBalancers.LoadBalancer[0], mdl)
	klog.Infof("[%s] find loadbalancer by tag, lbId [%s]", mdl.NamespacedName,
		resp.LoadBalancers.LoadBalancer[0].LoadBalancerId)
	return nil
}

func (p SLBProvider) findLoadBalancerByName(mdl *model.LoadBalancer) error {
	if mdl.LoadBalancerAttribute.LoadBalancerName == "" {
		klog.Warningf("[%s] find loadbalancer by name error: loadbalancer name is empty.", mdl.NamespacedName.String())
		return nil
	}
	klog.Infof("[%s] try to find loadbalancer by name %s",
		mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerName)
	req := slb.CreateDescribeLoadBalancersRequest()
	req.LoadBalancerName = mdl.LoadBalancerAttribute.LoadBalancerName
	resp, err := p.auth.SLB.DescribeLoadBalancers(req)
	if err != nil {
		return fmt.Errorf("[%s] find loadbalancer by name %s error: %s", mdl.NamespacedName,
			req.LoadBalancerName, util.FormatErrorMessage(err))
	}
	num := len(resp.LoadBalancers.LoadBalancer)
	if num == 0 {
		return nil
	}

	if num > 1 {
		var lbIds []string
		for _, lb := range resp.LoadBalancers.LoadBalancer {
			lbIds = append(lbIds, lb.LoadBalancerId)
		}
		return fmt.Errorf("[%s] find multiple loadbalances by name, lbIds[%s]", mdl.NamespacedName,
			strings.Join(lbIds, ","))
	}

	loadResponse(resp.LoadBalancers.LoadBalancer[0], mdl)
	klog.Infof("[%s] find loadbalancer by name, lbId [%s]", mdl.NamespacedName,
		mdl.LoadBalancerAttribute.LoadBalancerId)
	return nil
}

func (p SLBProvider) CreateLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	req := slb.CreateCreateLoadBalancerRequest()
	setRequest(req, mdl)
	req.ClientToken = utils.GetUUID()
	resp, err := p.auth.SLB.CreateLoadBalancer(req)
	if err != nil {
		return util.FormatErrorMessage(err)
	}
	mdl.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	mdl.LoadBalancerAttribute.Address = resp.Address
	return nil

}

func (p SLBProvider) DescribeLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	req := slb.CreateDescribeLoadBalancerAttributeRequest()
	req.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
	resp, err := p.auth.SLB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return util.FormatErrorMessage(err)
	}
	if resp == nil {
		klog.Errorf("RequestId: %s, lbId %s DescribeLoadBalancerAttribute response is nil",
			resp.RequestId, mdl.LoadBalancerAttribute.LoadBalancerId)
		return fmt.Errorf("DescribeLoadBalancer response is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s, lbId: %s", resp.RequestId, "DescribeLoadBalancer", req.LoadBalancerId)
	loadResponse(*resp, mdl)
	return nil
}

func (p SLBProvider) DeleteLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	req := slb.CreateDeleteLoadBalancerRequest()
	req.LoadBalancerId = mdl.LoadBalancerAttribute.LoadBalancerId
	_, err := p.auth.SLB.DeleteLoadBalancer(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) SetLoadBalancerDeleteProtection(ctx context.Context, lbId string, flag string) error {
	req := slb.CreateSetLoadBalancerDeleteProtectionRequest()
	req.LoadBalancerId = lbId
	req.DeleteProtection = flag
	_, err := p.auth.SLB.SetLoadBalancerDeleteProtection(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) ModifyLoadBalancerInstanceSpec(ctx context.Context, lbId string, spec string) error {
	req := slb.CreateModifyLoadBalancerInstanceSpecRequest()
	req.LoadBalancerId = lbId
	req.LoadBalancerSpec = spec
	_, err := p.auth.SLB.ModifyLoadBalancerInstanceSpec(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) SetLoadBalancerName(ctx context.Context, lbId string, name string) error {
	req := slb.CreateSetLoadBalancerNameRequest()
	req.LoadBalancerId = lbId
	req.LoadBalancerName = name
	_, err := p.auth.SLB.SetLoadBalancerName(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) ModifyLoadBalancerInternetSpec(ctx context.Context, lbId string, chargeType string, bandwidth int) error {
	req := slb.CreateModifyLoadBalancerInternetSpecRequest()
	req.LoadBalancerId = lbId
	req.InternetChargeType = chargeType
	req.Bandwidth = requests.NewInteger(bandwidth)
	_, err := p.auth.SLB.ModifyLoadBalancerInternetSpec(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) SetLoadBalancerModificationProtection(ctx context.Context, lbId string, flag string) error {
	req := slb.CreateSetLoadBalancerModificationProtectionRequest()
	req.LoadBalancerId = lbId
	req.ModificationProtectionStatus = flag
	if flag == string(model.OnFlag) {
		req.ModificationProtectionReason = model.ModificationProtectionReason
	}
	_, err := p.auth.SLB.SetLoadBalancerModificationProtection(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) AddTags(ctx context.Context, lbId string, tags string) error {
	req := slb.CreateAddTagsRequest()
	req.LoadBalancerId = lbId
	req.Tags = tags
	_, err := p.auth.SLB.AddTags(req)
	return util.FormatErrorMessage(err)
}

func (p SLBProvider) DescribeTags(ctx context.Context, lbId string) ([]model.Tag, error) {
	req := slb.CreateDescribeTagsRequest()
	req.LoadBalancerId = lbId
	resp, err := p.auth.SLB.DescribeTags(req)
	if err != nil {
		return nil, util.FormatErrorMessage(err)
	}
	var tags []model.Tag
	for _, tag := range resp.TagSets.TagSet {
		tags = append(tags, model.Tag{
			TagValue: tag.TagValue,
			TagKey:   tag.TagKey,
		})
	}
	return tags, nil
}

// UntagResources used for e2etest
func (p SLBProvider) UntagResources(ctx context.Context, lbId string, tagKey *[]string) error {
	req := slb.CreateUntagResourcesRequest()
	req.ResourceId = &[]string{lbId}
	req.ResourceType = "instance"
	req.TagKey = tagKey
	_, err := p.auth.SLB.UntagResources(req)
	return err
}

// DescribeAvailableResource used for e2etest
func (p SLBProvider) DescribeAvailableResource(ctx context.Context, addressType, AddressIPVersion string) ([]slb.AvailableResource, error) {
	req := slb.CreateDescribeAvailableResourceRequest()
	req.AddressType = addressType
	req.AddressIPVersion = AddressIPVersion
	resp, err := p.auth.SLB.DescribeAvailableResource(req)
	if err != nil {
		return nil, err
	}
	return resp.AvailableResources.AvailableResource, nil
}

// CreateAccessControlList used for e2etest
func (p SLBProvider) CreateAccessControlList(ctx context.Context, aclName string) (string, error) {
	req := slb.CreateCreateAccessControlListRequest()
	req.AclName = aclName
	resp, err := p.auth.SLB.CreateAccessControlList(req)
	if err != nil {
		return "", err
	}
	return resp.AclId, nil
}

// DeleteAccessControlList used for e2etest
func (p SLBProvider) DeleteAccessControlList(ctx context.Context, aclId string) error {
	req := slb.CreateDeleteAccessControlListRequest()
	req.AclId = aclId
	_, err := p.auth.SLB.DeleteAccessControlList(req)
	return err
}

// DescribeServerCertificates used for e2etest
func (p SLBProvider) DescribeServerCertificates(ctx context.Context) ([]string, error) {
	req := slb.CreateDescribeServerCertificatesRequest()
	resp, err := p.auth.SLB.DescribeServerCertificates(req)
	if err != nil {
		return nil, err
	}
	var certs []string
	for _, cert := range resp.ServerCertificates.ServerCertificate {
		certs = append(certs, cert.ServerCertificateId)
	}
	return certs, nil
}

func setRequest(request *slb.CreateLoadBalancerRequest, mdl *model.LoadBalancer) {
	if mdl.LoadBalancerAttribute.AddressType != "" {
		request.AddressType = string(mdl.LoadBalancerAttribute.AddressType)
	}
	if mdl.LoadBalancerAttribute.InternetChargeType != "" {
		request.InternetChargeType = string(mdl.LoadBalancerAttribute.InternetChargeType)
	}
	if mdl.LoadBalancerAttribute.Bandwidth != 0 {
		request.Bandwidth = requests.NewInteger(mdl.LoadBalancerAttribute.Bandwidth)
	}
	if mdl.LoadBalancerAttribute.LoadBalancerName != "" {
		request.LoadBalancerName = mdl.LoadBalancerAttribute.LoadBalancerName
	}
	if mdl.LoadBalancerAttribute.VpcId != "" {
		request.VpcId = mdl.LoadBalancerAttribute.VpcId
	}
	if mdl.LoadBalancerAttribute.VSwitchId != "" {
		request.VSwitchId = mdl.LoadBalancerAttribute.VSwitchId
	}
	if mdl.LoadBalancerAttribute.MasterZoneId != "" {
		request.MasterZoneId = mdl.LoadBalancerAttribute.MasterZoneId
	}
	if mdl.LoadBalancerAttribute.SlaveZoneId != "" {
		request.SlaveZoneId = mdl.LoadBalancerAttribute.SlaveZoneId
	}
	if mdl.LoadBalancerAttribute.LoadBalancerSpec != "" {
		request.LoadBalancerSpec = string(mdl.LoadBalancerAttribute.LoadBalancerSpec)
	}
	if mdl.LoadBalancerAttribute.ResourceGroupId != "" {
		request.ResourceGroupId = mdl.LoadBalancerAttribute.ResourceGroupId
	}
	if mdl.LoadBalancerAttribute.AddressIPVersion != "" {
		request.AddressIPVersion = string(mdl.LoadBalancerAttribute.AddressIPVersion)
	}
	if mdl.LoadBalancerAttribute.DeleteProtection != "" {
		request.DeleteProtection = string(mdl.LoadBalancerAttribute.DeleteProtection)
	}
	if mdl.LoadBalancerAttribute.ModificationProtectionStatus != "" {
		request.ModificationProtectionStatus = string(mdl.LoadBalancerAttribute.ModificationProtectionStatus)
		request.ModificationProtectionReason = mdl.LoadBalancerAttribute.ModificationProtectionReason
	}
}

func loadResponse(resp interface{}, lb *model.LoadBalancer) {
	v := reflect.ValueOf(resp)
	lb.LoadBalancerAttribute.LoadBalancerId = v.FieldByName("LoadBalancerId").String()
	lb.LoadBalancerAttribute.LoadBalancerName = v.FieldByName("LoadBalancerName").String()
	lb.LoadBalancerAttribute.Address = v.FieldByName("Address").String()
	lb.LoadBalancerAttribute.AddressType = model.AddressType(v.FieldByName("AddressType").String())
	lb.LoadBalancerAttribute.AddressIPVersion = model.AddressIPVersionType(v.FieldByName("AddressIPVersion").String())
	lb.LoadBalancerAttribute.NetworkType = v.FieldByName("NetworkType").String()
	lb.LoadBalancerAttribute.VpcId = v.FieldByName("VpcId").String()
	lb.LoadBalancerAttribute.VSwitchId = v.FieldByName("VSwitchId").String()
	lb.LoadBalancerAttribute.Bandwidth = int(v.FieldByName("Bandwidth").Int())
	lb.LoadBalancerAttribute.MasterZoneId = v.FieldByName("MasterZoneId").String()
	lb.LoadBalancerAttribute.SlaveZoneId = v.FieldByName("SlaveZoneId").String()
	lb.LoadBalancerAttribute.DeleteProtection = model.FlagType(v.FieldByName("DeleteProtection").String())
	lb.LoadBalancerAttribute.InternetChargeType = model.InternetChargeType(v.FieldByName("InternetChargeType").String())
	lb.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(v.FieldByName("LoadBalancerSpec").String())
	lb.LoadBalancerAttribute.ModificationProtectionStatus = model.ModificationProtectionType(
		v.FieldByName("ModificationProtectionStatus").String())
	lb.LoadBalancerAttribute.ResourceGroupId = v.FieldByName("ResourceGroupId").String()
}
