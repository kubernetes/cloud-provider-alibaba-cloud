package helper

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestIsServiceHashChanged(t *testing.T) {
	base := getDefaultService()
	base.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name"] = "slb-base"
	baseHash := GetServiceHash(base)

	svcAnnoChanged := base.DeepCopy()
	svcAnnoChanged.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name"] = "slb-anno-changed"
	annoHash := GetServiceHash(svcAnnoChanged)
	assert.NotEqual(t, baseHash, annoHash)

	svcMetaChanged := base.DeepCopy()
	svcMetaChanged.Labels = map[string]string{"app": "test"}
	hash := GetServiceHash(svcMetaChanged)
	assert.Equal(t, baseHash, hash)

	svcSpecChanged := base.DeepCopy()
	svcSpecChanged.Spec.ExternalTrafficPolicy = "Cluster"
	hash = GetServiceHash(svcSpecChanged)
	assert.NotEqual(t, baseHash, hash)

	svcNewAttrChanged := base.DeepCopy()
	svcNewAttrChanged.Spec.PublishNotReadyAddresses = true
	hash = GetServiceHash(svcNewAttrChanged)
	assert.Equal(t, baseHash, hash)
}

func getDefaultService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   v1.NamespaceDefault,
			Annotations: make(map[string]string),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "tcp",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					NodePort:   80,
					Protocol:   v1.ProtocolTCP,
				},
			},
			Type:                  v1.ServiceTypeLoadBalancer,
			ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
		},
	}
}

func TestIsTunnelTypeService(t *testing.T) {
	svc := getDefaultService()
	svc.Annotations = map[string]string{
		"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-tunnel-type": "tunnel",
	}
	assert.Equal(t, true, IsTunnelTypeService(svc))

	svc.Annotations = map[string]string{
		"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-tunnel-type": "",
	}
	assert.Equal(t, true, IsTunnelTypeService(svc))
}
