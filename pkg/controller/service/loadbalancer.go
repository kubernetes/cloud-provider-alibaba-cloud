package service

import (
	"encoding/json"
	"fmt"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/klog"
	"strconv"
)

func (reqCtx *RequestContext) FindLoadBalancer(mdl *model.LoadBalancer) error {
	if reqCtx == nil {
		klog.Infof("reqCtx is nil")
	}
	if reqCtx.anno == nil {
		klog.Infof("anno is nil")
	}
	// 1. set load balancer id
	if reqCtx.anno.Get(LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.anno.Get(LoadBalancerId)
	}

	// 2. set default loadbalancer name
	// it's safe to set loadbalancer name which will be overwritten in FindLoadBalancer func
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.anno.GetDefaultLoadBalancerName()

	// 3. set default loadbalancer tag
	mdl.LoadBalancerAttribute.Tags = reqCtx.anno.GetDefaultTags()
	return reqCtx.cloud.FindLoadBalancer(reqCtx.ctx, mdl)
}

func (reqCtx *RequestContext) BuildLoadBalancerAttributeForLocalModel(mdl *model.LoadBalancer) error {
	mdl.LoadBalancerAttribute.AddressType = model.AddressType(reqCtx.anno.Get(AddressType))
	mdl.LoadBalancerAttribute.InternetChargeType = model.InternetChargeType(reqCtx.anno.Get(ChargeType))
	bandwidth := reqCtx.anno.Get(Bandwidth)
	if bandwidth != "" {
		i, err := strconv.Atoi(bandwidth)
		if err != nil &&
			mdl.LoadBalancerAttribute.InternetChargeType == model.PayByBandwidth {
			return fmt.Errorf("bandwidth must be integer, got [%s], error: %s", bandwidth, err.Error())
		}
		mdl.LoadBalancerAttribute.Bandwidth = i
	}
	if reqCtx.anno.Get(LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.anno.Get(LoadBalancerId)
		mdl.LoadBalancerAttribute.IsUserManaged = true
	}
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.anno.Get(LoadBalancerName)
	mdl.LoadBalancerAttribute.VSwitchId = reqCtx.anno.Get(VswitchId)
	mdl.LoadBalancerAttribute.MasterZoneId = reqCtx.anno.Get(MasterZoneID)
	mdl.LoadBalancerAttribute.SlaveZoneId = reqCtx.anno.Get(SlaveZoneID)
	mdl.LoadBalancerAttribute.LoadBalancerSpec = model.LoadBalancerSpecType(reqCtx.anno.Get(Spec))
	mdl.LoadBalancerAttribute.ResourceGroupId = reqCtx.anno.Get(ResourceGroupId)
	mdl.LoadBalancerAttribute.AddressIPVersion = model.AddressIPVersionType(reqCtx.anno.Get(IPVersion))
	mdl.LoadBalancerAttribute.DeleteProtection = model.FlagType(reqCtx.anno.Get(DeleteProtection))
	mdl.LoadBalancerAttribute.ModificationProtectionStatus = model.ModificationProtectionType(reqCtx.anno.Get(ModificationProtection))
	return nil
}

func (reqCtx *RequestContext) BuildLoadBalancerAttributeForRemoteModel(mdl *model.LoadBalancer) error {
	return reqCtx.FindLoadBalancer(mdl)
}

func (reqCtx *RequestContext) EnsureLoadBalancerCreated(local *model.LoadBalancer) error {
	setModelDefaultValue(local, reqCtx.anno)
	err := reqCtx.cloud.CreateLoadBalancer(reqCtx.ctx, local)
	if err != nil {
		return fmt.Errorf("create slb error: %s", err.Error())
	}

	tags, err := json.Marshal(local.LoadBalancerAttribute.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags error: %s", err.Error())
	}
	return reqCtx.cloud.AddTags(reqCtx.ctx, local.LoadBalancerAttribute.LoadBalancerId, string(tags))
}

func (reqCtx *RequestContext) EnsureLoadBalancerDeleted(mdl *model.LoadBalancer) error {
	if mdl.LoadBalancerAttribute.LoadBalancerId == "" {
		return nil
	}

	// set delete protection off
	if mdl.LoadBalancerAttribute.DeleteProtection == model.OnFlag {
		if err := reqCtx.cloud.SetLoadBalancerDeleteProtection(
			reqCtx.ctx,
			mdl.LoadBalancerAttribute.LoadBalancerId,
			string(model.OffFlag),
		); err != nil {
			return fmt.Errorf("error to set slb id [%s] delete protection off, svc [%s], err: %s",
				mdl.LoadBalancerAttribute.LoadBalancerId, mdl.NamespacedName, err.Error())
		}
	}

	return reqCtx.cloud.DeleteLoadBalancer(reqCtx.ctx, mdl)

}

func (reqCtx *RequestContext) EnsureLoadBalancerUpdated(local, remote *model.LoadBalancer) error {
	lbId := remote.LoadBalancerAttribute.LoadBalancerId
	klog.Infof("found load balancer [%s], try to update load balancer attribute", lbId)

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
		klog.Infof("internet chargeType changed([%s] -> [%s]), update loadbalancer [%s]",
			remote.LoadBalancerAttribute.InternetChargeType, local.LoadBalancerAttribute.InternetChargeType, lbId)
	}
	if local.LoadBalancerAttribute.Bandwidth != 0 &&
		local.LoadBalancerAttribute.Bandwidth != remote.LoadBalancerAttribute.Bandwidth &&
		local.LoadBalancerAttribute.InternetChargeType == model.PayByBandwidth {
		needUpdate = true
		bandwidth = local.LoadBalancerAttribute.Bandwidth
		klog.Infof("bandwidth changed([%d] -> [%d]), update loadbalancer[%s]",
			remote.LoadBalancerAttribute.Bandwidth, local.LoadBalancerAttribute.Bandwidth, lbId)
	}
	if needUpdate {
		if remote.LoadBalancerAttribute.AddressType == model.InternetAddressType {
			klog.Infof("modify loadbalancer: chargeType=%s, bandwidth=%d", charge, bandwidth)
			return reqCtx.cloud.ModifyLoadBalancerInternetSpec(reqCtx.ctx, lbId, string(charge), bandwidth)
		} else {
			klog.Warningf("only internet loadbalancer is allowed to modify bandwidth and pay type")
		}
	}

	// update instance spec
	if local.LoadBalancerAttribute.LoadBalancerSpec != "" &&
		local.LoadBalancerAttribute.LoadBalancerSpec != remote.LoadBalancerAttribute.LoadBalancerSpec {
		klog.Infof("alicloud: loadbalancerSpec changed ([%s] -> [%s]), update loadbalancer [%s]",
			remote.LoadBalancerAttribute.LoadBalancerSpec, local.LoadBalancerAttribute.LoadBalancerSpec, lbId)
		return reqCtx.cloud.ModifyLoadBalancerInstanceSpec(reqCtx.ctx, lbId, string(local.LoadBalancerAttribute.LoadBalancerSpec))
	}

	// update slb delete protection
	if local.LoadBalancerAttribute.DeleteProtection != "" &&
		local.LoadBalancerAttribute.DeleteProtection != remote.LoadBalancerAttribute.DeleteProtection {
		klog.Infof("delete protection changed([%s] -> [%s]), update loadbalancer [%s]",
			remote.LoadBalancerAttribute.DeleteProtection, local.LoadBalancerAttribute.DeleteProtection, lbId)
		return reqCtx.cloud.SetLoadBalancerDeleteProtection(reqCtx.ctx, lbId, string(local.LoadBalancerAttribute.DeleteProtection))
	}

	// update modification protection
	if local.LoadBalancerAttribute.ModificationProtectionStatus != "" &&
		local.LoadBalancerAttribute.ModificationProtectionStatus != remote.LoadBalancerAttribute.ModificationProtectionStatus {
		klog.Infof("alicloud: loadbalancer modification protection changed([%s] -> [%s]) changed, update loadbalancer [%s]",
			remote.LoadBalancerAttribute.ModificationProtectionStatus, local.LoadBalancerAttribute.ModificationProtectionStatus,
			remote.LoadBalancerAttribute.LoadBalancerId)
		return reqCtx.cloud.SetLoadBalancerModificationProtection(reqCtx.ctx, lbId, string(local.LoadBalancerAttribute.ModificationProtectionStatus))
	}

	// update slb name
	// only user defined slb or slb which has "kubernetes.do.not.delete" tag can update name
	if local.LoadBalancerAttribute.LoadBalancerName != "" &&
		local.LoadBalancerAttribute.LoadBalancerName != remote.LoadBalancerAttribute.LoadBalancerName {
		//if isLoadBalancerHasTag(tags) || isUserDefinedLoadBalancer(service) {
		//	klog.Infof("alicloud: LoadBalancer name (%s -> %s) changed, update loadbalancer [%s]",
		//		remote.LoadBalancerAttribute.LoadBalancerName, local.LoadBalancerAttribute.LoadBalancerName, lbId)
		//	if err := slbClient.SetLoadBalancerName(context, lbId, local.LoadBalancerAttribute.LoadBalancerName); err != nil {
		//		return err
		//	}
		//}
		return reqCtx.cloud.SetLoadBalancerName(reqCtx.ctx, lbId, local.LoadBalancerAttribute.LoadBalancerName)
	}
	return nil
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

func setModelDefaultValue(mdl *model.LoadBalancer, anno *AnnotationRequest) {
	if mdl.LoadBalancerAttribute.AddressType == "" {
		mdl.LoadBalancerAttribute.AddressType = model.AddressType(anno.GetDefaultValue(AddressType))
	}

	if mdl.LoadBalancerAttribute.LoadBalancerName == "" {
		mdl.LoadBalancerAttribute.LoadBalancerName = anno.GetDefaultLoadBalancerName()
	}

	// TODO ecs模式下获取vpc id & vsw id
	if mdl.LoadBalancerAttribute.AddressType == model.IntranetAddressType {
		mdl.LoadBalancerAttribute.VpcId = ctx2.CFG.Global.VpcID
		if mdl.LoadBalancerAttribute.VSwitchId == "" {
			mdl.LoadBalancerAttribute.VSwitchId = ctx2.CFG.Global.VswitchID
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
}
