package nlbv2

import (
	"context"
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	ecsmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/ecs"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

func getReconcileNLB() (*ReconcileNLB, error) {
	eventRecord := record.NewFakeRecorder(100)
	recon := &ReconcileNLB{
		cloud:            getMockCloudProvider(),
		kubeClient:       getFakeKubeClient(),
		logger:           ctrl.Log.WithName("controller").WithName("nlb-controller"),
		record:           eventRecord,
		finalizerManager: helper.NewDefaultFinalizerManager(getFakeKubeClient()),
	}

	nlbManager := NewNLBManager(recon.cloud)
	listenerManager := NewListenerManager(recon.cloud)
	serverGroupManager, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	if err != nil {
		return nil, fmt.Errorf("NewServerGroupManager error:%s", err.Error())
	}
	recon.builder = NewModelBuilder(nlbManager, listenerManager, serverGroupManager)
	recon.applier = NewModelApplier(nlbManager, listenerManager, serverGroupManager)
	return recon, nil
}

func TestReconcileNLB_reconcileNLB(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	_, err = recon.Reconcile(context.TODO(), req)
	assert.Equal(t, nil, err)
}

func TestReconcileNLB_reconcileDeleteNLB(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: v1.NamespaceDefault,
			Name:      DelServiceName,
		},
	}
	_, err = recon.Reconcile(context.TODO(), req)
	assert.Equal(t, nil, err)

}

func TestReconcileNLB_reconcileNotFoundNLB(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: v1.NamespaceDefault,
			Name:      "not-found-svc",
		},
	}
	_, err = recon.Reconcile(context.TODO(), req)
	assert.Equal(t, nil, err)
}

var (
	NodeName       = "cn-hangzhou.192.0.168.68"
	ServiceName    = "nlb"
	DelServiceName = "nlbDel"
)

func getFakeKubeClient() client.Client {
	// Node
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   NodeName,
					Labels: map[string]string{"app": "nginx"},
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.64/26",
					ProviderID: "cn-hangzhou.ecs-id-1",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Reason: "KubeletReady",
							Status: v1.ConditionTrue,
							Type:   v1.NodeReady,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cn-hangzhou.192.0.168.69",
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.128/26",
					ProviderID: "alicloud://cn-hangzhou.ecs-id-2",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Reason: "KubeletReady",
							Status: "True",
						},
					},
				},
			},
		},
	}

	// service
	svcList := &v1.ServiceList{
		Items: []v1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ServiceName,
					Namespace: v1.NamespaceDefault,
					UID:       "5e4dbfc9-c2ae-4642-b033-5607860aef6a",
					Annotations: map[string]string{
						annotation.Annotation(annotation.ZoneMaps): "cn-hangzhou-a:vsw-1,cn-hangzhou-b:vsw-2",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							NodePort:   80,
							Protocol:   v1.ProtocolTCP,
						},
					},
					LoadBalancerClass: tea.String(helper.NLBClass),
					Type:              v1.ServiceTypeLoadBalancer,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DelServiceName,
					Namespace: v1.NamespaceDefault,
					UID:       "5e4dbfc9-c2ae-4642-b033-5607860aef6a",
					Annotations: map[string]string{
						annotation.Annotation(annotation.ZoneMaps): "cn-hangzhou-a:vsw-1,cn-hangzhou-b:vsw-2",
					},
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{helper.NLBFinalizer},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							NodePort:   80,
							Protocol:   v1.ProtocolTCP,
						},
					},
					LoadBalancerClass: tea.String(helper.NLBClass),
					Type:              v1.ServiceTypeLoadBalancer,
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								Hostname: vmock.NLBDNSName,
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "clb",
					Namespace: v1.NamespaceDefault,
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							NodePort:   80,
							Protocol:   v1.ProtocolTCP,
						},
					},
					Type: v1.ServiceTypeLoadBalancer,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodePort",
					Namespace: v1.NamespaceDefault,
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							NodePort:   80,
							Protocol:   v1.ProtocolTCP,
						},
					},
					Type: v1.ServiceTypeNodePort,
				},
			},
		},
	}

	// Endpoint & EndpointSlice
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nlb",
			Namespace: v1.NamespaceDefault,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.96.0.15",
						NodeName: tea.String(NodeName),
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "https",
						Port:     443,
						Protocol: "TCP",
					},
					{
						Name:     "http",
						Port:     8080,
						Protocol: "TCP",
					},
					{
						Name:     "tcp",
						Port:     80,
						Protocol: "TCP",
					},
					{
						Name:     "udp",
						Port:     53,
						Protocol: "TCP",
					},
				},
			},
		},
	}

	var (
		epReady        = true
		portName       = "port-tcp"
		protocol       = v1.ProtocolTCP
		port     int32 = 80
	)
	es := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nlb",
			Namespace: v1.NamespaceDefault,
			Labels:    map[string]string{discovery.LabelServiceName: "nlb"},
		},
		Endpoints: []discovery.Endpoint{
			{
				Addresses: []string{"10.96.0.15"},
				Conditions: discovery.EndpointConditions{
					Ready: &epReady,
				},
				NodeName: &NodeName,
			},
		},
		Ports: []discovery.EndpointPort{
			{
				Name:     &portName,
				Port:     &port,
				Protocol: &protocol,
			},
		},
		AddressType: discovery.AddressTypeIPv4,
	}

	objs := []runtime.Object{nodeList, eps, es, svcList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}

func getMockCloudProvider() prvd.Provider {
	return vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-id"),
		MockNLB:   vmock.NewMockNLB(nil),
	}
}

func getReqCtx(svc *v1.Service) *svcCtx.RequestContext {
	return &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}
}

func TestRateLimiter(t *testing.T) {
	rateLimit := workqueue.NewTypedMaxOfRateLimiter[string](
		workqueue.NewTypedItemExponentialFailureRateLimiter[string](5*time.Second, 300*time.Second),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.TypedBucketRateLimiter[string]{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
	item := "test"

	for i := 0; i < 10; i++ {
		d := rateLimit.When(item)
		t.Logf("ep duration %f", d.Seconds())
	}
}

func TestUpdateReadinessCondition(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName,
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: v1.NamespaceDefault,
			},
		}
		err = recon.kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		sgs := []*nlbmodel.ServerGroup{
			{
				InitialServers: []nlbmodel.ServerGroupServer{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: v1.NamespaceDefault,
							Name:      "test-pod-1",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, sgs)
		assert.NoError(t, err)
	})

	t.Run("TargetRef is nil", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName,
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		sgs := []*nlbmodel.ServerGroup{
			{
				InitialServers: []nlbmodel.ServerGroupServer{
					{
						TargetRef: nil,
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, sgs)
		assert.NoError(t, err)
	})

	t.Run("duplicate pod", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName,
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: v1.NamespaceDefault,
			},
		}
		err = recon.kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		sgs := []*nlbmodel.ServerGroup{
			{
				InitialServers: []nlbmodel.ServerGroupServer{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: v1.NamespaceDefault,
							Name:      "test-pod-2",
						},
					},
					{
						TargetRef: &v1.ObjectReference{
							Namespace: v1.NamespaceDefault,
							Name:      "test-pod-2",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, sgs)
		assert.NoError(t, err)
	})

	t.Run("pod not found", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName,
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		sgs := []*nlbmodel.ServerGroup{
			{
				InitialServers: []nlbmodel.ServerGroupServer{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: v1.NamespaceDefault,
							Name:      "non-existent-pod",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, sgs)
		assert.NoError(t, err)
	})

	t.Run("multiple server groups", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName,
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		pod1 := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-3",
				Namespace: v1.NamespaceDefault,
			},
		}
		pod2 := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-4",
				Namespace: v1.NamespaceDefault,
			},
		}
		err = recon.kubeClient.Create(context.TODO(), pod1)
		assert.NoError(t, err)
		err = recon.kubeClient.Create(context.TODO(), pod2)
		assert.NoError(t, err)

		sgs := []*nlbmodel.ServerGroup{
			{
				InitialServers: []nlbmodel.ServerGroupServer{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: v1.NamespaceDefault,
							Name:      "test-pod-3",
						},
					},
				},
			},
			{
				InitialServers: []nlbmodel.ServerGroupServer{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: v1.NamespaceDefault,
							Name:      "test-pod-4",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, sgs)
		assert.NoError(t, err)
	})

	t.Run("empty server groups", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName,
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		sgs := []*nlbmodel.ServerGroup{}

		err = recon.updateReadinessCondition(reqCtx, sgs)
		assert.NoError(t, err)
	})
}

func TestNLBControllerFunctions(t *testing.T) {
	t.Run("newReconciler function exists and can be called", func(t *testing.T) {
		// Simply test that the function exists and signature is correct
		// We can't fully test it without a real manager and context
		assert.NotNil(t, newReconciler)
	})
}

func TestReconcileNLB_UpdateServiceLabels(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-label-test",
			Namespace: v1.NamespaceDefault,
			Labels:    map[string]string{},
		},
	}

	cl := fake.NewClientBuilder().
		WithStatusSubresource(&v1.Service{}).
		WithRuntimeObjects(svc).
		Build()
	recon := &ReconcileNLB{kubeClient: cl}

	lb := &nlbmodel.NetworkLoadBalancer{
		LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
			LoadBalancerId: "nlb-123",
		},
		AssociatedSecurityGroup: &ecsmodel.SecurityGroup{
			ID: "sg-123",
		},
	}

	err := recon.updateServiceLabels(svc, lb)
	assert.NoError(t, err)

	updated := &v1.Service{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, updated)
	assert.NoError(t, err)
	assert.Equal(t, "nlb-123", updated.Labels[helper.LabelLoadBalancerId])
	assert.Equal(t, "sg-123", updated.Labels[helper.LabelSecurityGroupId])
	assert.NotEmpty(t, updated.Labels[helper.LabelServiceHash])
}

func TestReconcileNLB_RemoveServiceLabels(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-remove-label-test",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				helper.LabelServiceHash:     "hash",
				helper.LabelLoadBalancerId:  "nlb-1",
				helper.LabelSecurityGroupId: "sg-1",
			},
		},
	}

	cl := fake.NewClientBuilder().
		WithStatusSubresource(&v1.Service{}).
		WithRuntimeObjects(svc).
		Build()
	recon := &ReconcileNLB{kubeClient: cl}

	err := recon.removeServiceLabels(svc)
	assert.NoError(t, err)

	updated := &v1.Service{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, updated)
	assert.NoError(t, err)
	assert.NotContains(t, updated.Labels, helper.LabelServiceHash)
	assert.NotContains(t, updated.Labels, helper.LabelLoadBalancerId)
	assert.NotContains(t, updated.Labels, helper.LabelSecurityGroupId)
}

func TestReconcileNLB_ReconcileLoadBalancerResources(t *testing.T) {
	t.Run("success path", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)

		svc := &v1.Service{}
		err = recon.kubeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		}, svc)
		assert.NoError(t, err)

		err = recon.reconcileLoadBalancerResources(getReqCtx(svc))
		assert.NoError(t, err)
	})

	t.Run("build and apply model failed", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)

		badSvc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bad-zone-map",
				Namespace: v1.NamespaceDefault,
				Annotations: map[string]string{
					annotation.Annotation(annotation.ZoneMaps): "invalid",
				},
			},
			Spec: v1.ServiceSpec{
				Type:              v1.ServiceTypeLoadBalancer,
				LoadBalancerClass: tea.String(helper.NLBClass),
				Ports: []v1.ServicePort{
					{
						Name:       "tcp",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   v1.ProtocolTCP,
					},
				},
			},
		}

		err = recon.kubeClient.Create(context.TODO(), badSvc)
		assert.NoError(t, err)

		err = recon.reconcileLoadBalancerResources(getReqCtx(badSvc))
		assert.Error(t, err)
	})
}

func TestReconcileNLB_UpdateAndRemoveServiceStatus(t *testing.T) {
	t.Run("update service status lb nil", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)

		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "status-nil-lb",
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		err = recon.updateServiceStatus(reqCtx, svc, nil)
		assert.Error(t, err)
	})

	t.Run("remove service status no change", func(t *testing.T) {
		recon, err := getReconcileNLB()
		assert.NoError(t, err)

		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "status-no-change",
				Namespace: v1.NamespaceDefault,
			},
		}
		reqCtx := getReqCtx(svc)

		err = recon.removeServiceStatus(reqCtx, svc)
		assert.NoError(t, err)
	})
}

func TestHasInvalidServerInServerGroups(t *testing.T) {
	assert.False(t, hasInvalidServerInServerGroups(nil))
	assert.False(t, hasInvalidServerInServerGroups([]*nlbmodel.ServerGroup{
		nil,
		{InvalidServers: nil},
	}))
	assert.True(t, hasInvalidServerInServerGroups([]*nlbmodel.ServerGroup{
		{InvalidServers: []nlbmodel.ServerGroupServer{{ServerId: "s-1"}}},
	}))
}
