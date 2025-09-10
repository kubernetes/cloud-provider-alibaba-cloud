package clbv1

import (
	"context"
	"fmt"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/stretchr/testify/assert"
	ctrlcfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
)

type mockProvider struct {
	vmock.MockCloud
	vswIDs    []string
	vswitches map[string]vpc.VSwitch
}

func (m *mockProvider) VswitchIDs() ([]string, error) {
	if len(m.vswIDs) == 0 {
		return []string{}, nil
	}
	return m.vswIDs, nil
}

func (m *mockProvider) DescribeVswitchByID(ctx context.Context, vswId string) (vpc.VSwitch, error) {
	if vsw, ok := m.vswitches[vswId]; ok {
		return vsw, nil
	}
	return vpc.VSwitch{}, fmt.Errorf("vswitch %s not found", vswId)
}

func TestRandomVswitchID(t *testing.T) {
	cloud1 := &mockProvider{
		vswIDs: []string{},
	}
	mgr1 := NewLoadBalancerManager(cloud1)
	vswId1, err1 := mgr1.RandomVswitchID()
	assert.Error(t, err1)
	assert.Equal(t, "", vswId1)
	assert.Contains(t, err1.Error(), "no vswitches found")

	vswIDs := []string{"vsw-1", "vsw-2"}
	cloud2 := &mockProvider{
		vswIDs: vswIDs,
	}
	mgr2 := NewLoadBalancerManager(cloud2)
	vswId2, err2 := mgr2.RandomVswitchID()
	assert.NoError(t, err2)
	assert.Contains(t, vswIDs, vswId2)
}

func TestSetAvailableVswitchID(t *testing.T) {
	vswIDs := []string{"vsw-1", "vsw-2", "vsw-3"}
	vswitches := map[string]vpc.VSwitch{
		"vsw-1": {VSwitchId: "vsw-1", AvailableIpAddressCount: 0},
		"vsw-2": {VSwitchId: "vsw-2", AvailableIpAddressCount: 10},
		"vsw-3": {VSwitchId: "vsw-3", AvailableIpAddressCount: 20},
	}

	cloud := &mockProvider{
		vswIDs:    vswIDs,
		vswitches: vswitches,
	}
	mgr := NewLoadBalancerManager(cloud)

	reqCtx := &svcCtx.RequestContext{
		Ctx: context.TODO(),
		Log: util.ServiceLog,
	}

	local1 := &model.LoadBalancer{
		LoadBalancerAttribute: model.LoadBalancerAttribute{
			VSwitchId: "vsw-1",
		},
	}
	err1 := mgr.setAvailableVswitchID(reqCtx, local1)
	assert.NoError(t, err1)
	assert.Equal(t, "vsw-2", local1.LoadBalancerAttribute.VSwitchId)

	// Case 2: No available vswitch (all have 0 IP)
	vswitches2 := map[string]vpc.VSwitch{
		"vsw-1": {VSwitchId: "vsw-1", AvailableIpAddressCount: 0},
		"vsw-2": {VSwitchId: "vsw-2", AvailableIpAddressCount: 0},
	}
	cloud2 := &mockProvider{
		vswIDs:    []string{"vsw-1", "vsw-2"},
		vswitches: vswitches2,
	}
	mgr2 := NewLoadBalancerManager(cloud2)
	local2 := &model.LoadBalancer{
		LoadBalancerAttribute: model.LoadBalancerAttribute{
			VSwitchId: "vsw-1",
		},
	}
	err2 := mgr2.setAvailableVswitchID(reqCtx, local2)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "no available vsw found")
}

func TestEqualsAddressIPVersion(t *testing.T) {
	cases := []struct {
		name     string
		local    model.AddressIPVersionType
		remote   model.AddressIPVersionType
		expected bool
	}{
		{
			name:     "both empty",
			local:    "",
			remote:   "",
			expected: true,
		},
		{
			name:     "local empty, remote ipv4",
			local:    "",
			remote:   model.IPv4,
			expected: true,
		},
		{
			name:     "local ipv4, remote empty",
			local:    model.IPv4,
			remote:   "",
			expected: true,
		},
		{
			name:     "both ipv4",
			local:    model.IPv4,
			remote:   model.IPv4,
			expected: true,
		},
		{
			name:     "both ipv6",
			local:    model.IPv6,
			remote:   model.IPv6,
			expected: true,
		},
		{
			name:     "local ipv4, remote ipv6",
			local:    model.IPv4,
			remote:   model.IPv6,
			expected: false,
		},
		{
			name:     "local empty, remote ipv6",
			local:    "",
			remote:   model.IPv6,
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := equalsAddressIPVersion(c.local, c.remote)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestSetModelDefaultValue(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())

	t.Run("set default address type", func(t *testing.T) {
		mdl := &model.LoadBalancer{}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.Equal(t, model.InternetAddressType, mdl.LoadBalancerAttribute.AddressType)
	})

	t.Run("set default loadbalancer name", func(t *testing.T) {
		mdl := &model.LoadBalancer{}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.NotEmpty(t, mdl.LoadBalancerAttribute.LoadBalancerName)
	})

	t.Run("intranet with vswitch", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[annotation.Annotation(annotation.AddressType)] = string(model.IntranetAddressType)
		// Provide VSwitchId to avoid calling metadata VswitchID() which is unimplemented in mock
		svc.Annotations[annotation.Annotation(annotation.VswitchId)] = "vsw-test"
		reqCtx := getReqCtx(svc)
		mdl := &model.LoadBalancer{
			NamespacedName: util.NamespacedName(svc),
		}
		// Build local model first to populate fields from annotations
		err := mgr.BuildLocalModel(reqCtx, mdl)
		assert.NoError(t, err)
		// Then set default values
		err = setModelDefaultValue(mgr, mdl, reqCtx.Anno)
		assert.NoError(t, err)
		assert.NotEmpty(t, mdl.LoadBalancerAttribute.VpcId)
		assert.Equal(t, "vsw-test", mdl.LoadBalancerAttribute.VSwitchId)
	})

	t.Run("set default spec for paybybandwidth", func(t *testing.T) {
		mdl := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				InstanceChargeType: model.PayBySpec,
			},
		}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.NotEmpty(t, mdl.LoadBalancerAttribute.LoadBalancerSpec)
	})

	t.Run("set paybyclcu", func(t *testing.T) {
		mdl := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				InstanceChargeType: model.PayByCLCU,
			},
		}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.Equal(t, model.PayByCLCU, mdl.LoadBalancerAttribute.InstanceChargeType)
	})

	t.Run("intranet classic network sets vpc from config", func(t *testing.T) {
		oldVpcID := ctrlcfg.CloudCFG.Global.VpcID
		defer func() { ctrlcfg.CloudCFG.Global.VpcID = oldVpcID }()
		ctrlcfg.CloudCFG.Global.VpcID = "vpc-classic-test"
		mdl := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				AddressType:  model.IntranetAddressType,
				NetworkType:  model.ClassicNetworkType,
				VpcId:        "",
				VSwitchId:    "vsw-1",
			},
		}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.Equal(t, "vpc-classic-test", mdl.LoadBalancerAttribute.VpcId)
	})

	t.Run("set default delete protection and modification protection", func(t *testing.T) {
		mdl := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				DeleteProtection:             "",
				ModificationProtectionStatus: "",
			},
		}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.Equal(t, model.FlagType(model.OnFlag), mdl.LoadBalancerAttribute.DeleteProtection)
		assert.Equal(t, model.ModificationProtectionType(model.ConsoleProtection), mdl.LoadBalancerAttribute.ModificationProtectionStatus)
		assert.NotEmpty(t, mdl.LoadBalancerAttribute.ModificationProtectionReason)
	})

	t.Run("set default resource group id from config", func(t *testing.T) {
		oldRG := ctrlcfg.CloudCFG.Global.ResourceGroupID
		defer func() { ctrlcfg.CloudCFG.Global.ResourceGroupID = oldRG }()
		ctrlcfg.CloudCFG.Global.ResourceGroupID = "rg-default-test"
		mdl := &model.LoadBalancer{}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.Equal(t, "rg-default-test", mdl.LoadBalancerAttribute.ResourceGroupId)
	})

	t.Run("append default tags", func(t *testing.T) {
		mdl := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				Tags: []tag.Tag{{Key: "existing", Value: "tag"}},
			},
		}
		svc := getDefaultService()
		anno := annotation.NewAnnotationRequest(svc)
		err := setModelDefaultValue(mgr, mdl, anno)
		assert.NoError(t, err)
		assert.True(t, len(mdl.LoadBalancerAttribute.Tags) >= 1)
		assert.Equal(t, "existing", mdl.LoadBalancerAttribute.Tags[len(mdl.LoadBalancerAttribute.Tags)-1].Key)
	})
}

func TestAddTagIfNotExist(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())
	svc := getDefaultService()
	reqCtx := getReqCtx(svc)

	t.Run("tag not exist, should add", func(t *testing.T) {
		remote := model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "lb-test",
				Tags: []tag.Tag{
					{Key: "key1", Value: "value1"},
				},
			},
		}
		newTag := tag.Tag{Key: "key2", Value: "value2"}
		err := mgr.addTagIfNotExist(reqCtx, remote, newTag)
		assert.NoError(t, err)
	})

	t.Run("tag already exist, should skip", func(t *testing.T) {
		remote := model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "lb-test",
				Tags: []tag.Tag{
					{Key: "key1", Value: "value1"},
				},
			},
		}
		newTag := tag.Tag{Key: "key1", Value: "value2"}
		err := mgr.addTagIfNotExist(reqCtx, remote, newTag)
		assert.NoError(t, err)
	})

	t.Run("too many tags", func(t *testing.T) {
		var tags []tag.Tag
		for i := 0; i < MaxLBTagNum; i++ {
			tags = append(tags, tag.Tag{Key: "key", Value: "value"})
		}
		remote := model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "lb-test",
				Tags:           tags,
			},
		}
		newTag := tag.Tag{Key: "newkey", Value: "newvalue"}
		err := mgr.addTagIfNotExist(reqCtx, remote, newTag)
		assert.NoError(t, err)
	})
}

func TestUpdateInstanceChargeTypeAndInstanceSpec(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())

	t.Run("no change", func(t *testing.T) {
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)
		local := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				InstanceChargeType: model.PayBySpec,
				LoadBalancerSpec:   model.S1Small,
			},
		}
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId:     "lb-test",
				InstanceChargeType: model.PayBySpec,
				LoadBalancerSpec:   model.S1Small,
			},
		}
		err := mgr.updateInstanceChargeTypeAndInstanceSpec(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("change charge type", func(t *testing.T) {
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)
		local := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				InstanceChargeType: model.PayByCLCU,
			},
		}
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId:     "lb-test",
				InstanceChargeType: model.PayBySpec,
				LoadBalancerSpec:   model.S1Small,
			},
		}
		err := mgr.updateInstanceChargeTypeAndInstanceSpec(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("change spec", func(t *testing.T) {
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)
		local := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				InstanceChargeType: model.PayBySpec,
				LoadBalancerSpec:   model.S1Small,
			},
		}
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId:     "lb-test",
				InstanceChargeType: model.PayBySpec,
				LoadBalancerSpec:   model.S1Small,
			},
		}
		err := mgr.updateInstanceChargeTypeAndInstanceSpec(reqCtx, local, remote)
		assert.NoError(t, err)
	})
}

func TestLoadBalancerManager_BuildLocalModel(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())

	t.Run("build basic model", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[annotation.Annotation(annotation.AddressType)] = string(model.InternetAddressType)
		svc.Annotations[annotation.Annotation(annotation.Bandwidth)] = "5"
		reqCtx := getReqCtx(svc)
		mdl := &model.LoadBalancer{
			NamespacedName: util.NamespacedName(svc),
		}

		err := mgr.BuildLocalModel(reqCtx, mdl)
		assert.NoError(t, err)
		assert.Equal(t, model.InternetAddressType, mdl.LoadBalancerAttribute.AddressType)
		assert.Equal(t, 5, mdl.LoadBalancerAttribute.Bandwidth)
	})

	t.Run("build with user managed lb", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[annotation.Annotation(annotation.LoadBalancerId)] = "lb-test"
		reqCtx := getReqCtx(svc)
		mdl := &model.LoadBalancer{
			NamespacedName: util.NamespacedName(svc),
		}

		err := mgr.BuildLocalModel(reqCtx, mdl)
		assert.NoError(t, err)
		assert.True(t, mdl.LoadBalancerAttribute.IsUserManaged)
		assert.Equal(t, "lb-test", mdl.LoadBalancerAttribute.LoadBalancerId)
	})

	t.Run("invalid bandwidth", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[annotation.Annotation(annotation.Bandwidth)] = "invalid"
		reqCtx := getReqCtx(svc)
		mdl := &model.LoadBalancer{
			NamespacedName: util.NamespacedName(svc),
		}

		err := mgr.BuildLocalModel(reqCtx, mdl)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bandwidth must be integer")
	})
}

func TestLoadBalancerManager_UpdateLoadBalancerTags(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())

	t.Run("add new tags", func(t *testing.T) {
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)
		local := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				Tags: []tag.Tag{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
			},
		}
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "lb-test",
				Tags: []tag.Tag{
					{Key: "key1", Value: "value1"},
				},
			},
		}

		err := mgr.updateLoadBalancerTags(reqCtx, local, remote)
		assert.NoError(t, err)
	})

	t.Run("remove tags", func(t *testing.T) {
		svc := getDefaultService()
		reqCtx := getReqCtx(svc)
		local := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				Tags: []tag.Tag{
					{Key: "key1", Value: "value1"},
				},
			},
		}
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "lb-test",
				Tags: []tag.Tag{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
			},
		}

		err := mgr.updateLoadBalancerTags(reqCtx, local, remote)
		assert.NoError(t, err)
	})
}

func TestLoadBalancerManager_CleanupLoadBalancerTags(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())
	svc := getDefaultService()
	reqCtx := getReqCtx(svc)

	t.Run("cleanup tags", func(t *testing.T) {
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "lb-test",
				Tags: []tag.Tag{
					{Key: "kubernetes.io/cluster/test", Value: "owned"},
					{Key: "otherkey", Value: "value"},
				},
			},
		}

		err := mgr.CleanupLoadBalancerTags(reqCtx, remote)
		assert.NoError(t, err)
	})

	t.Run("empty lb id", func(t *testing.T) {
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId: "",
			},
		}

		err := mgr.CleanupLoadBalancerTags(reqCtx, remote)
		assert.NoError(t, err)
	})
}

func TestLoadBalancerManager_SetProtectionsOff(t *testing.T) {
	mgr := NewLoadBalancerManager(getMockCloudProvider())
	svc := getDefaultService()
	reqCtx := getReqCtx(svc)

	t.Run("turn off protections", func(t *testing.T) {
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId:               "lb-test",
				DeleteProtection:             model.OnFlag,
				ModificationProtectionStatus: model.ConsoleProtection,
			},
		}

		err := mgr.SetProtectionsOff(reqCtx, remote)
		assert.NoError(t, err)
	})

	t.Run("protections already off", func(t *testing.T) {
		remote := &model.LoadBalancer{
			LoadBalancerAttribute: model.LoadBalancerAttribute{
				LoadBalancerId:               "lb-test",
				DeleteProtection:             model.OffFlag,
				ModificationProtectionStatus: model.NonProtection,
			},
		}

		err := mgr.SetProtectionsOff(reqCtx, remote)
		assert.NoError(t, err)
	})
}
