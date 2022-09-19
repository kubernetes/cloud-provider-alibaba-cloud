package nlbv2

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"strings"
)

func NewNLBManager(cloud prvd.Provider) *NLBManager {
	return &NLBManager{
		cloud: cloud,
	}
}

type NLBManager struct {
	cloud prvd.Provider
}

func (mgr *NLBManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(annotation.LoadBalancerId)
		mdl.LoadBalancerAttribute.IsUserManaged = true
	}

	if reqCtx.Anno.Get(annotation.ZoneMaps) != "" {
		zoneMappings, err := parseZoneMappings(reqCtx.Anno.Get(annotation.ZoneMaps))
		if err != nil {
			return err
		}
		mdl.LoadBalancerAttribute.ZoneMappings = zoneMappings
	} else if !mdl.LoadBalancerAttribute.IsUserManaged {
		return fmt.Errorf("ParameterMissing, zone mappings are required")
	}

	mdl.LoadBalancerAttribute.AddressType = nlbmodel.GetAddressType(reqCtx.Anno.Get(annotation.AddressType))

	mdl.LoadBalancerAttribute.ResourceGroupId = reqCtx.Anno.Get(annotation.ResourceGroupId)

	mdl.LoadBalancerAttribute.AddressIpVersion = nlbmodel.GetAddressIpVersion(reqCtx.Anno.Get(annotation.IPVersion))

	mdl.LoadBalancerAttribute.Name = reqCtx.Anno.Get(annotation.LoadBalancerName)

	mdl.LoadBalancerAttribute.Tags = reqCtx.Anno.GetLoadBalancerAdditionalTags()
	return nil
}

func (mgr *NLBManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	return mgr.Find(reqCtx, mdl)
}

func (mgr *NLBManager) Find(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	if mdl.LoadBalancerAttribute == nil {
		mdl.LoadBalancerAttribute = &nlbmodel.LoadBalancerAttribute{}
	}

	// 1. set nlb id
	if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(annotation.LoadBalancerId)
	}

	// 2. set default loadbalancer name
	// it's safe to set loadbalancer name which will be overwritten in FindLoadBalancer func
	mdl.LoadBalancerAttribute.Name = reqCtx.Anno.GetDefaultLoadBalancerName()

	// 3. set default loadbalancer tag
	// filter tags using logic operator OR, so only TAGKEY tag can be added
	mdl.LoadBalancerAttribute.Tags = []tag.Tag{
		{
			Key:   helper.TAGKEY,
			Value: reqCtx.Anno.GetDefaultLoadBalancerName(),
		},
	}

	return mgr.cloud.FindNLB(reqCtx.Ctx, mdl)

}

func (mgr *NLBManager) Create(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	if err := setDefaultValueForLoadBalancer(mgr, mdl, reqCtx.Anno); err != nil {
		return fmt.Errorf("set model default value error: %s", err.Error())
	}

	err := mgr.cloud.CreateNLB(reqCtx.Ctx, mdl)
	if err != nil {
		return err
	}

	return mgr.cloud.TagNLBResource(reqCtx.Ctx, mdl.LoadBalancerAttribute.LoadBalancerId,
		nlbmodel.LoadBalancerTagType, mdl.LoadBalancerAttribute.Tags)
}

func (mgr *NLBManager) Delete(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	if mdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	return mgr.cloud.DeleteNLB(reqCtx.Ctx, mdl)
}

func (mgr *NLBManager) Update(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	local.LoadBalancerAttribute.LoadBalancerId = remote.LoadBalancerAttribute.LoadBalancerId
	// immutable attributes
	if local.LoadBalancerAttribute.AddressIpVersion != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.AddressIpVersion, remote.LoadBalancerAttribute.AddressIpVersion) {
		return fmt.Errorf("AddressIpVersion cannot be changed, service: %s, nlb: %s",
			local.LoadBalancerAttribute.AddressIpVersion, remote.LoadBalancerAttribute.AddressIpVersion)
	}

	if local.LoadBalancerAttribute.ResourceGroupId != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.ResourceGroupId, remote.LoadBalancerAttribute.ResourceGroupId) {
		return fmt.Errorf("ResourceGroupId cannot be changed, service: %s, nlb: %s",
			local.LoadBalancerAttribute.ResourceGroupId, remote.LoadBalancerAttribute.ResourceGroupId)
	}

	// mutable
	if local.LoadBalancerAttribute.AddressType != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.AddressType, remote.LoadBalancerAttribute.AddressType) {
		reqCtx.Log.Info(fmt.Sprintf("AddressType changed from [%s] to [%s]",
			local.LoadBalancerAttribute.AddressType, remote.LoadBalancerAttribute.AddressType))
		if err := mgr.cloud.UpdateNLBAddressType(reqCtx.Ctx, local); err != nil {
			return fmt.Errorf("UpdateNLBAddressType error: %s", err.Error())
		}
	}

	needUpdate := false
	for _, l := range local.LoadBalancerAttribute.ZoneMappings {
		found := false
		for _, r := range remote.LoadBalancerAttribute.ZoneMappings {
			if l.ZoneId == r.ZoneId && l.VSwitchId == r.VSwitchId {
				found = true
				break
			}
		}
		if !found {
			needUpdate = true
			break
		}
	}
	if needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("ZoneMappings changed from [%s] to [%s]",
			remote.LoadBalancerAttribute.ZoneMappings, local.LoadBalancerAttribute.ZoneMappings))
		if err := mgr.cloud.UpdateNLBZones(reqCtx.Ctx, local); err != nil {
			return fmt.Errorf("update zone mappings error: %s", err.Error())
		}
	}

	needUpdate = false
	if local.LoadBalancerAttribute.Name != "" &&
		local.LoadBalancerAttribute.Name != remote.LoadBalancerAttribute.Name {
		reqCtx.Log.Info(fmt.Sprintf("name changed from [%s] to [%s]",
			remote.LoadBalancerAttribute.Name, local.LoadBalancerAttribute.Name))
		needUpdate = true
	}

	if needUpdate {
		return mgr.cloud.UpdateNLB(reqCtx.Ctx, local)
	}

	return nil
}

func setDefaultValueForLoadBalancer(mgr *NLBManager, mdl *nlbmodel.NetworkLoadBalancer, anno *annotation.AnnotationRequest,
) error {
	if mdl.LoadBalancerAttribute.AddressType == "" {
		mdl.LoadBalancerAttribute.AddressType = anno.GetDefaultValue(annotation.AddressType)
	}

	if mdl.LoadBalancerAttribute.Name == "" {
		mdl.LoadBalancerAttribute.Name = anno.GetDefaultLoadBalancerName()
	}

	if mdl.LoadBalancerAttribute.VpcId == "" {
		vpcId, err := mgr.cloud.VpcID()
		if err != nil {
			return fmt.Errorf("get vpc id error: %s", err.Error())
		}
		mdl.LoadBalancerAttribute.VpcId = vpcId
	}

	mdl.LoadBalancerAttribute.Tags = append(anno.GetDefaultTags(), mdl.LoadBalancerAttribute.Tags...)
	return nil
}

func parseZoneMappings(zoneMaps string) ([]nlbmodel.ZoneMapping, error) {
	var ret []nlbmodel.ZoneMapping
	attrs := strings.Split(zoneMaps, ",")
	for _, attr := range attrs {
		items := strings.Split(attr, ":")
		if len(items) < 2 {
			return nil, fmt.Errorf("ZoneMapping format error, expect [zone-a:vsw-id-1,zone-b:vsw-id-2], got %s", zoneMaps)
		}
		zoneMap := nlbmodel.ZoneMapping{
			ZoneId:    items[0],
			VSwitchId: items[1],
		}

		if len(items) > 2 {
			zoneMap.IPv4Addr = items[2]
		}

		if len(items) > 3 {
			zoneMap.AllocationId = items[3]
		}
		ret = append(ret, zoneMap)
	}

	if len(ret) < 0 {
		return nil, fmt.Errorf("ZoneMapping format error, expect [zone-a:vsw-id-1,zone-b:vsw-id-2], got %s", zoneMaps)
	}
	return ret, nil
}
