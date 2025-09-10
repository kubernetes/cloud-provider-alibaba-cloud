package nlbv2

import (
	"context"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"testing"
)

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
		name          string
		remote        *nlbmodel.NetworkLoadBalancer
		expectError   bool
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
		name          string
		remote        *nlbmodel.NetworkLoadBalancer
		expectError   bool
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
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   nil,
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-remote"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "local equals remote",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-1"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-1"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "detach and attach",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-2"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-1"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "attach only",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-3"),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String(""),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			expectError: false,
		},
		{
			name: "detach only",
			local: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String(""),
				},
				NamespacedName: util.NamespacedName(svc),
			},
			remote: &nlbmodel.NetworkLoadBalancer{
				LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
					LoadBalancerId:       "nlb-test-id",
					BandwidthPackageId:   tea.String("bwp-4"),
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
				LoadBalancerId:   "nlb-test-id",
				ResourceGroupId: "rg-1",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:   "nlb-test-id",
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
				AddressType:   "Internet",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				AddressType:   "Intranet",
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
				LoadBalancerId:    "nlb-test-id",
				SecurityGroupIds: []string{"sg-1", "sg-2"},
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:    "nlb-test-id",
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
				LoadBalancerId:   "nlb-test-id",
				IPv6AddressType: "Internet",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId:   "nlb-test-id",
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
				Name:          "new-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				Name:          "old-name",
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
				Name:          "test-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		remote := &nlbmodel.NetworkLoadBalancer{
			LoadBalancerAttribute: &nlbmodel.LoadBalancerAttribute{
				LoadBalancerId: "nlb-test-id",
				Name:          "test-name",
			},
			NamespacedName: util.NamespacedName(svc),
		}
		err := mgr.Update(reqCtx, local, remote)
		assert.NoError(t, err)
	})
}

