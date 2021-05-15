package service

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/metadata"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"strconv"
	"strings"
)

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

// IListener listener interface
type IListener interface {
	Create(action CreateAction) error
	Update(action UpdateAction) error
}

type tcp struct {
	reqCtx *RequestContext
}

func (t *tcp) Create(action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.reqCtx.cloud.CreateLoadBalancerTCPListener(t.reqCtx.ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create tcp listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	return t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.listener.ListenerPort)
}

func (t *tcp) Update(action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	if !isNeedUpdate(action.local, &action.remote) {
		klog.Infof("tcp listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}
	// TODO fix me
	//return t.reqCtx.cloud.SetLoadBalancerTCPListenerAttribute(t.reqCtx.ctx, action.lbId, action.remote)
	return nil
}

type udp struct {
	reqCtx *RequestContext
}

func (t *udp) Create(action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.reqCtx.cloud.CreateLoadBalancerUDPListener(t.reqCtx.ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create udp listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	return t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.listener.ListenerPort)
}

func (t *udp) Update(action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	if !isNeedUpdate(action.local, &action.remote) {
		klog.Infof("udp listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}
	return t.reqCtx.cloud.SetLoadBalancerUDPListenerAttribute(t.reqCtx.ctx, action.lbId, action.remote)
}

type http struct {
	reqCtx *RequestContext
}

func (t *http) Create(action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.reqCtx.cloud.CreateLoadBalancerHTTPListener(t.reqCtx.ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create http listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	return t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.listener.ListenerPort)

}

func (t *http) Update(action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	// TODO forward
	//forward := forwardPort(def.ForwardPort, t.Port)
	//if forward != 0 {
	//	if response.ListenerForward != slb.OnFlag {
	//		needRecreate = true
	//		config.ListenerForward = slb.OnFlag
	//	}
	//} else {
	//	if response.ListenerForward != slb.OffFlag {
	//		needRecreate = true
	//		config.ListenerForward = slb.OffFlag
	//	}
	//}
	//config.ForwardPort = int(forward)
	//if !isNeedUpdate(t.local, t.remote) {
	//	klog.Infof("tcp listener did not change, skip [update], port=[%d]", t.port)
	//	// no recreate needed.  skip
	//	return nil
	//}
	//
	//if needRecreate {
	//
	//	config.BackendServerPort = int(t.NodePort)
	//	utils.Logf(t.Service, "HTTP listener checker [BackendServerPort]"+
	//		" changed, request=%d. response=%d. Recreate http listener.", t.NodePort, response.BackendServerPort)
	//	// The listener description has changed. It may be that multiple services reuse the same port of the same slb, and needs to record event.
	//	if response.Description != config.Description {
	//		record, err := utils.GetRecorderFromContext(ctx)
	//		if err != nil {
	//			klog.Warningf("get recorder error: %s", err.Error())
	//		} else {
	//			record.Eventf(
	//				t.Service,
	//				v1.EventTypeNormal,
	//				"RecreateListener",
	//				"Recreate HTTP listener [%s] -> [%s]",
	//				response.Description, config.Description,
	//			)
	//		}
	//	}
	//	err := t.Client.DeleteLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
	//	if err != nil {
	//		return err
	//	}
	//	err = t.Client.CreateLoadBalancerHTTPListener(ctx, (*slb.CreateLoadBalancerHTTPListenerArgs)(config))
	//	if err != nil {
	//		return err
	//	}
	//	return t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
	//}

	if action.remote.ListenerForward == model.OnFlag {
		klog.Infof("http %d ListenerForward is on, cannot update listener", action.local.ListenerPort)
		// no update needed.  skip
		return nil
	}

	if !isNeedUpdate(action.local, &action.remote) {
		klog.Infof("http listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}

	return t.reqCtx.cloud.SetLoadBalancerHTTPListenerAttribute(t.reqCtx.ctx, action.lbId, action.remote)

}

type https struct {
	reqCtx *RequestContext
}

func (t *https) Create(action CreateAction) error {
	setDefaultValueForListener(&action.listener)
	err := t.reqCtx.cloud.CreateLoadBalancerHTTPSListener(t.reqCtx.ctx, action.lbId, action.listener)
	if err != nil {
		return fmt.Errorf("create https listener %d error: %s", action.listener.ListenerPort, err.Error())
	}
	return t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.listener.ListenerPort)

}

func (t *https) Update(action UpdateAction) error {
	if action.remote.Status == model.Stopped {
		err := t.reqCtx.cloud.StartLoadBalancerListener(t.reqCtx.ctx, action.lbId, action.local.ListenerPort)
		if err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	if !isNeedUpdate(action.local, &action.remote) {
		klog.Infof("https listener did not change, skip [update], port=[%d]", action.local.ListenerPort)
		// no recreate needed.  skip
		return nil
	}
	return t.reqCtx.cloud.SetLoadBalancerHTTPSListenerAttribute(t.reqCtx.ctx, action.lbId, action.remote)
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
	klog.Infof("try to update listener. local: [%v], remote: [%v]", util.PrettyJson(local.Listeners), util.PrettyJson(remote.Listeners))
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

	// For updations and deletions
	for _, rlis := range remote.Listeners {
		found := false
		for _, llis := range local.Listeners {
			if rlis.ListenerPort == llis.ListenerPort {
				found = true
				// port matched. that is where the conflict case begin.
				//1. check protocol match.
				if rlis.Protocol == llis.Protocol {
					// protocol match, need to do update operate no matter managed by whom
					// consider override annotation & user defined loadbalancer
					//if !override && isUserDefinedLoadBalancer(svc) {
					//	// port conflict with user managed slb or listener.
					//	return nil, fmt.Errorf("PortProtocolConflict] port matched, but conflict with user managed listener. "+
					//		"Port:%d, ListenerName:%s, svc: %s. Protocol:[source:%s dst:%s]",
					//		remote.Port, remote.Name, local.NamedKey.Key(), remote.TransforedProto, local.TransforedProto)
					//}

					// do update operate
					updateActions = append(updateActions,
						UpdateAction{
							lbId:   remote.LoadBalancerAttribute.LoadBalancerId,
							local:  llis,
							remote: rlis,
						})
					klog.Infof("found listener %d with port & protocol match, do update", llis.ListenerPort)
				}
			} else {
				//// protocol not match, need to recreate
				//if !override && isUserDefinedLoadBalancer(svc) {
				//	return nil, fmt.Errorf("[PortProtocolConflict] port matched, "+
				//		"while protocol does not. force override listener %t. source[%v], target[%v]", override, local.NamedKey, remote.NamedKey)
				//}
				klog.Infof("found listener with protocol match, need recreate")
			}
		}
		// Do not delete any listener that no longer managed by my service
		// for safety. Only conflict case is taking care.
		if !found {
			//if isManagedByMyService(svc, remote) {
			//	// port has user managed nodes, do not delete
			//	hasUserNode, err := remote.listenerHasUserManagedNode(ctx)
			//	if err != nil {
			//		return nil, fmt.Errorf("check if listener has user managed node, error: %s", err.Error())
			//	}
			//	if hasUserNode {
			//		utils.Logf(svc, "svc %s do not delete port %d, because backends have user managed nodes", remote.NamedKey.Key(), remote.Port)
			//		continue
			//	}
			//	remote.Action = ACTION_DELETE
			//	deletion = append(deletion, remote)
			//	utils.Logf(svc, "found listener[%s] which is no longer needed "+
			//		"by my service[%s/%s], do delete", remote.NamedKey.Key(), svc.Namespace, svc.Name)
			//} else {
			//	utils.Logf(svc, "port [%d] not managed by my service [%s/%s], skip processing.", remote.Port, svc.Namespace, svc.Name)
			//}
			deleteActions = append(deleteActions, DeleteAction{
				lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
				listener: rlis,
			})
			klog.Infof("not found listener [%d], do delete", rlis.ListenerPort)
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
	if n.Bandwidth == 0 {
		n.Bandwidth = DEFAULT_LISTENER_BANDWIDTH
	}
}

// todo return update content, do not use remote directly
func isNeedUpdate(local model.ListenerAttribute, remote *model.ListenerAttribute) bool {
	needUpdate := false

	if remote.VGroupId != local.VGroupId {
		needUpdate = true
		remote.VGroupId = local.VGroupId
	}
	// acl
	if local.AclStatus != "" &&
		remote.AclStatus != local.AclStatus {
		needUpdate = true
		remote.AclStatus = local.AclStatus
	}
	if local.AclId != "" &&
		remote.AclId != local.AclId {
		needUpdate = true
		remote.AclId = local.AclId
	}
	if local.AclType != "" &&
		remote.AclType != local.AclType {
		needUpdate = true
		remote.AclType = local.AclType
	}
	if local.Scheduler != "" &&
		remote.Scheduler != local.Scheduler {
		needUpdate = true
		remote.Scheduler = local.Scheduler
	}
	// health check
	if local.HealthCheckType != "" &&
		remote.HealthCheckType != local.HealthCheckType {
		needUpdate = true
		remote.HealthCheckType = local.HealthCheckType
	}

	if local.HealthCheckURI != "" &&
		remote.HealthCheckURI != local.HealthCheckURI {
		needUpdate = true
		remote.HealthCheckURI = local.HealthCheckURI
	}
	if local.HealthCheckConnectPort != 0 &&
		remote.HealthCheckConnectPort != local.HealthCheckConnectPort {
		needUpdate = true
		remote.HealthCheckConnectPort = local.HealthCheckConnectPort
	}
	if local.HealthyThreshold != 0 &&
		remote.HealthyThreshold != local.HealthyThreshold {
		needUpdate = true
		remote.HealthyThreshold = local.HealthyThreshold
	}
	if local.UnhealthyThreshold != 0 &&
		remote.UnhealthyThreshold != local.UnhealthyThreshold {
		needUpdate = true
		remote.UnhealthyThreshold = local.UnhealthyThreshold
	}
	if local.HealthCheckConnectTimeout != 0 &&
		remote.HealthCheckConnectTimeout != local.HealthCheckConnectTimeout {
		needUpdate = true
		remote.HealthCheckConnectTimeout = local.HealthCheckConnectTimeout
	}
	if local.HealthCheckInterval != 0 &&
		remote.HealthCheckInterval != local.HealthCheckInterval {
		needUpdate = true
		remote.HealthCheckInterval = local.HealthCheckInterval
	}
	if local.PersistenceTimeout != nil &&
		remote.PersistenceTimeout != local.PersistenceTimeout {
		needUpdate = true
		remote.PersistenceTimeout = local.PersistenceTimeout
	}
	if local.HealthCheckHttpCode != "" &&
		remote.HealthCheckHttpCode != local.HealthCheckHttpCode {
		needUpdate = true
		remote.HealthCheckHttpCode = local.HealthCheckHttpCode
	}
	if local.HealthCheckDomain != "" &&
		remote.HealthCheckDomain != local.HealthCheckDomain {
		needUpdate = true
		remote.HealthCheckDomain = local.HealthCheckDomain
	}
	return needUpdate

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

func (reqCtx *RequestContext) BuildListenersFromService(mdl *model.LoadBalancer) error {
	for _, port := range reqCtx.svc.Spec.Ports {
		listener, err := buildListenerFromServicePort(reqCtx, port)
		if err != nil {
			return fmt.Errorf("build listener from servicePort %d error: %s", port.Port, err.Error())
		}
		mdl.Listeners = append(mdl.Listeners, listener)
	}
	return nil
}

func (reqCtx *RequestContext) BuildListenersFromCloud(mdl *model.LoadBalancer) error {
	listeners, err := reqCtx.cloud.DescribeLoadBalancerListeners(reqCtx.ctx, mdl.LoadBalancerAttribute.LoadBalancerId)
	if err != nil {
		return fmt.Errorf("DescribeLoadBalancerListeners error:%s", err.Error())
	}
	mdl.Listeners = listeners
	return nil
}

func (reqCtx *RequestContext) EnsureListenerCreated(action CreateAction) error {
	switch strings.ToLower(action.listener.Protocol) {
	case model.TCP:
		return (&tcp{reqCtx}).Create(action)
	case model.UDP:
		return (&udp{reqCtx}).Create(action)
	case model.HTTP:
		return (&http{reqCtx}).Create(action)
	case model.HTTPS:
		return (&https{reqCtx}).Create(action)
	default:
		return fmt.Errorf("%s protocol is not supported", action.listener.Protocol)
	}

}

func (reqCtx *RequestContext) EnsureListenerDeleted(action DeleteAction) error {
	return reqCtx.cloud.DeleteLoadBalancerListener(reqCtx.ctx, action.lbId, action.listener.ListenerPort)
}

func (reqCtx *RequestContext) EnsureListenerUpdated(action UpdateAction) error {
	switch strings.ToLower(action.local.Protocol) {
	case model.TCP:
		return (&tcp{reqCtx}).Update(action)
	case model.UDP:
		return (&udp{reqCtx}).Update(action)
	case model.HTTP:
		return (&http{reqCtx}).Update(action)
	case model.HTTPS:
		return (&https{reqCtx}).Update(action)
	default:
		return fmt.Errorf("%s protocol is not supported", action.local.Protocol)
	}
}

func buildListenerFromServicePort(req *RequestContext, port v1.ServicePort) (model.ListenerAttribute, error) {
	listener := model.ListenerAttribute{
		NamedKey: &model.ListenerNamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			CID:         metadata.CLUSTER_ID,
			Namespace:   req.svc.Namespace,
			ServiceName: req.svc.Name,
			Port:        port.Port,
		},
		ListenerPort: int(port.Port),
	}

	listener.Description = listener.NamedKey.Key()
	listener.VGroupName = getVGroupNamedKey(req.svc, port).Key()

	proto, err := protocol(req.anno.Get(ProtocolPort), port)
	if err != nil {
		return listener, err
	}
	listener.Protocol = proto

	if req.anno.Get(Scheduler) != "" {
		listener.Scheduler = req.anno.Get(Scheduler)
	}

	if req.anno.Get(PersistenceTimeout) != "" {
		timeout, err := strconv.Atoi(req.anno.Get(PersistenceTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(PersistenceTimeout), err.Error())
		}
		listener.PersistenceTimeout = &timeout
	}

	if req.anno.Get(CertID) != "" {
		listener.CertId = req.anno.Get(CertID)
	}

	if req.anno.Get(ForwardPort) != "" && listener.Protocol == model.HTTP {
		listener.ForwardPort = forwardPort(req.anno.Get(ForwardPort), int(port.Port))

	}

	// acl
	if req.anno.Get(AclStatus) != "" {
		listener.AclStatus = model.FlagType(req.anno.Get(AclStatus))
	}
	if req.anno.Get(AclType) != "" {
		listener.AclType = req.anno.Get(AclType)
	}
	if req.anno.Get(AclID) != "" {
		listener.AclId = req.anno.Get(AclID)
	}

	// connection drain
	if req.anno.Get(ConnectionDrain) != "" {
		listener.ConnectionDrain = model.FlagType(req.anno.Get(ConnectionDrain))
	}
	if req.anno.Get(ConnectionDrainTimeout) != "" {
		timeout, err := strconv.Atoi(req.anno.Get(ConnectionDrainTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation ConnectionDrainTimeout must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(ConnectionDrainTimeout), err.Error())
		}
		listener.ConnectionDrainTimeout = timeout
	}

	// cookie
	if req.anno.Get(Cookie) != "" {
		listener.Cookie = req.anno.Get(Cookie)
	}
	if req.anno.Get(CookieTimeout) != "" {
		timeout, err := strconv.Atoi(req.anno.Get(CookieTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation CookieTimeout must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(CookieTimeout), err.Error())
		}
		listener.CookieTimeout = timeout
	}
	if req.anno.Get(SessionStick) != "" {
		listener.StickySession = model.FlagType(req.anno.Get(SessionStick))
	}
	if req.anno.Get(SessionStickType) != "" {
		listener.StickySessionType = req.anno.Get(SessionStickType)
	}

	// health check
	if req.anno.Get(HealthyThreshold) != "" {
		t, err := strconv.Atoi(req.anno.Get(HealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthyThreshold must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(HealthyThreshold), err.Error())
		}
		listener.HealthyThreshold = t
	}
	if req.anno.Get(UnhealthyThreshold) != "" {
		t, err := strconv.Atoi(req.anno.Get(UnhealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("annotation UnhealthyThreshold must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(UnhealthyThreshold), err.Error())
		}
		listener.UnhealthyThreshold = t
	}
	if req.anno.Get(HealthCheckConnectTimeout) != "" {
		timeout, err := strconv.Atoi(req.anno.Get(HealthCheckConnectTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthCheckConnectTimeout must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(HealthCheckConnectTimeout), err.Error())
		}
		listener.HealthCheckConnectTimeout = timeout
	}
	if req.anno.Get(HealthCheckConnectPort) != "" {
		port, err := strconv.Atoi(req.anno.Get(HealthCheckConnectPort))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthCheckConnectPort must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(HealthCheckConnectPort), err.Error())
		}
		listener.HealthCheckConnectPort = port
	}
	if req.anno.Get(HealthCheckInterval) != "" {
		t, err := strconv.Atoi(req.anno.Get(HealthCheckInterval))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthCheckInterval must be integer, but got [%s]. message=[%s]\n",
				req.anno.Get(HealthCheckInterval), err.Error())
		}
		listener.HealthCheckInterval = t
	}
	if req.anno.Get(HealthCheckDomain) != "" {
		listener.HealthCheckDomain = req.anno.Get(HealthCheckDomain)
	}
	if req.anno.Get(HealthCheckURI) != "" {
		listener.HealthCheckURI = req.anno.Get(HealthCheckURI)
	}
	if req.anno.Get(HealthCheckHTTPCode) != "" {
		listener.HealthCheckHttpCode = req.anno.Get(HealthCheckHTTPCode)
	}
	if req.anno.Get(HealthCheckType) != "" {
		listener.HealthCheckType = req.anno.Get(HealthCheckType)
	}

	return listener, nil
}
