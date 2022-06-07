package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLoadBalancerId(t *testing.T) {
	var lb *LoadBalancer
	id := lb.GetLoadBalancerId()
	assert.Equal(t, id, "")

	lb2 := &LoadBalancer{}
	id2 := lb2.GetLoadBalancerId()
	assert.Equal(t, id2, "")

	lb3 := &LoadBalancer{
		LoadBalancerAttribute: LoadBalancerAttribute{
			LoadBalancerId: "lb-xxxx",
		},
	}
	id3 := lb3.GetLoadBalancerId()
	assert.Equal(t, id3, "lb-xxxx")
}
