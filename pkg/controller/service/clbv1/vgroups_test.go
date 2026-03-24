package clbv1

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/klog/v2/klogr"
)

func TestVGroupManager_BatchSyncVServerGroupBackendServers(t *testing.T) {
	vgroupManager, _ := getTestVGroupManager()

	var local, remote []model.BackendAttribute
	for i := 0; i < 200; i++ {
		remote = append(remote, model.BackendAttribute{
			ServerIp:    fmt.Sprintf("192.168.0.%d", i),
			ServerId:    fmt.Sprintf("eni-old-%d", i),
			Port:        8080,
			Weight:      100,
			Description: "k8s/svc-1/test/8080/clusterid",
		})
	}

	for i := 0; i < 200; i++ {
		local = append(local, model.BackendAttribute{
			ServerIp:    fmt.Sprintf("192.168.1.%d", i),
			ServerId:    fmt.Sprintf("eni-new-%d", i),
			Port:        8080,
			Weight:      100,
			Description: "k8s/svc-1/test/8080/clusterid",
		})
	}

	localVGroup := model.VServerGroup{
		VGroupId:   "rsp-test",
		VGroupName: "k8s/svc-1/test/8080/clusterid",
		Backends:   local,
	}
	remoteVGroup := model.VServerGroup{
		VGroupId:   "rsp-test",
		VGroupName: "k8s/svc-1/test/8080/clusterid",
		Backends:   remote,
	}
	ctx := context.WithValue(context.TODO(), dryrun.ContextKey("vgroup"), &remoteVGroup)
	reqCtx := &svcCtx.RequestContext{
		Ctx: ctx,
		Log: klogr.New(),
	}

	err := vgroupManager.UpdateVServerGroup(reqCtx, localVGroup, remoteVGroup)
	if err != nil {
		t.Error(err)
	}

	vgroup, ok := ctx.Value(dryrun.ContextKey("vgroup")).(*model.VServerGroup)
	if !ok {
		t.Errorf("no vgroup in context, %+v", ctx.Value("vGroup"))
	}

	assert.Equal(t, len(vgroup.Backends), 200)
}

func TestGetVGroupIDs(t *testing.T) {
	cases := []struct {
		name        string
		annotation  string
		expected    []string
		expectError bool
	}{
		{
			name:        "empty annotation",
			annotation:  "",
			expected:    nil,
			expectError: false,
		},
		{
			name:        "single vgroup",
			annotation:  "vsg-123:80",
			expected:    []string{"vsg-123"},
			expectError: false,
		},
		{
			name:        "multiple vgroups",
			annotation:  "vsg-123:80,vsg-456:443",
			expected:    []string{"vsg-123", "vsg-456"},
			expectError: false,
		},
		{
			name:        "invalid format",
			annotation:  "vsg-123",
			expected:    nil,
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := getVGroupIDs(c.annotation)
			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, result)
			}
		})
	}
}

func TestContainsIP(t *testing.T) {
	_, cidr1, _ := net.ParseCIDR("192.168.0.0/24")
	_, cidr2, _ := net.ParseCIDR("10.0.0.0/8")
	cidrs := []*net.IPNet{cidr1, cidr2}

	cases := []struct {
		name     string
		serverIp string
		expected bool
	}{
		{
			name:     "ip in first cidr",
			serverIp: "192.168.0.1",
			expected: true,
		},
		{
			name:     "ip in second cidr",
			serverIp: "10.0.0.1",
			expected: true,
		},
		{
			name:     "ip not in cidrs",
			serverIp: "172.16.0.1",
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := containsIP(cidrs, c.serverIp)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestPodNumberAlgorithm(t *testing.T) {
	t.Run("eni mode", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "eni-1", Weight: 50},
			{ServerId: "eni-2", Weight: 50},
		}
		result := podNumberAlgorithm(helper.ENITrafficPolicy, backends)
		assert.Equal(t, DefaultServerWeight, result[0].Weight)
		assert.Equal(t, DefaultServerWeight, result[1].Weight)
	})

	t.Run("cluster mode", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "ecs-1", Weight: 50},
			{ServerId: "ecs-2", Weight: 50},
		}
		result := podNumberAlgorithm(helper.ClusterTrafficPolicy, backends)
		assert.Equal(t, DefaultServerWeight, result[0].Weight)
		assert.Equal(t, DefaultServerWeight, result[1].Weight)
	})

	t.Run("local mode", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "ecs-1", Weight: 0, Type: model.ECSBackendType},
			{ServerId: "ecs-1", Weight: 0, Type: model.ECSBackendType},
			{ServerId: "ecs-2", Weight: 0, Type: model.ECSBackendType},
		}
		result := podNumberAlgorithm(helper.LocalTrafficPolicy, backends)
		assert.Equal(t, 2, result[0].Weight)
		assert.Equal(t, 2, result[1].Weight)
		assert.Equal(t, 1, result[2].Weight)
	})
}

func TestPodPercentAlgorithm(t *testing.T) {
	t.Run("empty backends", func(t *testing.T) {
		backends := []model.BackendAttribute{}
		result := podPercentAlgorithm(helper.ENITrafficPolicy, backends, 100)
		assert.Equal(t, 0, len(result))
	})

	t.Run("zero weight", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "eni-1"},
			{ServerId: "eni-2"},
		}
		result := podPercentAlgorithm(helper.ENITrafficPolicy, backends, 0)
		assert.Equal(t, 0, result[0].Weight)
		assert.Equal(t, 0, result[1].Weight)
	})

	t.Run("eni mode", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "eni-1"},
			{ServerId: "eni-2"},
		}
		result := podPercentAlgorithm(helper.ENITrafficPolicy, backends, 100)
		assert.Equal(t, 50, result[0].Weight)
		assert.Equal(t, 50, result[1].Weight)
	})

	t.Run("cluster mode", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "ecs-1"},
			{ServerId: "ecs-2"},
			{ServerId: "ecs-3"},
			{ServerId: "ecs-4"},
		}
		result := podPercentAlgorithm(helper.ClusterTrafficPolicy, backends, 100)
		assert.Equal(t, 25, result[0].Weight)
	})

	t.Run("local mode", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "ecs-1", Type: model.ECSBackendType},
			{ServerId: "ecs-1", Type: model.ECSBackendType},
			{ServerId: "ecs-2", Type: model.ECSBackendType},
		}
		result := podPercentAlgorithm(helper.LocalTrafficPolicy, backends, 100)
		assert.Equal(t, 66, result[0].Weight)
		assert.Equal(t, 66, result[1].Weight)
		assert.Equal(t, 33, result[2].Weight)
	})

	t.Run("minimum weight", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "eni-1"},
			{ServerId: "eni-2"},
		}
		result := podPercentAlgorithm(helper.ENITrafficPolicy, backends, 1)
		assert.Equal(t, 1, result[0].Weight)
		assert.Equal(t, 1, result[1].Weight)
	})
}

func TestRemoveDuplicatedECS(t *testing.T) {
	t.Run("no duplicates", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "ecs-1"},
			{ServerId: "ecs-2"},
		}
		result := removeDuplicatedECS(backends)
		assert.Equal(t, 2, len(result))
	})

	t.Run("with duplicates", func(t *testing.T) {
		backends := []model.BackendAttribute{
			{ServerId: "ecs-1", Type: model.ECSBackendType},
			{ServerId: "ecs-1", Type: model.ECSBackendType},
			{ServerId: "ecs-2", Type: model.ECSBackendType},
		}
		result := removeDuplicatedECS(backends)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "ecs-1", result[0].ServerId)
		assert.Equal(t, "ecs-2", result[1].ServerId)
	})

	t.Run("empty backends", func(t *testing.T) {
		backends := []model.BackendAttribute{}
		result := removeDuplicatedECS(backends)
		assert.Equal(t, 0, len(result))
	})
}

func TestDiff(t *testing.T) {
	t.Run("all same", func(t *testing.T) {
		remote := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
			},
		}
		local := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
			},
		}
		add, del, update := diff(remote, local)
		_ = add
		_ = del
		_ = update
		assert.Equal(t, 0, len(add))
		assert.Equal(t, 0, len(del))
		assert.Equal(t, 0, len(update))
	})

	t.Run("add backends", func(t *testing.T) {
		remote := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
			},
		}
		local := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
				{ServerId: "ecs-2", Port: 80, Weight: 100},
			},
		}
		add, del, update := diff(remote, local)
		assert.Equal(t, 1, len(add))
		assert.Equal(t, "ecs-2", add[0].ServerId)
		assert.Equal(t, 0, len(del))
		assert.Equal(t, 0, len(update))
	})

	t.Run("delete backends", func(t *testing.T) {
		remote := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
				{ServerId: "ecs-2", Port: 80, Weight: 100},
			},
		}
		local := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
			},
		}
		add, del, update := diff(remote, local)
		assert.Equal(t, 0, len(add))
		assert.Equal(t, 1, len(del))
		assert.Equal(t, "ecs-2", del[0].ServerId)
		assert.Equal(t, 0, len(update))
	})

	t.Run("update weight", func(t *testing.T) {
		remote := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100},
			},
		}
		local := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 50},
			},
		}
		add, del, update := diff(remote, local)
		assert.Equal(t, 0, len(add))
		assert.Equal(t, 0, len(del))
		assert.Equal(t, 1, len(update))
		assert.Equal(t, 50, update[0].Weight)
	})

	t.Run("skip user managed", func(t *testing.T) {
		remote := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "ecs-1", Port: 80, Weight: 100, IsUserManaged: true},
			},
		}
		local := model.VServerGroup{
			Backends: []model.BackendAttribute{},
		}
		add, del, update := diff(remote, local)
		assert.Equal(t, 0, len(add))
		assert.Equal(t, 0, len(del))
		assert.Equal(t, 0, len(update))
	})

	t.Run("eni backend", func(t *testing.T) {
		remote := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "eni-1", ServerIp: "192.168.0.1", Port: 8080, Type: "eni"},
			},
		}
		local := model.VServerGroup{
			Backends: []model.BackendAttribute{
				{ServerId: "eni-1", ServerIp: "192.168.0.1", Port: 8080, Type: "eni"},
				{ServerId: "eni-2", ServerIp: "192.168.0.2", Port: 8080, Type: "eni"},
			},
		}
		add, del, update := diff(remote, local)
		_ = del
		_ = update
		assert.Equal(t, 1, len(add))
		assert.Equal(t, "eni-2", add[0].ServerId)
	})
}

func TestSetBackendsFromEndpointSlices(t *testing.T) {
	mgr, err := getTestVGroupManager()
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     klogr.New(),
	}

	t.Run("empty endpoint slices", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			EndpointSlices: []discovery.EndpointSlice{},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("ready endpoint with int target port", func(t *testing.T) {
		ready := true
		nodeName := "node-1"
		portName := "http"
		port := int32(8080)
		protocol := v1.ProtocolTCP
		candidates := &backend.EndpointWithENI{
			EndpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-slice",
						Namespace: "default",
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 1, len(backends))
		assert.Equal(t, "10.0.0.1", backends[0].ServerIp)
		assert.Equal(t, 8080, backends[0].Port)
	})

	t.Run("ready endpoint with string target port", func(t *testing.T) {
		ready := true
		nodeName := "node-1"
		portName := "http"
		port := int32(8080)
		protocol := v1.ProtocolTCP
		candidates := &backend.EndpointWithENI{
			EndpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-slice",
						Namespace: "default",
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.2"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromString("http"),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 1, len(backends))
		assert.Equal(t, "10.0.0.2", backends[0].ServerIp)
		assert.Equal(t, 8080, backends[0].Port)
	})

	t.Run("terminating endpoint ignored", func(t *testing.T) {
		ready := true
		terminating := true
		nodeName := "node-1"
		portName := "http"
		port := int32(8080)
		protocol := v1.ProtocolTCP
		candidates := &backend.EndpointWithENI{
			EndpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-slice",
						Namespace: "default",
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.3"},
							Conditions: discovery.EndpointConditions{
								Ready:       &ready,
								Terminating: &terminating,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("duplicate endpoint addresses", func(t *testing.T) {
		ready := true
		nodeName := "node-1"
		portName := "http"
		port := int32(8080)
		protocol := v1.ProtocolTCP
		candidates := &backend.EndpointWithENI{
			EndpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-slice",
						Namespace: "default",
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.4"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
						{
							Addresses: []string{"10.0.0.4"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 1, len(backends))
	})

	t.Run("endpoint without target ref", func(t *testing.T) {
		ready := false
		nodeName := "node-1"
		portName := "http"
		port := int32(8080)
		protocol := v1.ProtocolTCP
		candidates := &backend.EndpointWithENI{
			EndpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-slice",
						Namespace: "default",
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.5"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName:  &nodeName,
							TargetRef: nil,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})
}

// TestSetBackendsFromEndpoints tests the setBackendsFromEndpoints function
func TestSetBackendsFromEndpoints(t *testing.T) {
	mgr, err := getTestVGroupManager()
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     klogr.New(),
	}

	t.Run("empty endpoints", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("ready endpoint with int target port", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.1",
								NodeName: stringPtr("node-1"),
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 1, len(backends))
		assert.Equal(t, "10.0.0.1", backends[0].ServerIp)
		assert.Equal(t, 8080, backends[0].Port)
	})

	t.Run("ready endpoint with string target port", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.2",
								NodeName: stringPtr("node-1"),
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromString("http"),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 1, len(backends))
		assert.Equal(t, "10.0.0.2", backends[0].ServerIp)
		assert.Equal(t, 8080, backends[0].Port)
	})

	t.Run("not ready endpoint without target ref", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:        "10.0.0.3",
								NodeName:  stringPtr("node-1"),
								TargetRef: nil,
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("not ready endpoint with pod but no readiness gate", func(t *testing.T) {
		// Create a fake kube client with a pod
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
		}
		kubeClient := getFakeKubeClient()
		err := kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		mgr, err := NewVGroupManager(kubeClient, getMockCloudProvider())
		assert.NoError(t, err)

		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.4",
								NodeName: stringPtr("node-1"),
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "test-pod",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("port name not found", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.5",
								NodeName: stringPtr("node-1"),
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "other",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromString("http"),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("not ready endpoint with non-pod target ref", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.6",
								NodeName: stringPtr("node-1"),
								TargetRef: &v1.ObjectReference{
									Kind:      "Node",
									Namespace: "default",
									Name:      "test-node",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 0, len(backends))
	})

	t.Run("not ready endpoint with pod not found", func(t *testing.T) {
		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.7",
								NodeName: stringPtr("node-1"),
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "not-found-pod",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.True(t, containsPotential) // Should be true when pod not found
		assert.Equal(t, 0, len(backends))
	})

	t.Run("not ready endpoint with pod with readiness gate but containers not ready", func(t *testing.T) {
		// Create a fake kube client with a pod that has readiness gate but containers not ready
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-not-ready",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				ReadinessGates: []v1.PodReadinessGate{
					{
						ConditionType: v1.PodConditionType(helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, "test")),
					},
				},
			},
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{
						Type:   v1.ContainersReady,
						Status: v1.ConditionFalse,
					},
				},
			},
		}
		kubeClient := getFakeKubeClient()
		err := kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		mgr, err := NewVGroupManager(kubeClient, getMockCloudProvider())
		assert.NoError(t, err)

		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.8",
								NodeName: stringPtr("node-1"),
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "test-pod-not-ready",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.True(t, containsPotential) // Should be true when containers not ready
		assert.Equal(t, 0, len(backends))
	})

	t.Run("not ready endpoint with pod with readiness gate and containers ready", func(t *testing.T) {
		// Create a fake kube client with a pod that has readiness gate and containers ready
		readinessGateName := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, "test")
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-ready",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				ReadinessGates: []v1.PodReadinessGate{
					{
						ConditionType: v1.PodConditionType(readinessGateName),
					},
				},
			},
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{
						Type:   v1.ContainersReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		kubeClient := getFakeKubeClient()
		err := kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		mgr, err := NewVGroupManager(kubeClient, getMockCloudProvider())
		assert.NoError(t, err)

		candidates := &backend.EndpointWithENI{
			Endpoints: &v1.Endpoints{
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.9",
								NodeName: stringPtr("node-1"),
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "test-pod-ready",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		vgroup := model.VServerGroup{
			VGroupName: "test-vgroup",
			ServicePort: v1.ServicePort{
				Port:       80,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
			},
		}
		backends, containsPotential, err := mgr.setBackendsFromEndpoints(reqCtx, candidates, vgroup)
		assert.NoError(t, err)
		assert.False(t, containsPotential)
		assert.Equal(t, 1, len(backends))
		assert.Equal(t, "10.0.0.9", backends[0].ServerIp)
	})
}

func stringPtr(s string) *string {
	return &s
}

// TestGetVgroupById tests the getVgroupById function
func TestGetVgroupById(t *testing.T) {
	mgr, err := getTestVGroupManager()
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Annotations: map[string]string{
				annotation.Annotation(annotation.LoadBalancerId): "lb-exist-id",
			},
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     klogr.New(),
	}

	t.Run("find existing vgroup", func(t *testing.T) {
		vgroup, err := getVgroupById(mgr, reqCtx, "rsp-tcp-80")
		assert.NoError(t, err)
		assert.NotNil(t, vgroup)
		assert.Equal(t, "rsp-tcp-80", vgroup.VGroupId)
	})

	t.Run("find non-existing vgroup", func(t *testing.T) {
		vgroup, err := getVgroupById(mgr, reqCtx, "rsp-not-exist")
		assert.NoError(t, err)
		assert.Nil(t, vgroup)
	})

	t.Run("find another existing vgroup", func(t *testing.T) {
		vgroup, err := getVgroupById(mgr, reqCtx, "rsp-udp-53")
		assert.NoError(t, err)
		assert.NotNil(t, vgroup)
		assert.Equal(t, "rsp-udp-53", vgroup.VGroupId)
	})
}

func TestVGroupManager_BuildLocalModel(t *testing.T) {
	mgr, err := getTestVGroupManager()
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)
		m := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{},
		}
		err = mgr.BuildLocalModel(reqCtx, m)
		assert.NoError(t, err)
		assert.NotNil(t, m.VServerGroups)
	})

	t.Run("with ENI backends", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[annotation.Annotation(annotation.BackendType)] = model.ENIBackendType
		reqCtx := getReqCtx(svc)
		m := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{},
		}
		err = mgr.BuildLocalModel(reqCtx, m)
		assert.NoError(t, err)
		assert.NotNil(t, m.VServerGroups)
	})

	t.Run("updateVServerGroupENIBackendID DescribeNetworkInterfaces error", func(t *testing.T) {
		oldEnv := os.Getenv("SERVICE_FORCE_BACKEND_ENI")
		defer func() { _ = os.Setenv("SERVICE_FORCE_BACKEND_ENI", oldEnv) }()
		_ = os.Setenv("SERVICE_FORCE_BACKEND_ENI", "true")
		cloud := prvd.Provider(vmock.MockCloud{
			MockECS:   vmock.NewMockECS(nil),
			MockVPC:   vmock.NewMockVPC(nil),
			IMetaData: vmock.NewMockMetaData("vpc-describe-eni-error"),
		})
		eniMgr, err := NewVGroupManager(getFakeKubeClient(), cloud)
		assert.NoError(t, err)
		svc := getDefaultService()
		svc.Annotations[annotation.Annotation(annotation.BackendType)] = model.ENIBackendType
		reqCtx := getReqCtx(svc)
		m := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{},
		}
		err = eniMgr.BuildLocalModel(reqCtx, m)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "DescribeNetworkInterfaces")
		}
	})
}
