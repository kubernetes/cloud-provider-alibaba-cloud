package nlbv2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
	"testing"
)

func singlePort(p int32) *nlbmodel.ListenerAttribute {
	return &nlbmodel.ListenerAttribute{
		ListenerPort: p,
	}
}
func rangePort(start, end int32) *nlbmodel.ListenerAttribute {
	return &nlbmodel.ListenerAttribute{
		StartPort: start,
		EndPort:   end,
	}
}

func TestIsListenerPortOverlapped(t *testing.T) {

	cases := []struct {
		a          *nlbmodel.ListenerAttribute
		b          *nlbmodel.ListenerAttribute
		overlapped bool
	}{
		{
			a:          singlePort(80),
			b:          singlePort(443),
			overlapped: false,
		},
		{
			a:          singlePort(80),
			b:          singlePort(80),
			overlapped: true,
		},
		{
			a:          singlePort(80),
			b:          rangePort(1, 100),
			overlapped: true,
		},
		{
			a:          rangePort(1, 100),
			b:          singlePort(80),
			overlapped: true,
		},
		{
			a:          rangePort(1, 100),
			b:          rangePort(101, 200),
			overlapped: false,
		},
		{
			a:          rangePort(101, 200),
			b:          rangePort(1, 100),
			overlapped: false,
		},
		{
			a:          rangePort(1, 101),
			b:          rangePort(101, 200),
			overlapped: true,
		},
		{
			a:          rangePort(101, 200),
			b:          rangePort(1, 101),
			overlapped: true,
		},
		{
			a:          rangePort(1, 102),
			b:          rangePort(50, 80),
			overlapped: true,
		},
		{
			a:          rangePort(50, 80),
			b:          rangePort(1, 102),
			overlapped: true,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			assert.Equal(t, c.overlapped, isListenerPortOverlapped(c.a, c.b))
		})
	}
}
