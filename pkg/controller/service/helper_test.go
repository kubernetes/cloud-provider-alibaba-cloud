package service

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestIsServiceHashChanged(t *testing.T) {
	base := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type: v1.ServiceTypeLoadBalancer,
		},
	}
	base.Annotations[Annotation(LoadBalancerName)] = "slb-base"
	baseHash := getServiceHash(base)

	svcAnnoChanged := base.DeepCopy()
	svcAnnoChanged.Annotations[Annotation(LoadBalancerName)] = "slb-anno-changed"
	annoHash := getServiceHash(svcAnnoChanged)
	assert.NotEqual(t, baseHash, annoHash)

	svcMetaChanged := base.DeepCopy()
	svcMetaChanged.Labels = map[string]string{"app": "test"}
	hash := getServiceHash(svcMetaChanged)
	assert.Equal(t, baseHash, hash)

	svcSpecChanged := base.DeepCopy()
	svcSpecChanged.Spec.ExternalTrafficPolicy = "Cluster"
	hash = getServiceHash(svcSpecChanged)
	assert.NotEqual(t, baseHash, hash)
}

func TestBatch(t *testing.T) {
	sum := 0
	addFunc := func(a []interface{}) error {
		for _, num := range a {
			i, _ := num.(int)
			sum += i
		}
		return nil
	}
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if err := Batch(nums, 3, addFunc); err != nil {
		t.Fatalf("Batch error: %s", err.Error())
	}
	assert.Equal(t, sum, 55)
}
