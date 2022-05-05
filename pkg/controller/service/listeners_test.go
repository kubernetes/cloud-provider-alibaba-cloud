package service

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"testing"
)

func getReqCtxForListener() *RequestContext {
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
				{
					Name:       "udp",
					Port:       53,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
	svc.Annotations[AclID] = "acl-id"
	svc.Annotations[AclStatus] = string(model.OnFlag)
	svc.Annotations[AclType] = "white"
	svc.Annotations[ProtocolPort] = "tcp:80,udp:53,http:8080,https:443"
	svc.Annotations[CertID] = "cert-id"

	return &RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    NewAnnotationRequest(svc),
		Log:     util.ServiceLog.WithValues("service", util.Key(svc)),
	}
}

func TestBuildLocalModelForListener(t *testing.T) {
	listenerManager := NewListenerManager(getMockCloudProvider())
	reqCtx := getReqCtxForListener()
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
	}
	if err := listenerManager.BuildLocalModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}

}

func TestBuildRemoteModelForListener(t *testing.T) {
	listenerManager := NewListenerManager(getMockCloudProvider())
	reqCtx := getReqCtxForListener()
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
		LoadBalancerAttribute: model.LoadBalancerAttribute{
			LoadBalancerId: "lb-id",
		},
	}
	if err := listenerManager.BuildRemoteModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}
}
