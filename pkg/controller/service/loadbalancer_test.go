package service

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"testing"
)

func getMockCloudProvider() prvd.Provider {
	return vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
}

func getReqCtxForLB() *RequestContext {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
	}
	svc.Annotations[LoadBalancerName] = "test-slb-name"

	return &RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    NewAnnotationRequest(svc),
		Log:     util.ServiceLog.WithValues("service", util.Key(svc)),
	}

}

func TestBuildLocalModelForLB(t *testing.T) {
	m := NewLoadBalancerManager(getMockCloudProvider())
	reqCtx := getReqCtxForLB()
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(reqCtx.Service),
	}
	if err := m.BuildLocalModel(reqCtx, lbMdl); err != nil {
		t.Error(err)
	}
}

func TestBuildRemoteModelForLB(t *testing.T) {
	m := NewLoadBalancerManager(getMockCloudProvider())
	reqCtx := getReqCtxForLB()
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
