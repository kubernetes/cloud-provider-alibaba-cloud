package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIPv4(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid IPv4 address",
			address:  "192.168.1.1",
			expected: true,
		},
		{
			name:     "invalid IPv4 address with too many dots",
			address:  "192.168.1.1.1",
			expected: true, // Still counts as IPv4 as it has less than 2 colons
		},
		{
			name:     "IPv6 address",
			address:  "2001:db8::1",
			expected: false,
		},
		{
			name:     "IPv6 address with brackets",
			address:  "[2001:db8::1]",
			expected: false,
		},
		{
			name:     "localhost IPv4",
			address:  "127.0.0.1",
			expected: true,
		},
		{
			name:     "localhost IPv6",
			address:  "::1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIPv4(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsIPv6(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid IPv4 address",
			address:  "192.168.1.1",
			expected: false,
		},
		{
			name:     "IPv6 address",
			address:  "2001:db8::1",
			expected: true,
		},
		{
			name:     "IPv6 address with brackets",
			address:  "[2001:db8::1]",
			expected: true,
		},
		{
			name:     "localhost IPv4",
			address:  "127.0.0.1",
			expected: false,
		},
		{
			name:     "localhost IPv6",
			address:  "::1",
			expected: true,
		},
		{
			name:     "IPv4-mapped IPv6 address",
			address:  "::ffff:192.0.2.1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIPv6(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}