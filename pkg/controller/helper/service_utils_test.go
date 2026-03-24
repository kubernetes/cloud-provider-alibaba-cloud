package helper

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/feature"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
)

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

func TestIsServiceHashChanged(t *testing.T) {
	base := getDefaultService()
	base.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name"] = "slb-base"
	baseHash := GetServiceHash(base)
	base.Labels = map[string]string{
		LabelServiceHash: baseHash,
	}

	t.Run("empty service hash label", func(t *testing.T) {
		svcEmptyLabel := base.DeepCopy()
		svcEmptyLabel.Labels = map[string]string{}
		assert.True(t, IsServiceHashChanged(svcEmptyLabel))
	})
	t.Run("service annotation changed", func(t *testing.T) {
		svcAnnoChanged := base.DeepCopy()
		svcAnnoChanged.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name"] = "slb-anno-changed"
		assert.True(t, IsServiceHashChanged(svcAnnoChanged))
	})

	t.Run("service label changed", func(t *testing.T) {
		svcLabelChanged := base.DeepCopy()
		svcLabelChanged.Labels["app"] = "test"
		assert.False(t, IsServiceHashChanged(svcLabelChanged))
	})

	t.Run("external traffic policy changed", func(t *testing.T) {
		svcSpecChanged := base.DeepCopy()
		svcSpecChanged.Spec.ExternalTrafficPolicy = "Cluster"
		assert.True(t, IsServiceHashChanged(svcSpecChanged))
	})

	t.Run("unimportant spec fields changed", func(t *testing.T) {
		svcNewAttrChanged := base.DeepCopy()
		svcNewAttrChanged.Spec.PublishNotReadyAddresses = true
		assert.False(t, IsServiceHashChanged(svcNewAttrChanged))
	})
}

func TestIsENIBackendType(t *testing.T) {
	oldEnv := os.Getenv("SERVICE_FORCE_BACKEND_ENI")
	oldServiceBackendType := ctrlCfg.CloudCFG.Global.ServiceBackendType
	defer func() {
		err := os.Setenv("SERVICE_FORCE_BACKEND_ENI", oldEnv)
		assert.NoError(t, err)
		ctrlCfg.CloudCFG.Global.ServiceBackendType = oldServiceBackendType
	}()

	err := os.Setenv("SERVICE_FORCE_BACKEND_ENI", "")
	assert.NoError(t, err)
	ctrlCfg.CloudCFG.Global.ServiceBackendType = model.ECSBackendType

	t.Run("not eni backend type", func(t *testing.T) {
		svc := getDefaultService()
		assert.False(t, IsENIBackendType(svc))
	})

	t.Run("backend type from service annotations", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations = map[string]string{
			BackendType: model.ENIBackendType,
		}
		assert.True(t, IsENIBackendType(svc))
	})

	t.Run("force set eni backend type from environment variable", func(t *testing.T) {
		err := os.Setenv("SERVICE_FORCE_BACKEND_ENI", "true")
		assert.NoError(t, err)
		svc := getDefaultService()
		assert.True(t, IsENIBackendType(svc))
		err = os.Setenv("SERVICE_FORCE_BACKEND_ENI", "")
		assert.NoError(t, err)
	})

	t.Run("eni backend type from cloud config", func(t *testing.T) {
		ctrlCfg.CloudCFG.Global.ServiceBackendType = model.ENIBackendType
		svc := getDefaultService()
		assert.True(t, IsENIBackendType(svc))
		ctrlCfg.CloudCFG.Global.ServiceBackendType = model.ECSBackendType
	})
}

func TestIsTunnelTypeService(t *testing.T) {
	t.Run("service with tunnel type annotation", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations = map[string]string{
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-tunnel-type": "tunnel",
		}
		assert.True(t, IsTunnelTypeService(svc))
	})

	t.Run("service with empty tunnel type annotation", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations = map[string]string{
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-tunnel-type": "",
		}
		assert.True(t, IsTunnelTypeService(svc))
	})

	t.Run("service without tunnel type annotation", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations = map[string]string{
			"other-annotation": "value",
		}
		assert.False(t, IsTunnelTypeService(svc))
	})

	t.Run("service with nil annotations", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations = nil
		assert.False(t, IsTunnelTypeService(svc))
	})
}

func TestRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		callCount := 0
		err := Retry(nil, func(svc *v1.Service) error {
			callCount++
			return nil
		}, nil)

		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("retry until success", func(t *testing.T) {
		callCount := 0
		err := Retry(&wait.Backoff{Steps: 3}, func(svc *v1.Service) error {
			callCount++
			if callCount < 2 {
				return fmt.Errorf("please try again")
			}
			return nil
		}, nil)

		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("retry until max attempts", func(t *testing.T) {
		callCount := 0
		err := Retry(&wait.Backoff{Steps: 3}, func(svc *v1.Service) error {
			callCount++
			return fmt.Errorf("please try again")
		}, nil)

		assert.Error(t, err) // ExponentialBackoff returns nil when it exhausts steps
		assert.Equal(t, 3, callCount)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		callCount := 0
		err := Retry(&wait.Backoff{Steps: 3}, func(svc *v1.Service) error {
			callCount++
			return fmt.Errorf("permanent error")
		}, nil)

		assert.NoError(t, err)        // ExponentialBackoff returns nil even for non-retryable errors
		assert.Equal(t, 1, callCount) // Should not retry
	})
}

func TestNeedCLB(t *testing.T) {
	t.Run("no load balancer type", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeClusterIP
		assert.False(t, NeedCLB(svc))
	})

	t.Run("has load balancer class", func(t *testing.T) {
		svc := getDefaultService()
		class := "some-class"
		svc.Spec.LoadBalancerClass = &class
		assert.False(t, NeedCLB(svc))
	})

	t.Run("has tunnel type annotation", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[TunnelType] = "tunnel"
		assert.False(t, NeedCLB(svc))
	})

	t.Run("has load balancer class annotation", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[LoadBalancerClass] = "class"
		assert.False(t, NeedCLB(svc))
	})

	t.Run("has LoadBalancerType annotation with feature gate enabled", func(t *testing.T) {
		oldValue := feature.DefaultMutableFeatureGate.Enabled(ctrlCfg.LoadBalancerTypeAnnotation)
		feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			ctrlCfg.LoadBalancerTypeAnnotation: true,
		})
		defer func() {
			feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				ctrlCfg.LoadBalancerTypeAnnotation: oldValue,
			})
		}()
		svc := getDefaultService()
		svc.Annotations[LoadBalancerType] = "clb"
		assert.False(t, NeedCLB(svc))
	})

	t.Run("default case - need CLB", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations = map[string]string{}
		assert.True(t, NeedCLB(svc))
	})
}

func TestNeedNLB(t *testing.T) {
	t.Run("not load balancer type", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeClusterIP
		assert.False(t, NeedNLB(svc))
	})

	t.Run("has NLB load balancer class", func(t *testing.T) {
		svc := getDefaultService()
		class := NLBClass
		svc.Spec.LoadBalancerClass = &class
		assert.True(t, NeedNLB(svc))
	})

	t.Run("has NLB type annotation without feature gate enabled", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[LoadBalancerType] = "nlb"
		assert.False(t, NeedNLB(svc))
	})

	t.Run("has NLB type annotation with feature gate enabled", func(t *testing.T) {
		oldValue := feature.DefaultMutableFeatureGate.Enabled(ctrlCfg.LoadBalancerTypeAnnotation)
		feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
			ctrlCfg.LoadBalancerTypeAnnotation: true,
		})
		defer func() {
			feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
				ctrlCfg.LoadBalancerTypeAnnotation: oldValue,
			})
		}()
		svc := getDefaultService()
		svc.Annotations[LoadBalancerType] = "nlb"
		assert.True(t, NeedNLB(svc))
	})

	t.Run("non NLB load balancer class", func(t *testing.T) {
		svc := getDefaultService()
		class := "other-class"
		svc.Spec.LoadBalancerClass = &class
		assert.False(t, NeedNLB(svc))
	})

	t.Run("non NLB type annotation", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[LoadBalancerType] = "clb"
		assert.False(t, NeedNLB(svc))
	})
}

func TestGetServiceTrafficPolicy(t *testing.T) {
	t.Run("ENI backend type", func(t *testing.T) {
		svc := getDefaultService()
		svc.Annotations[BackendType] = model.ENIBackendType
		policy, err := GetServiceTrafficPolicy(svc)
		assert.NoError(t, err)
		assert.Equal(t, ENITrafficPolicy, policy)
	})

	t.Run("ClusterIP service", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeClusterIP
		_, err := GetServiceTrafficPolicy(svc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster service type just support eni mode")
	})

	t.Run("Local traffic policy", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		policy, err := GetServiceTrafficPolicy(svc)
		assert.NoError(t, err)
		assert.Equal(t, LocalTrafficPolicy, policy)
	})

	t.Run("Cluster traffic policy", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
		policy, err := GetServiceTrafficPolicy(svc)
		assert.NoError(t, err)
		assert.Equal(t, ClusterTrafficPolicy, policy)
	})
}

func TestIsLocalModeService(t *testing.T) {
	t.Run("local mode", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		assert.True(t, IsLocalModeService(svc))
	})

	t.Run("cluster mode", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
		assert.False(t, IsLocalModeService(svc))
	})
}

func TestIsClusterIPService(t *testing.T) {
	t.Run("ClusterIP service", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeClusterIP
		assert.True(t, IsClusterIPService(svc))
	})

	t.Run("LoadBalancer service", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		assert.False(t, IsClusterIPService(svc))
	})

	t.Run("NodePort service", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeNodePort
		assert.False(t, IsClusterIPService(svc))
	})
}

func TestNeedDeleteLoadBalancer(t *testing.T) {
	now := metav1.Now()

	t.Run("service with deletion timestamp", func(t *testing.T) {
		svc := getDefaultService()
		svc.DeletionTimestamp = &now
		assert.True(t, NeedDeleteLoadBalancer(svc))
	})

	t.Run("service type changed to ClusterIP", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeClusterIP
		assert.True(t, NeedDeleteLoadBalancer(svc))
	})

	t.Run("service type changed to NodePort", func(t *testing.T) {
		svc := getDefaultService()
		svc.Spec.Type = v1.ServiceTypeNodePort
		assert.True(t, NeedDeleteLoadBalancer(svc))
	})

	t.Run("normal LoadBalancer service", func(t *testing.T) {
		svc := getDefaultService()
		assert.False(t, NeedDeleteLoadBalancer(svc))
	})
}

func TestRetryOnErrorContains(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		callCount := 0
		err := RetryOnErrorContains("retry", func() error {
			callCount++
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("retry on matching error", func(t *testing.T) {
		callCount := 0
		err := RetryOnErrorContains("retry", func() error {
			callCount++
			if callCount < 3 {
				return fmt.Errorf("please retry this operation")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("no retry on non-matching error", func(t *testing.T) {
		callCount := 0
		err := RetryOnErrorContains("retry", func() error {
			callCount++
			return fmt.Errorf("permanent error")
		})
		assert.Error(t, err)
		assert.Equal(t, 1, callCount)
	})
}

func TestIs7LayerProtocol(t *testing.T) {
	tests := []struct {
		protocol string
		want     bool
	}{
		{model.HTTP, true},
		{model.HTTPS, true},
		{model.TCP, false},
		{model.UDP, false},
		{"grpc", false},
	}

	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			assert.Equal(t, tt.want, Is7LayerProtocol(tt.protocol))
		})
	}
}

func TestIs4LayerProtocol(t *testing.T) {
	tests := []struct {
		protocol string
		want     bool
	}{
		{model.TCP, true},
		{model.UDP, true},
		{model.HTTP, false},
		{model.HTTPS, false},
		{"grpc", false},
	}

	for _, tt := range tests {
		t.Run(tt.protocol, func(t *testing.T) {
			assert.Equal(t, tt.want, Is4LayerProtocol(tt.protocol))
		})
	}
}

func TestIsServiceOwnIngress(t *testing.T) {
	t.Run("nil service", func(t *testing.T) {
		assert.False(t, IsServiceOwnIngress(nil))
	})

	t.Run("service without ingress", func(t *testing.T) {
		svc := getDefaultService()
		assert.False(t, IsServiceOwnIngress(svc))
	})

	t.Run("service with empty ingress list", func(t *testing.T) {
		svc := getDefaultService()
		svc.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{}
		assert.False(t, IsServiceOwnIngress(svc))
	})

	t.Run("service with ingress", func(t *testing.T) {
		svc := getDefaultService()
		svc.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
			{IP: "1.2.3.4"},
		}
		assert.True(t, IsServiceOwnIngress(svc))
	})
}
