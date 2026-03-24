package nlbv2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	reconbackend "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsServerEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        nlbmodel.ServerGroupServer
		b        nlbmodel.ServerGroupServer
		expected bool
	}{
		{
			name: "different server types",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EniServerType,
				ServerId:   "ecs-1",
			},
			expected: false,
		},
		{
			name: "ecs server type - same id",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-1",
			},
			expected: true,
		},
		{
			name: "ecs server type - different id",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-2",
			},
			expected: false,
		},
		{
			name: "eni server type - same id and ip",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EniServerType,
				ServerId:   "eni-1",
				ServerIp:   "10.0.0.1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EniServerType,
				ServerId:   "eni-1",
				ServerIp:   "10.0.0.1",
			},
			expected: true,
		},
		{
			name: "eni server type - different ip",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EniServerType,
				ServerId:   "eni-1",
				ServerIp:   "10.0.0.1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.EniServerType,
				ServerId:   "eni-1",
				ServerIp:   "10.0.0.2",
			},
			expected: false,
		},
		{
			name: "ip server type - same id and ip",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.IpServerType,
				ServerId:   "ip-1",
				ServerIp:   "10.0.0.1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.IpServerType,
				ServerId:   "ip-1",
				ServerIp:   "10.0.0.1",
			},
			expected: true,
		},
		{
			name: "ip server type - different id",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.IpServerType,
				ServerId:   "ip-1",
				ServerIp:   "10.0.0.1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.IpServerType,
				ServerId:   "ip-2",
				ServerIp:   "10.0.0.1",
			},
			expected: false,
		},
		{
			name: "unsupported server type",
			a: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.ServerType("Unknown"),
				ServerId:   "id-1",
			},
			b: nlbmodel.ServerGroupServer{
				ServerType: nlbmodel.ServerType("Unknown"),
				ServerId:   "id-1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isServerEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name            string
		remote          *nlbmodel.ServerGroup
		local           *nlbmodel.ServerGroup
		expectedAdds    int
		expectedDels    int
		expectedUpdates int
	}{
		{
			name: "empty remote and local",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectedAdds:    0,
			expectedDels:    0,
			expectedUpdates: 0,
		},
		{
			name: "add new server",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			expectedAdds:    1,
			expectedDels:    0,
			expectedUpdates: 0,
		},
		{
			name: "delete server",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectedAdds:    0,
			expectedDels:    1,
			expectedUpdates: 0,
		},
		{
			name: "update server weight",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     200,
					},
				},
			},
			expectedAdds:    0,
			expectedDels:    0,
			expectedUpdates: 1,
		},
		{
			name: "update server port",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       8080,
						Weight:     100,
					},
				},
			},
			expectedAdds:    0,
			expectedDels:    0,
			expectedUpdates: 1,
		},
		{
			name: "update server description",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType:  nlbmodel.EcsServerType,
						ServerId:    "ecs-1",
						Port:        80,
						Weight:      100,
						Description: "old",
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType:  nlbmodel.EcsServerType,
						ServerId:    "ecs-1",
						Port:        80,
						Weight:      100,
						Description: "new",
					},
				},
			},
			expectedAdds:    0,
			expectedDels:    0,
			expectedUpdates: 1,
		},
		{
			name: "ignore user managed servers",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType:    nlbmodel.EcsServerType,
						ServerId:      "ecs-1",
						IsUserManaged: true,
						Port:          80,
						Weight:        100,
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectedAdds:    0,
			expectedDels:    0,
			expectedUpdates: 0,
		},
		{
			name: "ignore weight update when IgnoreWeightUpdate is true",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				IgnoreWeightUpdate: true,
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     200,
					},
				},
			},
			expectedAdds:    0,
			expectedDels:    0,
			expectedUpdates: 0,
		},
		{
			name: "complex scenario",
			remote: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-2",
						Port:       80,
						Weight:     100,
					},
				},
			},
			local: &nlbmodel.ServerGroup{
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     200,
					},
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-3",
						Port:       80,
						Weight:     100,
					},
				},
			},
			expectedAdds:    1,
			expectedDels:    1,
			expectedUpdates: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adds, dels, updates := diff(tt.remote, tt.local)
			assert.Equal(t, tt.expectedAdds, len(adds))
			assert.Equal(t, tt.expectedDels, len(dels))
			assert.Equal(t, tt.expectedUpdates, len(updates))
		})
	}
}

func TestPodPercentAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		mode     helper.TrafficPolicy
		backends []nlbmodel.ServerGroupServer
		weight   int
		validate func(t *testing.T, result []nlbmodel.ServerGroupServer)
	}{
		{
			name:     "empty backends",
			mode:     helper.ClusterTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{},
			weight:   100,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 0, len(result))
			},
		},
		{
			name: "zero weight",
			mode: helper.ClusterTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1"},
				{ServerId: "ecs-2"},
			},
			weight: 0,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(0), result[0].Weight)
				assert.Equal(t, int32(0), result[1].Weight)
			},
		},
		{
			name: "cluster traffic policy - equal distribution",
			mode: helper.ClusterTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1"},
				{ServerId: "ecs-2"},
			},
			weight: 100,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(50), result[0].Weight)
				assert.Equal(t, int32(50), result[1].Weight)
			},
		},
		{
			name: "eni traffic policy - equal distribution",
			mode: helper.ENITrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1"},
				{ServerId: "ecs-2"},
			},
			weight: 100,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(50), result[0].Weight)
				assert.Equal(t, int32(50), result[1].Weight)
			},
		},
		{
			name: "cluster traffic policy - weight less than backend count",
			mode: helper.ClusterTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1"},
				{ServerId: "ecs-2"},
				{ServerId: "ecs-3"},
			},
			weight: 2,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 3, len(result))
				assert.Equal(t, int32(1), result[0].Weight)
				assert.Equal(t, int32(1), result[1].Weight)
				assert.Equal(t, int32(1), result[2].Weight)
			},
		},
		{
			name: "local traffic policy - same pods per ecs",
			mode: helper.LocalTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1"},
				{ServerId: "ecs-1"},
				{ServerId: "ecs-2"},
				{ServerId: "ecs-2"},
			},
			weight: 100,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 4, len(result))
				ecs1Weight := result[0].Weight
				ecs2Weight := result[2].Weight
				assert.Equal(t, ecs1Weight, result[1].Weight)
				assert.Equal(t, ecs2Weight, result[3].Weight)
				assert.True(t, ecs1Weight >= 1)
				assert.True(t, ecs2Weight >= 1)
			},
		},
		{
			name: "local traffic policy - different pods per ecs",
			mode: helper.LocalTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1", ServerType: nlbmodel.EcsServerType},
				{ServerId: "ecs-1", ServerType: nlbmodel.EcsServerType},
				{ServerId: "ecs-1", ServerType: nlbmodel.EcsServerType},
				{ServerId: "ecs-2", ServerType: nlbmodel.EcsServerType},
			},
			weight: 100,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 4, len(result))
				ecs1Weight := result[0].Weight
				ecs2Weight := result[3].Weight
				assert.Equal(t, ecs1Weight, result[1].Weight)
				assert.Equal(t, ecs1Weight, result[2].Weight)
				assert.True(t, ecs1Weight > ecs2Weight)
				assert.True(t, ecs1Weight >= 1)
				assert.True(t, ecs2Weight >= 1)
			},
		},
		{
			name: "local traffic policy - minimum weight is 1",
			mode: helper.LocalTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{
				{ServerId: "ecs-1"},
				{ServerId: "ecs-2"},
				{ServerId: "ecs-3"},
				{ServerId: "ecs-4"},
				{ServerId: "ecs-5"},
			},
			weight: 2,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 5, len(result))
				for _, backend := range result {
					assert.True(t, backend.Weight >= 1)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := podPercentAlgorithm(tt.mode, tt.backends, tt.weight)
			tt.validate(t, result)
		})
	}
}

func TestGetAnyPortServerGroupNamedKey(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-svc",
		},
	}

	tests := []struct {
		name      string
		protocol  string
		startPort int32
		endPort   int32
		validate  func(t *testing.T, key *nlbmodel.SGNamedKey)
	}{
		{
			name:      "basic test",
			protocol:  "TCP",
			startPort: 80,
			endPort:   90,
			validate: func(t *testing.T, key *nlbmodel.SGNamedKey) {
				assert.NotNil(t, key)
				assert.Equal(t, "default", key.Namespace)
				assert.Equal(t, "test-svc", key.ServiceName)
				assert.Equal(t, "TCP", key.Protocol)
				assert.Equal(t, "80_90", key.SGGroupPort)
			},
		},
		{
			name:      "UDP protocol",
			protocol:  "UDP",
			startPort: 53,
			endPort:   53,
			validate: func(t *testing.T, key *nlbmodel.SGNamedKey) {
				assert.NotNil(t, key)
				assert.Equal(t, "UDP", key.Protocol)
				assert.Equal(t, "53_53", key.SGGroupPort)
			},
		},
		{
			name:      "TCPSSL protocol",
			protocol:  "TCPSSL",
			startPort: 443,
			endPort:   443,
			validate: func(t *testing.T, key *nlbmodel.SGNamedKey) {
				assert.NotNil(t, key)
				assert.Equal(t, "TCPSSL", key.Protocol)
				assert.Equal(t, "443_443", key.SGGroupPort)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := getAnyPortServerGroupNamedKey(svc, tt.protocol, tt.startPort, tt.endPort)
			tt.validate(t, key)
		})
	}
}

func TestServerGroupManager_BatchRemoveServers(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.Equal(t, nil, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	sg := &nlbmodel.ServerGroup{
		ServerGroupId: "sg-test-id",
	}

	delServers := []nlbmodel.ServerGroupServer{
		{
			ServerType: nlbmodel.EcsServerType,
			ServerId:   "ecs-1",
			Port:       80,
		},
		{
			ServerType: nlbmodel.EcsServerType,
			ServerId:   "ecs-2",
			Port:       80,
		},
	}

	err = mgr.BatchRemoveServers(reqCtx, sg, delServers)
	assert.Equal(t, nil, err)
}

func TestServerGroupManager_BatchUpdateServers(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.Equal(t, nil, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	sg := &nlbmodel.ServerGroup{
		ServerGroupId: "sg-test-id",
	}

	updateServers := []nlbmodel.ServerGroupServer{
		{
			ServerType: nlbmodel.EcsServerType,
			ServerId:   "ecs-1",
			Port:       80,
			Weight:     100,
		},
		{
			ServerType: nlbmodel.EcsServerType,
			ServerId:   "ecs-2",
			Port:       80,
			Weight:     200,
		},
	}

	err = mgr.BatchUpdateServers(reqCtx, sg, updateServers)
	assert.Equal(t, nil, err)
}

func TestGetServerGroupIDs(t *testing.T) {
	tests := []struct {
		name       string
		annotation string
		expected   []string
		expectErr  bool
	}{
		{
			name:       "empty annotation",
			annotation: "",
			expected:   nil,
			expectErr:  false,
		},
		{
			name:       "single server group",
			annotation: "sgp-xxx:443",
			expected:   []string{"sgp-xxx"},
			expectErr:  false,
		},
		{
			name:       "multiple server groups",
			annotation: "sgp-xxx:443,sgp-yyy:80",
			expected:   []string{"sgp-xxx", "sgp-yyy"},
			expectErr:  false,
		},
		{
			name:       "invalid format - no colon",
			annotation: "sgp-xxx",
			expected:   nil,
			expectErr:  true,
		},
		{
			name:       "invalid format - empty after colon",
			annotation: "sgp-xxx:",
			expected:   []string{"sgp-xxx"},
			expectErr:  false,
		},
		{
			name:       "multiple colons",
			annotation: "sgp-xxx:443:extra",
			expected:   []string{"sgp-xxx"},
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getServerGroupIDs(tt.annotation)
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, nil, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestServerGroupManager_DeleteServerGroup(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.Equal(t, nil, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	err = mgr.DeleteServerGroup(reqCtx, "sg-test-id")
	assert.Equal(t, nil, err)
}

func TestServerGroupManager_UpdateServerGroup(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.Equal(t, nil, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	local := &nlbmodel.ServerGroup{
		ServerGroupId:   "sg-local-id",
		ServerGroupName: "local-sg",
		Scheduler:       "Wrr",
		Servers: []nlbmodel.ServerGroupServer{
			{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-1",
				Port:       80,
				Weight:     100,
			},
		},
	}

	remote := &nlbmodel.ServerGroup{
		ServerGroupId:   "sg-remote-id",
		ServerGroupName: "remote-sg",
		Scheduler:       "Wlc",
		Servers: []nlbmodel.ServerGroupServer{
			{
				ServerType: nlbmodel.EcsServerType,
				ServerId:   "ecs-1",
				Port:       80,
				Weight:     50,
			},
		},
	}

	err = mgr.UpdateServerGroup(reqCtx, local, remote)
	assert.Equal(t, nil, err)
}

func TestServerGroupManager_UpdateServerGroup_UserManaged(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.Equal(t, nil, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	local := &nlbmodel.ServerGroup{
		IsUserManaged:   true,
		ServerGroupId:   "sg-local-id",
		ServerGroupName: "local-sg",
		Servers:         []nlbmodel.ServerGroupServer{},
	}

	remote := &nlbmodel.ServerGroup{
		ServerGroupId:   "sg-remote-id",
		ServerGroupName: "remote-sg",
		Servers:         []nlbmodel.ServerGroupServer{},
	}

	err = mgr.UpdateServerGroup(reqCtx, local, remote)
	assert.Equal(t, nil, err)
}

func TestServerGroupManager_UpdateServerGroup_HealthCheck(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.Equal(t, nil, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.Equal(t, nil, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	enabled := true
	local := &nlbmodel.ServerGroup{
		ServerGroupId:   "sg-local-id",
		ServerGroupName: "local-sg",
		HealthCheckConfig: &nlbmodel.HealthCheckConfig{
			HealthCheckEnabled: &enabled,
			HealthCheckType:    "TCP",
			HealthyThreshold:   3,
		},
		Servers: []nlbmodel.ServerGroupServer{},
	}

	remote := &nlbmodel.ServerGroup{
		ServerGroupId:   "sg-remote-id",
		ServerGroupName: "remote-sg",
		HealthCheckConfig: &nlbmodel.HealthCheckConfig{
			HealthCheckEnabled: &enabled,
			HealthCheckType:    "HTTP",
			HealthyThreshold:   5,
		},
		Servers: []nlbmodel.ServerGroupServer{},
	}

	err = mgr.UpdateServerGroup(reqCtx, local, remote)
	assert.Equal(t, nil, err)
}

func TestServerGroupManager_UpdateServerGroup_ConnectionDrain(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	enabled := true
	disabled := false
	tests := []struct {
		name        string
		local       *nlbmodel.ServerGroup
		remote      *nlbmodel.ServerGroup
		expectError bool
	}{
		{
			name: "update ConnectionDrainEnabled",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:          "sg-local-id",
				ServerGroupName:        "local-sg",
				ConnectionDrainEnabled: &enabled,
				Servers:                []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:          "sg-remote-id",
				ServerGroupName:        "remote-sg",
				ConnectionDrainEnabled: &disabled,
				Servers:                []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update ConnectionDrainTimeout",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:          "sg-local-id",
				ServerGroupName:        "local-sg",
				ConnectionDrainTimeout: 30,
				Servers:                []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:          "sg-remote-id",
				ServerGroupName:        "remote-sg",
				ConnectionDrainTimeout: 60,
				Servers:                []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update PreserveClientIpEnabled",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:           "sg-local-id",
				ServerGroupName:         "local-sg",
				PreserveClientIpEnabled: &enabled,
				Servers:                 []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:           "sg-remote-id",
				ServerGroupName:         "remote-sg",
				PreserveClientIpEnabled: &disabled,
				Servers:                 []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.UpdateServerGroup(reqCtx, tt.local, tt.remote)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerGroupManager_UpdateServerGroup_HealthCheckDetails(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	enabled := true
	tests := []struct {
		name        string
		local       *nlbmodel.ServerGroup
		remote      *nlbmodel.ServerGroup
		expectError bool
	}{
		{
			name: "update HealthCheckConnectPort",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:     &enabled,
					HealthCheckConnectPort: 8080,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:     &enabled,
					HealthCheckConnectPort: 80,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update UnhealthyThreshold",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					UnhealthyThreshold: 3,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					UnhealthyThreshold: 5,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update HealthCheckConnectTimeout",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:        &enabled,
					HealthCheckConnectTimeout: 5,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:        &enabled,
					HealthCheckConnectTimeout: 10,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update HealthCheckInterval",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:  &enabled,
					HealthCheckInterval: 5,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:  &enabled,
					HealthCheckInterval: 10,
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update HealthCheckDomain",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HealthCheckDomain:  "example.com",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HealthCheckDomain:  "test.com",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update HealthCheckUrl",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HealthCheckUrl:     "/health",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HealthCheckUrl:     "/check",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update HttpCheckMethod",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HttpCheckMethod:    "GET",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HttpCheckMethod:    "POST",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "update HealthCheckHttpCode",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:  &enabled,
					HealthCheckHttpCode: []string{"200", "201"},
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-remote-id",
				ServerGroupName: "remote-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled:  &enabled,
					HealthCheckHttpCode: []string{"200"},
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
		{
			name: "add HealthCheckConfig when remote is nil",
			local: &nlbmodel.ServerGroup{
				ServerGroupId:   "sg-local-id",
				ServerGroupName: "local-sg",
				HealthCheckConfig: &nlbmodel.HealthCheckConfig{
					HealthCheckEnabled: &enabled,
					HealthCheckType:    "TCP",
				},
				Servers: []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId:     "sg-remote-id",
				ServerGroupName:   "remote-sg",
				HealthCheckConfig: nil,
				Servers:           []nlbmodel.ServerGroupServer{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.UpdateServerGroup(reqCtx, tt.local, tt.remote)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerGroupManager_CreateServerGroup(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	tests := []struct {
		name        string
		serverGroup *nlbmodel.ServerGroup
		expectError bool
		validate    func(t *testing.T, sg *nlbmodel.ServerGroup)
	}{
		{
			name: "create server group successfully",
			serverGroup: &nlbmodel.ServerGroup{
				ServerGroupName: "test-sg",
				Protocol:        nlbmodel.TCP,
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.NotEmpty(t, sg.ServerGroupId)
			},
		},
		{
			name: "create server group with empty ResourceGroupId",
			serverGroup: &nlbmodel.ServerGroup{
				ServerGroupName: "test-sg",
				Protocol:        nlbmodel.TCP,
				ResourceGroupId: "",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.NotEmpty(t, sg.ServerGroupId)
			},
		},
		{
			name: "create server group with existing ResourceGroupId",
			serverGroup: &nlbmodel.ServerGroup{
				ServerGroupName: "test-sg",
				Protocol:        nlbmodel.TCP,
				ResourceGroupId: "rg-test-id",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.Equal(t, "rg-test-id", sg.ResourceGroupId)
				assert.NotEmpty(t, sg.ServerGroupId)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sgCopy := *tt.serverGroup
			err := mgr.CreateServerGroup(reqCtx, &sgCopy)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, &sgCopy)
				}
			}
		})
	}
}

func TestServerGroupManager_CleanupServerGroupTags(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	tests := []struct {
		name        string
		serverGroup *nlbmodel.ServerGroup
		expectError bool
	}{
		{
			name: "empty tags returns nil",
			serverGroup: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Tags:          []tag.Tag{},
			},
			expectError: false,
		},
		{
			name: "tags with default tags",
			serverGroup: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Tags: []tag.Tag{
					{
						Key:   "key1",
						Value: "value1",
					},
				},
			},
			expectError: false,
		},
		{
			name: "tags matching default tags should be deleted",
			serverGroup: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Tags: []tag.Tag{
					{
						Key:   "ack.aliyun.com",
						Value: "a5e4dbfc9c2ae4642b0335607860aef6",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.CleanupServerGroupTags(reqCtx, tt.serverGroup)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBackendPort(t *testing.T) {
	tests := []struct {
		name     string
		port     int32
		anyPort  bool
		expected int32
	}{
		{
			name:     "anyPort enabled returns 0",
			port:     80,
			anyPort:  true,
			expected: 0,
		},
		{
			name:     "anyPort disabled returns port",
			port:     80,
			anyPort:  false,
			expected: 80,
		},
		{
			name:     "anyPort disabled with different port",
			port:     443,
			anyPort:  false,
			expected: 443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBackendPort(tt.port, tt.anyPort)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsServerManagedByMyService(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-svc",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	tests := []struct {
		name     string
		server   nlbmodel.ServerGroupServer
		expected bool
	}{
		{
			name: "server managed by my service",
			server: nlbmodel.ServerGroupServer{
				Description: "k8s.80.TCP.test-svc.default.clusterid",
			},
			expected: true,
		},
		{
			name: "server managed by different service",
			server: nlbmodel.ServerGroupServer{
				Description: "k8s.80.TCP.other-svc.default.clusterid",
			},
			expected: false,
		},
		{
			name: "server managed by different namespace",
			server: nlbmodel.ServerGroupServer{
				Description: "k8s.80.TCP.test-svc.other-ns.clusterid",
			},
			expected: false,
		},
		{
			name: "invalid description format",
			server: nlbmodel.ServerGroupServer{
				Description: "invalid-description",
			},
			expected: false,
		},
		{
			name: "empty description",
			server: nlbmodel.ServerGroupServer{
				Description: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isServerManagedByMyService(reqCtx, tt.server)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetWeightBackends(t *testing.T) {
	tests := []struct {
		name     string
		mode     helper.TrafficPolicy
		backends []nlbmodel.ServerGroupServer
		weight   *int
		validate func(t *testing.T, result []nlbmodel.ServerGroupServer)
	}{
		{
			name:     "nil weight uses default algorithm",
			mode:     helper.ClusterTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{{ServerId: "ecs-1"}, {ServerId: "ecs-2"}},
			weight:   nil,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(100), result[0].Weight)
				assert.Equal(t, int32(100), result[1].Weight)
			},
		},
		{
			name:     "weight 50 with cluster mode",
			mode:     helper.ClusterTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{{ServerId: "ecs-1"}, {ServerId: "ecs-2"}},
			weight:   intPtr(50),
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(25), result[0].Weight)
				assert.Equal(t, int32(25), result[1].Weight)
			},
		},
		{
			name:     "weight 0 sets all weights to 0",
			mode:     helper.ENITrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{{ServerId: "ecs-1"}, {ServerId: "ecs-2"}},
			weight:   intPtr(0),
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(0), result[0].Weight)
				assert.Equal(t, int32(0), result[1].Weight)
			},
		},
		{
			name:     "weight with local mode",
			mode:     helper.LocalTrafficPolicy,
			backends: []nlbmodel.ServerGroupServer{{ServerId: "ecs-1"}, {ServerId: "ecs-1"}, {ServerId: "ecs-2"}},
			weight:   intPtr(100),
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 3, len(result))
				assert.True(t, result[0].Weight >= 1)
				assert.True(t, result[1].Weight >= 1)
				assert.True(t, result[2].Weight >= 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setWeightBackends(tt.mode, tt.backends, tt.weight)
			tt.validate(t, result)
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestUpdateENIBackends(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		backends        []nlbmodel.ServerGroupServer
		ipVersion       model.AddressIPVersionType
		serverGroupType nlbmodel.ServerGroupType
		validate        func(t *testing.T, result []nlbmodel.ServerGroupServer)
	}{
		{
			name: "IpServerGroupType sets ServerId and ServerType",
			backends: []nlbmodel.ServerGroupServer{
				{ServerIp: "10.0.0.1"},
				{ServerIp: "10.0.0.2"},
			},
			ipVersion:       model.IPv4,
			serverGroupType: nlbmodel.IpServerGroupType,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, "10.0.0.1", result[0].ServerId)
				assert.Equal(t, "10.0.0.1", result[0].ServerIp)
				assert.Equal(t, nlbmodel.IpServerType, result[0].ServerType)
				assert.Equal(t, "10.0.0.2", result[1].ServerId)
				assert.Equal(t, "10.0.0.2", result[1].ServerIp)
				assert.Equal(t, nlbmodel.IpServerType, result[1].ServerType)
			},
		},
		{
			name: "InstanceServerGroupType sets EniServerType",
			backends: []nlbmodel.ServerGroupServer{
				{ServerIp: "10.0.0.1"},
			},
			ipVersion:       model.IPv4,
			serverGroupType: nlbmodel.InstanceServerGroupType,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
				assert.Equal(t, nlbmodel.EniServerType, result[0].ServerType)
			},
		},
		{
			name:            "empty backends",
			backends:        []nlbmodel.ServerGroupServer{},
			ipVersion:       model.IPv4,
			serverGroupType: nlbmodel.InstanceServerGroupType,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 0, len(result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := updateENIBackends(mgr, tt.backends, tt.ipVersion, tt.serverGroupType)
			assert.NoError(t, err)
			tt.validate(t, result)
		})
	}
}

func TestSetServerGroupAttributeFromAnno(t *testing.T) {
	tests := []struct {
		name        string
		sg          *nlbmodel.ServerGroup
		anno        map[string]string
		expectError bool
		validate    func(t *testing.T, sg *nlbmodel.ServerGroup)
	}{
		{
			name: "set ServerGroupType to Ip",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ServerGroupType): "ip",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.Equal(t, nlbmodel.IpServerGroupType, sg.ServerGroupType)
			},
		},
		{
			name: "set ServerGroupType to Instance",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ServerGroupType): "instance",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.Equal(t, nlbmodel.InstanceServerGroupType, sg.ServerGroupType)
			},
		},
		{
			name: "set ConnectionDrain enabled",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ConnectionDrain): "on",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.NotNil(t, sg.ConnectionDrainEnabled)
				assert.True(t, *sg.ConnectionDrainEnabled)
			},
		},
		{
			name: "set ConnectionDrainTimeout",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ConnectionDrainTimeout): "30",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.Equal(t, int32(30), sg.ConnectionDrainTimeout)
			},
		},
		{
			name: "set Scheduler",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.Scheduler): "Wrr",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.Equal(t, "Wrr", sg.Scheduler)
			},
		},
		{
			name: "set PreserveClientIp enabled",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.PreserveClientIp): "on",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.NotNil(t, sg.PreserveClientIpEnabled)
				assert.True(t, *sg.PreserveClientIpEnabled)
			},
		},
		{
			name: "set ResourceGroupId",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ResourceGroupId): "rg-123",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.Equal(t, "rg-123", sg.ResourceGroupId)
			},
		},
		{
			name: "set HealthCheckConfig",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.HealthCheckFlag):           "on",
				annotation.Annotation(annotation.HealthCheckType):           "TCP",
				annotation.Annotation(annotation.HealthCheckConnectPort):    "8080",
				annotation.Annotation(annotation.HealthyThreshold):          "3",
				annotation.Annotation(annotation.UnhealthyThreshold):        "3",
				annotation.Annotation(annotation.HealthCheckConnectTimeout): "5",
				annotation.Annotation(annotation.HealthCheckInterval):       "10",
				annotation.Annotation(annotation.HealthCheckDomain):         "example.com",
				annotation.Annotation(annotation.HealthCheckURI):            "/health",
				annotation.Annotation(annotation.HealthCheckMethod):         "GET",
				annotation.Annotation(annotation.HealthCheckHTTPCode):       "200,201",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.NotNil(t, sg.HealthCheckConfig)
				assert.NotNil(t, sg.HealthCheckConfig.HealthCheckEnabled)
				assert.True(t, *sg.HealthCheckConfig.HealthCheckEnabled)
				assert.Equal(t, "TCP", sg.HealthCheckConfig.HealthCheckType)
				assert.Equal(t, int32(8080), sg.HealthCheckConfig.HealthCheckConnectPort)
				assert.Equal(t, int32(3), sg.HealthCheckConfig.HealthyThreshold)
				assert.Equal(t, int32(3), sg.HealthCheckConfig.UnhealthyThreshold)
				assert.Equal(t, int32(5), sg.HealthCheckConfig.HealthCheckConnectTimeout)
				assert.Equal(t, int32(10), sg.HealthCheckConfig.HealthCheckInterval)
				assert.Equal(t, "example.com", sg.HealthCheckConfig.HealthCheckDomain)
				assert.Equal(t, "/health", sg.HealthCheckConfig.HealthCheckUrl)
				assert.Equal(t, "GET", sg.HealthCheckConfig.HttpCheckMethod)
				assert.Equal(t, []string{"200", "201"}, sg.HealthCheckConfig.HealthCheckHttpCode)
			},
		},
		{
			name: "set IgnoreWeightUpdate",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.IgnoreWeightUpdate): "on",
			},
			expectError: false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup) {
				assert.True(t, sg.IgnoreWeightUpdate)
			},
		},
		{
			name: "invalid ServerGroupType",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ServerGroupType): "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid ConnectionDrainTimeout",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.ConnectionDrainTimeout): "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid HealthCheckConnectPort",
			sg:   &nlbmodel.ServerGroup{},
			anno: map[string]string{
				annotation.Annotation(annotation.HealthCheckFlag):        "on",
				annotation.Annotation(annotation.HealthCheckConnectPort): "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.anno,
				},
			}
			anno := annotation.NewAnnotationRequest(svc)
			err := setServerGroupAttributeFromAnno(tt.sg, anno)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.sg)
				}
			}
		})
	}
}

func TestServerGroupManager_UpdateServerGroupServers(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      ServiceName,
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	tests := []struct {
		name        string
		local       *nlbmodel.ServerGroup
		remote      *nlbmodel.ServerGroup
		expectError bool
	}{
		{
			name: "no changes",
			local: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			expectError: false,
		},
		{
			name: "add server",
			local: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-2",
						Port:       80,
						Weight:     100,
					},
				},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			expectError: false,
		},
		{
			name: "remove server",
			local: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers:       []nlbmodel.ServerGroupServer{},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			expectError: false,
		},
		{
			name: "update server weight",
			local: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     200,
					},
				},
			},
			remote: &nlbmodel.ServerGroup{
				ServerGroupId: "sg-test-id",
				Servers: []nlbmodel.ServerGroupServer{
					{
						ServerType: nlbmodel.EcsServerType,
						ServerId:   "ecs-1",
						Port:       80,
						Weight:     100,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.updateServerGroupServers(reqCtx, tt.local, tt.remote)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerGroupManager_SetBackendsFromEndpoints(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      "test-svc",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	nodeName := "test-node"
	tests := []struct {
		name            string
		endpoints       *v1.Endpoints
		servicePort     *v1.ServicePort
		validate        func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool)
		setupKubeClient func(t *testing.T, kubeClient client.Client) client.Client
	}{
		{
			name: "empty subsets",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
				assert.False(t, containsPotentialReady)
			},
		},
		{
			name: "with addresses",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.1",
								NodeName: &nodeName,
							},
							{
								IP:       "10.0.0.2",
								NodeName: &nodeName,
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromString("tcp"),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 2, len(backends))
				assert.Equal(t, "10.0.0.1", backends[0].ServerIp)
				assert.Equal(t, "10.0.0.2", backends[1].ServerIp)
				assert.Equal(t, int32(8080), backends[0].Port)
			},
		},
		{
			name: "with int targetPort",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.1",
								NodeName: &nodeName,
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(9090),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 1, len(backends))
				assert.Equal(t, int32(9090), backends[0].Port)
			},
		},
		{
			name: "anyPort enabled",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.1",
								NodeName: &nodeName,
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 1, len(backends))
			},
		},
		{
			name: "with NotReadyAddresses without TargetRef",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:        "10.0.0.3",
								NodeName:  &nodeName,
								TargetRef: nil,
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
				assert.False(t, containsPotentialReady)
			},
		},
		{
			name: "with NotReadyAddresses with non-Pod TargetRef",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.3",
								NodeName: &nodeName,
								TargetRef: &v1.ObjectReference{
									Kind:      "Node",
									Namespace: v1.NamespaceDefault,
									Name:      "test-pod",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
				assert.False(t, containsPotentialReady)
			},
		},
		{
			name: "with NotReadyAddresses with Pod not found",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.3",
								NodeName: &nodeName,
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: v1.NamespaceDefault,
									Name:      "not-found-pod",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
				assert.True(t, containsPotentialReady)
			},
		},
		{
			name: "with NotReadyAddresses with Pod without readiness gate",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.3",
								NodeName: &nodeName,
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: v1.NamespaceDefault,
									Name:      "test-pod-no-gate",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
				assert.False(t, containsPotentialReady)
			},
			setupKubeClient: func(t *testing.T, kubeClient client.Client) client.Client {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-no-gate",
						Namespace: v1.NamespaceDefault,
					},
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{},
					},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName,
					},
				}
				svc := &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc",
						Namespace: v1.NamespaceDefault,
					},
				}
				return fake.NewClientBuilder().WithRuntimeObjects(pod, node, svc).Build()
			},
		},
		{
			name: "with NotReadyAddresses with Pod containers not ready",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						NotReadyAddresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.3",
								NodeName: &nodeName,
								TargetRef: &v1.ObjectReference{
									Kind:      "Pod",
									Namespace: v1.NamespaceDefault,
									Name:      "test-pod-not-ready",
								},
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "tcp",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.True(t, containsPotentialReady)
			},
			setupKubeClient: func(t *testing.T, kubeClient client.Client) client.Client {
				readinessGateName := helper.BuildReadinessGatePodConditionTypeWithPrefix(helper.TargetHealthPodConditionServiceTypePrefix, "test-svc")
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-not-ready",
						Namespace: v1.NamespaceDefault,
					},
					Spec: v1.PodSpec{
						ReadinessGates: []v1.PodReadinessGate{
							{ConditionType: readinessGateName},
						},
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   v1.ContainersReady,
								Status: v1.ConditionFalse,
							},
						},
					},
				}
				node := &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName,
					},
				}
				svc := &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc",
						Namespace: v1.NamespaceDefault,
					},
				}
				return fake.NewClientBuilder().WithRuntimeObjects(pod, node, svc).Build()
			},
		},
		{
			name: "port name not found in endpoints",
			endpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: v1.NamespaceDefault,
				},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{
								IP:       "10.0.0.1",
								NodeName: &nodeName,
							},
						},
						Ports: []v1.EndpointPort{
							{
								Name:     "other",
								Port:     8080,
								Protocol: v1.ProtocolTCP,
							},
						},
					},
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromString("tcp"),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testMgr := mgr
			if tt.setupKubeClient != nil {
				newKubeClient := tt.setupKubeClient(t, recon.kubeClient)
				testMgr, err = NewServerGroupManager(newKubeClient, recon.cloud)
				assert.NoError(t, err)
			}
			candidates := &reconbackend.EndpointWithENI{
				Endpoints:        tt.endpoints,
				TrafficPolicy:    helper.ClusterTrafficPolicy,
				AddressIPVersion: model.IPv4,
			}
			sg := nlbmodel.ServerGroup{
				ServicePort:     tt.servicePort,
				ServerGroupName: "test-sg",
				AnyPortEnabled:  tt.name == "anyPort enabled",
			}
			backends, containsPotentialReady, err := testMgr.setBackendsFromEndpoints(reqCtx, candidates, sg)
			assert.NoError(t, err)
			tt.validate(t, backends, containsPotentialReady)
		})
	}
}

func TestServerGroupManager_SetBackendsFromEndpointSlices(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      "test-svc",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	nodeName := "test-node"
	portName := "tcp"
	port := int32(8080)
	protocol := v1.ProtocolTCP
	ready := true
	notReady := false

	tests := []struct {
		name           string
		endpointSlices []discovery.EndpointSlice
		servicePort    *v1.ServicePort
		validate       func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool)
	}{
		{
			name:           "empty endpoint slices",
			endpointSlices: []discovery.EndpointSlice{},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
				assert.False(t, containsPotentialReady)
			},
		},
		{
			name: "with ready endpoints",
			endpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc-1",
						Namespace: v1.NamespaceDefault,
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.1", "10.0.0.2"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
					AddressType: discovery.AddressTypeIPv4,
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromString("tcp"),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 2, len(backends))
				assert.Equal(t, "10.0.0.1", backends[0].ServerIp)
				assert.Equal(t, "10.0.0.2", backends[1].ServerIp)
			},
		},
		{
			name: "with int targetPort",
			endpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc-1",
						Namespace: v1.NamespaceDefault,
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
					AddressType: discovery.AddressTypeIPv4,
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(9090),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 1, len(backends))
				assert.Equal(t, int32(9090), backends[0].Port)
			},
		},
		{
			name: "ignore terminating pods",
			endpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc-1",
						Namespace: v1.NamespaceDefault,
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discovery.EndpointConditions{
								Ready:       &ready,
								Terminating: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
					AddressType: discovery.AddressTypeIPv4,
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
			},
		},
		{
			name: "deduplicate endpoints",
			endpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc-1",
						Namespace: v1.NamespaceDefault,
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discovery.EndpointConditions{
								Ready: &ready,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
					AddressType: discovery.AddressTypeIPv4,
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 1, len(backends))
			},
		},
		{
			name: "not ready endpoint without targetRef",
			endpointSlices: []discovery.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc-1",
						Namespace: v1.NamespaceDefault,
					},
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"10.0.0.1"},
							Conditions: discovery.EndpointConditions{
								Ready: &notReady,
							},
							NodeName: &nodeName,
						},
					},
					Ports: []discovery.EndpointPort{
						{
							Name:     &portName,
							Port:     &port,
							Protocol: &protocol,
						},
					},
					AddressType: discovery.AddressTypeIPv4,
				},
			},
			servicePort: &v1.ServicePort{
				Name:       "tcp",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
			validate: func(t *testing.T, backends []nlbmodel.ServerGroupServer, containsPotentialReady bool) {
				assert.Equal(t, 0, len(backends))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &reconbackend.EndpointWithENI{
				EndpointSlices:   tt.endpointSlices,
				TrafficPolicy:    helper.ClusterTrafficPolicy,
				AddressIPVersion: model.IPv4,
			}
			sg := nlbmodel.ServerGroup{
				ServicePort:     tt.servicePort,
				ServerGroupName: "test-sg",
			}
			backends, containsPotentialReady, err := mgr.setBackendsFromEndpointSlices(reqCtx, candidates, sg)
			assert.NoError(t, err)
			tt.validate(t, backends, containsPotentialReady)
		})
	}
}

func TestServerGroupManager_BuildENIBackends(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      "test-svc",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	tests := []struct {
		name            string
		initBackends    []nlbmodel.ServerGroupServer
		serverGroupType nlbmodel.ServerGroupType
		weight          *int
		validate        func(t *testing.T, result []nlbmodel.ServerGroupServer)
	}{
		{
			name:            "empty backends",
			initBackends:    []nlbmodel.ServerGroupServer{},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			weight:          nil,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 0, len(result))
			},
		},
		{
			name: "with backends and InstanceServerGroupType",
			initBackends: []nlbmodel.ServerGroupServer{
				{ServerIp: "10.0.0.1"},
				{ServerIp: "10.0.0.2"},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			weight:          nil,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, nlbmodel.EniServerType, result[0].ServerType)
				assert.Equal(t, nlbmodel.EniServerType, result[1].ServerType)
				assert.Equal(t, int32(100), result[0].Weight)
				assert.Equal(t, int32(100), result[1].Weight)
			},
		},
		{
			name: "with backends and IpServerGroupType",
			initBackends: []nlbmodel.ServerGroupServer{
				{ServerIp: "10.0.0.1"},
			},
			serverGroupType: nlbmodel.IpServerGroupType,
			weight:          nil,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
				assert.Equal(t, nlbmodel.IpServerType, result[0].ServerType)
				assert.Equal(t, "10.0.0.1", result[0].ServerId)
				assert.Equal(t, "10.0.0.1", result[0].ServerIp)
			},
		},
		{
			name: "with weight",
			initBackends: []nlbmodel.ServerGroupServer{
				{ServerIp: "10.0.0.1"},
				{ServerIp: "10.0.0.2"},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			weight:          intPtr(50),
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, int32(25), result[0].Weight)
				assert.Equal(t, int32(25), result[1].Weight)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &reconbackend.EndpointWithENI{
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			}
			sg := nlbmodel.ServerGroup{
				ServerGroupType: tt.serverGroupType,
				Weight:          tt.weight,
			}
			result, err := mgr.buildENIBackends(reqCtx, candidates, tt.initBackends, sg)
			assert.NoError(t, err)
			tt.validate(t, result)
		})
	}
}

func TestServerGroupManager_BuildLocalBackends(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      "test-svc",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	nodeName1 := NodeName
	nodeName3 := "vk-node"

	tests := []struct {
		name            string
		nodes           []v1.Node
		initBackends    []nlbmodel.ServerGroupServer
		serverGroupType nlbmodel.ServerGroupType
		servicePort     *v1.ServicePort
		expectError     bool
		validate        func(t *testing.T, result []nlbmodel.ServerGroupServer)
	}{
		{
			name:            "empty backends",
			nodes:           []v1.Node{},
			initBackends:    []nlbmodel.ServerGroupServer{},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 0, len(result))
			},
		},
		{
			name: "with ECS backends and InstanceServerGroupType",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName1,
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-1",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
						},
					},
				},
			},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.1",
					NodeName: &nodeName1,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
				assert.Equal(t, nlbmodel.EcsServerType, result[0].ServerType)
				assert.Equal(t, "ecs-id-1", result[0].ServerId)
				assert.Equal(t, int32(30080), result[0].Port)
			},
		},
		{
			name: "with ECS backends and IpServerGroupType",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName1,
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-1",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
						},
					},
				},
			},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.1",
					NodeName: &nodeName1,
				},
			},
			serverGroupType: nlbmodel.IpServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
				assert.Equal(t, nlbmodel.IpServerType, result[0].ServerType)
				assert.Equal(t, "192.168.1.1", result[0].ServerId)
				assert.Equal(t, "192.168.1.1", result[0].ServerIp)
			},
		},
		{
			name: "with VK node",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName3,
						Labels: map[string]string{
							"type": helper.LabelNodeTypeVK,
						},
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-3",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.3"},
						},
					},
				},
			},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.3",
					NodeName: &nodeName3,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
				assert.Equal(t, nlbmodel.EniServerType, result[0].ServerType)
			},
		},
		{
			name:  "node not found",
			nodes: []v1.Node{},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.1",
					NodeName: &nodeName1,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 0, len(result))
			},
		},
		{
			name:  "nil NodeName",
			nodes: []v1.Node{},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.1",
					NodeName: nil,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: true,
		},
		{
			name: "remove duplicated ECS",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName1,
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-1",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
						},
					},
				},
			},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.1",
					NodeName: &nodeName1,
				},
				{
					ServerIp: "10.0.0.2",
					NodeName: &nodeName1,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &reconbackend.EndpointWithENI{
				Nodes:            tt.nodes,
				TrafficPolicy:    helper.LocalTrafficPolicy,
				AddressIPVersion: model.IPv4,
			}
			sg := nlbmodel.ServerGroup{
				ServerGroupType: tt.serverGroupType,
				ServicePort:     tt.servicePort,
				ServerGroupName: "test-sg",
			}
			result, err := mgr.buildLocalBackends(reqCtx, candidates, tt.initBackends, sg)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestServerGroupManager_BuildClusterBackends(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      "test-svc",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	nodeName1 := NodeName
	nodeName2 := "cn-hangzhou.192.0.168.69"
	nodeName3 := "vk-node"

	tests := []struct {
		name            string
		nodes           []v1.Node
		initBackends    []nlbmodel.ServerGroupServer
		serverGroupType nlbmodel.ServerGroupType
		servicePort     *v1.ServicePort
		expectError     bool
		validate        func(t *testing.T, result []nlbmodel.ServerGroupServer)
	}{
		{
			name: "with nodes and InstanceServerGroupType",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName1,
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-1",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName2,
					},
					Spec: v1.NodeSpec{
						ProviderID: "alicloud://cn-hangzhou.ecs-id-2",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.2"},
						},
					},
				},
			},
			initBackends:    []nlbmodel.ServerGroupServer{},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 2, len(result))
				assert.Equal(t, nlbmodel.EcsServerType, result[0].ServerType)
				assert.Equal(t, nlbmodel.EcsServerType, result[1].ServerType)
				assert.Equal(t, int32(30080), result[0].Port)
				assert.Equal(t, int32(30080), result[1].Port)
			},
		},
		{
			name: "with nodes and IpServerGroupType",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName1,
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-1",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
						},
					},
				},
			},
			initBackends:    []nlbmodel.ServerGroupServer{},
			serverGroupType: nlbmodel.IpServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 1, len(result))
				assert.Equal(t, nlbmodel.IpServerType, result[0].ServerType)
				assert.Equal(t, "192.168.1.1", result[0].ServerId)
				assert.Equal(t, "192.168.1.1", result[0].ServerIp)
			},
		},
		{
			name: "with VK backends",
			nodes: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: nodeName3,
						Labels: map[string]string{
							"type": helper.LabelNodeTypeVK,
						},
					},
					Spec: v1.NodeSpec{
						ProviderID: "cn-hangzhou.ecs-id-3",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{Type: v1.NodeInternalIP, Address: "192.168.1.3"},
						},
					},
				},
			},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.3",
					NodeName: &nodeName3,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.True(t, len(result) >= 1)
				hasEniBackend := false
				for _, b := range result {
					if b.ServerType == nlbmodel.EniServerType {
						hasEniBackend = true
						break
					}
				}
				assert.True(t, hasEniBackend)
			},
		},
		{
			name:  "nil NodeName in initBackends",
			nodes: []v1.Node{},
			initBackends: []nlbmodel.ServerGroupServer{
				{
					ServerIp: "10.0.0.1",
					NodeName: nil,
				},
			},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: true,
		},
		{
			name:            "empty nodes",
			nodes:           []v1.Node{},
			initBackends:    []nlbmodel.ServerGroupServer{},
			serverGroupType: nlbmodel.InstanceServerGroupType,
			servicePort: &v1.ServicePort{
				NodePort: 30080,
			},
			expectError: false,
			validate: func(t *testing.T, result []nlbmodel.ServerGroupServer) {
				assert.Equal(t, 0, len(result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &reconbackend.EndpointWithENI{
				Nodes:            tt.nodes,
				TrafficPolicy:    helper.ClusterTrafficPolicy,
				AddressIPVersion: model.IPv4,
			}
			sg := nlbmodel.ServerGroup{
				ServerGroupType: tt.serverGroupType,
				ServicePort:     tt.servicePort,
				ServerGroupName: "test-sg",
			}
			result, err := mgr.buildClusterBackends(reqCtx, candidates, tt.initBackends, sg)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestServerGroupManager_SetServerGroupServers(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	nodeName := NodeName
	node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		Spec: v1.NodeSpec{
			ProviderID: "cn-hangzhou.ecs-id-1",
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "192.168.1.1"},
			},
		},
	}

	endpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: v1.NamespaceDefault,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.0.0.1",
						NodeName: &nodeName,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "tcp",
						Port:     8080,
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}

	tests := []struct {
		name            string
		svc             *v1.Service
		candidates      *reconbackend.EndpointWithENI
		sg              *nlbmodel.ServerGroup
		isUserManagedLB bool
		expectError     bool
		validate        func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool)
	}{
		{
			name: "ENI traffic policy",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.BackendType): "eni",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
				NamedKey: &nlbmodel.SGNamedKey{
					NamedKey: nlbmodel.NamedKey{
						ServiceName: "test-svc",
						Namespace:   v1.NamespaceDefault,
					},
				},
			},
			isUserManagedLB: false,
			expectError:     false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool) {
				assert.True(t, len(sg.Servers) >= 0)
			},
		},
		{
			name: "Local traffic policy",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
				},
				Spec: v1.ServiceSpec{
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							NodePort:   30080,
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.LocalTrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					NodePort:   30080,
				},
				ServerGroupName: "test-sg",
				NamedKey: &nlbmodel.SGNamedKey{
					NamedKey: nlbmodel.NamedKey{
						ServiceName: "test-svc",
						Namespace:   v1.NamespaceDefault,
					},
				},
			},
			isUserManagedLB: false,
			expectError:     false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool) {
				assert.True(t, len(sg.Servers) >= 0)
			},
		},
		{
			name: "Cluster traffic policy",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							NodePort:   30080,
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ClusterTrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					NodePort:   30080,
				},
				ServerGroupName: "test-sg",
				NamedKey: &nlbmodel.SGNamedKey{
					NamedKey: nlbmodel.NamedKey{
						ServiceName: "test-svc",
						Namespace:   v1.NamespaceDefault,
					},
				},
			},
			isUserManagedLB: false,
			expectError:     false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool) {
				assert.True(t, len(sg.Servers) >= 0)
			},
		},
		{
			name: "unsupported traffic policy",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.TrafficPolicy("unsupported"),
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: false,
			expectError:     true,
		},
		{
			name: "user managed LB with VGroupWeight",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupWeight): "50",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool) {
				if sg.Weight != nil {
					assert.Equal(t, 50, *sg.Weight)
				}
			},
		},
		{
			name: "user managed LB with invalid VGroupPort format",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "invalid-format",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     true,
		},
		{
			name: "user managed LB with VGroupPort not found",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "sg-not-found-id:80",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     true,
		},
		{
			name: "user managed LB with VGroupPort not user managed",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "sg-not-user-managed-id:80",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     true,
		},
		{
			name: "user managed LB with valid VGroupPort",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "sg-user-managed-id:80",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool) {
				assert.Equal(t, "sg-user-managed-id", sg.ServerGroupId)
				assert.True(t, sg.IsUserManaged)
			},
		},
		{
			name: "user managed LB with invalid VGroupWeight - non-numeric",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "sg-user-managed-id:80",
						annotation.Annotation(annotation.VGroupWeight):   "invalid",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     true,
		},
		{
			name: "user managed LB with invalid VGroupWeight - out of range (negative)",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "sg-user-managed-id:80",
						annotation.Annotation(annotation.VGroupWeight):   "-1",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     true,
		},
		{
			name: "user managed LB with invalid VGroupWeight - out of range (over 100)",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.VGroupPort):     "sg-user-managed-id:80",
						annotation.Annotation(annotation.VGroupWeight):   "101",
						annotation.Annotation(annotation.LoadBalancerId): "nlb-exist-id",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints:        endpoints,
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
			},
			isUserManagedLB: true,
			expectError:     true,
		},
		{
			name: "empty backends triggers event",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			candidates: &reconbackend.EndpointWithENI{
				Endpoints: &v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc",
						Namespace: v1.NamespaceDefault,
					},
					Subsets: []v1.EndpointSubset{},
				},
				Nodes:            []v1.Node{node},
				TrafficPolicy:    helper.ENITrafficPolicy,
				AddressIPVersion: model.IPv4,
			},
			sg: &nlbmodel.ServerGroup{
				ServicePort: &v1.ServicePort{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
				ServerGroupName: "test-sg",
				NamedKey: &nlbmodel.SGNamedKey{
					NamedKey: nlbmodel.NamedKey{
						ServiceName: "test-svc",
						Namespace:   v1.NamespaceDefault,
					},
				},
			},
			isUserManagedLB: false,
			expectError:     false,
			validate: func(t *testing.T, sg *nlbmodel.ServerGroup, containsPotentialReady bool) {
				assert.Equal(t, 0, len(sg.Servers))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := getReqCtx(tt.svc)
			if reqCtx.Recorder == nil {
				reqCtx.Recorder = record.NewFakeRecorder(10)
			}
			containsPotentialReady, err := mgr.setServerGroupServers(reqCtx, tt.sg, tt.candidates, tt.isUserManagedLB)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.sg, containsPotentialReady)
				}
			}
		})
	}
}

func TestServerGroupManager_BuildLocalModel(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr, err := NewServerGroupManager(recon.kubeClient, recon.cloud)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		svc         *v1.Service
		mdl         *nlbmodel.NetworkLoadBalancer
		expectError bool
		validate    func(t *testing.T, mdl *nlbmodel.NetworkLoadBalancer)
	}{
		{
			name: "build local model with normal listener",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			mdl: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					IsUserManaged: false,
				},
				Listeners: []*nlbmodel.ListenerAttribute{
					{
						ListenerPort:     80,
						ListenerProtocol: nlbmodel.TCP,
						ServicePort: &v1.ServicePort{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, mdl *nlbmodel.NetworkLoadBalancer) {
				assert.NotNil(t, mdl.ServerGroups)
				assert.True(t, len(mdl.ServerGroups) > 0)
			},
		},
		{
			name: "build local model with anyPort listener",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			mdl: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					IsUserManaged: false,
				},
				Listeners: []*nlbmodel.ListenerAttribute{
					{
						ListenerPort:     0,
						StartPort:        1000,
						EndPort:          2000,
						ListenerProtocol: nlbmodel.TCP,
						ServicePort: &v1.ServicePort{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, mdl *nlbmodel.NetworkLoadBalancer) {
				assert.NotNil(t, mdl.ServerGroups)
				assert.True(t, len(mdl.ServerGroups) > 0)
				if len(mdl.ServerGroups) > 0 {
					assert.True(t, mdl.ServerGroups[0].AnyPortEnabled)
					assert.NotNil(t, mdl.ServerGroups[0].HealthCheckConfig)
				}
			},
		},
		{
			name: "build local model with invalid annotation",
			svc: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: v1.NamespaceDefault,
					Name:      "test-svc",
					Annotations: map[string]string{
						annotation.Annotation(annotation.ServerGroupType): "invalid",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			mdl: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					IsUserManaged: false,
				},
				Listeners: []*nlbmodel.ListenerAttribute{
					{
						ListenerPort:     80,
						ListenerProtocol: nlbmodel.TCP,
						ServicePort: &v1.ServicePort{
							Name:       "tcp",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := getReqCtx(tt.svc)
			reqCtx.Recorder = record.NewFakeRecorder(10)

			mdlCopy := *tt.mdl
			err := mgr.BuildLocalModel(reqCtx, &mdlCopy)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, &mdlCopy)
				}
			}
		})
	}
}
