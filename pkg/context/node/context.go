package node

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/context/base"
)

func NewNodeContext(
	node *v1.Node,
) *NodeContext {
	ctxs := NodeContext{}
	ctxs.SetKV(Node, node)
	return &ctxs
}

const (
	Node     = "Node"
	Task     = "Task"
	NodePool = "NodePool"
)

type NodeContext struct{ base.Context }

// Node
func (c *NodeContext) Node() (*v1.Node, error) {
	node, ok := c.Value(Node)
	if !ok {
		return nil, fmt.Errorf("node not found")
	}
	return node.(*v1.Node), nil
}
