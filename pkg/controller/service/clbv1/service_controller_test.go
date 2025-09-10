package clbv1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	NS          = "default"
	NodeName    = "cn-hangzhou.192.0.168.68"
	PodIP       = "10.96.0.15"
	SvcUID      = "5e4dbfc9-c2ae-4642-b033-5607860aef6a"
	SvcName     = "test"
	hostSvcName = "hostname-service"
	eipSvcName  = "eip-service"
	delSvcName  = "delete-service"
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
					ProviderID: "cn-hangzhou.ecs-id",
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
					ProviderID: "alicloud://cn-hangzhou.ecs-id",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
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
					Name:      SvcName,
					Namespace: NS,
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
					Name:        hostSvcName,
					Namespace:   NS,
					Annotations: map[string]string{annotation.Annotation(annotation.HostName): "foo.bar.com"},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        eipSvcName,
					Namespace:   NS,
					Annotations: map[string]string{annotation.Annotation(annotation.ExternalIPType): "eip"},
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              delSvcName,
					Namespace:         NS,
					Annotations:       make(map[string]string),
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{"service.k8s.alibaba/resources"},
					UID:               types.UID(SvcUID),
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{
								IP: vmock.LoadBalancerIP,
							},
						},
					},
				},
			},
		},
	}

	// Endpoint & EndpointSlice
	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SvcName,
			Namespace: NS,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       PodIP,
						NodeName: &NodeName,
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
			Name:      SvcName,
			Namespace: NS,
			Labels:    map[string]string{discovery.LabelServiceName: SvcName},
		},
		Endpoints: []discovery.Endpoint{
			{
				Addresses: []string{PodIP},
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
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
}

func getTestVGroupManager() (*VGroupManager, error) {
	return NewVGroupManager(getFakeKubeClient(), getMockCloudProvider())
}

func getDefaultService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        SvcName,
			Namespace:   NS,
			Annotations: make(map[string]string),
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
	}
}

func getReqCtx(svc *v1.Service) *svcCtx.RequestContext {
	return &svcCtx.RequestContext{
		Ctx:      context.TODO(),
		Service:  svc,
		Anno:     annotation.NewAnnotationRequest(svc),
		Log:      util.ServiceLog.WithValues("service", util.Key(svc)),
		Recorder: record.NewFakeRecorder(100),
	}
}

func getReconcileService() *ReconcileService {
	eventRecord := record.NewFakeRecorder(100)
	recon := &ReconcileService{
		cloud:            getMockCloudProvider(),
		kubeClient:       getFakeKubeClient(),
		record:           eventRecord,
		finalizerManager: helper.NewDefaultFinalizerManager(getFakeKubeClient()),
	}

	slbManager := NewLoadBalancerManager(recon.cloud)
	listenerManager := NewListenerManager(recon.cloud)
	vGroupManager, _ := NewVGroupManager(recon.kubeClient, recon.cloud)
	recon.builder = NewModelBuilder(slbManager, listenerManager, vGroupManager)
	recon.applier = NewModelApplier(slbManager, listenerManager, vGroupManager)

	return recon
}

func TestReconcileService(t *testing.T) {
	recon := getReconcileService()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      SvcName,
			Namespace: NS,
		},
	}
	_, err := recon.Reconcile(context.TODO(), req)
	if err != nil {
		t.Error(err)
	}
}

func TestReconcileHostNameService(t *testing.T) {
	recon := getReconcileService()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      hostSvcName,
			Namespace: NS,
		},
	}
	_, err := recon.Reconcile(context.TODO(), req)
	if err != nil {
		t.Error(err)
	}
}

func TestReconcileEIPService(t *testing.T) {
	recon := getReconcileService()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      eipSvcName,
			Namespace: NS,
		},
	}
	_, err := recon.Reconcile(context.TODO(), req)
	if err != nil {
		t.Error(err)
	}
}

func TestReconcileDeleteService(t *testing.T) {
	recon := getReconcileService()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      delSvcName,
			Namespace: NS,
		},
	}
	_, err := recon.Reconcile(context.TODO(), req)
	if err != nil {
		t.Error(err)
	}
}

func TestReconcileNotFoundService(t *testing.T) {
	recon := getReconcileService()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "not-found-service",
			Namespace: NS,
		},
	}
	_, err := recon.Reconcile(context.TODO(), req)
	if err != nil {
		t.Error(err)
	}
}

func TestDryRun(t *testing.T) {
	ctrlCfg.ControllerCFG.DryRun = true
	initMap(getFakeKubeClient())

	recon := getReconcileService()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      SvcName,
			Namespace: NS,
		},
	}
	_, err := recon.Reconcile(context.TODO(), req)
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateReadinessCondition(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: NS,
			},
		}
		err := recon.kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		vgroups := []model.VServerGroup{
			{
				InitialBackends: []model.BackendAttribute{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: NS,
							Name:      "test-pod-1",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("TargetRef is nil", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		vgroups := []model.VServerGroup{
			{
				InitialBackends: []model.BackendAttribute{
					{
						TargetRef: nil,
					},
				},
			},
		}

		err := recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("duplicate pod", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: NS,
			},
		}
		err := recon.kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		vgroups := []model.VServerGroup{
			{
				InitialBackends: []model.BackendAttribute{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: NS,
							Name:      "test-pod-2",
						},
					},
					{
						TargetRef: &v1.ObjectReference{
							Namespace: NS,
							Name:      "test-pod-2",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("pod not found", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		vgroups := []model.VServerGroup{
			{
				InitialBackends: []model.BackendAttribute{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: NS,
							Name:      "non-existent-pod",
						},
					},
				},
			},
		}

		err := recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("multiple vgroups", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		pod1 := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-3",
				Namespace: NS,
			},
		}
		pod2 := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-4",
				Namespace: NS,
			},
		}
		err := recon.kubeClient.Create(context.TODO(), pod1)
		assert.NoError(t, err)
		err = recon.kubeClient.Create(context.TODO(), pod2)
		assert.NoError(t, err)

		vgroups := []model.VServerGroup{
			{
				InitialBackends: []model.BackendAttribute{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: NS,
							Name:      "test-pod-3",
						},
					},
				},
			},
			{
				InitialBackends: []model.BackendAttribute{
					{
						TargetRef: &v1.ObjectReference{
							Namespace: NS,
							Name:      "test-pod-4",
						},
					},
				},
			},
		}

		err = recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("empty vgroups", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		vgroups := []model.VServerGroup{}

		err := recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("InvalidBackends ConditionFalse path", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "invalid-backend-pod", Namespace: NS},
		}
		err := recon.kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		vgroups := []model.VServerGroup{
			{
				InvalidBackends: []model.BackendAttribute{
					{
						TargetRef: &v1.ObjectReference{Namespace: NS, Name: "invalid-backend-pod"},
					},
				},
			},
		}
		err = recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})

	t.Run("ENI Backends path", func(t *testing.T) {
		recon := getReconcileService()
		svc := getDefaultService()
		svc.Annotations = map[string]string{helper.BackendType: model.ENIBackendType}
		reqCtx := getReqCtx(svc)

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "eni-backend-pod", Namespace: NS},
		}
		err := recon.kubeClient.Create(context.TODO(), pod)
		assert.NoError(t, err)

		vgroups := []model.VServerGroup{
			{
				Backends: []model.BackendAttribute{
					{TargetRef: &v1.ObjectReference{Namespace: NS, Name: "eni-backend-pod"}},
				},
			},
		}
		err = recon.updateReadinessCondition(reqCtx, vgroups)
		assert.NoError(t, err)
	})
}

func TestRemoveServiceStatus(t *testing.T) {
	t.Run("no change when status already empty", func(t *testing.T) {
		recon := getReconcileService()
		svc := &v1.Service{}
		err := recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: NS, Name: SvcName}, svc)
		assert.NoError(t, err)
		reqCtx := getReqCtx(svc)
		err = recon.removeServiceStatus(reqCtx, svc)
		assert.NoError(t, err)
	})

	t.Run("patch to clear status", func(t *testing.T) {
		recon := getReconcileService()
		svc := &v1.Service{}
		err := recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: NS, Name: SvcName}, svc)
		assert.NoError(t, err)
		svc.Status.LoadBalancer = v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{{IP: "1.2.3.4"}},
		}
		err = recon.kubeClient.Status().Update(context.TODO(), svc)
		assert.NoError(t, err)

		reqCtx := getReqCtx(svc)
		err = recon.removeServiceStatus(reqCtx, svc)
		assert.NoError(t, err)

		updated := &v1.Service{}
		err = recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: NS, Name: SvcName}, updated)
		assert.NoError(t, err)
		assert.Empty(t, updated.Status.LoadBalancer.Ingress)
	})
}

func TestRemoveServiceLabels(t *testing.T) {
	t.Run("no op when no target labels", func(t *testing.T) {
		recon := getReconcileService()
		svc := &v1.Service{}
		err := recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: NS, Name: SvcName}, svc)
		assert.NoError(t, err)
		err = recon.removeServiceLabels(svc)
		assert.NoError(t, err)
	})

	t.Run("remove labels when present", func(t *testing.T) {
		recon := getReconcileService()
		svc := &v1.Service{}
		err := recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: NS, Name: SvcName}, svc)
		assert.NoError(t, err)
		if svc.Labels == nil {
			svc.Labels = make(map[string]string)
		}
		svc.Labels[helper.LabelServiceHash] = "hash1"
		svc.Labels[helper.LabelLoadBalancerId] = "lb-1"
		err = recon.kubeClient.Update(context.TODO(), svc)
		assert.NoError(t, err)

		svc2 := &v1.Service{}
		err = recon.kubeClient.Get(context.TODO(), types.NamespacedName{Namespace: NS, Name: SvcName}, svc2)
		assert.NoError(t, err)
		err = recon.removeServiceLabels(svc2)
		assert.NoError(t, err)
	})
}
