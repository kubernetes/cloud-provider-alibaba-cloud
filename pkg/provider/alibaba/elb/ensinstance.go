package elb

import (
	"context"
	"fmt"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ens"
)

func (e ELBProvider) GetEnsRegionIdByNetwork(ctx context.Context, networkId string) (string, error) {
	if networkId == "" {
		return "", fmt.Errorf(" network id is empty")
	}
	req := ens.CreateDescribeNetworksRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.NetworkId = networkId
	resp, err := e.auth.ELB.DescribeNetworks(req)
	if err != nil {
		return "", util.SDKError("DescribeNetworks", err)
	}
	if len(resp.Networks.Network) == 0 || resp.Networks.Network[0].EnsRegionId == "" {
		return "", fmt.Errorf("edge network %s region is not found", networkId)
	}
	return resp.Networks.Network[0].EnsRegionId, nil
}

func (e ELBProvider) FindEnsInstancesByNetwork(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) (map[string]string, error) {
	if mdl.GetNetworkId() == "" || mdl.GetVSwitchId() == "" {
		return nil, fmt.Errorf("edge loadbalancer mode lacks significant attribute: network and vitual switch")
	}
	ret := make(map[string]string, 0)
	req := ens.CreateDescribeInstancesRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.NetworkId = mdl.GetNetworkId()
	req.VSwitchId = mdl.GetVSwitchId()
	resp, err := e.auth.ELB.DescribeInstances(req)
	if err != nil {
		return ret, util.SDKError("DescribeInstances", err)
	}
	for _, instance := range resp.Instances.Instance {
		ret[instance.InstanceId] = instance.Status
	}

	return ret, nil
}

func (e ELBProvider) FindNetWorkAndVSwitchByLoadBalancerId(ctx context.Context, lbId string) ([]string, error) {
	ret := make([]string, 0)
	req := ens.CreateDescribeLoadBalancerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	resp, err := e.auth.ELB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return ret, util.SDKError("DescribeLoadBalancerAttribute", err)
	}
	ret = append(ret, resp.NetworkId, resp.VSwitchId, resp.EnsRegionId)
	return ret, nil
}

func (e ELBProvider) DescribeNetwork(ctx context.Context, mdl *elbmodel.EdgeLoadBalancer) error {
	if mdl.GetNetworkId() == "" || mdl.GetVSwitchId() == "" {
		return fmt.Errorf("network or vswitch id is empty")
	}
	found := false
	req := ens.CreateDescribeNetworksRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.NetworkId = mdl.GetNetworkId()
	resp, err := e.auth.ELB.DescribeNetworks(req)
	if err != nil {
		return util.SDKError("DescribeNetworks", err)
	}
	networks := resp.Networks.Network
	if resp.TotalCount != 1 || len(networks) != 1 {
		return fmt.Errorf("find no network %s", mdl.GetNetworkId())
	}

	if resp.Networks.Network[0].VSwitchIds.VSwitchId == nil {
		return fmt.Errorf("find no vswitch %s in network %s", mdl.GetVSwitchId(), mdl.GetNetworkId())
	}
	for _, vsw := range networks[0].VSwitchIds.VSwitchId {
		if vsw == mdl.LoadBalancerAttribute.VSwitchId {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("find no vswitch %s in network %s", mdl.GetVSwitchId(), mdl.GetNetworkId())
	}
	mdl.LoadBalancerAttribute.EnsRegionId = networks[0].EnsRegionId
	return nil
}
