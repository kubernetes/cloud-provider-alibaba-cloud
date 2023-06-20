package clbv1

import (
	"context"
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/mohae/deepcopy"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"time"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"

	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/klog/v2"
)

const DefaultListenerBandwidth = -1

// IListener listener interface
type IListenerManager interface {
	Create(reqCtx *svcCtx.RequestContext, action CreateAction) error
	Update(reqCtx *svcCtx.RequestContext, action UpdateAction) error
}

type CreateAction struct {
	lbId     string
	listener model.ListenerAttribute
}

type UpdateAction struct {
	lbId   string
	local  model.ListenerAttribute
	remote model.ListenerAttribute
}

type DeleteAction struct {
	lbId     string
	listener model.ListenerAttribute
}

func NewListenerManager(cloud prvd.Provider) *ListenerManager {
	return &ListenerManager{
		cloud: cloud,
	}
}

type ListenerManager struct {
	cloud prvd.Provider
}

func (mgr *ListenerManager) Create(reqCtx *svcCtx.RequestContext, action CreateAction) error {
	switch strings.ToLower(action.listener.Protocol) {
	case model.TCP:
		return (&tcp{mgr}).Create(reqCtx, action)
	case model.UDP:
		return (&udp{mgr}).Create(reqCtx, action)
	case model.HTTP:
		return (&http{mgr}).Create(reqCtx, action)
	case model.HTTPS:
		return (&https{mgr}).Create(reqCtx, action)
	default:
		return fmt.Errorf("%s protocol is not supported", action.listener.Protocol)
	}

}

func (mgr *ListenerManager) Delete(reqCtx *svcCtx.RequestContext, action DeleteAction) error {
	reqCtx.Log.Info(fmt.Sprintf("delete listener %d", action.listener.ListenerPort))
	return mgr.cloud.DeleteLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort, action.listener.Protocol)
}

func (mgr *ListenerManager) Update(reqCtx *svcCtx.RequestContext, action UpdateAction) error {
	switch strings.ToLower(action.local.Protocol) {
	case model.TCP:
		return (&tcp{mgr}).Update(reqCtx, action)
	case model.UDP:
		return (&udp{mgr}).Update(reqCtx, action)
	case model.HTTP:
		return (&http{mgr}).Update(reqCtx, action)
	case model.HTTPS:
		return (&https{mgr}).Update(reqCtx, action)
	default:
		return fmt.Errorf("%s protocol is not supported", action.local.Protocol)
	}
}

// Describe Describe all listeners at once
func (mgr *ListenerManager) Describe(reqCtx *svcCtx.RequestContext, lbId string) ([]model.ListenerAttribute, error) {
	return mgr.cloud.DescribeLoadBalancerListeners(reqCtx.Ctx, lbId)
}

func (mgr *ListenerManager) BuildLocalModel(reqCtx *svcCtx.RequestContext, mdl *model.LoadBalancer) error {
	for _, port := range reqCtx.Service.Spec.Ports {
		listener, err := mgr.buildListenerFromServicePort(reqCtx, port)
		if err != nil {
			return fmt.Errorf("build listener from servicePort %d error: %s", port.Port, err.Error())
		}
		mdl.Listeners = append(mdl.Listeners, listener)
	}
	return nil
}

func (mgr *ListenerManager) BuildRemoteModel(reqCtx *svcCtx.RequestContext, mdl *model.LoadBalancer) error {
	listeners, err := mgr.Describe(reqCtx, mdl.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeLoadBalancerListeners error:%s", err.Error())
	}
	mdl.Listeners = listeners
	return nil
}

func (mgr *ListenerManager) buildListenerFromServicePort(reqCtx *svcCtx.RequestContext, port v1.ServicePort) (model.ListenerAttribute, error) {
	listener := model.ListenerAttribute{
		NamedKey: &model.ListenerNamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			CID:         base.CLUSTER_ID,
			Namespace:   reqCtx.Service.Namespace,
			ServiceName: reqCtx.Service.Name,
			Port:        port.Port,
		},
		ListenerPort: int(port.Port),
	}

	listener.Description = listener.NamedKey.Key()
	listener.VGroupName = getVGroupNamedKey(reqCtx.Service, port).Key()

	proto, err := protocol(reqCtx.Anno.Get(annotation.ProtocolPort), port)
	if err != nil {
		return listener, err
	}
	listener.Protocol = proto

	if reqCtx.Anno.Get(annotation.VGroupPort) != "" {
		vGroupId, err := vgroup(reqCtx.Anno.Get(annotation.VGroupPort), port)
		if err != nil {
			return listener, err
		}
		listener.VGroupId = vGroupId
	}

	if reqCtx.Anno.Get(annotation.Scheduler) != "" {
		listener.Scheduler = reqCtx.Anno.Get(annotation.Scheduler)
	}

	if reqCtx.Anno.Get(annotation.PersistenceTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.PersistenceTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation persistence timeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.PersistenceTimeout), err.Error())
		}
		listener.PersistenceTimeout = &timeout
	}

	if reqCtx.Anno.Get(annotation.EstablishedTimeout) != "" {
		establishedTimeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.EstablishedTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation EstablishedTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.EstablishedTimeout), err.Error())
		}
		listener.EstablishedTimeout = establishedTimeout
	}

	listener.CertId = reqCtx.Anno.Get(annotation.CertID)

	if reqCtx.Anno.Get(annotation.TLSCipherPolicy) != "" {
		listener.TLSCipherPolicy = reqCtx.Anno.Get(annotation.TLSCipherPolicy)
	}

	if reqCtx.Anno.Get(annotation.EnableHttp2) != "" {
		listener.EnableHttp2 = model.FlagType(reqCtx.Anno.Get(annotation.EnableHttp2))
	}

	if reqCtx.Anno.Get(annotation.ProxyProtocol) != "" {
		listener.EnableProxyProtocolV2 = tea.Bool(reqCtx.Anno.Get(annotation.ProxyProtocol) == string(model.OnFlag))
	}

	if reqCtx.Anno.Get(annotation.ForwardPort) != "" && listener.Protocol == model.HTTP {
		fp, err := forwardPort(reqCtx.Anno.Get(annotation.ForwardPort), int(port.Port))
		if err != nil {
			return listener, fmt.Errorf("Annotation ForwardPort error: %s ", err.Error())
		}
		if fp != 0 {
			listener.ForwardPort = fp
			listener.ListenerForward = model.OnFlag
		}
	}

	if reqCtx.Anno.Get(annotation.IdleTimeout) != "" {
		idleTimeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.IdleTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation IdleTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.IdleTimeout), err.Error())
		}
		listener.IdleTimeout = idleTimeout
	}

	if reqCtx.Anno.Get(annotation.RequestTimeout) != "" {
		requestTimeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.RequestTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation RequestTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.RequestTimeout), err.Error())
		}
		listener.RequestTimeout = requestTimeout
	}

	// acl
	if reqCtx.Anno.Get(annotation.AclStatus) != "" {
		listener.AclStatus = model.FlagType(reqCtx.Anno.Get(annotation.AclStatus))
	}
	if reqCtx.Anno.Get(annotation.AclType) != "" {
		listener.AclType = reqCtx.Anno.Get(annotation.AclType)
	}
	if reqCtx.Anno.Get(annotation.AclID) != "" {
		listener.AclId = reqCtx.Anno.Get(annotation.AclID)
	}

	// connection drain
	if reqCtx.Anno.Get(annotation.ConnectionDrain) != "" {
		listener.ConnectionDrain = model.FlagType(reqCtx.Anno.Get(annotation.ConnectionDrain))
	}
	if reqCtx.Anno.Get(annotation.ConnectionDrainTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.ConnectionDrainTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation ConnectionDrainTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.ConnectionDrainTimeout), err.Error())
		}
		listener.ConnectionDrainTimeout = timeout
	}

	// cookie
	if reqCtx.Anno.Get(annotation.Cookie) != "" {
		listener.Cookie = reqCtx.Anno.Get(annotation.Cookie)
	}
	if reqCtx.Anno.Get(annotation.CookieTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.CookieTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation CookieTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.CookieTimeout), err.Error())
		}
		listener.CookieTimeout = timeout
	}
	if reqCtx.Anno.Get(annotation.SessionStick) != "" {
		listener.StickySession = model.FlagType(reqCtx.Anno.Get(annotation.SessionStick))
	}
	if reqCtx.Anno.Get(annotation.SessionStickType) != "" {
		listener.StickySessionType = reqCtx.Anno.Get(annotation.SessionStickType)
	}

	// x-forwarded-for
	if reqCtx.Anno.Get(annotation.XForwardedForProto) != "" {
		listener.XForwardedForProto = model.FlagType(reqCtx.Anno.Get(annotation.XForwardedForProto))
	}

	// health check
	if reqCtx.Anno.Get(annotation.HealthyThreshold) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthyThreshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthyThreshold), err.Error())
		}
		listener.HealthyThreshold = t
	}
	if reqCtx.Anno.Get(annotation.UnhealthyThreshold) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(annotation.UnhealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("Annotation UnhealthyThreshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.UnhealthyThreshold), err.Error())
		}
		listener.UnhealthyThreshold = t
	}
	if reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckConnectTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckConnectTimeout), err.Error())
		}
		listener.HealthCheckConnectTimeout = timeout
	}
	if reqCtx.Anno.Get(annotation.HealthCheckConnectPort) != "" {
		port, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckConnectPort))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckConnectPort must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckConnectPort), err.Error())
		}
		listener.HealthCheckConnectPort = port
	}
	if reqCtx.Anno.Get(annotation.HealthCheckInterval) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckInterval))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckInterval must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckInterval), err.Error())
		}
		listener.HealthCheckInterval = t
	}
	if reqCtx.Anno.Get(annotation.HealthCheckDomain) != "" {
		listener.HealthCheckDomain = reqCtx.Anno.Get(annotation.HealthCheckDomain)
	}
	if reqCtx.Anno.Get(annotation.HealthCheckURI) != "" {
		listener.HealthCheckURI = reqCtx.Anno.Get(annotation.HealthCheckURI)
	}
	if reqCtx.Anno.Get(annotation.HealthCheckHTTPCode) != "" {
		listener.HealthCheckHttpCode = reqCtx.Anno.Get(annotation.HealthCheckHTTPCode)
	}
	if reqCtx.Anno.Get(annotation.HealthCheckType) != "" {
		listener.HealthCheckType = reqCtx.Anno.Get(annotation.HealthCheckType)
	}
	if reqCtx.Anno.Get(annotation.HealthCheckFlag) != "" {
		listener.HealthCheck = model.FlagType(reqCtx.Anno.Get(annotation.HealthCheckFlag))
	}
	if reqCtx.Anno.Get(annotation.HealthCheckTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(annotation.HealthCheckTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(annotation.HealthCheckTimeout), err.Error())
		}
		listener.HealthCheckTimeout = timeout
	}
	if reqCtx.Anno.Get(annotation.HealthCheckMethod) != "" {
		listener.HealthCheckMethod = reqCtx.Anno.Get(annotation.HealthCheckMethod)
	}

	if reqCtx.Anno.Get(annotation.HealthCheckSwitch) != "" {
		listener.HealthCheckSwitch = model.FlagType(reqCtx.Anno.Get(annotation.HealthCheckSwitch))
	}

	return listener, nil
}

type tcp struct {
	mgr *ListenerManager
}

func (t *tcp) Create(reqCtx *svcCtx.RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerTCPListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create tcp listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener tcp [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort, action.listener.Protocol)
}

func (t *tcp) Update(reqCtx *svcCtx.RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort, action.local.Protocol)
		if err != nil {
			return fmt.Errorf("start tcp listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: tcp [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}
	return t.mgr.cloud.SetLoadBalancerTCPListenerAttribute(reqCtx.Ctx, action.lbId, update)
}

type udp struct {
	mgr *ListenerManager
}

func (t *udp) Create(reqCtx *svcCtx.RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerUDPListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create udp listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener udp [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort, action.listener.Protocol)
}

func (t *udp) Update(reqCtx *svcCtx.RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort, action.local.Protocol)
		if err != nil {
			return fmt.Errorf("start udp listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: udp [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}
	return t.mgr.cloud.SetLoadBalancerUDPListenerAttribute(reqCtx.Ctx, action.lbId, update)
}

type http struct {
	mgr *ListenerManager
}

func (t *http) Create(reqCtx *svcCtx.RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerHTTPListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create http listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener http [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort, action.listener.Protocol)

}

func (t *http) Update(reqCtx *svcCtx.RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort, action.local.Protocol)
		if err != nil {
			return fmt.Errorf("start http listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}

	// The forwarding rule could not be updated. If the rule is changed, it needs to be recreated.
	needRecreate := false
	if action.local.ForwardPort != 0 {
		if action.remote.ListenerForward != model.OnFlag ||
			action.remote.ForwardPort != action.local.ForwardPort {
			needRecreate = true
		}
	} else {
		if action.remote.ListenerForward != model.OffFlag {
			needRecreate = true
			action.local.ListenerForward = model.OffFlag
		}
	}

	if needRecreate {
		reqCtx.Log.Info(fmt.Sprintf("port [%d] forward policy changed, need recreate", action.local.ListenerPort))
		err := t.mgr.Delete(reqCtx, DeleteAction{
			lbId:     action.lbId,
			listener: action.remote,
		})
		if err != nil {
			return fmt.Errorf("delete port [%d] error: %s", action.remote.ListenerPort, err.Error())
		}

		return t.Create(reqCtx, CreateAction{
			lbId:     action.lbId,
			listener: action.local,
		})
	}

	if action.remote.ListenerForward == model.OnFlag {
		reqCtx.Log.Info(fmt.Sprintf("update listener: http [%d] ListenerForward is on, skip update listener", action.local.ListenerPort))
		return nil
	}

	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: http [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}
	return t.mgr.cloud.SetLoadBalancerHTTPListenerAttribute(reqCtx.Ctx, action.lbId, update)

}

type https struct {
	mgr *ListenerManager
}

func (t *https) Create(reqCtx *svcCtx.RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerHTTPSListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create https listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener https [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort, action.listener.Protocol)

}

func (t *https) Update(reqCtx *svcCtx.RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort, action.local.Protocol)
		if err != nil {
			return fmt.Errorf("start https listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: https [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}

	if action.local.CertId != action.remote.CertId {
		err := checkCertValidity(t.mgr.cloud, action.remote.CertId, action.local.CertId)
		if err != nil {
			return err
		}
	}

	return t.mgr.cloud.SetLoadBalancerHTTPSListenerAttribute(reqCtx.Ctx, action.lbId, update)
}

func forwardPort(port string, target int) (int, error) {
	if port == "" {
		return 0, fmt.Errorf("forward port format error, get: %s, expect 80:443", port)
	}
	forwarded := ""
	tmps := strings.Split(port, ",")
	for _, v := range tmps {
		ports := strings.Split(v, ":")
		if len(ports) != 2 {
			return 0, fmt.Errorf("forward port format error: %s, expect 80:443,88:6443", port)
		}
		if ports[0] == strconv.Itoa(target) {
			forwarded = ports[1]
			break
		}
	}
	if forwarded != "" {
		forward, err := strconv.Atoi(forwarded)
		if err != nil {
			return 0, fmt.Errorf("forward port is not an integer, %s", forwarded)
		}
		klog.Infof("forward http port %d to %d", target, forward)
		return forward, nil
	}
	return 0, nil
}

func buildActionsForListeners(reqCtx *svcCtx.RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) ([]CreateAction, []UpdateAction, []DeleteAction, error) {
	var (
		createActions []CreateAction
		updateActions []UpdateAction
		deleteActions []DeleteAction
	)
	// associate listener and vGroup
	for i := range local.Listeners {
		if local.Listeners[i].VGroupId != "" {
			continue
		}
		if err := findVServerGroup(remote.VServerGroups, &local.Listeners[i]); err != nil {
			return createActions, updateActions, deleteActions, fmt.Errorf("find vservergroup error: %s", err.Error())
		}
	}

	// For update and deletions
	for _, rlis := range remote.Listeners {
		found := false
		for _, llis := range local.Listeners {
			if rlis.ListenerPort == llis.ListenerPort && rlis.Protocol == llis.Protocol {
				found = true
				// listener and protocol match, do update
				updateActions = append(updateActions,
					UpdateAction{
						lbId:   remote.LoadBalancerAttribute.LoadBalancerId,
						local:  llis,
						remote: rlis,
					})
			}
		}
		// Do not delete any listener that no longer managed by my service
		// for safety. Only conflict case is taking care.
		if !found {
			if !isPortManagedByMyService(local, rlis) {
				reqCtx.Log.Info(fmt.Sprintf("update listener: port [%d] namekey [%s] is managed by user, skip reconcile", rlis.ListenerPort, rlis.NamedKey))
				continue
			}

			deleteActions = append(deleteActions, DeleteAction{
				lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
				listener: rlis,
			})
		}
	}

	// For additions
	for _, llis := range local.Listeners {
		found := false
		for _, rlis := range remote.Listeners {
			if llis.ListenerPort == rlis.ListenerPort && llis.Protocol == rlis.Protocol {
				// port and protocol matched. updated. skip
				found = true
			}
		}
		if !found {
			createActions = append(createActions, CreateAction{
				lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
				listener: llis,
			})
		}
	}
	return createActions, updateActions, deleteActions, nil
}

// Protocol for protocol transform
func protocol(annotation string, port v1.ServicePort) (string, error) {

	if annotation == "" {
		return strings.ToLower(string(port.Protocol)), nil
	}
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return "", fmt.Errorf("port and "+
				"protocol format must be like 'https:443' with colon separated. got=[%+v]", pp)
		}

		if pp[0] != model.HTTP &&
			pp[0] != model.TCP &&
			pp[0] != model.HTTPS &&
			pp[0] != model.UDP {
			return "", fmt.Errorf("port protocol"+
				" format must be either [http|https|tcp|udp], protocol not supported wit [%s]\n", pp[0])
		}

		if pp[1] == fmt.Sprintf("%d", port.Port) {
			util.ServiceLog.Info(fmt.Sprintf("port [%d] transform protocol from %s to %s", port.Port, port.Protocol, pp[0]))
			return pp[0], nil
		}
	}
	return strings.ToLower(string(port.Protocol)), nil
}

// vgroup find the vGroup id associated with the specific ServicePort
func vgroup(annotation string, port v1.ServicePort) (string, error) {
	for _, v := range strings.Split(annotation, ",") {
		pp := strings.Split(v, ":")
		if len(pp) < 2 {
			return "", fmt.Errorf("vgroupid and "+
				"protocol format must be like 'vsp-xxx:443' with colon separated. got=[%+v]", pp)
		}

		if pp[1] == fmt.Sprintf("%d", port.Port) {
			return pp[0], nil
		}
	}
	return "", nil
}

func getVGroupNamedKey(svc *v1.Service, servicePort v1.ServicePort) *model.VGroupNamedKey {
	vGroupPort := ""
	if helper.IsENIBackendType(svc) {
		switch servicePort.TargetPort.Type {
		case intstr.Int:
			vGroupPort = fmt.Sprintf("%d", servicePort.TargetPort.IntValue())
		case intstr.String:
			vGroupPort = servicePort.TargetPort.StrVal
		}
	} else {
		vGroupPort = fmt.Sprintf("%d", servicePort.NodePort)
	}
	return &model.VGroupNamedKey{
		Prefix:      model.DEFAULT_PREFIX,
		Namespace:   svc.Namespace,
		CID:         base.CLUSTER_ID,
		VGroupPort:  vGroupPort,
		ServiceName: svc.Name}
}

func setDefaultValueForListener(n *model.ListenerAttribute) {
	// set default scheduler algorithm to rr
	// When the weight values are equal, the rr algorithm is better than the wrr algorithm
	if n.Scheduler == "" {
		n.Scheduler = "rr"
	}

	if n.Protocol == model.TCP || n.Protocol == model.UDP || n.Protocol == model.HTTPS {
		if n.Bandwidth == 0 {
			n.Bandwidth = DefaultListenerBandwidth
		}
	}

	if helper.Is7LayerProtocol(n.Protocol) {
		if n.HealthCheck == "" {
			n.HealthCheck = model.OffFlag
		}
		if n.StickySession == "" {
			n.StickySession = model.OffFlag
		}
	}
}

func isNeedUpdate(reqCtx *svcCtx.RequestContext, local model.ListenerAttribute, remote model.ListenerAttribute) (bool, model.ListenerAttribute) {
	update := deepcopy.Copy(remote).(model.ListenerAttribute)
	needUpdate := false
	updateDetail := ""

	if remote.Description != local.Description {
		needUpdate = true
		update.Description = local.Description
		updateDetail += fmt.Sprintf("Description %v should be changed to %v;",
			remote.Description, local.Description)
	}
	if remote.VGroupId != local.VGroupId {
		needUpdate = true
		update.VGroupId = local.VGroupId
		updateDetail += fmt.Sprintf("VGroupId %v should be changed to %v;",
			remote.VGroupId, local.VGroupId)
	}
	if local.Scheduler != "" &&
		remote.Scheduler != local.Scheduler {
		needUpdate = true
		update.Scheduler = local.Scheduler
		updateDetail += fmt.Sprintf("lb Scheduler %v should be changed to %v;",
			remote.Scheduler, local.Scheduler)
	}
	if local.Protocol == model.TCP &&
		local.PersistenceTimeout != nil && remote.PersistenceTimeout != nil &&
		*remote.PersistenceTimeout != *local.PersistenceTimeout {
		needUpdate = true
		update.PersistenceTimeout = local.PersistenceTimeout
		updateDetail += fmt.Sprintf("lb PersistenceTimeout %v should be changed to %v;",
			*remote.PersistenceTimeout, *local.PersistenceTimeout)
	}
	if local.Protocol == model.TCP &&
		local.EstablishedTimeout != 0 &&
		remote.EstablishedTimeout != local.EstablishedTimeout {
		needUpdate = true
		update.EstablishedTimeout = local.EstablishedTimeout
		updateDetail += fmt.Sprintf("EstablishedTimeout changed: %v - %v ;",
			remote.EstablishedTimeout, local.EstablishedTimeout)
	}
	if (local.Protocol == model.TCP || local.Protocol == model.UDP) &&
		local.EnableProxyProtocolV2 != nil &&
		tea.BoolValue(local.EnableProxyProtocolV2) != tea.BoolValue(remote.EnableProxyProtocolV2) {
		needUpdate = true
		update.EnableProxyProtocolV2 = tea.Bool(*local.EnableProxyProtocolV2)
		updateDetail += fmt.Sprintf("lb EnableProxyProtocolV2 %v should be changed to %v",
			tea.BoolValue(remote.EnableProxyProtocolV2), tea.BoolValue(local.EnableProxyProtocolV2))
	}

	// only for https
	if local.Protocol == model.HTTPS {
		// The cert id is necessary for https, so skip to check whether it is blank
		if remote.CertId != local.CertId {
			needUpdate = true
			update.CertId = local.CertId
			updateDetail += fmt.Sprintf("lb CertId %v should be changed to %v;",
				remote.CertId, local.CertId)
		}

		if local.EnableHttp2 != "" &&
			remote.EnableHttp2 != local.EnableHttp2 {
			needUpdate = true
			update.EnableHttp2 = local.EnableHttp2
			updateDetail += fmt.Sprintf("lb EnableHttp2 %v should be changed to %v;",
				remote.EnableHttp2, local.EnableHttp2)
		}

		if local.TLSCipherPolicy != "" &&
			remote.TLSCipherPolicy != local.TLSCipherPolicy {
			needUpdate = true
			update.TLSCipherPolicy = local.TLSCipherPolicy
			updateDetail += fmt.Sprintf("lb TLSCipherPolicy %v should be changed to %v;",
				remote.TLSCipherPolicy, local.TLSCipherPolicy)
		}

	}
	// acl
	if local.AclStatus != "" &&
		remote.AclStatus != local.AclStatus {
		needUpdate = true
		update.AclStatus = local.AclStatus
		updateDetail += fmt.Sprintf("lb AclStatus %v should be changed to %v;",
			remote.AclStatus, local.AclStatus)
	}
	if local.AclStatus == model.OnFlag &&
		local.AclId != "" &&
		remote.AclId != local.AclId {
		needUpdate = true
		update.AclId = local.AclId
		updateDetail += fmt.Sprintf("lb AclId %v should be changed to %v;",
			remote.AclId, local.AclId)
	}
	if local.AclStatus == model.OnFlag &&
		local.AclType != "" &&
		remote.AclType != local.AclType {
		needUpdate = true
		update.AclType = local.AclType
		updateDetail += fmt.Sprintf("lb AclType %v should be changed to %v;",
			remote.AclType, local.AclType)
	}
	// idle timeout
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.IdleTimeout != 0 &&
		remote.IdleTimeout != local.IdleTimeout {
		needUpdate = true
		update.IdleTimeout = local.IdleTimeout
		updateDetail += fmt.Sprintf("lb IdleTimeout %v should be changed to %v;",
			remote.IdleTimeout, local.IdleTimeout)
	}
	// request timeout
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.RequestTimeout != 0 &&
		remote.RequestTimeout != local.RequestTimeout {
		needUpdate = true
		update.RequestTimeout = local.RequestTimeout
		updateDetail += fmt.Sprintf("RequestTimeout changed: %v - %v ;",
			remote.RequestTimeout, local.RequestTimeout)
	}
	// session
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.StickySession != "" &&
		remote.StickySession != local.StickySession {
		needUpdate = true
		update.StickySession = local.StickySession
		updateDetail += fmt.Sprintf("lb StickySession %v should be changed to %v;",
			remote.StickySession, local.StickySession)
	}
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.StickySessionType != "" &&
		remote.StickySessionType != local.StickySessionType {
		needUpdate = true
		update.StickySessionType = local.StickySessionType
		updateDetail += fmt.Sprintf("lb StickySessionType %v should be changed to %v;",
			remote.StickySessionType, local.StickySessionType)
	}
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.Cookie != "" &&
		remote.Cookie != local.Cookie {
		needUpdate = true
		update.Cookie = local.Cookie
		updateDetail += fmt.Sprintf("lb Cookie %v should be changed to %v;",
			remote.Cookie, local.Cookie)
	}
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.CookieTimeout != 0 &&
		remote.CookieTimeout != local.CookieTimeout {
		needUpdate = true
		update.CookieTimeout = local.CookieTimeout
		updateDetail += fmt.Sprintf("lb CookieTimeout %v should be changed to %v;",
			remote.CookieTimeout, local.CookieTimeout)
	}
	// connection drain
	if helper.Is4LayerProtocol(local.Protocol) &&
		local.ConnectionDrain != "" &&
		remote.ConnectionDrain != local.ConnectionDrain {
		needUpdate = true
		update.ConnectionDrain = local.ConnectionDrain
		updateDetail += fmt.Sprintf("lb ConnectionDrain %v should be changed to %v;",
			remote.ConnectionDrain, local.ConnectionDrain)
	}
	if helper.Is4LayerProtocol(local.Protocol) &&
		local.ConnectionDrain == model.OnFlag &&
		local.ConnectionDrainTimeout != 0 &&
		remote.ConnectionDrainTimeout != local.ConnectionDrainTimeout {
		needUpdate = true
		update.ConnectionDrainTimeout = local.ConnectionDrainTimeout
		updateDetail += fmt.Sprintf("lb ConnectionDrainTimeout %v should be changed to %v;",
			remote.ConnectionDrainTimeout, local.ConnectionDrainTimeout)
	}

	// x-forwarded-for
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.XForwardedForProto != "" &&
		remote.XForwardedForProto != local.XForwardedForProto {
		needUpdate = true
		update.XForwardedForProto = local.XForwardedForProto
		updateDetail += fmt.Sprintf("lb XForwardedForProto %v should be changed to %v;",
			remote.XForwardedForProto, local.XForwardedForProto)
	}

	// health check
	if local.HealthCheckConnectPort != 0 &&
		remote.HealthCheckConnectPort != local.HealthCheckConnectPort {
		needUpdate = true
		update.HealthCheckConnectPort = local.HealthCheckConnectPort
		updateDetail += fmt.Sprintf("lb HealthCheckConnectPort %v should be changed to %v;",
			remote.HealthCheckConnectPort, local.HealthCheckConnectPort)
	}
	if local.HealthCheckInterval != 0 &&
		remote.HealthCheckInterval != local.HealthCheckInterval {
		needUpdate = true
		update.HealthCheckInterval = local.HealthCheckInterval
		updateDetail += fmt.Sprintf("lb HealthCheckInterval %v should be changed to %v;",
			remote.HealthCheckInterval, local.HealthCheckInterval)
	}
	if local.HealthyThreshold != 0 &&
		remote.HealthyThreshold != local.HealthyThreshold {
		needUpdate = true
		update.HealthyThreshold = local.HealthyThreshold
		updateDetail += fmt.Sprintf("lb HealthyThreshold %v should be changed to %v;",
			remote.HealthyThreshold, local.HealthyThreshold)
	}
	if local.UnhealthyThreshold != 0 &&
		remote.UnhealthyThreshold != local.UnhealthyThreshold {
		needUpdate = true
		update.UnhealthyThreshold = local.UnhealthyThreshold
		updateDetail += fmt.Sprintf("lb UnhealthyThreshold %v should be changed to %v;",
			remote.UnhealthyThreshold, local.UnhealthyThreshold)
	}
	if local.Protocol == model.TCP &&
		local.HealthCheckType != "" &&
		remote.HealthCheckType != local.HealthCheckType {
		needUpdate = true
		update.HealthCheckType = local.HealthCheckType
		updateDetail += fmt.Sprintf("lb HealthCheckType %v should be changed to %v;",
			remote.HealthCheckType, local.HealthCheckType)
	}
	if local.Protocol != model.UDP &&
		local.HealthCheckDomain != "" &&
		remote.HealthCheckDomain != local.HealthCheckDomain {
		needUpdate = true
		update.HealthCheckDomain = local.HealthCheckDomain
		updateDetail += fmt.Sprintf("lb HealthCheckDomain %v should be changed to %v;",
			remote.HealthCheckDomain, local.HealthCheckDomain)
	}
	if local.Protocol != model.UDP &&
		local.HealthCheckURI != "" &&
		remote.HealthCheckURI != local.HealthCheckURI {
		needUpdate = true
		update.HealthCheckURI = local.HealthCheckURI
		updateDetail += fmt.Sprintf("lb HealthCheckURI %v should be changed to %v;",
			remote.HealthCheckURI, local.HealthCheckURI)
	}
	if local.Protocol != model.UDP &&
		local.HealthCheckHttpCode != "" &&
		remote.HealthCheckHttpCode != local.HealthCheckHttpCode {
		needUpdate = true
		update.HealthCheckHttpCode = local.HealthCheckHttpCode
		updateDetail += fmt.Sprintf("lb HealthCheckHttpCode %v should be changed to %v;",
			remote.HealthCheckHttpCode, local.HealthCheckHttpCode)
	}
	if helper.Is4LayerProtocol(local.Protocol) &&
		local.HealthCheckConnectTimeout != 0 &&
		remote.HealthCheckConnectTimeout != local.HealthCheckConnectTimeout {
		needUpdate = true
		update.HealthCheckConnectTimeout = local.HealthCheckConnectTimeout
		updateDetail += fmt.Sprintf("lb HealthCheckConnectTimeout %v should be changed to %v;",
			remote.HealthCheckConnectTimeout, local.HealthCheckConnectTimeout)
	}
	if helper.Is4LayerProtocol(local.Protocol) &&
		local.HealthCheckSwitch != "" &&
		remote.HealthCheckSwitch != local.HealthCheckSwitch {
		needUpdate = true
		update.HealthCheckSwitch = local.HealthCheckSwitch
		updateDetail += fmt.Sprintf("lb HealthCheckSwitch %v should be changed to %v;",
			remote.HealthCheckSwitch, local.HealthCheckSwitch)
	}
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.HealthCheck != "" &&
		remote.HealthCheck != local.HealthCheck {
		needUpdate = true
		update.HealthCheck = local.HealthCheck
		updateDetail += fmt.Sprintf("lb HealthCheck %v should be changed to %v;",
			remote.HealthCheck, local.HealthCheck)
	}
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.HealthCheckTimeout != 0 &&
		remote.HealthCheckTimeout != local.HealthCheckTimeout {
		needUpdate = true
		update.HealthCheckTimeout = local.HealthCheckTimeout
		updateDetail += fmt.Sprintf("lb HealthCheckTimeout %v should be changed to %v;",
			remote.HealthCheckTimeout, local.HealthCheckTimeout)
	}
	if helper.Is7LayerProtocol(local.Protocol) &&
		local.HealthCheckMethod != "" &&
		remote.HealthCheckMethod != local.HealthCheckMethod {
		needUpdate = true
		update.HealthCheckMethod = local.HealthCheckMethod
		updateDetail += fmt.Sprintf("HealthCheckMethod changed: %v - %v ;",
			remote.HealthCheckMethod, local.HealthCheckMethod)
	}

	if needUpdate {
		reqCtx.Ctx = context.WithValue(reqCtx.Ctx, dryrun.ContextMessage, updateDetail)
		reqCtx.Log.Info(fmt.Sprintf("update listener: %s [%d] changed, detail %s", local.Protocol, local.ListenerPort, updateDetail))
	}
	return needUpdate, update
}

func findVServerGroup(vgs []model.VServerGroup, port *model.ListenerAttribute) error {
	for _, vg := range vgs {
		if vg.VGroupName == port.VGroupName {
			port.VGroupId = vg.VGroupId
			return nil
		}
	}
	return fmt.Errorf("can not find vgroup by name %s", port.VGroupName)
}

// ==========================================================================================

func isPortManagedByMyService(local *model.LoadBalancer, n model.ListenerAttribute) bool {
	if n.IsUserManaged || n.NamedKey == nil {
		return false
	}

	return n.NamedKey.ServiceName == local.NamespacedName.Name &&
		n.NamedKey.Namespace == local.NamespacedName.Namespace &&
		n.NamedKey.CID == base.CLUSTER_ID
}

func checkCertValidity(cloudClient prvd.Provider, oldCertId, newCertId string) error {
	oldCert, err := cloudClient.DescribeServerCertificateById(context.TODO(), oldCertId)
	if err != nil {
		return fmt.Errorf("describe old cert %s error: %s", oldCertId, err.Error())
	}

	newCert, err := cloudClient.DescribeServerCertificateById(context.TODO(), newCertId)
	if err != nil {
		return fmt.Errorf("describe new cert %s error: %s", newCertId, err.Error())
	}

	if oldCert == nil {
		return nil
	}

	if newCert == nil {
		return fmt.Errorf("can not found cert by id %s", newCertId)
	}

	if newCert.ExpireTimeStamp < time.Now().UnixMilli() {
		return fmt.Errorf("can not update cert %s because it is expired", newCertId)
	}

	return nil
}
