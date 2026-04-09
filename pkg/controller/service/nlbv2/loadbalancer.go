package nlbv2

import (
	"fmt"

	"strings"

	"github.com/aliyun/credentials-go/credentials/utils"
	"github.com/mohae/deepcopy"
	cmap "github.com/orcaman/concurrent-map"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	ecsmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/ecs"

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

const (
	errorCodeConflict                       = "Conflict"
	errorJoinSecurityGroupQuotaInsufficient = "QuotaInsufficient"
)

func NewNLBManager(cloud prvd.Provider) *NLBManager {
	return &NLBManager{
		cloud:      cloud,
		tokenCache: cmap.New(),
	}
}

type NLBManager struct {
	cloud      prvd.Provider
	tokenCache cmap.ConcurrentMap
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

	if enabled := reqCtx.Anno.Get(annotation.CrossZoneEnabled); enabled != "" {
		f, err := model.ParseFlagType(enabled)
		if err != nil {
			return fmt.Errorf("ParameterInvalid, cross zone enabled flag error: %s", err.Error())
		}
		mdl.LoadBalancerAttribute.CrossZoneEnabled = tea.Bool(f == model.OnFlag)
	}

	if len(reqCtx.Service.Spec.LoadBalancerSourceRanges) != 0 {
		if !mdl.LoadBalancerAttribute.IsUserManaged {
			mdl.LoadBalancerAttribute.SourceRanges = reqCtx.Service.Spec.LoadBalancerSourceRanges
		} else {
			reqCtx.Recorder.Eventf(reqCtx.Service, v1.EventTypeWarning, helper.FeatureNotSupported,
				"LoadBalancerSourceRanges is not supported for user-managed LB.")
		}
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

	key := reqCtx.Anno.GetDefaultLoadBalancerName()
	clientToken := ""
	if t, ok := mgr.tokenCache.Get(key); ok {
		clientToken = t.(string)
	} else {
		clientToken = utils.GetUUID()
		mgr.tokenCache.Set(key, clientToken)
	}

	err := mgr.cloud.CreateNLB(reqCtx.Ctx, mdl, clientToken)
	if err != nil {
		return err
	}

	return nil
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

	key := reqCtx.Anno.GetDefaultLoadBalancerName()
	mgr.tokenCache.Remove(key)
	return mgr.cloud.DeleteNLB(reqCtx.Ctx, mdl)
}

func (mgr *NLBManager) Update(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	local.LoadBalancerAttribute.LoadBalancerId = remote.LoadBalancerAttribute.LoadBalancerId
	reqCtx.Log.Info(fmt.Sprintf("found nlb [%s], try to update load balancer attribute", remote.LoadBalancerAttribute.LoadBalancerId))
	needUpdateAttribute := false
	update := &nlbmodel.NetworkLoadBalancer{
		LoadBalancerAttribute: deepcopy.Copy(local.LoadBalancerAttribute).(*nlbmodel.LoadBalancerAttribute),
	}
	var errs []error

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
				if (l.AllocationId == "" || l.AllocationId == r.AllocationId) &&
					(l.IPv4Addr == "" || l.IPv4Addr == r.IPv4Addr) {
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

	localSecurityGroupIds := local.LoadBalancerAttribute.SecurityGroupIds
	if local.LoadBalancerAttribute.SourceRangesSecurityGroupId != "" {
		localSecurityGroupIds = append(localSecurityGroupIds, local.LoadBalancerAttribute.SourceRangesSecurityGroupId)
	}
	if (local.LoadBalancerAttribute.SecurityGroupIds != nil || remote.AssociatedSecurityGroup != nil || len(localSecurityGroupIds) != 0) &&
		!util.IsStringSliceEqual(localSecurityGroupIds, remote.LoadBalancerAttribute.SecurityGroupIds) {
		reqCtx.Log.Info(fmt.Sprintf("SecurityGroupIds changed from %v to %v",
			remote.LoadBalancerAttribute.SecurityGroupIds, localSecurityGroupIds))
		// get difference
		var add, remove []string
		newMap := map[string]struct{}{}
		oldMap := map[string]struct{}{}
		for _, i := range localSecurityGroupIds {
			newMap[i] = struct{}{}
		}
		for _, i := range remote.LoadBalancerAttribute.SecurityGroupIds {
			oldMap[i] = struct{}{}
			if _, ok := newMap[i]; !ok {
				remove = append(remove, i)
			}
		}
		for _, i := range localSecurityGroupIds {
			if _, ok := oldMap[i]; !ok {
				add = append(add, i)
			}
		}

		reqCtx.Log.Info(fmt.Sprintf("security groups add %v, remove %v", add, remove))
		if err := mgr.UpdateSecurityGroupIds(reqCtx, local, add, remove); err != nil {
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
		update.LoadBalancerAttribute.Name = local.LoadBalancerAttribute.Name
		needUpdateAttribute = true
	}

	if local.LoadBalancerAttribute.CrossZoneEnabled != nil &&
		tea.BoolValue(local.LoadBalancerAttribute.CrossZoneEnabled) != tea.BoolValue(remote.LoadBalancerAttribute.CrossZoneEnabled) {
		reqCtx.Log.Info(fmt.Sprintf("CrossZoneEnabled changed from [%t] to [%t]",
			tea.BoolValue(remote.LoadBalancerAttribute.CrossZoneEnabled), tea.BoolValue(local.LoadBalancerAttribute.CrossZoneEnabled)))
		update.LoadBalancerAttribute.CrossZoneEnabled = local.LoadBalancerAttribute.CrossZoneEnabled
		needUpdateAttribute = true
	}

	if err := updateBandwidthPackageId(mgr, reqCtx, local, remote); err != nil {
		errs = append(errs, err)
	}

	if len(local.LoadBalancerAttribute.Tags) != 0 {
		if err := mgr.updateLoadBalancerTags(reqCtx, local, remote); err != nil {
			errs = append(errs, err)
		}
	}

	if needUpdateAttribute {
		if err := mgr.cloud.UpdateNLB(reqCtx.Ctx, update); err != nil {
			errs = append(errs, fmt.Errorf("UpdateNLBAttribute error: %s", err.Error()))
		}
	}

	return utilerrors.NewAggregate(errs)
}

func (mgr *NLBManager) updateLoadBalancerTags(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	lbId := remote.LoadBalancerAttribute.LoadBalancerId

	localTags := helper.FilterTags(local.LoadBalancerAttribute.Tags, reqCtx.Anno.GetDefaultTags(), local.LoadBalancerAttribute.IsUserManaged)
	remoteTags := helper.FilterTags(remote.LoadBalancerAttribute.Tags, reqCtx.Anno.GetDefaultTags(), local.LoadBalancerAttribute.IsUserManaged)
	needTag, needUntag := helper.DiffLoadBalancerTags(localTags, remoteTags)
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

func (mgr *NLBManager) UpdateSecurityGroupIds(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer, join, leave []string) error {
	retryAdd := false
	var errs []error
	if len(join) != 0 {
		err := helper.RetryOnErrorContains(errorCodeConflict, func() error {
			return mgr.cloud.NLBJoinSecurityGroup(reqCtx.Ctx, mdl.LoadBalancerAttribute.LoadBalancerId, join)
		})
		if err != nil {
			if strings.Contains(err.Error(), errorJoinSecurityGroupQuotaInsufficient) {
				retryAdd = true
			} else {
				errs = append(errs, err)
			}
		}
	}
	if len(leave) != 0 {
		err := mgr.cloud.NLBLeaveSecurityGroup(reqCtx.Ctx, mdl.LoadBalancerAttribute.LoadBalancerId, leave)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if retryAdd {
		err := helper.RetryOnErrorContains(errorCodeConflict, func() error {
			return mgr.cloud.NLBJoinSecurityGroup(reqCtx.Ctx, mdl.LoadBalancerAttribute.LoadBalancerId, join)
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func (mgr *NLBManager) FindAssociatedSecurityGroup(reqCtx *svcCtx.RequestContext) (*ecsmodel.SecurityGroup, error) {
	tags := []tag.Tag{
		{
			Key:   helper.TAGKEY,
			Value: reqCtx.Anno.GetDefaultLoadBalancerName(),
		},
	}

	sgs, err := mgr.cloud.DescribeSecurityGroups(reqCtx.Ctx, tags)
	if err != nil {
		return &ecsmodel.SecurityGroup{}, err
	}

	if len(sgs) == 0 {
		return nil, nil
	}
	if len(sgs) > 1 {
		return nil, fmt.Errorf("find associated security group error: expect 1, got %d", len(sgs))
	}
	if sgs[0].ID == "" {
		return nil, fmt.Errorf("find associated security group error: sg id is empty")
	}
	sg, err := mgr.cloud.DescribeSecurityGroupAttribute(reqCtx.Ctx, sgs[0].ID)
	if err != nil {
		return nil, err
	}

	return &sg, nil
}

func (mgr *NLBManager) CreateAssociatedSecurityGroup(reqCtx *svcCtx.RequestContext, mdl *nlbmodel.NetworkLoadBalancer) error {
	tags := []tag.Tag{
		{
			Key:   helper.TAGKEY,
			Value: reqCtx.Anno.GetDefaultLoadBalancerName(),
		},
		{
			Key:   util.ClusterTagKey,
			Value: ctrlCfg.CloudCFG.Global.ClusterID,
		},
	}

	vpcId := ctrlCfg.CloudCFG.Global.VpcID
	if mdl.LoadBalancerAttribute.VpcId != "" {
		vpcId = mdl.LoadBalancerAttribute.VpcId
	}
	rgId := ctrlCfg.CloudCFG.Global.ResourceGroupID
	if mdl.LoadBalancerAttribute.ResourceGroupId != "" {
		rgId = mdl.LoadBalancerAttribute.ResourceGroupId
	}

	sg := ecsmodel.SecurityGroup{
		Name:            fmt.Sprintf("k8s-nlb-%s", reqCtx.Anno.GetDefaultLoadBalancerName()),
		Description:     getSecurityGroupRuleDescription(reqCtx.Service),
		Type:            ecsmodel.SecurityGroupTypeEnterprise,
		VpcID:           vpcId,
		ResourceGroupID: rgId,
		Permissions:     buildSecurityGroupPermissionsFromSourceRanges(reqCtx, mdl.LoadBalancerAttribute.SourceRanges),
		Tags:            tags,
	}

	if err := mgr.cloud.CreateSecurityGroup(reqCtx.Ctx, sg); err != nil {
		return fmt.Errorf("error to create security group for nlb, svc [%s], err: %w",
			util.Key(reqCtx.Anno.Service), err)
	}

	return nil
}

func (mgr *NLBManager) UpdateAssociatedSecurityGroup(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	var errs []error
	if local.LoadBalancerAttribute.SourceRangesSecurityGroupId == "" || remote.AssociatedSecurityGroup == nil {
		return nil
	}

	needUpdate := false
	update := deepcopy.Copy(remote.AssociatedSecurityGroup).(*ecsmodel.SecurityGroup)
	desc := getSecurityGroupRuleDescription(reqCtx.Service)
	if update.Description != desc {
		needUpdate = true
		update.Description = desc
	}
	if update.Name != getSecurityGroupName(reqCtx) {
		needUpdate = true
		update.Name = getSecurityGroupName(reqCtx)
	}

	if needUpdate {
		if err := mgr.cloud.ModifySecurityGroupAttribute(reqCtx.Ctx, update.ID, update); err != nil {
			errs = append(errs, err)
		}
	}

	if err := mgr.updateAssociatedSecurityGroupRules(reqCtx, local, remote); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

func (mgr *NLBManager) DeleteAssociatedSecurityGroup(reqCtx *svcCtx.RequestContext, remote *nlbmodel.NetworkLoadBalancer) error {
	if remote.AssociatedSecurityGroup == nil || remote.AssociatedSecurityGroup.ID == "" {
		return nil
	}
	if err := mgr.cloud.DeleteSecurityGroup(reqCtx.Ctx, remote.AssociatedSecurityGroup.ID); err != nil {
		return fmt.Errorf("error to delete associated security group, sgId: %s, svc: %s, err: %w",
			remote.AssociatedSecurityGroup.ID, remote.NamespacedName, err)
	}
	return nil
}

func (mgr *NLBManager) updateAssociatedSecurityGroupRules(reqCtx *svcCtx.RequestContext, local, remote *nlbmodel.NetworkLoadBalancer) error {
	if local.LoadBalancerAttribute.SourceRangesSecurityGroupId == "" || remote.AssociatedSecurityGroup == nil {
		return nil
	}

	localGroups := buildSecurityGroupPermissionsFromSourceRanges(reqCtx, local.LoadBalancerAttribute.SourceRanges)
	toAdd, toDelete, toUpdate := diffAssociatedSecurityGroupPermissions(reqCtx, localGroups, remote.AssociatedSecurityGroup.Permissions)

	if len(toAdd) != 0 || len(toDelete) != 0 || len(toUpdate) != 0 {
		reqCtx.Log.Info("security group rules changed, need to update", "toAdd", toAdd, "toDelete", toDelete, "toUpdate", toUpdate)
	}
	retryAdd := false
	var errs []error
	if len(toAdd) != 0 {
		err := mgr.cloud.AuthorizeSecurityGroup(reqCtx.Ctx, local.LoadBalancerAttribute.SourceRangesSecurityGroupId, toAdd)
		if err != nil {
			if strings.Contains(err.Error(), "AuthorizationLimitExceed") {
				// quota exceeded, retry after delete
				retryAdd = true
			} else {
				errs = append(errs, err)
			}
		}
	}
	if len(toUpdate) != 0 {
		for _, u := range toUpdate {
			err := mgr.cloud.ModifySecurityGroupRule(reqCtx.Ctx, local.LoadBalancerAttribute.SourceRangesSecurityGroupId, u)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(toDelete) != 0 {
		err := mgr.cloud.RevokeSecurityGroup(reqCtx.Ctx, local.LoadBalancerAttribute.SourceRangesSecurityGroupId, toDelete)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if retryAdd {
		err := mgr.cloud.AuthorizeSecurityGroup(reqCtx.Ctx, local.LoadBalancerAttribute.SourceRangesSecurityGroupId, toAdd)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
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

func buildSecurityGroupPermissionsFromSourceRanges(reqCtx *svcCtx.RequestContext, sourceRanges []string) []ecsmodel.SecurityGroupPermission {
	var permissions []ecsmodel.SecurityGroupPermission
	description := getSecurityGroupRuleDescription(reqCtx.Service)
	for _, sourceRange := range sourceRanges {
		permissions = append(permissions, ecsmodel.SecurityGroupPermission{
			Policy:       ecsmodel.SecurityGroupPolicyAccept,
			SourceCidrIp: sourceRange,
			Priority:     "1",
			IpProtocol:   ecsmodel.SecurityGroupIpProtocolAll,
			PortRange:    "-1/-1",
			Description:  description,
		})
	}
	// add default drop policy
	permissions = append(permissions, ecsmodel.SecurityGroupPermission{
		Policy:       ecsmodel.SecurityGroupPolicyDrop,
		SourceCidrIp: "0.0.0.0/0",
		Priority:     "100",
		IpProtocol:   ecsmodel.SecurityGroupIpProtocolAll,
		PortRange:    "-1/-1",
		Description:  description,
	})

	return permissions
}

func diffAssociatedSecurityGroupPermissions(reqCtx *svcCtx.RequestContext, local, remote []ecsmodel.SecurityGroupPermission) ([]ecsmodel.SecurityGroupPermission, []ecsmodel.SecurityGroupPermission, []ecsmodel.SecurityGroupPermission) {
	var toAdd, toDelete, toUpdate []ecsmodel.SecurityGroupPermission
	for _, l := range local {
		found := false
		for _, r := range remote {
			if l.SourceCidrIp == r.SourceCidrIp && strings.EqualFold(l.Policy, r.Policy) {
				found = true
				if l.Description != r.Description ||
					l.Priority != r.Priority ||
					!strings.EqualFold(l.IpProtocol, r.IpProtocol) ||
					l.PortRange != r.PortRange {
					l.SecurityGroupRuleId = r.SecurityGroupRuleId
					toUpdate = append(toUpdate, l)
				}
				break
			}
		}
		if !found {
			toAdd = append(toAdd, l)
		}
	}
	desc := getSecurityGroupRuleDescription(reqCtx.Service)
	for _, r := range remote {
		found := false
		if r.Description != desc {
			reqCtx.Log.Info("security group rule is not managed by cluster, skip", "rule", r.SecurityGroupRuleId)
			continue
		}
		for _, l := range local {
			if r.SourceCidrIp == l.SourceCidrIp && strings.EqualFold(r.Policy, l.Policy) {
				found = true
				break
			}
		}
		if !found {
			toDelete = append(toDelete, r)
		}
	}
	return toAdd, toDelete, toUpdate
}

func getSecurityGroupName(reqCtx *svcCtx.RequestContext) string {
	return fmt.Sprintf("k8s-nlb-%s", reqCtx.Anno.GetDefaultLoadBalancerName())
}
func getSecurityGroupRuleDescription(svc *v1.Service) string {
	return fmt.Sprintf("k8s.%s.%s.%s", svc.Namespace, svc.Name, ctrlCfg.CloudCFG.Global.ClusterID)
}
