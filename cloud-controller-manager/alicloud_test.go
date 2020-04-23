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
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/denverdino/aliyungo/slb"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var keyid string
var keysecret string

var (
	vpcId                     = "vpc-2zeaybwqmvn6qgabfd3pe"
	regionId                  = "cn-beijing"
	zoneId                    = "cn-beijing-b"
	certID                    = "1745547945134207_157f665c830"
	listenPort1         int32 = 80
	listenPort2         int32 = 90
	targetPort1               = intstr.FromInt(8080)
	targetPort2               = intstr.FromInt(9090)
	nodePort1           int32 = 8080
	nodePort2           int32 = 9090
	protocolTcp               = v1.ProtocolTCP
	protocolUdp               = v1.ProtocolUDP
	node1                     = "i-bp1bcl00jhxr754tw8vx"
	node2                     = "i-bp1bcl00jhxr754tw8vy"
	clusterName               = "clusterName-random"
	serviceUIDNoneExist       = "UID-1234567890-0987654321-1234556"
	serviceUIDExist           = "c83f8bed-812e-11e9-a0ad-00163e0a3984"
	nodeName                  = "iZuf694l8lw6xvdx6gh7tkZ"
)

func TestInterface2Slice(t *testing.T) {
	am := []string{"a", "b"}

	err := Batch(
		am, 10,
		func(o []interface{}) error {
			for _, i := range o {
				m, ok := i.(string)
				if !ok {
					return fmt.Errorf("not string")
				}
				fmt.Printf("i:::: %s", m)
			}

			return nil
		},
	)
	if err != nil {
		fmt.Printf("batch error: %s", err.Error())
	}
}

// EnsureBasicLoadBalancer Configure
func TestEnsureLoadBalancerBasic(t *testing.T) {

	// Test case Example
	// Name: TestEnsureLoadBalancerBasic
	// Step 1: define your node & service object.
	prid := nodeid(string(REGION), INSTANCEID)
	prid2 := nodeid(string(REGION), INSTANCEID2)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "basic-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Port:       listenPort1,
						TargetPort: targetPort1,
						Protocol:   v1.ProtocolTCP,
						NodePort:   nodePort1,
					},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid2},
				Spec: v1.NodeSpec{
					ProviderID: prid2,
				},
			},
		},
	).WithEndpoints(
		&v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "basic-service",
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
		},
	)

	// Step2: Run start a test on specified function.
	f.RunDefault(t, "Basic LoadBalancer Creation Test")
}

// Test Http configuration.
func TestEnsureLoadBalancerAnnotation(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerProtocolPort: "http:80",
					ServiceAnnotationLoadBalancerAddressType:  string(slb.InternetAddressType),
					//ServiceAnnotationLoadBalancerVswitch: 		VSWITCH_ID,
					ServiceAnnotationLoadBalancerForwardPort: "80:443",
					//ServiceAnnotationLoadBalancerSLBNetworkType: "classic",
					ServiceAnnotationLoadBalancerChargeType: string(slb.PayByTraffic),
					//ServiceAnnotationLoadBalancerId: "ic",
					//ServiceAnnotationLoadBalancerBackendLabel: "key=value",
					ServiceAnnotationLoadBalancerRegion:       string(REGION),
					ServiceAnnotationLoadBalancerMasterZoneID: string(REGION_A),
					ServiceAnnotationLoadBalancerSlaveZoneID:  string(REGION_A),
					ServiceAnnotationLoadBalancerBandwidth:    "70",
					ServiceAnnotationLoadBalancerScheduler:    "wlc",

					//acl
					ServiceAnnotationLoadBalancerAclType:   "white",
					ServiceAnnotationLoadBalancerAclID:     "acl-idxxx",
					ServiceAnnotationLoadBalancerAclStatus: "on",
					//ServiceAnnotationLoadBalancerCertID:	   "certid",
					ServiceAnnotationLoadBalancerHealthCheckFlag:               "off",
					ServiceAnnotationLoadBalancerHealthCheckType:               "tcp",
					ServiceAnnotationLoadBalancerHealthCheckURI:                "/v1/check",
					ServiceAnnotationLoadBalancerHealthCheckConnectPort:        "80",
					ServiceAnnotationLoadBalancerHealthCheckHealthyThreshold:   "20",
					ServiceAnnotationLoadBalancerHealthCheckUnhealthyThreshold: "5",
					ServiceAnnotationLoadBalancerHealthCheckInterval:           "5",
					ServiceAnnotationLoadBalancerHealthCheckConnectTimeout:     "5",
					ServiceAnnotationLoadBalancerHealthCheckTimeout:            "5",
					ServiceAnnotationLoadBalancerHealthCheckDomain:             "aliyun.com",
					ServiceAnnotationLoadBalancerHealthCheckHTTPCode:           "200",
					ServiceAnnotationLoadBalancerAdditionalTags:                "k1=v1,k2=v2",
					ServiceAnnotationLoadBalancerOverrideListener:              "true",
					ServiceAnnotationLoadBalancerSpec:                          "lb-mini-spec",
					ServiceAnnotationLoadBalancerSessionStick:                  "on",
					ServiceAnnotationLoadBalancerSessionStickType:              "cookie",
					ServiceAnnotationLoadBalancerCookieTimeout:                 "5000",
					ServiceAnnotationLoadBalancerCookie:                        "none-cookie",
					ServiceAnnotationLoadBalancerPersistenceTimeout:            "7400",
					ServiceAnnotationLoadBalancerIPVersion:                     string(slb.IPv4),
					ServiceAnnotationLoadBalancerPrivateZoneName:               "",
					ServiceAnnotationLoadBalancerPrivateZoneId:                 "",
					ServiceAnnotationLoadBalancerPrivateZoneRecordName:         "",
					ServiceAnnotationLoadBalancerPrivateZoneRecordTTL:          "",
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "TestAnnotation")
}

// Test Http configuration.
func TestEnsureLoadBalancerSpec(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerSpec: "lb-spec-mini",
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "Create Loadbalancer With SPEC")
}

// Test Http configuration.
func TestEnsureLoadBalancerVswitchID(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerVswitch:     VSWITCH_ID,
					ServiceAnnotationLoadBalancerAddressType: string(slb.IntranetAddressType),
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
	)

	f.RunDefault(t, "Create Loadbalancer With VswitchID")
}

// Test Http configuration.
func TestEnsureLoadBalancerBackendLable(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerProtocolPort: "http:80",
					ServiceAnnotationLoadBalancerAddressType:  string(slb.InternetAddressType),
					//ServiceAnnotationLoadBalancerVswitch: 		VSWITCH_ID,
					ServiceAnnotationLoadBalancerBackendLabel: "key=value",
					//ServiceAnnotationLoadBalancerAdditionalTags: "k1=v1,k2=v2",
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: prid,
					Labels: map[string]string{
						"key":     "value",
						"ingress": "true",
					},
				},
				Spec: v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "TestBackendLabel")
}

// Test Http configuration.
func TestEnsureLoadBalancerHTTP(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerProtocolPort: fmt.Sprintf("http:%d", listenPort1),
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
	)

	f.RunDefault(t, "Create HTTPS Loadbalancer")
}

// Test Https configuration.
func TestEnsureLoadBalancerHTTPS(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerProtocolPort: fmt.Sprintf("http:%d", listenPort1),
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "Create HTTPS Loadbalancer")
}

// Test Https configuration.
func TestHTTPSFromHttp(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerProtocolPort: fmt.Sprintf("http:%d", listenPort1),
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "With HTTP Listener")

	f.SVC.Annotations[ServiceAnnotationLoadBalancerCertID] = certID
	f.SVC.Annotations[ServiceAnnotationLoadBalancerProtocolPort] = fmt.Sprintf("https:%d", listenPort1)
	f.RunDefault(t, "Change From Http to Https")
}

func TestEnsureLoadBalancerHTTPSHealthCheck(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
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
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunDefault(t, "HTTPS health check")
}

func TestEnsureLoadBalancerWithENI(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					"service.beta.kubernetes.io/backend-type": "eni",
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type: v1.ServiceTypeLoadBalancer,
			},
		},
	).WithEndpoints(
		&v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "default",
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP:       ENI_ADDR_1,
							NodeName: &prid,
						},
						{
							IP:       ENI_ADDR_2,
							NodeName: &prid,
						},
					},
					Ports: []v1.EndpointPort{
						{
							Port: listenPort1,
						},
					},
				},
			},
		},
	)
	f.RunDefault(t, "ENI backend Type test")
}

func TestEnsureLoadbalancerDeleted(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "my-service",
				Namespace:   "default",
				UID:         types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunCustomized(
		t, "Delete Loadbalancer",
		func(f *FrameWork) error {
			_, err := f.Cloud.EnsureLoadBalancer(context.Background(), CLUSTER_ID, f.SVC, f.Nodes)
			if err != nil {
				t.Fatalf("delete loadbalancer error: create %s", err.Error())
			}
			err = f.Cloud.EnsureLoadBalancerDeleted(context.Background(), CLUSTER_ID, f.SVC)
			if err != nil {
				t.Fatalf("ensure loadbalancer delete error, %s", err.Error())
			}
			exist, _, err := f.LoadBalancer().FindLoadBalancer(context.Background(), f.SVC)
			if err != nil || exist {
				t.Fatalf("Delete LoadBalancer error: %v, %t", err, exist)
			}
			return nil
		},
	)
}

func TestEnsureLoadBalancerDeleteWithUserDefined(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(nil)
	f.WithService(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "https-service",
				Namespace: "default",
				UID:       types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerId: LOADBALANCER_ID,
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
	).WithNodes(
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec:       v1.NodeSpec{ProviderID: prid},
			},
		},
	)

	f.RunCustomized(
		t, "Delete User Defined Loadbalancer",
		func(f *FrameWork) error {
			err := f.Cloud.EnsureLoadBalancerDeleted(context.Background(), CLUSTER_ID, f.SVC)
			if err != nil {
				t.Fatalf("ensure loadbalancer delete error, %s", err.Error())
			}
			exist, _, err := f.LoadBalancer().FindLoadBalancer(context.Background(), f.SVC)
			if err != nil || !exist {
				t.Fatalf("Delete LoadBalancer error: %v, %t", err, exist)
			}
			return nil
		},
	)
}

func TestNodeAddressAndInstanceID(t *testing.T) {

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
	)

	// New Mock Cloud to test
	cloud, err := NewMockCloud()

	if err != nil {
		t.Fatalf("TestNodeAddressAndInstanceID error: newcloud %s\n", err.Error())
	}
	prid := nodeid(string(REGION), INSTANCEID)

	n, e := cloud.NodeAddresses(context.Background(), types.NodeName(prid))
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

	n, e = cloud.NodeAddressesByProviderID(context.Background(), prid)
	if e != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: node address %s\n", e.Error())
	}
	if len(n) != 1 {
		t.Fatal("TestNodeAddressAndInstanceID error: node address returned must equal to 1")
	}
	if n[0].Type != "InternalIP" || n[0].Address != "192.168.211.130" {
		t.Fatal("TestNodeAddressAndInstanceID error: node address must equal to 192.168.211.130 InternalIP")
	}

	id, err := cloud.InstanceID(context.Background(), types.NodeName(prid))
	if err != nil {
		t.Errorf("TestNodeAddressAndInstanceID error: instanceid %s\n", err.Error())
	}
	t.Log(id)
	if id != INSTANCEID {
		t.Fatalf("TestNodeAddressAndInstanceID error: instance id must equal to %s" + node1)
	}
	iType, err := cloud.InstanceTypeByProviderID(context.Background(), prid)
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
