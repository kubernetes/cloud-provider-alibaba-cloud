package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLogEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		endpoint *v1.Endpoints
	}{
		{
			name:     "nil endpoints",
			endpoint: nil,
		},
		{
			name:     "empty endpoints",
			endpoint: &v1.Endpoints{},
		},
		{
			name: "endpoints with addresses",
			endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Subsets: []v1.EndpointSubset{
					{
						Addresses: []v1.EndpointAddress{
							{IP: "192.168.1.1"},
							{IP: "192.168.1.2"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				_ = LogEndpoints(tt.endpoint)
			})
		})
	}
}

func TestLogEndpointSlice(t *testing.T) {
	tests := []struct {
		name string
		es   *discovery.EndpointSlice
	}{
		{
			name: "nil endpoint slice",
			es:   nil,
		},
		{
			name: "empty endpoint slice",
			es:   &discovery.EndpointSlice{},
		},
		{
			name: "endpoint slice with endpoints",
			es: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Addresses: []string{"192.168.1.1", "192.168.1.2"},
					},
					{
						Addresses: []string{"192.168.1.3"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				_ = LogEndpointSlice(tt.es)
			})
		})
	}
}

func TestLogEndpointSliceList(t *testing.T) {
	tests := []struct {
		name   string
		esList []discovery.EndpointSlice
	}{
		{
			name:   "nil endpoint slice list",
			esList: nil,
		},
		{
			name:   "empty endpoint slice list",
			esList: []discovery.EndpointSlice{},
		},
		{
			name: "endpoint slice list with endpoints",
			esList: []discovery.EndpointSlice{
				{
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"192.168.1.1", "192.168.1.2"},
						},
					},
				},
				{
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"192.168.1.3"},
						},
					},
				},
			},
		},
		{
			name: "endpoint slice list with ready and not ready endpoints",
			esList: []discovery.EndpointSlice{
				{
					Endpoints: []discovery.Endpoint{
						{
							Addresses: []string{"192.168.1.1"},
							Conditions: discovery.EndpointConditions{
								Ready: func() *bool { r := true; return &r }(),
							},
						},
						{
							Addresses: []string{"192.168.1.2"},
							Conditions: discovery.EndpointConditions{
								Ready: func() *bool { r := false; return &r }(),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				_ = LogEndpointSliceList(tt.esList)
			})
		})
	}
}
