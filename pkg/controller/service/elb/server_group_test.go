package elb

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	svcCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service/reconcile/context"
	elbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/elb"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"testing"
)

func MockSGManager() *ServerGroupManager {
	return &ServerGroupManager{
		kubeClient: MockKubeClient(),
		cloud:      vmock.NewMockCloud(&base.ClientMgr{}),
	}
}

func TestServerGroupManager_BatchAddServerGroup(t *testing.T) {
	mgr := MockSGManager()
	nodeList := &corev1.NodeList{}
	err := mgr.kubeClient.List(context.TODO(), nodeList)
	if err != nil {
		t.Error("failed to list nodes")
	}
	sg := &elbmodel.EdgeServerGroup{}
	for _, node := range nodeList.Items {
		sg.Backends = append(sg.Backends, elbmodel.EdgeBackendAttribute{
			ServerId: node.Name,
			ServerIp: node.Status.Addresses[0].Address,
			Weight:   elbmodel.ServerGroupDefaultServerWeight,
			Port:     elbmodel.ServerGroupDefaultPort,
			Type:     elbmodel.ServerGroupDefaultType,
		})
	}

	err = mgr.batchAddServerGroup(&svcCtx.RequestContext{Ctx: context.TODO()}, "", sg)
	if err != nil {
		t.Error("failed to batch add server group")
	}

}
