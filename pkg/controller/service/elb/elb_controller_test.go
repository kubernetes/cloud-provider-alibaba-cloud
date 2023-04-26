package elb

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	EdgeNodeLabelKey = "alibabacloud.com/is-edge-worker"
	ENSNodeLabelKey  = "alibabacloud.com/ens-instance-id"
	MaxENSNumber     = 45
)

func MockKubeClient() client.Client {
	Items := make([]corev1.Node, 0, MaxENSNumber)
	for i := 0; i < MaxENSNumber; i++ {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("ens-%d", i),
				Labels: map[string]string{
					EdgeNodeLabelKey: "true",
					ENSNodeLabelKey:  fmt.Sprintf("ens-id-%d", i),
				},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{
						Type:    corev1.NodeInternalIP,
						Address: fmt.Sprintf("10.0.0.%d", i+1),
					},
				},
			},
		}
		Items = append(Items, node)
	}
	nodeList := &corev1.NodeList{
		Items: Items,
	}
	objs := []runtime.Object{nodeList}
	return fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
}
