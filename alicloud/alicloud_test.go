package alicloud

import (
	"encoding/json"
	"fmt"
	"github.com/denverdino/aliyungo/slb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/api/v1"
	"strings"
	"testing"
)

var keyid string = "LTAIdwVoOES5YYj2"
var keysecret string = "0NBaap8GkqBpo9TnJ7svT2sNDQ9MzZ"

var (
	listenPort1 int32 = 80
	listenPort2 int32 = 90
	targetPort1       = intstr.FromInt(8080)
	targetPort2       = intstr.FromInt(9090)
	nodePort1   int32 = 8080
	nodePort2   int32 = 9090
	protocolTcp       = v1.ProtocolTCP
	protocolUdp       = v1.ProtocolUDP
	node1             = "i-bp1bcl00jhxr754tw8vy"
	node2             = "i-bp1bcl00jhxr754tw8vy"
	clusterName       = "clusterName-random"
	serviceUID        = "UID-1234567890-0987654321-1234556"
	certID            = "1745547945134207_157f665c830"
	nodeName          = "iZuf694l8lw6xvdx6gh7tkZ"
)

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

func TestCloudConfig(t *testing.T) {
	_, err := newCloud()
	if err != nil {
		t.Errorf("newAliCloud error: %s\n", err.Error())
	}
}

func TestEnsureLoadBalancer(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancer error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
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
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancer error: %s\n", e.Error())
	}
}

func TestEnsureLoadBalancerHTTPS(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancerHTTPS error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
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
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerHTTPS error: %s\n", e.Error())
	}
}

func TestEnsureLoadBalancerHTTPSHealthCheck(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancerHTTPSHealthCheck error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
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
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerHTTPS error: %s\n", e.Error())
	}
}

func TestEnsureLoadBalancerHTTP(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancerHTTP error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
			Annotations: map[string]string{
				ServiceAnnotationLoadBalancerProtocolPort:    fmt.Sprintf("http:%d", listenPort1),
				ServiceAnnotationLoadBalancerHealthCheckFlag: "on",
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
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1, GenerateName: node1},
			Spec: v1.NodeSpec{ExternalID: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerHTTP error: %s\n", e.Error())
	}
}

func TestEnsureLoadBalancerWithListenPortChange(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancerWithListenPortChange error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort2, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerWithListenPortChange error: %s\n", e.Error())
	}
}

func TestEnsureLoadBalancerWithTargetPortChange(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancerWithTargetPortChange error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort2, TargetPort: targetPort2, Protocol: v1.ProtocolTCP, NodePort: nodePort2},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerWithTargetPortChange error: %s\n", e.Error())
	}
}

func TestEnsureLoadBalancerWithProtocolChange(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadBalancerWithTargetPortChange error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort1, TargetPort: targetPort2, Protocol: v1.ProtocolUDP, NodePort: nodePort2},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	_, e := c.EnsureLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureLoadBalancerWithTargetPortChange error: %s\n", e.Error())
	}
}

func TestUpdateLoadbalancer(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureBackendServers error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort2, TargetPort: targetPort2, Protocol: v1.ProtocolTCP, NodePort: nodePort2},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
	nodes := []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: node1}},
	}

	e := c.UpdateLoadBalancer(clusterName, service, nodes)
	if e != nil {
		t.Errorf("TestEnsureBackendServers error: %s\n", e.Error())
	}
}

func TestEnsureLoadbalancerDeleted(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error newCloud: %s\n", err.Error())
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-service",
			UID:  types.UID(serviceUID),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort2, TargetPort: targetPort2, Protocol: v1.ProtocolTCP, NodePort: nodePort2},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}

	e := c.EnsureLoadBalancerDeleted(clusterName, service)
	if e != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error: %s\n", e.Error())
	}
}

func TestNodeAddress(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error newCloud: %s\n", err.Error())
	}
	nodeName := types.NodeName("cd557aaeacb30474d90b8149403bb7611-node1")
	n, e := c.NodeAddresses(nodeName)
	if e != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error: %s\n", e.Error())
	}
	fmt.Printf("NodeAddress: %+v", n)
}

func TestFindInstance(t *testing.T) {
	c, err := newCloud()
	if err != nil {
		t.Errorf("TestEnsureLoadbalancerDeleted error newCloud: %s\n", err.Error())
	}
	nodeName := types.NodeName("c8330c59708d84f819a1ca0cd4c9bfd2b-node1")
	id, err := c.InstanceID(nodeName)
	if err != nil {
		t.Fatal("Instance Error: %s\n", err.Error())
	}
	fmt.Printf("%+v\n", id)
}
func newCloud() (*Cloud, error) {
	cfg := &CloudConfig{
		Global: struct {
			KubernetesClusterTag string

			AccessKeyID     string `json:"accessKeyID"`
			AccessKeySecret string `json:"accessKeySecret"`
			Region          string `json:"region"`
		}{
			AccessKeyID:     "",
			AccessKeySecret: "",
			Region:          "cn-zhangjiakou",
		},
	}
	return newAliCloud(cfg)
}

var vice string = `
{
    "kind": "Service",
    "apiVersion": "v1",
    "metadata": {
        "name": "my-service"
    },
    "spec": {
        "selector": {
            "app": "MyApp"
        },
        "ports": [
            {
                "protocol": "TCP",
                "port": 80,
                "targetPort": 9376
            }
        ],
        "type": "LoadBalancer"
    }
}
`

var con string = `
{
    "global": {
     "accessKeyID": "{{ access_key_id }}",
     "accessKeySecret": "{{ access_key_secret }}",
     "kubernetesClusterTag": "{{ region_id }}"
   }
 }
 `
