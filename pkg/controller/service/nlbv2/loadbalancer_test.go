package nlbv2

import (
	"context"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	ecsmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/ecs"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

func TestBuildSecurityGroupPermissionsFromSourceRanges(t *testing.T) {
	config.CloudCFG.Global.ClusterID = "test-cluster-id"

	tests := []struct {
		name         string
		service      *v1.Service
		sourceRanges []string
		wantLen      int
		wantDesc     string
	}{
		{
			name: "single source range",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "default",
				},
			},
			sourceRanges: []string{"192.168.1.0/24"},
			wantLen:      2, // 1 accept + 1 drop
			wantDesc:     "k8s.default.test-svc.test-cluster-id",
		},
		{
			name: "multiple source ranges",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "kube-system",
				},
			},
			sourceRanges: []string{"192.168.1.0/24", "10.0.0.0/16", "172.16.0.0/12"},
			wantLen:      4, // 3 accept + 1 drop
			wantDesc:     "k8s.kube-system.test-svc.test-cluster-id",
		},
		{
			name: "empty source ranges",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "default",
				},
			},
			sourceRanges: []string{},
			wantLen:      1, // only drop policy
			wantDesc:     "k8s.default.test-svc.test-cluster-id",
		},
		{
			name: "service with different namespace",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-service",
					Namespace: "production",
				},
			},
			sourceRanges: []string{"0.0.0.0/0"},
			wantLen:      2,
			wantDesc:     "k8s.production.my-service.test-cluster-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &svcCtx.RequestContext{
				Service: tt.service,
				Anno:    annotation.NewAnnotationRequest(tt.service),
				Log:     util.NLBLog.WithValues("service", util.Key(tt.service)),
			}

			permissions := buildSecurityGroupPermissionsFromSourceRanges(reqCtx, tt.sourceRanges)

			assert.Equal(t, tt.wantLen, len(permissions))

			for _, perm := range permissions {
				assert.Equal(t, tt.wantDesc, perm.Description)
			}

			acceptCount := 0
			for i, perm := range permissions {
				if i < len(tt.sourceRanges) {
					assert.Equal(t, ecsmodel.SecurityGroupPolicyAccept, perm.Policy)
					assert.Equal(t, tt.sourceRanges[i], perm.SourceCidrIp)
					assert.Equal(t, "1", perm.Priority)
					acceptCount++
				}
			}
			assert.Equal(t, len(tt.sourceRanges), acceptCount)

			if len(permissions) > 0 {
				lastPerm := permissions[len(permissions)-1]
				assert.Equal(t, ecsmodel.SecurityGroupPolicyDrop, lastPerm.Policy)
				assert.Equal(t, "0.0.0.0/0", lastPerm.SourceCidrIp)
				assert.Equal(t, "100", lastPerm.Priority)
			}
		})
	}
}

func TestGetSecurityGroupRuleDescription(t *testing.T) {
	config.CloudCFG.Global.ClusterID = "test-cluster-id"

	tests := []struct {
		name     string
		service  *v1.Service
		wantDesc string
	}{
		{
			name: "default namespace",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc1",
					Namespace: "default",
				},
			},
			wantDesc: "k8s.default.svc1.test-cluster-id",
		},
		{
			name: "custom namespace",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-app",
					Namespace: "production",
				},
			},
			wantDesc: "k8s.production.my-app.test-cluster-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := getSecurityGroupRuleDescription(tt.service)
			assert.Equal(t, tt.wantDesc, desc)
		})
	}
}

func TestDiffAssociatedSecurityGroupPermissions(t *testing.T) {
	config.CloudCFG.Global.ClusterID = "test-cluster-id"

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
		},
	}
	managedDesc := "k8s.default.test-svc.test-cluster-id"

	tests := []struct {
		name            string
		local           []ecsmodel.SecurityGroupPermission
		remote          []ecsmodel.SecurityGroupPermission
		wantAddLen      int
		wantDeleteLen   int
		wantUpdateLen   int
		wantAddCidrs    []string
		wantDeleteCidrs []string
		wantUpdateCidrs []string
	}{
		{
			name:          "both local and remote are empty",
			local:         []ecsmodel.SecurityGroupPermission{},
			remote:        []ecsmodel.SecurityGroupPermission{},
			wantAddLen:    0,
			wantDeleteLen: 0,
			wantUpdateLen: 0,
		},
		{
			name:  "local is empty, remote has managed rules - should delete",
			local: []ecsmodel.SecurityGroupPermission{},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
			},
			wantAddLen:      0,
			wantDeleteLen:   1,
			wantUpdateLen:   0,
			wantDeleteCidrs: []string{"192.168.1.0/24"},
		},
		{
			name: "remote is empty, local has rules - should add",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
			},
			remote:        []ecsmodel.SecurityGroupPermission{},
			wantAddLen:    1,
			wantDeleteLen: 0,
			wantUpdateLen: 0,
			wantAddCidrs:  []string{"192.168.1.0/24"},
		},
		{
			name: "local and remote are identical - no changes",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
			},
			wantAddLen:    0,
			wantDeleteLen: 0,
			wantUpdateLen: 0,
		},
		{
			name: "need to add new rules",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
				{
					SourceCidrIp: "10.0.0.0/16",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
			},
			wantAddLen:    1,
			wantDeleteLen: 0,
			wantUpdateLen: 0,
			wantAddCidrs:  []string{"10.0.0.0/16"},
		},
		{
			name: "need to delete rules",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
				{
					SourceCidrIp:        "10.0.0.0/16",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-2",
				},
			},
			wantAddLen:      0,
			wantDeleteLen:   1,
			wantUpdateLen:   0,
			wantDeleteCidrs: []string{"10.0.0.0/16"},
		},
		{
			name: "need to update priority",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "2",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
			},
			wantAddLen:      0,
			wantDeleteLen:   0,
			wantUpdateLen:   1,
			wantUpdateCidrs: []string{"192.168.1.0/24"},
		},
		{
			name: "need to update description",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  "new-description",
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
			},
			wantAddLen:      0,
			wantDeleteLen:   0,
			wantUpdateLen:   1,
			wantUpdateCidrs: []string{"192.168.1.0/24"},
		},
		{
			name: "remote has unmanaged rule - should skip and not delete",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
				{
					SourceCidrIp:        "10.0.0.0/16",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         "user-managed-rule",
					SecurityGroupRuleId: "rule-2",
				},
			},
			wantAddLen:    0,
			wantDeleteLen: 0,
			wantUpdateLen: 0,
		},
		{
			name: "complex scenario with add, delete and update",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
				{
					SourceCidrIp: "10.0.0.0/16",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "2",
					Description:  managedDesc,
				},
				{
					SourceCidrIp: "0.0.0.0/0",
					Policy:       ecsmodel.SecurityGroupPolicyDrop,
					Priority:     "100",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
				{
					SourceCidrIp:        "172.16.0.0/12",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-2",
				},
				{
					SourceCidrIp:        "10.0.0.0/16",
					Policy:              ecsmodel.SecurityGroupPolicyAccept,
					Priority:            "1",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-3",
				},
			},
			wantAddLen:      1,
			wantDeleteLen:   1,
			wantUpdateLen:   1,
			wantAddCidrs:    []string{"0.0.0.0/0"},
			wantDeleteCidrs: []string{"172.16.0.0/12"},
			wantUpdateCidrs: []string{"10.0.0.0/16"},
		},
		{
			name: "different policies with same cidr are different rules",
			local: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp: "192.168.1.0/24",
					Policy:       ecsmodel.SecurityGroupPolicyAccept,
					Priority:     "1",
					Description:  managedDesc,
				},
			},
			remote: []ecsmodel.SecurityGroupPermission{
				{
					SourceCidrIp:        "192.168.1.0/24",
					Policy:              ecsmodel.SecurityGroupPolicyDrop,
					Priority:            "100",
					Description:         managedDesc,
					SecurityGroupRuleId: "rule-1",
				},
			},
			wantAddLen:      1,
			wantDeleteLen:   1,
			wantUpdateLen:   0,
			wantAddCidrs:    []string{"192.168.1.0/24"},
			wantDeleteCidrs: []string{"192.168.1.0/24"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &svcCtx.RequestContext{
				Service: svc,
				Anno:    annotation.NewAnnotationRequest(svc),
				Log:     util.NLBLog.WithValues("service", util.Key(svc)),
			}

			toAdd, toDelete, toUpdate := diffAssociatedSecurityGroupPermissions(reqCtx, tt.local, tt.remote)

			assert.Equal(t, tt.wantAddLen, len(toAdd))
			assert.Equal(t, tt.wantDeleteLen, len(toDelete))
			assert.Equal(t, tt.wantUpdateLen, len(toUpdate))

			if tt.wantAddCidrs != nil {
				var addCidrs []string
				for _, p := range toAdd {
					addCidrs = append(addCidrs, p.SourceCidrIp)
				}
				assert.ElementsMatch(t, tt.wantAddCidrs, addCidrs)
			}

			if tt.wantDeleteCidrs != nil {
				var deleteCidrs []string
				for _, p := range toDelete {
					deleteCidrs = append(deleteCidrs, p.SourceCidrIp)
				}
				assert.ElementsMatch(t, tt.wantDeleteCidrs, deleteCidrs)
			}

			if tt.wantUpdateCidrs != nil {
				var updateCidrs []string
				for _, p := range toUpdate {
					updateCidrs = append(updateCidrs, p.SourceCidrIp)
				}
				assert.ElementsMatch(t, tt.wantUpdateCidrs, updateCidrs)
			}

			// Verify that update results contain SecurityGroupRuleId from remote
			for _, p := range toUpdate {
				assert.NotEmpty(t, p.SecurityGroupRuleId)
			}
		})
	}
}

func TestNLBManager_SetProtectionsOff(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr := NewNLBManager(recon.cloud)

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
		name        string
		remote      *nlbmodel.NetworkLoadBalancer
		expectError bool
	}{
		{
			name: "empty load balancer id returns nil",
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId: "",
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "valid load balancer id calls UpdateLoadBalancerProtection",
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId: "nlb-test-id",
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.SetProtectionsOff(reqCtx, tt.remote)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNLBManager_CleanupLoadBalancerTags(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr := NewNLBManager(recon.cloud)

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
		name        string
		remote      *nlbmodel.NetworkLoadBalancer
		expectError bool
	}{
		{
			name: "empty load balancer id returns nil",
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId: "",
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "valid load balancer id with tags",
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId: "nlb-test-id",
					Tags: []tag.Tag{
						{
							Key:   "key1",
							Value: "value1",
						},
					},
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.CleanupLoadBalancerTags(reqCtx, tt.remote)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateBandwidthPackageId(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr := NewNLBManager(recon.cloud)

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
		name        string
		local       *nlbmodel.NetworkLoadBalancer
		remote      *nlbmodel.NetworkLoadBalancer
		expectError bool
	}{
		{
			name: "local not set",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: nil,
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-remote"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "local equals remote",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-1"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-1"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "detach and attach",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-2"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-1"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "attach only",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-3"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String(""),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "detach only",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String(""),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:     "nlb-test-id",
					BandwidthPackageId: tea.String("bwp-4"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateBandwidthPackageId(mgr, reqCtx, tt.local, tt.remote)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNLBManager_Update(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)

	mgr := NewNLBManager(recon.cloud)

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

	t.Run("address ip version changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:   "nlb-test-id",
				AddressIpVersion: "Ipv4",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:   "nlb-test-id",
				AddressIpVersion: "Ipv6",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.Error(t, err)
	})

	t.Run("resource group id changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:  "nlb-test-id",
				ResourceGroupId: "rg-1",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:  "nlb-test-id",
				ResourceGroupId: "rg-2",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.Error(t, err)
	})

	t.Run("address type changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				AddressType:    "Internet",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				AddressType:    "Intranet",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("zone mappings changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				ZoneMappings: []nlbmodel.ZoneMapping{
					{
						ZoneId:    "cn-hangzhou-a",
						VSwitchId: "vsw-1",
					},
				},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				ZoneMappings: []nlbmodel.ZoneMapping{
					{
						ZoneId:    "cn-hangzhou-b",
						VSwitchId: "vsw-2",
					},
				},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("security group ids changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:   "nlb-test-id",
				SecurityGroupIds: []string{"sg-1", "sg-2"},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:   "nlb-test-id",
				SecurityGroupIds: []string{"sg-1"},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("ipv6 address type changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:  "nlb-test-id",
				IPv6AddressType: "Internet",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:  "nlb-test-id",
				IPv6AddressType: "Intranet",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("name changed", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				Name:           "new-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				Name:           "old-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("no changes", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				Name:           "test-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				Name:           "test-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})
}

func TestGetSecurityGroupName(t *testing.T) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "demo",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	assert.Equal(t, "k8s-nlb-"+reqCtx.Anno.GetDefaultLoadBalancerName(), getSecurityGroupName(reqCtx))
}

func TestAssociatedSecurityGroupOperations(t *testing.T) {
	recon, err := getReconcileNLB()
	assert.NoError(t, err)
	mgr := NewNLBManager(recon.cloud)

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      "sg-op-svc",
		},
	}
	reqCtx := &svcCtx.RequestContext{
		Ctx:     context.TODO(),
		Service: svc,
		Anno:    annotation.NewAnnotationRequest(svc),
		Log:     util.NLBLog.WithValues("service", util.Key(svc)),
	}

	t.Run("find associated security group", func(t *testing.T) {
		_, err := mgr.FindAssociatedSecurityGroup(reqCtx)
		assert.Error(t, err)
	})

	t.Run("create associated security group", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				SourceRanges: []string{"192.168.1.0/24"},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.CreateAssociatedSecurityGroup(reqCtx, local)
		assert.NoError(t, err)
	})

	t.Run("update associated security group early return", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				SourceRangesSecurityGroupId: "",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
			NamespacedName:        util.NamespacedName(svc),
		}
		err := mgr.UpdateAssociatedSecurityGroup(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("delete associated security group early return", func(t *testing.T) {
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{},
			NamespacedName:        util.NamespacedName(svc),
		}
		err := mgr.DeleteAssociatedSecurityGroup(reqCtx, remote)
		assert.NoError(t, err)
	})

	t.Run("update associated security group rules", func(t *testing.T) {
		local := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				SourceRangesSecurityGroupId: "sg-test",
				SourceRanges:                []string{"192.168.1.0/24"},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				SourceRangesSecurityGroupId: "sg-test",
			},
			AssociatedSecurityGroup: &ecsmodel.SecurityGroup{
				ID:          "sg-test",
				Permissions: []ecsmodel.SecurityGroupPermission{},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.updateAssociatedSecurityGroupRules(reqCtx, local, remote)
		assert.NoError(t, err)
	})
}
