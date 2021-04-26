package service

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// NewDefaultModelBuilder construct a new defaultModelBuilder
func NewClusterModelBuilder(kubeClient client.Client, cloud prvd.Provider, svc *v1.Service, anno *AnnotationRequest) *clusterModelBuilder {
	return &clusterModelBuilder{
		kubeClient: kubeClient,
		cloud:      cloud,
		svc:        svc,
		anno:       anno,
	}
}

type clusterModelBuilder struct {
	svc  *v1.Service
	anno *AnnotationRequest

	cloud      prvd.Provider
	kubeClient client.Client
}

func (c clusterModelBuilder) Build() (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{}
	if err := c.buildLoadBalancerAttribute(lbMdl); err != nil {
		return nil, fmt.Errorf("build cluster model error: %s", err.Error())
	}
	//buildBackend(lb)
	//buildListener(lb)
	return lbMdl, nil
}

func (c clusterModelBuilder) buildLoadBalancerAttribute(m *model.LoadBalancer) error {

	m.LoadBalancerAttribute.AddressType = c.anno.Get(AddressType)
	m.LoadBalancerAttribute.InternetChargeType = c.anno.Get(ChargeType)
	bandwidth := c.anno.Get(Bandwidth)
	if bandwidth != nil {
		i, err := strconv.Atoi(*bandwidth)
		if err != nil &&
			*(m.LoadBalancerAttribute.InternetChargeType) == string(model.PayByBandwidth) {
			return fmt.Errorf("bandwidth must be integer, got [%s], error: %s", *bandwidth, err.Error())
		}
		m.LoadBalancerAttribute.Bandwidth = &i
	}
	m.LoadBalancerAttribute.RefLoadBalancerId = c.anno.Get(LoadBalancerId)
	m.LoadBalancerAttribute.LoadBalancerName = c.anno.Get(LoadBalancerName)
	m.LoadBalancerAttribute.VSwitchId = c.anno.Get(VswitchId)
	m.LoadBalancerAttribute.MasterZoneId = c.anno.Get(MasterZoneID)
	m.LoadBalancerAttribute.SlaveZoneId = c.anno.Get(SlaveZoneID)
	m.LoadBalancerAttribute.LoadBalancerSpec = c.anno.Get(Spec)
	m.LoadBalancerAttribute.ResourceGroupId = c.anno.Get(ResourceGroupId)
	m.LoadBalancerAttribute.AddressIPVersion = c.anno.Get(IPVersion)
	m.LoadBalancerAttribute.DeleteProtection = c.anno.Get(DeleteProtection)
	m.LoadBalancerAttribute.ModificationProtectionStatus = c.anno.Get(ModificationProtection)

	return nil
}

func buildBackend(svc *v1.Service, model *model.LoadBalancer) error {
	return nil
}

func buildListener(svc *v1.Service, model *model.LoadBalancer) error {
	return nil
}

/*
type EndpointMgr struct {
	isEniBackendType bool

	req       *AnnotationRequest
	localMode bool
	nodes     []v1.Node
	svc       *v1.Service
	endpoints *v1.Endpoints
}

func NewEndpointMgr(
	mc client.Client, svc *v1.Service,
) (*EndpointMgr, error) {
	mgr := &EndpointMgr{
		svc:       svc,
		req:       &AnnotationRequest{svc: svc},
		localMode: isLocalModeService(svc),
		// backend type
		isEniBackendType: IsENIBackendType(svc),
	}
	nodes := v1.NodeList{}
	err := mc.List(context.TODO(), &nodes)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("svc: %s", key(svc)))
	}
	ns, err := filterOutByLabel(nodes.Items, mgr.req.Get(BackendLabel))
	if err != nil {
		return nil, errors.Wrap(err, BackendLabel)
	}
	mgr.nodes = ns

	eps := v1.Endpoints{}
	err = mc.Get(context.TODO(), client.ObjectKey{Namespace: svc.Namespace, Name: svc.Name}, &eps)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return nil, errors.Wrap(err, fmt.Sprintf("svc: %s", key(svc)))
		}
		klog.Warningf("endpoint not found: %s", key(svc))
	}
	mgr.endpoints = &eps
	return mgr, nil
}
*/
