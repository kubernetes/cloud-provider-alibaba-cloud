package clbv1

import (
	"context"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
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
				Topology: map[string]string{
					v1.LabelHostname: NodeName,
				},
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

func getTestVGroupManager() *VGroupManager {
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
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.ServiceLog.WithValues("service", util.Key(svc)),
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
	vGroupManager := NewVGroupManager(recon.kubeClient, recon.cloud)
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
