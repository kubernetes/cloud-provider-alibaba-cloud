package elb

import (
	"fmt"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"strings"

	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

func NewModelApplier(elbMgr *ELBManager, eipMgr *EIPManager, lisMgr *ListenerManager, sgMgr *ServerGroupManager) *ModelApplier {
	return &ModelApplier{
		ELBMgr: elbMgr,
		EIPMgr: eipMgr,
		LisMgr: lisMgr,
		SGMgr:  sgMgr,
	}
}

type ModelApplier struct {
	ELBMgr *ELBManager
	EIPMgr *EIPManager
	LisMgr *ListenerManager
	SGMgr  *ServerGroupManager
}

func (applier *ModelApplier) Apply(reqCtx *svcCtx.RequestContext, local *elbmodel.EdgeLoadBalancer) (*elbmodel.EdgeLoadBalancer, error) {
	remote := &elbmodel.EdgeLoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: elbmodel.EdgeLoadBalancerAttribute{
			IsUserManaged: true,
			IsReUsed:      false,
		},
		EipAttribute: elbmodel.EdgeEipAttribute{IsUserManaged: true},
		ServerGroup:  elbmodel.EdgeServerGroup{},
		Listeners:    elbmodel.EdgeListeners{},
	}
	err := applier.ELBMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return remote, fmt.Errorf("get load balancer attribute from cloud, error %s", err.Error())
	}
	if helper.IsServiceHashChanged(reqCtx.Service) || ctrlCfg.ControllerCFG.DryRun {
		if err = applier.applyLoadBalancerAttribute(reqCtx, local, remote); err != nil {
			return remote, fmt.Errorf("reconcile elb attribute error: %s", err.Error())
		}
	}
	//create and update event must need load balancer id
	if remote.GetLoadBalancerId() == "" && !needDeleteLoadBalancer(reqCtx.Service) {
		return remote, fmt.Errorf("alibaba cloud: can not find loadbalancer %s for service %s", remote.GetLoadBalancerName(), remote.NamespacedName.String())
	}

	err = applier.EIPMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return remote, fmt.Errorf("get EIP attribute from cloud, error %s", err.Error())
	}
	if helper.IsServiceHashChanged(reqCtx.Service) || ctrlCfg.ControllerCFG.DryRun {
		if err = applier.applyEdgeElasticIpAttribute(reqCtx, local, remote); err != nil {
			return remote, fmt.Errorf("reconcile eip attribute, error %s", err.Error())
		}
	}
	// delete event has released load balancer and eip, return
	if needDeleteLoadBalancer(reqCtx.Service) {
		return remote, nil
	}

	err = applier.SGMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return remote, fmt.Errorf("get server group from cloud, error %s", err.Error())
	}
	if err = applier.applyServerGroupAttribute(reqCtx, local, remote); err != nil {
		return remote, fmt.Errorf("reconcile server group error %s", err.Error())
	}

	err = applier.LisMgr.BuildRemoteModel(reqCtx, remote)
	if err != nil {
		return remote, fmt.Errorf("get listeners from cloud, error %s", err.Error())
	}
	if err = applier.applyListenersAttribute(reqCtx, local, remote); err != nil {
		return remote, fmt.Errorf("reconcile listener error %s", err.Error())
	}
	return remote, nil
}

func (applier *ModelApplier) applyLoadBalancerAttribute(reqCtx *svcCtx.RequestContext, localModel, remoteModel *elbmodel.EdgeLoadBalancer) error {
	if localModel == nil || remoteModel == nil {
		return fmt.Errorf("local model or remote model is nil")
	}

	if localModel.NamespacedName.String() != remoteModel.NamespacedName.String() {
		return fmt.Errorf("models for different svc, local [%s], remote [%s]",
			localModel.NamespacedName.String(), remoteModel.NamespacedName.String())
	}

	//delete elb when service is delete
	if needDeleteLoadBalancer(reqCtx.Service) {
		if remoteModel.GetLoadBalancerId() == "" {
			return fmt.Errorf("elb does not exist")
		}
		if !remoteModel.LoadBalancerAttribute.IsUserManaged {
			return applier.ELBMgr.DeleteELB(reqCtx, remoteModel)
		}
		reqCtx.Log.Info(fmt.Sprintf("elb %s is manageed by user, skip delete it", remoteModel.GetLoadBalancerId()))
		return nil
	}

	//create elb when loadbalancer id is not specified
	if localModel.GetLoadBalancerId() == "" && remoteModel.GetLoadBalancerId() == "" {
		if helper.IsServiceOwnIngress(reqCtx.Service) {
			return fmt.Errorf("alicloud: can not find loadbalancer, but it's defined in service [%v] "+
				"this may happen when you delete the loadbalancer", reqCtx.Service.Status.LoadBalancer.Ingress[0].IP)
		}
		if err := applier.ELBMgr.Create(reqCtx, localModel); err != nil {
			return fmt.Errorf("create elb error: %s", err.Error())
		}
		if err := applier.ELBMgr.waitFindLoadBalancer(reqCtx, localModel.GetLoadBalancerId(), remoteModel); err != nil {
			return fmt.Errorf("create elb %s error: %s", localModel.GetLoadBalancerId(), err.Error())
		}
		// active load balancer
		if remoteModel.LoadBalancerAttribute.LoadBalancerStatus == ELBInActive {
			err := applier.ELBMgr.cloud.SetEdgeLoadBalancerStatus(reqCtx.Ctx, ELBActive, remoteModel)
			if err != nil {
				return fmt.Errorf("active elb error: %s", err.Error())
			}
		}
		// update remote model
		reqCtx.Log.Info(fmt.Sprintf("successfully create elb %s", remoteModel.GetLoadBalancerId()))
		return nil
	}

	if canReusable, reason := isELBReusable(reqCtx, *remoteModel); !canReusable {
		return fmt.Errorf("the loadbalancer %s can not be reused, %s",
			remoteModel.GetLoadBalancerId(), reason)
	}

	err := applier.ELBMgr.Update(reqCtx, localModel, remoteModel)
	if err != nil {
		return fmt.Errorf("update service %s loadbalancer, error %s", localModel.NamespacedName.String(), err.Error())
	}

	return nil
}

func isELBReusable(reqCtx *svcCtx.RequestContext, remoteModel elbmodel.EdgeLoadBalancer) (bool, string) {
	if len(reqCtx.Service.Status.LoadBalancer.Ingress) < 1 {
		return true, ""
	}
	for _, ingress := range reqCtx.Service.Status.LoadBalancer.Ingress {
		if ingress.Hostname == remoteModel.GetLoadBalancerName() {
			return true, ""
		}
		if ingress.IP == remoteModel.LoadBalancerAttribute.Address {
			return true, ""
		}
		if ingress.IP == remoteModel.LoadBalancerAttribute.AssociatedEipAddress {
			return true, ""
		}
	}
	return false, fmt.Sprintf("service %s can not replace load balancer", remoteModel.NamespacedName.String())
}

func (applier *ModelApplier) applyEdgeElasticIpAttribute(reqCtx *svcCtx.RequestContext, localModel, remoteModel *elbmodel.EdgeLoadBalancer) error {
	if localModel == nil || remoteModel == nil {
		return fmt.Errorf("local model or remote model is nil")
	}

	if localModel.NamespacedName.String() != remoteModel.NamespacedName.String() {
		return fmt.Errorf("models for different svc, local [%s], remote [%s]",
			localModel.NamespacedName.String(), remoteModel.NamespacedName.String())
	}
	if reqCtx.Anno.Get(annotation.EdgeEipAssociate) != "" && !isTrue(reqCtx.Anno.Get(annotation.EdgeEipAssociate)) {
		if remoteModel.LoadBalancerAttribute.AssociatedEipId == "" {
			return nil
		}
		if strings.HasPrefix(remoteModel.LoadBalancerAttribute.AssociatedEipName, EipPrefix) {
			if err := applier.EIPMgr.DeleteEip(reqCtx, remoteModel); err != nil {
				return err
			}
		}
		return nil
	}

	if needDeleteLoadBalancer(reqCtx.Service) {
		if remoteModel.GetEipId() == "" {
			return fmt.Errorf("eip does not exist")
		}
		if !remoteModel.EipAttribute.IsUserManaged {
			if err := applier.EIPMgr.DeleteEip(reqCtx, remoteModel); err != nil {
				return err
			}
			reqCtx.Log.Info(fmt.Sprintf("successfully delete edge eip %s", remoteModel.GetEipId()))
			return nil
		}
		reqCtx.Log.Info(fmt.Sprintf("eip %s is managed by user, skip delete it", remoteModel.GetEipId()))
		return nil
	}

	//create eip when eip id is not specified
	if localModel.GetEipId() == "" && remoteModel.GetEipId() == "" {
		if remoteModel.GetAssociatedEipId() != "" {
			reqCtx.Log.Info(fmt.Sprintf("eip %s is associated to loadbalancer %s, can not overwrite it", remoteModel.GetAssociatedEipId(), remoteModel.GetLoadBalancerId()))
			return nil
		}
		if err := applier.EIPMgr.cloud.CreateEip(reqCtx.Ctx, localModel); err != nil {
			return fmt.Errorf("create eip error: %s", err.Error())
		}

		if err := applier.EIPMgr.waitFindEip(reqCtx, localModel.GetEipId(), remoteModel); err != nil {
			return fmt.Errorf("create eip error %s, error: %s", localModel.GetEipId(), err.Error())
		}

		if err := applier.EIPMgr.cloud.ModifyEipAttribute(reqCtx.Ctx, remoteModel.GetEipId(), localModel); err != nil {
			return fmt.Errorf("modify eip %s description %s error: %s",
				remoteModel.GetEipId(), remoteModel.EipAttribute.Description, err.Error())
		}

		if err := applier.EIPMgr.cloud.AssociateElbEipAddress(reqCtx.Ctx, remoteModel.GetEipId(), remoteModel.GetLoadBalancerId()); err != nil {
			return fmt.Errorf("associate eip %s to elb %s error: %s",
				remoteModel.GetEipId(), remoteModel.GetLoadBalancerId(), err.Error())
		}

		reqCtx.Log.Info(fmt.Sprintf("successfully create eip %s", remoteModel.GetEipId()))
		return nil
	}

	// update eip attribute
	if remoteModel.GetAssociatedEipId() == "" {
		if !remoteModel.EipAttribute.IsUserManaged {
			if err := applier.EIPMgr.Update(reqCtx, localModel, remoteModel); err != nil {
				return fmt.Errorf("update service %s eip %s, error %s", localModel.NamespacedName.String(), remoteModel.GetEipId(), err.Error())
			}
		}
		if err := applier.EIPMgr.cloud.AssociateElbEipAddress(reqCtx.Ctx, remoteModel.GetEipId(), remoteModel.GetLoadBalancerId()); err != nil {
			return fmt.Errorf("update service %s, associate eip %s to elb %s, error %s", remoteModel.NamespacedName.String(),
				remoteModel.GetEipId(), remoteModel.GetLoadBalancerId(), err.Error())
		}
		return nil
	}

	if remoteModel.GetAssociatedEipId() != "" && remoteModel.GetAssociatedEipId() == remoteModel.GetEipId() {
		if !remoteModel.EipAttribute.IsUserManaged {
			err := applier.EIPMgr.Update(reqCtx, localModel, remoteModel)
			if err != nil {
				return fmt.Errorf("update service %s eip %s, error %s", localModel.NamespacedName.String(), remoteModel.GetEipId(), err.Error())
			}
			return nil
		}
		reqCtx.Log.Info(fmt.Sprintf("can not support update service %s eip %s that is manageed by user",
			localModel.NamespacedName.String(), remoteModel.GetEipId()))
		return nil
	}

	if remoteModel.GetAssociatedEipId() != "" && remoteModel.GetAssociatedEipId() != remoteModel.GetEipId() {
		if !remoteModel.EipAttribute.IsUserManaged {
			if err := applier.EIPMgr.DeleteEip(reqCtx, remoteModel); err != nil {
				return fmt.Errorf("update service %s, release eip %s from elb %s, error %s", localModel.NamespacedName.String(),
					remoteModel.GetEipId(), remoteModel.GetLoadBalancerId(), err.Error())
			}
		}
		reqCtx.Log.Info(fmt.Sprintf("eip %s is associated to loadbalancer %s, can not overwrite it", remoteModel.GetEipId(), remoteModel.GetLoadBalancerId()))
		return nil
	}

	return nil
}

func (applier *ModelApplier) applyServerGroupAttribute(reqCtx *svcCtx.RequestContext, localModel, remoteModel *elbmodel.EdgeLoadBalancer) error {
	addServerGroup, removeServerGroup, updateServerGroup := getUpdateServerGroup(localModel, remoteModel)
	if !localModel.CanReUse() || !remoteModel.CanReUse() {
		if err := applier.SGMgr.batchRemoveServerGroup(reqCtx, remoteModel.GetLoadBalancerId(), removeServerGroup); err != nil {
			return fmt.Errorf("batch remove server group error : %s", err.Error())
		}
	}
	if err := applier.SGMgr.batchAddServerGroup(reqCtx, remoteModel.GetLoadBalancerId(), addServerGroup); err != nil {
		return fmt.Errorf("batch add server group error : %s", err.Error())
	}
	if err := applier.SGMgr.batchUpdateServerGroup(reqCtx, remoteModel.GetLoadBalancerId(), updateServerGroup); err != nil {
		return fmt.Errorf("batch update server group error : %s", err.Error())
	}
	return nil
}

func (applier *ModelApplier) applyListenersAttribute(reqCtx *svcCtx.RequestContext, localModel, remoteModel *elbmodel.EdgeLoadBalancer) error {
	addListener, removeListener, updateListener, err := getListeners(localModel, remoteModel)
	if err != nil {
		return fmt.Errorf("check for listener error %s", err.Error())
	}
	err = applier.LisMgr.batchRemoveListeners(reqCtx, remoteModel.GetLoadBalancerId(), &removeListener)
	if err != nil {
		return fmt.Errorf("batch remove listeners error : %s", err.Error())
	}

	err = applier.LisMgr.batchAddListeners(reqCtx, remoteModel.GetLoadBalancerId(), &addListener)
	if err != nil {
		return fmt.Errorf("batch add listeners error : %s", err.Error())
	}

	err = applier.LisMgr.batchUpdateListeners(reqCtx, remoteModel.GetLoadBalancerId(), &updateListener)
	if err != nil {
		return fmt.Errorf("batch update listeners error : %s", err.Error())
	}
	return nil
}
