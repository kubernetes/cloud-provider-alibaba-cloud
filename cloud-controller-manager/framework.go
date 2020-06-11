package alicloud

import (
	"context"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/pvtz"
	"github.com/denverdino/aliyungo/slb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/utils"
	"k8s.io/klog"
	controller "k8s.io/kube-aggregator/pkg/controllers"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// CloudDataMock
// is a function which set mocked Cloud
// initial data, include alibaba route/slb/instance
type CloudDataMock func()

func DefaultPreset() {
	InitCache()
	// Step 2: init Cloud cache data.
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
		WithENI(),
	)
}

func PreSetCloudData(sets ...CloudDataMock) {
	for _, initialSet := range sets {
		initialSet()
	}
}

var (
	INSTANCEID  = "i-xlakjbidlslkcdxxxx"
	INSTANCEID2 = "i-xlakjbidlslkcdxxxx2"
	ADDRESS     = "192.168.1.1"
)

var (
	LOADBALANCER_ID = "lb-bp1ids9hmq5924m6uk5w1"
	// do not change LOADBALANCER_NAME unless needed
	LOADBALANCER_NAME         = "ac83f8bed812e11e9a0ad00163e0a398"
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
				if strings.Contains(resource, metadata.VSWITCH_ID) {
					return VSWITCH_ID, nil
				}
				return "", fmt.Errorf("not found")
			},
		)
	}
	mgr := &ClientMgr{
		stop:         make(<-chan struct{}, 1),
		meta:         meta,
		loadbalancer: &LoadBalancerClient{c: slb, ins: ins, vpcid: VPCID},
		routes:       &RoutesClient{client: route, region: string(REGION)},
		instance:     &InstanceClient{c: ins},
	}

	return newAliCloud(mgr, "")
}

func NewDefaultFrameWork(
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
	return NewFrameWork(cloud, nil, nil, nil, nil)
}

func NewFrameWork(
	cloud *Cloud,
	svc *v1.Service,
	nodes []*v1.Node,
	endpoint *v1.Endpoints,
	preset func(),
) *FrameWork {
	if preset != nil {
		preset()
	}
	// set default service
	if svc == nil {
		svc = &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID("UID-1234567890-0987654321-1234556"),
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: 80, TargetPort: intstr.FromInt(8080), Protocol: v1.ProtocolTCP, NodePort: 8080},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		}
	}

	prid := nodeid(string(REGION), INSTANCEID)
	// set default nodes
	if nodes == nil {
		nodes = []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		}
	}
	// set default endpoint
	if endpoint == nil {
		endpoint = &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP:       ADDRESS,
							NodeName: &prid,
						},
					},
				},
			},
		}
	}
	return &FrameWork{
		Cloud:    cloud,
		SVC:      svc,
		Nodes:    nodes,
		Endpoint: endpoint,
	}
}

type OptionsFunc func(f *FrameWork)

func NewFrameWorkWithOptions(
	option OptionsFunc,
) *FrameWork {
	frame := &FrameWork{}
	option(frame)
	return frame
}

type FrameWork struct {
	Cloud         *Cloud
	SVC           *v1.Service
	Nodes         []*v1.Node
	Endpoint      *v1.Endpoints
	CloudDataMock func()
}

func (f *FrameWork) WithNodes(node []*v1.Node) *FrameWork { f.Nodes = node; return f }

func (f *FrameWork) WithEndpoints(endp *v1.Endpoints) *FrameWork { f.Endpoint = endp; return f }

func (f *FrameWork) WithService(svc *v1.Service) *FrameWork { f.SVC = svc; return f }

func (f *FrameWork) CloudImpl() *Cloud                 { return f.Cloud }
func (f *FrameWork) Instance() *InstanceClient         { return f.Cloud.climgr.Instances() }
func (f *FrameWork) Route() *RoutesClient              { return f.Cloud.climgr.Routes() }
func (f *FrameWork) LoadBalancer() *LoadBalancerClient { return f.Cloud.climgr.LoadBalancers() }

func (f *FrameWork) SLBSDK() ClientSLBSDK           { return f.Cloud.climgr.LoadBalancers().c }
func (f *FrameWork) RouteSDK() RouteSDK             { return f.Cloud.climgr.Routes().client }
func (f *FrameWork) InstanceSDK() ClientInstanceSDK { return f.Cloud.climgr.Instances().c }
func (f *FrameWork) PVTZSDK() ClientPVTZSDK         { return f.Cloud.climgr.PrivateZones().c }

func (f *FrameWork) hasAnnotation(anno string) bool { return serviceAnnotation(f.SVC, anno) != "" }

func (f *FrameWork) RunDefault(
	t *testing.T,
	describe string,
) {

	t.Log(fmt.Sprintf("RunDefault: %s, start", describe))
	err := f.Run(nil)
	if err != nil {
		t.Fatalf("RunDefault: %s, %s", describe, err.Error())
	}
}

func (f *FrameWork) RunCustomized(
	t *testing.T,
	describe string,
	custom CustomizedTest,
) {

	t.Log(describe)
	err := f.Run(custom)
	if err != nil {
		t.Fatalf("RunCustomized: %s, %s", describe, err.Error())
	}
}

func (f *FrameWork) Run(run CustomizedTest) error {
	//
	// initialize shared informer factory before run any test.
	f.Cloud.ifactory = informers.NewSharedInformerFactory(
		fake.NewSimpleClientset(f.Endpoint, f.SVC), 0,
	)
	// set informer
	inform := f.Cloud.ifactory.Core().V1().Endpoints().Informer()
	f.Cloud.ifactory.Start(nil)

	if !controller.WaitForCacheSync(
		"service", nil, inform.HasSynced,
	) {
		return fmt.Errorf("unable to initialize endpoint informer")
	}

	err := debug(f.Cloud.ifactory, f.SVC.Name, false)
	if err != nil {
		fmt.Printf("debug: %s\n", err.Error())
	}

	// override run if user defined specific test method
	if run != nil {
		// run customized testing
		// see DefaultTesting for more details.
		// usually,  call sequence:
		// 		EnsureLoadBalancer -> ExpectExistAndEqual -> Assert
		return run(f)
	}
	return DefaultTesting(f)
}

type CustomizedTest func(f *FrameWork) error

func DefaultTesting(f *FrameWork) error {
	status, err := f.CloudImpl().
		EnsureLoadBalancer(context.Background(), CLUSTER_ID, f.SVC, f.Nodes)
	if err != nil {
		return fmt.Errorf("EnsureLoadBalancer error: %s", err.Error())
	}
	if err := ExpectExistAndEqual(f); err != nil {
		return fmt.Errorf("test fail, expect equal, %s", err.Error())
	}

	if status == nil || len(status.Ingress) <= 0 {
		return fmt.Errorf("ingress status should not be nil: %v", status)
	}
	return nil
}

func debug(ifactory informers.SharedInformerFactory, name string, do bool) error {
	if !do {
		return nil
	}
	all, err := ifactory.
		Core().V1().
		Endpoints().
		Lister().
		Endpoints("default").
		List(labels.NewSelector())
	if err != nil {
		klog.Warningf("list: %s", err.Error())
	}
	fmt.Printf("AllObject: %v\n", utils.PrettyJson(all))
	endp, err := ifactory.
		Core().V1().
		Endpoints().
		Lister().
		Endpoints("default").
		Get(name)
	if err != nil {
		return fmt.Errorf("list: %s", err.Error())
	}
	fmt.Printf("GetObejct: %v\n", utils.PrettyJson(endp))
	return nil
}

// service & Cloud data must be consistent
func ExpectExistAndEqual(f *FrameWork) error {

	ctx := context.WithValue(context.Background(), utils.ContextService, f.SVC)
	exist, mlb, err := f.LoadBalancer().FindLoadBalancer(ctx, f.SVC)
	// 1. slb must exist
	if err != nil || !exist {
		return fmt.Errorf("slb must exist: %v, %v", exist, err)
	}

	// 1. port length equal
	if len(f.SVC.Spec.Ports) > len(mlb.ListenerPorts.ListenerPort) {
		return fmt.Errorf("not enough ports on loadbalancer: %d, %d", len(f.SVC.Spec.Ports), len(mlb.ListenerPorts.ListenerPort))
	}

	// 2. port & proto equal
	for _, p := range f.SVC.Spec.Ports {
		found := false
		for _, v := range mlb.ListenerPortsAndProtocol.ListenerPortAndProtocol {

			proto, err := Protocol(serviceAnnotation(f.SVC, ServiceAnnotationLoadBalancerProtocolPort), p)
			if err != nil {
				return fmt.Errorf("proto transfor error")
			}
			// If setting forward port, the BackendServerPort with http proto is 0.
			// So check whether the http port of slb equals the annotation's port.
			if f.hasAnnotation(ServiceAnnotationLoadBalancerForwardPort) && proto == "http" {
				ports := strings.Split(serviceAnnotation(f.SVC, ServiceAnnotationLoadBalancerForwardPort), ":")
				if len(ports) != 2 {
					return fmt.Errorf(fmt.Sprintf("forward-port format error: %s, expect 80:443,88:6443", ports))
				}
				httpPort, _ := strconv.Atoi(ports[0])
				if p.Port == int32(v.ListenerPort) && p.Port == int32(httpPort) {
					found = true
					break
				}
			} else {
				if p.Port == int32(v.ListenerPort) &&
					proto == v.ListenerProtocol {
					if err := f.ListenerEqual(ctx, mlb.LoadBalancerId, p, proto); err != nil {
						return fmt.Errorf(fmt.Sprintf("listener configuration not equal, %s", err.Error()))
					}
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("port not found %d", p.Port)
		}
	}

	// 3. backend server
	res, err := f.SLBSDK().DescribeVServerGroups(
		ctx,
		&slb.DescribeVServerGroupsArgs{
			LoadBalancerId: mlb.LoadBalancerId,
		},
	)
	if err != nil &&
		!strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("vserver group: %v", err)
	}
	if res == nil {
		return fmt.Errorf("vserver group can not be nil")
	}
	if len(res.VServerGroups.VServerGroup) < len(f.SVC.Spec.Ports) {
		return fmt.Errorf("vserver group count less than service: %d, %d", len(res.VServerGroups.VServerGroup), len(f.SVC.Spec.Ports))
	}

	sort.SliceStable(
		f.Nodes,
		func(i, j int) bool {
			_, ida, err := nodeFromProviderID(f.Nodes[i].Spec.ProviderID)
			if err != nil {
				panic("unexpected provider id")
			}
			_, idb, err := nodeFromProviderID(f.Nodes[j].Spec.ProviderID)
			if err != nil {
				panic("unexpected provider id")
			}
			if ida > idb {
				return true
			}
			return false
		},
	)

	defd, _ := ExtractAnnotationRequest(f.SVC)
	for _, v := range res.VServerGroups.VServerGroup {
		vg, err := f.SLBSDK().DescribeVServerGroupAttribute(
			ctx,
			&slb.DescribeVServerGroupAttributeArgs{
				VServerGroupId: v.VServerGroupId,
			},
		)
		if err != nil {
			return fmt.Errorf("vserver group attribute error: %v", err)
		}

		if isUserManagedVBackendServer(vg.VServerGroupName, f.SVC) {
			continue
		}

		backends := vg.BackendServers.BackendServer

		if f.SVC.Annotations[ServiceAnnotationLoadBalancerBackendType] == "eni" {

			if len(backends) != len(f.Endpoint.Subsets[0].Addresses) {
				return fmt.Errorf("Endpoint vgroup backend server must be %d", len(f.Nodes))
			}
			sort.SliceStable(
				backends,
				func(i, j int) bool {
					return backends[i].ServerIp < backends[j].ServerIp
				},
			)
			endpoints := f.Endpoint.Subsets[0].Addresses

			sort.SliceStable(
				endpoints,
				func(i, j int) bool {
					return endpoints[i].IP < endpoints[j].IP
				},
			)
			for k, v := range backends {
				if v.ServerIp != endpoints[k].IP {
					return fmt.Errorf("backend not equal Endpoint")
				}
			}
		} else if f.SVC.Spec.ExternalTrafficPolicy == "Local" {
			if len(f.Endpoint.Subsets) == 0 {
				return fmt.Errorf("Endpoint vgroup backend is 0. ")
			}

			//If multiple pods are running on one node,
			//there will be duplicate nodes in Endpoint.SubSets[0].Addresses.
			//The duplicate nodes need to be filtered.
			epNodeNameMap := make(map[string]string)
			for _, endpoint := range f.Endpoint.Subsets[0].Addresses {
				epNodeNameMap[*endpoint.NodeName] = *endpoint.NodeName
			}

			if len(backends) != len(epNodeNameMap) {
				return fmt.Errorf("Endpoint vgroup backend is not equal. ")
			}

			var endpointPrIds []string
			for _, epNodeName := range epNodeNameMap {
				found := false
				for _, node := range f.Nodes {
					if epNodeName == node.Name {
						endpointPrId := strings.Split(node.Spec.ProviderID, ".")
						if len(endpointPrId) != 2 {
							return fmt.Errorf("Node providerID %v format error. ", endpointPrId[1])
						}
						endpointPrIds = append(endpointPrIds, endpointPrId[1])
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Fail to find node: %s in Endpoint vgroup backend. ", epNodeName)
				}
			}
			sort.SliceStable(
				backends,
				func(i, j int) bool {
					return backends[i].ServerId < backends[j].ServerId
				},
			)

			sort.SliceStable(
				endpointPrIds,
				func(i, j int) bool {
					return endpointPrIds[i] < endpointPrIds[j]
				},
			)
			for k, v := range backends {
				if v.ServerId != endpointPrIds[k] {
					return fmt.Errorf("backend %v not equal Endpoint %v", v.ServerId, endpointPrIds[k])
				}
			}

		} else {
			f.Nodes = filterOutMaster(f.Nodes)
			if !f.hasAnnotation(ServiceAnnotationLoadBalancerBackendLabel) && len(backends) != len(f.Nodes) {
				return fmt.Errorf("node vgroup backend server must be %d", len(f.Nodes))
			}
			sort.SliceStable(
				backends,
				func(i, j int) bool {
					return backends[i].ServerId < backends[j].ServerId
				},
			)
			sort.SliceStable(
				f.Nodes,
				func(i, j int) bool {
					_, ida, err := nodeFromProviderID(f.Nodes[i].Spec.ProviderID)
					if err != nil {
						panic("xnode id error")
					}
					_, idb, err := nodeFromProviderID(f.Nodes[j].Spec.ProviderID)
					if err != nil {
						panic("ynode id error")
					}
					if ida < idb {
						return true
					}
					return false
				},
			)
			nodes := f.Nodes
			if f.hasAnnotation(ServiceAnnotationLoadBalancerBackendLabel) {
				var tpms []*v1.Node
				for _, v := range f.Nodes {
					if containsLabel(
						v,
						strings.Split(defd.BackendLabel, ","),
					) {
						tpms = append(tpms, v)
					}
				}
				nodes = tpms
			}
			if len(backends) > len(nodes) {
				return fmt.Errorf("backend:%v,node:%v not equal", len(backends), len(nodes))
			}
			for k, v := range backends {
				_, ida, err := nodeFromProviderID(nodes[k].Spec.ProviderID)
				if err != nil {
					return fmt.Errorf("unexpected provider id")
				}
				if v.ServerId != ida {
					return fmt.Errorf("backend not equal")
				}
			}

		}
	}

	// 4. describe tags
	if f.hasAnnotation(ServiceAnnotationLoadBalancerAdditionalTags) {
		arg := &slb.DescribeTagsArgs{LoadBalancerID: mlb.LoadBalancerId}
		tags, _, err := f.SLBSDK().DescribeTags(ctx, arg)
		if err != nil {
			return fmt.Errorf("describe tags:%s", err.Error())
		}

		if !tagsEqual(f.SVC.Annotations[ServiceAnnotationLoadBalancerAdditionalTags], tags) {
			return fmt.Errorf("tags not equal")
		}
	}

	// 5. private zone
	if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneRecordName) {
		var privateZoneRecordTTL int
		var selectedZoneId string

		if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneId) {
			selectedZoneId = defd.PrivateZoneId
		} else if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneName) {
			zones, err := f.PVTZSDK().DescribeZones(
				ctx,
				&pvtz.DescribeZonesArgs{
					Lang:    DEFAULT_LANG,
					Keyword: defd.PrivateZoneName,
				},
			)
			if err != nil {
				return fmt.Errorf("DescribeZones error: %s. ", err.Error())
			}
			if len(zones) == 0 {
				return fmt.Errorf("can not find zone by zone name %s. ", defd.PrivateZoneName)
			}

			for _, zone := range zones {
				if zone.ZoneName == defd.PrivateZoneName {
					selectedZoneId = zone.ZoneId
					break
				}
			}
		}

		pvrds, err := f.PVTZSDK().DescribeZoneRecords(
			ctx,
			&pvtz.DescribeZoneRecordsArgs{
				ZoneId: selectedZoneId,
			},
		)
		if err != nil {
			return fmt.Errorf("DescribeZoneRecords error: %s. ", err.Error())
		}
		found := false
		for _, record := range pvrds {
			if record.Rr == defd.PrivateZoneRecordName {
				privateZoneRecordTTL = record.Ttl
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("can not find private zone record %s", defd.PrivateZoneRecordName)
		}

		if f.hasAnnotation(ServiceAnnotationLoadBalancerPrivateZoneRecordTTL) {
			if privateZoneRecordTTL != defd.PrivateZoneRecordTTL {
				return fmt.Errorf("private zone record ttl not equal. ")
			}
		}
	}

	return f.SLBSpecEqual(mlb)
}

func (f *FrameWork) SLBSpecEqual(mlb *slb.LoadBalancerType) error {

	defd, _ := ExtractAnnotationRequest(f.SVC)
	if f.hasAnnotation(ServiceAnnotationLoadBalancerMasterZoneID) {
		if mlb.MasterZoneId != defd.MasterZoneID {
			return fmt.Errorf("master zoneid error: %s, %s", mlb.MasterZoneId, defd.MasterZoneID)
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerSlaveZoneID) {
		if mlb.SlaveZoneId != defd.SlaveZoneID {
			return fmt.Errorf(fmt.Sprintf("slave zoneid error:%s, %s", mlb.SlaveZoneId, defd.SlaveZoneID))
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerBandwidth) {
		if mlb.Bandwidth != defd.Bandwidth {
			return fmt.Errorf("bandwidth error: %d, %d", mlb.Bandwidth, defd.Bandwidth)
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerAddressType) {
		if mlb.AddressType != defd.AddressType {
			return fmt.Errorf("address type error: %s, %s", mlb.AddressType, defd.AddressType)
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerVswitch) {
		if mlb.VSwitchId != defd.VswitchID {
			return fmt.Errorf("vswitch id error: %s, %s", mlb.VSwitchId, defd.VswitchID)
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerSLBNetworkType) {
		if mlb.NetworkType != defd.SLBNetworkType {
			return fmt.Errorf("network type error")
		}
	}
	if f.hasAnnotation(ServiceAnnotationLoadBalancerChargeType) {
		if mlb.InternetChargeType != defd.ChargeType {
			return fmt.Errorf("charge type error: %s, %s", mlb.InternetChargeType, defd.ChargeType)
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerSpec) {
		if mlb.LoadBalancerSpec != defd.LoadBalancerSpec {
			return fmt.Errorf("loadbalancer spec error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerIPVersion) {
		if mlb.AddressIPVersion != defd.AddressIPVersion {
			return fmt.Errorf("address ip version error")
		}
	}
	return nil
}

func (f *FrameWork) ListenerEqual(ctx context.Context, id string, p v1.ServicePort, proto string) error {
	var (
		// Health check
		healthCheckInterval           int
		healthCheckTimeout            int
		healthCheckDomain             string
		healthCheckHTTPCode           string
		healthCheck                   string
		healthCheckType               string
		healthCheckURI                string
		healthCheckConnectPort        int
		healthCheckConnectTimeout     int
		healthCheckHealthyThreshold   int
		healthCheckUnhealthyThreshold int

		// session
		sessionStick       string
		sessionStickType   string
		cookieTimeout      int
		cookie             string
		persistenceTimeout int

		// acl
		aclStatus string
		aclId     string
		aclType   string

		scheduler string
	)
	defd, _ := ExtractAnnotationRequest(f.SVC)
	switch proto {
	case "tcp":
		resp, err := f.SLBSDK().DescribeLoadBalancerTCPListenerAttribute(ctx, id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort == 0 ||
			(!utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.NodePort)) ||
			(utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.TargetPort.IntVal)) {
			return fmt.Errorf("TCPBackendServerPortNotEqual")
		}

		if resp.PersistenceTimeout != nil {
			persistenceTimeout = *resp.PersistenceTimeout
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
		aclId = resp.AclId
		aclStatus = resp.AclStatus
		aclType = resp.AclType
		scheduler = string(resp.Scheduler)
	case "udp":
		resp, err := f.SLBSDK().DescribeLoadBalancerUDPListenerAttribute(ctx, id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort == 0 ||
			(!utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.NodePort)) ||
			(utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.TargetPort.IntVal)) {
			return fmt.Errorf("UDPBackendServerPortNotEqual")
		}

		healthCheckInterval = resp.HealthCheckInterval
		healthCheckConnectPort = resp.HealthCheckConnectPort
		healthCheckConnectTimeout = resp.HealthCheckConnectTimeout
		healthCheckHealthyThreshold = resp.HealthyThreshold
		healthCheckUnhealthyThreshold = resp.UnhealthyThreshold
		healthCheck = string(resp.HealthCheck)
		aclId = resp.AclId
		aclStatus = resp.AclStatus
		aclType = resp.AclType
		scheduler = string(resp.Scheduler)
	case "http":
		resp, err := f.SLBSDK().DescribeLoadBalancerHTTPListenerAttribute(ctx, id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort == 0 ||
			(!utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.NodePort)) ||
			(utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.TargetPort.IntVal)) {
			return fmt.Errorf("HTTPBackendServerPortNotEqual: %v, %v,%v", resp.BackendServerPort, p.NodePort, p.Port)
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
		sessionStick = string(resp.StickySession)
		sessionStickType = string(resp.StickySessionType)
		cookie = resp.Cookie
		cookieTimeout = resp.CookieTimeout
		aclId = resp.AclId
		aclStatus = resp.AclStatus
		aclType = resp.AclType
		scheduler = string(resp.Scheduler)
	case "https":
		resp, err := f.SLBSDK().DescribeLoadBalancerHTTPSListenerAttribute(ctx, id, int(p.Port))
		if err != nil {
			return err
		}
		if resp.BackendServerPort == 0 ||
			(!utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.NodePort)) ||
			(utils.IsENIBackendType(f.SVC) && resp.BackendServerPort != int(p.TargetPort.IntVal)) {
			return fmt.Errorf("HTTPSBackendServerPortNotEqual")
		}
		if resp.ServerCertificateId == "" ||
			resp.ServerCertificateId != defd.CertID {
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
		sessionStick = string(resp.StickySession)
		sessionStickType = string(resp.StickySessionType)
		cookie = resp.Cookie
		cookieTimeout = resp.CookieTimeout
		aclId = resp.AclId
		aclStatus = resp.AclStatus
		aclType = resp.AclType
		scheduler = string(resp.Scheduler)
	default:
		return fmt.Errorf("unknown proto: %s", proto)
	}
	// --------------------------- acl ---------------------------
	if f.hasAnnotation(ServiceAnnotationLoadBalancerAclID) {
		if aclId != string(defd.AclID) {
			return fmt.Errorf("acl id error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerAclStatus) {
		if aclStatus != string(defd.AclStatus) {
			return fmt.Errorf("acl status error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerAclType) {
		if aclType != defd.AclType {
			return fmt.Errorf("acl type error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerScheduler) {
		if scheduler != defd.Scheduler {
			return fmt.Errorf("scheduler type error")
		}
	}
	// --------------------------- SessionStick ----------------------------

	if (proto == "http" || proto == "https") &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerSessionStick) {
		if sessionStick != string(defd.StickySession) {
			return fmt.Errorf("sticky session error")
		}
	}

	if (proto == "http" || proto == "https") &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerSessionStickType) {
		if sessionStickType != string(defd.StickySessionType) {
			return fmt.Errorf("stick session type error")
		}
	}

	if (proto == "http" || proto == "https") &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerCookieTimeout) {
		if cookieTimeout != defd.CookieTimeout {
			return fmt.Errorf("cookie timeout error")
		}
	}

	if (proto == "http" || proto == "https") &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerCookie) {
		if cookie != string(defd.Cookie) {
			return fmt.Errorf("cookie error")
		}
	}

	if proto == "tcp" &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerPersistenceTimeout) {
		if persistenceTimeout != *defd.PersistenceTimeout {
			return fmt.Errorf("persistency timeout error: %d, %d", persistenceTimeout, defd.PersistenceTimeout)
		}
	}

	//=========================== Health Check Test ============================
	if proto == "tcp" &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckType) {
		if healthCheckType != string(defd.HealthCheckType) {
			return fmt.Errorf("health check type error: %s,%s", healthCheckType, defd.HealthCheckType)
		}
	}

	// Health checks with TCP type only work with listeners of the same protocol.
	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckConnectPort) {
		if healthCheckConnectPort != defd.HealthCheckConnectPort && healthCheckType == proto {
			return fmt.Errorf("health check connect port error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold) {
		if healthCheckHealthyThreshold != defd.HealthyThreshold && healthCheckType == proto {
			return fmt.Errorf("health check health threshold error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold) {
		if healthCheckUnhealthyThreshold != defd.UnhealthyThreshold && healthCheckType == proto {
			return fmt.Errorf("health check unhealthy threshold error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckInterval) {
		if healthCheckInterval != defd.HealthCheckInterval && healthCheckType == proto {
			return fmt.Errorf("health check interval error")
		}
	}

	if (proto == "tcp" || proto == "udp") &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckConnectTimeout) {
		if healthCheckConnectTimeout != defd.HealthCheckConnectTimeout {
			return fmt.Errorf("health check connect timeout error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckTimeout) {
		if healthCheckTimeout != defd.HealthCheckTimeout {
			return fmt.Errorf("health check timeout error")
		}
	}

	if f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckURI) {
		if healthCheckURI != string(defd.HealthCheckURI) {
			return fmt.Errorf("health check URI error")
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

	if (proto == "http" || proto == "https") &&
		f.hasAnnotation(ServiceAnnotationLoadBalancerHealthCheckFlag) {
		if healthCheck != string(defd.HealthCheck) {
			return fmt.Errorf("health check flag error")
		}
	}

	return nil
}

func containsLabel(node *v1.Node, lbl []string) bool {
	for _, m := range lbl {
		found := false
		for k, v := range node.Labels {
			label := fmt.Sprintf("%s=%s", k, v)
			if label == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func tagsEqual(tags string, items []slb.TagItemType) bool {
	stags := strings.Split(tags, ",")
	for _, m := range stags {
		found := false
		for _, v := range items {
			label := fmt.Sprintf("%s=%s", v.TagKey, v.TagValue)
			if label == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

const labelNodeRoleMaster = "node-role.kubernetes.io/master"

func filterOutMaster(nodes []*v1.Node) []*v1.Node {
	var result []*v1.Node
	for _, node := range nodes {
		found := false
		for k := range node.Labels {
			if k == labelNodeRoleMaster {
				found = true
			}
		}
		if !found {
			result = append(result, node)
		}
	}
	return result
}

func ExpectNotExist(f *FrameWork) error {
	ctx := context.WithValue(context.Background(), utils.ContextService, f.SVC)
	exist, _, err := f.LoadBalancer().FindLoadBalancer(ctx, f.SVC)
	if err != nil || exist {
		return fmt.Errorf("slb should not exist: %v, %t", err, exist)
	}
	return nil
}

func ExpectExist(f *FrameWork) error {
	ctx := context.WithValue(context.Background(), utils.ContextService, f.SVC)
	exist, _, err := f.LoadBalancer().FindLoadBalancer(ctx, f.SVC)
	if err != nil || !exist {
		return fmt.Errorf("ExpectExist: slb should exist, %v, %t", err, exist)
	}
	return nil
}

func ExpectAddressTypeNotEqual(f *FrameWork) error {
	ctx := context.WithValue(context.Background(), utils.ContextService, f.SVC)
	exist, mlb, err := f.LoadBalancer().FindLoadBalancer(ctx, f.SVC)
	if err != nil || !exist {
		return fmt.Errorf("ExpectAddressTypeNotEqual: slb should exist, isExist=%v, %v", exist, err)
	}
	defd, _ := ExtractAnnotationRequest(f.SVC)
	if f.hasAnnotation(ServiceAnnotationLoadBalancerAddressType) {
		if mlb.AddressType == defd.AddressType {
			return fmt.Errorf("address type mutate error: %s, %s", mlb.AddressType, defd.AddressType)
		}
	}
	return nil
}

func isUserManagedVBackendServer(VServerGroupName string, service *v1.Service) bool {
	for _, port := range service.Spec.Ports {
		vg := &vgroup{
			NamedKey: &NamedKey{
				CID:         CLUSTER_ID,
				Port:        vgroupPort(service, port),
				Namespace:   service.Namespace,
				ServiceName: service.Name,
				Prefix:      DEFAULT_PREFIX,
			},
		}
		if vg.NamedKey.Key() == VServerGroupName {
			return false
		}
	}

	return true
}
