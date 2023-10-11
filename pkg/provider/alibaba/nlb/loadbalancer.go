package nlb

import (
	"context"
	"fmt"
	nlb "github.com/alibabacloud-go/nlb-20220430/client"
	"github.com/alibabacloud-go/tea/tea"
	"k8s.io/apimachinery/pkg/util/wait"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	pkgUtil "k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

func NewNLBProvider(
	auth *base.ClientMgr,
) *NLBProvider {
	return &NLBProvider{auth: auth}
}

var _ prvd.INLB = &NLBProvider{}

type NLBProvider struct {
	auth *base.ClientMgr
}

type LoadBalancerStatus string

const (
	Active       = LoadBalancerStatus("Active")
	Provisioning = LoadBalancerStatus("Provisioning")
)

func (p *NLBProvider) FindNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	// 1. find by nlb id
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find nlb by id, nlb info [%+v]",
			mdl.NamespacedName, pkgUtil.PrettyJson(mdl))
		return p.DescribeNLB(ctx, mdl)
	}

	// 2. find by tags
	err := p.findNLBByTag(mdl)
	if err != nil {
		return err
	}
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find nlb by tag, nlb info [%+v]",
			mdl.NamespacedName, pkgUtil.PrettyJson(mdl))
		return nil
	}

	// 3. find by name
	err = p.findNLBByName(mdl)
	if err != nil {
		return err
	}
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find nlb by name, nlb info [%+v]",
			mdl.NamespacedName, pkgUtil.PrettyJson(mdl))
		return nil
	}

	klog.Infof("[%s] find no nlb", mdl.NamespacedName)
	return nil
}

func (p *NLBProvider) DescribeNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	resp, err := p.waitNLBActive(mdl.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return err
	}

	return loadResponse(resp.Body, mdl)

}

func (p *NLBProvider) CreateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	req := &nlb.CreateLoadBalancerRequest{
		AddressType:      tea.String(mdl.LoadBalancerAttribute.AddressType),
		LoadBalancerName: tea.String(mdl.LoadBalancerAttribute.Name),
		VpcId:            tea.String(mdl.LoadBalancerAttribute.VpcId),
		ZoneMappings:     []*nlb.CreateLoadBalancerRequestZoneMappings{},
	}
	if mdl.LoadBalancerAttribute.ResourceGroupId != "" {
		req.ResourceGroupId = tea.String(mdl.LoadBalancerAttribute.ResourceGroupId)
	}
	if mdl.LoadBalancerAttribute.AddressIpVersion != "" {
		req.AddressIpVersion = tea.String(mdl.LoadBalancerAttribute.AddressIpVersion)
	}
	for _, z := range mdl.LoadBalancerAttribute.ZoneMappings {
		req.ZoneMappings = append(req.ZoneMappings,
			&nlb.CreateLoadBalancerRequestZoneMappings{
				VSwitchId:          tea.String(z.VSwitchId),
				ZoneId:             tea.String(z.ZoneId),
				AllocationId:       tea.String(z.AllocationId),
				PrivateIPv4Address: tea.String(z.IPv4Addr),
			})
	}

	resp, err := p.auth.NLB.CreateLoadBalancer(req)
	if err != nil {
		return util.SDKError("CreateLoadBalancer", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI CreateLoadBalancer resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "CreateLoadBalancer")

	mdl.LoadBalancerAttribute.LoadBalancerId = tea.StringValue(resp.Body.LoadbalancerId)
	return nil
}

func (p *NLBProvider) DeleteNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	req := &nlb.DeleteLoadBalancerRequest{}
	req.LoadBalancerId = tea.String(mdl.LoadBalancerAttribute.LoadBalancerId)
	resp, err := p.auth.NLB.DeleteLoadBalancer(req)
	if err != nil {
		return util.SDKError("DeleteLoadBalancer", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI DeleteNLB resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DeleteLoadBalancer")

	return p.waitJobFinish("DeleteLoadBalancer", tea.StringValue(resp.Body.JobId), 20*time.Second, 3*time.Minute)
}

func (p *NLBProvider) UpdateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	req := &nlb.UpdateLoadBalancerAttributeRequest{}
	req.LoadBalancerId = tea.String(mdl.LoadBalancerAttribute.LoadBalancerId)
	if mdl.LoadBalancerAttribute.Name != "" {
		req.LoadBalancerName = tea.String(mdl.LoadBalancerAttribute.Name)
	}
	resp, err := p.auth.NLB.UpdateLoadBalancerAttribute(req)
	if err != nil {
		return util.SDKError("UpdateLoadBalancerAttribute", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI UpdateLoadBalancerAttribute resp is nil")
	}

	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateLoadBalancerAttribute")
	return nil
}

func (p *NLBProvider) UpdateNLBAddressType(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	req := &nlb.UpdateLoadBalancerAddressTypeConfigRequest{}
	req.LoadBalancerId = tea.String(mdl.LoadBalancerAttribute.LoadBalancerId)
	req.AddressType = tea.String(mdl.LoadBalancerAttribute.AddressType)

	resp, err := p.auth.NLB.UpdateLoadBalancerAddressTypeConfig(req)
	if err != nil {
		return util.SDKError("UpdateNLBAddressType", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI UpdateLoadBalancerAddressTypeConfig resp is nil")
	}

	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateLoadBalancerAddressTypeConfig")
	return nil
}

func (p *NLBProvider) UpdateNLBZones(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	req := &nlb.UpdateLoadBalancerZonesRequest{}
	req.LoadBalancerId = tea.String(mdl.LoadBalancerAttribute.LoadBalancerId)

	for _, z := range mdl.LoadBalancerAttribute.ZoneMappings {
		zoneMapping := &nlb.UpdateLoadBalancerZonesRequestZoneMappings{
			VSwitchId: tea.String(z.VSwitchId),
			ZoneId:    tea.String(z.ZoneId),
		}
		if z.IPv4Addr != "" {
			zoneMapping.PrivateIPv4Address = tea.String(z.IPv4Addr)
		}
		if z.AllocationId != "" {
			zoneMapping.AllocationId = tea.String(z.AllocationId)
		}
		req.ZoneMappings = append(req.ZoneMappings, zoneMapping)
	}

	resp, err := p.auth.NLB.UpdateLoadBalancerZones(req)
	if err != nil {
		return util.SDKError("UpdateLoadBalancerZones", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI UpdateLoadBalancerZones resp is nil")
	}

	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UpdateLoadBalancerZones")
	return nil
}

func (p *NLBProvider) UpdateNLBSecurityGroupIds(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer, added, removed []string) error {
	// leave first, then join
	if len(removed) != 0 {
		req := &nlb.LoadBalancerLeaveSecurityGroupRequest{}
		req.LoadBalancerId = tea.String(mdl.LoadBalancerAttribute.LoadBalancerId)
		req.SecurityGroupIds = tea.StringSlice(removed)
		resp, err := p.auth.NLB.LoadBalancerLeaveSecurityGroup(req)
		if err != nil {
			return util.SDKError("LoadBalancerLeaveSecurityGroup", err)
		}
		if resp == nil || resp.Body == nil {
			return fmt.Errorf("OpenAPI LoadBalancerLeaveSecurityGroup resp is nil")
		}
		klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "LoadBalancerLeaveSecurityGroup")

		err = p.waitJobFinish("LoadBalancerLeaveSecurityGroup", tea.StringValue(resp.Body.JobId))
		if err != nil {
			return err
		}
	}

	if len(added) != 0 {
		req := &nlb.LoadBalancerJoinSecurityGroupRequest{}
		req.LoadBalancerId = tea.String(mdl.LoadBalancerAttribute.LoadBalancerId)
		req.SecurityGroupIds = tea.StringSlice(added)
		resp, err := p.auth.NLB.LoadBalancerJoinSecurityGroup(req)
		if err != nil {
			return util.SDKError("LoadBalancerJoinSecurityGroup", err)
		}
		if resp == nil || resp.Body == nil {
			return fmt.Errorf("OpenAPI LoadBalancerJoinSecurityGroup resp is nil")
		}
		klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "LoadBalancerJoinSecurityGroup")

		err = p.waitJobFinish("LoadBalancerJoinSecurityGroup", tea.StringValue(resp.Body.JobId))
		if err != nil {
			return err
		}
	}

	return nil
}

// tag
func (p *NLBProvider) TagNLBResource(ctx context.Context, resourceId string, resourceType nlbmodel.TagResourceType, tags []tag.Tag,
) error {
	req := &nlb.TagResourcesRequest{}
	req.ResourceType = tea.String(string(resourceType))
	req.ResourceId = []*string{tea.String(resourceId)}
	for _, v := range tags {
		req.Tag = append(req.Tag, &nlb.TagResourcesRequestTag{
			Key:   tea.String(v.Key),
			Value: tea.String(v.Value),
		})
	}

	resp, err := p.auth.NLB.TagResources(req)
	if err != nil {
		return util.SDKError("TagResources", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI TagResources resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "TagResources")
	return nil
}

func (p *NLBProvider) ListNLBTagResources(ctx context.Context, lbId string) ([]tag.Tag, error) {
	req := &nlb.ListTagResourcesRequest{}
	req.ResourceType = tea.String("loadbalancer")
	req.ResourceId = []*string{tea.String(lbId)}

	resp, err := p.auth.NLB.ListTagResources(req)
	if err != nil {
		return nil, fmt.Errorf("list nlb %s tag error: %s", lbId, util.SDKError("ListTagResources", err))
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("OpenAPI ListTagResources resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "ListTagResources")

	var ret []tag.Tag
	for _, v := range resp.Body.TagResources {
		if v != nil {
			ret = append(ret, tag.Tag{
				Key:   tea.StringValue(v.TagKey),
				Value: tea.StringValue(v.TagValue),
			})
		}
	}
	return ret, nil
}

func (p *NLBProvider) findNLBByTag(mdl *nlbmodel.NetworkLoadBalancer) error {
	klog.Infof("[%s] try to find nlb by tag %+v", mdl.NamespacedName, mdl.LoadBalancerAttribute.Tags)
	req := &nlb.ListLoadBalancersRequest{}
	for _, v := range mdl.LoadBalancerAttribute.Tags {
		req.Tag = append(req.Tag,
			&nlb.ListLoadBalancersRequestTag{
				Key:   tea.String(v.Key),
				Value: tea.String(v.Value),
			},
		)
	}
	resp, err := p.auth.NLB.ListLoadBalancers(req)
	if err != nil {
		return fmt.Errorf("[%s] find nlb by tag error: %s", mdl.NamespacedName, util.SDKError("ListLoadBalancers", err))
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI ListLoadBalancers resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "ListLoadBalancers")

	num := len(resp.Body.LoadBalancers)
	if num == 0 {
		return nil
	}

	if resp.Body.LoadBalancers[0] == nil {
		return fmt.Errorf("ListLoadBalancers resp nlb is nil, resp: %+v", resp)
	}

	if num > 1 {
		var lbIds []string
		for _, lb := range resp.Body.LoadBalancers {
			if lb != nil && lb.LoadBalancerId != nil {
				lbIds = append(lbIds, tea.StringValue(lb.LoadBalancerId))
			}
		}
		return fmt.Errorf("[%s] find multiple loadbalances by tag, lbIds[%s]", mdl.NamespacedName,
			strings.Join(lbIds, ","))
	}

	if tea.StringValue(resp.Body.LoadBalancers[0].LoadBalancerStatus) == string(Active) {
		return loadResponse(resp.Body.LoadBalancers[0], mdl)
	}

	getResp, err := p.waitNLBActive(tea.StringValue(resp.Body.LoadBalancers[0].LoadBalancerId))
	if err != nil {
		return err
	}
	return loadResponse(getResp.Body, mdl)
}

func (p *NLBProvider) FindNLBByName(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return p.findNLBByName(mdl)
}

func (p *NLBProvider) findNLBByName(mdl *nlbmodel.NetworkLoadBalancer) error {
	if mdl.LoadBalancerAttribute.Name == "" {
		klog.Warningf("[%s] find nlb by name error: nlb name is empty.", mdl.NamespacedName.String())
		return nil
	}
	klog.Infof("[%s] try to find nlb by name %s",
		mdl.NamespacedName, mdl.LoadBalancerAttribute.Name)
	req := &nlb.ListLoadBalancersRequest{}
	req.LoadBalancerNames = []*string{tea.String(mdl.LoadBalancerAttribute.Name)}
	resp, err := p.auth.NLB.ListLoadBalancers(req)
	if err != nil {
		return fmt.Errorf("[%s] find loadbalancer by name %s error: %s", mdl.NamespacedName,
			mdl.LoadBalancerAttribute.Name, util.SDKError("ListLoadBalancers", err))
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI ListLoadBalancers resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "ListLoadBalancers")

	num := len(resp.Body.LoadBalancers)
	if num == 0 {
		return nil
	}

	if num > 1 {
		var lbIds []string
		for _, lb := range resp.Body.LoadBalancers {
			lbIds = append(lbIds, tea.StringValue(lb.LoadBalancerId))
		}
		return fmt.Errorf("[%s] find multiple loadbalances by name, lbIds[%s]", mdl.NamespacedName,
			strings.Join(lbIds, ","))
	}

	if tea.StringValue(resp.Body.LoadBalancers[0].LoadBalancerStatus) == string(Active) {
		return loadResponse(resp.Body.LoadBalancers[0], mdl)
	}

	getResp, err := p.waitNLBActive(tea.StringValue(resp.Body.LoadBalancers[0].LoadBalancerId))
	if err != nil {
		return err
	}
	return loadResponse(getResp.Body, mdl)
}

func loadResponse(resp interface{}, lb *nlbmodel.NetworkLoadBalancer) error {
	switch resp := resp.(type) {
	case *nlb.GetLoadBalancerAttributeResponseBody:
		lb.LoadBalancerAttribute.LoadBalancerId = tea.StringValue(resp.LoadBalancerId)
		lb.LoadBalancerAttribute.Name = tea.StringValue(resp.LoadBalancerName)
		lb.LoadBalancerAttribute.AddressType = tea.StringValue(resp.AddressType)
		lb.LoadBalancerAttribute.AddressIpVersion = tea.StringValue(resp.AddressIpVersion)
		lb.LoadBalancerAttribute.LoadBalancerStatus = tea.StringValue(resp.LoadBalancerStatus)
		lb.LoadBalancerAttribute.ResourceGroupId = tea.StringValue(resp.ResourceGroupId)
		lb.LoadBalancerAttribute.DNSName = tea.StringValue(resp.DNSName)
		lb.LoadBalancerAttribute.SecurityGroupIds = tea.StringSliceValue(resp.SecurityGroupIds)

		for _, z := range resp.ZoneMappings {
			lb.LoadBalancerAttribute.ZoneMappings = append(lb.LoadBalancerAttribute.ZoneMappings,
				nlbmodel.ZoneMapping{
					ZoneId:    tea.StringValue(z.ZoneId),
					VSwitchId: tea.StringValue(z.VSwitchId),
				},
			)
		}

	case *nlb.ListLoadBalancersResponseBodyLoadBalancers:
		lb.LoadBalancerAttribute.LoadBalancerId = tea.StringValue(resp.LoadBalancerId)
		lb.LoadBalancerAttribute.Name = tea.StringValue(resp.LoadBalancerName)
		lb.LoadBalancerAttribute.AddressType = tea.StringValue(resp.AddressType)
		lb.LoadBalancerAttribute.AddressIpVersion = tea.StringValue(resp.AddressIpVersion)
		lb.LoadBalancerAttribute.LoadBalancerStatus = tea.StringValue(resp.LoadBalancerStatus)
		lb.LoadBalancerAttribute.ResourceGroupId = tea.StringValue(resp.ResourceGroupId)
		lb.LoadBalancerAttribute.DNSName = tea.StringValue(resp.DNSName)
		lb.LoadBalancerAttribute.SecurityGroupIds = tea.StringSliceValue(resp.SecurityGroupIds)

		for _, z := range resp.ZoneMappings {
			lb.LoadBalancerAttribute.ZoneMappings = append(lb.LoadBalancerAttribute.ZoneMappings,
				nlbmodel.ZoneMapping{
					ZoneId:    tea.StringValue(z.ZoneId),
					VSwitchId: tea.StringValue(z.VSwitchId),
				},
			)
		}
	default:
		return fmt.Errorf("[%T] type not supported", resp)
	}
	return nil
}

const (
	DefaultRetryInterval = 3 * time.Second
	DefaultRetryTimeout  = 30 * time.Second
)

func (p *NLBProvider) waitJobFinish(api, jobId string, args ...time.Duration) error {
	var interval, timeout time.Duration
	if len(args) < 2 {
		interval = DefaultRetryInterval
		timeout = DefaultRetryTimeout
	} else {
		interval = args[0]
		timeout = args[1]
	}
	var (
		resp   *nlb.GetJobStatusResponse
		retErr error
	)
	_ = wait.PollImmediate(interval, timeout, func() (bool, error) {
		req := &nlb.GetJobStatusRequest{}
		req.JobId = tea.String(jobId)
		resp, retErr = p.auth.NLB.GetJobStatus(req)
		if retErr != nil {
			retErr = util.SDKError(fmt.Sprintf("%s-GetJobStatus", api), retErr)
			return false, retErr
		}
		if resp == nil || resp.Body == nil {
			retErr = fmt.Errorf("OpenAPI %s GetJobStatus resp is nil, JobId: %s", api, jobId)
			return false, nil
		}

		retErr = nil
		return tea.StringValue(resp.Body.Status) == "Succeeded", retErr
	})
	return retErr
}

func (p *NLBProvider) waitNLBActive(lbId string) (*nlb.GetLoadBalancerAttributeResponse, error) {
	var (
		retErr error
		resp   *nlb.GetLoadBalancerAttributeResponse
	)
	_ = wait.PollImmediate(20*time.Second, 2*time.Minute, func() (bool, error) {
		req := &nlb.GetLoadBalancerAttributeRequest{}
		req.LoadBalancerId = tea.String(lbId)

		resp, retErr = p.auth.NLB.GetLoadBalancerAttribute(req)
		if retErr != nil {
			retErr = util.SDKError("GetLoadBalancerAttribute", retErr)
			return false, retErr
		}

		if resp == nil || resp.Body == nil {
			retErr = fmt.Errorf("nlbId %s GetLoadBalancerAttribute response is nil, resp [%+v]", lbId, resp)
			return false, retErr
		}

		if tea.StringValue(resp.Body.LoadBalancerStatus) == string(Provisioning) {
			klog.V(5).Info("wait nlb to be active... ", "NLBId", lbId)
			retErr = fmt.Errorf("nlb %s is in creating status", lbId)
			return false, nil
		}

		retErr = nil
		return true, retErr
	})

	return resp, retErr
}

// NLBRegionIds used for e2etest
func (p *NLBProvider) NLBRegionIds() ([]string, error) {
	req := &nlb.DescribeRegionsRequest{}

	resp, err := p.auth.NLB.DescribeRegions(req)
	if err != nil {
		return nil, fmt.Errorf("describe nlb regions error: %s", err.Error())
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("OpenAPI DescribeRegions resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DescribeRegions")

	var ids []string
	for _, r := range resp.Body.Regions {
		if r.RegionId != nil {
			ids = append(ids, *r.RegionId)
		}
	}

	return ids, nil
}

// NLBZoneIds used for e2etest
func (p *NLBProvider) NLBZoneIds(regionId string) ([]string, error) {
	req := &nlb.DescribeZonesRequest{}
	req.RegionId = tea.String(regionId)

	resp, err := p.auth.NLB.DescribeZones(req)
	if err != nil {
		return nil, fmt.Errorf("describe nlb zones error: %s", err.Error())
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("OpenAPI DescribeZones resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "DescribeZones")

	var ids []string
	for _, z := range resp.Body.Zones {
		if z.ZoneId != nil {
			ids = append(ids, *z.ZoneId)
		}
	}

	return ids, nil
}

// UntagNLBResources used for e2etest
func (p *NLBProvider) UntagNLBResources(ctx context.Context, lbId string, tagKey []*string) error {
	req := &nlb.UntagResourcesRequest{}
	req.ResourceId = []*string{&lbId}
	req.ResourceType = tea.String("loadbalancer")
	req.TagKey = tagKey

	resp, err := p.auth.NLB.UntagResources(req)
	if err != nil {
		return util.SDKError("UntagResources", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("OpenAPI UntagResources resp is nil")
	}
	klog.V(5).Infof("RequestId: %s, API: %s", tea.StringValue(resp.Body.RequestId), "UntagResources")
	return nil
}
