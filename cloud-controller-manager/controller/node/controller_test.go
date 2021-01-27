package node

import (
	"context"
	"flag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
	"testing"
)

func TestPatchNode(t *testing.T) {
	var kubeConfig = flag.String("kubeconfig", "", "kubernetes config path")
	var nodeName = flag.String("nodeName", "", "patch node name")
	if kubeConfig == nil || nodeName == nil {
		klog.Info("kubeconfig or nodeName is nil, skip")
		t.Skip()
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf(err.Error())
	}

	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), *nodeName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeNetworkUnavailable &&
			condition.Status == v1.ConditionFalse {
			klog.Infof("node %s network is available", node.Name)
			return
		}
	}

	if err = nodeutil.SetNodeCondition(clientset, types.NodeName(node.Name), v1.NodeCondition{
		Type:               v1.NodeNetworkUnavailable,
		Status:             v1.ConditionFalse,
		Reason:             "NoRouteCreated",
		Message:            "Node created without a route",
		LastTransitionTime: metav1.Now(),
	}); err != nil {
		t.Fatalf(err.Error())
	}

	// fetch latest node from API server since alicloud-specific condition was set and informer cache may be stale
	_, err = clientset.CoreV1().Nodes().Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get node %s error: %s", node.Name, err.Error())
	}
}
