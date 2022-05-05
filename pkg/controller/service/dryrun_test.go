package service

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestMapFull(t *testing.T) {
	assert.Equal(t, mapfull(), true)

	initial.Store("test-ns/test-svc", 0)
	assert.Equal(t, mapfull(), false)

	initial.Store("test-ns/test-svc", 1)
	assert.Equal(t, mapfull(), true)
}

func TestInitMap(t *testing.T) {
	svcs := &v1.ServiceList{
		Items: []v1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
				},
			},
		},
	}

	objs := []runtime.Object{svcs}
	cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	initMap(cl)
}
