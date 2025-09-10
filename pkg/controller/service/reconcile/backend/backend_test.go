package backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBatch(t *testing.T) {
	sum := 0
	addFunc := func(a []int) error {
		for _, num := range a {
			sum += num
		}
		return nil
	}
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if err := Batch(nums, 3, addFunc); err != nil {
		t.Fatalf("Batch error: %s", err.Error())
	}
	assert.Equal(t, sum, 55)
}

func TestFilterOutByLabel(t *testing.T) {
	nodes := []v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
				Labels: map[string]string{
					"environment": "production",
					"zone":        "zone1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node2",
				Labels: map[string]string{
					"environment": "production",
					"zone":        "zone2",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node3",
				Labels: map[string]string{
					"environment": "development",
					"zone":        "zone1",
				},
			},
		},
	}

	t.Run("filter by single label", func(t *testing.T) {
		result, err := filterOutByLabel(nodes, "environment=production")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, []string{result[0].Name, result[1].Name}, "node1")
		assert.Contains(t, []string{result[0].Name, result[1].Name}, "node2")
	})

	t.Run("filter by multiple labels", func(t *testing.T) {
		result, err := filterOutByLabel(nodes, "environment=production,zone=zone1")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "node1", result[0].Name)
	})

	t.Run("filter by non-existent label", func(t *testing.T) {
		result, err := filterOutByLabel(nodes, "environment=staging")
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("invalid label format", func(t *testing.T) {
		_, err := filterOutByLabel(nodes, "environment")
		assert.Error(t, err)
	})

	t.Run("empty label string", func(t *testing.T) {
		result, err := filterOutByLabel(nodes, "")
		assert.NoError(t, err)
		assert.Len(t, result, 3)
	})
}

func TestNeedExcludeFromLB(t *testing.T) {
	t.Run("master node", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		masterNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-node",
				Labels: map[string]string{
					"node-role.kubernetes.io/master": "",
				},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.True(t, needExcludeFromLB(reqCtx, masterNode))
	})

	t.Run("node with ToBeDeletedTaint", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-to-be-deleted",
			},
			Spec: v1.NodeSpec{
				Taints: []v1.Taint{
					{
						Key:   "ToBeDeletedByClusterAutoscaler",
						Value: "some-value",
					},
				},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.True(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("unschedulable node with remove-unscheduled-backend on", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend": "on",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "unschedulable-node",
			},
			Spec: v1.NodeSpec{
				Unschedulable: true,
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.True(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("unschedulable node with remove-unscheduled-backend not on", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend": "off",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "unschedulable-node",
			},
			Spec: v1.NodeSpec{
				Unschedulable: true,
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.False(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("unschedulable node without remove-unscheduled-backend annotation", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test-service",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "unschedulable-node",
			},
			Spec: v1.NodeSpec{
				Unschedulable: true,
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.False(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("node without conditions", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-without-conditions",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{},
			},
		}
		assert.True(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("not ready node", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-ready-node",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionFalse,
					},
				},
			},
		}
		assert.True(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("vk node", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "vk-node",
				Labels: map[string]string{
					"type": "virtual-kubelet",
				},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.False(t, needExcludeFromLB(reqCtx, node))
	})

	t.Run("normal ready node", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "normal-node",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		assert.False(t, needExcludeFromLB(reqCtx, node))
	})
}

func TestSetTrafficPolicy(t *testing.T) {
	// Create a mock EndpointWithENI
	ep := &EndpointWithENI{}

	t.Run("eni backend type", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/backend-type": "eni",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		ep.setTrafficPolicy(reqCtx)
		assert.Equal(t, helper.ENITrafficPolicy, ep.TrafficPolicy)
	})

	t.Run("local traffic policy", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: v1.ServiceSpec{
				ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		ep.setTrafficPolicy(reqCtx)
		assert.Equal(t, helper.LocalTrafficPolicy, ep.TrafficPolicy)
	})

	t.Run("cluster traffic policy", func(t *testing.T) {
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: v1.ServiceSpec{
				ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeCluster,
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		ep.setTrafficPolicy(reqCtx)
		assert.Equal(t, helper.ClusterTrafficPolicy, ep.TrafficPolicy)
	})
}

func TestSetAddressIPVersion(t *testing.T) {
	t.Run("default ipv4", func(t *testing.T) {
		ep := &EndpointWithENI{}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		ep.setAddressIpVersion(reqCtx)
		assert.Equal(t, model.IPv4, ep.AddressIPVersion)
	})

	t.Run("ipv6 backend version with missing feature gates", func(t *testing.T) {
		ep := &EndpointWithENI{}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-ip-version": "ipv6",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		ep.setAddressIpVersion(reqCtx)
		assert.Equal(t, model.IPv4, ep.AddressIPVersion)
	})

	t.Run("ipv6 backend version with feature gates but no ip version annotation", func(t *testing.T) {
		ep := &EndpointWithENI{}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-ip-version": "ipv6",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		_ = utilfeature.DefaultMutableFeatureGate.Set("IPv6DualStack=true")
		_ = utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=true")
		defer func() {
			_ = utilfeature.DefaultMutableFeatureGate.Set("IPv6DualStack=false")
			_ = utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=false")
		}()

		ep.setAddressIpVersion(reqCtx)
		assert.Equal(t, model.IPv4, ep.AddressIPVersion)
	})

	t.Run("ipv6 backend version with all conditions met", func(t *testing.T) {
		ep := &EndpointWithENI{}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-ip-version": "ipv6",
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version":         "ipv6",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		_ = utilfeature.DefaultMutableFeatureGate.Set("IPv6DualStack=true")
		_ = utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=true")
		defer func() {
			_ = utilfeature.DefaultMutableFeatureGate.Set("IPv6DualStack=false")
			_ = utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=false")
		}()

		ep.setAddressIpVersion(reqCtx)
		assert.Equal(t, model.IPv6, ep.AddressIPVersion)
	})

	t.Run("dualstack for nlb with all conditions met", func(t *testing.T) {
		ep := &EndpointWithENI{}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-ip-version": "ipv6",
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version":         "dualstack",
				},
			},
			Spec: v1.ServiceSpec{
				Type:              v1.ServiceTypeLoadBalancer,
				LoadBalancerClass: pointer.String(helper.NLBClass),
			},
		}
		reqCtx := &svcCtx.RequestContext{
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		_ = utilfeature.DefaultMutableFeatureGate.Set("IPv6DualStack=true")
		_ = utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=true")
		defer func() {
			_ = utilfeature.DefaultMutableFeatureGate.Set("IPv6DualStack=false")
			_ = utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=false")
		}()

		ep.setAddressIpVersion(reqCtx)
		assert.Equal(t, model.IPv6, ep.AddressIPVersion)
	})
}

func TestNewEndpointWithENI(t *testing.T) {
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node2",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
		},
	}

	endpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "192.168.1.1",
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 80,
					},
				},
			},
		},
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}

	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	_ = discovery.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(endpoints, service).
		WithLists(nodeList).
		Build()

	t.Run("successfully create EndpointWithENI", func(t *testing.T) {
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		result, err := NewEndpointWithENI(reqCtx, fakeClient)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Nodes, 2)
		assert.Equal(t, helper.ClusterTrafficPolicy, result.TrafficPolicy)
		assert.Equal(t, model.IPv4, result.AddressIPVersion)
	})

	t.Run("using endpoint", func(t *testing.T) {
		err := utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=false")
		assert.NoError(t, err)
		defer func() {
			err := utilfeature.DefaultMutableFeatureGate.Set("EndpointSlice=true")
			assert.NoError(t, err)
		}()
		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: service,
			Anno:    annotation.NewAnnotationRequest(service),
		}

		result, err := NewEndpointWithENI(reqCtx, fakeClient)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Nodes, 2)
		assert.Equal(t, helper.ClusterTrafficPolicy, result.TrafficPolicy)
		assert.Equal(t, model.IPv4, result.AddressIPVersion)
	})

	t.Run("with eni backend type", func(t *testing.T) {
		eniService := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eni-service",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/backend-type": "eni",
				},
			},
		}

		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: eniService,
			Anno:    annotation.NewAnnotationRequest(eniService),
		}

		result, err := NewEndpointWithENI(reqCtx, fakeClient)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, helper.ENITrafficPolicy, result.TrafficPolicy)
	})

	t.Run("with local external traffic policy", func(t *testing.T) {
		localService := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "local-service",
				Namespace: "default",
			},
			Spec: v1.ServiceSpec{
				ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
			},
		}

		reqCtx := &svcCtx.RequestContext{
			Ctx:     context.Background(),
			Service: localService,
			Anno:    annotation.NewAnnotationRequest(localService),
		}

		result, err := NewEndpointWithENI(reqCtx, fakeClient)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, helper.LocalTrafficPolicy, result.TrafficPolicy)
	})
}

func TestGetNodes(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	readyNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
	}
	t.Run("success", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithLists(&v1.NodeList{Items: []v1.Node{*readyNode}}).Build()
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: service, Anno: annotation.NewAnnotationRequest(service)}
		nodes, err := GetNodes(reqCtx, client)
		assert.NoError(t, err)
		assert.Len(t, nodes, 1)
		assert.Equal(t, "node-1", nodes[0].Name)
	})
	t.Run("with backend label filter", func(t *testing.T) {
		labeledNode := readyNode.DeepCopy()
		labeledNode.Labels = map[string]string{"env": "prod"}
		client := fake.NewClientBuilder().WithScheme(scheme).WithLists(&v1.NodeList{Items: []v1.Node{*labeledNode}}).Build()
		svcWithAnno := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc",
				Namespace: "default",
				Annotations: map[string]string{
					"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-label": "env=prod",
				},
			},
		}
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: svcWithAnno, Anno: annotation.NewAnnotationRequest(svcWithAnno)}
		nodes, err := GetNodes(reqCtx, client)
		assert.NoError(t, err)
		assert.Len(t, nodes, 1)
	})
	t.Run("empty list", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: service, Anno: annotation.NewAnnotationRequest(service)}
		nodes, err := GetNodes(reqCtx, client)
		assert.NoError(t, err)
		assert.Len(t, nodes, 0)
	})
}

func TestGetEndpoints(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
	}
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Subsets:   []v1.EndpointSubset{{Addresses: []v1.EndpointAddress{{IP: "10.0.0.1"}}}},
	}
	t.Run("found", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(eps).Build()
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: service, Anno: annotation.NewAnnotationRequest(service)}
		result, err := getEndpoints(reqCtx, client)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Subsets, 1)
	})
	t.Run("not found", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: service, Anno: annotation.NewAnnotationRequest(service)}
		result, err := getEndpoints(reqCtx, client)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Subsets, 0)
	})
}

func TestGetEndpointByEndpointSlice(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)
	_ = discovery.AddToScheme(scheme)
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
	}
	esIPv4 := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-1",
			Namespace: "default",
			Labels:    map[string]string{discovery.LabelServiceName: "svc"},
		},
		AddressType: discovery.AddressTypeIPv4,
		Endpoints:   []discovery.Endpoint{{Addresses: []string{"10.0.0.1"}}},
	}
	esIPv6 := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-2",
			Namespace: "default",
			Labels:    map[string]string{discovery.LabelServiceName: "svc"},
		},
		AddressType: discovery.AddressTypeIPv6,
		Endpoints:   []discovery.Endpoint{{Addresses: []string{"fd00::1"}}},
	}
	t.Run("ipv4", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(esIPv4).Build()
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: service, Anno: annotation.NewAnnotationRequest(service)}
		result, err := getEndpointByEndpointSlice(reqCtx, client, model.IPv4)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, discovery.AddressTypeIPv4, result[0].AddressType)
	})
	t.Run("ipv6", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(esIPv6).Build()
		reqCtx := &svcCtx.RequestContext{Ctx: context.Background(), Service: service, Anno: annotation.NewAnnotationRequest(service)}
		result, err := getEndpointByEndpointSlice(reqCtx, client, model.IPv6)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, discovery.AddressTypeIPv6, result[0].AddressType)
	})
}
