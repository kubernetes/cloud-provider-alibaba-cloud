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

	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var keyid string
var keysecret string

var (
	vpcId             = "vpc-2zeaybwqmvn6qgabfd3pe"
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

func newMockCloud() (*Cloud, error) {
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
				return "", errors.New("not found")
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

func TestEnsureLoadBalancerBasic(t *testing.T) {

	// Test case Example
	// Name: TestEnsureLoadBalancerBasic
	// Step 1: define your node & service object.
	svc := &v1.Service{
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
	prid := nodeid(string(REGION), INSTANCEID)
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: prid},
			Spec: v1.NodeSpec{
				ProviderID: prid,
			},
		},
	}

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

	// Step3: New Mock cloud to test
	cloud, err := newMockCloud()
	if err != nil {
		t.Fatal(fmt.Sprintf("TestEnsureLoadBalancer error newCloud: %s\n", err.Error()))
	}

	// Step4: Call the func which you want to test. eg. EnsureLoadBalancer
	status, e := cloud.EnsureLoadBalancer(clusterName, svc, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancer error: %s\n", e.Error())
	}

	// Step5: do the check.
	exist, mlb, err := cloud.climgr.loadbalancer.findLoadBalancer(svc)
	if err != nil || !exist {
		t.Fatalf("slb not found, %v, %v", exist, err)
	}

	if len(mlb.ListenerPorts.ListenerPort) != 1 {
		t.Fatal("TestEnsureLoadBalancer error, expected only 1 listener port left")
	}
	if mlb.ListenerPorts.ListenerPort[0] != 80 {
		t.Fatalf("TestEnsureLoadBalancer error, expected to be port 80, Got %d", mlb.ListenerPorts.ListenerPort[0])
	}
	if len(status.Ingress) <= 0 {
		t.Fatalf("slb ingress <= 0, %v", status.Ingress)
	}

	args := slb.DescribeVServerGroupsArgs{
		LoadBalancerId: mlb.LoadBalancerId,
	}
	res, err := cloud.climgr.loadbalancer.c.DescribeVServerGroups(&args)
	if err != nil {
		t.Fatalf("vserver group: %v", err)
	}
	if len(res.VServerGroups.VServerGroup) != 1 {
		t.Fatalf("vserver group must be 1")
	}
	varg := slb.DescribeVServerGroupAttributeArgs{
		VServerGroupId: res.VServerGroups.VServerGroup[0].VServerGroupId,
	}
	vg, err := cloud.climgr.loadbalancer.c.DescribeVServerGroupAttribute(&varg)
	if err != nil {
		t.Fatalf("vserver group attribute error: %v", err)
	}
	if len(vg.BackendServers.BackendServer) != 1 {
		t.Fatalf("vgroup backend server must be 1")
	}

	if vg.BackendServers.BackendServer[0].ServerId != INSTANCEID {
		t.Fatalf("vgroup backend id error.")
	}
}

// EnsureLoadBalancer and HTTPS with UpdateLoadBalancer.
func TestEnsureLoadBalancerHTTPS(t *testing.T) {
	svc := &v1.Service{
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
	prid := nodeid(string(REGION), INSTANCEID)
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: prid},
			Spec: v1.NodeSpec{
				ProviderID: prid,
			},
		},
	}

	fmt.Printf("svc: %v, node:%v", svc, nodes)
}

func TestEnsureLoadBalancerWithPortChange(t *testing.T) {
	svc := &v1.Service{
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
	prid := nodeid(string(REGION), INSTANCEID)
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: prid},
			Spec: v1.NodeSpec{
				ProviderID: prid,
			},
		},
	}

	fmt.Printf("svc: %v, node:%v", svc, nodes)

}

func TestEnsureLoadBalancerHTTPSHealthCheck(t *testing.T) {

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerProtocolPort:           fmt.Sprintf("https:%d", listenPort1),
				ServiceAnnotationLoadBalancerCertID:                 certID,
				ServiceAnnotationLoadBalancerHealthCheckFlag:        string(slb.OnFlag),
				ServiceAnnotationLoadBalancerHealthCheckURI:         "/v2/check",
				ServiceAnnotationLoadBalancerHealthCheckConnectPort: targetPort1.String(),
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
	prid := nodeid(string(REGION), INSTANCEID)
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: prid},
			Spec: v1.NodeSpec{
				ProviderID: prid,
			},
		},
	}

	fmt.Printf("svc: %v, node:%v", svc, nodes)
}

func TestEnsureLoadbalancerDeleted(t *testing.T) {
	svc := &v1.Service{
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
	prid := nodeid(string(REGION), INSTANCEID)
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: prid},
			Spec: v1.NodeSpec{
				ProviderID: prid,
			},
		},
	}

	fmt.Printf("svc: %v, node:%v", svc, nodes)

}

func TestEnsureLoadBalancerDeleteWithUserDefined(t *testing.T) {
	svc := &v1.Service{
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
	prid := nodeid(string(REGION), INSTANCEID)
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: prid},
			Spec: v1.NodeSpec{
				ProviderID: prid,
			},
		},
	}

	fmt.Printf("svc: %v, node:%v", svc, nodes)

}

func TestNodeAddressAndInstanceID(t *testing.T) {

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

	// New Mock cloud to test
	cloud, err := newMockCloud()

	if err != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: newcloud %s\n", err.Error())
	}
	prid := nodeid(string(REGION), INSTANCEID)

	n, e := cloud.NodeAddresses(types.NodeName(prid))
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

	n, e = cloud.NodeAddressesByProviderID(prid)
	if e != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: node address %s\n", e.Error())
	}
	if len(n) != 1 {
		t.Fatal("TestNodeAddressAndInstanceID error: node address returned must equal to 1")
	}
	if n[0].Type != "InternalIP" || n[0].Address != "192.168.211.130" {
		t.Fatal("TestNodeAddressAndInstanceID error: node address must equal to 192.168.211.130 InternalIP")
	}

	id, err := cloud.InstanceID(types.NodeName(prid))
	if err != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: instanceid %s\n", err.Error())
	}
	t.Log(id)
	if id != INSTANCEID {
		t.Fatalf("TestNodeAddressAndInstanceID error: instance id must equal to %s" + node1)
	}
	iType, err := cloud.InstanceTypeByProviderID(prid)
	if err != nil {
		t.Fatal(err)
	}
	if "ecs.sn1ne.large" != iType {
		t.Fatalf("TestNodeAddressAndInstanceID error: instance type should be %s", "ecs.sn1ne.large")
	}

}

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

var con string = `
{
    "global": {
     "accessKeyID": "{{ access_key_id }}",
     "accessKeySecret": "{{ access_key_secret }}",
     "kubernetesClusterTag": "{{ region_id }}"
   }
 }
 `
