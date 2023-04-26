package elb

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"strconv"
	"strings"
	"time"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/klog/v2"
)

func NewEIPManager(cloud prvd.Provider) *EIPManager {
	return &EIPManager{
		cloud: cloud,
	}
}

type EIPManager struct {
	cloud prvd.Provider
}

func (mgr *EIPManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if reqCtx.Anno.Get(annotation.EdgeEipAssociate) != "" && !isTrue(reqCtx.Anno.Get(annotation.EdgeEipAssociate)) {
		return nil
	}
	err := setEipFromDefaultConfig(reqCtx, mdl)
	if err != nil {
		return fmt.Errorf("set default eip attribute error, %s", err.Error())
	}

	err = setEipFromAnnotation(reqCtx, &mdl.EipAttribute)
	if err != nil {
		return fmt.Errorf("set eip attribute error, %s", err.Error())
	}
	return nil
}

func (mgr *EIPManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	return mgr.Find(reqCtx, mdl)
}

func (mgr *EIPManager) Find(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	// set loadbalancer id
	if reqCtx.Anno.Get(annotation.EdgeEipId) != "" {
		mdl.EipAttribute.AllocationId = reqCtx.Anno.Get(annotation.EdgeEipId)
	}
	mdl.EipAttribute.Name = getDefaultEIPName(reqCtx)
	if mdl.GetEipId() != "" {
		err := mgr.cloud.DescribeEnsEipAddressesById(reqCtx.Ctx, mdl.GetEipId(), mdl)
		if err != nil {
			klog.Infof("[%s] find no edge eip by id", mdl.NamespacedName)
			return err
		}
		klog.Infof("[%s] find edge eip by id, edge eip ip [%s]", mdl.NamespacedName, mdl.GetEipId())
		return nil
	}
	err := mgr.cloud.DescribeEnsEipAddressesByName(reqCtx.Ctx, mdl.GetEipName(), mdl)
	if err != nil {
		klog.Infof("[%s] find no edge eip by name", mdl.NamespacedName)
		return err
	}
	klog.Infof("[%s] find eip by name, edge eip name [%s]", mdl.NamespacedName, mdl.GetEipName())
	return nil
}

func (mgr *EIPManager) Update(reqCtx *svcCtx.RequestContext, lMdl, rMdl *elbmodel.EdgeLoadBalancer) error {
	var isChanged = false
	if lMdl.EipAttribute.Bandwidth > 0 && lMdl.EipAttribute.Bandwidth != rMdl.EipAttribute.Bandwidth {
		isChanged = true
	}
	if lMdl.EipAttribute.Description != rMdl.EipAttribute.Description {
		isChanged = true
	}
	if isChanged {
		return mgr.cloud.ModifyEipAttribute(reqCtx.Ctx, rMdl.GetEipId(), lMdl)
	}
	return nil
}

func (mgr *EIPManager) waitFindEip(reqCtx *svcCtx.RequestContext, eipId string, mdl *elbmodel.EdgeLoadBalancer) error {
	waitErr := util.RetryImmediateOnError(10*time.Second, 2*time.Minute, canSkipError, func() error {
		err := mgr.cloud.DescribeEnsEipAddressesById(reqCtx.Ctx, eipId, mdl)
		if err != nil {
			return err
		}
		if mdl.GetEipId() == "" {
			return fmt.Errorf("service %s find no eip", mdl.NamespacedName.String())
		}
		if mdl.EipAttribute.Status == EipAvailable || mdl.EipAttribute.Status == EipInUse {
			return nil
		}
		return fmt.Errorf("eip %s status is aberrant", mdl.GetEipId())
	})

	if waitErr != nil {
		return fmt.Errorf("failed to find eip %s", mdl.GetEipId())
	}
	return nil
}

func (mgr *EIPManager) DeleteEip(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	waitErr := util.RetryImmediateOnError(10*time.Second, 2*time.Minute, canSkipError, func() error {
		err := mgr.cloud.DescribeEnsEipAddressesById(reqCtx.Ctx, mdl.GetEipId(), mdl)
		if err != nil {
			return err
		}
		if mdl.GetEipId() == "" {
			return fmt.Errorf("service %s find no eip", mdl.NamespacedName.String())
		}
		if mdl.EipAttribute.Status == EipAvailable {
			return nil
		}
		if mdl.EipAttribute.Status == EipInUse {
			err = mgr.cloud.UnAssociateElbEipAddress(reqCtx.Ctx, mdl.GetEipId())
			if err != nil {
				return fmt.Errorf("service %s delete eip error %s", mdl.NamespacedName.String(), err.Error())
			}
		}
		return fmt.Errorf("eip %s status is aberrant", mdl.GetEipId())
	})
	if waitErr != nil {
		return fmt.Errorf("failed to unassociate and release eip %s due to eip status", mdl.GetEipId())
	}
	if err := mgr.cloud.ReleaseEip(reqCtx.Ctx, mdl.GetEipId()); err != nil {
		return fmt.Errorf("service %s delete eip error %s", mdl.NamespacedName.String(), err.Error())
	}
	return nil
}

func setEipFromDefaultConfig(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if mdl.LoadBalancerAttribute.EnsRegionId == "" {
		return fmt.Errorf("remote model %s lacks edge load balancer attributes region", mdl.NamespacedName.String())
	}
	mdl.EipAttribute.Name = getDefaultEIPName(reqCtx)
	mdl.EipAttribute.NamedKey = elbmodel.NamedKey{
		Prefix:      model.DEFAULT_PREFIX,
		CID:         base.CLUSTER_ID,
		Namespace:   reqCtx.Service.Namespace,
		ServiceName: reqCtx.Service.Name,
	}
	mdl.EipAttribute.Description = mdl.EipAttribute.NamedKey.String()
	mdl.EipAttribute.EnsRegionId = mdl.LoadBalancerAttribute.EnsRegionId
	mdl.EipAttribute.Bandwidth = elbmodel.EipDefaultBandwidth
	mdl.EipAttribute.InstanceChargeType = elbmodel.EipDefaultInstanceChargeType
	mdl.EipAttribute.InternetChargeType = elbmodel.EipDefaultInternetChargeType
	return nil
}

func setEipFromAnnotation(reqCtx *svcCtx.RequestContext, eip *elbmodel.EdgeEipAttribute) error {
	if reqCtx.Anno.Get(annotation.EdgeEipId) != "" {
		eip.AllocationId = reqCtx.Anno.Get(annotation.EdgeEipId)
	} else {
		eip.IsUserManaged = false
		if reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse) != "" && isTrue(reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse)) {
			if reqCtx.Anno.Get(annotation.EdgeEipAssociate) == "" {
				return fmt.Errorf("service %s reuse elb not support ccm manage eip", util.Key(reqCtx.Service))
			}
			if isTrue(reqCtx.Anno.Get(annotation.EdgeEipAssociate)) {
				return fmt.Errorf("service %s reuse elb not support ccm manage eip", util.Key(reqCtx.Service))
			}
		}
	}
	if reqCtx.Anno.Get(annotation.EdgeEipBandwidth) != "" {
		bandwidth, err := strconv.Atoi(reqCtx.Anno.Get(annotation.EdgeEipBandwidth))
		if err != nil {
			return fmt.Errorf("Annotation eip bandwidth must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.EdgeEipBandwidth), err.Error())
		}
		eip.Bandwidth = bandwidth
	}
	if reqCtx.Anno.Get(annotation.EdgeEipInstanceChargeType) != "" {
		eip.InstanceChargeType = reqCtx.Anno.Get(annotation.EdgeEipInstanceChargeType)
	}
	if reqCtx.Anno.Get(annotation.EdgeEipInternetChargeType) != "" {
		eip.InternetChargeType = reqCtx.Anno.Get(annotation.EdgeEipInternetChargeType)
	}
	return nil
}

func getDefaultEIPName(reqCtx *svcCtx.RequestContext) string {
	//GCE requires that the name of a load balancer starts with a lower case letter.
	ret := EipPrefix + string(reqCtx.Service.UID)
	ret = strings.Replace(ret, "-", "", -1)
	//AWS requires that the name of a load balancer is shorter than 32 bytes.
	if len(ret) > 32 {
		ret = ret[:32]
	}
	return ret
}
