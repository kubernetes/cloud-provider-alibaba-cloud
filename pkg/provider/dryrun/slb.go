package dryrun

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/slb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
)

func NewDryRunSLB(
	auth *base.ClientMgr, slb *slb.SLBProvider,
) *DryRunSLB {
	return &DryRunSLB{auth: auth, slb: slb}
}

var _ prvd.ILoadBalancer = &DryRunSLB{}

type DryRunSLB struct {
	auth *base.ClientMgr
	slb  *slb.SLBProvider
}

func (m *DryRunSLB) FindLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return m.slb.FindLoadBalancer(ctx, mdl)
}

func (m *DryRunSLB) CreateLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	mtype := "CreateLoadBalancer"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), "", "CreateSLB", ERROR, "")
	return fmt.Errorf("api %s should not be called", mtype)
}

func (m *DryRunSLB) DescribeLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return m.slb.DescribeLoadBalancer(ctx, mdl)
}

func (m *DryRunSLB) DeleteLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	mtype := "DeleteLoadBalancer"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), mdl.LoadBalancerAttribute.LoadBalancerId, "DeleteSLB", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) ModifyLoadBalancerInstanceSpec(ctx context.Context, lbId string, spec string) error {
	mtype := "ModifyLoadBalancerInstanceSpec"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), lbId, "ModifySLBSpec", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerDeleteProtection(ctx context.Context, lbId string, flag string) error {
	mtype := "SetLoadBalancerDeleteProtection"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), lbId, "SetSLBDeleteProtection", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerName(ctx context.Context, lbId string, name string) error {
	mtype := "SetLoadBalancerName"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), lbId, "SetSLBName", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) ModifyLoadBalancerInternetSpec(ctx context.Context, lbId string, chargeType string, bandwidth int) error {
	mtype := "ModifyLoadBalancerInternetSpec"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), lbId, "ModifyInternetSpec", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerModificationProtection(ctx context.Context, lbId string, flag string) error {
	mtype := "SetLoadBalancerModificationProtection"
	svc := getService(ctx)
	AddEvent(SLB, util.Key(svc), lbId, "SetSLBModificationProtection", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) AddTags(ctx context.Context, lbId string, tags string) error {
	return m.slb.AddTags(ctx, lbId, tags)
}

func (m *DryRunSLB) DescribeTags(ctx context.Context, lbId string) ([]model.Tag, error) {
	return m.slb.DescribeTags(ctx, lbId)
}

// Listener
func (m *DryRunSLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	return m.slb.DescribeLoadBalancerListeners(ctx, lbId)
}

func (m *DryRunSLB) StartLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	mtype := "StartLoadBalancerListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), port), lbId, "StartListener",
		ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) StopLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	mtype := "StopLoadBalancerListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), port), lbId, "StopListener",
		ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) DeleteLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	mtype := "DeleteLoadBalancerListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), port), lbId, "DeleteListener", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) CreateLoadBalancerTCPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "CreateLoadBalancerTCPListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"CreateListener", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerTCPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "SetLoadBalancerTCPListenerAttribute"
	svc := getService(ctx)
	reason := getDryRunMsg(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"UpdateListener", ERROR, reason)
	return fmt.Errorf("function should not be called on start %s, called reason: %s", mtype, reason)
}

func (m *DryRunSLB) CreateLoadBalancerUDPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "CreateLoadBalancerUDPListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"CreateListener", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerUDPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "SetLoadBalancerUDPListenerAttribute"
	svc := getService(ctx)
	reason := getDryRunMsg(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"UpdateListener", ERROR, reason)
	return fmt.Errorf("function should not be called on start %s, called reason: %s", mtype, reason)
}

func (m *DryRunSLB) CreateLoadBalancerHTTPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "CreateLoadBalancerHTTPListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"CreateListener", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerHTTPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "SetLoadBalancerHTTPListenerAttribute"
	svc := getService(ctx)
	reason := getDryRunMsg(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"UpdateListener", ERROR, reason)
	return fmt.Errorf("function should not be called on start %s, called reason: %s", mtype, reason)
}

func (m *DryRunSLB) CreateLoadBalancerHTTPSListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "CreateLoadBalancerHTTPSListener"
	svc := getService(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId, "CreateListener",
		ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetLoadBalancerHTTPSListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	mtype := "SetLoadBalancerHTTPSListenerAttribute"
	svc := getService(ctx)
	reason := getDryRunMsg(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%d", util.Key(svc), listener.ListenerPort), lbId,
		"UpdateListener", ERROR, reason)
	return fmt.Errorf("function should not be called on start %s, called reason: %s", mtype, reason)
}

// VServerGroup
func (m *DryRunSLB) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	return m.slb.DescribeVServerGroups(ctx, lbId)
}

/*
From v1.9.3.313-g748f81e-aliyun, ccm sets backend type to eni in newly created Terway clusters.
*/
func (m *DryRunSLB) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
	mtype := "CreateVServerGroup"
	AddEvent(SLB, vg.VGroupName, lbId, "CreateVgroup", NORMAL, "")
	klog.Warningf("%s try to call %s function, lb id: %s", vg.VGroupName, mtype, lbId)
	return nil
}

func (m *DryRunSLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	return m.slb.DescribeVServerGroupAttribute(ctx, vGroupId)
}

func (m *DryRunSLB) DeleteVServerGroup(ctx context.Context, vGroupId string) error {
	mtype := "DeleteVServerGroup"
	svc := getService(ctx)
	lbId := getSlb(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%s", util.Key(svc), vGroupId), lbId,
		"DeleteVgroup", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

/*
 From v1.9.3.239-g40d97e1-aliyun, ccm support ecs and eni together.
 If a svc who has ecs and eni backends together, it's normal to call the AddVServerGroupBackendServers api to add eci backend.
*/
func (m *DryRunSLB) AddVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	mtype := "AddVServerGroupBackendServers"
	svc := getService(ctx)
	lbId := getSlb(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%s", util.Key(svc), vGroupId), lbId,
		"AddVgroup", NORMAL, "")
	klog.Warningf("%s try to call %s function, vgroup id: %s, lb id: %s", util.Key(svc), mtype, vGroupId, lbId)
	return nil
}

func (m *DryRunSLB) RemoveVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	mtype := "RemoveVServerGroupBackendServers"
	svc := getService(ctx)
	lbId := getSlb(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/%s", util.Key(svc), vGroupId), lbId,
		"RemoveVgroup", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func (m *DryRunSLB) SetVServerGroupAttribute(ctx context.Context, vGroupId string, backends string) error {
	return m.slb.SetVServerGroupAttribute(ctx, vGroupId, backends)
}

func (m *DryRunSLB) ModifyVServerGroupBackendServers(ctx context.Context, vGroupId string, old string, new string) error {
	mtype := "ModifyVServerGroupBackendServers"
	svc := getService(ctx)
	lbId := getSlb(ctx)
	AddEvent(SLB, fmt.Sprintf("%s/VGroupID/%s", util.Key(svc), vGroupId), lbId,
		"ModifyVgroup", ERROR, "")
	return fmt.Errorf("function should not be called on start %s", mtype)
}

func getService(ctx context.Context) *v1.Service {
	isvc := ctx.Value(ContextService)
	if isvc == nil {
		return unknown()
	}
	svc, ok := isvc.(*v1.Service)
	if !ok {
		return unknown()
	}
	return svc
}

func getSlb(ctx context.Context) string {
	islb := ctx.Value(ContextSLB)
	if islb == nil {
		return ""
	}
	lbId, ok := islb.(string)
	if !ok {
		return ""
	}
	return lbId
}

func unknown() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unknown",
			Namespace: "unkonwn",
		},
	}
}

func getDryRunMsg(ctx context.Context) string {
	isMsg := ctx.Value(ContextMessage)
	if isMsg == nil {
		return ""
	}
	msg, ok := isMsg.(string)
	if !ok {
		return ""
	}
	return msg
}
