package service

import (
	"context"
	"fmt"
	"github.com/mohae/deepcopy"
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

// IListener listener interface
type IListenerManager interface {
	Create(reqCtx *RequestContext, action CreateAction) error
	Update(reqCtx *RequestContext, action UpdateAction) error
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

func (mgr *ListenerManager) Create(reqCtx *RequestContext, action CreateAction) error {
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

func (mgr *ListenerManager) Delete(reqCtx *RequestContext, action DeleteAction) error {
	reqCtx.Log.Info(fmt.Sprintf("delete listener %d", action.listener.ListenerPort))
	return mgr.cloud.DeleteLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)
}

func (mgr *ListenerManager) Update(reqCtx *RequestContext, action UpdateAction) error {
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
func (mgr *ListenerManager) Describe(reqCtx *RequestContext, lbId string) ([]model.ListenerAttribute, error) {
	return mgr.cloud.DescribeLoadBalancerListeners(reqCtx.Ctx, lbId)
}

func (mgr *ListenerManager) BuildLocalModel(reqCtx *RequestContext, mdl *model.LoadBalancer) error {
	for _, port := range reqCtx.Service.Spec.Ports {
		listener, err := mgr.buildListenerFromServicePort(reqCtx, port)
		if err != nil {
			return fmt.Errorf("build listener from servicePort %d error: %s", port.Port, err.Error())
		}
		mdl.Listeners = append(mdl.Listeners, listener)
	}
	return nil
}

func (mgr *ListenerManager) BuildRemoteModel(reqCtx *RequestContext, mdl *model.LoadBalancer) error {
	listeners, err := mgr.Describe(reqCtx, mdl.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeLoadBalancerListeners error:%s", err.Error())
	}
	mdl.Listeners = listeners
	return nil
}

func (mgr *ListenerManager) buildListenerFromServicePort(reqCtx *RequestContext, port v1.ServicePort) (model.ListenerAttribute, error) {
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

	proto, err := protocol(reqCtx.Anno.Get(ProtocolPort), port)
	if err != nil {
		return listener, err
	}
	listener.Protocol = proto

	if reqCtx.Anno.Get(VGroupPort) != "" {
		vGroupId, err := vgroup(reqCtx.Anno.Get(VGroupPort), port)
		if err != nil {
			return listener, err
		}
		listener.VGroupId = vGroupId
	}

	if reqCtx.Anno.Get(Scheduler) != "" {
		listener.Scheduler = reqCtx.Anno.Get(Scheduler)
	}

	if reqCtx.Anno.Get(PersistenceTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(PersistenceTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation persistence timeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(PersistenceTimeout), err.Error())
		}
		listener.PersistenceTimeout = &timeout
	}

	if reqCtx.Anno.Get(CertID) != "" {
		listener.CertId = reqCtx.Anno.Get(CertID)
	}

	if reqCtx.Anno.Get(EnableHttp2) != "" {
		listener.EnableHttp2 = model.FlagType(reqCtx.Anno.Get(EnableHttp2))
	}

	if reqCtx.Anno.Get(ForwardPort) != "" && listener.Protocol == model.HTTP {
		forwardPort, err := forwardPort(reqCtx.Anno.Get(ForwardPort), int(port.Port))
		if err != nil {
			return listener, fmt.Errorf("Annotation ForwardPort error: %s ", err.Error())
		}
		listener.ForwardPort = forwardPort
		listener.ListenerForward = model.OnFlag
	}

	if reqCtx.Anno.Get(IdleTimeout) != "" {
		idleTimeout, err := strconv.Atoi(reqCtx.Anno.Get(IdleTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation IdleTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(IdleTimeout), err.Error())
		}
		listener.IdleTimeout = idleTimeout
	}

	// acl
	if reqCtx.Anno.Get(AclStatus) != "" {
		listener.AclStatus = model.FlagType(reqCtx.Anno.Get(AclStatus))
	}
	if reqCtx.Anno.Get(AclType) != "" {
		listener.AclType = reqCtx.Anno.Get(AclType)
	}
	if reqCtx.Anno.Get(AclID) != "" {
		listener.AclId = reqCtx.Anno.Get(AclID)
	}

	// connection drain
	if reqCtx.Anno.Get(ConnectionDrain) != "" {
		listener.ConnectionDrain = model.FlagType(reqCtx.Anno.Get(ConnectionDrain))
	}
	if reqCtx.Anno.Get(ConnectionDrainTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(ConnectionDrainTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation ConnectionDrainTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(ConnectionDrainTimeout), err.Error())
		}
		listener.ConnectionDrainTimeout = timeout
	}

	// cookie
	if reqCtx.Anno.Get(Cookie) != "" {
		listener.Cookie = reqCtx.Anno.Get(Cookie)
	}
	if reqCtx.Anno.Get(CookieTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(CookieTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation CookieTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(CookieTimeout), err.Error())
		}
		listener.CookieTimeout = timeout
	}
	if reqCtx.Anno.Get(SessionStick) != "" {
		listener.StickySession = model.FlagType(reqCtx.Anno.Get(SessionStick))
	}
	if reqCtx.Anno.Get(SessionStickType) != "" {
		listener.StickySessionType = reqCtx.Anno.Get(SessionStickType)
	}

	// x-forwarded-for
	if reqCtx.Anno.Get(XForwardedForProto) != "" {
		listener.XForwardedForProto = model.FlagType(reqCtx.Anno.Get(XForwardedForProto))
	}

	// health check
	if reqCtx.Anno.Get(HealthyThreshold) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(HealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthyThreshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(HealthyThreshold), err.Error())
		}
		listener.HealthyThreshold = t
	}
	if reqCtx.Anno.Get(UnhealthyThreshold) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(UnhealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("Annotation UnhealthyThreshold must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(UnhealthyThreshold), err.Error())
		}
		listener.UnhealthyThreshold = t
	}
	if reqCtx.Anno.Get(HealthCheckConnectTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckConnectTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckConnectTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(HealthCheckConnectTimeout), err.Error())
		}
		listener.HealthCheckConnectTimeout = timeout
	}
	if reqCtx.Anno.Get(HealthCheckConnectPort) != "" {
		port, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckConnectPort))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckConnectPort must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(HealthCheckConnectPort), err.Error())
		}
		listener.HealthCheckConnectPort = port
	}
	if reqCtx.Anno.Get(HealthCheckInterval) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckInterval))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckInterval must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(HealthCheckInterval), err.Error())
		}
		listener.HealthCheckInterval = t
	}
	if reqCtx.Anno.Get(HealthCheckDomain) != "" {
		listener.HealthCheckDomain = reqCtx.Anno.Get(HealthCheckDomain)
	}
	if reqCtx.Anno.Get(HealthCheckURI) != "" {
		listener.HealthCheckURI = reqCtx.Anno.Get(HealthCheckURI)
	}
	if reqCtx.Anno.Get(HealthCheckHTTPCode) != "" {
		listener.HealthCheckHttpCode = reqCtx.Anno.Get(HealthCheckHTTPCode)
	}
	if reqCtx.Anno.Get(HealthCheckType) != "" {
		listener.HealthCheckType = reqCtx.Anno.Get(HealthCheckType)
	}
	if reqCtx.Anno.Get(HealthCheckFlag) != "" {
		listener.HealthCheck = model.FlagType(reqCtx.Anno.Get(HealthCheckFlag))
	}
	if reqCtx.Anno.Get(HealthCheckTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckTimeout must be integer, but got [%s]. message=[%s] ",
				reqCtx.Anno.Get(HealthCheckTimeout), err.Error())
		}
		listener.HealthCheckTimeout = timeout
	}

	return listener, nil
}

type tcp struct {
	mgr *ListenerManager
}

func (t *tcp) Create(reqCtx *RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerTCPListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create tcp listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener tcp [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)
}

func (t *tcp) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: tcp [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}
	reqCtx.Log.Info(fmt.Sprintf("update listener: tcp [%d] try to update [%#v]", update.ListenerPort, update))
	return t.mgr.cloud.SetLoadBalancerTCPListenerAttribute(reqCtx.Ctx, action.lbId, update)
}

type udp struct {
	mgr *ListenerManager
}

func (t *udp) Create(reqCtx *RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerUDPListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create udp listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener udp [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)
}

func (t *udp) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start udp listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: udp [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}
	reqCtx.Log.Info(fmt.Sprintf("update listener: udp [%d] updated [%v]", update.ListenerPort, update))
	return t.mgr.cloud.SetLoadBalancerUDPListenerAttribute(reqCtx.Ctx, action.lbId, update)
}

type http struct {
	mgr *ListenerManager
}

func (t *http) Create(reqCtx *RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerHTTPListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create http listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener http [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)

}

func (t *http) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
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
	reqCtx.Log.Info(fmt.Sprintf("update listener: http [%d] update [%+v]", update.ListenerPort, update))
	return t.mgr.cloud.SetLoadBalancerHTTPListenerAttribute(reqCtx.Ctx, action.lbId, update)

}

type https struct {
	mgr *ListenerManager
}

func (t *https) Create(reqCtx *RequestContext, action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.mgr.cloud.CreateLoadBalancerHTTPSListener(reqCtx.Ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create https listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	reqCtx.Log.Info(fmt.Sprintf("create listener https [%d]", action.listener.ListenerPort))
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)

}

func (t *https) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start https listener %d error: %s", action.local.ListenerPort, err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(reqCtx, action.local, action.remote)
	if !needUpdate {
		reqCtx.Log.Info(fmt.Sprintf("update listener: https [%d] did not change, skip", action.local.ListenerPort))
		return nil
	}
	reqCtx.Log.Info(fmt.Sprintf("update listener: https [%d] updated [%v]", update.ListenerPort, update))
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
		if ports[0] == strconv.Itoa(int(target)) {
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
	return 0, fmt.Errorf("forward port format error: %s, expect 80:443,88:6443", port)
}

func buildActionsForListeners(reqCtx *RequestContext, local *model.LoadBalancer, remote *model.LoadBalancer) ([]CreateAction, []UpdateAction, []DeleteAction, error) {
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

	// For updations and deletions
	for _, rlis := range remote.Listeners {
		found := false
		for _, llis := range local.Listeners {
			if rlis.ListenerPort == llis.ListenerPort {
				found = true
				// protocol match, do update
				if rlis.Protocol == llis.Protocol {
					updateActions = append(updateActions,
						UpdateAction{
							lbId:   remote.LoadBalancerAttribute.LoadBalancerId,
							local:  llis,
							remote: rlis,
						})
				} else {
					// protocol not match, need to recreate
					deleteActions = append(deleteActions,
						DeleteAction{
							lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
							listener: rlis,
						})
					reqCtx.Log.Info(fmt.Sprintf("update listener: port [%d] match while protocol not, need recreate", llis.ListenerPort))
				}
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
			if llis.ListenerPort == rlis.ListenerPort {
				// port match
				if llis.Protocol != rlis.Protocol {
					// protocol does not match, do add listener
					break
				}
				// port matched. updated. skip
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
	if isENIBackendType(svc) {
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

	if Is7LayerProtocol(n.Protocol) {
		if n.HealthCheck == "" {
			n.HealthCheck = model.OffFlag
		}
		if n.StickySession == "" {
			n.StickySession = model.OffFlag
		}
	}
}

func isNeedUpdate(reqCtx *RequestContext, local model.ListenerAttribute, remote model.ListenerAttribute) (bool, model.ListenerAttribute) {
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
	// The cert id is necessary for https, so skip to check whether it is blank
	if local.Protocol == model.HTTPS &&
		remote.CertId != local.CertId {
		needUpdate = true
		update.CertId = local.CertId
		updateDetail += fmt.Sprintf("lb CertId %v should be changed to %v;",
			remote.CertId, local.CertId)
	}
	if local.Protocol == model.HTTPS &&
		local.EnableHttp2 != "" &&
		remote.EnableHttp2 != local.EnableHttp2 {
		needUpdate = true
		update.EnableHttp2 = local.EnableHttp2
		updateDetail += fmt.Sprintf("lb EnableHttp2 %v should be changed to %v;",
			remote.EnableHttp2, local.EnableHttp2)
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
	if Is7LayerProtocol(local.Protocol) &&
		local.IdleTimeout != 0 &&
		remote.IdleTimeout != local.IdleTimeout {
		needUpdate = true
		update.IdleTimeout = local.IdleTimeout
		updateDetail += fmt.Sprintf("lb IdleTimeout %v should be changed to %v;",
			remote.IdleTimeout, local.IdleTimeout)
	}
	// session
	if Is7LayerProtocol(local.Protocol) &&
		local.StickySession != "" &&
		remote.StickySession != local.StickySession {
		needUpdate = true
		update.StickySession = local.StickySession
		updateDetail += fmt.Sprintf("lb StickySession %v should be changed to %v;",
			remote.StickySession, local.StickySession)
	}
	if Is7LayerProtocol(local.Protocol) &&
		local.StickySessionType != "" &&
		remote.StickySessionType != local.StickySessionType {
		needUpdate = true
		update.StickySessionType = local.StickySessionType
		updateDetail += fmt.Sprintf("lb StickySessionType %v should be changed to %v;",
			remote.StickySessionType, local.StickySessionType)
	}
	if Is7LayerProtocol(local.Protocol) &&
		local.Cookie != "" &&
		remote.Cookie != local.Cookie {
		needUpdate = true
		update.Cookie = local.Cookie
		updateDetail += fmt.Sprintf("lb Cookie %v should be changed to %v;",
			remote.Cookie, local.Cookie)
	}
	if Is7LayerProtocol(local.Protocol) &&
		local.CookieTimeout != 0 &&
		remote.CookieTimeout != local.CookieTimeout {
		needUpdate = true
		update.CookieTimeout = local.CookieTimeout
		updateDetail += fmt.Sprintf("lb CookieTimeout %v should be changed to %v;",
			remote.CookieTimeout, local.CookieTimeout)
	}
	// connection drain
	if Is4LayerProtocol(local.Protocol) &&
		local.ConnectionDrain != "" &&
		remote.ConnectionDrain != local.ConnectionDrain {
		needUpdate = true
		update.ConnectionDrain = local.ConnectionDrain
		updateDetail += fmt.Sprintf("lb ConnectionDrain %v should be changed to %v;",
			remote.ConnectionDrain, local.ConnectionDrain)
	}
	if Is4LayerProtocol(local.Protocol) &&
		local.ConnectionDrain == model.OnFlag &&
		local.ConnectionDrainTimeout != 0 &&
		remote.ConnectionDrainTimeout != local.ConnectionDrainTimeout {
		needUpdate = true
		update.ConnectionDrainTimeout = local.ConnectionDrainTimeout
		updateDetail += fmt.Sprintf("lb ConnectionDrainTimeout %v should be changed to %v;",
			remote.ConnectionDrainTimeout, local.ConnectionDrainTimeout)
	}

	//x-forwarded-for
	if Is7LayerProtocol(local.Protocol) &&
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
	if Is4LayerProtocol(local.Protocol) &&
		local.HealthCheckConnectTimeout != 0 &&
		remote.HealthCheckConnectTimeout != local.HealthCheckConnectTimeout {
		needUpdate = true
		update.HealthCheckConnectTimeout = local.HealthCheckConnectTimeout
		updateDetail += fmt.Sprintf("lb HealthCheckConnectTimeout %v should be changed to %v;",
			remote.HealthCheckConnectTimeout, local.HealthCheckConnectTimeout)
	}
	if Is7LayerProtocol(local.Protocol) &&
		local.HealthCheck != "" &&
		remote.HealthCheck != local.HealthCheck {
		needUpdate = true
		update.HealthCheck = local.HealthCheck
		updateDetail += fmt.Sprintf("lb HealthCheck %v should be changed to %v;",
			remote.HealthCheck, local.HealthCheck)
	}
	if Is7LayerProtocol(local.Protocol) &&
		local.HealthCheckTimeout != 0 &&
		remote.HealthCheckTimeout != local.HealthCheckTimeout {
		needUpdate = true
		update.HealthCheckTimeout = local.HealthCheckTimeout
		updateDetail += fmt.Sprintf("lb HealthCheckTimeout %v should be changed to %v;",
			remote.HealthCheckTimeout, local.HealthCheckTimeout)
	}

	reqCtx.Ctx = context.WithValue(reqCtx.Ctx, dryrun.ContextMessage, updateDetail)
	reqCtx.Log.Info(fmt.Sprintf("try to update listener %d, detail %s", local.ListenerPort, updateDetail))
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
