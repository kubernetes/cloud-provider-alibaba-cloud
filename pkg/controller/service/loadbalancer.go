package service

import (
	"encoding/json"
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"strconv"
)

func NewLoadBalancerManager(cloud prvd.Provider) *LoadBalancerManager {
	return &LoadBalancerManager{
		cloud: cloud,
	}
}

type LoadBalancerManager struct {
	cloud prvd.Provider
}

func (mgr *LoadBalancerManager) Find(reqCtx *RequestContext, mdl *model.LoadBalancer) error {
	// 1. set load balancer id
	if reqCtx.Anno.Get(LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(LoadBalancerId)
	}

	// 2. set default loadbalancer name
	// it's safe to set loadbalancer name which will be overwritten in FindLoadBalancer func
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.Anno.GetDefaultLoadBalancerName()

	// 3. set default loadbalancer tag
	// filter tags using logic operator OR, so only TAGKEY tag can be added
	mdl.LoadBalancerAttribute.Tags = []model.Tag{
		{
			TagKey:   TAGKEY,
			TagValue: reqCtx.Anno.GetDefaultLoadBalancerName(),
		},
	}
	return mgr.cloud.FindLoadBalancer(reqCtx.Ctx, mdl)
}

func (mgr *LoadBalancerManager) Create(reqCtx *RequestContext, local *model.LoadBalancer) error {
	if err := setModelDefaultValue(mgr, local, reqCtx.Anno); err != nil {
		return fmt.Errorf("set model default value error: %s", err.Error())
	}
	err := mgr.cloud.CreateLoadBalancer(reqCtx.Ctx, local)
	if err != nil {
		return fmt.Errorf("create slb error: %s", err.Error())
	}

	tags, err := json.Marshal(local.LoadBalancerAttribute.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags error: %s", err.Error())
	}
	return mgr.cloud.AddTags(reqCtx.Ctx, local.LoadBalancerAttribute.LoadBalancerId, string(tags))
}

func (mgr *LoadBalancerManager) Delete(reqCtx *RequestContext, remote *model.LoadBalancer) error {
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

	return mgr.cloud.DeleteLoadBalancer(reqCtx.Ctx, remote)

}

func (mgr *LoadBalancerManager) Update(reqCtx *RequestContext, local, remote *model.LoadBalancer) error {
	lbId := remote.LoadBalancerAttribute.LoadBalancerId
	reqCtx.Log.Info(fmt.Sprintf("found load balancer [%s], try to update load balancer attribute", lbId))

	if err := updateLBDefaultTag(mgr, reqCtx, local, remote); err != nil {
		return fmt.Errorf("alicloud: update lb tag error: %s", err.Error())
	}

	if local.LoadBalancerAttribute.MasterZoneId != "" &&
		local.LoadBalancerAttribute.MasterZoneId != remote.LoadBalancerAttribute.MasterZoneId {
		return fmt.Errorf("alicloud: can not change LoadBalancer master zone id once created")
	}
	if local.LoadBalancerAttribute.SlaveZoneId != "" &&
		local.LoadBalancerAttribute.SlaveZoneId != remote.LoadBalancerAttribute.SlaveZoneId {
		return fmt.Errorf("alicloud: can not change LoadBalancer slave zone id once created")
	}
	if local.LoadBalancerAttribute.AddressType != "" &&
		local.LoadBalancerAttribute.AddressType != remote.LoadBalancerAttribute.AddressType {
		return fmt.Errorf("alicloud: can not change LoadBalancer AddressType once created. delete and retry")
	}
	if !equalsAddressIPVersion(local.LoadBalancerAttribute.AddressIPVersion, remote.LoadBalancerAttribute.AddressIPVersion) {
		return fmt.Errorf("alicloud: can not change LoadBalancer AddressIPVersion once created")
	}
	if local.LoadBalancerAttribute.ResourceGroupId != "" &&
		local.LoadBalancerAttribute.ResourceGroupId != remote.LoadBalancerAttribute.ResourceGroupId {
		return fmt.Errorf("alicloud: can not change ResourceGroupId once created")
	}

	// update chargeType & bandwidth
	needUpdate, charge, bandwidth := false, remote.LoadBalancerAttribute.InternetChargeType, remote.LoadBalancerAttribute.Bandwidth
	if local.LoadBalancerAttribute.InternetChargeType != "" &&
		local.LoadBalancerAttribute.InternetChargeType != remote.LoadBalancerAttribute.InternetChargeType {
		needUpdate = true
		charge = local.LoadBalancerAttribute.InternetChargeType
		reqCtx.Log.WithName(lbId).Info(fmt.Sprintf("update lb: internet chargeType changed([%s] -> [%s])",
			remote.LoadBalancerAttribute.InternetChargeType, local.LoadBalancerAttribute.InternetChargeType))
	}
	if local.LoadBalancerAttribute.Bandwidth != 0 &&
		local.LoadBalancerAttribute.Bandwidth != remote.LoadBalancerAttribute.Bandwidth &&
		local.LoadBalancerAttribute.InternetChargeType == model.PayByBandwidth {
		needUpdate = true
		bandwidth = local.LoadBalancerAttribute.Bandwidth
		reqCtx.Log.WithName(lbId).Info(fmt.Sprintf("update lb: bandwidth changed([%d] -> [%d])",
			remote.LoadBalancerAttribute.Bandwidth, local.LoadBalancerAttribute.Bandwidth))
	}
	if needUpdate {
		if remote.LoadBalancerAttribute.AddressType == model.InternetAddressType {
			reqCtx.Log.WithName(lbId).Info(fmt.Sprintf("update lb: modify loadbalancer: chargeType=%s, bandwidth=%d", charge, bandwidth))
			return mgr.cloud.ModifyLoadBalancerInternetSpec(reqCtx.Ctx, lbId, string(charge), bandwidth)
		} else {
			reqCtx.Log.WithName(lbId).Info("update lb: only internet loadbalancer is allowed to modify bandwidth and pay type")
		}
	}

	// update instance spec
	if local.LoadBalancerAttribute.LoadBalancerSpec != "" &&
		local.LoadBalancerAttribute.LoadBalancerSpec != remote.LoadBalancerAttribute.LoadBalancerSpec {
		reqCtx.Log.WithName(lbId).Info("update lb: loadbalancerSpec changed ([%s] -> [%s])",
			remote.LoadBalancerAttribute.LoadBalancerSpec, local.LoadBalancerAttribute.LoadBalancerSpec)
		return mgr.cloud.ModifyLoadBalancerInstanceSpec(reqCtx.Ctx, lbId, string(local.LoadBalancerAttribute.LoadBalancerSpec))
	}

	// update slb delete protection
	if local.LoadBalancerAttribute.DeleteProtection != "" &&
		local.LoadBalancerAttribute.DeleteProtection != remote.LoadBalancerAttribute.DeleteProtection {
		reqCtx.Log.WithName(lbId).Info("update lb: delete protection changed([%s] -> [%s])",
			remote.LoadBalancerAttribute.DeleteProtection, local.LoadBalancerAttribute.DeleteProtection)
		return mgr.cloud.SetLoadBalancerDeleteProtection(reqCtx.Ctx, lbId, string(local.LoadBalancerAttribute.DeleteProtection))
	}

	// update modification protection
	if local.LoadBalancerAttribute.ModificationProtectionStatus != "" &&
		local.LoadBalancerAttribute.ModificationProtectionStatus != remote.LoadBalancerAttribute.ModificationProtectionStatus {
		reqCtx.Log.WithName(lbId).Info("update lb: loadbalancer modification protection changed([%s] -> [%s])",
			remote.LoadBalancerAttribute.ModificationProtectionStatus, local.LoadBalancerAttribute.ModificationProtectionStatus, )
		return mgr.cloud.SetLoadBalancerModificationProtection(reqCtx.Ctx, lbId, string(local.LoadBalancerAttribute.ModificationProtectionStatus))
	}

	// update slb name
	// only user defined slb or slb which has "kubernetes.do.not.delete" tag can update name
	if local.LoadBalancerAttribute.LoadBalancerName != "" &&
		local.LoadBalancerAttribute.LoadBalancerName != remote.LoadBalancerAttribute.LoadBalancerName {
		reqCtx.Log.WithName(lbId).Info("update lb: loadbalancer name changed([%s]-[%s])",
			remote.LoadBalancerAttribute.LoadBalancerName, local.LoadBalancerAttribute.LoadBalancerName)
		return mgr.cloud.SetLoadBalancerName(reqCtx.Ctx, lbId, local.LoadBalancerAttribute.LoadBalancerName)
	}
	return nil
}

// Build build load balancer attribute for local model
func (mgr *LoadBalancerManager) BuildLocalModel(reqCtx *RequestContext, mdl *model.LoadBalancer) error {
	mdl.LoadBalancerAttribute.AddressType = model.AddressType(reqCtx.Anno.Get(AddressType))
	mdl.LoadBalancerAttribute.InternetChargeType = model.InternetChargeType(reqCtx.Anno.Get(ChargeType))
	bandwidth := reqCtx.Anno.Get(Bandwidth)
	if bandwidth != "" {
		i, err := strconv.Atoi(bandwidth)
		if err != nil &&
			mdl.LoadBalancerAttribute.InternetChargeType == model.PayByBandwidth {
			return fmt.Errorf("bandwidth must be integer, got [%s], error: %s", bandwidth, err.Error())
		}
		mdl.LoadBalancerAttribute.Bandwidth = i
	}
	if reqCtx.Anno.Get(LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(LoadBalancerId)
		mdl.LoadBalancerAttribute.IsUserManaged = true
	}
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.Anno.Get(LoadBalancerName)
	mdl.LoadBalancerAttribute.VSwitchId = reqCtx.Anno.Get(VswitchId)
	mdl.LoadBalancerAttribute.MasterZoneId = reqCtx.Anno.Get(MasterZoneID)
	mdl.LoadBalancerAttribute.SlaveZoneId = reqCtx.Anno.Get(SlaveZoneID)
	mdl.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(reqCtx.Anno.Get(Spec))
	mdl.LoadBalancerAttribute.ResourceGroupId = reqCtx.Anno.Get(ResourceGroupId)
	mdl.LoadBalancerAttribute.AddressIPVersion = model.AddressIPVersionType(reqCtx.Anno.Get(IPVersion))
	mdl.LoadBalancerAttribute.DeleteProtection = model.FlagType(reqCtx.Anno.Get(DeleteProtection))
	mdl.LoadBalancerAttribute.ModificationProtectionStatus = model.ModificationProtectionType(reqCtx.Anno.Get(ModificationProtection))
	return nil
}

func (mgr *LoadBalancerManager) BuildRemoteModel(reqCtx *RequestContext, mdl *model.LoadBalancer) error {
	return mgr.Find(reqCtx, mdl)
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

func setModelDefaultValue(mgr *LoadBalancerManager, mdl *model.LoadBalancer, anno *AnnotationRequest) error {
	if mdl.LoadBalancerAttribute.AddressType == "" {
		mdl.LoadBalancerAttribute.AddressType = model.AddressType(anno.GetDefaultValue(AddressType))
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

	if mdl.LoadBalancerAttribute.LoadBalancerSpec == "" {
		mdl.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(anno.GetDefaultValue(Spec))
	}

	if mdl.LoadBalancerAttribute.DeleteProtection == "" {
		mdl.LoadBalancerAttribute.DeleteProtection = model.FlagType(anno.GetDefaultValue(DeleteProtection))
	}

	if mdl.LoadBalancerAttribute.ModificationProtectionStatus == "" {
		mdl.LoadBalancerAttribute.ModificationProtectionStatus = model.ModificationProtectionType(anno.GetDefaultValue(ModificationProtection))
		mdl.LoadBalancerAttribute.ModificationProtectionReason = model.ModificationProtectionReason
	}

	mdl.LoadBalancerAttribute.Tags = append(mdl.LoadBalancerAttribute.Tags, anno.GetDefaultTags()...)
	return nil
}
