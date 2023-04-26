package elb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

func NewELBManager(cloud prvd.Provider) *ELBManager {
	return &ELBManager{
		cloud: cloud,
	}
}

type ELBManager struct {
	cloud prvd.Provider
}

func (mgr *ELBManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {

	// private network and virtual switch is must be specified when lb is not specified
	if reqCtx.Anno.Get(annotation.LoadBalancerId) == "" && (reqCtx.Anno.Get(annotation.EdgeNetWorkId) == "" || reqCtx.Anno.Get(annotation.VswitchId) == "") {
		return fmt.Errorf("%s lacks annotation about edge private network or switch", util.Key(reqCtx.Service))
	}
	if err := mgr.setELBFromDefaultConfig(reqCtx, mdl); err != nil {
		return err
	}
	if err := mgr.setELBFromAnnotation(reqCtx, mdl); err != nil {
		return err
	}

	return nil
}

func (mgr *ELBManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if err := mgr.Find(reqCtx, mdl); err != nil {
		return err
	}
	if err := mgr.cloud.FindAssociatedInstance(reqCtx.Ctx, mdl); err != nil {
		return fmt.Errorf("check lb %s has associated with eip, error %s", mdl.GetLoadBalancerId(), err.Error())
	}
	return nil
}

func (mgr *ELBManager) Find(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(annotation.LoadBalancerId)
	}
	if reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse) != "" {
		mdl.LoadBalancerAttribute.IsReUsed = isTrue(reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse))
	}
	mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.Anno.GetDefaultLoadBalancerName()
	err := mgr.cloud.FindEdgeLoadBalancer(reqCtx.Ctx, mdl)
	if err != nil {
		// load balancer id is set empty if it can not find load balancer instance by id or name
		mdl.LoadBalancerAttribute.LoadBalancerId = ""
		return err
	}
	return nil
}

func (mgr *ELBManager) Create(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if err := mgr.checkBeforeCreate(reqCtx, mdl); err != nil {
		return err
	}
	if err := mgr.cloud.CreateEdgeLoadBalancer(reqCtx.Ctx, mdl); err != nil {
		return err
	}
	return nil
}

func (mgr *ELBManager) checkBeforeCreate(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if mdl.GetNetworkId() == "" {
		return fmt.Errorf("check error, lock network id")
	}
	if mdl.GetVSwitchId() == "" {
		return fmt.Errorf("check error, lock vswitch id")
	}
	if mdl.LoadBalancerAttribute.EnsRegionId == "" {
		err := mgr.cloud.DescribeNetwork(reqCtx.Ctx, mdl)
		if err != nil {
			return fmt.Errorf("check error, lock region id")
		}
	}
	if mdl.LoadBalancerAttribute.LoadBalancerSpec == "" {
		mdl.LoadBalancerAttribute.LoadBalancerSpec = elbmodel.ELBDefaultSpec
	}

	if mdl.LoadBalancerAttribute.PayType == "" {
		mdl.LoadBalancerAttribute.PayType = elbmodel.ELBDefaultPayType
	}

	if mdl.LoadBalancerAttribute.LoadBalancerName == "" {
		mdl.LoadBalancerAttribute.LoadBalancerName = reqCtx.Anno.GetDefaultLoadBalancerName()
	}
	return nil
}

func (mgr *ELBManager) setELBFromDefaultConfig(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	mdl.LoadBalancerAttribute.PayType = elbmodel.ELBDefaultPayType
	mdl.LoadBalancerAttribute.LoadBalancerSpec = elbmodel.ELBDefaultSpec
	return nil
}
func (mgr *ELBManager) setELBFromAnnotation(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if reqCtx.Anno.Get(annotation.LoadBalancerId) != "" {
		mdl.LoadBalancerAttribute.LoadBalancerId = reqCtx.Anno.Get(annotation.LoadBalancerId)
		ret, err := mgr.cloud.FindNetWorkAndVSwitchByLoadBalancerId(reqCtx.Ctx, mdl.LoadBalancerAttribute.LoadBalancerId)
		if err != nil {
			return fmt.Errorf("%s has incorrect annotation about %s to find network and vswitch, err: %s", util.Key(reqCtx.Service), annotation.LoadBalancerId, err.Error())
		}
		mdl.LoadBalancerAttribute.NetworkId = ret[0]
		mdl.LoadBalancerAttribute.VSwitchId = ret[1]
	} else {
		if reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse) != "" && isTrue(reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse)) {
			return fmt.Errorf("service %s reuse elb, must specify the elb that have been created by user", util.Key(reqCtx.Service))
		}
		mdl.LoadBalancerAttribute.NetworkId = reqCtx.Anno.Get(annotation.EdgeNetWorkId)
		mdl.LoadBalancerAttribute.VSwitchId = reqCtx.Anno.Get(annotation.VswitchId)
		mdl.LoadBalancerAttribute.IsUserManaged = false
	}
	err := mgr.cloud.DescribeNetwork(reqCtx.Ctx, mdl)
	if err != nil {
		return fmt.Errorf("%s has incorrect annotation about network id %s and vswitch id %s to get region by network, err: %s",
			util.Key(reqCtx.Service), mdl.GetNetworkId(), mdl.GetVSwitchId(), err.Error())
	}
	if reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse) != "" {
		mdl.LoadBalancerAttribute.IsReUsed = isTrue(reqCtx.Anno.Get(annotation.EdgeLoadBalancerReUse))
	}
	mdl.LoadBalancerAttribute.LoadBalancerSpec = reqCtx.Anno.Get(annotation.Spec)
	mdl.LoadBalancerAttribute.PayType = reqCtx.Anno.Get(annotation.EdgePayType)
	return nil
}

func (mgr *ELBManager) Update(reqCtx *svcCtx.RequestContext, lMdl, rMdl *elbmodel.EdgeLoadBalancer) error {
	if rMdl.GetLoadBalancerId() == "" {
		return fmt.Errorf("found no load balancer %s, try to update load balancer", rMdl.GetLoadBalancerId())
	}
	reqCtx.Log.Info(fmt.Sprintf("load balancer [%s] start to update load balancer attribute", rMdl.GetLoadBalancerId()))
	var isChanged = false
	if lMdl.LoadBalancerAttribute.LoadBalancerSpec != "" && !isEqual(lMdl.LoadBalancerAttribute.LoadBalancerSpec, rMdl.LoadBalancerAttribute.LoadBalancerSpec) {
		isChanged = true
	}
	if lMdl.LoadBalancerAttribute.PayType != "" && !isEqual(lMdl.LoadBalancerAttribute.PayType, rMdl.LoadBalancerAttribute.PayType) {
		isChanged = true
	}
	if !isChanged {
		return nil
	}
	reqCtx.Log.Info(fmt.Sprintf("edge load balancer [%s] not support upgrade", rMdl.GetLoadBalancerId()))
	return nil
}

func (mgr *ELBManager) waitFindLoadBalancer(reqCtx *svcCtx.RequestContext, lbId string, mdl *elbmodel.EdgeLoadBalancer) error {
	waitErr := util.RetryImmediateOnError(10*time.Second, time.Minute, canSkipError, func() error {
		err := mgr.cloud.DescribeEdgeLoadBalancerById(reqCtx.Ctx, lbId, mdl)
		if err != nil {
			return err
		}
		if mdl.GetLoadBalancerId() == "" {
			return fmt.Errorf("service %s %s load balancer", InstanceNotFound, mdl.NamespacedName.String())
		}
		if mdl.LoadBalancerAttribute.LoadBalancerStatus == ELBActive || mdl.LoadBalancerAttribute.LoadBalancerStatus == ELBInActive {
			return nil
		}
		return fmt.Errorf("elb %s status is %s", mdl.GetLoadBalancerId(), StatusAberrant)
	})
	if waitErr != nil {
		return fmt.Errorf("failed to find elb %s", mdl.GetLoadBalancerId())
	}
	return nil
}

func (mgr *ELBManager) DeleteELB(reqCtx *svcCtx.RequestContext, mdl *elbmodel.EdgeLoadBalancer) error {
	if mdl.LoadBalancerAttribute.LoadBalancerStatus == ELBActive {
		err := mgr.cloud.SetEdgeLoadBalancerStatus(reqCtx.Ctx, ELBInActive, mdl)
		if err != nil {
			return fmt.Errorf("delete elb [%s] error: %s", mdl.GetLoadBalancerId(), err.Error())
		}
	}
	waitErr := util.RetryImmediateOnError(10*time.Second, 2*time.Minute, canSkipError, func() error {
		err := mgr.cloud.DescribeEdgeLoadBalancerById(reqCtx.Ctx, mdl.GetLoadBalancerId(), mdl)
		if err != nil {
			return err
		}
		if mdl.LoadBalancerAttribute.LoadBalancerStatus == ELBInActive {
			return nil
		}
		return fmt.Errorf("elb %s status is aberrant", mdl.GetLoadBalancerId())
	})
	if waitErr != nil {
		return fmt.Errorf("failed to inactive elb %s", mdl.GetLoadBalancerId())
	}
	err := mgr.cloud.DeleteEdgeLoadBalancer(reqCtx.Ctx, mdl)
	if err != nil {
		return fmt.Errorf("delete elb [%s] error: %s", mdl.GetLoadBalancerId(), err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("successfully delete elb %s", mdl.GetLoadBalancerId()))
	return nil
}

func canSkipError(err error) bool {
	return strings.Contains(err.Error(), InstanceNotFound) || strings.Contains(err.Error(), StatusAberrant)
}

func isTrue(s string) bool {
	if s == "false" || s == "False" || s == "FALSE" {
		return false
	}
	return true
}

func isEqual(s1, s2 interface{}) bool {
	if reflect.TypeOf(s1) != reflect.TypeOf(s2) {
		return false
	}
	return reflect.DeepEqual(s1, s2)
}
