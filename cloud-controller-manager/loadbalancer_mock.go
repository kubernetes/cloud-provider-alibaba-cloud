package alicloud

import (
	"encoding/json"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/slb"
	"github.com/denverdino/aliyungo/util"
	"reflect"
	"strings"
	"sync"
	"time"
)

func WithNewLoadBalancerStore() CloudDataMock {
	return func() {
		LOADBALANCER = LBStore{}
	}
}

func WithLoadBalancer() CloudDataMock {
	return func() {
		LOADBALANCER.loadbalancer.Store(
			LOADBALANCER_ID,
			slb.LoadBalancerType{
				LoadBalancerId:     LOADBALANCER_ID,
				LoadBalancerName:   LOADBALANCER_NAME,
				RegionId:           REGION,
				LoadBalancerSpec:   LOADBALANCER_SPEC,
				Bandwidth:          100,
				InternetChargeType: slb.PayByBandwidth,
				Address:            LOADBALANCER_ADDRESS,
				VSwitchId:          VSWITCH_ID,
				VpcId:              VPCID,
				MasterZoneId:       fmt.Sprintf("%s-a", REGION),
				SlaveZoneId:        fmt.Sprintf("%s-b", REGION),
			},
		)
		listener := &slb.DescribeLoadBalancerTCPListenerAttributeResponse{
			DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{},
			TCPListenerType: slb.TCPListenerType{
				LoadBalancerId:    LOADBALANCER_ID,
				ListenerPort:      80,
				BackendServerPort: 32999,
				Bandwidth:         50,
				Description:       "",
				VServerGroupId:    "",
				VServerGroup:      "",
				HealthCheck:       "on",
				HealthCheckURI:    "",
				//HealthCheckConnectPort:    args.HealthCheckConnectPort,
				//HealthCheckConnectTimeout: args.HealthCheckConnectTimeout,
				//HealthCheckDomain:         args.HealthCheckDomain,
				//HealthCheckHttpCode:       args.HealthCheckHttpCode,
				//HealthCheckInterval:       args.HealthCheckInterval,
				//HealthCheckType:           args.HealthCheckType,
				//HealthyThreshold:          args.HealthyThreshold,
				//UnhealthyThreshold:        args.UnhealthyThreshold,
			},
		}
		LOADBALANCER.listeners.Store(listenerKey(LOADBALANCER_ID, 80), listener)
	}
}

type mockClientSLB struct {
	describeLoadBalancers          func(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error)
	createLoadBalancer             func(args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error)
	deleteLoadBalancer             func(loadBalancerId string) (err error)
	modifyLoadBalancerInternetSpec func(args *slb.ModifyLoadBalancerInternetSpecArgs) (err error)
	modifyLoadBalancerInstanceSpec func(args *slb.ModifyLoadBalancerInstanceSpecArgs) (err error)
	describeLoadBalancerAttribute  func(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error)
	removeBackendServers           func(loadBalancerId string, backendServers []string) (result []slb.BackendServerType, err error)
	addBackendServers              func(loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error)

	stopLoadBalancerListener                   func(loadBalancerId string, port int) (err error)
	startLoadBalancerListener                  func(loadBalancerId string, port int) (err error)
	createLoadBalancerTCPListener              func(args *slb.CreateLoadBalancerTCPListenerArgs) (err error)
	createLoadBalancerUDPListener              func(args *slb.CreateLoadBalancerUDPListenerArgs) (err error)
	deleteLoadBalancerListener                 func(loadBalancerId string, port int) (err error)
	createLoadBalancerHTTPSListener            func(args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error)
	createLoadBalancerHTTPListener             func(args *slb.CreateLoadBalancerHTTPListenerArgs) (err error)
	describeLoadBalancerHTTPSListenerAttribute func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error)
	describeLoadBalancerTCPListenerAttribute   func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error)
	describeLoadBalancerUDPListenerAttribute   func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error)
	describeLoadBalancerHTTPListenerAttribute  func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error)

	setLoadBalancerHTTPListenerAttribute  func(args *slb.SetLoadBalancerHTTPListenerAttributeArgs) (err error)
	setLoadBalancerHTTPSListenerAttribute func(args *slb.SetLoadBalancerHTTPSListenerAttributeArgs) (err error)
	setLoadBalancerTCPListenerAttribute   func(args *slb.SetLoadBalancerTCPListenerAttributeArgs) (err error)
	setLoadBalancerUDPListenerAttribute   func(args *slb.SetLoadBalancerUDPListenerAttributeArgs) (err error)
	removeTags                            func(args *slb.RemoveTagsArgs) error
	describeTags                          func(args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error)
	addTags                               func(args *slb.AddTagsArgs) error

	createVServerGroup               func(args *slb.CreateVServerGroupArgs) (response *slb.CreateVServerGroupResponse, err error)
	describeVServerGroups            func(args *slb.DescribeVServerGroupsArgs) (response *slb.DescribeVServerGroupsResponse, err error)
	deleteVServerGroup               func(args *slb.DeleteVServerGroupArgs) (response *slb.DeleteVServerGroupResponse, err error)
	setVServerGroupAttribute         func(args *slb.SetVServerGroupAttributeArgs) (response *slb.SetVServerGroupAttributeResponse, err error)
	describeVServerGroupAttribute    func(args *slb.DescribeVServerGroupAttributeArgs) (response *slb.DescribeVServerGroupAttributeResponse, err error)
	modifyVServerGroupBackendServers func(args *slb.ModifyVServerGroupBackendServersArgs) (response *slb.ModifyVServerGroupBackendServersResponse, err error)
	addVServerGroupBackendServers    func(args *slb.AddVServerGroupBackendServersArgs) (response *slb.AddVServerGroupBackendServersResponse, err error)
	removeVServerGroupBackendServers func(args *slb.RemoveVServerGroupBackendServersArgs) (response *slb.RemoveVServerGroupBackendServersResponse, err error)
}

type LBStore struct {
	loadbalancer sync.Map
	listeners    sync.Map
	tags         sync.Map
	vgroups      sync.Map
}

// LOADBALANCER slb cloud mock storage
// string: *slb.LoadBalancerType{}
var LOADBALANCER = LBStore{}

func newBaseLoadbalancer() []*slb.LoadBalancerType {
	return []*slb.LoadBalancerType{
		{
			LoadBalancerId:     LOADBALANCER_ID,
			LoadBalancerName:   LOADBALANCER_NAME,
			LoadBalancerStatus: "active",
			Address:            LOADBALANCER_ADDRESS,
			RegionId:           REGION,
			RegionIdAlias:      string(REGION),
			AddressType:        "internet",
			VSwitchId:          "",
			VpcId:              "",
			NetworkType:        LOADBALANCER_NETWORK_TYPE,
			Bandwidth:          0,
			InternetChargeType: "4",
			CreateTime:         "2018-03-14T17:16Z",
			CreateTimeStamp:    util.NewISO6801Time(time.Now()),
			LoadBalancerSpec:   slb.S1Small,
		},
	}
}

func (c *mockClientSLB) DescribeLoadBalancers(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error) {
	if c.describeLoadBalancers != nil {
		return c.describeLoadBalancers(args)
	}
	var results []slb.LoadBalancerType
	LOADBALANCER.loadbalancer.Range(
		func(key, value interface{}) bool {

			v, ok := value.(slb.LoadBalancerType)
			if !ok {
				fmt.Printf("API: DescribeLoadBalancers, "+
					"unexpected type %s, not slb.LoadBalancerType", reflect.TypeOf(value))
				return true
			}
			if args.LoadBalancerId != "" &&
				v.LoadBalancerId != args.LoadBalancerId {
				// continue next
				return true
			}
			if args.LoadBalancerName != "" &&
				v.LoadBalancerName != args.LoadBalancerName {
				// continue next
				return true
			}
			if args.RegionId != "" &&
				args.RegionId != v.RegionId {
				// continue next
				return true
			}
			if args.Tags != "" {
				bytag := &slb.DescribeTagsArgs{
					LoadBalancerID: args.LoadBalancerId,
					Tags:           args.Tags,
				}
				tag, _, _ := c.DescribeTags(bytag)
				if len(tag) <= 0 {
					return true
				}
			}
			results = append(results, v)
			return true
		},
	)
	return results, nil
}

func (c *mockClientSLB) StopLoadBalancerListener(loadBalancerId string, port int) (err error) {
	if c.stopLoadBalancerListener != nil {
		return c.stopLoadBalancerListener(loadBalancerId, port)
	}
	key := listenerKey(loadBalancerId, port)
	listenerObj, ok := LOADBALANCER.listeners.Load(key)
	if !ok || listenerObj == nil {
		return fmt.Errorf("not found listener: %s %d ", loadBalancerId, port)
	}
	switch listenerObj.(type) {
	case *slb.DescribeLoadBalancerTCPListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerTCPListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Stopped
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	case *slb.DescribeLoadBalancerUDPListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerUDPListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Stopped
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	case *slb.DescribeLoadBalancerHTTPListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerHTTPListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Stopped
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	case *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerHTTPSListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Stopped
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	default:
		return fmt.Errorf("StopLoadBalancerListener() listener type error")
	}

	// return nil indicate no stop success
	return nil
}

func (c *mockClientSLB) CreateLoadBalancer(args *slb.CreateLoadBalancerArgs) (response *slb.CreateLoadBalancerResponse, err error) {
	if c.createLoadBalancer != nil {
		return c.createLoadBalancer(args)
	}
	if args.LoadBalancerName == "" {
		return nil, fmt.Errorf("slb name must not be empty")
	}
	addrtype := slb.InternetAddressType
	if args.AddressType != "" {
		addrtype = args.AddressType
	}
	ipver := slb.IPv4
	if args.AddressIPVersion != "" {
		ipver = args.AddressIPVersion
	}
	ins := slb.LoadBalancerType{
		LoadBalancerId:     newid(),
		LoadBalancerName:   args.LoadBalancerName,
		RegionId:           args.RegionId,
		LoadBalancerSpec:   args.LoadBalancerSpec,
		Bandwidth:          args.Bandwidth,
		InternetChargeType: args.InternetChargeType,
		Address:            LOADBALANCER_ADDRESS,
		AddressType:        addrtype,
		VSwitchId:          args.VSwitchId,
		VpcId:              VPCID,
		AddressIPVersion:   ipver,
		MasterZoneId:       args.MasterZoneId,
		SlaveZoneId:        args.SlaveZoneId,
	}
	LOADBALANCER.loadbalancer.Store(ins.LoadBalancerId, ins)
	return &slb.CreateLoadBalancerResponse{
		LoadBalancerId:   ins.LoadBalancerId,
		Address:          ins.Address,
		NetworkType:      string(ins.InternetChargeType),
		LoadBalancerName: ins.LoadBalancerName,
	}, nil
}

func (c *mockClientSLB) DeleteLoadBalancer(loadBalancerId string) (err error) {
	if c.deleteLoadBalancer != nil {
		return c.deleteLoadBalancer(loadBalancerId)
	}
	LOADBALANCER.loadbalancer.Delete(loadBalancerId)
	return nil
}

func (c *mockClientSLB) ModifyLoadBalancerInternetSpec(args *slb.ModifyLoadBalancerInternetSpecArgs) (err error) {
	if c.modifyLoadBalancerInternetSpec != nil {
		return c.modifyLoadBalancerInternetSpec(args)
	}
	if args.LoadBalancerId == "" {
		return fmt.Errorf("loadbalancer id must not be empty")
	}
	v, ok := LOADBALANCER.loadbalancer.Load(args.LoadBalancerId)
	if !ok {
		return fmt.Errorf("loadbalancer not found by id %s", args.LoadBalancerId)
	}
	ins, ok := v.(slb.LoadBalancerType)
	if !ok {
		return fmt.Errorf("not slb.LoadBalancerType")
	}
	if args.Bandwidth != 0 {
		ins.Bandwidth = args.Bandwidth
	}
	if args.InternetChargeType != "" {
		ins.InternetChargeType = args.InternetChargeType
	}
	LOADBALANCER.loadbalancer.Store(ins.LoadBalancerId, ins)
	return nil
}

func (c *mockClientSLB) ModifyLoadBalancerInstanceSpec(args *slb.ModifyLoadBalancerInstanceSpecArgs) (err error) {
	if c.modifyLoadBalancerInstanceSpec != nil {
		return c.modifyLoadBalancerInstanceSpec(args)
	}
	if args.LoadBalancerId == "" {
		return fmt.Errorf("loadbalancer id must not be empty")
	}
	v, ok := LOADBALANCER.loadbalancer.Load(args.LoadBalancerId)
	if !ok {
		return fmt.Errorf("loadbalancer not found by id %s", args.LoadBalancerId)
	}
	ins, ok := v.(slb.LoadBalancerType)
	if !ok {
		return fmt.Errorf("not slb.LoadBalancerType")
	}
	if args.LoadBalancerSpec != "" {
		ins.LoadBalancerSpec = args.LoadBalancerSpec
	}
	LOADBALANCER.loadbalancer.Store(ins.LoadBalancerId, ins)
	return nil
}

func (c *mockClientSLB) DescribeLoadBalancerAttribute(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error) {
	if c.describeLoadBalancerAttribute != nil {
		return c.describeLoadBalancerAttribute(loadBalancerId)
	}

	if loadBalancerId == "" {
		return nil, fmt.Errorf("loadbalancer id must not be empty")
	}

	v, ok := LOADBALANCER.loadbalancer.Load(loadBalancerId)
	if !ok {
		return nil, fmt.Errorf("loadbalancer not found by id %s", loadBalancerId)
	}
	ins, ok := v.(slb.LoadBalancerType)
	if !ok {
		return nil, fmt.Errorf("not slb.LoadBalancerType")
	}
	var ports []int
	var pproto []slb.ListenerPortAndProtocolType
	LOADBALANCER.listeners.Range(
		func(key, value interface{}) bool {
			k := key.(string)
			if strings.Contains(k, loadBalancerId) {
				port := 0
				descrip := ""
				proto := ""
				switch value.(type) {
				case *slb.DescribeLoadBalancerTCPListenerAttributeResponse:
					v := value.(*slb.DescribeLoadBalancerTCPListenerAttributeResponse)
					port = v.ListenerPort
					descrip = v.Description
					proto = "tcp"
				case *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse:
					v := value.(*slb.DescribeLoadBalancerHTTPSListenerAttributeResponse)
					port = v.ListenerPort
					descrip = v.Description
					proto = "https"
				case *slb.DescribeLoadBalancerUDPListenerAttributeResponse:
					v := value.(*slb.DescribeLoadBalancerUDPListenerAttributeResponse)
					port = v.ListenerPort
					descrip = v.Description
					proto = "udp"
				case *slb.DescribeLoadBalancerHTTPListenerAttributeResponse:
					v := value.(*slb.DescribeLoadBalancerHTTPListenerAttributeResponse)
					port = v.ListenerPort
					descrip = v.Description
					proto = "http"
				}
				ports = append(ports, port)
				pp := slb.ListenerPortAndProtocolType{
					ListenerPort:     port,
					ListenerProtocol: proto,
					Description:      descrip,
				}
				pproto = append(pproto, pp)
			}
			return true
		},
	)
	ins.ListenerPorts.ListenerPort = ports
	ins.ListenerPortsAndProtocol.ListenerPortAndProtocol = pproto
	return &ins, nil
}

func (c *mockClientSLB) RemoveBackendServers(loadBalancerId string, backendServers []string) (result []slb.BackendServerType, err error) {
	if c.removeBackendServers != nil {
		return c.removeBackendServers(loadBalancerId, backendServers)
	}
	v, ok := LOADBALANCER.loadbalancer.Load(loadBalancerId)
	if !ok {
		return nil, fmt.Errorf("loadbalancer not found %s addbackend", loadBalancerId)
	}

	ins, ok := v.(slb.LoadBalancerType)
	if !ok {
		return nil, fmt.Errorf("not slb.BackendServerType")
	}

	for _, bc := range ins.BackendServers.BackendServer {
		found := false
		for _, del := range backendServers {
			if bc.ServerId == del {
				found = true
			}
		}
		if !found {
			ins.BackendServers.BackendServer = append(ins.BackendServers.BackendServer, bc)
		}
	}
	LOADBALANCER.loadbalancer.Store(loadBalancerId, ins)
	return backend(backendServers), nil
}

func backend(bc []string) []slb.BackendServerType {
	var result []slb.BackendServerType
	for _, back := range bc {
		server := slb.BackendServerType{
			Weight:   100,
			Type:     "ecs",
			ServerId: back,
		}
		result = append(result, server)
	}
	return result
}

func (c *mockClientSLB) AddBackendServers(loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error) {
	if c.addBackendServers != nil {
		return c.addBackendServers(loadBalancerId, backendServers)
	}

	v, ok := LOADBALANCER.loadbalancer.Load(loadBalancerId)
	if !ok {
		return nil, fmt.Errorf("loadbalancer not found %s addbackend", loadBalancerId)
	}

	ins, ok := v.(slb.LoadBalancerType)
	if !ok {
		return nil, fmt.Errorf("not slb.BackendServerType")
	}

	for _, add := range backendServers {
		found := false
		for _, bc := range ins.BackendServers.BackendServer {
			if bc.ServerId == add.ServerId {
				found = true
			}
		}
		if !found {
			fmt.Printf("AddBackend: %v\n", add)
			ins.BackendServers.BackendServer = append(ins.BackendServers.BackendServer, add)
		}
	}
	LOADBALANCER.loadbalancer.Store(loadBalancerId, ins)

	return nil, nil
}

func listenerKey(id string, port int) string {
	return fmt.Sprintf("%s/%d", id, port)
}

func (c *mockClientSLB) StartLoadBalancerListener(loadBalancerId string, port int) (err error) {
	if c.startLoadBalancerListener != nil {
		return c.startLoadBalancerListener(loadBalancerId, port)
	}
	key := listenerKey(loadBalancerId, port)
	listenerObj, ok := LOADBALANCER.listeners.Load(key)
	if !ok {
		return fmt.Errorf("not found listener: %s %d ", loadBalancerId, port)
	}
	switch listenerObj.(type) {
	case *slb.DescribeLoadBalancerTCPListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerTCPListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Running
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	case *slb.DescribeLoadBalancerUDPListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerUDPListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Running
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	case *slb.DescribeLoadBalancerHTTPListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerHTTPListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Running
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	case *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse:
		if listener, ok := listenerObj.(*slb.DescribeLoadBalancerHTTPSListenerAttributeResponse); ok {
			listener.DescribeLoadBalancerListenerAttributeResponse.Status = slb.Running
			LOADBALANCER.listeners.Store(key, listener)
		}
		break
	default:
		return fmt.Errorf("StartLoadBalancerListener() listener type error")
	}
	return nil
}
func (c *mockClientSLB) CreateLoadBalancerTCPListener(args *slb.CreateLoadBalancerTCPListenerArgs) (err error) {
	if c.createLoadBalancerTCPListener != nil {
		return c.createLoadBalancerTCPListener(args)
	}
	listener := &slb.DescribeLoadBalancerTCPListenerAttributeResponse{
		DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{},
		TCPListenerType: slb.TCPListenerType{
			LoadBalancerId:            args.LoadBalancerId,
			ListenerPort:              args.ListenerPort,
			BackendServerPort:         args.BackendServerPort,
			Bandwidth:                 args.Bandwidth,
			PersistenceTimeout:        args.PersistenceTimeout,
			Description:               args.Description,
			VServerGroupId:            args.VServerGroupId,
			VServerGroup:              args.VServerGroup,
			HealthCheck:               args.HealthCheck,
			HealthCheckURI:            args.HealthCheckURI,
			HealthCheckConnectPort:    args.HealthCheckConnectPort,
			HealthCheckConnectTimeout: args.HealthCheckConnectTimeout,
			HealthCheckDomain:         args.HealthCheckDomain,
			HealthCheckHttpCode:       args.HealthCheckHttpCode,
			HealthCheckInterval:       args.HealthCheckInterval,
			HealthCheckType:           args.HealthCheckType,
			HealthyThreshold:          args.HealthyThreshold,
			UnhealthyThreshold:        args.UnhealthyThreshold,
			AclType:                   args.AclType,
			AclId:                     args.AclId,
			AclStatus:                 args.AclStatus,
			Scheduler:                 args.Scheduler,
		},
	}
	key := listenerKey(args.LoadBalancerId, args.ListenerPort)
	_, ok := LOADBALANCER.listeners.Load(key)
	if ok {
		return fmt.Errorf("tcp listener exist %d", args.ListenerPort)
	}
	LOADBALANCER.listeners.Store(key, listener)
	return nil
}

func (c *mockClientSLB) CreateLoadBalancerUDPListener(args *slb.CreateLoadBalancerUDPListenerArgs) (err error) {
	if c.createLoadBalancerUDPListener != nil {
		return c.createLoadBalancerUDPListener(args)
	}

	listener := &slb.DescribeLoadBalancerUDPListenerAttributeResponse{
		DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{},
		UDPListenerType: slb.UDPListenerType{
			LoadBalancerId:            args.LoadBalancerId,
			ListenerPort:              args.ListenerPort,
			BackendServerPort:         args.BackendServerPort,
			Bandwidth:                 args.Bandwidth,
			PersistenceTimeout:        args.PersistenceTimeout,
			Description:               args.Description,
			VServerGroupId:            args.VServerGroupId,
			VServerGroup:              args.VServerGroup,
			HealthCheck:               args.HealthCheck,
			HealthCheckConnectPort:    args.HealthCheckConnectPort,
			HealthCheckConnectTimeout: args.HealthCheckConnectTimeout,
			HealthCheckInterval:       args.HealthCheckInterval,
			HealthyThreshold:          args.HealthyThreshold,
			UnhealthyThreshold:        args.UnhealthyThreshold,
			AclType:                   args.AclType,
			AclId:                     args.AclId,
			AclStatus:                 args.AclStatus,
			Scheduler:                 args.Scheduler,
		},
	}
	key := listenerKey(args.LoadBalancerId, args.ListenerPort)
	_, ok := LOADBALANCER.listeners.Load(key)
	if ok {
		return fmt.Errorf("listener exist %d", args.ListenerPort)
	}
	LOADBALANCER.listeners.Store(key, listener)
	return nil
}
func (c *mockClientSLB) DeleteLoadBalancerListener(loadBalancerId string, port int) (err error) {
	if c.deleteLoadBalancerListener != nil {
		return c.deleteLoadBalancerListener(loadBalancerId, port)
	}
	LOADBALANCER.listeners.Delete(listenerKey(loadBalancerId, port))
	return nil
}
func (c *mockClientSLB) CreateLoadBalancerHTTPSListener(args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error) {
	if c.createLoadBalancerHTTPSListener != nil {
		return c.createLoadBalancerHTTPSListener(args)
	}

	listener := &slb.DescribeLoadBalancerHTTPSListenerAttributeResponse{
		DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{},
		HTTPSListenerType: slb.HTTPSListenerType{
			HTTPListenerType: slb.HTTPListenerType{
				LoadBalancerId:         args.LoadBalancerId,
				ListenerPort:           args.ListenerPort,
				BackendServerPort:      args.BackendServerPort,
				Bandwidth:              args.Bandwidth,
				Description:            args.Description,
				VServerGroupId:         args.VServerGroupId,
				VServerGroup:           args.VServerGroup,
				StickySession:          args.StickySession,
				StickySessionType:      args.StickySessionType,
				Cookie:                 args.Cookie,
				CookieTimeout:          args.CookieTimeout,
				HealthCheckTimeout:     args.HealthCheckTimeout,
				HealthCheck:            args.HealthCheck,
				HealthCheckURI:         args.HealthCheckURI,
				HealthCheckConnectPort: args.HealthCheckConnectPort,
				HealthCheckDomain:      args.HealthCheckDomain,
				HealthCheckHttpCode:    args.HealthCheckHttpCode,
				HealthCheckInterval:    args.HealthCheckInterval,
				HealthyThreshold:       args.HealthyThreshold,
				UnhealthyThreshold:     args.UnhealthyThreshold,
				AclType:                args.AclType,
				AclId:                  args.AclId,
				AclStatus:              args.AclStatus,
				Scheduler:              args.Scheduler,
			},
			ServerCertificateId: args.ServerCertificateId,
		},
	}
	key := listenerKey(args.LoadBalancerId, args.ListenerPort)
	_, ok := LOADBALANCER.listeners.Load(key)
	if ok {
		return fmt.Errorf("https listener exist %d", args.ListenerPort)
	}
	LOADBALANCER.listeners.Store(key, listener)

	return nil
}
func (c *mockClientSLB) CreateLoadBalancerHTTPListener(args *slb.CreateLoadBalancerHTTPListenerArgs) (err error) {
	if c.createLoadBalancerHTTPListener != nil {
		return c.createLoadBalancerHTTPListener(args)
	}
	listener := &slb.DescribeLoadBalancerHTTPListenerAttributeResponse{
		DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{},
		HTTPListenerType: slb.HTTPListenerType{
			LoadBalancerId:         args.LoadBalancerId,
			ListenerPort:           args.ListenerPort,
			BackendServerPort:      args.BackendServerPort,
			Bandwidth:              args.Bandwidth,
			Description:            args.Description,
			VServerGroupId:         args.VServerGroupId,
			VServerGroup:           args.VServerGroup,
			StickySession:          args.StickySession,
			StickySessionType:      args.StickySessionType,
			Cookie:                 args.Cookie,
			CookieTimeout:          args.CookieTimeout,
			HealthCheckTimeout:     args.HealthCheckTimeout,
			HealthCheck:            args.HealthCheck,
			HealthCheckURI:         args.HealthCheckURI,
			HealthCheckConnectPort: args.HealthCheckConnectPort,
			HealthCheckDomain:      args.HealthCheckDomain,
			HealthCheckHttpCode:    args.HealthCheckHttpCode,
			HealthCheckInterval:    args.HealthCheckInterval,
			HealthyThreshold:       args.HealthyThreshold,
			UnhealthyThreshold:     args.UnhealthyThreshold,
			AclType:                args.AclType,
			AclId:                  args.AclId,
			AclStatus:              args.AclStatus,
			Scheduler:              args.Scheduler,
		},
	}
	key := listenerKey(args.LoadBalancerId, args.ListenerPort)
	_, ok := LOADBALANCER.listeners.Load(key)
	if ok {
		return fmt.Errorf("http listener exist %d", args.ListenerPort)
	}
	LOADBALANCER.listeners.Store(key, listener)
	return nil
}
func (c *mockClientSLB) DescribeLoadBalancerHTTPSListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error) {
	if c.describeLoadBalancerHTTPSListenerAttribute != nil {
		return c.describeLoadBalancerHTTPSListenerAttribute(loadBalancerId, port)
	}
	v, ok := LOADBALANCER.listeners.Load(listenerKey(loadBalancerId, port))
	if !ok {
		fmt.Printf("listener not found, %s, %d\n", loadBalancerId, port)
		return nil, nil
	}
	result, ok := v.(*slb.DescribeLoadBalancerHTTPSListenerAttributeResponse)
	if !ok {
		return nil, fmt.Errorf("not type HTTPS listener. %s", reflect.TypeOf(v))
	}
	return result, nil
}

func (c *mockClientSLB) DescribeLoadBalancerTCPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {
	if c.describeLoadBalancerTCPListenerAttribute != nil {
		return c.describeLoadBalancerTCPListenerAttribute(loadBalancerId, port)
	}
	v, ok := LOADBALANCER.listeners.Load(listenerKey(loadBalancerId, port))
	if !ok {
		fmt.Printf("listener not found, %s, %d\n", loadBalancerId, port)
		return nil, nil
	}
	result, ok := v.(*slb.DescribeLoadBalancerTCPListenerAttributeResponse)
	if !ok {
		return nil, fmt.Errorf("not type TCP listener. %s", reflect.TypeOf(v))
	}
	return result, nil
}

func (c *mockClientSLB) DescribeLoadBalancerUDPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error) {
	if c.describeLoadBalancerUDPListenerAttribute != nil {
		return c.describeLoadBalancerUDPListenerAttribute(loadBalancerId, port)
	}
	v, ok := LOADBALANCER.listeners.Load(listenerKey(loadBalancerId, port))
	if !ok {
		fmt.Printf("listener not found, %s, %d\n", loadBalancerId, port)
		return nil, nil
	}
	result, ok := v.(*slb.DescribeLoadBalancerUDPListenerAttributeResponse)
	if !ok {
		return nil, fmt.Errorf("not type UDP listener. %s", reflect.TypeOf(v))
	}
	return result, nil
}

func (c *mockClientSLB) DescribeLoadBalancerHTTPListenerAttribute(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error) {
	if c.describeLoadBalancerHTTPListenerAttribute != nil {
		return c.describeLoadBalancerHTTPListenerAttribute(loadBalancerId, port)
	}
	v, ok := LOADBALANCER.listeners.Load(listenerKey(loadBalancerId, port))
	if !ok {
		fmt.Printf("listener not found, %s, %d\n", loadBalancerId, port)
		return nil, nil
	}
	result, ok := v.(*slb.DescribeLoadBalancerHTTPListenerAttributeResponse)
	if !ok {
		return nil, fmt.Errorf("not type HTTP listener. %s", reflect.TypeOf(v))
	}
	return result, nil
}

func (c *mockClientSLB) SetLoadBalancerHTTPListenerAttribute(args *slb.SetLoadBalancerHTTPListenerAttributeArgs) (err error) {
	if c.setLoadBalancerHTTPListenerAttribute != nil {
		return c.setLoadBalancerHTTPListenerAttribute(args)
	}
	lb, err := c.DescribeLoadBalancerHTTPListenerAttribute(args.LoadBalancerId, args.ListenerPort)
	if err != nil {
		return err
	}
	lb.CookieTimeout = args.CookieTimeout
	lb.Cookie = args.Cookie
	lb.StickySessionType = args.StickySessionType
	lb.StickySession = args.StickySession
	lb.HealthCheckTimeout = args.HealthCheckTimeout
	lb.HealthCheck = args.HealthCheck
	lb.HealthyThreshold = args.HealthyThreshold
	lb.HealthCheckInterval = args.HealthCheckInterval
	lb.HealthCheckHttpCode = args.HealthCheckHttpCode
	lb.HealthCheckDomain = args.HealthCheckDomain
	lb.HealthCheckConnectPort = args.HealthCheckConnectPort
	lb.HealthCheckURI = args.HealthCheckURI
	lb.UnhealthyThreshold = args.UnhealthyThreshold
	lb.ListenerPort = args.ListenerPort
	lb.VServerGroup = args.VServerGroup
	lb.VServerGroupId = args.VServerGroupId
	lb.Description = args.Description
	lb.Bandwidth = args.Bandwidth
	lb.BackendServerPort = args.BackendServerPort
	lb.AclStatus = args.AclStatus
	lb.AclId = args.AclId
	lb.AclType = args.AclType
	lb.Scheduler = args.Scheduler
	LOADBALANCER.listeners.Store(listenerKey(args.LoadBalancerId, args.ListenerPort), lb)
	return nil
}

func (c *mockClientSLB) SetLoadBalancerHTTPSListenerAttribute(args *slb.SetLoadBalancerHTTPSListenerAttributeArgs) (err error) {
	if c.setLoadBalancerHTTPSListenerAttribute != nil {
		return c.setLoadBalancerHTTPSListenerAttribute(args)
	}
	lb, err := c.DescribeLoadBalancerHTTPSListenerAttribute(args.LoadBalancerId, args.ListenerPort)
	if err != nil {
		return err
	}
	lb.CookieTimeout = args.CookieTimeout
	lb.Cookie = args.Cookie
	lb.StickySessionType = args.StickySessionType
	lb.StickySession = args.StickySession
	lb.HealthCheckTimeout = args.HealthCheckTimeout
	lb.HealthCheck = args.HealthCheck
	lb.HealthyThreshold = args.HealthyThreshold
	lb.HealthCheckInterval = args.HealthCheckInterval
	lb.HealthCheckHttpCode = args.HealthCheckHttpCode
	lb.HealthCheckDomain = args.HealthCheckDomain
	lb.HealthCheckConnectPort = args.HealthCheckConnectPort
	lb.HealthCheckURI = args.HealthCheckURI
	lb.UnhealthyThreshold = args.UnhealthyThreshold
	lb.ListenerPort = args.ListenerPort
	lb.VServerGroup = args.VServerGroup
	lb.VServerGroupId = args.VServerGroupId
	lb.Description = args.Description
	lb.Bandwidth = args.Bandwidth
	lb.BackendServerPort = args.BackendServerPort
	lb.AclStatus = args.AclStatus
	lb.AclId = args.AclId
	lb.AclType = args.AclType
	lb.Scheduler = args.Scheduler
	LOADBALANCER.listeners.Store(listenerKey(args.LoadBalancerId, args.ListenerPort), lb)
	return nil
}

func (c *mockClientSLB) SetLoadBalancerTCPListenerAttribute(args *slb.SetLoadBalancerTCPListenerAttributeArgs) (err error) {
	if c.setLoadBalancerTCPListenerAttribute != nil {
		return c.setLoadBalancerTCPListenerAttribute(args)
	}

	lb, err := c.DescribeLoadBalancerTCPListenerAttribute(args.LoadBalancerId, args.ListenerPort)
	if err != nil {
		return err
	}
	lb.HealthCheckConnectTimeout = args.HealthCheckConnectTimeout
	lb.HealthCheck = args.HealthCheck
	lb.HealthyThreshold = args.HealthyThreshold
	lb.HealthCheckInterval = args.HealthCheckInterval
	lb.HealthCheckHttpCode = args.HealthCheckHttpCode
	lb.HealthCheckDomain = args.HealthCheckDomain
	lb.HealthCheckConnectPort = args.HealthCheckConnectPort
	lb.HealthCheckURI = args.HealthCheckURI
	lb.UnhealthyThreshold = args.UnhealthyThreshold
	lb.ListenerPort = args.ListenerPort
	lb.VServerGroup = args.VServerGroup
	lb.VServerGroupId = args.VServerGroupId
	lb.Description = args.Description
	lb.Bandwidth = args.Bandwidth
	lb.PersistenceTimeout = args.PersistenceTimeout
	lb.BackendServerPort = args.BackendServerPort
	lb.AclStatus = args.AclStatus
	lb.AclId = args.AclId
	lb.AclType = args.AclType
	lb.Scheduler = args.Scheduler
	LOADBALANCER.listeners.Store(listenerKey(args.LoadBalancerId, args.ListenerPort), lb)
	return nil
}

func (c *mockClientSLB) SetLoadBalancerUDPListenerAttribute(args *slb.SetLoadBalancerUDPListenerAttributeArgs) (err error) {
	if c.setLoadBalancerUDPListenerAttribute != nil {
		return c.setLoadBalancerUDPListenerAttribute(args)
	}

	lb, err := c.DescribeLoadBalancerUDPListenerAttribute(args.LoadBalancerId, args.ListenerPort)
	if err != nil {
		return err
	}
	lb.HealthCheckConnectTimeout = args.HealthCheckConnectTimeout
	lb.HealthCheck = args.HealthCheck
	lb.HealthyThreshold = args.HealthyThreshold
	lb.HealthCheckInterval = args.HealthCheckInterval
	lb.HealthCheckConnectPort = args.HealthCheckConnectPort
	lb.UnhealthyThreshold = args.UnhealthyThreshold
	lb.ListenerPort = args.ListenerPort
	lb.VServerGroup = args.VServerGroup
	lb.VServerGroupId = args.VServerGroupId
	lb.Description = args.Description
	lb.Bandwidth = args.Bandwidth
	lb.PersistenceTimeout = lb.PersistenceTimeout
	lb.BackendServerPort = args.BackendServerPort
	lb.AclStatus = args.AclStatus
	lb.AclId = args.AclId
	lb.AclType = args.AclType
	lb.Scheduler = args.Scheduler
	LOADBALANCER.listeners.Store(listenerKey(args.LoadBalancerId, args.ListenerPort), lb)
	return nil
}

func (c *mockClientSLB) RemoveTags(args *slb.RemoveTagsArgs) error {
	if c.removeTags != nil {
		return c.removeTags(args)
	}
	v, ok := LOADBALANCER.tags.Load(args.LoadBalancerID)
	if !ok {
		return nil
	}
	tags := &[]slb.TagItem{}
	err := json.Unmarshal([]byte(args.Tags), tags)
	if err != nil {
		return err
	}

	ins, ok := v.([]slb.TagItem)
	if !ok {
		return fmt.Errorf("not TagItem type %s", reflect.TypeOf(v))
	}
	var result []slb.TagItem
	for _, t := range ins {
		found := false
		for _, m := range *tags {
			if t.TagKey == m.TagKey &&
				t.TagValue == m.TagValue {
				found = true
				break
			}
		}
		if !found {
			result = append(result, t)
		}
	}
	LOADBALANCER.tags.Store(args.LoadBalancerID, result)
	return nil
}

func (c *mockClientSLB) DescribeTags(args *slb.DescribeTagsArgs) (tags []slb.TagItemType, pagination *common.PaginationResult, err error) {
	if c.describeTags != nil {
		return c.describeTags(args)
	}
	v, ok := LOADBALANCER.tags.Load(args.LoadBalancerID)
	if !ok {
		return []slb.TagItemType{}, nil, nil
	}
	ins, ok := v.([]slb.TagItemType)
	if !ok {
		return nil, nil, fmt.Errorf("TagItem not found type %s", reflect.TypeOf(v))
	}

	return ins, nil, nil
}
func (c *mockClientSLB) AddTags(args *slb.AddTagsArgs) error {
	if c.addTags != nil {
		return c.addTags(args)
	}
	tags := &[]slb.TagItem{}
	err := json.Unmarshal([]byte(args.Tags), tags)
	if err != nil {
		return err
	}
	var tagstype []slb.TagItemType
	for _, tag := range *tags {
		tagstype = append(tagstype, slb.TagItemType{TagItem: tag})
	}
	v, ok := LOADBALANCER.tags.Load(args.LoadBalancerID)
	if !ok {
		LOADBALANCER.tags.Store(args.LoadBalancerID, tagstype)
		return nil
	}
	ins, ok := v.([]slb.TagItemType)
	if !ok {
		return fmt.Errorf("TagItem not found type %s", reflect.TypeOf(v))
	}
	for _, tag := range *tags {
		found := false

		for _, i := range ins {
			if tag.TagKey == i.TagKey &&
				tag.TagValue == i.TagValue {
				found = true
				break
			}
		}
		if !found {
			ins = append(ins, slb.TagItemType{TagItem: tag})
		}
	}
	LOADBALANCER.tags.Store(args.LoadBalancerID, ins)
	return nil
}

func vgroupKey(id, vgroupid string) string {
	return fmt.Sprintf("%s/%s", id, vgroupid)
}

func (c *mockClientSLB) CreateVServerGroup(args *slb.CreateVServerGroupArgs) (response *slb.CreateVServerGroupResponse, err error) {
	if c.createVServerGroup != nil {
		return c.createVServerGroup(args)
	}
	vgroup := slb.CreateVServerGroupResponse{
		VServerGroupId:   newid(),
		VServerGroupName: args.VServerGroupName,
	}
	LOADBALANCER.vgroups.Store(vgroupKey(args.LoadBalancerId, vgroup.VServerGroupId), vgroup)
	return &vgroup, nil
}

func (c *mockClientSLB) DescribeVServerGroups(args *slb.DescribeVServerGroupsArgs) (response *slb.DescribeVServerGroupsResponse, err error) {

	if c.describeVServerGroups != nil {
		return c.describeVServerGroups(args)
	}
	var vgr []slb.VServerGroup
	LOADBALANCER.vgroups.Range(
		func(key, value interface{}) bool {
			k := key.(string)
			v := value.(slb.CreateVServerGroupResponse)
			if strings.Contains(k, args.LoadBalancerId) {
				vgr = append(vgr, slb.VServerGroup{VServerGroupId: v.VServerGroupId, VServerGroupName: v.VServerGroupName})
			}
			return true
		},
	)

	return &slb.DescribeVServerGroupsResponse{
		VServerGroups: struct {
			VServerGroup []slb.VServerGroup
		}{
			VServerGroup: vgr,
		},
	}, nil
}

func (c *mockClientSLB) DeleteVServerGroup(args *slb.DeleteVServerGroupArgs) (response *slb.DeleteVServerGroupResponse, err error) {
	if c.deleteVServerGroup != nil {
		return c.deleteVServerGroup(args)
	}
	ikey := ""
	LOADBALANCER.vgroups.Range(
		func(key, value interface{}) bool {
			k := key.(string)
			if strings.Contains(k, args.VServerGroupId) {
				ikey = k
				return false
			}
			return true
		},
	)
	LOADBALANCER.vgroups.Delete(ikey)
	return nil, nil
}

func (c *mockClientSLB) SetVServerGroupAttribute(args *slb.SetVServerGroupAttributeArgs) (response *slb.SetVServerGroupAttributeResponse, err error) {
	if c.setVServerGroupAttribute != nil {
		return c.setVServerGroupAttribute(args)
	}
	return nil, nil
}

func (c *mockClientSLB) DescribeVServerGroupAttribute(args *slb.DescribeVServerGroupAttributeArgs) (response *slb.DescribeVServerGroupAttributeResponse, err error) {
	if c.describeVServerGroupAttribute != nil {
		return c.describeVServerGroupAttribute(args)
	}
	ikey := ""
	LOADBALANCER.vgroups.Range(
		func(key, value interface{}) bool {
			k := key.(string)

			if strings.Contains(k, args.VServerGroupId) {
				ikey = k
				return false
			}
			return true
		},
	)
	if ikey == "" {
		return nil, fmt.Errorf("vgroup not found, %s", args.VServerGroupId)
	}
	v, _ := LOADBALANCER.vgroups.Load(ikey)
	vgr := v.(slb.CreateVServerGroupResponse)

	return &slb.DescribeVServerGroupAttributeResponse{
		VServerGroupId:   vgr.VServerGroupId,
		VServerGroupName: vgr.VServerGroupName,
		BackendServers:   vgr.BackendServers,
	}, nil
}

func (c *mockClientSLB) ModifyVServerGroupBackendServers(args *slb.ModifyVServerGroupBackendServersArgs) (response *slb.ModifyVServerGroupBackendServersResponse, err error) {
	if c.modifyVServerGroupBackendServers != nil {
		return c.modifyVServerGroupBackendServers(args)
	}
	return nil, nil
}
func (c *mockClientSLB) AddVServerGroupBackendServers(args *slb.AddVServerGroupBackendServersArgs) (response *slb.AddVServerGroupBackendServersResponse, err error) {
	if c.addVServerGroupBackendServers != nil {
		return c.addVServerGroupBackendServers(args)
	}
	ikey := ""
	LOADBALANCER.vgroups.Range(
		func(key, value interface{}) bool {
			k := key.(string)
			if strings.Contains(k, args.VServerGroupId) {
				ikey = k
				return false
			}
			return true
		},
	)
	if ikey == "" {
		return nil, fmt.Errorf("add: vgroup not found, %s", args.VServerGroupId)
	}
	v, _ := LOADBALANCER.vgroups.Load(ikey)
	vgr := v.(slb.CreateVServerGroupResponse)
	backends := &[]slb.VBackendServerType{}
	err = json.Unmarshal([]byte(args.BackendServers), backends)
	if err != nil {
		return nil, err
	}

	for _, b := range *backends {
		found := false
		for _, cac := range vgr.BackendServers.BackendServer {
			if b.ServerId == cac.ServerId &&
				b.ServerIp == cac.ServerIp {
				found = true
				break
			}
		}
		if !found {
			vgr.BackendServers.BackendServer = append(vgr.BackendServers.BackendServer, b)
		}
	}
	LOADBALANCER.vgroups.Store(ikey, vgr)
	return &slb.AddVServerGroupBackendServersResponse{
		VServerGroupId:   vgr.VServerGroupId,
		VServerGroupName: vgr.VServerGroupName,
		BackendServers:   vgr.BackendServers,
	}, nil
}

func (c *mockClientSLB) RemoveVServerGroupBackendServers(args *slb.RemoveVServerGroupBackendServersArgs) (response *slb.RemoveVServerGroupBackendServersResponse, err error) {
	if c.removeVServerGroupBackendServers != nil {
		return c.removeVServerGroupBackendServers(args)
	}
	ikey := ""
	LOADBALANCER.vgroups.Range(
		func(key, value interface{}) bool {
			k := key.(string)
			if strings.Contains(k, args.VServerGroupId) {
				ikey = k
				return false
			}
			return true
		},
	)
	if ikey == "" {
		return nil, fmt.Errorf("remove: vgroup not found, %s", args.VServerGroupId)
	}
	v, _ := LOADBALANCER.vgroups.Load(ikey)
	vgr := v.(slb.CreateVServerGroupResponse)
	backends := &[]slb.VBackendServerType{}
	err = json.Unmarshal([]byte(args.BackendServers), backends)
	if err != nil {
		return nil, err
	}
	var result []slb.VBackendServerType
	for _, b := range vgr.BackendServers.BackendServer {
		found := false
		for _, cac := range *backends {
			if b.ServerId == cac.ServerId &&
				b.ServerIp == cac.ServerIp {
				found = true
				break
			}
		}
		if !found {
			result = append(result, b)
		}
	}
	vgr.BackendServers.BackendServer = result
	LOADBALANCER.vgroups.Store(ikey, vgr)
	return &slb.RemoveVServerGroupBackendServersResponse{
		VServerGroupName: vgr.VServerGroupName,
		VServerGroupId:   vgr.VServerGroupId,
		BackendServers:   vgr.BackendServers,
	}, nil
}
