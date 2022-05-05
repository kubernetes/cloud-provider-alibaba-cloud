package service

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func getTestVGroupManager() *VGroupManager {
	nodeName1 := "cn-hangzhou.192.0.168.68"

	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeName1,
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.64/26",
					ProviderID: "cn-hangzhou.i-m44444444",
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

	eps := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.96.0.15",
						NodeName: &nodeName1,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "http",
						Port:     443,
						Protocol: "TCP",
					},
				},
			},
		},
	}
	objs := []runtime.Object{nodeList, eps}
	cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	return NewVGroupManager(cl, getMockCloudProvider())
}

func getReqCtxForVGroup() *RequestContext {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: v1.ServiceSpec{
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
	svc.Annotations[Annotation(LoadBalancerName)] = "test-slb-name"

	return &RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    NewAnnotationRequest(svc),
		Log:     util.ServiceLog.WithValues("service", util.Key(svc)),
	}
}

func TestBuildLocalModelForVGroup(t *testing.T) {
	m := getTestVGroupManager()
	reqCtx := getReqCtxForVGroup()
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
	}
	// cluster
	reqCtx.Service.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyType(v1.ServiceInternalTrafficPolicyCluster)
	if err := m.BuildLocalModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}
	// local
	reqCtx.Service.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyType(v1.ServiceInternalTrafficPolicyLocal)
	if err := m.BuildLocalModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}
	// eni
	reqCtx.Service.Annotations[BackendType] = model.ENIBackendType
	if err := m.BuildLocalModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}
}

func TestBuildRemoteModelForVGroup(t *testing.T) {
	m := getTestVGroupManager()
	reqCtx := getReqCtxForVGroup()
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: model.LoadBalancerAttribute{
			LoadBalancerId: "lb-id",
		},
	}
	if err := m.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}
}
