package util

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiextfakes "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	discoveryfake "k8s.io/client-go/discovery/fake"
)

type testObject struct {
	metav1.ObjectMeta
}

func TestNamespacedName(t *testing.T) {
	obj := &testObject{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	}

	namespacedName := NamespacedName(obj)
	assert.Equal(t, "test-namespace", namespacedName.Namespace)
	assert.Equal(t, "test-name", namespacedName.Name)
}

func TestKey(t *testing.T) {
	obj := &testObject{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
	}

	key := Key(obj)
	assert.Equal(t, "test-namespace/test-name", key)
}

func TestPrettyJson(t *testing.T) {
	obj := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "test-user",
		Age:  25,
	}

	result := PrettyJson(obj)
	expected := `{
    "name": "test-user",
    "age": 25
}`
	assert.Equal(t, expected, result)

	invalidObj := make(chan int)
	result = PrettyJson(invalidObj)
	assert.Equal(t, "", result)
}

func TestMergeStringMap(t *testing.T) {
	// Test basic merge
	map1 := map[string]string{"a": "1", "b": "2"}
	map2 := map[string]string{"c": "3", "d": "4"}
	result := MergeStringMap(map1, map2)
	expected := map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"}
	assert.Equal(t, expected, result)

	// Test with overlapping keys (first value should be kept)
	map3 := map[string]string{"a": "1", "b": "2"}
	map4 := map[string]string{"a": "3", "d": "4"}
	result = MergeStringMap(map3, map4)
	expected = map[string]string{"a": "1", "b": "2", "d": "4"}
	assert.Equal(t, expected, result)

	// Test with empty maps
	empty := map[string]string{}
	result = MergeStringMap(empty, map1)
	assert.Equal(t, map1, result)

	result = MergeStringMap(map1, empty)
	assert.Equal(t, map1, result)

	// Test with multiple maps
	map5 := map[string]string{"e": "5"}
	result = MergeStringMap(map1, map2, map5)
	expected = map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5"}
	assert.Equal(t, expected, result)
}

func TestIsStringSliceEqual(t *testing.T) {
	// Test equal slices
	s1 := []string{"a", "b", "c"}
	s2 := []string{"a", "b", "c"}
	assert.True(t, IsStringSliceEqual(s1, s2))

	// Test equal slices with different order
	s3 := []string{"c", "b", "a"}
	assert.True(t, IsStringSliceEqual(s1, s3))

	// Test equal slices with case differences
	s4 := []string{"A", "B", "C"}
	assert.True(t, IsStringSliceEqual(s1, s4))

	// Test different length slices
	s5 := []string{"a", "b"}
	assert.False(t, IsStringSliceEqual(s1, s5))

	// Test different content
	s6 := []string{"a", "b", "d"}
	assert.False(t, IsStringSliceEqual(s1, s6))

	// Test empty slices
	s7 := []string{}
	s8 := []string{}
	assert.True(t, IsStringSliceEqual(s7, s8))

	// Test one empty slice
	assert.False(t, IsStringSliceEqual(s1, s7))
}

func TestClusterVersionAtLeast(t *testing.T) {
	client := apiextfakes.NewSimpleClientset()
	fakeDiscovery, ok := client.Discovery().(*discoveryfake.FakeDiscovery)
	if !ok {
		t.Fatal("failed to create fake discovery client")
	}
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "20",
		GitVersion: "v1.20.0",
	}

	ok, err := ClusterVersionAtLeast(client, "v1.18.0")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = ClusterVersionAtLeast(client, "v1.22.0")
	assert.NoError(t, err)
	assert.False(t, ok)

	_, err = ClusterVersionAtLeast(client, "invalid")
	assert.Error(t, err)

	fakeDiscovery.FakedServerVersion = &version.Info{
		GitVersion: "invalid",
	}
	_, err = ClusterVersionAtLeast(client, "v1.18.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected error parsing running Kubernetes version")
}

func TestRetryImmediateOnError(t *testing.T) {
	// Test successful function on first try
	var attempts int
	fn := func() error {
		attempts++
		return nil
	}

	err := RetryImmediateOnError(time.Millisecond, time.Millisecond*10, nil, fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)

	// Test retryable error that eventually succeeds
	attempts = 0
	var shouldFail = true
	fn = func() error {
		attempts++
		if shouldFail && attempts < 3 {
			return &testError{"retryable error"}
		}
		return nil
	}

	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "retryable")
	}

	err = RetryImmediateOnError(time.Millisecond, time.Millisecond*50, retryable, fn)
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)

	// Test non-retryable error
	attempts = 0
	fn = func() error {
		attempts++
		return &testError{"normal error"}
	}

	err = RetryImmediateOnError(time.Millisecond, time.Millisecond*50, retryable, fn)
	assert.Error(t, err)
	assert.Equal(t, 1, attempts)
	assert.Contains(t, err.Error(), "normal error")
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
