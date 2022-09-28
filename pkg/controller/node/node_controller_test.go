package node

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/cloud-provider/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	NodeName = fmt.Sprintf("cn-hangzhou.%s", vmock.InstanceIP)
)

func TestSyncCloudNode(t *testing.T) {
	recon := getReconcileNode()
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: NodeName,
		},
	}
	err := getFakeKubeClient().Get(context.TODO(), util.NamespacedName(node), node)
	if err != nil {
		t.Error(err)
	}

	if err = recon.syncCloudNode(node); err != nil {
		t.Error(err)
	}

	updatedNode := &v1.Node{}
	if err := recon.client.Get(context.TODO(), util.NamespacedName(node), updatedNode); err != nil {
		t.Error(err)
	}

	if len(updatedNode.Spec.Taints) != 0 {
		t.Errorf("remove taint error, taints: %+v", updatedNode.Spec.Taints)
	}
	if instanceType, ok := updatedNode.Labels[v1.LabelInstanceType]; !ok || instanceType != vmock.InstanceType {
		t.Errorf("node label LabelInstanceType not equal, expect %s, got %s", vmock.InstanceType, instanceType)
	}
	if zone, ok := updatedNode.Labels[v1.LabelTopologyZone]; !ok || zone != vmock.ZoneID {
		t.Errorf("node label LabelTopologyZone not equal, expect %s, got %s", vmock.ZoneID, zone)
	}
	if region, ok := updatedNode.Labels[v1.LabelTopologyRegion]; !ok || region != vmock.RegionID {
		t.Errorf("node label LabelTopologyRegion not equal, expect %s, got %s", vmock.RegionID, region)
	}
	if len(updatedNode.Status.Addresses) == 0 {
		t.Error("node address is empty")
	}
	for _, addr := range updatedNode.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			if addr.Address != vmock.InstanceIP {
				t.Errorf("node internal ip address not equal, expect %s, got %s", vmock.InstanceIP, addr.Address)
			}
		}
	}
}

func getReconcileNode() *ReconcileNode {
	eventRecord := record.NewFakeRecorder(100)
	recon := &ReconcileNode{
		cloud:  getMockCloudProvider(),
		client: getFakeKubeClient(),
		record: eventRecord,
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
					Taints: []v1.Taint{
						{
							Key:   api.TaintExternalCloudProvider,
							Value: "true",
						},
					},
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
