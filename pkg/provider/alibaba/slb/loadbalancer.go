package slb

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"os"
	"reflect"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
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
		lbId := mdl.LoadBalancerAttribute.LoadBalancerId
		klog.Infof("[%s] find loadbalancer by id, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		err := p.DescribeLoadBalancer(ctx, mdl)
		if err != nil {
			return err
		}
		// Check LoadBalancer ID once more to prevent from an abnormal response from API.
		if mdl.LoadBalancerAttribute.LoadBalancerId != lbId {
			return fmt.Errorf("[%s] find loadbalancer by id error: loadbalancer id from API not match, expect [%s], actual [%s]",
				mdl.NamespacedName, lbId, mdl.LoadBalancerAttribute.LoadBalancerId)
		}
		return nil
	}

	// 2. find by tags
	err := p.findLoadBalancerByTag(mdl)
	if err != nil {
		return err
	}
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find loadbalancer by tag, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return nil
	}

	// 3. find by loadbalancer name
	err = p.FindLoadBalancerByName(mdl)
	if err != nil {
		return err
	}
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find loadbalancer by name, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return nil
	}

	klog.Infof("[%s] find no loadbalancer", mdl.NamespacedName)
	return nil
}

func (p SLBProvider) findLoadBalancerByTag(mdl *model.LoadBalancer) error {
	var tags []slb.Tag
	for _, t := range mdl.LoadBalancerAttribute.Tags {
		tags = append(tags, slb.Tag{
			TagKey:   t.Key,
			TagValue: t.Value,
		})
	}
	items, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("tags marshal error: %s", err.Error())
	}

	klog.Infof("[%s] try to find loadbalancer by tag %s", mdl.NamespacedName, string(items))
	req := slb.CreateDescribeLoadBalancersRequest()
	req.Tags = string(items)
	resp, err := p.auth.SLB.DescribeLoadBalancers(req)
	if err != nil {
		return fmt.Errorf("[%s] find loadbalancer by tag error: %s", mdl.NamespacedName,
			util.SDKError("DescribeLoadBalancers", err).Error())
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
	return nil
}

func (p SLBProvider) FindLoadBalancerByName(mdl *model.LoadBalancer) error {
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
			req.LoadBalancerName, util.SDKError("DescribeLoadBalancers", err).Error())
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
	return nil
}

func (p SLBProvider) CreateLoadBalancer(ctx context.Context, mdl *model.LoadBalancer, clientToken string) error {
	req := slb.CreateCreateLoadBalancerRequest()
	setRequest(req, mdl)
	if clientToken != "" {
		req.ClientToken = clientToken
	}
	if ascmContext := os.Getenv("X-ACSPROXY-ASCM-CONTEXT"); ascmContext != "" {
		req.GetHeaders()["x-acsproxy-ascm-context"] = ascmContext
	}
	resp, err := p.auth.SLB.CreateLoadBalancer(req)
	if err != nil {
		return util.SDKError("CreateLoadBalancer", err)
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
		return util.SDKError("DescribeLoadBalancerAttribute", err)
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
	return util.SDKError("DeleteLoadBalancer", err)
}

func (p SLBProvider) SetLoadBalancerDeleteProtection(ctx context.Context, lbId string, flag string) error {
	req := slb.CreateSetLoadBalancerDeleteProtectionRequest()
	req.LoadBalancerId = lbId
	req.DeleteProtection = flag
	_, err := p.auth.SLB.SetLoadBalancerDeleteProtection(req)
	return util.SDKError("SetLoadBalancerDeleteProtection", err)
}

func (p SLBProvider) ModifyLoadBalancerInstanceSpec(ctx context.Context, lbId string, spec string) error {
	req := slb.CreateModifyLoadBalancerInstanceSpecRequest()
	req.LoadBalancerId = lbId
	req.LoadBalancerSpec = spec
	_, err := p.auth.SLB.ModifyLoadBalancerInstanceSpec(req)
	return util.SDKError("ModifyLoadBalancerInstanceSpec", err)
}

func (p SLBProvider) SetLoadBalancerName(ctx context.Context, lbId string, name string) error {
	req := slb.CreateSetLoadBalancerNameRequest()
	req.LoadBalancerId = lbId
	req.LoadBalancerName = name
	_, err := p.auth.SLB.SetLoadBalancerName(req)
	return util.SDKError("SetLoadBalancerName", err)
}

func (p SLBProvider) ModifyLoadBalancerInternetSpec(ctx context.Context, lbId string, chargeType string, bandwidth int) error {
	req := slb.CreateModifyLoadBalancerInternetSpecRequest()
	req.LoadBalancerId = lbId
	req.InternetChargeType = chargeType
	req.Bandwidth = requests.NewInteger(bandwidth)
	_, err := p.auth.SLB.ModifyLoadBalancerInternetSpec(req)
	return util.SDKError("ModifyLoadBalancerInternetSpec", err)
}

func (p SLBProvider) SetLoadBalancerModificationProtection(ctx context.Context, lbId string, flag string) error {
	req := slb.CreateSetLoadBalancerModificationProtectionRequest()
	req.LoadBalancerId = lbId
	req.ModificationProtectionStatus = flag
	if flag == string(model.OnFlag) {
		req.ModificationProtectionReason = model.ModificationProtectionReason
	}
	_, err := p.auth.SLB.SetLoadBalancerModificationProtection(req)
	return util.SDKError("SetLoadBalancerModificationProtection", err)
}

func (p SLBProvider) ModifyLoadBalancerInstanceChargeType(ctx context.Context, lbId string, instanceChargeType string, spec string) error {
	req := slb.CreateModifyLoadBalancerInstanceChargeTypeRequest()
	req.LoadBalancerId = lbId
	req.InstanceChargeType = instanceChargeType
	if spec != "" {
		req.LoadBalancerSpec = spec
	}
	_, err := p.auth.SLB.ModifyLoadBalancerInstanceChargeType(req)
	return util.SDKError("ModifyLoadBalancerInstanceChargeType", err)
}

func (p SLBProvider) TagCLBResource(ctx context.Context, resourceId string, tags []tag.Tag) error {
	req := slb.CreateTagResourcesRequest()
	req.ResourceId = &[]string{resourceId}
	req.ResourceType = "instance"
	var reqTags []slb.TagResourcesTag
	for _, t := range tags {
		reqTags = append(reqTags, slb.TagResourcesTag{
			Key:   t.Key,
			Value: t.Value,
		})
	}
	req.Tag = &reqTags

	_, err := p.auth.SLB.TagResources(req)
	return util.SDKError("TagResources", err)
}

func (p SLBProvider) ListCLBTagResources(ctx context.Context, lbId string) ([]tag.Tag, error) {
	req := slb.CreateListTagResourcesRequest()
	req.ResourceId = &[]string{lbId}
	req.ResourceType = "instance"

	resp, err := p.auth.SLB.ListTagResources(req)
	if err != nil {
		return nil, util.SDKError("ListTagResources", err)
	}
	var tags []tag.Tag
	for _, v := range resp.TagResources.TagResource {
		tags = append(tags, tag.Tag{
			Value: v.TagValue,
			Key:   v.TagKey,
		})
	}
	return tags, nil
}

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

// DescribeAccessControlList used for e2etest
func (p SLBProvider) DescribeAccessControlList(ctx context.Context, aclName string) (string, error) {
	req := slb.CreateDescribeAccessControlListsRequest()
	req.AclName = aclName
	resp, err := p.auth.SLB.DescribeAccessControlLists(req)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return "", nil
		}
		return "", err
	}
	if len(resp.Acls.Acl) == 0 {
		return "", nil
	}
	return resp.Acls.Acl[0].AclId, nil
}

// DeleteAccessControlList used for e2etest
func (p SLBProvider) DeleteAccessControlList(ctx context.Context, aclId string) error {
	req := slb.CreateDeleteAccessControlListRequest()
	req.AclId = aclId
	_, err := p.auth.SLB.DeleteAccessControlList(req)
	return err
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
	if mdl.LoadBalancerAttribute.InstanceChargeType != "" {
		request.InstanceChargeType = string(mdl.LoadBalancerAttribute.InstanceChargeType)
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
	if mdl.LoadBalancerAttribute.Address != "" {
		request.Address = mdl.LoadBalancerAttribute.Address
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
	lb.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(v.FieldByName("LoadBalancerSpec").String())
	lb.LoadBalancerAttribute.ModificationProtectionStatus = model.ModificationProtectionType(
		v.FieldByName("ModificationProtectionStatus").String())
	lb.LoadBalancerAttribute.ResourceGroupId = v.FieldByName("ResourceGroupId").String()
	lb.LoadBalancerAttribute.InstanceChargeType = model.InstanceChargeType(v.FieldByName("InstanceChargeType").String())

	switch t := resp.(type) {
	// DescribeLoadBalancers
	case slb.LoadBalancer:
		lb.LoadBalancerAttribute.InternetChargeType = model.InternetChargeType(v.FieldByName("InternetChargeTypeAlias").String())
		var tags []tag.Tag
		for _, tg := range t.Tags.Tag {
			tags = append(tags, tag.Tag{Key: tg.TagKey, Value: tg.TagValue})
		}
		lb.LoadBalancerAttribute.Tags = tags
	// DescribeLoadBalancerAttribute
	case slb.DescribeLoadBalancerAttributeResponse:
		lb.LoadBalancerAttribute.InternetChargeType = model.InternetChargeType(v.FieldByName("InternetChargeType").String())
		var tags []tag.Tag
		for _, tg := range t.Tags.Tag {
			tags = append(tags, tag.Tag{Key: tg.TagKey, Value: tg.TagValue})
		}
		lb.LoadBalancerAttribute.Tags = tags
	default:
		klog.Errorf("not support type: %T", t)
	}

}
