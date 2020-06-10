/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alicloud

import (
	"context"
	//"errors"
	"fmt"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/klog"
	"sort"
	"strconv"
	"strings"
)

// DEFAULT_LISTENER_BANDWIDTH default listener bandwidth
var DEFAULT_LISTENER_BANDWIDTH = -1

/*
	Author: @aoxn
	Date:   2018-11-20

	ListenerV1 Path:
	 -----------------------------------------------------------------------------
    |																			  |
	|					Named Listener 80                                         |
	| LoadBalancer ->   Named Listener 443          -> Defaulted Server Group     |
	|					Named Listener 8080			                              |
	|																			  |
	 ----------------------------------------------------------------------------

	ListenerV2 Path:
	 -----------------------------------------------------------------------------
    |																			  |
	|					Named Listener 80           -> Named Vserver group 80     |
	| LoadBalancer ->   Named Listener 443          -> Named Vserver group 443    |
	|					Named Listener 8080			-> Named Vserver group 8080   |
	|																			  |
	 ----------------------------------------------------------------------------
	ListenerV2 Name Format:     k8s/Port/ServiceName/Namespace/ClusterID
	VServer Group Name Format:  k8s/NodePort/ServiceName/Namespace/ClusterID
*/

// Protocol for protocol transform
func Protocol(annotation string, port v1.ServicePort) (string, error) {

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

		if pp[0] != "http" &&
			pp[0] != "tcp" &&
			pp[0] != "https" &&
			pp[0] != "udp" {
			return "", fmt.Errorf("port protocol"+
				" format must be either [http|https|tcp|udp], protocol not supported wit [%s]\n", pp[0])
		}

		if pp[1] == fmt.Sprintf("%d", port.Port) {
			klog.Infof("transfor protocol from %s to %s", string(port.Protocol), pp[0])
			return pp[0], nil
		}
	}
	return strings.ToLower(string(port.Protocol)), nil
}

// IListener listener interface
type IListener interface {
	Describe(ctx context.Context) error
	Add(ctx context.Context) error
	Remove(ctx context.Context) error
	Update(ctx context.Context) error
}

// FORMAT_ERROR format error message
var FORMAT_ERROR = "ListenerName Format Error: k8s/${port}/${service}/${namespace}/${clusterid} format is expected"

type formatError struct{ key string }

func (f formatError) Error() string { return fmt.Sprintf("%s. Got [%s]", FORMAT_ERROR, f.key) }

// DEFAULT_PREFIX default prefix for listener
var DEFAULT_PREFIX = "k8s"

// NamedKey identify listeners on grouped attributes
type NamedKey struct {
	Prefix      string
	CID         string
	Namespace   string
	ServiceName string
	Port        int32
	TargetPort  int32
}

func (n *NamedKey) String() string {
	if n == nil {
		return ""
	}
	return n.Key()
}

// Key key of NamedKey
func (n *NamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%d/%s/%s/%s", n.Prefix, n.Port, n.ServiceName, n.Namespace, n.CID)
}

//ServiceURI service URI for the NamedKey.
func (n *NamedKey) ServiceURI() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%s/%s/%s", n.Prefix, n.ServiceName, n.Namespace, n.CID)
}

// Reference reference
func (n *NamedKey) Reference(backport int32) string {
	return (&NamedKey{
		Prefix:      n.Prefix,
		Namespace:   n.Namespace,
		CID:         n.CID,
		Port:        backport,
		ServiceName: n.ServiceName}).Key()
}

// URIfromService build ServiceURI from service
func URIfromService(svc *v1.Service) string {
	return fmt.Sprintf("%s/%s/%s/%s", DEFAULT_PREFIX, svc.Name, svc.Namespace, CLUSTER_ID)
}

// LoadNamedKey build NamedKey from string.
func LoadNamedKey(key string) (*NamedKey, error) {
	metas := strings.Split(key, "/")
	if len(metas) != 5 || metas[0] != DEFAULT_PREFIX {
		return nil, formatError{key: key}
	}
	port, err := strconv.Atoi(metas[1])
	if err != nil {
		return nil, err
	}
	return &NamedKey{
		CID:         metas[4],
		Namespace:   metas[3],
		ServiceName: metas[2],
		Port:        int32(port),
		Prefix:      DEFAULT_PREFIX}, nil
}

// Listener loadbalancer listener
type Listener struct {
	Name string
	// NamedKey Map between ServiceName and Listener from console view.
	NamedKey *NamedKey

	// Proto is the protocol from console view
	Proto string

	// TransforedProto is the real protocol that a listener indicated.
	TransforedProto string

	Port int32

	// NodePort Backend server port
	NodePort int32

	// ServiceName reference from k8s service
	Service *v1.Service

	// LoadBalancerID service connected SLB.
	LoadBalancerID string

	// Action indicate the operate method. ADD UPDATE DELETE
	Action string

	Client ClientSLBSDK

	VGroups *vgroups

	VServerGroupId string
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
	switch strings.ToUpper(n.TransforedProto) {
	case "TCP":
		return &tcp{n}
	case "UDP":
		return &udp{n}
	case "HTTP":
		return &http{n}
	case "HTTPS":
		return &https{n}
	}
	return &tcp{n}
}

// Apply apply listener operate . add/update/delete etc.
func (n *Listener) Apply(ctx context.Context) error {
	klog.Infof("apply %s listener for %v with trans protocol %s", n.Action, n.NamedKey, n.TransforedProto)
	switch n.Action {
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
	return fmt.Errorf("UnKnownAction: %s, %s/%s", n.Action, n.Service.Namespace, n.Service.Name)
}

// Start start listener
func (n *Listener) Start(ctx context.Context) error {
	return n.Client.StartLoadBalancerListener(
		ctx, n.LoadBalancerID, int(n.Port),
	)
}

// Describe describe listener
func (n *Listener) Describe(ctx context.Context) error {

	return fmt.Errorf("unimplemented")
}

// Remove remove Listener
func (n *Listener) Remove(ctx context.Context) error {
	err := n.Client.StopLoadBalancerListener(ctx, n.LoadBalancerID, int(n.Port))
	if err != nil {
		return err
	}
	return n.Client.DeleteLoadBalancerListener(ctx, n.LoadBalancerID, int(n.Port))
}

func (n *Listener) findVgroup(key string) string {
	for _, v := range *n.VGroups {
		if v.NamedKey.Key() == key {
			klog.Infof("found: key=%s, groupid=%s, try use vserver group mode.", key, v.VGroupId)
			return v.VGroupId
		}
	}
	klog.Infof("find: vserver group [%s] does not found. use default backend group.", key)
	return STRINGS_EMPTY
}

// STRINGS_EMPTY empty string
var STRINGS_EMPTY = ""

// Listeners listeners collection
type Listeners []*Listener

// EnsureListeners make sure listeners reconciled
// 1. First, build listeners config from aliyun API output.
// 2. Second, build listeners from k8s service object.
// 3. Third, Merge the up two listeners to decide whether add/update/remove is needed.
// 4. Do update.  Clean unused vserver group.
func EnsureListeners(
	ctx context.Context,
	slbins *LoadBalancerClient,
	service *v1.Service,
	lb *slb.LoadBalancerType,
	vgs *vgroups,
) error {

	local, err := BuildListenersFromService(service, lb, slbins.c, vgs)
	if err != nil {
		return fmt.Errorf("build listener from service: %s", err.Error())
	}

	// Merge listeners generate an listener list to be updated/deleted/added.
	updates, err := BuildActionsForListeners(ctx, service, local, BuildListenersFromAPI(service, lb, slbins.c, vgs))
	if err != nil {
		return fmt.Errorf("merge listener: %s", err.Error())
	}
	utils.Logf(service, "ensure listener: %d updates for %s", len(updates), lb.LoadBalancerId)

	// make https come first.
	// ensure https listeners to be created first for http forward
	sort.SliceStable(
		updates,
		func(i, j int) bool {
			// 1. https comes first.
			// 2. DELETE action comes before https
			if isDeleteAction(updates[i].Action) {
				return true
			}
			if isDeleteAction(updates[j].Action) {
				return false
			}
			if strings.ToUpper(
				updates[i].TransforedProto,
			) == "HTTPS" {
				return true
			}
			return false
		},
	)
	// do update/add/delete
	for _, up := range updates {
		err := up.Apply(ctx)
		if err != nil {
			return fmt.Errorf("ensure listener: %s", err.Error())
		}
	}

	return CleanUPVGroupMerged(ctx, slbins, service, lb, vgs)
}

func isDeleteAction(action string) bool { return action == ACTION_DELETE }

// EnsureListenersDeleted Only listener which owned by my service was deleted.
func EnsureListenersDeleted(
	ctx context.Context,
	client ClientSLBSDK,
	service *v1.Service,
	lb *slb.LoadBalancerType,
	vgs *vgroups,
) error {

	local, err := BuildListenersFromService(service, lb, client, vgs)
	if err != nil {
		return fmt.Errorf("build listener from service: %s", err.Error())
	}
	remote := BuildListenersFromAPI(service, lb, client, vgs)

	// make http come first.
	// ensure http listeners to be removed before https
	sort.SliceStable(
		local,
		func(i, j int) bool {
			// 1. http comes first.
			if strings.ToUpper(
				local[i].TransforedProto,
			) == "HTTP" {
				return true
			}
			return false
		},
	)

	for _, loc := range local {
		for _, rem := range remote {
			hasUserNode, err := rem.listenerHasUserManagedNode(ctx)
			if err != nil {
				return fmt.Errorf("ensure listener: %s", err.Error())
			}
			if hasUserNode {
				klog.Infof("%s port %d vgroup has user managed node, skip", rem.NamedKey, rem.Port)
				continue
			}
			if !isManagedByMyService(service, rem) {
				continue
			}
			if loc.Port == rem.Port {
				err := loc.Remove(ctx)
				if err != nil {
					return fmt.Errorf("ensure listener: %s", err.Error())
				}
			}
		}
	}

	return CleanUPVGroupDirect(ctx, vgs)
}

func isManagedByMyService(svc *v1.Service, remote *Listener) bool {

	return remote.NamedKey != nil &&
		remote.NamedKey.ServiceURI() == URIfromService(svc)
}

func isProtocolMatch(local, remote *Listener) bool {
	return local.TransforedProto == remote.TransforedProto
}

func isPortMatch(local, remote *Listener) bool {
	return local.Port == remote.Port
}

// 1. We update listener to the latest version2 when updation is needed.
// 2. We assume listener with an empty name to be legacy version.
// 3. We assume listener with an arbitrary name to be user managed listener.
// 4. LoadBalancer created by kubernetes is not allowed to be reused.
func BuildActionsForListeners(ctx context.Context, svc *v1.Service, service, console Listeners) (Listeners, error) {
	override := isOverrideListeners(svc)
	var (
		addition = Listeners{}
		updation = Listeners{}
		deletion = Listeners{}
	)
	// For updations and deletions
	for _, remote := range console {
		found := false
		for _, local := range service {
			if remote.Port == local.Port {
				found = true
				// port matched. that is where the conflict case begin.
				// 1. check protocol match.
				if isProtocolMatch(local, remote) {
					// protocol match, need to do update operate no matter managed by whom
					// consider override annotation & user defined loadbalancer
					if !override && isUserDefinedLoadBalancer(svc) {
						// port conflict with user managed slb or listener.
						return nil, fmt.Errorf("PortProtocolConflict] port matched, but conflict with user managed listener. "+
							"Port:%d, ListenerName:%s, svc: %s. Protocol:[source:%s dst:%s]",
							remote.Port, remote.Name, local.NamedKey.Key(), remote.TransforedProto, local.TransforedProto)
					}

					// do update operate
					local.Action = ACTION_UPDATE
					updation = append(updation, local)
					utils.Logf(svc, "found listener with port & protocol match, do update %s", local.NamedKey.Key())
				} else {
					// protocol not match, need to recreate
					if !override && isUserDefinedLoadBalancer(svc) {
						return nil, fmt.Errorf("[PortProtocolConflict] port matched, "+
							"while protocol does not. force override listener %t. source[%v], target[%v]", override, local.NamedKey, remote.NamedKey)
					}
					remote.Action = ACTION_DELETE
					deletion = append(deletion, remote)
					utils.Logf(svc, "found listener with port match while protocol not, do delete & add %s", local.NamedKey.Key())
				}
			}
		}
		// Do not delete any listener that no longer managed by my service
		// for safety. Only conflict case is taking care.
		if !found {
			if isManagedByMyService(svc, remote) {
				// port has user managed nodes, do not delete
				hasUserNode, err := remote.listenerHasUserManagedNode(ctx)
				if err != nil {
					return nil, err
				}
				if hasUserNode {
					utils.Logf(svc, "svc %s do not delete port %d, because backends have user managed nodes", remote.NamedKey.Key(), remote.Port)
					continue
				}
				remote.Action = ACTION_DELETE
				deletion = append(deletion, remote)
				utils.Logf(svc, "found listener[%s] which is no longer needed "+
					"by my service[%s/%s], do delete", remote.NamedKey.Key(), svc.Namespace, svc.Name)
			} else {
				utils.Logf(svc, "port [%d] not managed by my service [%s/%s], skip processing.", remote.Port, svc.Namespace, svc.Name)
			}
		}
	}
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// For additions
	for _, local := range service {
		found := false
		for _, remote := range console {
			if isPortMatch(remote, local) {
				// port match
				if !isProtocolMatch(remote, local) {
					// protocol does not match, do add listener
					break
				}
				// port matched. updated . skip
				found = true
				break
			}
		}
		if !found {
			local.Action = ACTION_ADD
			addition = append(addition, local)
			utils.Logf(svc, "add new listener %s", local.NamedKey.Key())
		}
	}

	// Pls be careful of the sequence. deletion first,then addition, last updation
	return append(append(deletion, addition...), updation...), nil
}

// BuildListenersFromService Build expected listeners
func BuildListenersFromService(
	svc *v1.Service,
	lb *slb.LoadBalancerType,
	client ClientSLBSDK,
	vgrps *vgroups,
) (Listeners, error) {
	listeners := Listeners{}

	for _, port := range svc.Spec.Ports {
		proto, err := Protocol(serviceAnnotation(svc, ServiceAnnotationLoadBalancerProtocolPort), port)
		if err != nil {
			return nil, err
		}
		n := Listener{
			NamedKey: &NamedKey{
				CID:         CLUSTER_ID,
				Namespace:   svc.Namespace,
				ServiceName: svc.Name,
				Port:        port.Port,
				Prefix:      DEFAULT_PREFIX},
			Port:            port.Port,
			NodePort:        port.NodePort,
			Proto:           string(port.Protocol),
			Service:         svc,
			TransforedProto: proto,
			Client:          client,
			VGroups:         vgrps,
			LoadBalancerID:  lb.LoadBalancerId,
		}
		if utils.IsENIBackendType(svc) {
			n.NodePort = port.TargetPort.IntVal
		}
		n.Name = n.NamedKey.Key()
		listeners = append(listeners, &n)
	}
	return listeners, nil
}

// BuildListenersFromAPI Load current listeners
func BuildListenersFromAPI(
	service *v1.Service,
	lb *slb.LoadBalancerType,
	client ClientSLBSDK,
	vgrps *vgroups,
) (listeners Listeners) {
	ports := lb.ListenerPortsAndProtocol.ListenerPortAndProtocol
	for _, port := range ports {
		key, err := LoadNamedKey(port.Description)
		if err != nil {
			klog.Warningf("alicloud: error parse listener description[%s]. %s", port.Description, err.Error())
		}
		proto := port.ListenerProtocol
		if strings.ToUpper(proto) == "HTTP" ||
			strings.ToUpper(proto) == "HTTPS" {
			proto = "TCP"
		}
		n := Listener{
			Name:            port.Description,
			NamedKey:        key,
			Port:            int32(port.ListenerPort),
			Proto:           proto,
			TransforedProto: port.ListenerProtocol,
			LoadBalancerID:  lb.LoadBalancerId,
			Service:         service,
			Client:          client,
			VGroups:         vgrps,
		}
		listeners = append(listeners, &n)
	}
	return listeners
}

type tcp struct{ *Listener }

func (t *tcp) Add(ctx context.Context) error {
	def, _ := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerTCPListener(
		ctx,
		&slb.CreateLoadBalancerTCPListenerArgs{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      int(t.Port),
			BackendServerPort: int(t.NodePort),
			//Health Check
			Scheduler:          slb.SchedulerType(def.Scheduler),
			Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
			PersistenceTimeout: def.PersistenceTimeout,
			Description:        t.NamedKey.Key(),

			VServerGroupId:            t.findVgroup(t.NamedKey.Reference(t.NodePort)),
			AclType:                   def.AclType,
			AclStatus:                 def.AclStatus,
			AclId:                     def.AclID,
			HealthCheckType:           def.HealthCheckType,
			HealthCheckURI:            def.HealthCheckURI,
			HealthCheckConnectPort:    def.HealthCheckConnectPort,
			HealthyThreshold:          def.HealthyThreshold,
			UnhealthyThreshold:        def.UnhealthyThreshold,
			HealthCheckConnectTimeout: def.HealthCheckConnectTimeout,
			HealthCheckInterval:       def.HealthCheckInterval,
			HealthCheck:               def.HealthCheck,
			HealthCheckDomain:         def.HealthCheckDomain,
			HealthCheckHttpCode:       def.HealthCheckHttpCode,
		})
}

func (t *tcp) Update(ctx context.Context) error {
	def, request := ExtractAnnotationRequest(t.Service)

	response, err := t.Client.DescribeLoadBalancerTCPListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return fmt.Errorf("update tcp listener: %s", err.Error())
	}
	utils.Logf(t.Service, "tcp listener %d status is %s.", t.Port, response.Status)
	if response.Status == slb.Stopped {
		if err = t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port)); err != nil {
			return fmt.Errorf("start tcp listener error: %s", err.Error())
		}
	}
	config := &slb.SetLoadBalancerTCPListenerAttributeArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		Description:       t.NamedKey.Key(),
		//Health Check
		Scheduler:          slb.SchedulerType(response.Scheduler),
		Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
		PersistenceTimeout: response.PersistenceTimeout,
		VServerGroup:       slb.OnFlag,
		VServerGroupId:     t.findVgroup(t.NamedKey.Reference(t.NodePort)),

		AclType:                   response.AclType,
		AclStatus:                 response.AclStatus,
		AclId:                     response.AclId,
		HealthCheckType:           response.HealthCheckType,
		HealthCheckURI:            response.HealthCheckURI,
		HealthCheckConnectPort:    response.HealthCheckConnectPort,
		HealthyThreshold:          response.HealthyThreshold,
		UnhealthyThreshold:        response.UnhealthyThreshold,
		HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
		HealthCheckInterval:       response.HealthCheckInterval,
		HealthCheck:               response.HealthCheck,
		HealthCheckHttpCode:       response.HealthCheckHttpCode,
		HealthCheckDomain:         response.HealthCheckDomain,
	}
	needUpdate := false
	/*
		if request.Bandwidth != 0 &&
			def.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = def.Bandwidth
			klog.V(2).Infof("TCP listener checker [bandwidth] changed, request=%d. response=%d", def.Bandwidth, response.Bandwidth)
		}
	*/

	if request.AclStatus != "" &&
		def.AclStatus != response.AclStatus {
		needUpdate = true
		config.AclStatus = def.AclStatus
	}
	if request.AclID != "" &&
		def.AclID != response.AclId {
		needUpdate = true
		config.AclId = def.AclID
	}
	if request.AclType != "" &&
		def.AclType != response.AclType {
		needUpdate = true
		config.AclType = def.AclType
	}

	if request.Scheduler != "" &&
		def.Scheduler != string(response.Scheduler) {
		needUpdate = true
		config.Scheduler = slb.SchedulerType(def.Scheduler)
	}

	// todo: perform healthcheck update.
	if request.HealthCheckType != "" &&
		def.HealthCheckType != response.HealthCheckType {
		needUpdate = true
		config.HealthCheckType = def.HealthCheckType
	}
	if request.HealthCheckURI != "" &&
		def.HealthCheckURI != response.HealthCheckURI {
		needUpdate = true
		config.HealthCheckURI = def.HealthCheckURI
	}
	if request.HealthCheckConnectPort != 0 &&
		def.HealthCheckConnectPort != response.HealthCheckConnectPort {
		needUpdate = true
		config.HealthCheckConnectPort = def.HealthCheckConnectPort
	}
	if request.HealthyThreshold != 0 &&
		def.HealthyThreshold != response.HealthyThreshold {
		needUpdate = true
		config.HealthyThreshold = def.HealthyThreshold
	}
	if request.UnhealthyThreshold != 0 &&
		def.UnhealthyThreshold != response.UnhealthyThreshold {
		needUpdate = true
		config.UnhealthyThreshold = def.UnhealthyThreshold
	}
	if request.HealthCheckConnectTimeout != 0 &&
		def.HealthCheckConnectTimeout != response.HealthCheckConnectTimeout {
		needUpdate = true
		config.HealthCheckConnectTimeout = def.HealthCheckConnectTimeout
	}
	if request.HealthCheckInterval != 0 &&
		def.HealthCheckInterval != response.HealthCheckInterval {
		needUpdate = true
		config.HealthCheckInterval = def.HealthCheckInterval
	}
	if request.PersistenceTimeout != nil &&
		*def.PersistenceTimeout != *response.PersistenceTimeout {
		needUpdate = true
		config.PersistenceTimeout = def.PersistenceTimeout
	}
	if request.HealthCheckHttpCode != "" &&
		def.HealthCheckHttpCode != response.HealthCheckHttpCode {
		needUpdate = true
		config.HealthCheckHttpCode = def.HealthCheckHttpCode
	}
	if request.HealthCheckDomain != "" &&
		def.HealthCheckDomain != response.HealthCheckDomain {
		needUpdate = true
		config.HealthCheckDomain = def.HealthCheckDomain
	}
	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		config.BackendServerPort = int(t.NodePort)
		klog.V(2).Infof("tcp listener [BackendServerPort] changed, request=%d. response=%d, recreate.", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		err = t.Client.CreateLoadBalancerTCPListener(ctx, (*slb.CreateLoadBalancerTCPListenerArgs)(config))
		if err != nil {
			return err
		}
		return t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
	}
	if !needUpdate {
		utils.Logf(t.Service, "tcp listener did not change, skip [update], port=[%d], nodeport=[%d]", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	utils.Logf(t.Service, "TCP listener checker changed, request update listener attribute [%s]", t.LoadBalancerID)
	klog.V(5).Infof(PrettyJson(def))
	klog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerTCPListenerAttribute(ctx, config)
}

func (t *tcp) Describe(ctx context.Context) error {
	response, err := t.Client.DescribeLoadBalancerTCPListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return fmt.Errorf("%s DescribeLoadBalancer tcp listener %d: %s", t.Name, t.Port, err.Error())
	}
	if response == nil {
		return fmt.Errorf("%s DescribeLoadBalancer tcp listener %d: response is nil", t.Name, t.Port)
	}
	// other fields to be filled
	t.VServerGroupId = response.VServerGroupId
	return nil
}

type udp struct{ *Listener }

func (t *udp) Describe(ctx context.Context) error {
	response, err := t.Client.DescribeLoadBalancerUDPListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return fmt.Errorf("%s DescribeLoadBalancer udp listener %d: %s", t.Name, t.Port, err.Error())
	}
	if response == nil {
		return fmt.Errorf("%s DescribeLoadBalancer udp listener %d: response is nil", t.Name, t.Port)
	}
	// other fields to be filled
	t.VServerGroupId = response.VServerGroupId
	return nil
}

func (t *udp) Add(ctx context.Context) error {
	def, _ := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerUDPListener(
		ctx,
		&slb.CreateLoadBalancerUDPListenerArgs{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      int(t.Port),
			BackendServerPort: int(t.NodePort),
			Description:       t.NamedKey.Key(),
			VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
			//Health Check
			Scheduler:          slb.SchedulerType(def.Scheduler),
			Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
			PersistenceTimeout: def.PersistenceTimeout,

			AclType:   def.AclType,
			AclStatus: def.AclStatus,
			AclId:     def.AclID,
			//HealthCheckType:           request.HealthCheckType,
			//HealthCheckURI:            request.HealthCheckURI,
			HealthCheckConnectPort:    def.HealthCheckConnectPort,
			HealthyThreshold:          def.HealthyThreshold,
			UnhealthyThreshold:        def.UnhealthyThreshold,
			HealthCheckConnectTimeout: def.HealthCheckConnectTimeout,
			HealthCheckInterval:       def.HealthCheckInterval,
			HealthCheck:               def.HealthCheck,
		},
	)
}

func (t *udp) Update(ctx context.Context) error {
	def, request := ExtractAnnotationRequest(t.Service)
	response, err := t.Client.DescribeLoadBalancerUDPListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	utils.Logf(t.Service, "udp listener %d status is %s.", t.Port, response.Status)
	if response.Status == slb.Stopped {
		if err = t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port)); err != nil {
			return fmt.Errorf("start udp listener error: %s", err.Error())
		}
	}
	config := &slb.SetLoadBalancerUDPListenerAttributeArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		Description:       t.NamedKey.Key(),
		VServerGroup:      slb.OnFlag,
		VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
		AclType:           response.AclType,
		AclStatus:         response.AclStatus,
		AclId:             response.AclId,
		//Health Check
		Scheduler:          slb.SchedulerType(response.Scheduler),
		Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
		PersistenceTimeout: response.PersistenceTimeout,
		//HealthCheckType:           response.HealthCheckType,
		//HealthCheckURI:            response.HealthCheckURI,
		HealthCheckConnectPort:    response.HealthCheckConnectPort,
		HealthyThreshold:          response.HealthyThreshold,
		UnhealthyThreshold:        response.UnhealthyThreshold,
		HealthCheckConnectTimeout: response.HealthCheckConnectTimeout,
		HealthCheckInterval:       response.HealthCheckInterval,
		HealthCheck:               response.HealthCheck,
	}
	needUpdate := false
	/*
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			klog.V(2).Infof("UDP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
	*/
	if request.AclStatus != "" &&
		def.AclStatus != response.AclStatus {
		needUpdate = true
		config.AclStatus = def.AclStatus
	}
	if request.AclID != "" &&
		def.AclID != response.AclId {
		needUpdate = true
		config.AclId = def.AclID
	}
	if request.AclType != "" &&
		def.AclType != response.AclType {
		needUpdate = true
		config.AclType = def.AclType
	}

	if request.Scheduler != "" &&
		def.Scheduler != string(response.Scheduler) {
		needUpdate = true
		config.Scheduler = slb.SchedulerType(def.Scheduler)
	}
	// todo: perform healthcheck update.
	if request.HealthCheckConnectPort != 0 &&
		def.HealthCheckConnectPort != response.HealthCheckConnectPort {
		needUpdate = true
		config.HealthCheckConnectPort = def.HealthCheckConnectPort
	}
	if request.HealthyThreshold != 0 &&
		def.HealthyThreshold != response.HealthyThreshold {
		needUpdate = true
		config.HealthyThreshold = def.HealthyThreshold
	}
	if request.UnhealthyThreshold != 0 &&
		def.UnhealthyThreshold != response.UnhealthyThreshold {
		needUpdate = true
		config.UnhealthyThreshold = def.UnhealthyThreshold
	}
	if request.HealthCheckConnectTimeout != 0 &&
		def.HealthCheckConnectTimeout != response.HealthCheckConnectTimeout {
		needUpdate = true
		config.HealthCheckConnectTimeout = def.HealthCheckConnectTimeout
	}
	if request.HealthCheckInterval != 0 &&
		def.HealthCheckInterval != response.HealthCheckInterval {
		needUpdate = true
		config.HealthCheckInterval = def.HealthCheckInterval
	}
	if request.PersistenceTimeout != nil &&
		*def.PersistenceTimeout != *response.PersistenceTimeout {
		needUpdate = true
		config.PersistenceTimeout = def.PersistenceTimeout
	}
	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		config.BackendServerPort = int(t.NodePort)
		utils.Logf(t.Service, "udp listener checker [BackendServerPort] changed, "+
			"request=%d. response=%d", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		err = t.Client.CreateLoadBalancerUDPListener(ctx, (*slb.CreateLoadBalancerUDPListenerArgs)(config))
		if err != nil {
			return err
		}
		return t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
	}

	if !needUpdate {
		utils.Logf(t.Service, "udp listener did not change, skip "+
			"[update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	utils.Logf(t.Service, "UDP listener checker changed, request recreate [%s]\n", t.LoadBalancerID)
	klog.V(5).Infof(PrettyJson(request))
	klog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerUDPListenerAttribute(ctx, config)
}

type http struct{ *Listener }

func (t *http) Describe(ctx context.Context) error {
	response, err := t.Client.DescribeLoadBalancerHTTPListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return fmt.Errorf("%s DescribeLoadBalancer http listener %d: %s", t.Name, t.Port, err.Error())
	}
	if response == nil {
		return fmt.Errorf(" %s DescribeLoadBalancer http listener %d: response is nil", t.Name, t.Port)
	}
	// other fields to be filled
	t.VServerGroupId = response.VServerGroupId
	return nil
}

func (t *http) Add(ctx context.Context) error {
	def, request := ExtractAnnotationRequest(t.Service)
	httpc := &slb.CreateLoadBalancerHTTPListenerArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		Description:       t.NamedKey.Key(),
		VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
		//Health Check
		Scheduler:         slb.SchedulerType(def.Scheduler),
		Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
		StickySession:     def.StickySession,
		StickySessionType: def.StickySessionType,
		CookieTimeout:     def.CookieTimeout,
		Cookie:            def.Cookie,

		AclType:   def.AclType,
		AclStatus: def.AclStatus,
		AclId:     def.AclID,
		//HealthCheckType:           request.HealthCheckType,
		HealthCheckURI:         request.HealthCheckURI,
		HealthCheckConnectPort: request.HealthCheckConnectPort,
		HealthyThreshold:       request.HealthyThreshold,
		UnhealthyThreshold:     request.UnhealthyThreshold,
		//HealthCheckConnectTimeout: request.HealthCheckConnectTimeout,
		HealthCheckInterval: request.HealthCheckInterval,
		HealthCheckDomain:   def.HealthCheckDomain,
		HealthCheck:         def.HealthCheck,
		HealthCheckTimeout:  def.HealthCheckTimeout,
		HealthCheckHttpCode: def.HealthCheckHttpCode,
	}
	forward := forwardPort(def.ForwardPort, t.Port)
	if forward != 0 {
		httpc.ListenerForward = slb.OnFlag
	} else {
		httpc.ListenerForward = slb.OffFlag
	}
	httpc.ForwardPort = int(forward)
	return t.Client.CreateLoadBalancerHTTPListener(ctx, httpc)
}

func forwardPort(port string, target int32) int32 {
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
		return int32(forward)
	}
	return 0
}

func (t *http) Update(ctx context.Context) error {

	def, request := ExtractAnnotationRequest(t.Service)
	response, err := t.Client.DescribeLoadBalancerHTTPListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	utils.Logf(t.Service, "http listener %d status is %s.", t.Port, response.Status)
	if response.Status == slb.Stopped {
		if err = t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port)); err != nil {
			return fmt.Errorf("start http listener error: %s", err.Error())
		}
	}
	config := &slb.SetLoadBalancerHTTPListenerAttributeArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		//Health Check
		Scheduler:         slb.SchedulerType(response.Scheduler),
		Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
		StickySession:     response.StickySession,
		StickySessionType: response.StickySessionType,
		CookieTimeout:     response.CookieTimeout,
		Cookie:            response.Cookie,
		Description:       t.NamedKey.Key(),
		VServerGroup:      slb.OnFlag,
		VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),

		AclType:                response.AclType,
		AclStatus:              response.AclStatus,
		AclId:                  response.AclId,
		HealthCheck:            response.HealthCheck,
		HealthCheckURI:         response.HealthCheckURI,
		HealthCheckConnectPort: response.HealthCheckConnectPort,
		HealthyThreshold:       response.HealthyThreshold,
		UnhealthyThreshold:     response.UnhealthyThreshold,
		HealthCheckTimeout:     response.HealthCheckTimeout,
		HealthCheckDomain:      response.HealthCheckDomain,
		HealthCheckHttpCode:    response.HealthCheckHttpCode,
		HealthCheckInterval:    response.HealthCheckInterval,
	}
	needUpdate := false
	needRecreate := false
	/*
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			klog.V(2).Infof("HTTP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
	*/
	if request.AclStatus != "" &&
		def.AclStatus != response.AclStatus {
		needUpdate = true
		config.AclStatus = def.AclStatus
	}
	if request.AclID != "" &&
		def.AclID != response.AclId {
		needUpdate = true
		config.AclId = def.AclID
	}
	if request.AclType != "" &&
		def.AclType != response.AclType {
		needUpdate = true
		config.AclType = def.AclType
	}
	if request.Scheduler != "" &&
		def.Scheduler != string(response.Scheduler) {
		needUpdate = true
		config.Scheduler = slb.SchedulerType(def.Scheduler)
	}
	// todo: perform healthcheck update.
	if request.HealthCheck != "" &&
		def.HealthCheck != response.HealthCheck {
		needUpdate = true
		config.HealthCheck = def.HealthCheck
	}
	if request.HealthCheckURI != "" &&
		def.HealthCheckURI != response.HealthCheckURI {
		needUpdate = true
		config.HealthCheckURI = def.HealthCheckURI
	}
	if request.HealthCheckConnectPort != 0 &&
		def.HealthCheckConnectPort != response.HealthCheckConnectPort {
		needUpdate = true
		config.HealthCheckConnectPort = def.HealthCheckConnectPort
	}
	if request.HealthyThreshold != 0 &&
		def.HealthyThreshold != response.HealthyThreshold {
		needUpdate = true
		config.HealthyThreshold = def.HealthyThreshold
	}
	if request.UnhealthyThreshold != 0 &&
		def.UnhealthyThreshold != response.UnhealthyThreshold {
		needUpdate = true
		config.UnhealthyThreshold = def.UnhealthyThreshold
	}
	if request.HealthCheckTimeout != 0 &&
		def.HealthCheckTimeout != response.HealthCheckTimeout {
		needUpdate = true
		config.HealthCheckTimeout = def.HealthCheckTimeout
	}
	if request.HealthCheckInterval != 0 &&
		def.HealthCheckInterval != response.HealthCheckInterval {
		needUpdate = true
		config.HealthCheckInterval = def.HealthCheckInterval
	}
	if string(request.StickySession) != "" &&
		def.StickySession != response.StickySession {
		needUpdate = true
		config.StickySession = def.StickySession
	}
	if string(request.StickySessionType) != "" &&
		def.StickySessionType != response.StickySessionType {
		needUpdate = true
		config.StickySessionType = def.StickySessionType
	}
	if request.Cookie != "" &&
		def.Cookie != response.Cookie {
		needUpdate = true
		config.Cookie = def.Cookie
	}
	if request.CookieTimeout != 0 &&
		def.CookieTimeout != response.CookieTimeout {
		needUpdate = true
		config.CookieTimeout = def.CookieTimeout
	}
	if request.HealthCheckHttpCode != "" &&
		def.HealthCheckHttpCode != response.HealthCheckHttpCode {
		needUpdate = true
		config.HealthCheckHttpCode = def.HealthCheckHttpCode
	}
	if request.HealthCheckDomain != "" &&
		def.HealthCheckDomain != response.HealthCheckDomain {
		needUpdate = true
		config.HealthCheckDomain = def.HealthCheckDomain
	}
	forward := forwardPort(def.ForwardPort, t.Port)
	if forward != 0 {
		if response.ListenerForward != slb.OnFlag {
			needRecreate = true
			config.ListenerForward = slb.OnFlag
		}
	} else {
		if response.ListenerForward != slb.OffFlag {
			needRecreate = true
			config.ListenerForward = slb.OffFlag
		}
	}
	config.ForwardPort = int(forward)

	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		// listener with listenerforward status on, no need to reRecreate
		if response.ListenerForward == slb.OffFlag {
			needRecreate = true
		}
	}

	if needRecreate {

		config.BackendServerPort = int(t.NodePort)
		utils.Logf(t.Service, "HTTP listener checker [BackendServerPort]"+
			" changed, request=%d. response=%d. Recreate http listener.", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		err = t.Client.CreateLoadBalancerHTTPListener(ctx, (*slb.CreateLoadBalancerHTTPListenerArgs)(config))
		if err != nil {
			return err
		}
		return t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
	}

	if response.ListenerForward == slb.OnFlag {
		utils.Logf(t.Service, "%d ListenerForward is on, cannot update listener", t.Port)
		// no update needed.  skip
		return nil
	}

	if !needUpdate {
		utils.Logf(t.Service, "http listener did not change, skip [update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	utils.Logf(t.Service, "http listener checker changed, request update [%s]\n", t.LoadBalancerID)
	klog.V(5).Infof(PrettyJson(request))
	klog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerHTTPListenerAttribute(ctx, config)
}

type https struct{ *Listener }

func (t *https) Describe(ctx context.Context) error {
	response, err := t.Client.DescribeLoadBalancerHTTPSListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return fmt.Errorf("%s DescribeLoadBalancer https listener %d: %s", t.Name, t.Port, err.Error())
	}
	if response == nil {
		return fmt.Errorf("%s DescribeLoadBalancer https listener %d: response is nil", t.Name, t.Port)
	}
	// other fields to be filled
	t.VServerGroupId = response.VServerGroupId
	return nil
}

func (t *https) Add(ctx context.Context) error {

	def, request := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerHTTPSListener(
		ctx,
		&slb.CreateLoadBalancerHTTPSListenerArgs{
			HTTPListenerType: slb.HTTPListenerType{
				LoadBalancerId:    t.LoadBalancerID,
				ListenerPort:      int(t.Port),
				BackendServerPort: int(t.NodePort),
				Description:       t.NamedKey.Key(),
				VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
				AclType:           def.AclType,
				AclStatus:         def.AclStatus,
				AclId:             def.AclID,
				//Health Check
				Scheduler:         slb.SchedulerType(def.Scheduler),
				HealthCheck:       def.HealthCheck,
				Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
				StickySession:     def.StickySession,
				StickySessionType: def.StickySessionType,
				Cookie:            def.Cookie,
				CookieTimeout:     def.CookieTimeout,

				HealthCheckURI:         def.HealthCheckURI,
				HealthCheckConnectPort: def.HealthCheckConnectPort,
				HealthyThreshold:       def.HealthyThreshold,
				UnhealthyThreshold:     def.UnhealthyThreshold,
				HealthCheckTimeout:     def.HealthCheckTimeout,
				HealthCheckInterval:    def.HealthCheckInterval,
				HealthCheckDomain:      def.HealthCheckDomain,
				HealthCheckHttpCode:    def.HealthCheckHttpCode,
			},
			ServerCertificateId: request.CertID,
		},
	)
}

func (t *https) Update(ctx context.Context) error {
	def, request := ExtractAnnotationRequest(t.Service)
	response, err := t.Client.DescribeLoadBalancerHTTPSListenerAttribute(ctx, t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	utils.Logf(t.Service, "https listener %d status is %s.", t.Port, response.Status)
	if response.Status == slb.Stopped {
		if err = t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port)); err != nil {
			return fmt.Errorf("start https listener error: %s", err.Error())
		}
	}
	config := &slb.SetLoadBalancerHTTPSListenerAttributeArgs{
		HTTPListenerType: slb.HTTPListenerType{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      response.ListenerPort,
			BackendServerPort: response.BackendServerPort,
			Description:       t.NamedKey.Key(),
			VServerGroup:      slb.OnFlag,
			VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
			//Health Check
			Scheduler:         slb.SchedulerType(response.Scheduler),
			HealthCheck:       response.HealthCheck,
			Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
			StickySession:     response.StickySession,
			StickySessionType: response.StickySessionType,
			CookieTimeout:     response.CookieTimeout,
			Cookie:            response.Cookie,

			AclType:                response.AclType,
			AclStatus:              response.AclStatus,
			AclId:                  response.AclId,
			HealthCheckURI:         response.HealthCheckURI,
			HealthCheckConnectPort: response.HealthCheckConnectPort,
			HealthyThreshold:       response.HealthyThreshold,
			UnhealthyThreshold:     response.UnhealthyThreshold,
			HealthCheckTimeout:     response.HealthCheckTimeout,
			HealthCheckInterval:    response.HealthCheckInterval,
			HealthCheckHttpCode:    response.HealthCheckHttpCode,
			HealthCheckDomain:      response.HealthCheckDomain,
		},
		ServerCertificateId: response.ServerCertificateId,
	}

	needUpdate := false
	/*
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			klog.Infof("HTTPS listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
	*/
	// todo: perform healthcheck update.
	if request.AclStatus != "" &&
		def.AclStatus != response.AclStatus {
		needUpdate = true
		config.AclStatus = def.AclStatus
	}
	if request.AclID != "" &&
		def.AclID != response.AclId {
		needUpdate = true
		config.AclId = def.AclID
	}
	if request.AclType != "" &&
		def.AclType != response.AclType {
		needUpdate = true
		config.AclType = def.AclType
	}
	if request.Scheduler != "" &&
		def.Scheduler != string(response.Scheduler) {
		needUpdate = true
		config.Scheduler = slb.SchedulerType(def.Scheduler)
	}
	if request.HealthCheck != "" &&
		def.HealthCheck != response.HealthCheck {
		needUpdate = true
		config.HealthCheck = def.HealthCheck
	}
	if request.HealthCheckURI != "" &&
		def.HealthCheckURI != response.HealthCheckURI {
		needUpdate = true
		config.HealthCheckURI = def.HealthCheckURI
	}
	if request.HealthCheckConnectPort != 0 &&
		def.HealthCheckConnectPort != response.HealthCheckConnectPort {
		needUpdate = true
		config.HealthCheckConnectPort = def.HealthCheckConnectPort
	}
	if request.HealthyThreshold != 0 &&
		def.HealthyThreshold != response.HealthyThreshold {
		needUpdate = true
		config.HealthyThreshold = def.HealthyThreshold
	}
	if request.UnhealthyThreshold != 0 &&
		def.UnhealthyThreshold != response.UnhealthyThreshold {
		needUpdate = true
		config.UnhealthyThreshold = def.UnhealthyThreshold
	}
	if request.HealthCheckTimeout != 0 &&
		def.HealthCheckTimeout != response.HealthCheckTimeout {
		needUpdate = true
		config.HealthCheckTimeout = def.HealthCheckTimeout
	}
	if request.HealthCheckInterval != 0 &&
		def.HealthCheckInterval != response.HealthCheckInterval {
		needUpdate = true
		config.HealthCheckInterval = def.HealthCheckInterval
	}

	if string(request.StickySession) != "" &&
		def.StickySession != response.StickySession {
		needUpdate = true
		config.StickySession = def.StickySession
	}
	if string(request.StickySessionType) != "" &&
		def.StickySessionType != response.StickySessionType {
		needUpdate = true
		config.StickySessionType = def.StickySessionType
	}
	if request.Cookie != "" &&
		def.Cookie != response.Cookie {
		needUpdate = true
		config.Cookie = def.Cookie
	}
	if request.CookieTimeout != 0 &&
		def.CookieTimeout != response.CookieTimeout {
		needUpdate = true
		config.CookieTimeout = def.CookieTimeout
	}
	if request.HealthCheckHttpCode != "" &&
		def.HealthCheckHttpCode != response.HealthCheckHttpCode {
		needUpdate = true
		config.HealthCheckHttpCode = def.HealthCheckHttpCode
	}
	if request.HealthCheckDomain != "" &&
		def.HealthCheckDomain != response.HealthCheckDomain {
		needUpdate = true
		config.HealthCheckDomain = def.HealthCheckDomain
	}
	if request.CertID != "" &&
		def.CertID != response.ServerCertificateId {
		needUpdate = true
		config.ServerCertificateId = def.CertID
	}
	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		needUpdate = true
		config.BackendServerPort = int(t.NodePort)
		utils.Logf(t.Service, "listener checker [BackendServerPort] changed, request=%d. response=%d", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		err = t.Client.CreateLoadBalancerHTTPSListener(ctx, (*slb.CreateLoadBalancerHTTPSListenerArgs)(config))
		if err != nil {
			return err
		}
		return t.Client.StartLoadBalancerListener(ctx, t.LoadBalancerID, int(t.Port))
	}

	if !needUpdate {
		utils.Logf(t.Service, "https listener did not change, skip [update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	utils.Logf(t.Service, "https listener checker changed, request recreate [%s]\n", t.LoadBalancerID)
	klog.V(5).Infof(PrettyJson(request))
	klog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerHTTPSListenerAttribute(ctx, config)
}

func (n *Listener) listenerHasUserManagedNode(ctx context.Context) (bool, error) {
	err := n.Instance().Describe(ctx)
	if err != nil {
		return false, err
	}

	remoteVg, err := n.Client.DescribeVServerGroupAttribute(
		ctx,
		&slb.DescribeVServerGroupAttributeArgs{
			VServerGroupId: n.VServerGroupId,
		})
	if err != nil {
		return false, fmt.Errorf("DescribeVServerGroupAttribute: "+
			"failed to DescribeVServerGroupAttribute vgroup id[%s], error: %s", n.VServerGroupId, err.Error())
	}

	for _, backend := range remoteVg.BackendServers.BackendServer {
		klog.Infof("%s,%s", backend.Description, n.NamedKey.Reference(n.NodePort))
		if isUserManagedNode(backend.Description) {
			klog.Infof("%s vgroup %s has user managed node, node ip is %s, node id is %s",
				n.NamedKey, n.VServerGroupId, backend.ServerIp, backend.ServerId)
			return true, nil
		}
	}

	return false, nil

}
