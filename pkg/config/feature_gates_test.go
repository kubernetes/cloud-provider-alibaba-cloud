package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/version"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	k8stesting "k8s.io/client-go/discovery/fake"
	"k8s.io/component-base/featuregate"
)

func TestBindFeatureGates(t *testing.T) {
	tests := []struct {
		name         string
		features     string
		k8sVersion   string
		expectError  bool
		validateFunc func(*testing.T)
	}{
		{
			name:        "valid feature gates",
			features:    "EndpointSlice=true,IPv6DualStack=false",
			k8sVersion:  "v1.20.0",
			expectError: false,
			validateFunc: func(t *testing.T) {
				assert.True(t, utilfeature.DefaultMutableFeatureGate.Enabled(EndpointSlice))
				assert.False(t, utilfeature.DefaultMutableFeatureGate.Enabled(IPv6DualStack))
			},
		},
		{
			name:        "invalid feature gates",
			features:    "EndpointSlice=true,IPv6DualStack",
			k8sVersion:  "v1.20.0",
			expectError: true,
		},
		{
			name:        "invalid feature gates 2",
			features:    "EndpointSlice=123",
			k8sVersion:  "v1.20.0",
			expectError: true,
		},
		{
			name:        "disable EndpointSlice for k8s < v1.20.0",
			features:    "EndpointSlice=true",
			k8sVersion:  "v1.19.0",
			expectError: false,
			validateFunc: func(t *testing.T) {
				assert.False(t, utilfeature.DefaultMutableFeatureGate.Enabled(EndpointSlice))
			},
		},
		{
			name:        "disable IPv6DualStack for k8s < v1.20.0",
			features:    "IPv6DualStack=true",
			k8sVersion:  "v1.19.0",
			expectError: false,
			validateFunc: func(t *testing.T) {
				assert.False(t, utilfeature.DefaultMutableFeatureGate.Enabled(IPv6DualStack))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFeatureGates()

			client := apiextfake.NewSimpleClientset()
			fakeDiscovery, ok := client.Discovery().(*k8stesting.FakeDiscovery)
			assert.True(t, ok)

			fakeDiscovery.FakedServerVersion = &version.Info{
				GitVersion: tt.k8sVersion,
			}

			err := BindFeatureGates(client, tt.features)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}
		})
	}
}

func resetFeatureGates() {
	utilfeature.DefaultMutableFeatureGate = featuregate.NewFeatureGate()
	_ = utilfeature.DefaultMutableFeatureGate.Add(CloudProviderFeatureGates)
}
