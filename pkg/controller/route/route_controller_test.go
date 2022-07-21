package route

import (
	"context"
	cmap "github.com/orcaman/concurrent-map"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	NodeName = "cn-hangzhou.192.0.168.68"
)

func TestUpdateNetworkingCondition(t *testing.T) {
	r := getReconcileRoute()
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: NodeName,
		},
		Spec: v1.NodeSpec{
			PodCIDR:    "10.96.0.64/26",
			ProviderID: "cn-hangzhou.ecs-id",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Reason: string(v1.NodeReady),
					Status: v1.ConditionTrue,
				},
				{
					Reason: string(v1.NodeNetworkUnavailable),
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	err := r.updateNetworkingCondition(context.TODO(), node, true)
	if err != nil {
		t.Error(err)
	}

	updatedNode := &v1.Node{}
	err = r.client.Get(context.TODO(), util.NamespacedName(node), updatedNode)
	if err != nil {
		t.Error(err)
	}

	networkCondition, ok := helper.FindCondition(node.Status.Conditions, v1.NodeNetworkUnavailable)
	if !ok || networkCondition.Status != v1.ConditionFalse {
		t.Error("node condition update failed")
	}

}

func getReconcileRoute() *ReconcileRoute {
	eventRecord := record.NewFakeRecorder(100)
	recon := &ReconcileRoute{
		cloud:        getMockCloudProvider(),
		client:       getFakeKubeClient(),
		record:       eventRecord,
		nodeCache:    cmap.New(),
		configRoutes: true,
	}

	return recon
}

func getMockCloudProvider() prvd.Provider {
	return vmock.MockCloud{
		MockVPC:   vmock.NewMockVPC(nil),
		IMetaData: vmock.NewMockMetaData("vpc-single-route-table"),
	}
}

func getFakeKubeClient() client.Client {
	// Node
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: NodeName,
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.64/26",
					ProviderID: "cn-hangzhou.ecs-id",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
						},
						{
							Reason: string(v1.NodeNetworkUnavailable),
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cn-hangzhou.192.0.168.69",
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.128/26",
					ProviderID: "alicloud://cn-hangzhou.ecs-id",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
						},
						{
							Reason: string(v1.NodeNetworkUnavailable),
							Status: v1.ConditionFalse,
						},
					},
				},
			},
		},
	}

	objs := []runtime.Object{nodeList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}
