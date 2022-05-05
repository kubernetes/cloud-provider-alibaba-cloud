package vmock

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
)

func NewMockCLB(
	auth *base.ClientMgr,
) *MockCLB {
	return &MockCLB{auth: auth}
}

type MockCLB struct {
	auth *base.ClientMgr
}

func (m *MockCLB) FindLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	mdl.LoadBalancerAttribute.LoadBalancerId = "lb-id"
	return nil
}
func (m *MockCLB) CreateLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
	return nil
}
func (m *MockCLB) DescribeLoadBalancer(ctx context.Context, mdl *model.LoadBalancer) error {
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
func (m *MockCLB) SetLoadBalancerModificationProtection(ctx context.Context, lbId string, flag string) error {
	return nil
}
func (m *MockCLB) AddTags(ctx context.Context, lbId string, tags string) error {
	return nil
}

func (m *MockCLB) DescribeTags(ctx context.Context, lbId string) ([]model.Tag, error) {
	return nil, nil
}

// Listener
func (m *MockCLB) DescribeLoadBalancerListeners(ctx context.Context, lbId string) ([]model.ListenerAttribute, error) {

	listeners := []model.ListenerAttribute{
		{
			ListenerPort:        443,
			Description:         "k8s/443/test/default/clusterid",
			Status:              "running",
			ListenerForward:     model.OffFlag,
			VGroupName:          "k8s/443/test/default/clusterid",
			VGroupId:            "rsp-443bad4h3w111",
			Protocol:            model.HTTPS,
			Scheduler:           "rr",
			CertId:              "1609088142240951_17ce3ecf535_-97773175_-xxxxxxxxxx",
			Bandwidth:           -1,
			EnableHttp2:         "on",
			StickySession:       "off",
			XForwardedFor:       "off",
			AclId:               "acl-bp1oq8t0wn5ljniegxxxx",
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
			VGroupId:            "rsp-8080ad4h3w111",
			Protocol:            model.HTTP,
			Scheduler:           "rr",
			Bandwidth:           -1,
			StickySession:       "off",
			XForwardedFor:       "off",
			AclId:               "acl-bp1oq8t0wn5ljniegxxxx",
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
			ListenerPort:        80,
			Description:         "k8s/80/test/default/clusterid",
			Status:              "running",
			ListenerForward:     model.OffFlag,
			VGroupName:          "k8s/80/test/default/clusterid",
			VGroupId:            "rsp-801bad4h3w111",
			Protocol:            model.TCP,
			Scheduler:           "rr",
			Bandwidth:           -1,
			AclId:               "acl-bp1oq8t0wn5ljniegxxxx",
			AclType:             "white",
			AclStatus:           model.OnFlag,
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
			Status:             "running",
			ListenerForward:    model.OffFlag,
			VGroupName:         "k8s/53/test/default/clusterid",
			VGroupId:           "rsp-531bad4h3w111",
			Protocol:           model.UDP,
			Scheduler:          "rr",
			Bandwidth:          -1,
			AclId:              "acl-bp1oq8t0wn5ljniegxxxx",
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
	vgroups := []model.VServerGroup{
		{
			IsUserManaged: false,
			ServicePort: v1.ServicePort{
				Name: "udp",
				Port: 53,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 53,
				},
			},

			VGroupId:   "rsp-531bad4h3w111",
			VGroupName: "k8s/53/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					IsUserManaged: false,
					Description:   "k8s/53/test/default/clusterid",
					ServerId:      "eni-id",
					ServerIp:      "10.96.0.15",
					Weight:        100,
					Port:          53,
					Type:          "eni",
				},
			},
		},
		{
			IsUserManaged: false,
			ServicePort: v1.ServicePort{
				Name: "tcp",
				Port: 80,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 80,
				},
			},

			VGroupId:   "rsp-801bad4h3w111",
			VGroupName: "k8s/80/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					IsUserManaged: false,
					Description:   "k8s/80/test/default/clusterid",
					ServerId:      "eni-id",
					ServerIp:      "10.96.0.15",
					Weight:        100,
					Port:          80,
					Type:          "eni",
				},
			},
		},
		{
			IsUserManaged: false,
			ServicePort: v1.ServicePort{
				Name: "http",
				Port: 8080,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8080,
				},
			},

			VGroupId:   "rsp-8080ad4h3w111",
			VGroupName: "k8s/8080/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					IsUserManaged: false,
					Description:   "k8s/8080/test/default/clusterid",
					ServerId:      "eni-id",
					ServerIp:      "10.96.0.15",
					Weight:        100,
					Port:          8080,
					Type:          "eni",
				},
			},
		},
		{
			IsUserManaged: false,
			ServicePort: v1.ServicePort{
				Name: "https",
				Port: 443,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 443,
				},
			},

			VGroupId:   "rsp-443bad4h3w111",
			VGroupName: "k8s/443/test/default/clusterid",
			Backends: []model.BackendAttribute{
				{
					IsUserManaged: false,
					Description:   "k8s/443/test/default/clusterid",
					ServerId:      "eni-id",
					ServerIp:      "10.96.0.15",
					Weight:        100,
					Port:          443,
					Type:          "eni",
				},
			},
		},
	}
	return vgroups, nil
}
func (m *MockCLB) CreateVServerGroup(ctx context.Context, vg *model.VServerGroup, lbId string) error {
	return nil
}
func (m *MockCLB) DescribeVServerGroupAttribute(ctx context.Context, vGroupId string) (model.VServerGroup, error) {
	return model.VServerGroup{}, nil
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
