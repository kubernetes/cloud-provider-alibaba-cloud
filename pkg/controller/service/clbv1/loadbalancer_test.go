package clbv1

import (
	"context"
	"fmt"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/stretchr/testify/assert"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
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
