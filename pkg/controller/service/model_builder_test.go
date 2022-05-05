package service

import (
	"testing"
)

func TestBuildModel(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}
	reqCtx := getReqCtxForListener()
	if _, err := builder.Instance(LocalModel).Build(reqCtx); err != nil {
		t.Error(err)
	}

	if _, err := builder.Instance(RemoteModel).Build(reqCtx); err != nil {
		t.Error(err)
	}

}
