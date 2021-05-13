package service

import (
	"context"
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
type IListener interface {
	Describe(ctx context.Context) error
	Add(ctx context.Context) error
	Remove(ctx context.Context) error
	Update(ctx context.Context) error
}

type Listener struct {
	cloud    prvd.Provider
	namedKey *model.ListenerNamedKey
	lbId     string
	port     int
	protocol string

	// Action indicate the operate method. ADD UPDATE DELETE
	action string

	local  *model.ListenerAttribute
	remote *model.ListenerAttribute
}

var (
	// ACTION_ADD actions add
	ACTION_ADD = "ADD"

	// ACTION_UPDATE update
	ACTION_UPDATE = "UPDATE"

	// ACTION_DELETE delete
	ACTION_DELETE = "DELETE"
)

// Instance  listener instance
func (n *Listener) Instance() IListener {
	switch strings.ToLower(n.protocol) {
	case model.TCP:
		return &tcp{n}
	case model.UDP:
		return &udp{n}
	case model.HTTP:
		return &http{n}
	case model.HTTPS:
		return &https{n}
	}
	return &tcp{n}
}

// Apply apply listener operate . add/update/delete etc.
func (n *Listener) Apply(ctx context.Context) error {
	klog.Infof("apply %s listener for %v with trans protocol %s", n.action, n.namedKey, n.protocol)
	switch n.action {
	case ACTION_UPDATE:
		return n.Instance().Update(ctx)
	case ACTION_ADD:
		err := n.Instance().Add(ctx)
		if err != nil {
			return err
		}
		return n.Start(ctx)
	case ACTION_DELETE:
		return n.Instance().Remove(ctx)
	}
	return fmt.Errorf("UnKnownAction: %s, %s", n.action, n.namedKey)
}

// Start start listener
func (n *Listener) Start(ctx context.Context) error {
	return n.cloud.StartLoadBalancerListener(ctx, n.lbId, n.port)
}

// Describe describe listener
func (n *Listener) Describe(ctx context.Context) error {
	return fmt.Errorf("unimplemented")
}

// Remove remove Listener
func (n *Listener) Remove(ctx context.Context) error {
	err := n.cloud.StopLoadBalancerListener(ctx, n.lbId, n.port)
	if err != nil {
		return err
	}
	return n.cloud.DeleteLoadBalancerListener(ctx, n.lbId, n.port)
}

type tcp struct{ *Listener }

func (t *tcp) Describe(ctx context.Context) error {
	return fmt.Errorf("unimplemented")
}

func (t *tcp) Add(ctx context.Context) error {
	setDefaultValueForListener(t.local)
	return t.cloud.CreateLoadBalancerTCPListener(ctx, t.lbId, t.local)
}

func (t *tcp) Update(ctx context.Context) error {
	if t.remote.Status == string(model.Stopped) {
		if err := t.cloud.StartLoadBalancerListener(ctx, t.lbId, t.port); err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	if !isNeedUpdate(t.local, t.remote) {
		klog.Infof("tcp listener did not change, skip [update], port=[%d]", t.port)
		// no recreate needed.  skip
		return nil
	}
	//return t.cloud.SetLoadBalancerTCPListenerAttribute(ctx, t.lbId, t.remote)
	return nil
}

type udp struct{ *Listener }

func (t *udp) Describe(ctx context.Context) error {
	return fmt.Errorf("unimplemented")
}

func (t *udp) Add(ctx context.Context) error {
	setDefaultValueForListener(t.local)
	return t.cloud.CreateLoadBalancerUDPListener(ctx, t.lbId, t.local)
}

func (t *udp) Update(ctx context.Context) error {
	if t.remote.Status == string(model.Stopped) {
		if err := t.cloud.StartLoadBalancerListener(ctx, t.lbId, t.port); err != nil {
			return fmt.Errorf("start udp listener error: %s", err.Error())
		}
	}
	if !isNeedUpdate(t.local, t.remote) {
		klog.Infof("udp listener did not change, skip [update], port=[%d]", t.port)
		// no recreate needed.  skip
		return nil
	}
	return t.cloud.SetLoadBalancerUDPListenerAttribute(ctx, t.lbId, t.remote)
}

type http struct{ *Listener }

func (t *http) Describe(ctx context.Context) error {
	return fmt.Errorf("unimplemented")
}

func (t *http) Add(ctx context.Context) error {
	setDefaultValueForListener(t.local)
	return t.cloud.CreateLoadBalancerHTTPListener(ctx, t.lbId, t.local)
}

func (t *http) Update(ctx context.Context) error {

	if t.remote.Status == string(model.Stopped) {
		if err := t.cloud.StartLoadBalancerListener(ctx, t.lbId, t.port); err != nil {
			return fmt.Errorf("start http listener error: %s", err.Error())
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

	if t.remote.ListenerForward == string(model.OnFlag) {
		klog.Infof("http %d ListenerForward is on, cannot update listener", t.port)
		// no update needed.  skip
		return nil
	}

	if !isNeedUpdate(t.local, t.remote) {
		klog.Infof("http listener did not change, skip [update], port=[%d]", t.port)
		// no recreate needed.  skip
		return nil
	}

	return t.cloud.SetLoadBalancerHTTPListenerAttribute(ctx, t.lbId, t.remote)

}

type https struct{ *Listener }

func (t *https) Describe(ctx context.Context) error {
	return fmt.Errorf("unimplemented")
}

func (t *https) Add(ctx context.Context) error {
	setDefaultValueForListener(t.local)
	return t.cloud.CreateLoadBalancerHTTPSListener(ctx, t.lbId, t.local)
}

func (t *https) Update(ctx context.Context) error {
	if t.remote.Status == string(model.Stopped) {
		if err := t.cloud.StartLoadBalancerListener(ctx, t.lbId, t.port); err != nil {
			return fmt.Errorf("start https listener error: %s", err.Error())
		}
	}
	if !isNeedUpdate(t.local, t.remote) {
		klog.Infof("https listener did not change, skip [update], port=[%d]", t.port)
		// no recreate needed.  skip
		return nil
	}
	return t.cloud.SetLoadBalancerHTTPSListenerAttribute(ctx, t.lbId, t.remote)
}

// ==========================================================================================
func buildListenersFromService(c localModel, port v1.ServicePort) (model.ListenerAttribute, error) {
	listener := model.ListenerAttribute{
		NamedKey: &model.ListenerNamedKey{
			Prefix:      model.DEFAULT_PREFIX,
			CID:         metadata.CLUSTER_ID,
			Namespace:   c.svc.Namespace,
			ServiceName: c.svc.Name,
			Port:        port.Port,
		},
		ListenerPort: int(port.Port),
	}

	listener.Description = listener.NamedKey.Key()
	listener.VGroupName = getVGroupNamedKey(c.svc, port).Key()

	proto, err := protocol(c.anno.Get(ProtocolPort), port)
	if err != nil {
		return listener, err
	}
	listener.Protocol = proto

	if c.anno.Get(Scheduler) != nil {
		listener.Scheduler = c.anno.Get(Scheduler)
	}

	if c.anno.Get(PersistenceTimeout) != nil {
		timeout, err := strconv.Atoi(*c.anno.Get(PersistenceTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation persistence timeout must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(PersistenceTimeout), err.Error())
		}
		listener.PersistenceTimeout = &timeout
	}

	if c.anno.Get(CertID) != nil {
		listener.CertId = c.anno.Get(CertID)
	}

	if c.anno.Get(ForwardPort) != nil && listener.Protocol == model.HTTP {
		forward := forwardPort(*c.anno.Get(ForwardPort), int(port.Port))
		listener.ForwardPort = &forward

	}

	// acl
	if c.anno.Get(AclStatus) != nil {
		listener.AclStatus = c.anno.Get(AclStatus)
	}
	if c.anno.Get(AclType) != nil {
		listener.AclType = c.anno.Get(AclType)
	}
	if c.anno.Get(AclID) != nil {
		listener.AclId = c.anno.Get(AclID)
	}

	// connection drain
	if c.anno.Get(ConnectionDrain) != nil {
		listener.ConnectionDrain = c.anno.Get(ConnectionDrain)
	}
	if c.anno.Get(ConnectionDrainTimeout) != nil {
		timeout, err := strconv.Atoi(*c.anno.Get(ConnectionDrainTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation ConnectionDrainTimeout must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(ConnectionDrainTimeout), err.Error())
		}
		listener.ConnectionDrainTimeout = &timeout
	}

	// cookie
	if c.anno.Get(Cookie) != nil {
		listener.Cookie = c.anno.Get(Cookie)
	}
	if c.anno.Get(CookieTimeout) != nil {
		timeout, err := strconv.Atoi(*c.anno.Get(CookieTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation CookieTimeout must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(CookieTimeout), err.Error())
		}
		listener.CookieTimeout = &timeout
	}
	if c.anno.Get(SessionStick) != nil {
		listener.StickySession = c.anno.Get(SessionStick)
	}
	if c.anno.Get(SessionStickType) != nil {
		listener.StickySessionType = c.anno.Get(SessionStickType)
	}

	// health check
	if c.anno.Get(HealthyThreshold) != nil {
		t, err := strconv.Atoi(*c.anno.Get(HealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthyThreshold must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(HealthyThreshold), err.Error())
		}
		listener.HealthyThreshold = &t
	}
	if c.anno.Get(UnhealthyThreshold) != nil {
		t, err := strconv.Atoi(*c.anno.Get(UnhealthyThreshold))
		if err != nil {
			return listener, fmt.Errorf("annotation UnhealthyThreshold must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(UnhealthyThreshold), err.Error())
		}
		listener.UnhealthyThreshold = &t
	}
	if c.anno.Get(HealthCheckConnectTimeout) != nil {
		timeout, err := strconv.Atoi(*c.anno.Get(HealthCheckConnectTimeout))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthCheckConnectTimeout must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(HealthCheckConnectTimeout), err.Error())
		}
		listener.HealthCheckConnectTimeout = &timeout
	}
	if c.anno.Get(HealthCheckConnectPort) != nil {
		port, err := strconv.Atoi(*c.anno.Get(HealthCheckConnectPort))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthCheckConnectPort must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(HealthCheckConnectPort), err.Error())
		}
		listener.HealthCheckConnectPort = &port
	}
	if c.anno.Get(HealthCheckInterval) != nil {
		t, err := strconv.Atoi(*c.anno.Get(HealthCheckInterval))
		if err != nil {
			return listener, fmt.Errorf("annotation HealthCheckInterval must be integer, but got [%s]. message=[%s]\n",
				*c.anno.Get(HealthCheckInterval), err.Error())
		}
		listener.HealthCheckInterval = &t
	}
	if c.anno.Get(HealthCheckDomain) != nil {
		listener.HealthCheckDomain = c.anno.Get(HealthCheckDomain)
	}
	if c.anno.Get(HealthCheckURI) != nil {
		listener.HealthCheckURI = c.anno.Get(HealthCheckURI)
	}
	if c.anno.Get(HealthCheckHTTPCode) != nil {
		listener.HealthCheckHttpCode = c.anno.Get(HealthCheckHTTPCode)
	}
	if c.anno.Get(HealthCheckType) != nil {
		listener.HealthCheckType = c.anno.Get(HealthCheckType)
	}

	return listener, nil
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

type Listeners []*Listener

func buildActionsForListeners(m *modelApplier, local *model.LoadBalancer, remote *model.LoadBalancer) (Listeners, error) {
	klog.Infof("try to update listener. local: [%v], remote: [%v]", util.PrettyJson(local.Listeners), util.PrettyJson(remote.Listeners))
	// make https come first.
	// ensure https listeners to be created first for http forward
	//sort.SliceStable(
	//	updates,
	//	func(i, j int) bool {
	//		// 1. https comes first.
	//		// 2. DELETE action comes before https
	//		if isDeleteAction(updates[i].Action) {
	//			return true
	//		}
	//		if isDeleteAction(updates[j].Action) {
	//			return false
	//		}
	//		if strings.ToUpper(
	//			updates[i].TransforedProto,
	//		) == "HTTPS" {
	//			return true
	//		}
	//		return false
	//	},
	//)
	var (
		addition Listeners
		updation Listeners
		deletion Listeners
	)
	for i := range local.Listeners {
		if err := findVServerGroup(remote.VServerGroups, &local.Listeners[i]); err != nil {
			return nil, fmt.Errorf("find vservergroup error: %s", err.Error())
		}
	}

	// For updations and deletions
	for i, rlis := range remote.Listeners {
		found := false
		for j, llis := range local.Listeners {
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
					updation = append(updation,
						&Listener{
							cloud:    m.cloud,
							namedKey: llis.NamedKey,
							lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
							port:     llis.ListenerPort,
							protocol: llis.Protocol,
							local:    &local.Listeners[j],
							remote:   &remote.Listeners[i],
							action:   ACTION_UPDATE,
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
			deletion = append(deletion, &Listener{
				cloud:    m.cloud,
				namedKey: rlis.NamedKey,
				lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
				port:     rlis.ListenerPort,
				protocol: rlis.Protocol,
				remote:   &remote.Listeners[i],
				action:   ACTION_DELETE,
			})
			klog.Infof("not found listener [%d], do delete", rlis.ListenerPort)
		}
	}

	// For additions
	for i, llis := range local.Listeners {
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
			addition = append(addition, &Listener{
				cloud:    m.cloud,
				namedKey: llis.NamedKey,
				lbId:     remote.LoadBalancerAttribute.LoadBalancerId,
				port:     llis.ListenerPort,
				protocol: llis.Protocol,
				local:    &local.Listeners[i],
				action:   ACTION_ADD,
			})
		}
	}
	// Pls be careful of the sequence. deletion first,then addition, last updation
	return append(append(deletion, addition...), updation...), nil
}

// Protocol for protocol transform
func protocol(annotation *string, port v1.ServicePort) (string, error) {

	if annotation == nil {
		klog.Infof("transfor protocol, empty annotation %d/%s", port.Port, port.Protocol)
		return strings.ToLower(string(port.Protocol)), nil
	}
	for _, v := range strings.Split(*annotation, ",") {
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
	if n.Bandwidth == nil {
		defaultBandwidth := DEFAULT_LISTENER_BANDWIDTH
		n.Bandwidth = &defaultBandwidth
	}
}

func isNeedUpdate(local *model.ListenerAttribute, remote *model.ListenerAttribute) bool {
	needUpdate := false

	if remote.VGroupId != local.VGroupId {
		needUpdate = true
		remote.VGroupId = local.VGroupId
	}
	// acl
	if local.AclStatus != nil &&
		remote.AclStatus != local.AclStatus {
		needUpdate = true
		remote.AclStatus = local.AclStatus
	}
	if local.AclId != nil &&
		remote.AclId != local.AclId {
		needUpdate = true
		remote.AclId = local.AclId
	}
	if local.AclType != nil &&
		remote.AclType != local.AclType {
		needUpdate = true
		remote.AclType = local.AclType
	}
	if local.Scheduler != nil &&
		remote.Scheduler != local.Scheduler {
		needUpdate = true
		remote.Scheduler = local.Scheduler
	}
	// health check
	if local.HealthCheckType != nil &&
		remote.HealthCheckType != local.HealthCheckType {
		needUpdate = true
		remote.HealthCheckType = local.HealthCheckType
	}

	if local.HealthCheckURI != nil &&
		remote.HealthCheckURI != local.HealthCheckURI {
		needUpdate = true
		remote.HealthCheckURI = local.HealthCheckURI
	}
	if local.HealthCheckConnectPort != nil &&
		remote.HealthCheckConnectPort != local.HealthCheckConnectPort {
		needUpdate = true
		remote.HealthCheckConnectPort = local.HealthCheckConnectPort
	}
	if local.HealthyThreshold != nil &&
		remote.HealthyThreshold != local.HealthyThreshold {
		needUpdate = true
		remote.HealthyThreshold = local.HealthyThreshold
	}
	if local.UnhealthyThreshold != nil &&
		remote.UnhealthyThreshold != local.UnhealthyThreshold {
		needUpdate = true
		remote.UnhealthyThreshold = local.UnhealthyThreshold
	}
	if local.HealthCheckConnectTimeout != nil &&
		remote.HealthCheckConnectTimeout != local.HealthCheckConnectTimeout {
		needUpdate = true
		remote.HealthCheckConnectTimeout = local.HealthCheckConnectTimeout
	}
	if local.HealthCheckInterval != nil &&
		remote.HealthCheckInterval != local.HealthCheckInterval {
		needUpdate = true
		remote.HealthCheckInterval = local.HealthCheckInterval
	}
	if local.PersistenceTimeout != nil &&
		remote.PersistenceTimeout != local.PersistenceTimeout {
		needUpdate = true
		remote.PersistenceTimeout = local.PersistenceTimeout
	}
	if local.HealthCheckHttpCode != nil &&
		remote.HealthCheckHttpCode != local.HealthCheckHttpCode {
		needUpdate = true
		remote.HealthCheckHttpCode = local.HealthCheckHttpCode
	}
	if local.HealthCheckDomain != nil &&
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
