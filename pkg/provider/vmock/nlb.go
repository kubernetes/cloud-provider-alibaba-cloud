package vmock

import (
	"context"
	"github.com/alibabacloud-go/tea/tea"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/klog/v2"
)

func NewMockNLB(
	auth *base.ClientMgr,
) *MockNLB {
	return &MockNLB{auth: auth}
}

var _ prvd.INLB = &MockNLB{}

type MockNLB struct {
	auth *base.ClientMgr
}

const (
	NLBDNSName = "nlb-id.cn-hangzhou.nlb.aliyuncs.com"
	ExistNLBID = "nlb-exist-id"
)

func (m MockNLB) TagNLBResource(ctx context.Context, resourceId string, resourceType nlbmodel.TagResourceType, tags []tag.Tag) error {
	return nil
}

func (m MockNLB) ListNLBTagResources(ctx context.Context, lbId string) ([]tag.Tag, error) {
	return nil, nil
}

func (m MockNLB) FindNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find loadbalancer by id, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return m.DescribeNLB(ctx, mdl)
	}

	if mdl.LoadBalancerAttribute.Name == "a5e4dbfc9c2ae4642b0335607860aef6" {
		mdl.LoadBalancerAttribute.LoadBalancerId = ExistNLBID
		klog.Infof("[%s] find loadbalancer by name, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return m.DescribeNLB(ctx, mdl)
	}
	return nil
}

func (m MockNLB) DescribeNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mdl.LoadBalancerAttribute.LoadBalancerId = ExistNLBID
	mdl.LoadBalancerAttribute.Name = "nlb-name"
	mdl.LoadBalancerAttribute.AddressType = nlbmodel.InternetAddressType
	mdl.LoadBalancerAttribute.DNSName = NLBDNSName
	mdl.LoadBalancerAttribute.AddressIpVersion = nlbmodel.IPv4
	mdl.LoadBalancerAttribute.VpcId = "vpc-id"
	mdl.LoadBalancerAttribute.ResourceGroupId = "rg-id"
	mdl.LoadBalancerAttribute.ZoneMappings = []nlbmodel.ZoneMapping{
		{
			ZoneId:    "cn-hangzhou-a",
			VSwitchId: "vsw-1",
		},
		{
			ZoneId:    "cn-hangzhou-b",
			VSwitchId: "vsw-2",
		},
	}
	return nil
}

func (m MockNLB) CreateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	mdl.LoadBalancerAttribute.LoadBalancerId = "nlb-new-created-id"
	return nil
}

func (m MockNLB) DeleteNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return nil
}

func (m MockNLB) UpdateNLB(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return nil
}

func (m MockNLB) UpdateNLBAddressType(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return nil
}

func (m MockNLB) UpdateNLBZones(ctx context.Context, mdl *nlbmodel.NetworkLoadBalancer) error {
	return nil
}

func (m MockNLB) ListNLBServerGroups(ctx context.Context, tags []tag.Tag) ([]*nlbmodel.ServerGroup, error) {
	found := false
	for _, t := range tags {
		if t.Key == helper.TAGKEY && t.Value == "a5e4dbfc9c2ae4642b0335607860aef6" {
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}
	sgs := []*nlbmodel.ServerGroup{
		{
			ServerGroupId:           "rsp-udp-53",
			ServerGroupName:         "k8s.53.nlb.default.clusterid",
			Protocol:                nlbmodel.UDP,
			ServerGroupType:         nlbmodel.InstanceServerGroupType,
			AddressIPVersion:        nlbmodel.IPv4,
			ResourceGroupId:         "rg-id",
			Scheduler:               "rr",
			ConnectionDrainEnabled:  tea.Bool(true),
			ConnectionDrainTimeout:  30,
			PreserveClientIpEnabled: tea.Bool(true),
			HealthCheckConfig: &nlbmodel.HealthCheckConfig{
				HealthCheckEnabled:        tea.Bool(true),
				HealthCheckType:           nlbmodel.TCP,
				HealthCheckConnectTimeout: 3,
				HealthCheckInterval:       5,
				HealthyThreshold:          6,
				UnhealthyThreshold:        5,
			},
			Servers: []nlbmodel.ServerGroupServer{
				{
					ServerId:   "ecs-id-1",
					ServerType: "Ecs",
					Port:       53,
					Weight:     53,
				},
				{
					ServerId:   "ecs-id-2",
					ServerType: "Ecs",
					Port:       53,
					Weight:     53,
				},
			},
		},
		{
			ServerGroupId:           "rsp-tcp-80",
			ServerGroupName:         "k8s.80.nlb.default.clusterid",
			Protocol:                nlbmodel.TCP,
			ServerGroupType:         nlbmodel.InstanceServerGroupType,
			AddressIPVersion:        nlbmodel.IPv4,
			ResourceGroupId:         "rg-id",
			Scheduler:               "sch",
			ConnectionDrainEnabled:  tea.Bool(false),
			PreserveClientIpEnabled: tea.Bool(false),
			HealthCheckConfig: &nlbmodel.HealthCheckConfig{
				HealthCheckEnabled:        tea.Bool(true),
				HealthCheckType:           nlbmodel.TCP,
				HealthCheckConnectPort:    80,
				HealthCheckConnectTimeout: 4,
				HealthCheckInterval:       6,
				HealthyThreshold:          1,
				UnhealthyThreshold:        2,
			},
		},
		{
			ServerGroupId:           "rsp-tcpssl-443",
			ServerGroupName:         "k8s.443.nlb.default.clusterid",
			Protocol:                nlbmodel.TCPSSL,
			ServerGroupType:         nlbmodel.InstanceServerGroupType,
			AddressIPVersion:        nlbmodel.IPv4,
			ResourceGroupId:         "rg-id",
			Scheduler:               "rr",
			ConnectionDrainEnabled:  tea.Bool(true),
			ConnectionDrainTimeout:  30,
			PreserveClientIpEnabled: tea.Bool(true),
			HealthCheckConfig: &nlbmodel.HealthCheckConfig{
				HealthCheckEnabled:        tea.Bool(true),
				HealthCheckType:           nlbmodel.TCP,
				HealthCheckConnectTimeout: 3,
				HealthCheckInterval:       5,
				HealthyThreshold:          6,
				UnhealthyThreshold:        5,
			},
			Servers: []nlbmodel.ServerGroupServer{
				{
					ServerId:   "ecs-id-1",
					ServerType: "Ecs",
					Port:       443,
					Weight:     100,
				},
				{
					ServerId:   "ecs-id-to-be-deleted",
					ServerType: "Ecs",
					Port:       443,
					Weight:     100,
				},
			},
		},
	}
	for i := range sgs {
		namedKey, err := nlbmodel.LoadNLBSGNamedKey(sgs[i].ServerGroupName)
		if err != nil {
			sgs[i].IsUserManaged = true
		}
		sgs[i].NamedKey = namedKey
	}
	return sgs, nil
}

func (m MockNLB) CreateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	sg.ServerGroupId = "sg-created-id"
	return nil
}

func (m MockNLB) DeleteNLBServerGroup(ctx context.Context, sgId string) error {
	return nil
}

func (m MockNLB) UpdateNLBServerGroup(ctx context.Context, sg *nlbmodel.ServerGroup) error {
	return nil
}

func (m MockNLB) AddNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	return nil
}

func (m MockNLB) RemoveNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	return nil
}

func (m MockNLB) UpdateNLBServers(ctx context.Context, sgId string, backends []nlbmodel.ServerGroupServer) error {
	return nil
}

func (m MockNLB) ListNLBListeners(ctx context.Context, lbId string) ([]*nlbmodel.ListenerAttribute, error) {
	if lbId == ExistNLBID {
		listeners := []*nlbmodel.ListenerAttribute{
			{
				ListenerId:           "lsn-tcpssl-id@443",
				ListenerPort:         443,
				ListenerDescription:  "k8s.443.TCPSSL.nlb.default.clusterid",
				ListenerStatus:       "running",
				ServerGroupName:      "k8s.443.nlb.default.clusterid",
				ServerGroupId:        "rsp-tcpssl-443",
				ListenerProtocol:     nlbmodel.TCPSSL,
				CaEnabled:            tea.Bool(true),
				CertificateIds:       []string{"cert-id"},
				CaCertificateIds:     []string{"cacert-id"},
				SecurityPolicyId:     "tls_cipher_policy_1_2",
				ProxyProtocolEnabled: tea.Bool(true),
				IdleTimeout:          15,
				Cps:                  tea.Int32(60),
			},
			{
				ListenerId:           "lsn-udp-id@53",
				ListenerPort:         53,
				ListenerDescription:  "k8s.53.UDP.nlb.default.clusterid",
				ListenerStatus:       "stopped",
				ServerGroupName:      "k8s.53.nlb.default.clusterid",
				ServerGroupId:        "rsp-udp-53",
				ListenerProtocol:     nlbmodel.UDP,
				ProxyProtocolEnabled: tea.Bool(false),
				IdleTimeout:          60,
				Cps:                  tea.Int32(30),
			},
			{
				ListenerId:           "lsn-tcp-id@80",
				ListenerPort:         80,
				ListenerDescription:  "k8s.80.TCP.nlb.default.clusterid",
				ListenerStatus:       "running",
				ServerGroupName:      "k8s.80.nlb.default.clusterid",
				ServerGroupId:        "rsp-tcp-80",
				ListenerProtocol:     nlbmodel.TCP,
				ProxyProtocolEnabled: tea.Bool(true),
				IdleTimeout:          15,
				Cps:                  tea.Int32(60),
			},
			{
				ListenerId:           "lsn-tcp-id@82",
				ListenerPort:         82,
				ListenerDescription:  "-",
				ListenerStatus:       "running",
				ServerGroupId:        "rsp-user-managed-id",
				ListenerProtocol:     nlbmodel.TCP,
				ProxyProtocolEnabled: tea.Bool(true),
			},
		}

		for i := range listeners {
			namedKey, err := nlbmodel.LoadNLBListenerNamedKey(listeners[i].ListenerDescription)
			if err != nil {
				listeners[i].IsUserManaged = true
			}
			listeners[i].NamedKey = namedKey
		}
		return listeners, nil
	}
	return nil, nil
}

func (m MockNLB) CreateNLBListener(ctx context.Context, lbId string, lis *nlbmodel.ListenerAttribute) error {
	return nil
}

func (m MockNLB) UpdateNLBListener(ctx context.Context, lis *nlbmodel.ListenerAttribute) error {
	return nil
}

func (m MockNLB) DeleteNLBListener(ctx context.Context, listenerId string) error {
	return nil
}

func (m MockNLB) StartNLBListener(ctx context.Context, listenerId string) error {
	return nil
}
