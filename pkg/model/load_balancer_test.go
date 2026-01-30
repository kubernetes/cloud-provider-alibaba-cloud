package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestParseFlagType(t *testing.T) {
	cases := []struct {
		flag    string
		value   FlagType
		wantErr bool
	}{
		{
			flag:  "on",
			value: OnFlag,
		},
		{
			flag:  "off",
			value: OffFlag,
		},
		{
			flag:  "ON",
			value: OnFlag,
		},
		{
			flag:  "Off",
			value: OffFlag,
		},
		{
			flag:    "true",
			wantErr: true,
		},
	}

	for _, c := range cases {
		value, err := ParseFlagType(c.flag)
		if c.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, c.value, value)
		}
	}
}
