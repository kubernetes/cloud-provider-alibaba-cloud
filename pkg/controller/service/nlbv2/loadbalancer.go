package nlbv2

import (
	"fmt"
	"strings"

	"github.com/alibabacloud-go/tea/tea"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
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

	if reqCtx.Anno.Has(annotation.SecurityGroupIds) {
		mdl.LoadBalancerAttribute.SecurityGroupIds = []string{}
		anno := reqCtx.Anno.Get(annotation.SecurityGroupIds)
		if anno != "" {
			mdl.LoadBalancerAttribute.SecurityGroupIds = strings.Split(anno, ",")
		}
	}

	mdl.LoadBalancerAttribute.AddressType = nlbmodel.GetAddressType(reqCtx.Anno.Get(annotation.AddressType))

	mdl.LoadBalancerAttribute.ResourceGroupId = reqCtx.Anno.Get(annotation.ResourceGroupId)

	mdl.LoadBalancerAttribute.AddressIpVersion = nlbmodel.GetAddressIpVersion(reqCtx.Anno.Get(annotation.IPVersion))
	mdl.LoadBalancerAttribute.IPv6AddressType = nlbmodel.GetAddressType(reqCtx.Anno.Get(annotation.IPv6AddressType))

	mdl.LoadBalancerAttribute.Name = reqCtx.Anno.Get(annotation.LoadBalancerName)

	mdl.LoadBalancerAttribute.Tags = reqCtx.Anno.GetLoadBalancerAdditionalTags()

	if reqCtx.Anno.Has(annotation.BandwidthPackageId) {
		mdl.LoadBalancerAttribute.BandwidthPackageId = tea.String(reqCtx.Anno.Get(annotation.BandwidthPackageId))
	}

	if reqCtx.Anno.Get(annotation.PreserveLBOnDelete) != "" {
		mdl.LoadBalancerAttribute.PreserveOnDelete = true
	}

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

	// disable delete protection
	err := mgr.cloud.UpdateLoadBalancerProtection(reqCtx.Ctx, mdl.LoadBalancerAttribute.LoadBalancerId,
		&nlbmodel.DeletionProtectionConfig{Enabled: false}, nil)
	if err != nil {
		return fmt.Errorf("disable delete protection error: %s", err.Error())
	}

	return mgr.cloud.DeleteNLB(reqCtx.Ctx, mdl)
}

func (mgr *NLBManager) Update(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	local.LoadBalancerAttribute.LoadBalancerId = remote.LoadBalancerAttribute.LoadBalancerId
	reqCtx.Log.Info(fmt.Sprintf("found nlb [%s], try to update load balancer attribute", remote.LoadBalancerAttribute.LoadBalancerId))
	errs := []error{}

	// immutable attributes
	if local.LoadBalancerAttribute.AddressIpVersion != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.AddressIpVersion, remote.LoadBalancerAttribute.AddressIpVersion) {
		errs = append(errs, fmt.Errorf("AddressIpVersion cannot be changed, service: %s, nlb: %s",
			local.LoadBalancerAttribute.AddressIpVersion, remote.LoadBalancerAttribute.AddressIpVersion))
	}

	if local.LoadBalancerAttribute.ResourceGroupId != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.ResourceGroupId, remote.LoadBalancerAttribute.ResourceGroupId) {
		errs = append(errs, fmt.Errorf("ResourceGroupId cannot be changed, service: %s, nlb: %s",
			local.LoadBalancerAttribute.ResourceGroupId, remote.LoadBalancerAttribute.ResourceGroupId))
	}

	// mutable
	if local.LoadBalancerAttribute.AddressType != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.AddressType, remote.LoadBalancerAttribute.AddressType) {
		reqCtx.Log.Info(fmt.Sprintf("AddressType changed from [%s] to [%s]",
			local.LoadBalancerAttribute.AddressType, remote.LoadBalancerAttribute.AddressType))
		if err := mgr.cloud.UpdateNLBAddressType(reqCtx.Ctx, local); err != nil {
			errs = append(errs, fmt.Errorf("UpdateNLBAddressType error: %s", err.Error()))
		}
	}

	for _, l := range local.LoadBalancerAttribute.ZoneMappings {
		match := false
		for _, r := range remote.LoadBalancerAttribute.ZoneMappings {
			if l.ZoneId == r.ZoneId && l.VSwitchId == r.VSwitchId {
				if l.AllocationId == "" || l.AllocationId != "" && l.AllocationId == r.AllocationId {
					match = true
					break
				}
			}
		}
		if !match {
			reqCtx.Log.Info(fmt.Sprintf("ZoneMappings changed from [%s] to [%s]",
				remote.LoadBalancerAttribute.ZoneMappings, local.LoadBalancerAttribute.ZoneMappings))
			if err := mgr.cloud.UpdateNLBZones(reqCtx.Ctx, local); err != nil {
				errs = append(errs, fmt.Errorf("update zone mappings error: %s", err.Error()))
			}
			break
		}
	}

	if local.LoadBalancerAttribute.SecurityGroupIds != nil &&
		!util.IsStringSliceEqual(local.LoadBalancerAttribute.SecurityGroupIds, remote.LoadBalancerAttribute.SecurityGroupIds) {
		reqCtx.Log.Info(fmt.Sprintf("SecurityGroupIds changed from %v to %v",
			remote.LoadBalancerAttribute.SecurityGroupIds, local.LoadBalancerAttribute.SecurityGroupIds))
		// get difference
		var added, removed []string
		newMap := map[string]struct{}{}
		oldMap := map[string]struct{}{}
		for _, i := range local.LoadBalancerAttribute.SecurityGroupIds {
			newMap[i] = struct{}{}
		}
		for _, i := range remote.LoadBalancerAttribute.SecurityGroupIds {
			oldMap[i] = struct{}{}
			if _, ok := newMap[i]; !ok {
				removed = append(removed, i)
			}
		}
		for _, i := range local.LoadBalancerAttribute.SecurityGroupIds {
			if _, ok := oldMap[i]; !ok {
				added = append(added, i)
			}
		}

		reqCtx.Log.Info(fmt.Sprintf("security groups added %v, removed %v", added, removed))
		if err := mgr.cloud.UpdateNLBSecurityGroupIds(reqCtx.Ctx, local, added, removed); err != nil {
			errs = append(errs, fmt.Errorf("update security group ids error: %s", err.Error()))
		}
	}

	if local.LoadBalancerAttribute.IPv6AddressType != "" &&
		!strings.EqualFold(local.LoadBalancerAttribute.IPv6AddressType, remote.LoadBalancerAttribute.IPv6AddressType) {
		reqCtx.Log.Info(fmt.Sprintf("IPv6AddressType changed from [%s] to [%s]",
			remote.LoadBalancerAttribute.IPv6AddressType, local.LoadBalancerAttribute.IPv6AddressType))
		if err := mgr.cloud.UpdateNLBIPv6AddressType(reqCtx.Ctx, local); err != nil {
			errs = append(errs, fmt.Errorf("UpdateNLBIPv6AddressType error: %s", err.Error()))
		}
	}

	if local.LoadBalancerAttribute.Name != "" &&
		local.LoadBalancerAttribute.Name != remote.LoadBalancerAttribute.Name {
		reqCtx.Log.Info(fmt.Sprintf("name changed from [%s] to [%s]",
			remote.LoadBalancerAttribute.Name, local.LoadBalancerAttribute.Name))
		errs = append(errs, mgr.cloud.UpdateNLB(reqCtx.Ctx, local))
	}

	if err := updateBandwidthPackageId(mgr, reqCtx, local, remote); err != nil {
		errs = append(errs, err)
	}

	if len(local.LoadBalancerAttribute.Tags) != 0 {
		if err := mgr.updateLoadBalancerTags(reqCtx, local, remote); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (mgr *NLBManager) updateLoadBalancerTags(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	var localTags, remoteTags []tag.Tag
	lbId := remote.LoadBalancerAttribute.LoadBalancerId
	defaultTags := reqCtx.Anno.GetDefaultTags()
	if local.LoadBalancerAttribute.IsUserManaged {
		defaultTags = append(defaultTags, tag.Tag{Key: helper.REUSEKEY, Value: "true"})
	}

	for _, r := range remote.LoadBalancerAttribute.Tags {
		found := false
		for _, d := range defaultTags {
			if r.Key == d.Key {
				found = true
				break
			}
		}
		if !found {
			remoteTags = append(remoteTags, r)
		}
	}

	for _, l := range local.LoadBalancerAttribute.Tags {
		found := false
		for _, d := range defaultTags {
			if l.Key == d.Key {
				found = true
				break
			}
		}
		if !found {
			localTags = append(localTags, l)
		}
	}

	needTag, needUntag := util.DiffLoadBalancerTags(local.LoadBalancerAttribute.Tags, remoteTags)
	if len(needTag) != 0 || len(needUntag) != 0 {
		reqCtx.Log.Info("tag changed", "lb", lbId, "needTag", needTag, "needUntag", needUntag)
	}

	if len(needTag) != 0 {
		if err := mgr.cloud.TagNLBResource(reqCtx.Ctx, lbId, nlbmodel.LoadBalancerTagType, needTag); err != nil {
			return fmt.Errorf("error to tag slb id [%s] with tags %v, svc [%s], err: %s",
				lbId, needTag, remote.NamespacedName, err.Error())
		}
	}

	if len(needUntag) != 0 {
		var untags []*string
		for _, t := range needUntag {
			untags = append(untags, tea.String(t.Key))
		}
		if err := mgr.cloud.UntagNLBResources(reqCtx.Ctx, lbId, nlbmodel.LoadBalancerTagType, untags); err != nil {
			return fmt.Errorf("error to untag slb id [%s] with tags %v, svc [%s], err: %s",
				lbId, needUntag, remote.NamespacedName, err.Error())
		}
	}

	return nil
}

func (mgr *NLBManager) SetProtectionsOff(reqCtx *svcCtx.RequestContext, remote *nlbmodel.NetworkLoadBalancer) error {
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	modCfg := &nlbmodel.ModificationProtectionConfig{Status: nlbmodel.NonProtection}
	delCfg := &nlbmodel.DeletionProtectionConfig{Enabled: false}
	if err := mgr.cloud.UpdateLoadBalancerProtection(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId, delCfg, modCfg); err != nil {
		return fmt.Errorf("error to set nlb id [%s] protections off, svc [%s], err: %s",
			remote.LoadBalancerAttribute.LoadBalancerId, remote.NamespacedName, err.Error())
	}

	return nil
}

func (mgr *NLBManager) CleanupLoadBalancerTags(reqCtx *svcCtx.RequestContext, remote *nlbmodel.NetworkLoadBalancer) error {
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	defaultTags := reqCtx.Anno.GetDefaultTags()
	var removedTags []*string
	for _, r := range remote.LoadBalancerAttribute.Tags {
		for _, l := range defaultTags {
			if l.Key == r.Key && l.Value == r.Value {
				removedTags = append(removedTags, tea.String(r.Key))
			}
		}
	}

	if len(removedTags) == 0 {
		return nil
	}
	return mgr.cloud.UntagNLBResources(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId, nlbmodel.LoadBalancerTagType, removedTags)
}

func updateBandwidthPackageId(mgr *NLBManager, reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	// local not set, return
	if local.LoadBalancerAttribute.BandwidthPackageId == nil {
		return nil
	}

	// local == remote, return
	if tea.StringValue(local.LoadBalancerAttribute.BandwidthPackageId) ==
		tea.StringValue(remote.LoadBalancerAttribute.BandwidthPackageId) {
		return nil
	}

	// local != remote
	// detach
	if tea.StringValue(remote.LoadBalancerAttribute.BandwidthPackageId) != "" {
		reqCtx.Log.Info(fmt.Sprintf("detach bandwidthPackageId [%s]",
			*remote.LoadBalancerAttribute.BandwidthPackageId))
		if err := mgr.cloud.DetachCommonBandwidthPackageFromLoadBalancer(reqCtx.Ctx,
			local.LoadBalancerAttribute.LoadBalancerId, *remote.LoadBalancerAttribute.BandwidthPackageId); err != nil {
			return err
		}
	}

	// attach
	if tea.StringValue(local.LoadBalancerAttribute.BandwidthPackageId) != "" {
		reqCtx.Log.Info(fmt.Sprintf("attach bandwidthPackageId [%s]",
			*local.LoadBalancerAttribute.BandwidthPackageId))
		if err := mgr.cloud.AttachCommonBandwidthPackageToLoadBalancer(reqCtx.Ctx,
			local.LoadBalancerAttribute.LoadBalancerId, *local.LoadBalancerAttribute.BandwidthPackageId); err != nil {
			return err
		}
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

	if mdl.LoadBalancerAttribute.ResourceGroupId == "" {
		mdl.LoadBalancerAttribute.ResourceGroupId = ctrlCfg.CloudCFG.Global.ResourceGroupID
	}

	if mdl.LoadBalancerAttribute.DeletionProtectionConfig == nil {
		mdl.LoadBalancerAttribute.DeletionProtectionConfig = &nlbmodel.DeletionProtectionConfig{
			Enabled: true,
		}
	}

	if mdl.LoadBalancerAttribute.ModificationProtectionConfig == nil {
		mdl.LoadBalancerAttribute.ModificationProtectionConfig = &nlbmodel.ModificationProtectionConfig{
			Status: nlbmodel.ConsoleProtection,
			Reason: nlbmodel.ModificationProtectionReason,
		}
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

	if len(ret) == 0 {
		return nil, fmt.Errorf("ZoneMapping format error, expect [zone-a:vsw-id-1,zone-b:vsw-id-2], got %s", zoneMaps)
	}
	return ret, nil
}
