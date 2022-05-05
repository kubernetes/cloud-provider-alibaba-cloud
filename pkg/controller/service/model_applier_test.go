package service

import "testing"

func TestApply(t *testing.T) {
	builder := &ModelBuilder{
		LoadBalancerMgr: NewLoadBalancerManager(getMockCloudProvider()),
		ListenerMgr:     NewListenerManager(getMockCloudProvider()),
		VGroupMgr:       getTestVGroupManager(),
	}

	applier := NewModelApplier(NewLoadBalancerManager(getMockCloudProvider()),
		NewListenerManager(getMockCloudProvider()), getTestVGroupManager())
	reqCtx := getReqCtxForListener()

	localModel, err := builder.Instance(LocalModel).Build(reqCtx)
	if err != nil {
		t.Error(err)
	}
	_, err = applier.Apply(reqCtx, localModel)
	if err != nil {
		t.Error(err)

	}
}
