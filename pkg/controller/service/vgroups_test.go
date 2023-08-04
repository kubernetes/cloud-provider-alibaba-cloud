package service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/dryrun"
	"k8s.io/klog/v2/klogr"
	"testing"
)

func TestVGroupManager_BatchSyncVServerGroupBackendServers(t *testing.T) {
	vgroupManager := getTestVGroupManager()

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
	reqCtx := &RequestContext{
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
