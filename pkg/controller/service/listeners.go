package service

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"strconv"
	"strings"
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

// private func
func (mgr *ListenerManager) buildListenerFromServicePort(reqCtx *RequestContext, port v1.ServicePort) (model.ListenerAttribute, error) {
	listener := model.ListenerAttribute{
		NamedKey: &model.ListenerNamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			CID:         metadata.CLUSTER_ID,
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

	if reqCtx.Anno.Get(Scheduler) != "" {
		listener.Scheduler = reqCtx.Anno.Get(Scheduler)
	}

	if reqCtx.Anno.Get(PersistenceTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(PersistenceTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
				reqCtx.Anno.Get(PersistenceTimeout), err.Error())
		}
		listener.PersistenceTimeout = &timeout
	}

	if reqCtx.Anno.Get(CertID) != "" {
		listener.CertId = reqCtx.Anno.Get(CertID)
	}

	if reqCtx.Anno.Get(ForwardPort) != "" && listener.Protocol == model.HTTP {
		listener.ForwardPort = forwardPort(reqCtx.Anno.Get(ForwardPort), int(port.Port))
		listener.ListenerForward = model.OnFlag
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
			return listener, fmt.Errorf("Annotation ConnectionDrainTimeout must be integer, but got [%s]. message=[%s]\n",
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
			return listener, fmt.Errorf("Annotation CookieTimeout must be integer, but got [%s]. message=[%s]\n",
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

	// health check
	if reqCtx.Anno.Get(HealthyThreshold) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(HealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthyThreshold must be integer, but got [%s]. message=[%s]\n",
				reqCtx.Anno.Get(HealthyThreshold), err.Error())
		}
		listener.HealthyThreshold = t
	}
	if reqCtx.Anno.Get(UnhealthyThreshold) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(UnhealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("Annotation UnhealthyThreshold must be integer, but got [%s]. message=[%s]\n",
				reqCtx.Anno.Get(UnhealthyThreshold), err.Error())
		}
		listener.UnhealthyThreshold = t
	}
	if reqCtx.Anno.Get(HealthCheckConnectTimeout) != "" {
		timeout, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckConnectTimeout))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckConnectTimeout must be integer, but got [%s]. message=[%s]\n",
				reqCtx.Anno.Get(HealthCheckConnectTimeout), err.Error())
		}
		listener.HealthCheckConnectTimeout = timeout
	}
	if reqCtx.Anno.Get(HealthCheckConnectPort) != "" {
		port, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckConnectPort))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckConnectPort must be integer, but got [%s]. message=[%s]\n",
				reqCtx.Anno.Get(HealthCheckConnectPort), err.Error())
		}
		listener.HealthCheckConnectPort = port
	}
	if reqCtx.Anno.Get(HealthCheckInterval) != "" {
		t, err := strconv.Atoi(reqCtx.Anno.Get(HealthCheckInterval))
		if err != nil {
			return listener, fmt.Errorf("Annotation HealthCheckInterval must be integer, but got [%s]. message=[%s]\n",
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
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)
}

func (t *tcp) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(action.local, action.remote)
	if !needUpdate {
		klog.Infof("tcp listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}
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
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)
}

func (t *udp) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(action.local, action.remote)
	if !needUpdate {
		klog.Infof("udp listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}
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
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)

}

func (t *http) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
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
		klog.Infof("http listener [%d] need recreate", action.local.ListenerPort)
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
		klog.Infof("http %d ListenerForward is on, cannot update listener", action.local.ListenerPort)
		// no update needed.  skip
		return nil
	}

	needUpdate, update := isNeedUpdate(action.local, action.remote)
	if !needUpdate {
		klog.Infof("http listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}

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
	return t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.listener.ListenerPort)

}

func (t *https) Update(reqCtx *RequestContext, action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.mgr.cloud.StartLoadBalancerListener(reqCtx.Ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	needUpdate, update := isNeedUpdate(action.local, action.remote)
	if !needUpdate {
		klog.Infof("https listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}
	return t.mgr.cloud.SetLoadBalancerHTTPSListenerAttribute(reqCtx.Ctx, action.lbId, update)
}

func forwardPort(port string, target int) int {
	if port == "" {
		return 0
	}
	forwarded := ""
	tmps := strings.Split(port, ",")
	for _, v := range tmps {
		ports := strings.Split(v, ":")
		if len(ports) != 2 {
			klog.Infof("forward-port format error: %s, expect 80:443,88:6443", port)
			continue
		}
		if ports[0] == strconv.Itoa(int(target)) {
			forwarded = ports[1]
			break
		}
	}
	if forwarded != "" {
		forward, err := strconv.Atoi(forwarded)
		if err != nil {
			klog.Errorf("forward port is not an integer, %s", forwarded)
			return 0
		}
		klog.Infof("forward http port %d to %d", target, forward)
		return forward
	}
	return 0
}

func buildActionsForListeners(local *model.LoadBalancer, remote *model.LoadBalancer) ([]CreateAction, []UpdateAction, []DeleteAction, error) {
	var (
		createActions []CreateAction
		updateActions []UpdateAction
		deleteActions []DeleteAction
	)
	for i := range local.Listeners {
		if err := findVServerGroup(remote.VServerGroups, &local.Listeners[i]); err != nil {
			return createActions, updateActions, deleteActions, fmt.Errorf("find vservergroup error: %s", err.Error())
		}
	}
	klog.Infof("try to update listener. local: [%v], remote: [%v]", util.PrettyJson(local.Listeners), util.PrettyJson(remote.Listeners))

	// For updations and deletions
	for _, rlis := range remote.Listeners {
		found := false
		for _, llis := range local.Listeners {
			if rlis.ListenerPort == llis.ListenerPort {
				found = true
				// port matched. that is where the conflict case begin.
				//1. check protocol match.
				if rlis.Protocol == llis.Protocol {
					// do update operate
					updateActions = append(updateActions,
						UpdateAction{
							lbId:   remote.LoadBalancerAttribute.LoadBalancerId,
							local:  llis,
							remote: rlis,
						})
					klog.Infof("found listener %d with port & protocol match, do update", llis.ListenerPort)
				} else {
					// protocol not match, need to recreate
					deleteActions = append(deleteActions,
						DeleteAction{
							lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
							listener: rlis,
						})
					klog.Infof("found listener with port match while protocol not, do delete & add %s", llis.NamedKey)
				}
			}
		}
		// Do not delete any listener that no longer managed by my service
		// for safety. Only conflict case is taking care.
		if !found {
			if local.LoadBalancerAttribute.IsUserManaged &&
				!isManagedByMyService(local, rlis) {
				klog.Infof("port [%d] not managed by my service [%s], skip processing.", rlis.ListenerPort, rlis.NamedKey)
				continue
			}

			//// port has user managed nodes, do not delete
			//hasUserNode, err := remote.listenerHasUserManagedNode(Ctx)
			//if err != nil {
			//	return nil, fmt.Errorf("check if listener has user managed node, error: %s", err.Error())
			//}
			//if hasUserNode {
			//	utils.Logf(svc, "svc %s do not delete port %d, because backends have user managed nodes", remote.NamedKey.Key(), remote.Port)
			//	continue
			//}
			deleteActions = append(deleteActions, DeleteAction{
				lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
				listener: rlis,
			})
			klog.Infof("found listener[%s] which is no longer needed "+
				"by my service[%s], do delete", rlis.ListenerPort, rlis.NamedKey)
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
		klog.Infof("transfor protocol, empty annotation %d/%s", port.Port, port.Protocol)
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
			klog.Infof("transform protocol from %s to %s", port.Protocol, pp[0])
			return pp[0], nil
		}
	}
	return strings.ToLower(string(port.Protocol)), nil
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
		CID:         metadata.CLUSTER_ID,
		VGroupPort:  vGroupPort,
		ServiceName: svc.Name}
}

func setDefaultValueForListener(n *model.ListenerAttribute) {
	if n.Protocol == model.TCP || n.Protocol == model.UDP || n.Protocol == model.HTTPS {
		if n.Bandwidth == 0 {
			n.Bandwidth = DEFAULT_LISTENER_BANDWIDTH
		}
	}

	if n.Protocol == model.HTTP || n.Protocol == model.HTTPS {
		if n.HealthCheck == "" {
			n.HealthCheck = model.OffFlag
		}
		if n.StickySession == "" {
			n.StickySession = model.OffFlag
		}
	}
}

func isNeedUpdate(local model.ListenerAttribute, remote model.ListenerAttribute) (bool, model.ListenerAttribute) {
	update := model.ListenerAttribute{
		ListenerPort: local.ListenerPort,
	}
	needUpdate := false

	if remote.Description != local.Description {
		needUpdate = true
		update.Description = local.Description
	}
	if remote.VGroupId != local.VGroupId {
		needUpdate = true
		update.VGroupId = local.VGroupId
	}
	// acl
	if local.AclStatus != "" &&
		remote.AclStatus != local.AclStatus {
		needUpdate = true
		update.AclStatus = local.AclStatus
	}
	if local.AclId != "" &&
		remote.AclId != local.AclId {
		needUpdate = true
		update.AclId = local.AclId
	}
	if local.AclType != "" &&
		remote.AclType != local.AclType {
		needUpdate = true
		update.AclType = local.AclType
	}
	if local.Scheduler != "" &&
		remote.Scheduler != local.Scheduler {
		needUpdate = true
		update.Scheduler = local.Scheduler
	}
	// health check
	if local.HealthCheckType != "" &&
		remote.HealthCheckType != local.HealthCheckType {
		needUpdate = true
		update.HealthCheckType = local.HealthCheckType
	}
	if local.HealthCheckURI != "" &&
		remote.HealthCheckURI != local.HealthCheckURI {
		needUpdate = true
		update.HealthCheckURI = local.HealthCheckURI
	}
	if local.HealthCheckConnectPort != 0 &&
		remote.HealthCheckConnectPort != local.HealthCheckConnectPort {
		needUpdate = true
		update.HealthCheckConnectPort = local.HealthCheckConnectPort
	}
	if local.HealthyThreshold != 0 &&
		remote.HealthyThreshold != local.HealthyThreshold {
		needUpdate = true
		update.HealthyThreshold = local.HealthyThreshold
	}
	if local.UnhealthyThreshold != 0 &&
		remote.UnhealthyThreshold != local.UnhealthyThreshold {
		needUpdate = true
		update.UnhealthyThreshold = local.UnhealthyThreshold
	}
	if local.HealthCheckConnectTimeout != 0 &&
		remote.HealthCheckConnectTimeout != local.HealthCheckConnectTimeout {
		needUpdate = true
		update.HealthCheckConnectTimeout = local.HealthCheckConnectTimeout
	}
	if local.HealthCheckInterval != 0 &&
		remote.HealthCheckInterval != local.HealthCheckInterval {
		needUpdate = true
		update.HealthCheckInterval = local.HealthCheckInterval
	}
	if local.PersistenceTimeout != nil &&
		remote.PersistenceTimeout != local.PersistenceTimeout {
		needUpdate = true
		update.PersistenceTimeout = local.PersistenceTimeout
	}
	if local.HealthCheckHttpCode != "" &&
		remote.HealthCheckHttpCode != local.HealthCheckHttpCode {
		needUpdate = true
		update.HealthCheckHttpCode = local.HealthCheckHttpCode
	}
	if local.HealthCheckDomain != "" &&
		remote.HealthCheckDomain != local.HealthCheckDomain {
		needUpdate = true
		update.HealthCheckDomain = local.HealthCheckDomain
	}
	if local.HealthCheckTimeout != 0 &&
		remote.HealthCheckTimeout != local.HealthCheckTimeout {
		needUpdate = true
		update.HealthCheckTimeout = local.HealthCheckTimeout
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

func isManagedByMyService(local *model.LoadBalancer, n model.ListenerAttribute) bool {
	return n.NamedKey.ServiceName == local.NamespacedName.Name &&
		n.NamedKey.Namespace == local.NamespacedName.Namespace &&
		n.NamedKey.CID == metadata.CLUSTER_ID
}
