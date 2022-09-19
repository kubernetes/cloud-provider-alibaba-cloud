package vmock

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/klog/v2"
)

func NewMockCLB(
	auth *base.ClientMgr,
) *MockCLB {
	return &MockCLB{auth: auth}
}

var _ prvd.ILoadBalancer = &MockCLB{}

type MockCLB struct {
	auth *base.ClientMgr
}

const (
	LoadBalancerIP = "47.168.0.1"
	ExistLBID      = "lb-exist-id"
	ExistVGroupID  = "rsp-reuse-id"
)

func (m *MockCLB) FindLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	if mdl.LoadBalancerAttribute.LoadBalancerId != "" {
		klog.Infof("[%s] find loadbalancer by id, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return m.DescribeLoadBalancer(ctx, mdl)
	}

	if mdl.LoadBalancerAttribute.LoadBalancerName == "a5e4dbfc9c2ae4642b0335607860aef6" {
		mdl.LoadBalancerAttribute.LoadBalancerId = ExistLBID
		klog.Infof("[%s] find loadbalancer by name, LoadBalancerId [%s]",
			mdl.NamespacedName, mdl.LoadBalancerAttribute.LoadBalancerId)
		return m.DescribeLoadBalancer(ctx, mdl)
	}
	return nil
}
func (m *MockCLB) CreateLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	mdl.LoadBalancerAttribute.LoadBalancerId = "lb-new-created-id"
	return nil
}
func (m *MockCLB) DescribeLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	mdl.LoadBalancerAttribute.LoadBalancerName = "lb-name"
	mdl.LoadBalancerAttribute.LoadBalancerSpec = "slb.s1.small"
	mdl.LoadBalancerAttribute.Address = LoadBalancerIP
	mdl.LoadBalancerAttribute.AddressType = "internet"
	mdl.LoadBalancerAttribute.AddressIPVersion = "ipv4"
	mdl.LoadBalancerAttribute.NetworkType = "classic"
	mdl.LoadBalancerAttribute.VpcId = ""
	mdl.LoadBalancerAttribute.VSwitchId = ""
	mdl.LoadBalancerAttribute.Bandwidth = 5120
	mdl.LoadBalancerAttribute.MasterZoneId = "cn-hangzhou-i"
	mdl.LoadBalancerAttribute.SlaveZoneId = "cn-hangzhou-h"
	mdl.LoadBalancerAttribute.DeleteProtection = "on"
	mdl.LoadBalancerAttribute.ModificationProtectionStatus = "ConsoleProtection"
	mdl.LoadBalancerAttribute.ModificationProtectionReason = "managed.by.ack"
	mdl.LoadBalancerAttribute.ResourceGroupId = "rg-id"
	mdl.LoadBalancerAttribute.InternetChargeType = "paybytraffic"
	return nil
}
func (m *MockCLB) DeleteLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return nil
}
func (m *MockCLB) ModifyLoadBalancerInstanceSpec(ctx context.Context, lbId string, spec string) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerDeleteProtection(ctx context.Context, lbId string, flag string) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerName(ctx context.Context, lbId string, name string) error {
	return nil
}
func (m *MockCLB) ModifyLoadBalancerInternetSpec(ctx context.Context, lbId string, chargeType string, bandwidth int) error {
	return nil
}
func (m *MockCLB) ModifyLoadBalancerInstanceChargeType(ctx context.Context, lbId string, instanceChargeType string) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerModificationProtection(ctx context.Context, lbId string, flag string) error {
	return nil
}
func (m *MockCLB) TagCLBResource(ctx context.Context, resourceId string, tags []tag.Tag) error {
	return nil
}
func (m *MockCLB) ListCLBTagResources(ctx context.Context, lbId string) ([]tag.Tag, error) {
	return nil, nil
}

// Listener
func (m *MockCLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {
	if lbId == ExistLBID {
		listeners := []model.ListenerAttribute{
			{
				ListenerPort:        443,
				Description:         "k8s/443/test/default/clusterid",
				Status:              "running",
				ListenerForward:     model.OffFlag,
				VGroupName:          "k8s/443/test/default/clusterid",
				VGroupId:            "rsp-https-443",
				Protocol:            model.HTTPS,
				Scheduler:           "rr",
				CertId:              "cert-id",
				Bandwidth:           -1,
				EnableHttp2:         "on",
				StickySession:       "off",
				XForwardedFor:       "off",
				AclId:               "acl-id",
				AclType:             "white",
				AclStatus:           model.OnFlag,
				ConnectionDrain:     model.OffFlag,
				IdleTimeout:         15,
				RequestTimeout:      60,
				HealthyThreshold:    5,
				UnhealthyThreshold:  4,
				HealthCheckTimeout:  10,
				HealthCheck:         model.OnFlag,
				HealthCheckDomain:   "foo.bar.com",
				HealthCheckURI:      "/test/index.html",
				HealthCheckHttpCode: "http_2xx",
				HealthCheckMethod:   "head",
			},
			{
				ListenerPort:        8080,
				Description:         "k8s/8080/test/default/clusterid",
				Status:              "running",
				ListenerForward:     model.OffFlag,
				VGroupName:          "k8s/8080/test/default/clusterid",
				VGroupId:            "rsp-http-8080",
				Protocol:            model.HTTP,
				Scheduler:           "rr",
				Bandwidth:           -1,
				StickySession:       "off",
				XForwardedFor:       "off",
				AclId:               "acl-id",
				AclType:             "black",
				AclStatus:           model.OnFlag,
				ConnectionDrain:     model.OffFlag,
				IdleTimeout:         15,
				RequestTimeout:      60,
				HealthyThreshold:    5,
				UnhealthyThreshold:  4,
				HealthCheckTimeout:  10,
				HealthCheck:         model.OnFlag,
				HealthCheckDomain:   "foo.bar.com",
				HealthCheckURI:      "/test/index.html",
				HealthCheckHttpCode: "http_2xx",
				HealthCheckMethod:   "head",
			},
			{
				ListenerPort:        80,
				Description:         "k8s/80/test/default/clusterid",
				Status:              "running",
				ListenerForward:     model.OffFlag,
				VGroupName:          "k8s/80/test/default/clusterid",
				VGroupId:            "rsp-tcp-80",
				Protocol:            model.TCP,
				Scheduler:           "rr",
				Bandwidth:           -1,
				AclStatus:           model.OffFlag,
				ConnectionDrain:     model.OffFlag,
				HealthyThreshold:    5,
				UnhealthyThreshold:  4,
				HealthCheckTimeout:  10,
				HealthCheckDomain:   "foo.bar.com",
				HealthCheckURI:      "/test/index.html",
				HealthCheckHttpCode: "http_2xx",
				HealthCheckMethod:   "head",
			},
			{
				ListenerPort:       53,
				Description:        "k8s/53/test/default/clusterid",
				Status:             model.Stopped,
				ListenerForward:    model.OffFlag,
				VGroupName:         "k8s/53/test/default/clusterid",
				VGroupId:           "rsp-for-53",
				Protocol:           model.UDP,
				Scheduler:          "rr",
				Bandwidth:          -1,
				AclId:              "acl-wrong-id",
				AclType:            "white",
				AclStatus:          model.OnFlag,
				ConnectionDrain:    model.OffFlag,
				HealthyThreshold:   5,
				UnhealthyThreshold: 4,
				HealthCheckTimeout: 10,
			},
		}

		for i := range listeners {
			namedKey, err := model.LoadListenerNamedKey(listeners[i].Description)
			if err != nil {
				listeners[i].IsUserManaged = true
			}
			listeners[i].NamedKey = namedKey
		}
		return listeners, nil
	}

	return nil, nil
}
func (m *MockCLB) StartLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	return nil
}
func (m *MockCLB) StopLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	return nil
}
func (m *MockCLB) DeleteLoadBalancerListener(ctx context.Context, lbId string, port int) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerTCPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerTCPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerUDPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerUDPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerHTTPListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerHTTPListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) CreateLoadBalancerHTTPSListener(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}
func (m *MockCLB) SetLoadBalancerHTTPSListenerAttribute(ctx context.Context, lbId string, listener model.ListenerAttribute) error {
	return nil
}

// VServerGroup
func (m *MockCLB) DescribeVServerGroups(ctx context.Context, lbId string) ([]model.VServerGroup, error) {
	if lbId == ExistLBID {
		vgroups := []model.VServerGroup{
			{
				ServicePort: v1.ServicePort{
					Name:       "udp",
					Port:       53,
					TargetPort: intstr.FromInt(53),
				},

				VGroupId:   "rsp-udp-53",
				VGroupName: "k8s/53/test/default/clusterid",
			},
			{
				ServicePort: v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
				VGroupId:   "rsp-tcp-80",
				VGroupName: "k8s/80/test/default/clusterid",
			},
			{
				ServicePort: v1.ServicePort{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				VGroupId:   "rsp-http-8080",
				VGroupName: "k8s/8080/test/default/clusterid",
			},
			{
				ServicePort: v1.ServicePort{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(443),
				},
				VGroupId:   "rsp-https-443",
				VGroupName: "k8s/443/test/default/clusterid",
			},
			{
				VGroupId:   ExistVGroupID,
				VGroupName: ExistVGroupID,
			},
		}
		for i := range vgroups {
			namedKey, err := model.LoadVGroupNamedKey(vgroups[i].VGroupName)
			if err != nil {
				vgroups[i].IsUserManaged = true
			}
			vgroups[i].NamedKey = namedKey
		}
		return vgroups, nil
	}

	return nil, nil
}
func (m *MockCLB) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
	return nil
}
func (m *MockCLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	switch vGroupId {
	case "rsp-https-443":
		return model.VServerGroup{
			VGroupId:   "rsp-https-443",
			VGroupName: "k8s/443/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					Description: "k8s/443/test/default/clusterid",
					ServerId:    "eni-id",
					ServerIp:    "10.96.0.15",
					Weight:      100,
					Port:        443,
					Type:        "eni",
				},
				{
					Description: "k8s/443/test/default/clusterid",
					ServerId:    "ecs-id",
					ServerIp:    "",
					Weight:      50,
					Port:        443,
					Type:        "ecs",
				},
			}}, nil
	case "rsp-http-8080":
		return model.VServerGroup{
			VGroupId:   "rsp-http-8080",
			VGroupName: "k8s/8080/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					Description: "k8s/8080/test/default/clusterid",
					ServerId:    "eni-id",
					ServerIp:    "10.96.0.15",
					Weight:      100,
					Port:        8080,
					Type:        "eni",
				},
				{
					Description: "k8s/8080/test/default/clusterid",
					ServerId:    "eni-id",
					ServerIp:    "10.96.0.16",
					Weight:      100,
					Port:        8080,
					Type:        "eni",
				},
			}}, nil
	case "rsp-tcp-80":
		return model.VServerGroup{
			VGroupId:   "rsp-tcp-80",
			VGroupName: "k8s/80/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					Description: "k8s/80/test/default/clusterid",
					ServerId:    "eni-id",
					ServerIp:    "10.96.0.15",
					Weight:      80,
					Port:        80,
					Type:        "eni",
				},
			},
		}, nil
	case "rsp-udp-53":
		return model.VServerGroup{
			VGroupId:   "rsp-udp-53",
			VGroupName: "k8s/53/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					Description: "k8s/53/test/default/clusterid",
					ServerId:    "eni-id",
					ServerIp:    "10.96.0.15",
					Weight:      100,
					Port:        54,
					Type:        "eni",
				},
			},
		}, nil
	case ExistVGroupID:
		return model.VServerGroup{
			VGroupId:   ExistVGroupID,
			VGroupName: ExistVGroupID,
			Backends: []model.BackendAttribute{
				{
					Description: "",
					ServerId:    "eni-id",
					ServerIp:    "10.96.0.16",
					Weight:      100,
					Port:        88,
					Type:        "eni",
				},
			}}, nil
	default:
		return model.VServerGroup{}, nil
	}
}
func (m *MockCLB) DeleteVServerGroup(ctx context.Context, vGroupId string) error {
	return nil
}
func (m *MockCLB) AddVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	return nil
}
func (m *MockCLB) RemoveVServerGroupBackendServers(ctx context.Context, vGroupId string, backends string) error {
	return nil
}
func (m *MockCLB) SetVServerGroupAttribute(ctx context.Context, vGroupId string, backends string) error {
	return nil
}
func (m *MockCLB) ModifyVServerGroupBackendServers(ctx context.Context, vGroupId string, old string, new string) error {
	return nil
}
