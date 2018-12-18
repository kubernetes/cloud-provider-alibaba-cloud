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
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

var keyid string
var keysecret string

var (
	vpcId             = "vpc-2zeaybwqmvn6qgabfd3pe"
	vswitchid         = "vsw-2zeclpmxy66zzxj4cg4ls"
	regionId          = "cn-beijing"
	zoneId            = "cn-beijing-b"
	certID            = "1745547945134207_157f665c830"
	listenPort1 int32 = 80
	listenPort2 int32 = 90
	targetPort1       = intstr.FromInt(8080)
	targetPort2       = intstr.FromInt(9090)
	nodePort1   int32 = 8080
	nodePort2   int32 = 9090
	protocolTcp       = v1.ProtocolTCP
	protocolUdp       = v1.ProtocolUDP
	node1             = "i-bp1bcl00jhxr754tw8vx"
	node2             = "i-bp1bcl00jhxr754tw8vy"
	clusterName       = "clusterName-random"
	serviceUID        = "UID-1234567890-0987654321-1234556"
	nodeName          = "iZuf694l8lw6xvdx6gh7tkZ"
)

func TestBase64(t *testing.T) {
	data := "YWJjCg=="
	key, err := b64.StdEncoding.DecodeString(data)
	if err != nil {
		t.Fail()
	}
	t.Log(string(key))
}

func TestCloudConfigInit(t *testing.T) {
	config := strings.NewReader(con)
	var cfg CloudConfig
	if err := json.NewDecoder(config).Decode(&cfg); err != nil {
		t.Error(err)
	}
	if cfg.Global.AccessKeyID == "" || cfg.Global.AccessKeySecret == "" {
		t.Error("AccessKeyID or AccessKeySecret Must not null")
	}
}

func newMockCloud(slb ClientSLBSDK, route RouteSDK, ins ClientInstanceSDK, meta *metadata.MetaData) (*Cloud, error) {
	if meta == nil {
		meta = metadata.NewMockMetaData(nil, func(resource string) (string, error) {
			if strings.Contains(resource, metadata.REGION) {
				return regionId, nil
			}
			if strings.Contains(resource, metadata.VPC_ID) {
				return vswitchid, nil
			}
			return "", errors.New("not found")
		})
	}
	mgr := &ClientMgr{
		stop: make(<-chan struct{}, 1),
		meta: meta,
		loadbalancer: &LoadBalancerClient{
			c: slb,
		},
		routes: &RoutesClient{
			client: route,
		},
		instance: &InstanceClient{
			c: ins,
		},
	}

	return newAliCloud(mgr)
}

func newBasicService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "basic-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
}
func newHTTPSSerice() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "https-service",
			UID:  types.UID(serviceUID),
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerProtocolPort: fmt.Sprintf("https:%d", listenPort1),
				ServiceAnnotationLoadBalancerCertID:       certID,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
}
func newPortRangeService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "port-range-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1,
				}, {
					Port: 443, TargetPort: intstr.FromInt(443), Protocol: v1.ProtocolTCP, NodePort: 30443,
				},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
}

func newReadyService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-service",
			UID:         types.UID(serviceUID),
			Annotations: map[string]string{
			//ServiceAnnotationLoadBalancerId: "idbllll",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: 31789},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
		Status: v1.ServiceStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "1.1.1.1",
					},
				},
			},
		},
	}
}
func newNode1() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: node1},
		Spec: v1.NodeSpec{
			ProviderID: nodeid(regionId, node1),
		},
	}
}

func newMockClientInstanceSDK(instanceid string) ClientInstanceSDK {
	return &mockClientInstanceSDK{
		describeInstances: func(args *ecs.DescribeInstancesArgs) (instances []ecs.InstanceAttributesType, pagination *common.PaginationResult, err error) {
			if !strings.Contains(args.InstanceIds, instanceid) {
				return nil, nil, errors.New("not found")
			}
			instances = []ecs.InstanceAttributesType{
				{
					InstanceId:          instanceid,
					ImageId:             "centos_7_04_64_20G_alibase_201701015.vhd",
					RegionId:            common.Region(regionId),
					ZoneId:              zoneId,
					InstanceType:        "ecs.sn1ne.large",
					InstanceTypeFamily:  "ecs.sn1ne",
					Status:              "running",
					InstanceNetworkType: "vpc",
					VpcAttributes: ecs.VpcAttributesType{
						VpcId:     vpcId,
						VSwitchId: vswitchid,
						PrivateIpAddress: ecs.IpAddressSetType{
							IpAddress: []string{"192.168.211.130"},
						},
					},
					InstanceChargeType: common.PostPaid,
				},
			}
			return instances, nil, nil
		},
	}

}
func newMockClientSLB(service *v1.Service, nodes []*v1.Node, base *[]slb.LoadBalancerType, detail *slb.LoadBalancerType) *mockClientSLB {
	grp := vgroups{
		{
			NamedKey:       &NamedKey{CID: CLUSTER_ID, ServiceName: service.Name, Namespace: service.Namespace, Port: listenPort1},
			LoadBalancerId: detail.LoadBalancerId,
			RegionId:       detail.RegionId,
			VGroupId:       "v-idvgroup",
		},
	}
	return &mockClientSLB{
		describeLoadBalancers: func(args *slb.DescribeLoadBalancersArgs) (loadBalancers []slb.LoadBalancerType, err error) {

			if args.LoadBalancerId != "" {
				(*base)[0].LoadBalancerId = args.LoadBalancerId
				return *base, nil
			}
			if args.LoadBalancerName != "" {
				(*base)[0].LoadBalancerName = args.LoadBalancerName
				return *base, nil
			}
			if len(args.Tags) > 0 {
				(*base)[0].LoadBalancerName = cloudprovider.GetLoadBalancerName(service)
			} else {
				return nil, errors.New("loadbalancerid or loadbanancername must be specified")
			}
			return *base, nil
		},
		describeLoadBalancerAttribute: func(loadBalancerId string) (loadBalancer *slb.LoadBalancerType, err error) {
			return loadbalancerAttrib(&((*base)[0])), nil
		},
		createLoadBalancerTCPListener: func(args *slb.CreateLoadBalancerTCPListenerArgs) (err error) {
			li := slb.ListenerPortAndProtocolType{
				ListenerPort:     args.ListenerPort,
				ListenerProtocol: "tcp",
			}
			detail.ListenerPorts.ListenerPort = append(detail.ListenerPorts.ListenerPort, args.ListenerPort)
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = append(detail.ListenerPortsAndProtocol.ListenerPortAndProtocol, li)
			return nil
		},
		createLoadBalancerHTTPSListener: func(args *slb.CreateLoadBalancerHTTPSListenerArgs) (err error) {
			// check certid
			if args.ServerCertificateId != certID {
				return fmt.Errorf("server cert must be provided and equals to [%s]", certID)
			}
			li := slb.ListenerPortAndProtocolType{
				ListenerPort:     args.ListenerPort,
				ListenerProtocol: "https",
			}
			detail.ListenerPorts.ListenerPort = append(detail.ListenerPorts.ListenerPort, args.ListenerPort)
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = append(detail.ListenerPortsAndProtocol.ListenerPortAndProtocol, li)
			return nil
		},
		deleteLoadBalancerListener: func(loadBalancerId string, port int) (err error) {
			response := []slb.ListenerPortAndProtocolType{}
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol
			for _, p := range ports {
				if p.ListenerPort == port {
					continue
				}
				response = append(response, p)
			}

			listports := []int{}
			lports := detail.ListenerPorts.ListenerPort
			for _, po := range lports {
				if po == port {
					continue
				}
				listports = append(listports, po)
			}
			detail.ListenerPortsAndProtocol.ListenerPortAndProtocol = response
			detail.ListenerPorts.ListenerPort = listports
			return nil
		},
		describeLoadBalancerTCPListenerAttribute: func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol

			for _, p := range ports {
				if p.ListenerPort == port {
					return &slb.DescribeLoadBalancerTCPListenerAttributeResponse{
						DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{
							Status: slb.Running,
						},
						TCPListenerType: slb.TCPListenerType{
							LoadBalancerId:    loadBalancerId,
							ListenerPort:      port,
							BackendServerPort: 31789,
							Bandwidth:         50,
						},
					}, nil
				}
			}
			return nil, errors.New("not found")
		},
		describeLoadBalancerHTTPSListenerAttribute: func(loadBalancerId string, port int) (response *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error) {
			ports := detail.ListenerPortsAndProtocol.ListenerPortAndProtocol

			for _, p := range ports {
				if p.ListenerPort == port {
					return &slb.DescribeLoadBalancerHTTPSListenerAttributeResponse{
						DescribeLoadBalancerListenerAttributeResponse: slb.DescribeLoadBalancerListenerAttributeResponse{
							Status: slb.Running,
						},
						HTTPSListenerType: slb.HTTPSListenerType{
							HTTPListenerType: slb.HTTPListenerType{
								LoadBalancerId:    loadBalancerId,
								ListenerPort:      port,
								BackendServerPort: 31789,
								Bandwidth:         50,
							},
							ServerCertificateId: certID,
						},
					}, nil
				}
			}
			return nil, errors.New("not found")
		},
		removeBackendServers: func(loadBalancerId string, backendServers []string) (result []slb.BackendServerType, err error) {
			servers := detail.BackendServers.BackendServer
			target := []slb.BackendServerType{}
			for _, server := range servers {
				found := false
				for _, backendServer := range backendServers {
					if server.ServerId == backendServer {
						found = true
					}
				}
				if !found {
					target = append(target, server)
				}
			}
			detail.BackendServers.BackendServer = target
			return target, nil
		},
		addBackendServers: func(loadBalancerId string, backendServers []slb.BackendServerType) (result []slb.BackendServerType, err error) {
			detail.BackendServers.BackendServer = append(detail.BackendServers.BackendServer, backendServers...)
			return detail.BackendServers.BackendServer, nil
		},
		describeVServerGroups: func(args *slb.DescribeVServerGroupsArgs) (response *slb.DescribeVServerGroupsResponse, err error) {
			return &slb.DescribeVServerGroupsResponse{
				VServerGroups: struct {
					VServerGroup []slb.VServerGroup
				}{
					[]slb.VServerGroup{{
						VServerGroupId:   grp[0].VGroupId,
						VServerGroupName: grp[0].NamedKey.Key(),
					}},
				},
			}, nil
		},
		createVServerGroup: func(args *slb.CreateVServerGroupArgs) (response *slb.CreateVServerGroupResponse, err error) {
			return &slb.CreateVServerGroupResponse{
				VServerGroupName: args.VServerGroupName,
				VServerGroupId:   grp[0].VGroupId,
			}, nil
		},
		deleteLoadBalancer: func(loadBalancerId string) (err error) {
			*base = []slb.LoadBalancerType{}
			return nil
		},
	}
}
func TestEnsureLoadBalancerBasic(t *testing.T) {
	service := newBasicService()
	nodes := []*v1.Node{
		newNode1(),
	}
	base := newBaseLoadbalancer()
	detail := loadbalancerAttrib(&base[0])
	t.Log(PrettyJson(detail))

	// New Mock cloud to test
	cloud, err := newMockCloud(newMockClientSLB(service, nodes, &base, detail), nil, newMockClientInstanceSDK(node1), nil)

	if err != nil {
		t.Fatal(fmt.Sprintf("TestEnsureLoadBalancer error newCloud: %s\n", err.Error()))
	}

	status, e := cloud.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancer error: %s\n", e.Error())
	}
	t.Log(PrettyJson(status))
	t.Log(PrettyJson(detail))
	if len(detail.BackendServers.BackendServer) != 1 {
		t.Fatal("TestEnsureLoadBalancer error, expected only 1 backend left")
	}
	if detail.BackendServers.BackendServer[0].ServerId != node1 {
		t.Fatal(fmt.Sprintf("TestEnsureLoadBalancer error, expected to be instance [%s]", node1))
	}
	if len(detail.ListenerPorts.ListenerPort) != 1 {
		t.Fatal("TestEnsureLoadBalancer error, expected only 1 listener port left")
	}
	if detail.ListenerPorts.ListenerPort[0] != 80 {
		t.Fatal("TestEnsureLoadBalancer error, expected to be port 80")
	}
}

// EnsureLoadBalancer and HTTPS with UpdateLoadBalancer.
func TestEnsureLoadBalancerHTTPS(t *testing.T) {
	service := newHTTPSSerice()
	nodes := []*v1.Node{
		newNode1(),
	}
	base := newBaseLoadbalancer()
	detail := loadbalancerAttrib(&base[0])
	t.Log(PrettyJson(detail))
	// New Mock cloud to test
	cloud, err := newMockCloud(newMockClientSLB(service, nodes, &base, detail), nil, newMockClientInstanceSDK(node1), nil)

	if err != nil {
		t.Fatal(fmt.Sprintf("TestEnsureLoadBalancer error newCloud: %s\n", err.Error()))
	}
	status, e := cloud.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerHTTPS error: %s\n", e.Error())
	}
	t.Log(PrettyJson(status))
	t.Log(PrettyJson(detail))
	if len(detail.ListenerPortsAndProtocol.ListenerPortAndProtocol) != 1 {
		t.Fatal("TestEnsureLoadBalancerHTTPS error, expected only 1 listener port left")
	}
	if detail.ListenerPortsAndProtocol.ListenerPortAndProtocol[0].ListenerProtocol != "https" {
		t.Fatal("TestEnsureLoadBalancerHTTPS error, expected to be protocol https")
	}
}

func TestEnsureLoadBalancerWithPortChange(t *testing.T) {
	service := newPortRangeService()
	nodes := []*v1.Node{
		newNode1(),
	}

	base := newBaseLoadbalancer()
	detail := loadbalancerAttrib(&base[0])
	grp := vgroups{
		{
			NamedKey:       &NamedKey{CID: CLUSTER_ID, ServiceName: service.Name, Namespace: service.Namespace, Port: listenPort1},
			LoadBalancerId: detail.LoadBalancerId,
			RegionId:       detail.RegionId,
			VGroupId:       "v-idvgroup",
		},
	}
	t.Log(PrettyJson(detail))
	// New Mock cloud to test
	cloud, err := newMockCloud(newMockClientSLB(service, nodes, &base, detail), nil, newMockClientInstanceSDK(node1), nil)

	if err != nil {
		t.Fatal(fmt.Sprintf("TestEnsureLoadBalancer error newCloud: %s\n", err.Error()))
	}

	status, e := cloud.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancer error: %s\n", e.Error())
	}
	t.Log("--------------------- Status ---------------------")
	t.Log(PrettyJson(status))

	t.Log("+++++++++++++++++++++ Detail +++++++++++++++++++++++")
	t.Log(PrettyJson(detail))

	if len(detail.ListenerPorts.ListenerPort) != 2 {
		t.Fatal("TestEnsureLoadBalancerWithPortChange error, expected only 1 listener port left")
	}
	if !Contains(detail.ListenerPorts.ListenerPort, 80) {
		t.Fatal("TestEnsureLoadBalancerWithPortChange error, expected to be port 80")
	}
	if !Contains(detail.ListenerPorts.ListenerPort, 443) {
		t.Fatal("TestEnsureLoadBalancerWithPortChange error, expected to be port 443")
	}
}

//
//func TestEnsureLoadBalancerHTTPSHealthCheck(t *testing.T) {
//	c, err := newMockCloud(nil,nil,nil,nil)
//	if err != nil {
//		t.Errorf("TestEnsureLoadBalancerHTTPSHealthCheck error newCloud: %s\n", err.Error())
//	}
//
//	service := &v1.ServiceName{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "my-service",
//			UID:  types.UID(serviceUID),
//			Annotations: map[string]string{
//				ServiceAnnotationLoadBalancerProtocolPort:           fmt.Sprintf("https:%d", listenPort1),
//				ServiceAnnotationLoadBalancerCertID:                 certID,
//				ServiceAnnotationLoadBalancerHealthCheckFlag:        string(slb.OnFlag),
//				ServiceAnnotationLoadBalancerHealthCheckURI:         "/v2/check",
//				ServiceAnnotationLoadBalancerHealthCheckConnectPort: targetPort1.String(),
//			},
//		},
//		Spec: v1.ServiceSpec{
//			Ports: []v1.ServicePort{
//				{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
//			},
//			Type:            v1.ServiceTypeLoadBalancer,
//			SessionAffinity: v1.ServiceAffinityNone,
//		},
//	}
//	nodes := []*v1.Node{
//		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
//	}
//
//	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
//	if e != nil {
//		t.Errorf("TestEnsureLoadBalancerHTTPS error: %s\n", e.Error())
//	}
//}
//
//
//

func TestEnsureLoadbalancerDeleted(t *testing.T) {
	service := newReadyService()

	base := newBaseLoadbalancer()
	// New Mock cloud to test
	cloud, err := newMockCloud(newMockClientSLB(service, nil, &base, nil), nil, newMockClientInstanceSDK(node1), nil)

	if err != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error newCloud: %s\n", err.Error())
	}

	e := cloud.EnsureLoadBalancerDeleted(clusterName, service)
	if e != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error: %s\n", e.Error())
	}

	t.Log(PrettyJson(base))

	if len(base) != 0 {
		t.Fatal("TestEnsureLoadbalancerDeleted error: base length must equal to 0")
	}
}

func TestEnsureLoadBalancerDeleteWithUserDefined(t *testing.T) {
	base := newBaseLoadbalancer()
	service := newReadyService()
	t.Log(PrettyJson(base))
	// New Mock cloud to test
	cloud, err := newMockCloud(newMockClientSLB(service, nil, &base, nil), nil, nil, nil)
	if err != nil {
		t.Fatal(fmt.Sprintf("TestEnsureLoadBalancerDeleteWithUserDefined error newCloud: %s\n", err.Error()))
	}

	e := cloud.EnsureLoadBalancerDeleted(clusterName, service)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerDeleteWithUserDefined error: %s\n", e.Error())
	}
	t.Log(PrettyJson(base))
	if len(base) != 0 {
		t.Fatal("TestEnsureLoadBalancerDeleteWithUserDefined error, expected not to be deleted")
	}
}

func TestNodeAddressAndInstanceID(t *testing.T) {
	// New Mock cloud to test
	cloud, err := newMockCloud(nil, nil, newMockClientInstanceSDK(node1), nil)

	if err != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: newcloud %s\n", err.Error())
	}
	providerID := nodeid(regionId, node1)
	n, e := cloud.NodeAddresses(types.NodeName(providerID))
	if e != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: node address %s\n", e.Error())
	}
	t.Log(PrettyJson(n))

	if len(n) != 1 {
		t.Fatal("TestNodeAddressAndInstanceID error: node address returned must equal to 1")
	}
	if n[0].Type != "InternalIP" || n[0].Address != "192.168.211.130" {
		t.Fatal("TestNodeAddressAndInstanceID error: node address must equal to 192.168.211.130 InternalIP")
	}

	n, e = cloud.NodeAddressesByProviderID(providerID)
	if e != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: node address %s\n", e.Error())
	}
	if len(n) != 1 {
		t.Fatal("TestNodeAddressAndInstanceID error: node address returned must equal to 1")
	}
	if n[0].Type != "InternalIP" || n[0].Address != "192.168.211.130" {
		t.Fatal("TestNodeAddressAndInstanceID error: node address must equal to 192.168.211.130 InternalIP")
	}

	id, err := cloud.InstanceID(types.NodeName(providerID))
	if err != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: instanceid %s\n", err.Error())
	}
	t.Log(id)
	if id != node1 {
		t.Fatalf("TestNodeAddressAndInstanceID error: instance id must equal to %s" + node1)
	}
	iType, err := cloud.InstanceTypeByProviderID(providerID)
	if err != nil {
		t.Fatal(err)
	}
	if "ecs.sn1ne.large" != iType {
		t.Fatalf("TestNodeAddressAndInstanceID error: instance type should be %s", "ecs.sn1ne.large")
	}

}

var con string = `
{
    "global": {
     "accessKeyID": "{{ access_key_id }}",
     "accessKeySecret": "{{ access_key_secret }}",
     "kubernetesClusterTag": "{{ region_id }}"
   }
 }
 `
