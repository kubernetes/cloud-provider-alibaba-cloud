package route

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestUpdateNetworkingCondition(t *testing.T) {
	cl, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}
	r := &ReconcileRoute{client: cl}

	nodes := &v1.NodeList{}
	if err := cl.List(context.TODO(), nodes); err != nil {
		t.Fatalf(err.Error())
	}
	for _, n := range nodes.Items {
		if err := r.updateNetworkingCondition(context.TODO(), &n, true); err != nil {
			t.Fatalf(err.Error())
		}
	}

}
