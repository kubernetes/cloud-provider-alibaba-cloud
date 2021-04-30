package service

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	v1 "k8s.io/api/core/v1"
	ctx2 "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

// NewDefaultModelBuilder construct a new defaultModelBuilder
func NewClusterModelBuilder(ctx context.Context, kubeClient client.Client, cloud prvd.Provider, svc *v1.Service, anno *AnnotationRequest) *clusterModelBuilder {
	return &clusterModelBuilder{
		ctx:        ctx,
		kubeClient: kubeClient,
		cloud:      cloud,
		svc:        svc,
		anno:       anno,
	}
}

type clusterModelBuilder struct {
	ctx  context.Context
	svc  *v1.Service
	anno *AnnotationRequest

	cloud      prvd.Provider
	kubeClient client.Client
}

func (c clusterModelBuilder) Build() (*model.LoadBalancer, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(c.svc),
	}

	if err := c.buildLoadBalancerAttribute(lbMdl); err != nil {
		return nil, fmt.Errorf("build slb attribute error: %s", err.Error())
	}
	//if err := c.buildBackend(lbMdl); err != nil {
	//	return nil, fmt.Errorf("build slb backend error: %s", err.Error())
	//}
	//if err := c.buildListener(lbMdl); err != nil {
	//	return nil, fmt.Errorf("build slb listener error: %s", err.Error())
	//}
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
	if c.anno.Get(LoadBalancerId) != nil {
		m.LoadBalancerAttribute.LoadBalancerId = *c.anno.Get(LoadBalancerId)
		m.LoadBalancerAttribute.IsUserManaged = true
	}
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

func (c clusterModelBuilder) buildBackend(mdl *model.LoadBalancer) error {
	ctx := context.TODO()

	nodes := v1.NodeList{}
	err := c.kubeClient.List(ctx, &nodes)
	if err != nil {
		return fmt.Errorf("get nodes from k8s error: %s", err.Error())
	}
	nds, err := filterNodes(nodes.Items, *c.anno.Get(BackendLabel))
	if err != nil {
		return fmt.Errorf("filter nodes error: %s", err.Error())
	}

	eps := &v1.Endpoints{}
	err = c.kubeClient.Get(ctx, util.NamespacedName(c.svc), eps)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("get endpoints %s from k8s error: %s", util.Key(c.svc), err.Error())
		}
		klog.Warningf("endpoint not found: %s", util.Key(c.svc))
	}

	backends, err := getBackends(c, c.svc, eps, nds)
	if err != nil {
		return fmt.Errorf("build backends error: %s", err.Error())
	}
	mdl.Backends = backends
	return nil
}

func getBackends(c clusterModelBuilder, svc *v1.Service, eps *v1.Endpoints, nodes []v1.Node) ([]model.BackendAttribute, error) {
	var backends []model.BackendAttribute
	// ENI mode
	if isENIBackendType(svc) {
		if len(eps.Subsets) == 0 {
			klog.Warningf("%s endpoint is nil in eni mode", util.Key(svc))
			return nil, nil
		}
		klog.Infof("[ENI] mode service: %s", util.Key(svc))
		var privateIpAddress []string
		for _, ep := range eps.Subsets {
			for _, addr := range ep.Addresses {
				privateIpAddress = append(privateIpAddress, addr.IP)
			}
		}
		err := Batch(privateIpAddress, 40, c.buildFunc(&backends))
		if err != nil {
			return backends, fmt.Errorf("batch process eni fail: %s", err.Error())
		}
	}
	// Local mode
	if isLocalModeService(svc) {
		if len(eps.Subsets) == 0 {
			klog.Warningf("%s endpoint is nil in local mode", util.Key(svc))
			return nil, nil
		}
		klog.Infof("[Local] mode service: %s", util.Key(svc))
		// 1. add duplicate ecs backends
		for _, sub := range eps.Subsets {
			for _, add := range sub.Addresses {
				if add.NodeName == nil {
					return nil, fmt.Errorf("add ecs backends for service[%s] error, NodeName is nil for ip %s ", util.Key(svc), add.IP)
				}
				node := findNodeByNodeName(nodes, *add.NodeName)
				if node == nil {
					klog.Warningf("can not find correspond node %s for endpoint %s", *add.NodeName, add.IP)
					continue
				}
				if isExcludeNode(node) {
					// filter vk node
					continue
				}
				_, id, err := nodeFromProviderID(node.Spec.ProviderID)
				if err != nil {
					return backends, fmt.Errorf("parse providerid: %s. "+
						"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
				}
				backends = append(
					backends,
					model.BackendAttribute{
						ServerId: id,
						Weight:   DEFAULT_SERVER_WEIGHT,
						Type:     "ecs",
					},
				)
			}
		}
		// 2. add eci backends
		return addECIBackends(backends)
	}

	// Cluster mode
	// When ecs and eci are deployed in a cluster, add ecs first and then add eci
	klog.Infof("[Cluster] mode service: %s", util.Key(svc))
	// 1. add ecs backends
	for _, node := range nodes {
		if isExcludeNode(&node) {
			continue
		}
		_, id, err := nodeFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return backends, fmt.Errorf("normal parse providerid: %s. "+
				"expected: ${regionid}.${nodeid}, %s", node.Spec.ProviderID, err.Error())
		}

		backends = append(
			backends,
			model.BackendAttribute{
				ServerId: id,
				Weight:   DEFAULT_SERVER_WEIGHT,
				Type:     "ecs",
			},
		)
	}
	// 2. add eci backends
	return addECIBackends(backends)

}

func addECIBackends(backends []model.BackendAttribute) ([]model.BackendAttribute, error) {
	// TODO
	klog.Infof("implement me!")
	return backends, nil

}

func (c clusterModelBuilder) buildFunc(backend *[]model.BackendAttribute) func(o []interface{}) error {

	// backend build function
	return func(o []interface{}) error {
		var ips []string
		for _, i := range o {
			ip, ok := i.(string)
			if !ok {
				return fmt.Errorf("not string: %v", i)
			}
			ips = append(ips, ip)
		}
		// TODO FIX ME
		resp, err := c.cloud.DescribeNetworkInterfaces(ctx2.CFG.Global.VpcID, &ips)
		if err != nil {
			return fmt.Errorf("call DescribeNetworkInterfaces: %s", err.Error())
		}
		for _, ip := range ips {
			eniid, err := findENIbyAddrIP(resp, ip)
			if err != nil {
				return err
			}
			*backend = append(
				*backend,
				model.BackendAttribute{
					ServerId: eniid,
					Weight:   DEFAULT_SERVER_WEIGHT,
					Type:     "eni",
					ServerIp: ip,
				},
			)
		}
		return nil
	}
}

func (c clusterModelBuilder) buildListener(mdl *model.LoadBalancer) error {
	for _, port := range c.svc.Spec.Ports {
		listener := model.ListenerAttribute{
			ListenerPort: int(port.Port),
		}
		mdl.Listeners = append(mdl.Listeners, listener)
	}
	return nil
}

// NewCloudModelBuilder construct a new defaultModelBuilder

func NewCloudModelBuilder(ctx context.Context, cloud prvd.Provider, svc *v1.Service, anno *AnnotationRequest) *cloudModelBuilder {
	return &cloudModelBuilder{
		ctx:   ctx,
		cloud: cloud,
		svc:   svc,
		anno:  anno,
	}
}

type cloudModelBuilder struct {
	ctx   context.Context
	svc   *v1.Service
	anno  *AnnotationRequest
	cloud prvd.Provider
}

func (c cloudModelBuilder) Build() (*model.LoadBalancer, error) {
	lb := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(c.svc),
	}

	_, lb, err := c.buildLoadBalancerAttribute(lb)
	if err != nil {
		return nil, fmt.Errorf("can not get load balancer attribute from cloud, error: %s", err.Error())
	}
	//buildBackend(svc, lb)
	//buildListener(svc, lb)
	return lb, nil
}

func (c cloudModelBuilder) buildLoadBalancerAttribute(lb *model.LoadBalancer) (bool, *model.LoadBalancer, error) {
	// Initialize the slb model to find the slb associated with the svc
	// 1. set loadbalancer id
	if c.anno.Get(LoadBalancerId) != nil {
		lb.LoadBalancerAttribute.LoadBalancerId = *c.anno.Get(LoadBalancerId)
	}
	// 2. set default loadbalancer name
	v := c.anno.GetDefaultLoadBalancerName()
	lb.LoadBalancerAttribute.LoadBalancerName = &v
	// 3. set default loadbalancer tag
	lb.LoadBalancerAttribute.Tags = []slb.Tag{
		{
			TagKey:   TAGKEY,
			TagValue: v,
		},
	}
	return c.cloud.FindSLB(c.ctx, lb)
}
