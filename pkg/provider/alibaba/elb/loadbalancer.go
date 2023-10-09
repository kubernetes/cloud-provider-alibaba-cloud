package elb

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ens"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"
	"k8s.io/klog/v2"
	"time"
)

const (
	connectionTimeout = 30 * time.Second
	readTimeout       = 30 * time.Second
)

func NewELBProvider(auth *base.ClientMgr) *ELBProvider {
	return &ELBProvider{auth: auth}
}

var _ prvd.IELB = &ELBProvider{}

type ELBProvider struct {
	auth *base.ClientMgr
}

func (e ELBProvider) FindEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {

	if mdl.GetLoadBalancerId() == "" {
		err := e.DescribeEdgeLoadBalancerByName(ctx, mdl.GetLoadBalancerName(), mdl)
		if err != nil {
			klog.Infof("[%s] find no loadbalancer by name", mdl.NamespacedName)
			return err
		}
		klog.Infof("[%s] find elb by name, LoadBalancerId [%s]", mdl.NamespacedName, mdl.GetLoadBalancerName())
	}

	if mdl.GetLoadBalancerId() != "" {
		err := e.DescribeEdgeLoadBalancerById(ctx, mdl.GetLoadBalancerId(), mdl)
		if err != nil {
			klog.Infof("[%s] find no loadbalancer by id", mdl.NamespacedName)
			return err
		}
		klog.Infof("[%s] find elb by id, LoadBalancerId [%s]", mdl.NamespacedName, mdl.GetLoadBalancerId())
	}
	return nil
}

func (e ELBProvider) CreateEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateCreateLoadBalancerRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.NetworkId = mdl.GetNetworkId()
	req.VSwitchId = mdl.GetVSwitchId()
	req.LoadBalancerSpec = mdl.LoadBalancerAttribute.LoadBalancerSpec
	req.LoadBalancerName = mdl.LoadBalancerAttribute.LoadBalancerName
	req.EnsRegionId = mdl.LoadBalancerAttribute.EnsRegionId
	req.PayType = mdl.LoadBalancerAttribute.PayType
	resp, err := e.auth.ELB.CreateLoadBalancer(req)
	if err != nil {
		return util.SDKError("CreateLoadBalancer", err)
	}
	mdl.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	return nil
}

func (e ELBProvider) SetEdgeLoadBalancerStatus(ctx context.Context, status string, mdl *elbmodel.EdgeLoadBalancer) error {
	if mdl.GetLoadBalancerId() == "" {
		return fmt.Errorf("loadbalancer id is empty")
	}
	req := ens.CreateSetLoadBalancerStatusRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = mdl.GetLoadBalancerId()
	req.LoadBalancerStatus = status

	_, err := e.auth.ELB.SetLoadBalancerStatus(req)
	if err != nil {
		return util.SDKError("SetLoadBalancerStatus", err)
	}
	return nil
}

func (e ELBProvider) DescribeEdgeLoadBalancerById(ctx context.Context, lbId string, mdl *elbmodel.EdgeLoadBalancer) error {
	req := ens.CreateDescribeLoadBalancerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	resp, err := e.auth.ELB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerAttribute", err)
	}
	if resp == nil {
		klog.Errorf("RequestId: %s, load balancer Id %s DescribeLoadBalancerAttribute response is nil", resp.RequestId, lbId)
		return fmt.Errorf("find no load balancer by id %s", lbId)
	}
	if resp.LoadBalancerId == "" {
		return fmt.Errorf("find no load balancer by elb Id %s", lbId)
	}
	loadELBRespId(resp, mdl)
	return nil
}

func (e ELBProvider) DescribeEdgeLoadBalancerByName(ctx context.Context, lbName string, mdl *elbmodel.EdgeLoadBalancer) error {
	if lbName == "" {
		return fmt.Errorf("loadbalancer name is empty")
	}
	req := ens.CreateDescribeLoadBalancersRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	resp, err := e.auth.ELB.DescribeLoadBalancers(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancers", err)
	}
	if len(resp.LoadBalancers.LoadBalancer) < 1 {
		return nil
	}

	for _, lb := range resp.LoadBalancers.LoadBalancer {
		if lb.LoadBalancerName == lbName {
			loadElbRespName(lb, mdl)
			break
		}
	}
	return nil
}

func (e ELBProvider) DeleteEdgeLoadBalancer(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	if mdl.GetLoadBalancerId() == "" {
		return fmt.Errorf("loadbalancer id is empty")
	}
	req := ens.CreateReleaseInstanceRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.InstanceId = mdl.GetLoadBalancerId()
	_, err := e.auth.ELB.ReleaseInstance(req)
	if err != nil {
		return util.SDKError("ReleaseInstance", err)
	}
	return nil
}

func loadELBRespId(resp *ens.DescribeLoadBalancerAttributeResponse, mdl *elbmodel.EdgeLoadBalancer) {
	mdl.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	mdl.LoadBalancerAttribute.LoadBalancerName = resp.LoadBalancerName
	mdl.LoadBalancerAttribute.EnsRegionId = resp.EnsRegionId
	mdl.LoadBalancerAttribute.NetworkId = resp.NetworkId
	mdl.LoadBalancerAttribute.VSwitchId = resp.VSwitchId
	mdl.LoadBalancerAttribute.Address = resp.Address
	mdl.LoadBalancerAttribute.PayType = resp.PayType
	mdl.LoadBalancerAttribute.AddressIPVersion = resp.AddressIPVersion
	mdl.LoadBalancerAttribute.LoadBalancerStatus = resp.LoadBalancerStatus
	mdl.LoadBalancerAttribute.CreateTime = resp.CreateTime
	mdl.LoadBalancerAttribute.LoadBalancerSpec = resp.LoadBalancerSpec
}
func loadElbRespName(resp ens.LoadBalancer, mdl *elbmodel.EdgeLoadBalancer) {
	mdl.LoadBalancerAttribute.LoadBalancerId = resp.LoadBalancerId
	mdl.LoadBalancerAttribute.LoadBalancerName = resp.LoadBalancerName
	mdl.LoadBalancerAttribute.EnsRegionId = resp.EnsRegionId
	mdl.LoadBalancerAttribute.NetworkId = resp.NetworkId
	mdl.LoadBalancerAttribute.VSwitchId = resp.VSwitchId
	mdl.LoadBalancerAttribute.Address = resp.Address
	mdl.LoadBalancerAttribute.PayType = resp.PayType
	mdl.LoadBalancerAttribute.AddressIPVersion = resp.AddressIPVersion
	mdl.LoadBalancerAttribute.LoadBalancerStatus = resp.LoadBalancerStatus
	mdl.LoadBalancerAttribute.CreateTime = resp.CreateTime
	mdl.LoadBalancerAttribute.IsUserManaged = false
}
