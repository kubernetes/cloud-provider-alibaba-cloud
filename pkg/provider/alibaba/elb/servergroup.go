package elb

import (
	"context"
	"strconv"

	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ens"
)

func (e ELBProvider) FindBackendFromLoadBalancer(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	req := ens.CreateDescribeLoadBalancerAttributeRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	resp, err := e.auth.ELB.DescribeLoadBalancerAttribute(req)
	if err != nil {
		return util.SDKError("DescribeLoadBalancerAttribute", err)
	}
	loadServerGroupResponse(resp, sg)
	return nil
}
func (e ELBProvider) UpdateEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	BackendServers := make([]ens.SetBackendServersBackendServers, 0, len(sg.Backends))
	for _, backend := range sg.Backends {
		BackendServers = append(BackendServers, ens.SetBackendServersBackendServers{
			ServerId: backend.ServerId,
			Weight:   strconv.Itoa(backend.Weight),
			Type:     backend.Type,
		})
	}
	req := ens.CreateSetBackendServersRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.BackendServers = &BackendServers
	_, err := e.auth.ELB.SetBackendServers(req)
	if err != nil {
		return util.SDKError("SetBackendServers", err)
	}
	return nil
}

func (e ELBProvider) AddBackendToEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	BackendServers := make([]ens.AddBackendServersBackendServers, 0, len(sg.Backends))
	for _, backend := range sg.Backends {
		BackendServers = append(BackendServers, ens.AddBackendServersBackendServers{
			ServerId: backend.ServerId,
			Weight:   strconv.Itoa(backend.Weight),
			Type:     backend.Type,
			Ip:       backend.ServerIp,
			Port:     backend.Port,
		})
	}
	req := ens.CreateAddBackendServersRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.BackendServers = &BackendServers
	_, err := e.auth.ELB.AddBackendServers(req)
	if err != nil {
		return util.SDKError("AddBackendServers", err)
	}
	return nil
}

func (e ELBProvider) RemoveBackendFromEdgeServerGroup(ctx context.Context, lbId string, sg *elbmodel.EdgeServerGroup) error {
	BackendServers := make([]ens.RemoveBackendServersBackendServers, 0, len(sg.Backends))
	for _, backend := range sg.Backends {
		BackendServers = append(BackendServers, ens.RemoveBackendServersBackendServers{
			ServerId: backend.ServerId,
			Weight:   strconv.Itoa(backend.Weight),
			Type:     backend.Type,
			Ip:       backend.ServerIp,
			Port:     backend.Port,
		})
	}
	req := ens.CreateRemoveBackendServersRequest()
	req.Scheme = "https"
	req.ConnectTimeout = connectionTimeout
	req.ReadTimeout = readTimeout
	req.LoadBalancerId = lbId
	req.BackendServers = &BackendServers
	_, err := e.auth.ELB.RemoveBackendServers(req)
	if err != nil {
		return util.SDKError("RemoveBackendServers", err)
	}
	return nil
}

func loadServerGroupResponse(backends *ens.DescribeLoadBalancerAttributeResponse, sg *elbmodel.EdgeServerGroup) {
	for _, server := range backends.BackendServers {
		backend := elbmodel.EdgeBackendAttribute{
			ServerId: server.ServerId,
			ServerIp: server.Ip,
			Type:     server.Type,
			Port:     server.Port,
			Weight:   server.Weight,
		}
		sg.Backends = append(sg.Backends, backend)
	}
}
