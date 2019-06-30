package alicloud

import (
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/slb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sort"
	"strings"
	"testing"
)

// cloudDataMock
// is a function which set mocked cloud
// initial data, include alibaba route/slb/instance
type CloudDataMock func()

func DefaultPreset() {
	// Step 2: init cloud cache data.
	PreSetCloudData(
		// LoadBalancer
		WithNewLoadBalancerStore(),
		WithLoadBalancer(),

		// VPC & Route
		WithNewRouteStore(),
		WithVpcs(),
		WithVRouter(),

		// Instance Store
		WithNewInstanceStore(),
		WithInstance(),
	)
}

func PreSetCloudData(sets ...CloudDataMock) {
	for _, initialSet := range sets {
		initialSet()
	}
}

var (
	INSTANCEID = "i-xlakjbidlslkcdxxxx"
)

var (
	LOADBALANCER_ID           = "lb-bp1ids9hmq5924m6uk5w1"
	LOADBALANCER_NAME         = "a2cb99d47cc8311e899db00163e12560"
	LOADBALANCER_ADDRESS      = "47.97.241.114"
	LOADBALANCER_NETWORK_TYPE = "classic"
	LOADBALANCER_SPEC         = slb.LoadBalancerSpecType(slb.S1Small)

	SERVICE_UID = types.UID("2cb99d47-cc83-11e8-99db-00163e125603")
)

var (
	VPCID          = "vpc-2zeaybwqmvn6qgabfd3pe"
	VROUTER_ID     = "vrt-2zegcm0ty46mq243fmxoj"
	ROUTE_TABLE_ID = "vtb-2zedne8cr43rp5oqsr9xg"
	REGION         = common.Hangzhou
	REGION_A       = "cn-hangzhou-a"
	VSWITCH_ID     = "vsw-2zeclpmxy66zzxj4cg4ls"
	ROUTE_ENTRIES  = []ecs.RouteEntrySetType{
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "172.16.3.0/24",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "Instance",
			InstanceId:           "i-2zee0h6bdcgrocv2n9jb",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "172.16.2.0/24",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "Instance",
			InstanceId:           "i-2zecarjjmtkx3oru4233",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "172.16.0.0/24",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "Instance",
			InstanceId:           "i-2ze7q4vl8cosjsd56j0h",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "0.0.0.0/0",
			Type:                 ecs.RouteTableCustom,
			NextHopType:          "NatGateway",
			InstanceId:           "ngw-2zetlvdtq0zt9ubez3zz3",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "192.168.0.0/16",
			Type:                 ecs.RouteTableSystem,
			NextHopType:          "local",
			Status:               ecs.RouteEntryStatusAvailable,
		},
		{
			RouteTableId:         ROUTE_TABLE_ID,
			DestinationCidrBlock: "100.64.0.0/10",
			Type:                 ecs.RouteTableSystem,
			NextHopType:          "service",
			Status:               ecs.RouteEntryStatusAvailable,
		},
	}
)

func NewMockCloud() (*Cloud, error) {
	return newMockCloudWithSDK(
		&mockClientSLB{},
		&mockRouteSDK{},
		&mockClientInstanceSDK{},
		nil,
	)
}

func newMockCloudWithSDK(
	slb ClientSLBSDK,
	route RouteSDK,
	ins ClientInstanceSDK,
	meta *metadata.MetaData,
) (*Cloud, error) {

	if meta == nil {
		meta = metadata.NewMockMetaData(
			nil,
			func(resource string) (string, error) {
				if strings.Contains(resource, metadata.REGION) {
					return string(REGION), nil
				}
				if strings.Contains(resource, metadata.VPC_ID) {
					return VPCID, nil
				}
				return "", fmt.Errorf("not found")
			},
		)
	}
	mgr := &ClientMgr{
		stop:         make(<-chan struct{}, 1),
		meta:         meta,
		loadbalancer: &LoadBalancerClient{c: slb},
		routes:       &RoutesClient{client: route, region: string(REGION)},
		instance:     &InstanceClient{c: ins},
	}

	return newAliCloud(mgr, "")
}

func NewDefaultFrameWork(
	svc *v1.Service,
	nodes []*v1.Node,
	preset func(),
) *FrameWork {
	if preset == nil {
		preset = DefaultPreset
	}
	preset()
	cloud, err := NewMockCloud()
	if err != nil {
		panic(err)
	}
	return NewFrameWork(cloud, svc, nodes, nil)
}

func NewFrameWork(
	cloud *Cloud,
	svc *v1.Service,
	nodes []*v1.Node,
	preset func(),
) *FrameWork {
	if preset != nil {
		preset()
	}
	return &FrameWork{
		cloud: cloud,
		svc:   svc,
		nodes: nodes,
	}
}

type FrameWork struct {
	cloud         *Cloud
	svc           *v1.Service
	nodes         []*v1.Node
	cloudDataMock func()
}

func (f *FrameWork) Cloud() *Cloud                     { return f.cloud }
func (f *FrameWork) Instance() *InstanceClient         { return f.cloud.climgr.Instances() }
func (f *FrameWork) Route() *RoutesClient              { return f.cloud.climgr.Routes() }
func (f *FrameWork) LoadBalancer() *LoadBalancerClient { return f.cloud.climgr.LoadBalancers() }

func (f *FrameWork) SLBSDK() ClientSLBSDK           { return f.cloud.climgr.LoadBalancers().c }
func (f *FrameWork) RouteSDK() RouteSDK             { return f.cloud.climgr.Routes().client }
func (f *FrameWork) InstanceSDK() ClientInstanceSDK { return f.cloud.climgr.Instances().c }

func (f *FrameWork) hasAnnotation(anno string) bool { return serviceAnnotation(f.svc, anno) != "" }

func (f *FrameWork) RunDefault(
	t *testing.T,
	describe string,
) {
	f.Run(t, describe, nil)
}

func (f *FrameWork) Run(
	t *testing.T,
	describe string,
	run func(),
) {
	t.Log(describe)
	if run == nil {
		run = func() {
			status, e := f.cloud.EnsureLoadBalancer(CLUSTER_ID, f.svc, f.nodes)
			if e != nil {
				t.Errorf("TestEnsureLoadBalancerHTTPS error: %s\n", e.Error())
			}
			if !f.ExpectExistAndEqual(t) {
				t.Fatalf("https test fail")
			}

			if status == nil || len(status.Ingress) <= 0 {
				t.Fatalf("status nil")
			}
		}
	}
	run()
}

// service & cloud data must be consistent
func (f *FrameWork) ExpectExistAndEqual(t *testing.T) bool {

	exist, mlb, err := f.LoadBalancer().findLoadBalancer(f.svc)
	// 1. slb must exist
	if err != nil || !exist {
		t.Fatalf("slb not found, %v, %v", exist, err)
	}

	// 1. port length equal
	if len(f.svc.Spec.Ports) != len(mlb.ListenerPorts.ListenerPort) {
		t.Fatal("port length not equal")
	}

	// 2. port & proto equal
	for _, p := range f.svc.Spec.Ports {
		found := false
		for _, v := range mlb.ListenerPortsAndProtocol.ListenerPortAndProtocol {

			proto, err := Protocol(serviceAnnotation(f.svc, ServiceAnnotationLoadBalancerProtocolPort), p)
			if err != nil {
				t.Fatalf("proto transfor error")
			}
			if p.Port == int32(v.ListenerPort) &&
				proto == v.ListenerProtocol {
				if err := f.ListenerEqual(mlb.LoadBalancerId, p, proto); err != nil {
					t.Fatalf("listener configuration not equal")
				}
				found = true
				break
			}

		}
		if !found {
			t.Fatalf("port not found %d", p.Port)
		}
	}

	// 3. backend server
	res, err := f.SLBSDK().DescribeVServerGroups(&slb.DescribeVServerGroupsArgs{
		LoadBalancerId: mlb.LoadBalancerId,
	})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Fatalf("vserver group: %v", err)
	}
	if res == nil {
		t.Fatalf("vserver group can not be nil")
	}
	if len(res.VServerGroups.VServerGroup) != len(f.svc.Spec.Ports) {
		t.Fatalf("vserver group must be equal to port count")
	}

	sort.SliceStable(
		f.nodes,
		func(i, j int) bool {
			_, ida, err := nodeFromProviderID(f.nodes[i].Spec.ProviderID)
			if err != nil {
				t.Fatalf("unexpected provider id")
			}
			_, idb, err := nodeFromProviderID(f.nodes[j].Spec.ProviderID)
			if err != nil {
				t.Fatalf("unexpected provider id")
			}
			if ida > idb {
				return true
			}
			return false
		},
	)

	for _, v := range res.VServerGroups.VServerGroup {
		vg, err := f.SLBSDK().DescribeVServerGroupAttribute(
			&slb.DescribeVServerGroupAttributeArgs{
				VServerGroupId: v.VServerGroupId,
			},
		)
		if err != nil {
			t.Fatalf("vserver group attribute error: %v", err)
		}

		backends := vg.BackendServers.BackendServer
		if len(backends) != len(f.nodes) {
			t.Fatalf("vgroup backend server must be %d", len(f.nodes))
		}
		sort.SliceStable(
			backends,
			func(i, j int) bool {
				if backends[i].ServerId > backends[j].ServerId {
					return true
				}
				return false
			},
		)
		for k, v := range backends {
			_, ida, err := nodeFromProviderID(f.nodes[k].Spec.ProviderID)
			if err != nil {
				t.Fatalf("unexpected provider id")
			}
			if v.ServerId != ida {
				t.Fatalf("backend not equal")
			}
		}
	}

	return f.SLBSpecEqual(t, mlb)
}

func (f *FrameWork) SLBSpecEqual(t *testing.T, mlb *slb.LoadBalancerType) bool {

	defd, _ := ExtractAnnotationRequest(f.svc)
	if f.hasAnnotation(ServiceAnnotationLoadBalancerMasterZoneID) {
		if mlb.MasterZoneId != defd.MasterZoneID {
			t.Fatalf("master zoneid error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerSlaveZoneID) {
		if mlb.SlaveZoneId != defd.SlaveZoneID {
			t.Fatalf("slave zoneid error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerBandwidth) {
		if mlb.Bandwidth != defd.Bandwidth {
			t.Fatalf("bandwidth error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerAddressType) {
		if mlb.AddressType != defd.AddressType {
			t.Fatalf("address type error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerVswitch) {
		if mlb.VSwitchId != defd.VswitchID {
			t.Fatalf("vswitch id error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerSLBNetworkType) {
		if mlb.NetworkType != defd.SLBNetworkType {
			t.Fatalf("network type error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerChargeType) {
		if mlb.InternetChargeType != defd.ChargeType {
			t.Fatalf("charge type error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerSpec) {
		if mlb.LoadBalancerSpec != defd.LoadBalancerSpec {
			t.Fatalf("loadbalancer spec error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerIPVersion) {
		if mlb.AddressIPVersion != defd.AddressIPVersion {
			t.Fatalf("address ip version error")
		}
	}
	return true
}

func (f *FrameWork) ListenerEqual(id string, p v1.ServicePort, proto string) error {
	var (
		// Health check
		healthCheckInterval           = 0
		healthCheckTimeout            = 0
		healthCheckDomain             = ""
		healthCheckHTTPCode           = ""
		healthCheck                   = ""
		healthCheckType               = ""
		healthCheckURI                = ""
		healthCheckConnectPort        = 0
		healthCheckConnectTimeout     = 0
		healthCheckHealthyThreshold   = 0
		healthCheckUnhealthyThreshold = 0

		sessionStick       = ""
		sessionStickType   = ""
		cookieTimeout      = ""
		cookie             = ""
		persistenceTimeout = ""

		privateZoneName       = ""
		privateZoneId         = ""
		privateZoneRecordName = ""
		privateZoneRecordTTL  = ""
	)
	defd, _ := ExtractAnnotationRequest(f.svc)
	switch proto {
	case "tcp":
		resp, err := f.SLBSDK().DescribeLoadBalancerTCPListenerAttribute(id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort != int(p.NodePort) {
			return fmt.Errorf("TCPBackendServerPortNotEqual")
		}

		healthCheckInterval = resp.HealthCheckInterval
		healthCheckDomain = resp.HealthCheckDomain
		healthCheckHTTPCode = string(resp.HealthCheckHttpCode)
		healthCheckType = string(resp.HealthCheckType)
		healthCheckURI = resp.HealthCheckURI
		healthCheckConnectPort = resp.HealthCheckConnectPort
		healthCheckConnectTimeout = resp.HealthCheckConnectTimeout
		healthCheckHealthyThreshold = resp.HealthyThreshold
		healthCheckUnhealthyThreshold = resp.UnhealthyThreshold
		healthCheck = string(resp.HealthCheck)
	case "udp":
		resp, err := f.SLBSDK().DescribeLoadBalancerUDPListenerAttribute(id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort != int(p.NodePort) {
			return fmt.Errorf("UDPBackendServerPortNotEqual")
		}

		healthCheckInterval = resp.HealthCheckInterval
		healthCheckConnectPort = resp.HealthCheckConnectPort
		healthCheckConnectTimeout = resp.HealthCheckConnectTimeout
		healthCheckHealthyThreshold = resp.HealthyThreshold
		healthCheckUnhealthyThreshold = resp.UnhealthyThreshold
		healthCheck = string(resp.HealthCheck)
	case "http":
		resp, err := f.SLBSDK().DescribeLoadBalancerHTTPListenerAttribute(id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort != int(p.NodePort) {
			return fmt.Errorf("HTTPBackendServerPortNotEqual")
		}

		healthCheckTimeout = resp.HealthCheckTimeout
		healthCheckInterval = resp.HealthCheckInterval
		healthCheckDomain = resp.HealthCheckDomain
		healthCheckHTTPCode = string(resp.HealthCheckHttpCode)
		healthCheckURI = resp.HealthCheckURI
		healthCheckConnectPort = resp.HealthCheckConnectPort
		healthCheckHealthyThreshold = resp.HealthyThreshold
		healthCheckUnhealthyThreshold = resp.UnhealthyThreshold
		healthCheck = string(resp.HealthCheck)
	case "https":
		resp, err := f.SLBSDK().DescribeLoadBalancerHTTPSListenerAttribute(id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort != int(p.NodePort) {
			return fmt.Errorf("HTTPSBackendServerPortNotEqual")
		}
		if resp.ServerCertificateId != defd.CertID {
			return fmt.Errorf("HTTPSCertIDNotEqual")
		}
		healthCheckTimeout = resp.HealthCheckTimeout
		healthCheckInterval = resp.HealthCheckInterval
		healthCheckDomain = resp.HealthCheckDomain
		healthCheckHTTPCode = string(resp.HealthCheckHttpCode)
		healthCheckURI = resp.HealthCheckURI
		healthCheckConnectPort = resp.HealthCheckConnectPort
		healthCheckHealthyThreshold = resp.HealthyThreshold
		healthCheckUnhealthyThreshold = resp.UnhealthyThreshold
		healthCheck = string(resp.HealthCheck)
	default:
		return fmt.Errorf("unknown proto: %s", proto)
	}

	// --------------------------- SessionStick ----------------------------
	if f.hasAnnotation(ServiceAnnotationLoadBalancerSessionStick) {
		if sessionStick != string(defd.StickySession) {
			return fmt.Errorf("sticky session error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerSessionStickType) {
		if sessionStickType != string(defd.StickySessionType) {
			return fmt.Errorf("stick session type error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerCookieTimeout) {
		if cookieTimeout != string(defd.CookieTimeout) {
			return fmt.Errorf("cookie timeout error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerCookie) {
		if cookie != string(defd.Cookie) {
			return fmt.Errorf("cookie error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerPersistenceTimeout) {
		if persistenceTimeout != string(defd.PersistenceTimeout) {
			return fmt.Errorf("persistency timeout error")
		}
	}

	// ++++++++++++++++++++++++++++ Private Zone +++++++++++++++++++++++++++++
	if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneName) {
		if privateZoneName != string(defd.PrivateZoneName) {
			return fmt.Errorf("private zone name error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneId) {
		if privateZoneId != string(defd.PrivateZoneId) {
			return fmt.Errorf("private zone id error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneRecordName) {
		if privateZoneRecordName != string(defd.PrivateZoneRecordName) {
			return fmt.Errorf("private zone record name error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneRecordTTL) {
		if privateZoneRecordTTL != string(defd.PrivateZoneRecordTTL) {
			return fmt.Errorf("private zone record ttl error")
		}
	}

	//=========================== Health Check Test ============================
	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckType) {
		if healthCheckType != string(defd.HealthCheckType) {
			return fmt.Errorf("health check type error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckURI) {
		if healthCheckURI != string(defd.HealthCheckURI) {
			return fmt.Errorf("health check URI error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckConnectPort) {
		if healthCheckConnectPort != defd.HealthCheckConnectPort {
			return fmt.Errorf("health check connect port error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold) {
		if healthCheckHealthyThreshold != defd.HealthyThreshold {
			return fmt.Errorf("health check health threshhold error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold) {
		if healthCheckUnhealthyThreshold != defd.UnhealthyThreshold {
			return fmt.Errorf("health check unhealthy threshhold error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckInterval) {
		if healthCheckInterval != defd.HealthCheckInterval {
			return fmt.Errorf("health check interval error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckConnectTimeout) {
		if healthCheckConnectTimeout != defd.HealthCheckConnectTimeout {
			return fmt.Errorf("health check connect timeout error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckTimeout) {
		if healthCheckTimeout != defd.HealthCheckTimeout {
			return fmt.Errorf("health check timeout error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckDomain) {
		if healthCheckDomain != string(defd.HealthCheckDomain) {
			return fmt.Errorf("health check domain error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckHTTPCode) {
		if healthCheckHTTPCode != string(defd.HealthCheckHttpCode) {
			return fmt.Errorf("health check http code error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckFlag) {
		if healthCheck != string(defd.HealthCheck) {
			return fmt.Errorf("health check flag error")
		}
	}
	//=========================== Health Check Test ============================

	return nil
}
