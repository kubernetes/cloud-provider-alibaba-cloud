package node

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/controller/helper"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
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

	nodes, err := NodeList(client)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(nodes.Items))
	assert.Equal(t, "normal-node", nodes.Items[0].Name)
}
