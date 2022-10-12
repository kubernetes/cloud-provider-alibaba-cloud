package helper

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestPatchM(t *testing.T) {
	diff := func(copy runtime.Object) (client.Object, error) {
		nins := copy.(*v1.Node)
		condition, ok := FindCondition(nins.Status.Conditions, v1.NodeNetworkUnavailable)
		condition.Type = v1.NodeNetworkUnavailable
		condition.Status = v1.ConditionFalse
		condition.Reason = "RouteCreated"
		condition.Message = "RouteController created a route"

		if !ok {
			nins.Status.Conditions = append(nins.Status.Conditions, *condition)
		}
		return nins, nil
	}

	mclient := getFakeKubeClient()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-1",
			Labels: map[string]string{"app": "nginx"},
		},
	}

	err := PatchM(mclient, node1, diff, PatchStatus)
	if err != nil {
		t.Error(err)
	}
	node1Patched := &v1.Node{}
	_ = mclient.Get(context.TODO(), client.ObjectKey{
		Name: node1.Name,
	}, node1Patched)

	for _, con := range node1Patched.Status.Conditions {
		if con.Type == v1.NodeNetworkUnavailable {
			assert.Equal(t, v1.ConditionFalse, con.Status)
			return
		}
	}
	t.Error("Fail to patch node network status")

}

func getFakeKubeClient() client.Client {
	// Node
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-1",
					Labels: map[string]string{"app": "nginx"},
				},
				Spec: v1.NodeSpec{
					PodCIDR:    "10.96.0.64/26",
					ProviderID: "cn-hangzhou.ecs-id",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Reason: string(v1.NodeReady),
							Status: v1.ConditionTrue,
						},
						{
							Type:   v1.NodeNetworkUnavailable,
							Reason: "No Route Created",
							Status: v1.ConditionTrue,
						},
					},
				},
			},
		},
	}

	objs := []runtime.Object{nodeList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}
