package clbv1

import (
	"fmt"
	"github.com/aliyun/credentials-go/credentials/utils"
	cmap "github.com/orcaman/concurrent-map"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrlcfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"strconv"
)

const MaxLBTagNum = 10

func NewLoadBalancerManager(cloud prvd.Provider) *LoadBalancerManager {
	return &LoadBalancerManager{
		cloud:      cloud,
		tokenCache: cmap.New(),
	}
}

type LoadBalancerManager struct {
	cloud      prvd.Provider
	tokenCache cmap.ConcurrentMap
}

func (mgr *LoadBalancerManager) Find(reqCtx *svcCtx.RequestContext, mdl *model.LoadBalancer) error {
	// 1. set load balancer id
	if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(annotation.LoadBalancerId)
	}

	// 2. set default loadbalancer name
	// it's safe to set loadbalancer name which will be overwritten in FindLoadBalancer func
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.Anno.GetDefaultLoadBalancerName()

	// 3. set default loadbalancer tag
	// filter tags using logic operator OR, so only TAGKEY tag can be added
	mdl.LoadBalancerAttribute.Tags = []tag.Tag{
		{
			Key:   helper.TAGKEY,
			Value: reqCtx.Anno.GetDefaultLoadBalancerName(),
		},
	}
	return mgr.cloud.FindLoadBalancer(reqCtx.Ctx, mdl)
}

func (mgr *LoadBalancerManager) Create(reqCtx *svcCtx.RequestContext, local *model.LoadBalancer) error {
	if err := setModelDefaultValue(mgr, local, reqCtx.Anno); err != nil {
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

	err := mgr.cloud.CreateLoadBalancer(reqCtx.Ctx, local, clientToken)
	if err != nil {
		return fmt.Errorf("create slb error: %s", err.Error())
	}

	return nil
}

func (mgr *LoadBalancerManager) Delete(reqCtx *svcCtx.RequestContext, remote *model.LoadBalancer) error {
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	// set delete protection off
	if remote.LoadBalancerAttribute.DeleteProtection == model.OnFlag {
		if err := mgr.cloud.SetLoadBalancerDeleteProtection(
			reqCtx.Ctx,
			remote.LoadBalancerAttribute.LoadBalancerId,
			string(model.OffFlag),
		); err != nil {
			return fmt.Errorf("error to set slb id [%s] delete protection off, svc [%s], err: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, remote.NamespacedName, err.Error())
		}
	}

	key := reqCtx.Anno.GetDefaultLoadBalancerName()
	mgr.tokenCache.Remove(key)
	return mgr.cloud.DeleteLoadBalancer(reqCtx.Ctx, remote)

}

func (mgr *LoadBalancerManager) Update(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) error {
	lbId := remote.LoadBalancerAttribute.LoadBalancerId
	reqCtx.Log.Info(fmt.Sprintf("found load balancer [%s], try to update load balancer attribute", lbId))
	errs := []error{}

	// update tag
	var lbTag tag.Tag
	if local.LoadBalancerAttribute.IsUserManaged {
		lbTag = tag.Tag{Key: helper.REUSEKEY, Value: "true"}
	} else {
		lbTag = tag.Tag{Key: helper.TAGKEY, Value: reqCtx.Anno.GetDefaultLoadBalancerName()}
	}
	if err := mgr.addTagIfNotExist(reqCtx, *remote, lbTag); err != nil {
		errs = append(errs, fmt.Errorf("AddTags: %s", err.Error()))
	}

	if local.LoadBalancerAttribute.MasterZoneId != "" &&
		local.LoadBalancerAttribute.MasterZoneId != remote.LoadBalancerAttribute.MasterZoneId {
		errs = append(errs, fmt.Errorf("alicloud: can not change LoadBalancer master zone id once created"))
	}
	if local.LoadBalancerAttribute.SlaveZoneId != "" &&
		local.LoadBalancerAttribute.SlaveZoneId != remote.LoadBalancerAttribute.SlaveZoneId {
		errs = append(errs, fmt.Errorf("alicloud: can not change LoadBalancer slave zone id once created"))
	}
	if local.LoadBalancerAttribute.AddressType != "" &&
		local.LoadBalancerAttribute.AddressType != remote.LoadBalancerAttribute.AddressType {
		errs = append(errs, fmt.Errorf("alicloud: can not change LoadBalancer AddressType once created. delete and retry"))
	}
	if !equalsAddressIPVersion(local.LoadBalancerAttribute.AddressIPVersion, remote.LoadBalancerAttribute.AddressIPVersion) {
		errs = append(errs, fmt.Errorf("alicloud: can not change LoadBalancer AddressIPVersion once created"))
	}
	if local.LoadBalancerAttribute.ResourceGroupId != "" &&
		local.LoadBalancerAttribute.ResourceGroupId != remote.LoadBalancerAttribute.ResourceGroupId {
		errs = append(errs, fmt.Errorf("alicloud: can not change ResourceGroupId once created"))
	}
	if local.LoadBalancerAttribute.Address != "" &&
		local.LoadBalancerAttribute.Address != remote.LoadBalancerAttribute.Address {
		errs = append(errs, fmt.Errorf("alicloud: can not change LoadBalancer address once created"))
	}

	// update instanceChargeType && instanceSpec
	if err := mgr.updateInstanceChargeTypeAndInstanceSpec(reqCtx, local, remote); err != nil {
		errs = append(errs, fmt.Errorf("updateInstanceChargeTypeAndInstanceSpec error: %s", err.Error()))
	}

	// update internet chargeType & bandwidth
	needUpdate, charge, bandwidth := false, remote.LoadBalancerAttribute.InternetChargeType, remote.LoadBalancerAttribute.Bandwidth
	if local.LoadBalancerAttribute.InternetChargeType != "" &&
		local.LoadBalancerAttribute.InternetChargeType != remote.LoadBalancerAttribute.InternetChargeType {
		needUpdate = true
		charge = local.LoadBalancerAttribute.InternetChargeType
		reqCtx.Log.Info(fmt.Sprintf("update lb: internet chargeType changed([%s] - [%s])",
			remote.LoadBalancerAttribute.InternetChargeType, local.LoadBalancerAttribute.InternetChargeType),
			"lbId", lbId)
	}
	if local.LoadBalancerAttribute.Bandwidth != 0 &&
		local.LoadBalancerAttribute.Bandwidth != remote.LoadBalancerAttribute.Bandwidth &&
		local.LoadBalancerAttribute.InternetChargeType == model.PayByBandwidth {
		needUpdate = true
		bandwidth = local.LoadBalancerAttribute.Bandwidth
		reqCtx.Log.Info(fmt.Sprintf("update lb: bandwidth changed([%d] - [%d])",
			remote.LoadBalancerAttribute.Bandwidth, local.LoadBalancerAttribute.Bandwidth),
			"lbId", lbId)
	}

	if needUpdate {
		if remote.LoadBalancerAttribute.AddressType == model.InternetAddressType {
			reqCtx.Log.Info(fmt.Sprintf("update lb: modify loadbalancer: chargeType=%s, bandwidth=%d", charge, bandwidth),
				"lbId", lbId)
			if err := mgr.cloud.ModifyLoadBalancerInternetSpec(reqCtx.Ctx, lbId, string(charge), bandwidth); err != nil {
				errs = append(errs, fmt.Errorf("ModifyLoadBalancerInternetSpec: %s", err.Error()))
			}
		} else {
			reqCtx.Log.Info("update lb: only internet loadbalancer is allowed to modify bandwidth and pay type",
				"lbId", lbId)
		}
	}

	// update slb delete protection
	if local.LoadBalancerAttribute.DeleteProtection != "" &&
		local.LoadBalancerAttribute.DeleteProtection != remote.LoadBalancerAttribute.DeleteProtection {
		reqCtx.Log.Info(fmt.Sprintf("update lb: delete protection changed([%s] - [%s])",
			remote.LoadBalancerAttribute.DeleteProtection, local.LoadBalancerAttribute.DeleteProtection),
			"lbId", lbId)
		if err := mgr.cloud.SetLoadBalancerDeleteProtection(reqCtx.Ctx, lbId,
			string(local.LoadBalancerAttribute.DeleteProtection)); err != nil {
			errs = append(errs, fmt.Errorf("SetLoadBalancerDeleteProtection: %s", err.Error()))
		}
	}

	// update modification protection
	if local.LoadBalancerAttribute.ModificationProtectionStatus != "" &&
		local.LoadBalancerAttribute.ModificationProtectionStatus != remote.LoadBalancerAttribute.ModificationProtectionStatus {
		reqCtx.Log.WithName(lbId).Info(fmt.Sprintf("update lb: loadbalancer modification protection changed([%s] - [%s])",
			remote.LoadBalancerAttribute.ModificationProtectionStatus,
			local.LoadBalancerAttribute.ModificationProtectionStatus),
			"lbId", lbId)
		if err := mgr.cloud.SetLoadBalancerModificationProtection(reqCtx.Ctx, lbId,
			string(local.LoadBalancerAttribute.ModificationProtectionStatus)); err != nil {
			errs = append(errs, fmt.Errorf("SetLoadBalancerModificationProtection: %s", err.Error()))
		}
	}

	// update slb name
	// only user defined slb or slb which has "kubernetes.do.not.delete" tag can update name
	if local.LoadBalancerAttribute.LoadBalancerName != "" &&
		local.LoadBalancerAttribute.LoadBalancerName != remote.LoadBalancerAttribute.LoadBalancerName {
		reqCtx.Log.Info(fmt.Sprintf("update lb: loadbalancer name changed([%s]-[%s])",
			remote.LoadBalancerAttribute.LoadBalancerName, local.LoadBalancerAttribute.LoadBalancerName),
			"lbId", lbId)
		if err := mgr.cloud.SetLoadBalancerName(reqCtx.Ctx, lbId,
			local.LoadBalancerAttribute.LoadBalancerName); err != nil {
			errs = append(errs, fmt.Errorf("SetLoadBalancerName: %s", err.Error()))
		}
	}

	// update additional tags
	if len(local.LoadBalancerAttribute.Tags) != 0 {
		if err := mgr.updateLoadBalancerTags(reqCtx, local, remote); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

// Build build load balancer attribute for local model
func (mgr *LoadBalancerManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *model.LoadBalancer) error {
	mdl.LoadBalancerAttribute.AddressType = model.AddressType(reqCtx.Anno.Get(annotation.AddressType))
	mdl.LoadBalancerAttribute.InternetChargeType = model.InternetChargeType(reqCtx.Anno.Get(annotation.ChargeType))
	mdl.LoadBalancerAttribute.InstanceChargeType = model.InstanceChargeType(reqCtx.Anno.Get(annotation.InstanceChargeType))
	mdl.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(reqCtx.Anno.Get(annotation.Spec))
	bandwidth := reqCtx.Anno.Get(annotation.Bandwidth)
	if bandwidth != "" {
		i, err := strconv.Atoi(bandwidth)
		if err != nil {
			return fmt.Errorf("bandwidth must be integer, got [%s], error: %s", bandwidth, err.Error())
		}
		mdl.LoadBalancerAttribute.Bandwidth = i
	}
	mdl.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(reqCtx.Anno.Get(annotation.Spec))

	if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(annotation.LoadBalancerId)
		mdl.LoadBalancerAttribute.IsUserManaged = true
	}
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.Anno.Get(annotation.LoadBalancerName)
	mdl.LoadBalancerAttribute.VSwitchId = reqCtx.Anno.Get(annotation.VswitchId)
	mdl.LoadBalancerAttribute.MasterZoneId = reqCtx.Anno.Get(annotation.MasterZoneID)
	mdl.LoadBalancerAttribute.SlaveZoneId = reqCtx.Anno.Get(annotation.SlaveZoneID)
	mdl.LoadBalancerAttribute.ResourceGroupId = reqCtx.Anno.Get(annotation.ResourceGroupId)
	mdl.LoadBalancerAttribute.AddressIPVersion = model.AddressIPVersionType(reqCtx.Anno.Get(annotation.IPVersion))
	mdl.LoadBalancerAttribute.DeleteProtection = model.FlagType(reqCtx.Anno.Get(annotation.DeleteProtection))
	mdl.LoadBalancerAttribute.ModificationProtectionStatus =
		model.ModificationProtectionType(reqCtx.Anno.Get(annotation.ModificationProtection))
	mdl.LoadBalancerAttribute.Tags = reqCtx.Anno.GetLoadBalancerAdditionalTags()
	mdl.LoadBalancerAttribute.Address = reqCtx.Anno.Get(annotation.IP)
	if reqCtx.Anno.Get(annotation.PreserveLBOnDelete) != "" {
		mdl.LoadBalancerAttribute.PreserveOnDelete = true
	}
	return nil
}

func (mgr *LoadBalancerManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *model.LoadBalancer) error {
	return mgr.Find(reqCtx, mdl)
}

func (mgr *LoadBalancerManager) updateLoadBalancerTags(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) error {
	lbId := remote.LoadBalancerAttribute.LoadBalancerId

	localTags := helper.FilterTags(local.LoadBalancerAttribute.Tags, reqCtx.Anno.GetDefaultTags(), local.LoadBalancerAttribute.IsUserManaged)
	remoteTags := helper.FilterTags(remote.LoadBalancerAttribute.Tags, reqCtx.Anno.GetDefaultTags(), local.LoadBalancerAttribute.IsUserManaged)
	needTag, needUntag := helper.DiffLoadBalancerTags(localTags, remoteTags)
	if len(needTag) != 0 || len(needUntag) != 0 {
		reqCtx.Log.Info("tag changed", "lb", lbId, "needTag", needTag, "needUntag", needUntag)
	}

	if len(needTag) != 0 {
		if err := mgr.cloud.TagCLBResource(reqCtx.Ctx, lbId, needTag); err != nil {
			return fmt.Errorf("error to tag slb id [%s] with tags %v, svc [%s], err: %s",
				lbId, needTag, remote.NamespacedName, err.Error())
		}
	}

	if len(needUntag) != 0 {
		var untags []string
		for _, t := range needUntag {
			untags = append(untags, t.Key)
		}
		if err := mgr.cloud.UntagResources(reqCtx.Ctx, lbId, &untags); err != nil {
			return fmt.Errorf("error to untag slb id [%s] with tags %v, svc [%s], err: %s",
				lbId, needUntag, remote.NamespacedName, err.Error())
		}
	}
	return nil
}

// SetProtectionsOff turns off modification protection and deletion protection of load balancer
func (mgr *LoadBalancerManager) SetProtectionsOff(reqCtx *svcCtx.RequestContext, remote *model.LoadBalancer) error {
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	if remote.LoadBalancerAttribute.DeleteProtection == model.OnFlag {
		if err := mgr.cloud.SetLoadBalancerDeleteProtection(
			reqCtx.Ctx,
			remote.LoadBalancerAttribute.LoadBalancerId,
			string(model.OffFlag),
		); err != nil {
			return fmt.Errorf("error to set slb id [%s] delete protection off, svc [%s], err: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, remote.NamespacedName, err.Error())
		}
	}
	if remote.LoadBalancerAttribute.ModificationProtectionStatus == model.ConsoleProtection {
		if err := mgr.cloud.SetLoadBalancerModificationProtection(reqCtx.Ctx,
			remote.LoadBalancerAttribute.LoadBalancerId,
			string(model.NonProtection)); err != nil {
			return fmt.Errorf("error to set slb id [%s] modification protection off, svc [%s], err: %s",
				remote.LoadBalancerAttribute.LoadBalancerId, remote.NamespacedName, err.Error())
		}
	}

	return nil
}

// CleanupLoadBalancerTags removes service-related tags from remote slb
func (mgr *LoadBalancerManager) CleanupLoadBalancerTags(reqCtx *svcCtx.RequestContext, remote *model.LoadBalancer) error {
	if remote.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	defaultTags := reqCtx.Anno.GetDefaultTags()
	var removedTags []string
	for _, r := range remote.LoadBalancerAttribute.Tags {
		for _, l := range defaultTags {
			if r.Key == l.Key && r.Value == l.Value {
				removedTags = append(removedTags, r.Key)
			}
		}
	}

	if len(defaultTags) == 0 {
		return nil
	}
	return mgr.cloud.UntagResources(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId, &removedTags)
}

func equalsAddressIPVersion(local, remote model.AddressIPVersionType) bool {
	if local == "" {
		local = model.IPv4
	}

	if remote == "" {
		remote = model.IPv4
	}
	return local == remote
}

func setModelDefaultValue(mgr *LoadBalancerManager, mdl *model.LoadBalancer, anno *annotation.AnnotationRequest) error {
	if mdl.LoadBalancerAttribute.AddressType == "" {
		mdl.LoadBalancerAttribute.AddressType = model.AddressType(anno.GetDefaultValue(annotation.AddressType))
	}

	if mdl.LoadBalancerAttribute.LoadBalancerName == "" {
		mdl.LoadBalancerAttribute.LoadBalancerName = anno.GetDefaultLoadBalancerName()
	}

	if mdl.LoadBalancerAttribute.AddressType == model.IntranetAddressType {
		vpcId, err := mgr.cloud.VpcID()
		if err != nil {
			return fmt.Errorf("get vpc id from metadata error: %s", err.Error())
		}
		mdl.LoadBalancerAttribute.VpcId = vpcId
		if mdl.LoadBalancerAttribute.VSwitchId == "" {
			vswId, err := mgr.cloud.VswitchID()
			if err != nil {
				return fmt.Errorf("get vsw id from metadata error: %s", err.Error())
			}
			mdl.LoadBalancerAttribute.VSwitchId = vswId
		}
	}

	if mdl.LoadBalancerAttribute.InstanceChargeType.IsPayBySpec() &&
		mdl.LoadBalancerAttribute.LoadBalancerSpec == "" {
		mdl.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(anno.GetDefaultValue(annotation.Spec))
	}

	if mdl.LoadBalancerAttribute.DeleteProtection == "" {
		mdl.LoadBalancerAttribute.DeleteProtection = model.FlagType(anno.GetDefaultValue(annotation.DeleteProtection))
	}

	if mdl.LoadBalancerAttribute.ModificationProtectionStatus == "" {
		mdl.LoadBalancerAttribute.ModificationProtectionStatus = model.ModificationProtectionType(anno.GetDefaultValue(annotation.ModificationProtection))
		mdl.LoadBalancerAttribute.ModificationProtectionReason = model.ModificationProtectionReason
	}

	if mdl.LoadBalancerAttribute.ResourceGroupId == "" {
		mdl.LoadBalancerAttribute.ResourceGroupId = ctrlcfg.CloudCFG.Global.ResourceGroupID
	}

	mdl.LoadBalancerAttribute.Tags = append(anno.GetDefaultTags(), mdl.LoadBalancerAttribute.Tags...)
	return nil
}

func (mgr *LoadBalancerManager) addTagIfNotExist(reqCtx *svcCtx.RequestContext, remote model.LoadBalancer, newTag tag.Tag) error {
	for _, tag := range remote.LoadBalancerAttribute.Tags {
		if tag.Key == newTag.Key {
			return nil
		}
	}
	if len(remote.LoadBalancerAttribute.Tags) < MaxLBTagNum {
		return mgr.cloud.TagCLBResource(reqCtx.Ctx, remote.LoadBalancerAttribute.LoadBalancerId, []tag.Tag{newTag})
	}
	util.ServiceLog.Info("warning: total tags more than 10, can not add more tags")
	return nil
}

func (mgr *LoadBalancerManager) updateInstanceChargeTypeAndInstanceSpec(reqCtx *svcCtx.RequestContext, local, remote *model.LoadBalancer) error {
	lbId := remote.GetLoadBalancerId()
	var instanceChargeType model.InstanceChargeType

	if local.LoadBalancerAttribute.InstanceChargeType != "" {
		if local.LoadBalancerAttribute.InstanceChargeType != remote.LoadBalancerAttribute.InstanceChargeType {
			spec := ""
			if local.LoadBalancerAttribute.InstanceChargeType.IsPayBySpec() {
				if local.LoadBalancerAttribute.LoadBalancerSpec == "" {
					spec = reqCtx.Anno.GetDefaultValue(annotation.Spec)
				} else {
					spec = string(local.LoadBalancerAttribute.LoadBalancerSpec)
				}
			}

			reqCtx.Log.Info(fmt.Sprintf("update lb: InstanceChargeType changed ([%s] - [%s],  spec: [%s])",
				remote.LoadBalancerAttribute.InstanceChargeType, local.LoadBalancerAttribute.InstanceChargeType, spec),
				"lbId", lbId)

			return mgr.cloud.ModifyLoadBalancerInstanceChargeType(reqCtx.Ctx, lbId,
				string(local.LoadBalancerAttribute.InstanceChargeType), spec)
		} else {
			instanceChargeType = local.LoadBalancerAttribute.InstanceChargeType
		}
	} else {
		instanceChargeType = remote.LoadBalancerAttribute.InstanceChargeType
	}

	// if chargeType is paybyspec, update instance spec
	if instanceChargeType.IsPayBySpec() {
		if local.LoadBalancerAttribute.LoadBalancerSpec != "" &&
			local.LoadBalancerAttribute.LoadBalancerSpec != remote.LoadBalancerAttribute.LoadBalancerSpec {
			reqCtx.Log.Info(fmt.Sprintf("update lb: loadbalancerSpec changed ([%s] - [%s])",
				remote.LoadBalancerAttribute.LoadBalancerSpec, local.LoadBalancerAttribute.LoadBalancerSpec),
				"lbId", lbId)
			return mgr.cloud.ModifyLoadBalancerInstanceSpec(reqCtx.Ctx, lbId,
				string(local.LoadBalancerAttribute.LoadBalancerSpec))
		}
	}
	return nil
}
