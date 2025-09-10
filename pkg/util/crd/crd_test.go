package crd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
)

func TestConf_getName(t *testing.T) {
	conf := &Conf{
		NamePlural: "examples",
		Group:      "test.example.com",
	}

	expected := "examples.test.example.com"
	actual := conf.getName()

	assert.Equal(t, expected, actual)
}

func TestNewClient(t *testing.T) {
	fakeClient := apiextfake.NewSimpleClientset()
	client := NewClient(fakeClient)

	assert.NotNil(t, client)
	assert.Equal(t, fakeClient, client.client)
}

func TestNewCustomClient(t *testing.T) {
	fakeClient := apiextfake.NewSimpleClientset()
	client := NewCustomClient(fakeClient)

	assert.NotNil(t, client)
	assert.Equal(t, fakeClient, client.client)
}

func TestAddDefaultCaregories(t *testing.T) {
	client := &Client{}
	testCases := []struct {
		name           string
		input          []string
		expectedLength int
	}{
		{
			name:           "Empty categories",
			input:          []string{},
			expectedLength: 2, // Should add "all" and "kooper"
		},
		{
			name:           "Existing categories without defaults",
			input:          []string{"custom"},
			expectedLength: 3, // Should add "all" and "kooper"
		},
		{
			name:           "Existing categories with some defaults",
			input:          []string{"custom", "all"},
			expectedLength: 3, // Should add "kooper"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.addDefaultCaregories(tc.input)
			assert.Len(t, result, tc.expectedLength)
		})
	}
}

func TestCreateSubresources(t *testing.T) {
	client := &Client{}

	t.Run("Both subresources disabled", func(t *testing.T) {
		conf := Conf{
			EnableStatusSubresource: false,
			EnableScaleSubresource:  nil,
		}
		result := client.createSubresources(conf)
		assert.Nil(t, result)
	})

	t.Run("Status subresrouce enabled", func(t *testing.T) {
		conf := Conf{
			EnableStatusSubresource: true,
			EnableScaleSubresource:  nil,
		}
		result := client.createSubresources(conf)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Status)
	})

	t.Run("Scale subresource enabled", func(t *testing.T) {
		scale := &apiextv1beta1.CustomResourceSubresourceScale{
			SpecReplicasPath: ".spec.replicas",
		}
		conf := Conf{
			EnableStatusSubresource: false,
			EnableScaleSubresource:  scale,
		}
		result := client.createSubresources(conf)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Scale)
	})

	t.Run("Both subresrouces enabled", func(t *testing.T) {
		scale := &apiextv1beta1.CustomResourceSubresourceScale{
			SpecReplicasPath: ".spec.replicas",
		}
		conf := Conf{
			EnableStatusSubresource: true,
			EnableScaleSubresource:  scale,
		}
		result := client.createSubresources(conf)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Status)
		assert.NotNil(t, result.Scale)
	})

}

func TestValidateCRD(t *testing.T) {
	// Create a fake client that simulates server version response
	fakeClient := apiextfake.NewSimpleClientset()
	client := &Client{client: fakeClient}

	err := client.validateCRD()
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	fakeClient := apiextfake.NewSimpleClientset()
	client := &Client{client: fakeClient}

	err := client.Delete("test.crd.com")
	assert.Error(t, err)
}

func TestWaitToBePresent_Timeout(t *testing.T) {
	// Create a fake client
	fakeClient := apiextfake.NewSimpleClientset()
	client := &Client{client: fakeClient}

	// Test timeout scenario - CRD doesn't exist
	timeout := 10 * time.Millisecond
	err := client.WaitToBePresent("nonexistent.crd.com", timeout)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for CRD")
}

func TestEnsurePresent_CreationError(t *testing.T) {
	fakeClient := apiextfake.NewSimpleClientset()
	client := &Client{client: fakeClient}
	fakeDiscovery, ok := fakeClient.Discovery().(*fake.FakeDiscovery)
	if !ok {
		t.Fatal("Failed to convert fake client to DiscoveryInterface")
	}
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "19",
		GitVersion: "v1.19.0",
	}

	conf := Conf{
		Kind:                    "TestKind",
		NamePlural:              "test",
		Group:                   "test.example.com",
		Version:                 "v1",
		Scope:                   NamespaceScoped,
		EnableStatusSubresource: false,
	}

	err := client.EnsurePresent(conf)
	assert.NoError(t, err)
}
