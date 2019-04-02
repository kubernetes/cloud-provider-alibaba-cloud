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
	//"errors"
	"fmt"
	"github.com/denverdino/aliyungo/slb"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"strconv"
	"strings"
)

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

func Protocol(annotation string, port v1.ServicePort) (string, error) {

	if annotation == "" {
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
			return pp[0], nil
		}
	}
	return strings.ToLower(string(port.Protocol)), nil
}

type IListener interface {
	Describe() error
	Add() error
	Remove() error
	Update() error
}

var FORMAT_ERROR = "ListenerName Format Error: k8s/${port}/${service}/${namespace}/${clusterid} format is expected"

type formatError struct{ key string }

func (f formatError) Error() string { return fmt.Sprintf("%s. Got [%s]", FORMAT_ERROR, f.key) }

var DEFAULT_PREFIX = "k8s"

type NamedKey struct {
	Prefix      string
	CID         string
	Namespace   string
	ServiceName string
	Port        int32
}

func (n *NamedKey) Key() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%d/%s/%s/%s", n.Prefix, n.Port, n.ServiceName, n.Namespace, n.CID)
}

func (n *NamedKey) ServiceURI() string {
	if n.Prefix == "" {
		n.Prefix = DEFAULT_PREFIX
	}
	return fmt.Sprintf("%s/%s/%s/%s", n.Prefix, n.ServiceName, n.Namespace, n.CID)
}

func (n *NamedKey) Reference(backport int32) string {
	return (&NamedKey{
		Prefix:      n.Prefix,
		Namespace:   n.Namespace,
		CID:         n.CID,
		Port:        backport,
		ServiceName: n.ServiceName}).Key()
}

func URIfromService(svc *v1.Service) string {
	return fmt.Sprintf("%s/%s/%s/%s", DEFAULT_PREFIX, svc.Name, svc.Namespace, CLUSTER_ID)
}

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
}

var (
	ACTION_ADD    = "ADD"
	ACTION_UPDATE = "UPDATE"
	ACTION_DELETE = "DELETE"
)

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

func (n *Listener) Apply() error {
	glog.Infof("apply: check listener for %s, name:[%s]", n.Action, n.Name)
	glog.V(6).Infof("Listener: %s => \n%+v\n", n.Action, PrettyJson(n))
	switch n.Action {
	case ACTION_UPDATE:
		err := n.Instance().Update()
		if err != nil {
			return err
		}
		return n.Start()
	case ACTION_ADD:
		err := n.Instance().Add()
		if err != nil {
			return err
		}
		return n.Start()
	case ACTION_DELETE:
		return n.Instance().Remove()
	}
	return fmt.Errorf("UnKnownAction: %s, %s/%s", n.Action, n.Service.Namespace, n.Service.Name)
}

func (n *Listener) Start() error {
	return n.Client.StartLoadBalancerListener(
		n.LoadBalancerID, int(n.Port))
}

func (t *Listener) Describe() error {

	return fmt.Errorf("unimplemented")
}

func (t *Listener) Remove() error {
	err := t.Client.StopLoadBalancerListener(t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	return t.Client.DeleteLoadBalancerListener(t.LoadBalancerID, int(t.Port))
}

func (t *Listener) findVgroup(key string) string {
	for _, v := range *t.VGroups {
		if v.NamedKey.Key() == key {
			glog.Infof("found: key=%s, groupid=%s, try use vserver group mode.", key, v.VGroupId)
			return v.VGroupId
		}
	}
	glog.Infof("find: vserver group [%s] does not found. use default backend group.", key)
	return STRINGS_EMPTY
}

var STRINGS_EMPTY = ""

type Listeners []*Listener

// 1. First, build listeners config from aliyun API output.
// 2. Second, build listeners from k8s service object.
// 3. Third, Merge the up two listeners to decide whether add/update/remove is needed.
// 4. Do update.  Clean unused vserver group.
func EnsureListeners(client ClientSLBSDK,
	service *v1.Service,
	lb *slb.LoadBalancerType, vgs *vgroups) error {

	local, err := buildListenersFromService(service, lb, client, vgs)
	if err != nil {
		return fmt.Errorf("build listener from service: %s", err.Error())
	}

	// Merge listeners generate an listener list to be updated/deleted/added.
	updates, err := mergeListeners(service, local, buildListenersFromAPI(service, lb, client, vgs))
	if err != nil {
		return fmt.Errorf("merge listener: %s", err.Error())
	}
	glog.Infof("ensure listener: %d updates for %s", len(updates), lb.LoadBalancerId)
	// do update/add/delete
	for _, up := range updates {
		err := up.Apply()
		if err != nil {
			return fmt.Errorf("ensure listener: %s", err.Error())
		}
	}

	return CleanUPVGroupMerged(service, lb, client, vgs)
}

// Only listener which owned by my service was deleted.
func EnsureListenersDeleted(client ClientSLBSDK,
	service *v1.Service,
	lb *slb.LoadBalancerType, vgs *vgroups) error {

	local, err := buildListenersFromService(service, lb, client, vgs)
	if err != nil {
		return fmt.Errorf("build listener from service: %s", err.Error())
	}
	remote := buildListenersFromAPI(service, lb, client, vgs)

	for _, loc := range local {
		for _, rem := range remote {
			if !isManagedByMyService(service, rem) {
				continue
			}
			if loc.Port == rem.Port {
				err := loc.Remove()
				if err != nil {
					return fmt.Errorf("ensure listener: %s", err.Error())
				}
			}
		}
	}

	return CleanUPVGroupDirect(vgs)
}

func isManagedByMyService(svc *v1.Service, remote *Listener) bool {
	if remote.Name == STRINGS_EMPTY ||
		remote.Name == "-" {
		// Assume listener without a name or named '-' to be k8s managed listener.
		// This is normally for service update. make a transform
		return true
	}
	return remote.NamedKey != nil &&
		remote.NamedKey.ServiceURI() == URIfromService(svc)
}

func isUserManagedListener(remote *Listener) bool {
	return remote.NamedKey == nil && remote.Name != STRINGS_EMPTY && remote.Name != "-"
}

// 1. We update listener to the latest version2 when updation is needed.
// 2. We assume listener with an empty name to be legacy version.
// 3. We assume listener with an arbitary name to be user managed listener.
// 4. LoadBalancer created by kubernetes is not allowed to be reused.
func mergeListeners(svc *v1.Service, service, console Listeners) (Listeners, error) {
	override := isOverrideListeners(serviceAnnotation(svc, ServiceAnnotationLoadBalancerOverrideListener))
	addition, updation, deletion := Listeners{}, Listeners{}, Listeners{}
	// For updations and deletions
	for _, remote := range console {

		skipDeletion, overridePort := false, false
		for _, local := range service {
			if remote.Port == local.Port {
				// port matched. that is where the conflict case begin.
				if isUserManagedListener(remote) {
					if !override {
						// port conflict with user managed port.
						return nil, fmt.Errorf("port matched, but conflict with user managed port. "+
							"Port:%d, ListenerName:%s, Service: %s. Protocol:[source:%s dst:%s]",
							remote.Port, remote.Name, local.NamedKey.Key(), remote.TransforedProto, local.TransforedProto)
					} else {
						overridePort = true
						break
					}
				}
				if !isManagedByMyService(svc, remote) {
					if !override {
						// port conflict with other service
						return nil, fmt.Errorf("port matched. but not managed by this service[%s]. "+
							"conflict with service[%s]", local.NamedKey.Key(), remote.NamedKey.Key())
					} else {
						overridePort = true
						break
					}
				}
				if remote.TransforedProto == local.TransforedProto {
					// protocol matched. do update.
					local.Action = ACTION_UPDATE
					skipDeletion = true
					updation = append(updation, local)
				}
				// not match , do delete
				break
			}
		}

		if !skipDeletion {
			if overridePort {
				remote.Action = ACTION_DELETE
				deletion = append(deletion, remote)
			} else {
				// only delete listeners which managed by kubernetes and by my service.
				if isManagedByMyService(svc, remote) && !isUserManagedListener(remote) {
					remote.Action = ACTION_DELETE
					deletion = append(deletion, remote)
				} else {
					glog.Infof("managed by other service or user managed listener, skip delete. [%s]", remote.Name)
				}
			}
		}
	}
	// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
	// For additions
	for _, local := range service {
		found := false
		for _, remote := range console {
			if remote.Port == local.Port {
				if !isManagedByMyService(svc, remote) {
					if override {
						// do add
						break
					} else {
						return addition, fmt.Errorf("error: service[%s] trying "+
							"to declare a port belongs to somebody else [%s]", local.NamedKey.Key(), remote.Name)
					}
				}
				if remote.TransforedProto != local.TransforedProto {
					// do add
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
		}
	}

	// Pls be careful of the sequence. deletion first,then addition, last updation
	return append(append(deletion, addition...), updation...), nil
}

// buildListenersFromService Build expected listeners
func buildListenersFromService(service *v1.Service,
	lb *slb.LoadBalancerType,
	client ClientSLBSDK, vgrps *vgroups) (Listeners, error) {
	listeners := Listeners{}
	for _, port := range service.Spec.Ports {
		proto, err := Protocol(serviceAnnotation(service, ServiceAnnotationLoadBalancerProtocolPort), port)
		if err != nil {
			return nil, err
		}
		n := Listener{
			NamedKey: &NamedKey{
				CID:         CLUSTER_ID,
				Namespace:   service.Namespace,
				ServiceName: service.Name,
				Port:        port.Port,
				Prefix:      DEFAULT_PREFIX},
			Port:            port.Port,
			NodePort:        port.NodePort,
			Proto:           string(port.Protocol),
			Service:         service,
			TransforedProto: proto,
			Client:          client,
			VGroups:         vgrps,
			LoadBalancerID:  lb.LoadBalancerId,
		}
		n.Name = n.NamedKey.Key()
		listeners = append(listeners, &n)
	}
	return listeners, nil
}

// buildListenersFromAPI Load current listeners
func buildListenersFromAPI(service *v1.Service,
	lb *slb.LoadBalancerType,
	client ClientSLBSDK, vgrps *vgroups) (listeners Listeners) {
	ports := lb.ListenerPortsAndProtocol.ListenerPortAndProtocol
	for _, port := range ports {
		key, err := LoadNamedKey(port.Description)
		if err != nil {
			glog.Warningf("alicloud: error parse listener description[%s]. %s", port.Description, err.Error())
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

func (t *tcp) Add() error {
	def, _ := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerTCPListener(
		&slb.CreateLoadBalancerTCPListenerArgs{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      int(t.Port),
			BackendServerPort: int(t.NodePort),
			//Health Check
			Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
			PersistenceTimeout: def.PersistenceTimeout,
			Description:        t.NamedKey.Key(),

			VServerGroupId:            t.findVgroup(t.NamedKey.Reference(t.NodePort)),
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

func (t *tcp) Update() error {
	def, request := ExtractAnnotationRequest(t.Service)

	response, err := t.Client.DescribeLoadBalancerTCPListenerAttribute(t.LoadBalancerID, int(t.Port))
	if err != nil {
		return fmt.Errorf("update tcp listener: %s", err.Error())
	}
	config := &slb.SetLoadBalancerTCPListenerAttributeArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		Description:       t.NamedKey.Key(),
		//Health Check
		Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
		PersistenceTimeout: response.PersistenceTimeout,
		VServerGroupId:     t.findVgroup(t.NamedKey.Reference(t.NodePort)),

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
			glog.V(2).Infof("TCP listener checker [bandwidth] changed, request=%d. response=%d", def.Bandwidth, response.Bandwidth)
		}
	*/

	// todo: perform healthcheck update.
	if def.HealthCheckType != response.HealthCheckType {
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
	if def.PersistenceTimeout != response.PersistenceTimeout {
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
		glog.V(2).Infof("tcp listener [BackendServerPort] changed, request=%d. response=%d, recreate.", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		return t.Client.CreateLoadBalancerTCPListener((*slb.CreateLoadBalancerTCPListenerArgs)(config))
	}
	if !needUpdate {
		glog.V(2).Infof("alicloud: tcp listener did not change, skip [update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	glog.V(2).Infof("TCP listener checker changed, request update listener attribute [%s]\n", t.LoadBalancerID)
	glog.V(5).Infof(PrettyJson(def))
	glog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerTCPListenerAttribute(config)
}

type udp struct{ *Listener }

func (t *udp) Describe() error {
	return fmt.Errorf("unimplemented")
}
func (t *udp) Add() error {
	def, _ := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerUDPListener(
		&slb.CreateLoadBalancerUDPListenerArgs{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      int(t.Port),
			BackendServerPort: int(t.NodePort),
			Description:       t.NamedKey.Key(),
			VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
			//Health Check
			Bandwidth:          DEFAULT_LISTENER_BANDWIDTH,
			PersistenceTimeout: def.PersistenceTimeout,

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

func (t *udp) Update() error {
	def, request := ExtractAnnotationRequest(t.Service)
	response, err := t.Client.DescribeLoadBalancerUDPListenerAttribute(t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	config := &slb.SetLoadBalancerUDPListenerAttributeArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		Description:       t.NamedKey.Key(),
		VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
		//Health Check
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
			glog.V(2).Infof("UDP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
	*/

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
	if def.PersistenceTimeout != response.PersistenceTimeout {
		needUpdate = true
		config.PersistenceTimeout = def.PersistenceTimeout
	}
	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		config.BackendServerPort = int(t.NodePort)
		glog.V(2).Infof("UDP listener checker [BackendServerPort] changed, request=%d. response=%d", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		return t.Client.CreateLoadBalancerUDPListener((*slb.CreateLoadBalancerUDPListenerArgs)(config))
	}

	if !needUpdate {
		glog.V(2).Infof("alicloud: udp listener did not change, skip [update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	glog.V(2).Infof("UDP listener checker changed, request recreate [%s]\n", t.LoadBalancerID)
	glog.V(5).Infof(PrettyJson(request))
	glog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerUDPListenerAttribute(config)
}

type http struct{ *Listener }

func (t *http) Describe() error {
	return fmt.Errorf("unimplemented")
}
func (t *http) Add() error {
	def, request := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerHTTPListener(
		&slb.CreateLoadBalancerHTTPListenerArgs{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      int(t.Port),
			BackendServerPort: int(t.NodePort),
			Description:       t.NamedKey.Key(),
			VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
			//Health Check
			Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
			StickySession:     def.StickySession,
			StickySessionType: def.StickySessionType,
			CookieTimeout:     def.CookieTimeout,
			Cookie:            def.Cookie,

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
		})
}

func (t *http) Update() error {

	def, request := ExtractAnnotationRequest(t.Service)
	response, err := t.Client.DescribeLoadBalancerHTTPListenerAttribute(t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	config := &slb.SetLoadBalancerHTTPListenerAttributeArgs{
		LoadBalancerId:    t.LoadBalancerID,
		ListenerPort:      int(t.Port),
		BackendServerPort: int(t.NodePort),
		//Health Check
		Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
		StickySession:     response.StickySession,
		StickySessionType: response.StickySessionType,
		CookieTimeout:     response.CookieTimeout,
		Cookie:            response.Cookie,
		Description:       t.NamedKey.Key(),
		VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),

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
	/*
		if request.Bandwidth != 0 &&
			request.Bandwidth != response.Bandwidth {
			needUpdate = true
			config.Bandwidth = request.Bandwidth
			glog.V(2).Infof("HTTP listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
	*/

	// todo: perform healthcheck update.
	if def.HealthCheck != response.HealthCheck {
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
	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		config.BackendServerPort = int(t.NodePort)
		glog.V(2).Infof("HTTP listener checker [BackendServerPort] changed, request=%d. response=%d", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		return t.Client.CreateLoadBalancerHTTPListener((*slb.CreateLoadBalancerHTTPListenerArgs)(config))
	}

	if !needUpdate {
		glog.V(2).Infof("alicloud: http listener did not change, skip [update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	glog.V(2).Infof("HTTP listener checker changed, request recreate [%s]\n", t.LoadBalancerID)
	glog.V(5).Infof(PrettyJson(request))
	glog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerHTTPListenerAttribute(config)
}

type https struct{ *Listener }

func (t *https) Describe() error {
	return fmt.Errorf("unimplemented")
}
func (t *https) Add() error {

	def, request := ExtractAnnotationRequest(t.Service)
	return t.Client.CreateLoadBalancerHTTPSListener(
		&slb.CreateLoadBalancerHTTPSListenerArgs{
			HTTPListenerType: slb.HTTPListenerType{
				LoadBalancerId:    t.LoadBalancerID,
				ListenerPort:      int(t.Port),
				BackendServerPort: int(t.NodePort),
				Description:       t.NamedKey.Key(),
				VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
				//Health Check
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

func (t *https) Update() error {
	def, request := ExtractAnnotationRequest(t.Service)
	response, err := t.Client.DescribeLoadBalancerHTTPSListenerAttribute(t.LoadBalancerID, int(t.Port))
	if err != nil {
		return err
	}
	config := &slb.SetLoadBalancerHTTPSListenerAttributeArgs{
		HTTPListenerType: slb.HTTPListenerType{
			LoadBalancerId:    t.LoadBalancerID,
			ListenerPort:      response.ListenerPort,
			BackendServerPort: response.BackendServerPort,
			Description:       t.NamedKey.Key(),
			VServerGroupId:    t.findVgroup(t.NamedKey.Reference(t.NodePort)),
			//Health Check
			HealthCheck:       response.HealthCheck,
			Bandwidth:         DEFAULT_LISTENER_BANDWIDTH,
			StickySession:     response.StickySession,
			StickySessionType: response.StickySessionType,
			CookieTimeout:     response.CookieTimeout,
			Cookie:            response.Cookie,

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
			glog.V(2).Infof("HTTPS listener checker [bandwidth] changed, request=%d. response=%d", request.Bandwidth, response.Bandwidth)
		}
	*/
	// todo: perform healthcheck update.
	if def.HealthCheck != response.HealthCheck {
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
	// backend server port has changed.
	if int(t.NodePort) != response.BackendServerPort {
		needUpdate = true
		config.BackendServerPort = int(t.NodePort)
		glog.V(2).Infof("HTTPS listener checker [BackendServerPort] changed, request=%d. response=%d", t.NodePort, response.BackendServerPort)
		err := t.Client.DeleteLoadBalancerListener(t.LoadBalancerID, int(t.Port))
		if err != nil {
			return err
		}
		return t.Client.CreateLoadBalancerHTTPSListener((*slb.CreateLoadBalancerHTTPSListenerArgs)(config))
	}

	if !needUpdate {
		glog.V(2).Infof("alicloud: https listener did not change, skip [update], port=[%d], nodeport=[%d]\n", t.Port, t.NodePort)
		// no recreate needed.  skip
		return nil
	}
	glog.V(2).Infof("HTTPS listener checker changed, request recreate [%s]\n", t.LoadBalancerID)
	glog.V(5).Infof(PrettyJson(request))
	glog.V(5).Infof(PrettyJson(response))
	return t.Client.SetLoadBalancerHTTPSListenerAttribute(config)
}
