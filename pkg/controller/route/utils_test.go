package route

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetIPv4RouteForNode(t *testing.T) {
	var testNodeGrid = []struct {
		node       v1.Node
		expectCidr string
		err        bool
	}{
		{
			node:       v1.Node{Spec: v1.NodeSpec{PodCIDR: "192.168.0.0/24", PodCIDRs: []string{"192.168.0.0/24", "fe80::1/96"}}},
			expectCidr: "192.168.0.0/24",
			err:        false,
		},
		{
			node:       v1.Node{Spec: v1.NodeSpec{PodCIDR: "", PodCIDRs: []string{"192.168.0.0/24", "fe80::1/96"}}},
			expectCidr: "192.168.0.0/24",
			err:        false,
		},
		{
			node:       v1.Node{Spec: v1.NodeSpec{PodCIDR: "192.168.0.0/24"}},
			expectCidr: "192.168.0.0/24",
			err:        false,
		},
		{
			node:       v1.Node{Spec: v1.NodeSpec{}},
			expectCidr: "",
			err:        false,
		},
		{
			node:       v1.Node{Spec: v1.NodeSpec{PodCIDR: "192.168.111.11111/24", PodCIDRs: []string{"192.168.111.11111/24", "fe80::1/96"}}},
			expectCidr: "",
			err:        true,
		},
	}
	for _, testcase := range testNodeGrid {
		ipnet, cidr, err := getIPv4RouteForNode(&testcase.node)
		assert.Equal(t, testcase.expectCidr, cidr)
		if !testcase.err && testcase.expectCidr != "" {
			assert.NotEmpty(t, ipnet)
		}
		assert.Equal(t, testcase.err, err != nil)
	}
}
