package nlbv2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	nlbmodel "k8s.io/cloud-provider-alibaba-cloud/pkg/model/nlb"
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

func TestServerGroup(t *testing.T) {
	tests := []struct {
		name        string
		annotation  string
		port        v1.ServicePort
		expected    string
		expectError bool
	}{
		{
			name:       "single server group match",
			annotation: "sg-123:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expected:    "sg-123",
			expectError: false,
		},
		{
			name:       "multiple server groups match first",
			annotation: "sg-123:443,sg-456:80",
			port: v1.ServicePort{
				Port: 443,
			},
			expected:    "sg-123",
			expectError: false,
		},
		{
			name:       "multiple server groups match second",
			annotation: "sg-123:443,sg-456:80",
			port: v1.ServicePort{
				Port: 80,
			},
			expected:    "sg-456",
			expectError: false,
		},
		{
			name:       "no match returns empty",
			annotation: "sg-123:443,sg-456:80",
			port: v1.ServicePort{
				Port: 8080,
			},
			expected:    "",
			expectError: false,
		},
		{
			name:       "empty annotation returns error",
			annotation: "",
			port: v1.ServicePort{
				Port: 443,
			},
			expected:    "",
			expectError: true,
		},
		{
			name:       "invalid format missing colon",
			annotation: "sg-123",
			port: v1.ServicePort{
				Port: 443,
			},
			expected:    "",
			expectError: true,
		},
		{
			name:       "empty server group id returns empty",
			annotation: ":443",
			port: v1.ServicePort{
				Port: 443,
			},
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := serverGroup(tt.annotation, tt.port)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPortRange(t *testing.T) {
	tests := []struct {
		name        string
		annotation  string
		port        v1.ServicePort
		expectedStart int32
		expectedEnd   int32
		expectError   bool
	}{
		{
			name: "valid port range match",
			annotation: "1000-2000:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 1000,
			expectedEnd:   2000,
			expectError:   false,
		},
		{
			name: "multiple port ranges match first",
			annotation: "1000-2000:443,3000-4000:80",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 1000,
			expectedEnd:   2000,
			expectError:   false,
		},
		{
			name: "multiple port ranges match second",
			annotation: "1000-2000:443,3000-4000:80",
			port: v1.ServicePort{
				Port: 80,
			},
			expectedStart: 3000,
			expectedEnd:   4000,
			expectError:   false,
		},
		{
			name: "no match returns zero",
			annotation: "1000-2000:443,3000-4000:80",
			port: v1.ServicePort{
				Port: 8080,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   false,
		},
		{
			name: "invalid format missing colon",
			annotation: "1000-2000",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "invalid format missing dash",
			annotation: "1000:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "invalid format start port not number",
			annotation: "abc-2000:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "invalid format end port not number",
			annotation: "1000-xyz:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "invalid format start port >= end port",
			annotation: "2000-1000:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "invalid format start port < 1",
			annotation: "0-2000:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "invalid format end port > 65535",
			annotation: "1000-65536:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 0,
			expectedEnd:   0,
			expectError:   true,
		},
		{
			name: "valid boundary values",
			annotation: "1-65535:443",
			port: v1.ServicePort{
				Port: 443,
			},
			expectedStart: 1,
			expectedEnd:   65535,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := portRange(tt.annotation, tt.port)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStart, start)
				assert.Equal(t, tt.expectedEnd, end)
			}
		})
	}
}
