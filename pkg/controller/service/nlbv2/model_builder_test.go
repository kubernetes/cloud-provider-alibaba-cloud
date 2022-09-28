package nlbv2

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestModelBuilder_BuildModel(t *testing.T) {
	nlbManager := NewNLBManager(getMockCloudProvider())
	listenerManager := NewListenerManager(getMockCloudProvider())
	serverGroupManager, err := NewServerGroupManager(getFakeKubeClient(), getMockCloudProvider())
	if err != nil {
		t.Error(err)
	}
	builder := NewModelBuilder(nlbManager, listenerManager, serverGroupManager)
	cl := getFakeKubeClient()
	svc := &v1.Service{}
	_ = cl.Get(context.TODO(), types.NamespacedName{Namespace: v1.NamespaceDefault, Name: ServiceName}, svc)

	_, err = builder.Instance(RemoteModel).Build(getReqCtx(svc))
	if err != nil {
		t.Error(err)
	}
}
