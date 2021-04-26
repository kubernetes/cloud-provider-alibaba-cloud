package service

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
)

func NewCloudModelBuilder(cloud prvd.Provider, svc *v1.Service, anno *AnnotationRequest) *cloudModelBuilder {
	return &cloudModelBuilder{
		cloud: cloud,
		svc:   svc,
		anno:  anno,
	}
}

type cloudModelBuilder struct {
	svc   *v1.Service
	anno  *AnnotationRequest
	cloud prvd.Provider
}

func (c cloudModelBuilder) Build() (*model.LoadBalancer, error) {
	lb := &model.LoadBalancer{}
	buildLoadBalancerAttribute(c.svc, lb)
	//buildBackend(svc, lb)
	//buildListener(svc, lb)
	return lb, nil
}

func buildLoadBalancerAttribute(svc *v1.Service, lb *model.LoadBalancer) error {
	return nil

}

//
//func (c cloudModelBuilder) findLoadBalancer() (bool, *slb.LoadBalancerType, error) {
//	return false, nil, nil
//}
