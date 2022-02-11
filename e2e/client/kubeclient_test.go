package client

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"testing"
)

func CreateKubeClient() *KubeClient {
	cfg := config.GetConfigOrDie()
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return &KubeClient{client}
}

func TestKubeClient_GetLatestNode(t *testing.T) {
	client := CreateKubeClient()
	node, err := client.GetLatestNode()
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("node name: %s", node.Name)
}

func TestKubeClient_PatchNodeStatus(t *testing.T) {
	client := CreateKubeClient()
	oldNode, err := client.GetLatestNode()
	if err != nil {
		t.Fatalf(err.Error())
	}
	newNode := oldNode.DeepCopy()
	for index, value := range newNode.Status.Addresses {
		if value.Type == v1.NodeInternalIP {
			newNode.Status.Addresses[index].Address = "123.123.123.123"
		}
	}

	updated, err := client.PatchNodeStatus(oldNode, newNode)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("updated address: %s", updated.Status.Addresses)
}
