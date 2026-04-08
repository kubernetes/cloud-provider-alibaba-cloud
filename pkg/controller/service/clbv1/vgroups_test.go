package clbv1

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/annotation"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/backend"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/klog/v2/klogr"
)

func TestVGroupManager_BatchSyncVServerGroupBackendServers(t *testing.T) {
	vgroupManager, _ := getTestVGroupManager()

	var local, remote []model.BackendAttribute
	for i := 0; i < 200; i++ {
		remote = append(remote, model.BackendAttribute{
			ServerIp:    fmt.Sprintf("192.168.0.%d", i),
			ServerId:    fmt.Sprintf("eni-old-%d", i),
			Port:        8080,
			Weight:      100,
			Description: "k8s/svc-1/test/8080/clusterid",
		})
	}

	for i := 0; i < 200; i++ {
		local = append(local, model.BackendAttribute{
			ServerIp:    fmt.Sprintf("192.168.1.%d", i),
			ServerId:    fmt.Sprintf("eni-new-%d", i),
			Port:        8080,
			Weight:      100,
			Description: "k8s/svc-1/test/8080/clusterid",
		})
	}

	localVGroup := model.VServerGroup{
		VGroupId:   "rsp-test",
		VGroupName: "k8s/svc-1/test/8080/clusterid",
		Backends:   local,
	}
	remoteVGroup := model.VServerGroup{
		VGroupId:   "rsp-test",
		VGroupName: "k8s/svc-1/test/8080/clusterid",
		Backends:   remote,
	}
	ctx := context.WithValue(context.TODO(), dryrun.ContextKey("vgroup"), &remoteVGroup)
	reqCtx := &svcCtx.RequestContext{
		Ctx: ctx,
		Log: klogr.New(),
	}

	err := vgroupManager.UpdateVServerGroup(reqCtx, localVGroup, remoteVGroup)
	if err != nil {
		t.Error(err)
	}

	vgroup, ok := ctx.Value(dryrun.ContextKey("vgroup")).(*model.VServerGroup)
	if !ok {
		t.Errorf("no vgroup in context, %+v", ctx.Value("vGroup"))
	}

	assert.Equal(t, len(vgroup.Backends), 200)
}

// --- setWeightBackends ---

func TestSetWeightBackends_NilWeightNilDefault_UsesDefaultServerWeight(t *testing.T) {
	backends := []model.BackendAttribute{
		{ServerId: "ecs-1", Type: model.ECSBackendType},
		{ServerId: "ecs-2", Type: model.ECSBackendType},
	}
	result := setWeightBackends(helper.ENITrafficPolicy, backends, nil, nil)
	for _, b := range result {
		assert.Equal(t, DefaultServerWeight, b.Weight)
	}
}

func TestSetWeightBackends_NilWeightWithCustomDefault_UsesCustomDefaultWeight(t *testing.T) {
	backends := []model.BackendAttribute{
		{ServerId: "ecs-1", Type: model.ECSBackendType},
		{ServerId: "ecs-2", Type: model.ECSBackendType},
	}
	customDefault := 50
	result := setWeightBackends(helper.ENITrafficPolicy, backends, nil, &customDefault)
	for _, b := range result {
		assert.Equal(t, 50, b.Weight)
	}
}

func TestSetWeightBackends_WithWeight_IgnoresDefaultWeight(t *testing.T) {
	backends := []model.BackendAttribute{
		{ServerId: "ecs-1", Type: model.ECSBackendType},
		{ServerId: "ecs-2", Type: model.ECSBackendType},
	}
	weight := 60
	customDefault := 50
	// podPercentAlgorithm: 60 total / 2 backends = 30 each
	result := setWeightBackends(helper.ENITrafficPolicy, backends, &weight, &customDefault)
	for _, b := range result {
		assert.Equal(t, 30, b.Weight)
	}
}

// --- podNumberAlgorithm ---

func TestPodNumberAlgorithm_ENIMode_AppliesCustomDefaultWeight(t *testing.T) {
	backends := []model.BackendAttribute{
		{ServerId: "eni-1"},
		{ServerId: "eni-2"},
		{ServerId: "eni-3"},
	}
	result := podNumberAlgorithm(helper.ENITrafficPolicy, backends, 75)
	for _, b := range result {
		assert.Equal(t, 75, b.Weight)
	}
}

func TestPodNumberAlgorithm_ClusterMode_AppliesCustomDefaultWeight(t *testing.T) {
	backends := []model.BackendAttribute{
		{ServerId: "ecs-1", Type: model.ECSBackendType},
		{ServerId: "ecs-2", Type: model.ECSBackendType},
	}
	result := podNumberAlgorithm(helper.ClusterTrafficPolicy, backends, 30)
	for _, b := range result {
		assert.Equal(t, 30, b.Weight)
	}
}

func TestPodNumberAlgorithm_LocalMode_WeightByPodCount(t *testing.T) {
	// defaultWeight is not used in LocalMode; weight = number of pods on the node
	backends := []model.BackendAttribute{
		{ServerId: "ecs-1", Type: model.ECSBackendType},
		{ServerId: "ecs-1", Type: model.ECSBackendType},
		{ServerId: "ecs-2", Type: model.ECSBackendType},
	}
	result := podNumberAlgorithm(helper.LocalTrafficPolicy, backends, 50)
	assert.Equal(t, 2, result[0].Weight) // ecs-1 has 2 pods
	assert.Equal(t, 2, result[1].Weight)
	assert.Equal(t, 1, result[2].Weight) // ecs-2 has 1 pod
}

// --- buildVGroupForServicePort annotation parsing ---

func TestBuildVGroupForServicePort_InvalidDefaultWeight_NotInteger(t *testing.T) {
	svc := getDefaultService()
	svc.Annotations[annotation.Annotation(annotation.DefaultWeight)] = "abc"
	reqCtx := getReqCtx(svc)
	mgr, _ := getTestVGroupManager()
	candidates := &backend.EndpointWithENI{}

	_, _, err := mgr.buildVGroupForServicePort(reqCtx, svc.Spec.Ports[0], candidates, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default weight must be integer in range [0,100]")
}

func TestBuildVGroupForServicePort_InvalidDefaultWeight_TooLarge(t *testing.T) {
	svc := getDefaultService()
	svc.Annotations[annotation.Annotation(annotation.DefaultWeight)] = "150"
	reqCtx := getReqCtx(svc)
	mgr, _ := getTestVGroupManager()
	candidates := &backend.EndpointWithENI{}

	_, _, err := mgr.buildVGroupForServicePort(reqCtx, svc.Spec.Ports[0], candidates, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default weight must be integer in range [0,100]")
}

func TestBuildVGroupForServicePort_InvalidDefaultWeight_Negative(t *testing.T) {
	svc := getDefaultService()
	svc.Annotations[annotation.Annotation(annotation.DefaultWeight)] = "-1"
	reqCtx := getReqCtx(svc)
	mgr, _ := getTestVGroupManager()
	candidates := &backend.EndpointWithENI{}

	_, _, err := mgr.buildVGroupForServicePort(reqCtx, svc.Spec.Ports[0], candidates, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default weight must be integer in range [0,100]")
}
