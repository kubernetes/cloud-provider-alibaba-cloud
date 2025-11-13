package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNodeList(t *testing.T) {
	nodeList := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "normal-node",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "exclude-label",
					Labels: map[string]string{
						helper.LabelNodeExcludeNode: "true",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "legacy-exclude-label",
					Labels: map[string]string{
						helper.LabelNodeExcludeNodeDeprecated: "true",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "empty-provider-id",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "hybrid-node",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "ack-hybrid://123456",
				},
			},
		},
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(nodeList).Build()

	nodes, err := NodeList(client, false)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(nodes.Items))
	assert.Equal(t, "normal-node", nodes.Items[0].Name)
}

func TestGetNodeType(t *testing.T) {
	cases := []struct {
		name string
		node *v1.Node
		want NodeType
		ok   bool
	}{
		{
			name: "normal ecs node",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "normal-node",
					Labels: map[string]string{},
				},
				Spec: v1.NodeSpec{
					ProviderID: "cn-hangzhou.i-123456",
				},
			},
			want: NodeTypeECS,
			ok:   true,
		},
		{
			name: "lingjun node",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lingjun-node",
					Labels: map[string]string{
						LabelLingJunWorker: "true",
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "e01-123456",
				},
			},
			want: NodeTypeLingJun,
			ok:   true,
		},
		{
			name: "unknown node",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "unknown-node",
				},
				Spec: v1.NodeSpec{
					ProviderID: "unknown-provider-id",
				},
			},
			ok: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, ok := getNodeType(c.node)
			assert.Equal(t, c.ok, ok)
			if c.ok {
				assert.Equal(t, c.want, actual)
			}
		})
	}
}
