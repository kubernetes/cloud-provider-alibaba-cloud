package hash

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHashObject(t *testing.T) {
	o := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"key": "value",
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "True",
					Reason: "OK",
				},
				{
					Type:   "DiskFull",
					Reason: "OK",
				},
			},
		},
	}

	arr := []interface{}{o.Status.Conditions, o.Labels, o.Spec.Unschedulable}

	assert.Equal(t, HashObject(arr), "0bca66e81b3124c0ea1c8a6e16c4a12092546c00bb0e75db53de5c2b")
}

func TestHashLabel(t *testing.T) {
	o := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				ReconcileHashLable: "fbd209056832bd23dc0334e04c2eea5bf2c5541e501ca2ab747ce7e3",
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   "Ready",
					Status: "True",
					Reason: "OK",
				},
			},
		},
	}

	arr := []interface{}{o.Status.Conditions, o.Labels, o.Spec.Unschedulable}
	//fmt.Printf("%s\n\n\n",PrettyYaml(arr))
	//fmt.Printf("%s\n\n\n",HashString(arr))
	//fmt.Printf(HashObject(arr))

	assert.Equal(t, HashObject(arr), "03e8cc3ccf0b0b0bff80371bb738d5e8fa755e9da3d022a7617ce612")
}
