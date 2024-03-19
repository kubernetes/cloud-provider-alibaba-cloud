package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func TestIsProxyProtocolEnabledOn(t *testing.T) {
	type test struct {
		annotation string
		protocol   string
		port       int32

		wantErr error
		result  bool
	}
	tests := []test{
		{annotation: "", protocol: "tcp", port: 80, wantErr: nil, result: true},
		{annotation: "udp:53,tcp:53", protocol: "tcp", port: 80, wantErr: nil, result: false},
		{annotation: "tcp|80,", protocol: "tcp", port: 80, wantErr: fmt.Errorf("port and "+
			"protocol format must be like 'https:443' with colon separated. got=[%+v]", "[tcp|80]"), result: false},
		{annotation: "tcp:80|udp:53", protocol: "tcp", port: 80, wantErr: nil, result: false},
		{annotation: "tcp:80,tcp:81", protocol: "tcp", port: 81, wantErr: nil, result: true},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("[%v/%v]", i+1, len(tests)), func(t *testing.T) {
			result, err := IsProxyProtocolEnabledOn(tc.annotation, tc.protocol, tc.port)

			assert.Equal(t, result, tc.result)
			if err != tc.wantErr && err.Error() != tc.wantErr.Error() {
				t.Errorf("%s not captured. %s", tc.wantErr, err.Error())
			}
		})
	}
}
